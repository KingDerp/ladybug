package server

import (
	"context"

	"ladybug/database"
)

type Message struct {
	Id          string `json:"id"`
	BuyerSent   bool   `json:"buyerSent"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"createdAt"`
}

func MessageFromDB(message *database.Message) *Message {
	return &Message{
		Id:          message.Id,
		BuyerSent:   message.BuyerSent,
		Description: message.Description,
		CreatedAt:   message.CreatedAt.Unix(),
	}
}

func MessagesFromDB(messages []*database.Message) []*Message {
	out := []*Message{}
	for _, m := range messages {
		out = append(out, MessageFromDB(m))
	}

	return out
}

type PagedBuyerMessagesByConversationIdReq struct {
	Offset         int64  `json:"offset"`
	ConversationId string `json:"conversationId"`
}

type PagedBuyerMessagesByConversationIdResp struct {
	Offset   int64      `json:"offset"`
	Messages []*Message `json:"messages"`
}

func (u *BuyerServer) PagedBuyerMessagesByConversationId(ctx context.Context,
	req *PagedBuyerMessagesByConversationIdReq) (resp *PagedBuyerMessagesByConversationIdResp, err error) {

	var messages []*database.Message
	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		conversation, err := tx.Get_Conversation_By_Id(ctx, database.Conversation_Id(req.ConversationId))
		if err != nil {
			return err
		}

		messages, err = tx.Limited_Message_By_ConversationPk_OrderBy_Desc_CreatedAt(ctx,
			database.Message_ConversationPk(conversation.Pk), messageRequestLimit, req.Offset)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	offset := req.Offset + messageRequestLimit

	return &PagedBuyerMessagesByConversationIdResp{
		Messages: MessagesFromDB(messages),
		Offset:   offset,
	}, nil
}
