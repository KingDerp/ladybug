package server

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/zeebo/errs"

	"ladybug/database"
	"ladybug/validate"
)

const (
	maxExecutiveContacts = 12
)

//NOTE: Executive contacts have full access to a vendor's portal. there are presently no roles set up
type ExecutiveContact struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	CountryCode int    `json:"countryCode"`
	AreaCode    int    `json:"areaCode"`
	PhoneNumber int    `json:"phoneNumber"`
	Email       string `json:"email"`
	Password    string `json:"password"`
}

type VendorSignUpRequest struct {
	Fein              string              `json:"fein"`
	BillingAddress    *validate.Address   `json:"billingAddress"`
	ShippingAddress   *validate.Address   `json"shippingAddress"`
	ExecutiveContacts []*ExecutiveContact `json"executiveContact"`
}

type VendorSignUpResponse struct {
	Session  *database.VendorSession `json:"-"`
	VendorId string                  `json:"vendorId"`
}

func (v *VendorServer) VendorSignUp(ctx context.Context, req *VendorSignUpRequest) (
	resp *VendorSignUpResponse, err error) {

	err = ValidateVendorSignUpRequest(req)
	if err != nil {
		return nil, err
	}

	ship_addr_is_empty := validate.AddressIsEmpty(req.ShippingAddress)

	if !ship_addr_is_empty {
		if err := validate.CheckAddress(req.ShippingAddress); err != nil {
			return nil, err
		}
	}

	var vendor_session *database.VendorSession
	var vendor_id string
	err = v.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {

		vendor, err := tx.Create_Vendor(ctx,
			database.Vendor_Id(uuid.NewV4().String()),
			database.Vendor_Fein(req.Fein))
		if err != nil {
			return err
		}

		vendor_id = vendor.Id

		err = tx.CreateNoReturn_VendorAddress(ctx,
			database.VendorAddress_VendorPk(vendor.Pk),
			database.VendorAddress_StreetAddress(req.BillingAddress.StreetAddress),
			database.VendorAddress_City(req.BillingAddress.City),
			database.VendorAddress_State(req.BillingAddress.State),
			database.VendorAddress_Zip(req.BillingAddress.Zip),
			database.VendorAddress_IsBilling(true),
			database.VendorAddress_Id(uuid.NewV4().String()))
		if err != nil {
			return err
		}

		if !ship_addr_is_empty {
			err = tx.CreateNoReturn_VendorAddress(ctx,
				database.VendorAddress_VendorPk(vendor.Pk),
				database.VendorAddress_StreetAddress(req.ShippingAddress.StreetAddress),
				database.VendorAddress_City(req.ShippingAddress.City),
				database.VendorAddress_State(req.ShippingAddress.State),
				database.VendorAddress_Zip(req.ShippingAddress.Zip),
				database.VendorAddress_IsBilling(false),
				database.VendorAddress_Id(uuid.NewV4().String()))
			if err != nil {
				return err
			}
		}

		for _, e := range req.ExecutiveContacts {
			hash, err := hashPassword(e.Password)
			if err != nil {
				return err
			}

			exec, err := tx.Create_ExecutiveContact(ctx,
				database.ExecutiveContact_Id(uuid.NewV4().String()),
				database.ExecutiveContact_VendorPk(vendor.Pk),
				database.ExecutiveContact_FirstName(e.FirstName),
				database.ExecutiveContact_LastName(e.LastName))
			if err != nil {
				return err
			}

			err = tx.CreateNoReturn_VendorEmail(ctx,
				database.VendorEmail_Id(uuid.NewV4().String()),
				database.VendorEmail_ExecutiveContactPk(exec.Pk),
				database.VendorEmail_Address(e.Email),
				database.VendorEmail_SaltedHash(hash))
			if err != nil {
				return err
			}

			err = tx.CreateNoReturn_VendorPhone(ctx,
				database.VendorPhone_Id(uuid.NewV4().String()),
				database.VendorPhone_ExecutiveContactPk(exec.Pk),
				database.VendorPhone_PhoneNumber(e.PhoneNumber),
				database.VendorPhone_CountryCode(e.CountryCode),
				database.VendorPhone_AreaCode(e.AreaCode))
			if err != nil {
				return err
			}
		}

		vendor_session, err = tx.Create_VendorSession(ctx,
			database.VendorSession_VendorPk(vendor.Pk),
			database.VendorSession_Id(uuid.NewV4().String()))
		if err != nil {
			return err
		}

		return nil

	})
	if err != nil {
		return nil, err
	}

	return &VendorSignUpResponse{Session: vendor_session, VendorId: vendor_id}, nil
}

//ValidateVendorSignUpRequest verifies the required data for regisering as a vendor. returns an
//error is there is a problem or incoming data does not match the requirement
func ValidateVendorSignUpRequest(vsr *VendorSignUpRequest) error {

	if len(vsr.ExecutiveContacts) > maxExecutiveContacts {
		return errs.New("only a max of %d contacts are allowed", maxExecutiveContacts)
	}

	for _, e := range vsr.ExecutiveContacts {
		if err := validate.CheckFullName(e.FirstName, e.LastName); err != nil {
			return err
		}

		if err := validate.CheckPassword(e.Password); err != nil {
			return err
		}

		if err := validate.CheckEmail(e.Email); err != nil {
			return err
		}
	}

	if err := validate.CheckAddress(vsr.BillingAddress); err != nil {
		return err
	}

	if !validate.AddressIsEmpty(vsr.ShippingAddress) {
		if err := validate.CheckAddress(vsr.ShippingAddress); err != nil {
			return err
		}
	}

	return nil

}
