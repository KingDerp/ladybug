package server

import (
	"context"
	"ladybug/database"

	uuid "github.com/satori/go.uuid"
)

type StartProductTrialReq struct {
	BuyerPk   int64
	VendorId  string `json:"vendorId"`
	ProductId string `json:"productId"`
}

type StartProductTrialResp struct {
	TrialProduct *TrialProduct `json:"trialProduct"`
}

func (u *BuyerServer) StartProductTrial(ctx context.Context, req *StartProductTrialReq) (
	resp *StartProductTrialResp, err error) {

	var trial_product *database.TrialProduct
	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {

		vendor_pk_field, err := tx.Get_Vendor_Pk_By_Id(ctx,
			database.Vendor_Id(req.VendorId))
		if err != nil {
			return err
		}

		product, err := tx.Get_Product_Pk_Product_Price_By_Id(ctx,
			database.Product_Id(req.ProductId))
		if err != nil {
			return err
		}

		trial_product, err = tx.Create_TrialProduct(ctx,
			database.TrialProduct_Id(uuid.NewV4().String()),
			database.TrialProduct_VendorPk(vendor_pk_field.Pk),
			database.TrialProduct_BuyerPk(req.BuyerPk),
			database.TrialProduct_ProductPk(product.Pk),
			database.TrialProduct_TrialPrice(product.Price),
			database.TrialProduct_IsReturned(false),
		)
		if err != nil {
			return err
		}

		return nil
	})

	return &StartProductTrialResp{
		TrialProduct: TrialFromDB(trial_product),
	}, nil
}
