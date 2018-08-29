package server

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/zeebo/errs"
	"gopkg.in/spacemonkeygo/dbx.v1/prettyprint"

	"ladybug/database"
	"ladybug/validate"
)

type SignUpRequest struct {
	FirstName       string            `json:"firstName"`
	LastName        string            `json:"lastName"`
	Password        string            `json:"password"`
	Email           string            `json:"email"`
	BillingAddress  *validate.Address `json:"billingAddress"`
	ShippingAddress *validate.Address `json"shippingAddress"`
}

type SignUpResponse struct {
	Session *database.BuyerSession
}

func (u *BuyerServer) BuyerSignUp(ctx context.Context, req *SignUpRequest) (resp *SignUpResponse,
	err error) {

	fmt.Println("entered sign up server")
	prettyprint.Println(req)
	err = ValidateBuyerSignUpRequest(req)
	if err != nil {
		return nil, err
	}

	hash, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	var session *database.BuyerSession
	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {

		buyer, err := tx.Create_Buyer(ctx, database.Buyer_Id(uuid.NewV4().String()),
			database.Buyer_FirstName(req.FirstName), database.Buyer_LastName(req.LastName))
		if err != nil {
			fmt.Println("error is here")
			return err
		}

		fmt.Println("BLAH")
		err = tx.CreateNoReturn_BuyerEmail(ctx,
			database.BuyerEmail_BuyerPk(buyer.Pk),
			database.BuyerEmail_Address(strings.ToLower(req.Email)),
			database.BuyerEmail_SaltedHash(hash),
			database.BuyerEmail_Id(uuid.NewV4().String()),
		)
		if database.IsConstraintViolationError(err) {
			logrus.Error(err)
			return errs.New("that email is already in use")
		}
		if err != nil {
			return err
		}

		fmt.Println("BLAH1")
		err = tx.CreateNoReturn_Address(ctx, database.Address_BuyerPk(buyer.Pk),
			database.Address_StreetAddress(req.BillingAddress.StreetAddress),
			database.Address_City(req.BillingAddress.City),
			database.Address_State(req.BillingAddress.State),
			database.Address_Zip(req.BillingAddress.Zip),
			database.Address_IsBilling(true),
			database.Address_Id(uuid.NewV4().String()))
		if err != nil {
			return err
		}
		fmt.Println("BLAH2")

		if !validate.AddressIsEmpty(req.ShippingAddress) {
			err = tx.CreateNoReturn_Address(ctx, database.Address_BuyerPk(buyer.Pk),
				database.Address_StreetAddress(req.ShippingAddress.StreetAddress),
				database.Address_City(req.ShippingAddress.City),
				database.Address_State(req.ShippingAddress.State),
				database.Address_Zip(req.ShippingAddress.Zip),
				database.Address_IsBilling(false),
				database.Address_Id(uuid.NewV4().String()))
		}
		fmt.Println("BLAH3")

		session, err = tx.Create_BuyerSession(ctx, database.BuyerSession_BuyerPk(buyer.Pk),
			database.BuyerSession_Id(uuid.NewV4().String()))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &SignUpResponse{Session: session}, nil
}

//ValidateBuyerSignUpRequest is an internal function that is used to verify only the required data in a sign up
//request. For example a billing address is required a shipping address is not.
func ValidateBuyerSignUpRequest(sur *SignUpRequest) error {

	//validate buyer name
	if err := validate.CheckFullName(sur.FirstName, sur.LastName); err != nil {
		return err
	}

	//validate password
	if err := validate.CheckPassword(sur.Password); err != nil {
		return err
	}

	if err := validate.CheckEmail(sur.Email); err != nil {
		return err
	}

	if err := validate.CheckAddress(sur.BillingAddress); err != nil {
		return err
	}

	if !validate.AddressIsEmpty(sur.ShippingAddress) {
		if reflect.DeepEqual(sur.ShippingAddress, sur.BillingAddress) {
			return errs.New("shipping address is the same as billing address")
		}

		if err := validate.CheckAddress(sur.ShippingAddress); err != nil {
			return err
		}
	}

	return nil
}
