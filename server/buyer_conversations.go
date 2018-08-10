package server

import (
	"context"
	"ladybug/database"
)

type PagedBuyerConversationsReq struct {
	BuyerPk   int64
	PageToken string `json:"pageToken"`
}

type Conversation struct {
	Id string `json:"id"`
}

type PagedBuyerConversationResp struct {
	PageToken     string          `json:"pageToken"`
	Conversations []*Conversation `json:"conversations"`
}

func ConversationFromDB(conv *database.Conversation) *Conversation {
	return &Conversation{
		Id: conv.Id,
	}
}

func ConversationsFromDB(conversations []*database.Conversation) []*Conversation {
	out := []*Conversation{}
	for _, c := range conversations {
		out = append(out, ConversationFromDB(c))
	}

	return out
}

func (u *BuyerServer) GetPagedBuyerConversations(ctx context.Context, req *PagedBuyerConversationsReq) (
	resp *PagedBuyerConversationResp, err error) {

	var conversations []*database.Conversation
	var ctoken string
	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		conversations, ctoken, err = tx.Paged_Conversation_By_BuyerPk(ctx,
			database.Conversation_BuyerPk(req.BuyerPk), conversationRequestLimit, req.PageToken)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &PagedBuyerConversationResp{
		PageToken:     ctoken,
		Conversations: ConversationsFromDB(conversations),
	}, nil

}
