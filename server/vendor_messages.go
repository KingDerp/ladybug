package server

import (
	"context"
	"ladybug/database"

	uuid "github.com/satori/go.uuid"
)

type PostVendorMessageToConversationReq struct {
	VendorPk           int64
	BuyerId            string `json:"vendorId"`
	MessageDescription string `json:"messageDescription"`
}

type PostVendorMessageToConversationResp struct {
	Message *Message `json:"messages"`
}

func (v *VendorServer) PostVendorMessageToConversation(ctx context.Context,
	req *PostVendorMessageToConversationReq) (resp *PostVendorMessageToConversationResp, err error) {

	var message *database.Message
	err = v.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {

		buyer_pk_field, err := tx.Get_Buyer_Pk_By_Id(ctx,
			database.Buyer_Id(req.BuyerId))
		if err != nil {
			return err
		}

		conversation, err := tx.Find_Conversation_By_VendorPk_And_BuyerPk(ctx,
			database.Conversation_VendorPk(req.VendorPk),
			database.Conversation_BuyerPk(buyer_pk_field.Pk))
		if err != nil {
			return err
		}

		//if buyer and vendor have not had a conversation before create new conversation
		if conversation == nil {
			conversation, err = tx.Create_Conversation(ctx,
				database.Conversation_VendorPk(req.VendorPk),
				database.Conversation_BuyerPk(buyer_pk_field.Pk),
				database.Conversation_BuyerUnread(true),
				database.Conversation_VendorUnread(false),
				database.Conversation_MessageCount(1),
				database.Conversation_Id(uuid.NewV4().String()))
			if err != nil {
				return err
			}
		} else {
			conversation_updates := database.Conversation_Update_Fields{
				MessageCount: database.Conversation_MessageCount(conversation.MessageCount + 1),
				BuyerUnread:  database.Conversation_BuyerUnread(true),
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
			database.Message_BuyerSent(false),
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

	return &PostVendorMessageToConversationResp{
		Message: MessageFromDB(message),
	}, nil
}

type PagedVendorMessagesByConversationIdReq struct {
	Offset         int64  `json:"offset"`
	ConversationId string `json:"conversationId"`
}

type PagedVendorMessagesByConversationIdResp struct {
	Offset   int64      `json:"offset"`
	Messages []*Message `json:"messages"`
}

func (v *VendorServer) PagedVendorMessagesByConversationId(ctx context.Context,
	req *PagedVendorMessagesByConversationIdReq) (resp *PagedVendorMessagesByConversationIdResp, err error) {

	var messages []*database.Message
	err = v.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
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

	return &PagedVendorMessagesByConversationIdResp{
		Messages: MessagesFromDB(messages),
		Offset:   offset,
	}, nil
}
