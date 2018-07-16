package server

import (
	"context"
	"ladybug/database"

	uuid "github.com/satori/go.uuid"
)

type VendorServer struct {
	db *database.DB
}

func NewVendorServer(db *database.DB) *VendorServer {
	return &VendorServer{db: db}
}

//NOTE: Executive contacts have full access to a vendor's portal. there are presently no agent roles
//set up
type ExecutiveContact struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	CountryCode int    `json:"countryCode"`
	AreaCode    int    `json:"areaCode"`
	PhoneNumber int    `json:"phoneNumber"`
	Email       string `json:"email"`
	Password    string `json:"password"`
}

//TODO(mac): I should look for the form of a Fein and validate it.
type VendorSignUpRequest struct {
	Fein              string              `json:"fein"`
	BillingAddress    *Address            `json:"billingAddress"`
	ShippingAddress   *Address            `json"shippingAddress"`
	ExecutiveContacts []*ExecutiveContact `json"executiveContact"`
}

type VendorSignUpResponse struct {
	Session *database.VendorSession
}

func (v *VendorServer) VendorSignUp(ctx context.Context, req *VendorSignUpRequest) (
	resp *VendorSignUpResponse, err error) {

	err = validateVendorSignUpRequest(req)
	if err != nil {
		return nil, err
	}

	ship_addr_is_empty := addressIsEmpty(req.ShippingAddress)

	if !ship_addr_is_empty {
		if err := validateAddress(req.ShippingAddress); err != nil {
			return nil, err
		}
	}

	var vendor_session *database.VendorSession
	err = v.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {

		vendor, err := tx.Create_Vendor(ctx,
			database.Vendor_Id(uuid.NewV4().String()),
			database.Vendor_Fein(req.Fein))
		if err != nil {
			return err
		}

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

	return &VendorSignUpResponse{Session: vendor_session}, nil
}
