package server

import (
	"context"
	"ladybug/database"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/zeebo/errs"
	"golang.org/x/crypto/bcrypt"
)

const (
	productRequestLimit = 10
)

type BuyerServer struct {
	db *database.DB
}

func NewBuyerServer(db *database.DB) *BuyerServer {
	return &BuyerServer{db: db}
}

type GetBuyerRequest struct {
	BuyerPk int64
}

type GetBuyerResponse struct {
	Buyer *Buyer
}

type Email struct {
	Address string
}

type Buyer struct {
	FullName  string   `json:"fullName"`
	FirstName string   `jsons:"firstName"`
	LastName  string   `jsons:"lastName"`
	Emails    []*Email `json:"emails"`
}

func EmailFromDB(email *database.Email) *Email {
	return &Email{
		Address: email.Address,
	}
}

func EmailsFromDB(emails []*database.Email) []*Email {
	out := []*Email{}
	for _, email := range emails {
		out = append(out, EmailFromDB(email))
	}
	return out

}

func BuyerFromDB(buyer *database.Buyer, emails []*database.Email) *Buyer {
	return &Buyer{
		FirstName: buyer.FirstName,
		LastName:  buyer.LastName,
		Emails:    EmailsFromDB(emails),
	}
}

func (u *BuyerServer) GetBuyer(ctx context.Context, req *GetBuyerRequest) (
	resp *GetBuyerResponse, err error) {

	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		buyer, err := tx.Get_Buyer_By_Pk(ctx, database.Buyer_Pk(req.BuyerPk))
		if err != nil {
			return err
		}

		emails, err := tx.All_Email_By_BuyerPk(ctx, database.Email_BuyerPk(req.BuyerPk))
		if err != nil {
			return err
		}

		resp = &GetBuyerResponse{Buyer: BuyerFromDB(buyer, emails)}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type UpdateBuyerRequest struct {
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
	Email           string `json:"email"`
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"password"`
}

type UpdateBuyerResponse struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

func (u *BuyerServer) UpdateBuyer(ctx context.Context, req *UpdateBuyerRequest) (
	resp *UpdateBuyerResponse, err error) {

	var email *database.Email
	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		email, err = tx.Get_Email_By_Address(ctx, database.Email_Address(req.Email))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	hash, err := hashPassword(req.CurrentPassword)
	if err != nil {
		return nil, err
	}

	if err := comparePasswordHash(hash, email.SaltedHash); err != nil {
		return nil, err
	}

	buyer_updates := database.Buyer_Update_Fields{
		FirstName: database.Buyer_FirstName(req.FirstName),
		LastName:  database.Buyer_LastName(req.LastName),
	}

	email_updates := database.Email_Update_Fields{
		Address:    database.Email_Address(req.Email),
		SaltedHash: database.Email_SaltedHash(hash),
	}

	var buyer *database.Buyer
	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		buyer, err = tx.Update_Buyer_By_Pk(ctx, database.Buyer_Pk(email.BuyerPk), buyer_updates)
		if err != nil {
			return err
		}

		email, err = tx.Update_Email_By_Pk(ctx, database.Email_Pk(email.Pk), email_updates)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &UpdateBuyerResponse{
		FirstName: buyer.FirstName,
		LastName:  buyer.LastName,
		Email:     email.Address,
	}, nil
}

type LogInRequest struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

func (u *BuyerServer) BuyerLogIn(ctx context.Context, req *LogInRequest) (
	resp *database.BuyerSession, err error) {

	var email *database.Email
	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		email, err = tx.Find_Email_By_Address(ctx, database.Email_Address(
			strings.ToLower(req.Email)))
		if err != nil {
			return err
		}

		if email == nil {
			return errs.New("No email exists with that address")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := comparePasswordHash(req.Password, email.SaltedHash); err != nil {
		return nil, err
	}

	var session *database.BuyerSession
	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {

		session, err = tx.Create_BuyerSession(ctx, database.BuyerSession_BuyerPk(email.BuyerPk),
			database.BuyerSession_Id(uuid.NewV4().String()))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return session, nil
}

//serverProduct represents data from database.Products that is safe for public consumption
//TODO(mac): I don't like the name of this struct is there a better name to differentiate it from
//the database object?
type serverProduct struct {
	Price          float32 `json:"price"`
	Discount       float32 `json:"discount"`
	DiscountActive bool    `json:"discountActive"`
	Sku            string  `json:"sku"`
	GoogleBucketId string  `json:"googleBucketId"`
	NumInStock     int     `json:"numInStock"`
	Description    string  `json:"description"`
}

type ProductResponse struct {
	Products []*serverProduct `json:"products"`
}

type ProductRequest struct {
	ProductCategories []string `json:"productCategory"`
}

func ProductsFromDB(db_products []*database.Product) []*serverProduct {

	products := []*serverProduct{}
	for _, p := range db_products {
		products = append(products, &serverProduct{
			Price:          p.Price,
			Discount:       p.Discount,
			DiscountActive: p.DiscountActive,
			Sku:            p.Sku,
			GoogleBucketId: p.GoogleBucketId,
			NumInStock:     p.NumInStock,
			Description:    p.Description,
		})
	}

	return products
}

//TODO(mac): this endpoint needs to be paginated
func (u *BuyerServer) BuyerProducts(ctx context.Context, req *ProductRequest) (
	resp *ProductResponse, err error) {

	//TODO(mac): eventually we need to use the request to search for products by category
	product_response := &ProductResponse{}
	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {

		db_products, err := tx.All_Product_By_ProductActive_Equal_True(ctx)
		if err != nil {
			return err
		}

		product_response.Products = ProductsFromDB(db_products)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return product_response, nil
}

type GetBuyerMessageRequest struct {
	BuyerPk int64
}

type GetBuyerMessagesResponse struct {
	Messages []*serverMessage
}

type serverMessage struct {
	Id        string    `json:"messageId"`
	CreatedAt time.Time `json:"createdAt"`
	BuyerSent bool      `json:"buyerSent"`
	Message   string    `json:"message"`
}

func MessagesFromDB(db_messages []*database.Message) []*serverMessage {
	server_messages := []*serverMessage{}
	for _, m := range db_messages {
		server_messages = append(server_messages, MessageFromDB(m))
	}

	return server_messages
}

func (u *BuyerServer) GetBuyerMessages(ctx context.Context, req *GetBuyerMessageRequest) (
	resp *GetBuyerMessagesResponse, err error) {

	message_response := &GetBuyerMessagesResponse{}
	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {

		messages, err := tx.All_Message_By_BuyerPk(ctx,
			database.Message_BuyerPk(req.BuyerPk))
		if err != nil {
			return err
		}

		message_response.Messages = MessagesFromDB(messages)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return message_response, nil
}

type PostBuyerMessageRequest struct {
	BuyerPk  int64  `json:"-"`
	VendorId string `json:"vendorId"`
	Message  string `json:"message"`
}

type PostBuyerMessageResponse struct {
	Message *serverMessage `json:"message"`
}

func MessageFromDB(db_message *database.Message) *serverMessage {
	return &serverMessage{
		Id:        db_message.Id,
		CreatedAt: db_message.CreatedAt,
		BuyerSent: db_message.BuyerSent,
		Message:   db_message.Message,
	}
}

func (u *BuyerServer) PostBuyerMessage(ctx context.Context, req *PostBuyerMessageRequest) (
	resp *PostBuyerMessageResponse, err error) {

	var message *database.Message
	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {

		pk_row, err := tx.Get_Vendor_Pk_By_Id(ctx,
			database.Vendor_Id(req.VendorId))
		if err != nil {
			return err
		}

		message, err = tx.Create_Message(ctx,
			database.Message_VendorPk(pk_row.Pk),
			database.Message_BuyerPk(req.BuyerPk),
			database.Message_Id(uuid.NewV4().String()),
			database.Message_BuyerSent(true),
			database.Message_Message(req.Message),
			database.Message_Create_Fields{},
		)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &PostBuyerMessageResponse{Message: MessageFromDB(message)}, nil
}

//HashPassword takes a string, creates a hash using that string and returns the hash in a string
//format. One should note that GenerateFromPassword uses base64 encoding in it's logic.
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errs.Wrap(err)
	}
	return string(hash), err
	//return base64.URLEncoding.EncodeToString(hash), nil
}

//ComparePasswordHash - this funtion takes the unhashed version of the password and the hash to
//compare it against. returns an error if there was an internal problem or if the hashed and
//unhashed password do not match
func comparePasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return errs.New("email or password does not match")
	}
	return nil
}
