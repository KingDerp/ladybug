package server

import (
	"context"

	"ladybug/database"

	uuid "github.com/satori/go.uuid"
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

type PostBuyerMessageToConversationReq struct {
	BuyerPk            int64
	VendorId           string `json:"vendorId"`
	MessageDescription string `json:"messageDescription"`
}

type PostBuyerMessageToConversationResp struct {
	Message *Message `json:"messages"`
}

func (u *BuyerServer) PostBuyerMessageToConversation(ctx context.Context,
	req *PostBuyerMessageToConversationReq) (resp *PostBuyerMessageToConversationResp, err error) {

	var message *database.Message
	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {

		vendor_pk_field, err := tx.Get_Vendor_Pk_By_Id(ctx,
			database.Vendor_Id(req.VendorId))
		if err != nil {
			return err
		}

		conversation, err := tx.Find_Conversation_By_VendorPk_And_BuyerPk(ctx,
			database.Conversation_VendorPk(vendor_pk_field.Pk),
			database.Conversation_BuyerPk(req.BuyerPk))
		if err != nil {
			return err
		}

		//if buyer and vendor have not had a conversation before create new conversation
		if conversation == nil {
			conversation, err = tx.Create_Conversation(ctx,
				database.Conversation_VendorPk(vendor_pk_field.Pk),
				database.Conversation_BuyerPk(req.BuyerPk),
				database.Conversation_BuyerUnread(false),
				database.Conversation_VendorUnread(true),
				database.Conversation_MessageCount(1),
				database.Conversation_Id(uuid.NewV4().String()))
			if err != nil {
				return err
			}
		} else {
			conversation_updates := database.Conversation_Update_Fields{
				MessageCount: database.Conversation_MessageCount(conversation.MessageCount + 1),
				VendorUnread: database.Conversation_VendorUnread(true),
			}

			conversation, err = tx.Update_Conversation_By_Pk(ctx,
				database.Conversation_Pk(conversation.Pk),
				database.Conversation_Update_Fields(conversation_updates))
			if err != nil {
				return err
			}
		}

		message, err = tx.Create_Message(ctx,
			database.Message_Id(uuid.NewV4().String()),
			database.Message_BuyerSent(true),
			database.Message_Description(req.MessageDescription),
			database.Message_ConversationPk(conversation.Pk),
			database.Message_ConversationNumber(conversation.MessageCount))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &PostBuyerMessageToConversationResp{
		Message: MessageFromDB(message),
	}, nil
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

		//TODO(mac): once you've tested the incrementing property of conversation message count come
		//back here and change this query to get messages by message count in descending order
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
