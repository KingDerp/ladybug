package server

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"ladybug/database"
	"ladybug/validate"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

var (
	defaultPassword = "Password8%"
)

func TestProductReview(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	ctx := context.Background()
	vendor := test.createVendorInDB(ctx)
	buyer := test.createBuyer(ctx, &createBuyerInDBOptions{})
	product := test.createActiveAndApprovedProductInStock(ctx, vendor.Pk)

	req := &ProductReviewReq{
		BuyerPk:     buyer.Pk,
		ProductId:   product.Id,
		Stars:       1,
		Description: "product did not fill the gaping hole consumerism has left!",
	}

	//buyer has not purchased the product
	_, err := test.BuyerServer.ReviewProduct(ctx, req)
	require.EqualError(t, err, "you cannot leave a review for a product you have not purchased")

	//buyer has purchased
	test.purchaseProduct(ctx, buyer.Pk, vendor.Pk, product)
	resp, err := test.BuyerServer.ReviewProduct(ctx, req)
	require.NoError(t, err)
	require.Equal(t, resp.ReviewResponeMessage, "Thank you for leaving a review!")

	//buyer has already left a review
	_, err = test.BuyerServer.ReviewProduct(ctx, req)
	require.EqualError(t, err, "you have already left a review for this product")
}

func TestStartTrialProduct(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	ctx := context.Background()
	vendor := test.createVendorInDB(ctx)
	buyer := test.createBuyer(ctx, &createBuyerInDBOptions{})
	product := test.createActiveAndApprovedProductInStock(ctx, vendor.Pk)

	req := &StartProductTrialReq{
		BuyerPk:   buyer.Pk,
		VendorId:  vendor.Id,
		ProductId: product.Id,
	}

	resp, err := test.BuyerServer.StartProductTrial(ctx, req)
	require.NoError(t, err)
	require.Equal(t, resp.TrialProduct.TrialPrice, product.Price)
	require.Equal(t, resp.TrialProduct.TrialEndDate, trialExpirationInUnixTime(product.CreatedAt))
}

func TestIncrementConversationCount(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	//set up
	ctx := context.Background()
	vendor := test.createVendorInDB(ctx)
	buyer := test.createBuyer(ctx, &createBuyerInDBOptions{})

	//first contact from buyer
	req := &PostBuyerMessageToConversationReq{
		BuyerPk:            buyer.Pk,
		VendorId:           vendor.Id,
		MessageDescription: "Your Mother was a hamster",
	}
	resp, err := test.BuyerServer.PostBuyerMessageToConversation(ctx, req)
	require.NoError(t, err)
	require.Equal(t, resp.Message.Description, req.MessageDescription)
	require.True(t, resp.Message.BuyerSent)

	conversation := test.getConversation(ctx, vendor.Pk, buyer.Pk)
	require.Equal(t, conversation.MessageCount, int64(1))

	resp, err = test.BuyerServer.PostBuyerMessageToConversation(ctx, req)
	require.NoError(t, err)

	conversation = test.getConversation(ctx, vendor.Pk, buyer.Pk)
	require.Equal(t, conversation.MessageCount, int64(2))
	require.Equal(t, resp.Message.MessageNumber, int64(2))
}

func TestGetBuyerMessagesByConversationId(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	//set up
	ctx := context.Background()
	vendor := test.createVendorInDB(ctx)
	buyer := test.createBuyer(ctx, &createBuyerInDBOptions{})
	conversation := test.createConversationInDB(buyer, vendor)
	test.createMessageHistory(ctx, conversation, 210)

	//offset 0
	req := &PagedBuyerMessagesByConversationIdReq{
		Offset:         int64(0),
		ConversationId: conversation.Id,
	}
	resp, err := test.BuyerServer.PagedBuyerMessagesByConversationId(ctx, req)
	require.NoError(t, err)
	require.Equal(t, len(resp.Messages), messageRequestLimit)
	require.Equal(t, resp.Offset, int64(messageRequestLimit))

	//subsequent request with new offset
	req.Offset = resp.Offset
	resp, err = test.BuyerServer.PagedBuyerMessagesByConversationId(ctx, req)
	require.NoError(t, err)
	require.Equal(t, len(resp.Messages), messageRequestLimit)
	require.Equal(t, resp.Offset, int64(messageRequestLimit*2))

	//offset exceeds number of messages in history
	req.Offset = int64(1000)
	resp, err = test.BuyerServer.PagedBuyerMessagesByConversationId(ctx, req)
	require.NoError(t, err)
	require.Equal(t, len(resp.Messages), 0) //no messages because offest exceeds message history

	//offset result is less than messageRequestLimit
	req.Offset = 200
	resp, err = test.BuyerServer.PagedBuyerMessagesByConversationId(ctx, req)
	require.NoError(t, err)
	require.Equal(t, len(resp.Messages), 10)
}

func TestGetBuyerConversationsUnread(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	//set up
	ctx := context.Background()
	vendors := test.createVendorsInDB(ctx, 50)
	buyer := test.createBuyer(ctx, &createBuyerInDBOptions{})
	conversations := test.createConversationsWithVendors(buyer, vendors)
	test.createDefaultMessagesFromVendor(ctx, conversations[:20])

	//get unread conversations
	req := &BuyerConversationsUnreadReq{BuyerPk: buyer.Pk}
	resp, err := test.BuyerServer.GetBuyerConversationsUnread(ctx, req)
	require.NoError(t, err)
	require.Equal(t, len(resp.Conversations), 20)
}

func TestGetPagedBuyerConversations(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	//set up
	ctx := context.Background()
	vendors := test.createVendorsInDB(ctx, 50)
	buyer := test.createBuyer(ctx, &createBuyerInDBOptions{})
	test.createConversationsWithVendors(buyer, vendors)

	//get first set of paged results
	req := &PagedBuyerConversationsReq{BuyerPk: buyer.Pk}
	resp, err := test.BuyerServer.GetPagedBuyerConversations(ctx, req)
	require.NoError(t, err)
	require.Equal(t, resp.PageToken, strconv.Itoa(conversationRequestLimit))

	//get second set of paged results
	req.PageToken = resp.PageToken
	resp, err = test.BuyerServer.GetPagedBuyerConversations(ctx, req)
	require.NoError(t, err)
	require.Equal(t, resp.PageToken, strconv.Itoa(conversationRequestLimit*2))
}

func TestGetBuyerProducts(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	//set up
	ctx := context.Background()
	vendor := test.createVendorInDB(ctx)
	test.createActiveAndApprovedProductsInStock(ctx, 100, vendor.Pk)

	//productRequestLimit test
	resp, err := test.BuyerServer.BuyerProducts(ctx, &ProductRequest{})
	require.NoError(t, err)
	require.Equal(t, len(resp.Products), productRequestLimit)

	//token number test
	token_num, err := strconv.Atoi(resp.PageToken)
	require.NoError(t, err)
	require.Equal(t, token_num, productRequestLimit)
}

func TestGetBuyerProductsActiveOnly(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	//set up
	ctx := context.Background()
	vendor := test.createVendorInDB(ctx)
	test.createActiveAndApprovedProductsInStock(ctx, 100, vendor.Pk)
	test.createInactiveAndApprovedProductsInStock(ctx, 50, vendor.Pk)

	//only active products in response
	resp, err := test.BuyerServer.BuyerProducts(ctx, &ProductRequest{})
	require.NoError(t, err)
	require.True(t, test.allProductsAreActive(resp.Products))

	//inactive products match count
	count, err := test.db.Count_Product_By_ProductActive_Equal_False(ctx)
	require.NoError(t, err)
	require.Equal(t, count, int64(50))
}

func TestGetBuyerProductsNotApproved(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	//set up
	ctx := context.Background()
	vendor := test.createVendorInDB(ctx)
	test.createActiveProductsNotApprovedInStock(ctx, 50, vendor.Pk)

	//no products returned
	resp, err := test.BuyerServer.BuyerProducts(ctx, &ProductRequest{})
	require.NoError(t, err)
	require.Equal(t, len(resp.Products), 0)
}

func TestGetBuyerInfo(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	ctx := context.Background()

	//buyer doesn't exist
	b, err := test.db.Find_Buyer_By_Pk(ctx, database.Buyer_Pk(100))
	require.NoError(t, err)
	require.Nil(t, b)

	buyer := test.createFullTestBuyer(ctx)
	req := &GetBuyerRequest{BuyerPk: buyer.Pk}

	//check response matches request
	resp, err := test.BuyerServer.GetBuyer(ctx, req)
	require.NoError(t, err)
	require.Equal(t, resp.Buyer.FirstName, buyer.FirstName)
	require.Equal(t, resp.Buyer.LastName, buyer.LastName)
	require.Equal(t, resp.Buyer.Emails[0].Address, buyer.emails[0].Address)
}

func TestBuyerLogIn(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	ctx := context.Background()
	req := &LogInRequest{Email: "non_existant@email.com", Password: "Password6*"}

	//no email exists
	resp, err := test.BuyerServer.BuyerLogIn(ctx, req)
	require.EqualError(t, err, "No email exists with that address")
	require.Nil(t, resp)

	buyer := test.createFullTestBuyer(ctx)

	//password mismatch
	req.Email = buyer.emails[0].Address
	resp, err = test.BuyerServer.BuyerLogIn(ctx, req)
	require.EqualError(t, err, "email or password does not match")
	require.Nil(t, resp)

	//valid request
	req.Password = buyer.emails[0].unsaltedPassword
	resp, err = test.BuyerServer.BuyerLogIn(ctx, req)
	require.NoError(t, err)
}

func TestSuccesfulBuyerSignUp(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()
	ctx := context.Background()
	req := getCompleteSignUpRequest()

	resp, err := test.BuyerServer.BuyerSignUp(ctx, req)
	require.NoError(t, err)

	test.compareSignUpWithDatabase(ctx, resp.Session.BuyerPk, req)
}

func TestBuyerPasswordSignUp(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	ctx := context.Background()
	req := getCompleteSignUpRequest()

	//password is empty
	req.Password = ""
	_, err := test.BuyerServer.BuyerSignUp(ctx, req)
	require.EqualError(t, err, "Password must not be empty")

	//no upper case letter
	req.Password = "no_upper_letter"
	_, err = test.BuyerServer.BuyerSignUp(ctx, req)
	require.EqualError(t, err, "Password must contain an upper case letter")

	//no lower case letter
	req.Password = "NO_LOWER_CASE_LETTER"
	_, err = test.BuyerServer.BuyerSignUp(ctx, req)
	require.EqualError(t, err, "Password must contain a lower case letter")

	//no number
	req.Password = "Password"
	_, err = test.BuyerServer.BuyerSignUp(ctx, req)
	require.EqualError(t, err, "Password must contain a number")

	//password longer than password max
	req.Password = "PASSWORD_is_longer_than_50_characters_and_therefore_will_not_work!"
	_, err = test.BuyerServer.BuyerSignUp(ctx, req)
	require.EqualError(t, err, fmt.Sprintf(
		"Password must be a maximum of %d characters", validate.MaxPasswordLen))

	//valid password
	req.Password = "Password8*"
	_, err = test.BuyerServer.BuyerSignUp(ctx, req)
	require.NoError(t, err)
}

func TestBuyerName(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	ctx := context.Background()
	req := getCompleteSignUpRequest()

	//missing First Name
	req.FirstName = ""
	_, err := test.BuyerServer.BuyerSignUp(ctx, req)
	require.EqualError(t, err, "name must not be empty")

	//missing Last Name
	req.FirstName = "Joey"
	req.LastName = ""
	_, err = test.BuyerServer.BuyerSignUp(ctx, req)
	require.EqualError(t, err, "name must not be empty")

	//name exceeds 50 characters
	req.LastName = "longer_than_50_characters_shouldn't_be_allowed_12345"
	_, err = test.BuyerServer.BuyerSignUp(ctx, req)
	require.EqualError(t, err, "name cannot exceed 50 characters")

	//name with no error
	req.LastName = "Calzone"
	_, err = test.BuyerServer.BuyerSignUp(ctx, req)
	require.NoError(t, err)
}

func TestBuyerEmail(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	ctx := context.Background()
	req := getCompleteSignUpRequest()

	//email empty
	req.Email = ""
	_, err := test.BuyerServer.BuyerSignUp(ctx, req)
	require.EqualError(t, err, "email address cannot be empty")

	//missing top level domain
	req.Email = "joey@calzone"
	_, err = test.BuyerServer.BuyerSignUp(ctx, req)
	require.EqualError(t, err, fmt.Sprintf("%s is not a valid email address", req.Email))

	//missing before @
	req.Email = "@calzone.com"
	_, err = test.BuyerServer.BuyerSignUp(ctx, req)
	require.EqualError(t, err, fmt.Sprintf("%s is not a valid email address", req.Email))

	//missing @
	req.Email = "joeycalzone.com"
	_, err = test.BuyerServer.BuyerSignUp(ctx, req)
	require.EqualError(t, err, fmt.Sprintf("%s is not a valid email address", req.Email))

	//TLD is not .com
	req.Email = "joey@calzone.marketing"
	_, err = test.BuyerServer.BuyerSignUp(ctx, req)
	require.NoError(t, err)

	//other valid email
	req.Email = "joey@calzone.com"
	_, err = test.BuyerServer.BuyerSignUp(ctx, req)
	require.NoError(t, err)
}

func TestBuyerAddress(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	ctx := context.Background()
	req := getCompleteSignUpRequest()

	//missing street address
	req.BillingAddress = &validate.Address{}
	_, err := test.BuyerServer.BuyerSignUp(ctx, req)
	require.EqualError(t, err, "city, state, street, or zip fields are blank for billing address")

	//missing city
	req.BillingAddress.StreetAddress = "21 heartbreak ln"
	_, err = test.BuyerServer.BuyerSignUp(ctx, req)
	require.EqualError(t, err, "city, state, street, or zip fields are blank for billing address")

	//missing State
	req.BillingAddress.City = "Paris"
	_, err = test.BuyerServer.BuyerSignUp(ctx, req)
	require.EqualError(t, err, "city, state, street, or zip fields are blank for billing address")

	//missing Zip
	req.BillingAddress.State = "FL"
	_, err = test.BuyerServer.BuyerSignUp(ctx, req)
	require.EqualError(t, err, "city, state, street, or zip fields are blank for billing address")

	//valide address Zip
	req.BillingAddress.Zip = 98563
	_, err = test.BuyerServer.BuyerSignUp(ctx, req)
	require.NoError(t, err)

	//billing and shipping the same
	req.ShippingAddress = req.BillingAddress
	_, err = test.BuyerServer.BuyerSignUp(ctx, req)
	require.EqualError(t, err, "shipping address is the same as billing address")

	//address is nil
	req.BillingAddress = nil
	_, err = test.BuyerServer.BuyerSignUp(ctx, req)
	require.EqualError(t, err, "no address was submitted")
}

func TestPasswordMatches(t *testing.T) {
	password := "Password1!"
	hash, err := hashPassword(password)
	require.NoError(t, err)

	err = comparePasswordHash(password, hash)
	require.NoError(t, err)

}

func TestPasswordNotMatch(t *testing.T) {
	password := "Password2!"
	hash, err := hashPassword("password3@")
	require.NoError(t, err)

	err = comparePasswordHash(password, hash)
	require.EqualError(t, err, "email or password does not match")
}

//---------------------------------- helpers -----------------------------------------------//

type FullTestBuyer struct {
	*database.Buyer
	emails    []*BuyerTestEmail
	addresses []*database.Address
	session   *database.BuyerSession
}

type BuyerTestEmail struct {
	*database.BuyerEmail
	unsaltedPassword string
}

func (s *serverTest) createFullTestBuyer(ctx context.Context) *FullTestBuyer {
	req := getCompleteSignUpRequest()

	resp, err := s.BuyerServer.BuyerSignUp(ctx, req)
	require.NoError(s.t, err)

	return s.compareSignUpWithDatabase(ctx, resp.Session.BuyerPk, req)
}

//compareSignUpWithDatabase will compare a reqest with objects created in the database on a
//succesful request. returns a pointer to FullTestBuyer which includes all relevant data pertaining
//to the new buyer. The function assumes that the buyer is a new buyer with on previous account
//data
func (s *serverTest) compareSignUpWithDatabase(ctx context.Context, buyer_pk int64,
	req *SignUpRequest) *FullTestBuyer {
	//buyer object was created
	buyer, err := s.db.Get_Buyer_By_Pk(ctx, database.Buyer_Pk(buyer_pk))
	require.NoError(s.t, err)
	require.Equal(s.t, buyer.FirstName, req.FirstName)
	require.Equal(s.t, buyer.LastName, req.LastName)

	//emails created
	emails, err := s.db.All_BuyerEmail_By_BuyerPk(ctx, database.BuyerEmail_BuyerPk(buyer_pk))
	require.NoError(s.t, err)
	require.Equal(s.t, len(emails), 1)
	require.Equal(s.t, emails[0].Address, req.Email)

	//password matches
	require.NoError(s.t, comparePasswordHash(req.Password, emails[0].SaltedHash))

	//Addresses
	billing_adds, err := s.db.All_Address_By_IsBilling_Equal_True_And_BuyerPk(
		ctx, database.Address_BuyerPk(buyer_pk))
	require.NoError(s.t, err)
	require.Equal(s.t, len(billing_adds), 1)
	require.Equal(s.t, billing_adds[0].StreetAddress, req.BillingAddress.StreetAddress)

	var shipping_adds []*database.Address
	if !validate.AddressIsEmpty(req.ShippingAddress) {
		shipping_adds, err = s.db.All_Address_By_IsBilling_Equal_False_And_BuyerPk(
			ctx, database.Address_BuyerPk(buyer_pk))
		require.NoError(s.t, err)
		require.Equal(s.t, len(shipping_adds), 1)
		require.Equal(s.t, shipping_adds[0].StreetAddress, req.ShippingAddress.StreetAddress)
	}

	session, err := s.db.Get_BuyerSession_By_BuyerPk(ctx, database.BuyerSession_BuyerPk(buyer.Pk))
	require.NoError(s.t, err)

	test_emails := []*BuyerTestEmail{}
	test_emails = append(test_emails, &BuyerTestEmail{
		BuyerEmail:       emails[0],
		unsaltedPassword: req.Password,
	})

	return &FullTestBuyer{
		Buyer:     buyer,
		emails:    test_emails,
		addresses: append(billing_adds, shipping_adds...),
		session:   session,
	}
}

//getCompleteSignUpRequest will return a pointer to a SignUpRequest that is intended to be complete.
//Meaning is a request to sign up a new buyer is made with the returned SignUpRequest there should
//be no errors with the request
func getCompleteSignUpRequest() *SignUpRequest {
	return &SignUpRequest{
		FirstName: "Joey",
		LastName:  "Calzone",
		Password:  defaultPassword,
		Email:     "joey@calzone.com",
		BillingAddress: &validate.Address{
			StreetAddress: "21 heartbreak ln",
			City:          "Paris",
			State:         "Florida",
			Zip:           87569,
		},
		ShippingAddress: &validate.Address{
			StreetAddress: "P.O. Box 32",
			City:          "Paris",
			State:         "Florida",
			Zip:           87569,
		},
	}
}

func randFloat(min, max float32) float32 {
	return min + rand.Float32()*(max-min)
}

func (s *serverTest) getConversation(ctx context.Context, vendor_pk, buyer_pk int64) (
	conversation *database.Conversation) {
	conversation, err := s.db.Get_Conversation_By_VendorPk_And_BuyerPk(ctx,
		database.Conversation_VendorPk(vendor_pk),
		database.Conversation_BuyerPk(buyer_pk))
	require.NoError(s.t, err)

	return conversation
}

func (s *serverTest) createConversationInDB(buyer *database.Buyer, vendor *database.Vendor) *database.Conversation {
	c, err := s.db.Create_Conversation(context.Background(),
		database.Conversation_VendorPk(vendor.Pk),
		database.Conversation_BuyerPk(buyer.Pk),
		database.Conversation_BuyerUnread(false),
		database.Conversation_VendorUnread(false),
		database.Conversation_MessageCount(0),
		database.Conversation_Id(uuid.NewV4().String()),
	)
	require.NoError(s.t, err)

	return c
}

func (s *serverTest) createConversationsWithBuyers(vendor *database.Vendor, buyers []*database.Buyer) (
	conversations []*database.Conversation) {

	conversations = []*database.Conversation{}
	for _, b := range buyers {
		conversations = append(conversations, s.createConversationInDB(b, vendor))
	}

	return conversations
}

func (s *serverTest) createConversationsWithVendors(buyer *database.Buyer, vendors []*database.Vendor) (
	conversations []*database.Conversation) {

	conversations = []*database.Conversation{}
	for _, v := range vendors {
		conversations = append(conversations, s.createConversationInDB(buyer, v))
	}

	return conversations
}

type newMessageOptions struct {
	Id          string
	BuyerSent   bool
	Description string
}

//createMessageHistory takes context, conversation model and a number. The number inicates the
//number of messages that should be created between the buyer and vendor. The message history will
//be back and forth with the buyer initiating the contact
func (s *serverTest) createMessageHistory(ctx context.Context, conversation *database.Conversation,
	num int) {
	for i := 0; i < num; i += 2 {
		s.createDefaultBuyerMessageToVendor(ctx, conversation)
		s.createDefaultVendorMessageToBuyer(ctx, conversation)
	}
}

func (s *serverTest) createDefaultVendorMessageToBuyer(ctx context.Context,
	conversation *database.Conversation) *database.Message {

	return s.createNewMessage(ctx, conversation, nil)
}

func (s *serverTest) createDefaultBuyerMessageToVendor(ctx context.Context,
	conversation *database.Conversation) *database.Message {

	return s.createNewMessage(ctx, conversation, &newMessageOptions{BuyerSent: true})
}

func (s *serverTest) createNewMessage(ctx context.Context, conversation *database.Conversation,
	options *newMessageOptions) (
	message *database.Message) {

	require.NotNil(s.t, conversation)

	if options == nil {
		options = &newMessageOptions{}
	}

	if options.Id == "" {
		options.Id = uuid.NewV4().String()
	}

	if options.Description == "" {
		options.Description = "default message body"
	}

	message, err := s.db.Create_Message(ctx,
		database.Message_Id(uuid.NewV4().String()),
		database.Message_BuyerSent(options.BuyerSent),
		database.Message_Description(options.Description),
		database.Message_ConversationPk(conversation.Pk),
		database.Message_ConversationNumber(conversation.MessageCount+1),
	)
	require.NoError(s.t, err)

	updates := database.Conversation_Update_Fields{
		MessageCount: database.Conversation_MessageCount(conversation.MessageCount + 1),
	}
	if options.BuyerSent == false {
		updates.BuyerUnread = database.Conversation_BuyerUnread(true)
	} else {
		updates.VendorUnread = database.Conversation_VendorUnread(true)
	}

	err = s.db.UpdateNoReturn_Conversation_By_Pk(ctx,
		database.Conversation_Pk(conversation.Pk), updates)
	require.NoError(s.t, err)

	return message
}

func (s *serverTest) createDefaultMessagesFromBuyer(ctx context.Context,
	conversations []*database.Conversation) {

	messages := []*database.Message{}
	for _, c := range conversations {
		messages = append(messages, s.createNewMessage(ctx, c, &newMessageOptions{BuyerSent: true}))
	}
}

func (s *serverTest) createDefaultMessagesFromVendor(ctx context.Context,
	conversations []*database.Conversation) {

	messages := []*database.Message{}
	for _, c := range conversations {
		messages = append(messages, s.createNewMessage(ctx, c, nil))
	}
}
