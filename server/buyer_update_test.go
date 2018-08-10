package server

import (
	"context"
	"ladybug/database"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuyerRequestFieldsFromUpdateRequest(t *testing.T) {
	//set up
	req := &UpdateBuyerRequest{
		CurrentEmail: "test@email.com",
		FirstName:    "Sarah",
		NewPassword:  "Password1!",
	}
	req_fields := BuyerRequestFieldsFromUpdateRequest(req)

	//assert first name
	require.True(t, req_fields.FirstName.set)
	require.Equal(t, *req_fields.FirstName.Value(), "Sarah")

	//assert current email
	require.True(t, req_fields.CurrentEmail.set)
	require.Equal(t, *req_fields.CurrentEmail.Value(), "test@email.com")

	//assert new password
	require.True(t, req_fields.NewPassword.set)
	require.Equal(t, *req_fields.NewPassword.Value(), "Password1!")

	//assert last name
	require.False(t, req_fields.LastName.set)
	require.Nil(t, req_fields.LastName.Value())

	//assert new email
	require.False(t, req_fields.NewEmail.set)
	require.Nil(t, req_fields.NewEmail.Value())

	//assert current password
	require.False(t, req_fields.CurrentPassword.set)
	require.Nil(t, req_fields.CurrentPassword.Value())
}

func TestValidateBuyerRequestFieldsEmptyRequest(t *testing.T) {
	//set up
	req := &UpdateBuyerRequest{}
	req_fields := BuyerRequestFieldsFromUpdateRequest(req)

	//empty request
	err := ValidateUpdateBuyerRequestFields(req_fields)
	require.EqualError(t, err, "all fields are empty. nothing to update")
}

func TestValidateBuyerRequestFieldsNameExceedsMax(t *testing.T) {
	//set up
	req := &UpdateBuyerRequest{}
	req_fields := BuyerRequestFieldsFromUpdateRequest(req)

	//first name too long
	req_fields.FirstName = SetRequestField("this_name_is_toooooooooooooooooooooooooooooooo_long")
	err := ValidateUpdateBuyerRequestFields(req_fields)
	require.EqualError(t, err, "name cannot exceed 50 characters")
}

func TestValidateBuyerRequestFieldsCurrentEmailNotSet(t *testing.T) {
	//set up
	req := &UpdateBuyerRequest{}
	req_fields := BuyerRequestFieldsFromUpdateRequest(req)

	//current email not set
	req_fields.LastName = SetRequestField("Carter")
	err := ValidateUpdateBuyerRequestFields(req_fields)
	require.EqualError(t, err, "current email must be set")
}

func TestValidateBuyerRequestFieldsCurrentPassNotSet(t *testing.T) {
	//set up
	req := &UpdateBuyerRequest{}
	req_fields := BuyerRequestFieldsFromUpdateRequest(req)

	//current password not set
	req_fields.CurrentEmail = SetRequestField("email@email.com")
	err := ValidateUpdateBuyerRequestFields(req_fields)
	require.EqualError(t, err, "current password must be set")
}

func TestValidateBuyerRequestFieldsInvalidPass(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	//set up
	ctx := context.Background()
	buyer := test.createFullTestBuyer(ctx)
	req := getFullUpdateBuyerRequest(buyer)
	req_fields := BuyerRequestFieldsFromUpdateRequest(req)

	//faulty new password
	req_fields.CurrentPassword = SetRequestField("Password2@")
	req_fields.NewPassword = SetRequestField("Password!")
	err := ValidateUpdateBuyerRequestFields(req_fields)
	require.EqualError(t, err, "Password must contain a number")
}

func TestValidateBuyerRequestFieldsSuccess(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	//set up
	ctx := context.Background()
	buyer := test.createFullTestBuyer(ctx)
	req := getFullUpdateBuyerRequest(buyer)
	req_fields := BuyerRequestFieldsFromUpdateRequest(req)

	//succesfully validated
	req_fields.NewPassword = SetRequestField("Password1!")
	err := ValidateUpdateBuyerRequestFields(req_fields)
	require.NoError(t, err)
}

func TestUpdateBuyerFirstNameOnly(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	//set up
	ctx := context.Background()
	buyer := test.createFullTestBuyer(ctx)
	req := getFullUpdateBuyerRequest(buyer)

	//update first name only
	req.FirstName = "Harpo"
	resp, err := test.BuyerServer.UpdateBuyer(ctx, req)
	require.NoError(t, err)
	require.Equal(t, req.FirstName, resp.Buyer.FirstName)
	require.NotEqual(t, req.FirstName, buyer.FirstName)
}

func TestUpdateBuyerLastNameOnly(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	//set up
	ctx := context.Background()
	buyer := test.createFullTestBuyer(ctx)
	req := getFullUpdateBuyerRequest(buyer)

	//update last name only
	req.LastName = "jojo"
	resp, err := test.BuyerServer.UpdateBuyer(ctx, req)
	require.NoError(t, err)
	require.Equal(t, req.LastName, resp.Buyer.LastName)
	require.NotEqual(t, req.LastName, buyer.LastName)
}

func TestUpdateBuyerFirstAndLastName(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	//set up
	ctx := context.Background()
	buyer := test.createFullTestBuyer(ctx)
	req := getFullUpdateBuyerRequest(buyer)

	//update first and last name together
	req.FirstName = "Osh Kosh"
	req.LastName = "Bgosh"
	resp, err := test.BuyerServer.UpdateBuyer(ctx, req)
	require.NoError(t, err)
	require.Equal(t, req.FirstName, resp.Buyer.FirstName)
	require.NotEqual(t, req.FirstName, buyer.FirstName)
	require.Equal(t, req.LastName, resp.Buyer.LastName)
	require.NotEqual(t, req.LastName, buyer.LastName)
}

func TestUpdateBuyerEmailOnly(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	//set up
	ctx := context.Background()
	buyer := test.createFullTestBuyer(ctx)
	req := getFullUpdateBuyerRequest(buyer)

	//update email
	req.NewEmail = "test@testemail.com"
	resp, err := test.BuyerServer.UpdateBuyer(ctx, req)
	require.NoError(t, err)
	require.Equal(t, req.NewEmail, resp.Buyer.Emails[0].Address)
	require.NotEqual(t, req.CurrentEmail, resp.Buyer.Emails[0].Address)
}

func TestUpdateBuyerEmailAndPassword(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	//set up
	ctx := context.Background()
	buyer := test.createFullTestBuyer(ctx)
	req := getFullUpdateBuyerRequest(buyer)

	//update email and password
	req.NewEmail = "newemail@email.com"
	req.NewPassword = "Password5%"
	resp, err := test.BuyerServer.UpdateBuyer(ctx, req)
	require.NoError(t, err)
	require.Equal(t, req.NewEmail, resp.Buyer.Emails[0].Address)
	require.NotEqual(t, req.CurrentEmail, resp.Buyer.Emails[0].Address)
}

func TestUpdateBuyerAllFields(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	//set up
	ctx := context.Background()
	buyer := test.createFullTestBuyer(ctx)
	req := getFullUpdateBuyerRequest(buyer)

	req.FirstName = "new first name"
	req.LastName = "new last name"
	req.NewEmail = "new@email.com"
	req.NewPassword = "Password1!"

	resp, err := test.BuyerServer.UpdateBuyer(ctx, req)
	require.NoError(t, err)

	//verify change in first name
	require.Equal(t, req.FirstName, resp.Buyer.FirstName)
	require.NotEqual(t, req.FirstName, buyer.FirstName)

	//verify change in last name
	require.Equal(t, req.LastName, resp.Buyer.LastName)
	require.NotEqual(t, req.LastName, buyer.LastName)

	//verify change in email
	require.Equal(t, req.NewEmail, resp.Buyer.Emails[0].Address)
	require.NotEqual(t, req.CurrentEmail, resp.Buyer.Emails[0].Address)

	//verify password change
	email, err := test.db.Get_BuyerEmail_By_Address(ctx, database.BuyerEmail_Address(
		resp.Buyer.Emails[0].Address))
	require.NoError(t, err)
	require.NoError(t, comparePasswordHash(req.NewPassword, email.SaltedHash))
}

//---------------------------------- helpers -----------------------------------------------//

func getFullUpdateBuyerRequest(b *FullTestBuyer) *UpdateBuyerRequest {
	return &UpdateBuyerRequest{
		BuyerPk:         b.Pk,
		FirstName:       b.FirstName,
		LastName:        b.LastName,
		CurrentEmail:    b.emails[0].Address,
		NewEmail:        "",
		CurrentPassword: defaultPassword,
		NewPassword:     "",
	}
}
