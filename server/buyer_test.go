package server

import (
	"context"
	"fmt"
	"testing"

	"ladybug/database"
	"ladybug/validate"

	"github.com/stretchr/testify/require"
)

func TestSuccesfulBuyerSignUp(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()
	ctx := context.Background()
	req := getCompleteSignUpRequest()

	resp, err := test.BuyerServer.SignUp(ctx, req)
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
	_, err := test.BuyerServer.SignUp(ctx, req)
	require.EqualError(t, err, "Password must not be empty")

	//no upper case letter
	req.Password = "no_upper_letter"
	_, err = test.BuyerServer.SignUp(ctx, req)
	require.EqualError(t, err, "Password must contain an upper case letter")

	//no lower case letter
	req.Password = "NO_LOWER_CASE_LETTER"
	_, err = test.BuyerServer.SignUp(ctx, req)
	require.EqualError(t, err, "Password must contain a lower case letter")

	//no number
	req.Password = "Password"
	_, err = test.BuyerServer.SignUp(ctx, req)
	require.EqualError(t, err, "Password must contain a number")

	//password longer than password max
	req.Password = "PASSWORD_is_longer_than_50_characters_and_therefore_will_not_work!"
	_, err = test.BuyerServer.SignUp(ctx, req)
	require.EqualError(t, err, fmt.Sprintf(
		"Password must be a maximum of %d characters", validate.MaxPasswordLen))

	//valid password
	req.Password = "Password8*"
	_, err = test.BuyerServer.SignUp(ctx, req)
	require.NoError(t, err)
}

func TestBuyerName(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	ctx := context.Background()
	req := getCompleteSignUpRequest()

	//missing First Name
	req.FirstName = ""
	_, err := test.BuyerServer.SignUp(ctx, req)
	require.EqualError(t, err, "name must not be empty")

	//missing Last Name
	req.FirstName = "Joey"
	req.LastName = ""
	_, err = test.BuyerServer.SignUp(ctx, req)
	require.EqualError(t, err, "name must not be empty")

	//name exceeds 50 characters
	req.LastName = "longer_than_50_characters_shouldn't_be_allowed_12345"
	_, err = test.BuyerServer.SignUp(ctx, req)
	require.EqualError(t, err, "name cannot exceed 50 characters")

	//name with no error
	req.LastName = "Calzone"
	_, err = test.BuyerServer.SignUp(ctx, req)
	require.NoError(t, err)
}

func TestBuyerEmail(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	ctx := context.Background()
	req := getCompleteSignUpRequest()

	//email empty
	req.Email = ""
	_, err := test.BuyerServer.SignUp(ctx, req)
	require.EqualError(t, err, "email address cannot be empty")

	//missing top level domain
	req.Email = "joey@calzone"
	_, err = test.BuyerServer.SignUp(ctx, req)
	require.EqualError(t, err, fmt.Sprintf("%s is not a valid email address", req.Email))

	//missing before @
	req.Email = "@calzone.com"
	_, err = test.BuyerServer.SignUp(ctx, req)
	require.EqualError(t, err, fmt.Sprintf("%s is not a valid email address", req.Email))

	//missing @
	req.Email = "joeycalzone.com"
	_, err = test.BuyerServer.SignUp(ctx, req)
	require.EqualError(t, err, fmt.Sprintf("%s is not a valid email address", req.Email))

	//TLD is not .com
	req.Email = "joey@calzone.marketing"
	_, err = test.BuyerServer.SignUp(ctx, req)
	require.NoError(t, err)

	//other valid email
	req.Email = "joey@calzone.com"
	_, err = test.BuyerServer.SignUp(ctx, req)
	require.NoError(t, err)
}

func TestBuyerAddress(t *testing.T) {
	test := newTest(t)
	defer test.tearDown()

	ctx := context.Background()
	req := getCompleteSignUpRequest()

	//missing street address
	req.BillingAddress = &validate.Address{}
	_, err := test.BuyerServer.SignUp(ctx, req)
	require.EqualError(t, err, "city, state, street, or zip fields are blank for billing address")

	//missing city
	req.BillingAddress.StreetAddress = "21 heartbreak ln"
	_, err = test.BuyerServer.SignUp(ctx, req)
	require.EqualError(t, err, "city, state, street, or zip fields are blank for billing address")

	//missing State
	req.BillingAddress.City = "Paris"
	_, err = test.BuyerServer.SignUp(ctx, req)
	require.EqualError(t, err, "city, state, street, or zip fields are blank for billing address")

	//missing Zip
	req.BillingAddress.State = "FL"
	_, err = test.BuyerServer.SignUp(ctx, req)
	require.EqualError(t, err, "city, state, street, or zip fields are blank for billing address")

	//valide address Zip
	req.BillingAddress.Zip = 98563
	_, err = test.BuyerServer.SignUp(ctx, req)
	require.NoError(t, err)

	//billing and shipping the same
	req.ShippingAddress = req.BillingAddress
	_, err = test.BuyerServer.SignUp(ctx, req)
	require.EqualError(t, err, "shipping address is the same as billing address")

	//address is nil
	req.BillingAddress = nil
	_, err = test.BuyerServer.SignUp(ctx, req)
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

//compareSignUpWithDatabase will compare a reqest with objects created in the database on a
//succesful request. The function assumes that the buyer is a new buyer with on previous account
//data
func (s *serverTest) compareSignUpWithDatabase(ctx context.Context, buyer_pk int64,
	req *SignUpRequest) {
	//buyer object was created
	buyer, err := s.db.Get_Buyer_By_Pk(ctx, database.Buyer_Pk(buyer_pk))
	require.NoError(s.t, err)
	require.Equal(s.t, buyer.FirstName, req.FirstName)
	require.Equal(s.t, buyer.LastName, req.LastName)

	//emails created
	emails, err := s.db.All_Email_By_BuyerPk(ctx, database.Email_BuyerPk(buyer_pk))
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

	if !validate.AddressIsEmpty(req.ShippingAddress) {
		shipping_adds, err := s.db.All_Address_By_IsBilling_Equal_False_And_BuyerPk(
			ctx, database.Address_BuyerPk(buyer_pk))
		require.NoError(s.t, err)
		require.Equal(s.t, len(shipping_adds), 1)
		require.Equal(s.t, shipping_adds[0].StreetAddress, req.ShippingAddress.StreetAddress)
	}
}

func getCompleteSignUpRequest() *SignUpRequest {
	return &SignUpRequest{
		FirstName: "Joey",
		LastName:  "Calzone",
		Password:  "Password8%",
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