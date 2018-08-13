package server

import (
	"context"
	"ladybug/database"

	uuid "github.com/satori/go.uuid"
	"github.com/zeebo/errs"
)

type ProductReviewReq struct {
	BuyerPk     int64
	ProductId   string `json:"productId"`
	Stars       int    `json:"stars"`
	Description string `json:"description"`
}

type ProductReviewResp struct {
	ReviewResponeMessage string `json:"reviewResponseMessage"`
}

func (u *BuyerServer) ReviewProduct(ctx context.Context, req *ProductReviewReq) (
	resp *ProductReviewResp, err error) {

	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		has_purchased, err := tx.Has_PurchasedProduct_By_BuyerPk(ctx,
			database.PurchasedProduct_BuyerPk(req.BuyerPk))
		if err != nil {
			return err
		}

		if !has_purchased {
			return errs.New("you cannot leave a review for a product you have not purchased")
		}

		has_review, err := tx.Has_ProductReview_By_Product_Id_And_ProductReview_BuyerPk(ctx,
			database.Product_Id(req.ProductId),
			database.ProductReview_BuyerPk(req.BuyerPk))
		if err != nil {
			return err
		}

		if has_review {
			return errs.New("you have already left a review for this product")
		}

		product_pk_field, err := tx.Get_Product_By_Id(ctx, database.Product_Id(req.ProductId))
		if err != nil {
			return err
		}

		err = tx.CreateNoReturn_ProductReview(ctx,
			database.ProductReview_Id(uuid.NewV4().String()),
			database.ProductReview_BuyerPk(req.BuyerPk),
			database.ProductReview_ProductPk(product_pk_field.Pk),
			database.ProductReview_Rating(req.Stars),
			database.ProductReview_Description(req.Description))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &ProductReviewResp{ReviewResponeMessage: "Thank you for leaving a review!"}, nil
}

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
