package server

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPostVendorMessagesToConversations(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	//set up
	ctx := context.Background()
	vendor := test.createVendorInDB(ctx)
	buyer := test.createBuyer(ctx, &createBuyerInDBOptions{})
	conversation := test.createConversationInDB(buyer, vendor)
	test.createMessageHistory(ctx, conversation, 210)

	req := &PostVendorMessageToConversationReq{
		VendorPk:           vendor.Pk,
		BuyerId:            buyer.Id,
		MessageDescription: "stop! can't touch this.",
	}

	resp, err := test.VendorServer.PostVendorMessageToConversation(ctx, req)
	require.NoError(t, err)
	require.Equal(t, resp.Message.Description, req.MessageDescription)
	require.False(t, resp.Message.BuyerSent)
}

func TestGetVendorMessagesByConversationId(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	//set up
	ctx := context.Background()
	vendor := test.createVendorInDB(ctx)
	buyer := test.createBuyer(ctx, &createBuyerInDBOptions{})
	conversation := test.createConversationInDB(buyer, vendor)
	test.createMessageHistory(ctx, conversation, 210)

	//offset 0
	req := &PagedVendorMessagesByConversationIdReq{
		Offset:         int64(0),
		ConversationId: conversation.Id,
	}
	resp, err := test.VendorServer.PagedVendorMessagesByConversationId(ctx, req)
	require.NoError(t, err)
	require.Equal(t, len(resp.Messages), messageRequestLimit)
	require.Equal(t, resp.Offset, int64(messageRequestLimit))

	//subsequent request with new offset
	req.Offset = resp.Offset
	resp, err = test.VendorServer.PagedVendorMessagesByConversationId(ctx, req)
	require.NoError(t, err)
	require.Equal(t, len(resp.Messages), messageRequestLimit)
	require.Equal(t, resp.Offset, int64(messageRequestLimit*2))

	//offset exceeds number of messages in history
	req.Offset = int64(1000)
	resp, err = test.VendorServer.PagedVendorMessagesByConversationId(ctx, req)
	require.NoError(t, err)
	require.Equal(t, len(resp.Messages), 0) //no messages because offest exceeds message history

	//offset result is less than messageRequestLimit
	req.Offset = 200
	resp, err = test.VendorServer.PagedVendorMessagesByConversationId(ctx, req)
	require.NoError(t, err)
	require.Equal(t, len(resp.Messages), 10)
}

func TestGetPagedVendorConversations(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	//set up
	ctx := context.Background()
	vendor := test.createVendorInDB(ctx)
	buyers := test.createDefaultBuyers(ctx, 53)
	test.createConversationsWithBuyers(vendor, buyers)

	//get first set of paged results
	req := &PagedVendorConversationReq{VendorPk: vendor.Pk}
	resp, err := test.VendorServer.GetPagedVendorConversations(ctx, req)
	require.NoError(t, err)
	require.Equal(t, resp.PageToken, strconv.Itoa(conversationRequestLimit))
	require.Equal(t, len(resp.Conversations), conversationRequestLimit)

	//get second set of paged results
	req.PageToken = resp.PageToken
	resp, err = test.VendorServer.GetPagedVendorConversations(ctx, req)
	require.NoError(t, err)
	require.Equal(t, resp.PageToken, strconv.Itoa(conversationRequestLimit*2))
	require.Equal(t, len(resp.Conversations), conversationRequestLimit)

	//third page should only include 3 results
	req.PageToken = resp.PageToken
	resp, err = test.VendorServer.GetPagedVendorConversations(ctx, req)
	require.NoError(t, err)
	require.Equal(t, len(resp.Conversations), 13)
}

func TestGetVendorConversationsUnread(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	//set up
	ctx := context.Background()
	vendor := test.createVendorInDB(ctx)
	buyers := test.createDefaultBuyers(ctx, 53)
	conversations := test.createConversationsWithBuyers(vendor, buyers)
	test.createDefaultMessagesFromBuyer(ctx, conversations[:20])

	//get unread conversations
	req := &VendorConversationsUnreadReq{VendorPk: vendor.Pk}
	resp, err := test.VendorServer.GetVendorCoversationsUnread(ctx, req)
	require.NoError(t, err)
	require.Equal(t, len(resp.Conversations), 20)
}
