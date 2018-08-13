package server

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

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
