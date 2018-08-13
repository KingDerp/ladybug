package server

import (
	"context"

	"ladybug/database"
)

type VendorConversationsUnreadReq struct {
	VendorPk int64
}

type VendorConversationsUnreadResp struct {
	Conversations []*Conversation `json:"conversations"`
}

func (u *VendorServer) GetVendorCoversationsUnread(ctx context.Context,
	req *VendorConversationsUnreadReq) (resp *VendorConversationsUnreadResp, err error) {

	var conversations []*database.Conversation
	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		conversations, err = tx.All_Conversation_By_VendorPk_And_VendorUnread_Equal_True(ctx,
			database.Conversation_VendorPk(req.VendorPk))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &VendorConversationsUnreadResp{
		Conversations: ConversationsFromDB(conversations),
	}, nil
}

type PagedVendorConversationReq struct {
	VendorPk  int64
	PageToken string `json:"pageToken"`
}

type PagedVendorConversationResp struct {
	PageToken     string          `json:"pageToken"`
	Conversations []*Conversation `json:"conversations"`
}

func (u *VendorServer) GetPagedVendorConversations(ctx context.Context,
	req *PagedVendorConversationReq) (resp *PagedVendorConversationResp, err error) {

	var conversations []*database.Conversation
	var ctoken string
	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		conversations, ctoken, err = tx.Paged_Conversation_By_VendorPk(ctx,
			database.Conversation_VendorPk(req.VendorPk), conversationRequestLimit, req.PageToken)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &PagedVendorConversationResp{
		PageToken:     ctoken,
		Conversations: ConversationsFromDB(conversations),
	}, nil
}
