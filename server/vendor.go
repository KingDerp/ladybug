package server

import (
	"context"
	"fmt"
	"ladybug/database"

	uuid "github.com/satori/go.uuid"
	"github.com/zeebo/errs"
	"gopkg.in/spacemonkeygo/dbx.v1/prettyprint"
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

	return &RegisterProductResponse{Response: "Product has been succesfully registered"}, nil
}

type GetVendorMessageRequest struct {
	VendorPk int64
}

type GetVendorMessageResponse struct {
	Messages []*serverMessage `json:"messages"`
}

func (v *VendorServer) GetVendorMessages(ctx context.Context, req *GetVendorMessageRequest) (
	resp *GetVendorMessageResponse, err error) {

	fmt.Printf("req.VendorPk: %d\n", req.VendorPk)

	message_response := &GetVendorMessageResponse{}
	err = v.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {

		messages, err := tx.All_Message_By_VendorPk(ctx, database.Message_VendorPk(req.VendorPk))
		if err != nil {
			return err
		}
		fmt.Println("all vendor messages from db")
		prettyprint.Println(messages)
		message_response.Messages = MessagesFromDB(messages)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return message_response, nil
}

type PostVendorMessageRequest struct {
	VendorPk  int64  `json:"-"`
	MessageId string `json:"messageId"`
	Message   string `json:"message"`
}

type PostVendorMessageResponse struct {
	Message *serverMessage `json:"message"`
}

func (v *VendorServer) PostVendorMessage(ctx context.Context, req *PostVendorMessageRequest) (
	resp *PostVendorMessageResponse, err error) {

	var message *database.Message
	err = v.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {

		parent_message, err := tx.Get_Message_By_Id(ctx, database.Message_Id(req.MessageId))
		if err != nil {
			return err
		}

		if parent_message.VendorPk != req.VendorPk {
			return errs.New("unauthorized")
		}

		message, err = tx.Create_Message(ctx,
			database.Message_VendorPk(req.VendorPk),
			database.Message_UserPk(parent_message.UserPk),
			database.Message_Id(uuid.NewV4().String()),
			database.Message_BuyerSent(false),
			database.Message_Message(req.Message),
			database.Message_Create_Fields{
				ParentPk: database.Message_ParentPk(parent_message.Pk),
			},
		)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &PostVendorMessageResponse{Message: MessageFromDB(message)}, nil
}
