package server

import (
	"context"

	"ladybug/database"
	"ladybug/validate"

	"github.com/zeebo/errs"
)

type UpdateBuyerRequest struct {
	BuyerPk         int64
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
	CurrentEmail    string `json:"currentEmail"`
	NewEmail        string `json:"newEmail"`
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"password"`
}

type UpdateBuyerResponse struct {
	Buyer *Buyer
}

type UpdateRequestField struct {
	set   bool
	value string
}

type BuyerRequestFields struct {
	FirstName       UpdateRequestField
	LastName        UpdateRequestField
	NewEmail        UpdateRequestField
	CurrentEmail    UpdateRequestField
	CurrentPassword UpdateRequestField
	NewPassword     UpdateRequestField
}

func (u *BuyerServer) UpdateBuyer(ctx context.Context, req *UpdateBuyerRequest) (
	resp *UpdateBuyerResponse, err error) {

	buyer_req_fields := BuyerRequestFieldsFromUpdateRequest(req)

	if err := ValidateUpdateBuyerRequestFields(buyer_req_fields); err != nil {
		return nil, err
	}

	//set buyer updates
	has_buyer_updates := buyer_req_fields.hasBuyerUpdates()
	buyer_updates := &database.Buyer_Update_Fields{}
	if has_buyer_updates {
		buyer_updates = buyer_req_fields.makeBuyerUpdateFields()
	}

	//set buyer email updates
	has_email_updates := buyer_req_fields.hasEmailUpdates()
	email_updates := &database.BuyerEmail_Update_Fields{}
	if has_email_updates {
		email_updates, err = buyer_req_fields.makeEmailUpdateFields()
		if err != nil {
			return nil, err
		}
	}

	var buyer *database.Buyer
	var emails []*database.BuyerEmail
	if has_buyer_updates || has_email_updates {
		err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {

			if has_buyer_updates {
				buyer, err = tx.Update_Buyer_By_Pk(ctx, database.Buyer_Pk(req.BuyerPk),
					database.Buyer_Update_Fields(*buyer_updates))
				if err != nil {
					return err
				}
			}

			if has_email_updates {
				email, err := tx.Get_BuyerEmail_By_Address(ctx, database.BuyerEmail_Address(
					*buyer_req_fields.CurrentEmail.Value()))
				if err != nil {
					return err
				}

				err = comparePasswordHash(*buyer_req_fields.CurrentPassword.Value(),
					email.SaltedHash)
				if err != nil {
					return err
				}

				err = tx.UpdateNoReturn_BuyerEmail_By_Address(ctx,
					database.BuyerEmail_Address(*buyer_req_fields.CurrentEmail.Value()),
					database.BuyerEmail_Update_Fields(*email_updates))
				if err != nil {
					return err
				}

				emails, err = tx.All_BuyerEmail_By_BuyerPk(ctx, database.BuyerEmail_BuyerPk(
					email.BuyerPk))
				if err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return &UpdateBuyerResponse{Buyer: BuyerFromDB(buyer, emails)}, nil
}

func SetRequestField(v string) UpdateRequestField {
	return UpdateRequestField{set: true, value: v}
}

func (f UpdateRequestField) Value() *string {
	if !f.set {
		return nil
	}

	return &f.value
}

func (f BuyerRequestFields) isEmpty() bool {
	if f.FirstName.set == false &&
		f.LastName.set == false &&
		f.CurrentEmail.set == false &&
		f.NewEmail.set == false &&
		f.CurrentPassword.set == false &&
		f.NewPassword.set == false {
		return true
	}

	return false
}

func (f *BuyerRequestFields) makeBuyerUpdateFields() *database.Buyer_Update_Fields {

	buyer_updates := &database.Buyer_Update_Fields{}

	if f.FirstName.set {
		buyer_updates.FirstName = database.Buyer_FirstName(*f.FirstName.Value())
	}

	if f.LastName.set {
		buyer_updates.LastName = database.Buyer_LastName(*f.LastName.Value())
	}

	return buyer_updates
}

func (f BuyerRequestFields) makeEmailUpdateFields() (updates *database.BuyerEmail_Update_Fields,
	err error) {

	email_updates := &database.BuyerEmail_Update_Fields{}

	if f.NewEmail.set {
		email_updates.Address = database.BuyerEmail_Address(*f.NewEmail.Value())
	}

	if f.NewPassword.set {
		hash, err := hashPassword(*f.NewPassword.Value())
		if err != nil {
			return nil, err
		}

		email_updates.SaltedHash = database.BuyerEmail_SaltedHash(hash)
	}

	return email_updates, nil
}

func (f BuyerRequestFields) hasBuyerUpdates() bool {
	return f.FirstName.set || f.LastName.set
}

func (f BuyerRequestFields) hasEmailUpdates() bool {
	return f.NewEmail.set || f.NewPassword.set
}

func BuyerRequestFieldsFromUpdateRequest(req *UpdateBuyerRequest) *BuyerRequestFields {
	buyer_req_fields := &BuyerRequestFields{}

	if req.FirstName != "" {
		buyer_req_fields.FirstName = SetRequestField(req.FirstName)
	}
	if req.LastName != "" {
		buyer_req_fields.LastName = SetRequestField(req.LastName)
	}
	if req.CurrentEmail != "" {
		buyer_req_fields.CurrentEmail = SetRequestField(req.CurrentEmail)
	}
	if req.NewEmail != "" {
		buyer_req_fields.NewEmail = SetRequestField(req.NewEmail)
	}
	if req.CurrentPassword != "" {
		buyer_req_fields.CurrentPassword = SetRequestField(req.CurrentPassword)
	}
	if req.NewPassword != "" {
		buyer_req_fields.NewPassword = SetRequestField(req.NewPassword)
	}

	return buyer_req_fields

}

func ValidateUpdateBuyerRequestFields(req *BuyerRequestFields) error {
	if req.isEmpty() {
		return errs.New("all fields are empty. nothing to update")
	}

	if req.FirstName.set {
		if err := validate.CheckName(*req.FirstName.Value()); err != nil {
			return err
		}
	}

	if req.LastName.set {
		if err := validate.CheckName(*req.LastName.Value()); err != nil {
			return err
		}
	}

	if !req.CurrentEmail.set {
		return errs.New("current email must be set")
	}

	if req.NewEmail.set {
		if err := validate.CheckEmail(*req.NewEmail.Value()); err != nil {
			return err
		}
	}

	if !req.CurrentPassword.set {
		return errs.New("current password must be set")
	}

	if req.NewPassword.set {
		if err := validate.CheckPassword(*req.NewPassword.Value()); err != nil {
			return err
		}
	}

	return nil
}

func updateBuyerRequestIsEmpty(req *UpdateBuyerRequest) bool {
	if req.FirstName == "" &&
		req.LastName == "" &&
		req.CurrentEmail == "" &&
		req.NewEmail == "" &&
		req.CurrentPassword == "" &&
		req.NewPassword == "" {
		return true
	}

	return false

}
