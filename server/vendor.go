package server

import (
	"context"
	"fmt"
	"ladybug/database"

	uuid "github.com/satori/go.uuid"
)

type VendorServer struct {
	db *database.DB
}

func NewVendorServer(db *database.DB) *VendorServer {
	return &VendorServer{db: db}
}

type RegisterProductRequest struct {
	VendorPk       int64
	UnitPrice      float32 `json:"unitPrice"`
	Discount       float32 `json:"discountPrice"`
	DiscountActive bool    `json:"discountActive"`
	SKU            string  `json:"sku"`
	GoogleBucketId string  `json:"googleBucketId"`
	ProductActive  bool    `json:"productActive"`
	NumberInStock  int     `json:"numberInStock"`
	Description    string  `json:"description"`
}

type RegisterProductResponse struct {
	Response string `json:"response"`
}

func (v *VendorServer) RegisterProduct(ctx context.Context, req *RegisterProductRequest) (
	resp *RegisterProductResponse, err error) {

	err = v.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {

		err = tx.CreateNoReturn_Product(ctx,
			database.Product_Id(uuid.NewV4().String()),
			database.Product_VendorPk(req.VendorPk),
			database.Product_Price(req.UnitPrice),
			database.Product_Discount(req.Discount),
			database.Product_DiscountActive(req.DiscountActive),
			database.Product_Sku(req.SKU),
			database.Product_GoogleBucketId(req.GoogleBucketId),
			database.Product_LadybugApproved(false),
			database.Product_ProductActive(req.ProductActive),
			database.Product_NumInStock(req.NumberInStock),
			database.Product_Description(req.Description))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	fmt.Println("product successfully registered")
	return &RegisterProductResponse{Response: "Product has been succesfully registered"}, nil
}
