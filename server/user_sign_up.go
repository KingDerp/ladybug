package server

import (
	"context"
	"ladybug/database"
	"strings"

	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/zeebo/errs"
)

type SignUpRequest struct {
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	Password       string `json:"password"`
	Email          string `json:"email"`
	BillingStreet  string `json:"billingStreet"`
	BillingCity    string `json:"billingCity"`
	BillingState   string `json:"billingState"`
	BillingZip     int    `json:"billingZip"`
	ShippingStreet string `json:"shippingStreet"`
	ShippingCity   string `json:"shippingCity"`
	ShippingState  string `json:"shippingState"`
	ShippingZip    int    `json:"shippingZip"`
}

type SignUpResponse struct {
	Session *database.Session
}

func (u *UserServer) SignUp(ctx context.Context, req *SignUpRequest) (resp *SignUpResponse,
	err error) {

	err = validateSignUpRequest(req)
	if err != nil {
		return nil, err
	}

	ship_addr_is_empty := shippingAddressIsEmpty(req)

	if !ship_addr_is_empty {
		if err := validateShippingAddress(req); err != nil {
			return nil, err
		}
	}

	hash, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	var session *database.Session
	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {

		user, err := tx.Create_User(ctx, database.User_Id(uuid.NewV4().String()),
			database.User_FirstName(req.FirstName), database.User_LastName(req.LastName))
		if err != nil {
			return err
		}

		err = tx.CreateNoReturn_Email(ctx, database.Email_UserPk(user.Pk),
			database.Email_Address(strings.ToLower(req.Email)), database.Email_SaltedHash(hash),
			database.Email_Id(uuid.NewV4().String()))
		if database.IsConstraintViolationError(err) {
			logrus.Error(err)
			return errs.New("that email is already in use")
		}
		if err != nil {
			return err
		}

		err = tx.CreateNoReturn_Address(ctx, database.Address_UserPk(user.Pk),
			database.Address_StreetAddress(req.BillingStreet),
			database.Address_City(req.BillingCity),
			database.Address_State(req.BillingState),
			database.Address_Zip(req.BillingZip),
			database.Address_IsBilling(true),
			database.Address_Id(uuid.NewV4().String()))

		if !ship_addr_is_empty {
			err = tx.CreateNoReturn_Address(ctx, database.Address_UserPk(user.Pk),
				database.Address_StreetAddress(req.ShippingStreet),
				database.Address_City(req.ShippingCity),
				database.Address_State(req.ShippingCity),
				database.Address_Zip(req.ShippingZip),
				database.Address_IsBilling(false),
				database.Address_Id(uuid.NewV4().String()))
		}

		session, err = tx.Create_Session(ctx, database.Session_UserPk(user.Pk),
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
