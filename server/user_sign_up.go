package server

import (
	"context"
	"strings"

	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/zeebo/errs"

	"ladybug/database"
)

type SignUpRequest struct {
	FirstName       string   `json:"firstName"`
	LastName        string   `json:"lastName"`
	Password        string   `json:"password"`
	Email           string   `json:"email"`
	BillingAddress  *Address `json:"billingAddress"`
	ShippingAddress *Address `json"shippingAddress"`
}

type SignUpResponse struct {
	Session *database.Session
}

func (u *BuyerServer) SignUp(ctx context.Context, req *SignUpRequest) (resp *SignUpResponse,
	err error) {

	err = validateSignUpRequest(req)
	if err != nil {
		return nil, err
	}

	hash, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	var session *database.Session
	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {

		buyer, err := tx.Create_Buyer(ctx, database.Buyer_Id(uuid.NewV4().String()),
			database.Buyer_FirstName(req.FirstName), database.Buyer_LastName(req.LastName))
		if err != nil {
			return err
		}

		err = tx.CreateNoReturn_Email(ctx, database.Email_BuyerPk(buyer.Pk),
			database.Email_Address(strings.ToLower(req.Email)), database.Email_SaltedHash(hash),
			database.Email_Id(uuid.NewV4().String()))
		if database.IsConstraintViolationError(err) {
			logrus.Error(err)
			return errs.New("that email is already in use")
		}
		if err != nil {
			return err
		}

		err = tx.CreateNoReturn_Address(ctx, database.Address_BuyerPk(buyer.Pk),
			database.Address_StreetAddress(req.BillingAddress.StreetAddress),
			database.Address_City(req.BillingAddress.City),
			database.Address_State(req.BillingAddress.State),
			database.Address_Zip(req.BillingAddress.Zip),
			database.Address_IsBilling(true),
			database.Address_Id(uuid.NewV4().String()))

		if !addressIsEmpty(req.ShippingAddress) {
			err = tx.CreateNoReturn_Address(ctx, database.Address_BuyerPk(buyer.Pk),
				database.Address_StreetAddress(req.ShippingAddress.StreetAddress),
				database.Address_City(req.ShippingAddress.City),
				database.Address_State(req.ShippingAddress.State),
				database.Address_Zip(req.ShippingAddress.Zip),
				database.Address_IsBilling(false),
				database.Address_Id(uuid.NewV4().String()))
		}

		session, err = tx.Create_Session(ctx, database.Session_BuyerPk(buyer.Pk),
			database.Session_Id(uuid.NewV4().String()))
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
