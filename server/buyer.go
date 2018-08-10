package server

import (
	"context"
	"strings"

	"ladybug/database"

	uuid "github.com/satori/go.uuid"
	"github.com/zeebo/errs"
	"golang.org/x/crypto/bcrypt"
)

const (
	productRequestLimit      = 25
	conversationRequestLimit = 20
)

type BuyerServer struct {
	db *database.DB
}

func NewBuyerServer(db *database.DB) *BuyerServer {
	return &BuyerServer{db: db}
}

type BuyerEmail struct {
	Address string
}

type Buyer struct {
	FirstName string        `jsons:"firstName"`
	LastName  string        `jsons:"lastName"`
	Emails    []*BuyerEmail `json:"emails"`
}

func BuyerEmailFromDB(email *database.BuyerEmail) *BuyerEmail {
	return &BuyerEmail{
		Address: email.Address,
	}
}

func BuyerEmailsFromDB(emails []*database.BuyerEmail) []*BuyerEmail {
	out := []*BuyerEmail{}
	for _, email := range emails {
		out = append(out, BuyerEmailFromDB(email))
	}
	return out

}

func BuyerFromDB(buyer *database.Buyer, emails []*database.BuyerEmail) *Buyer {
	return &Buyer{
		FirstName: buyer.FirstName,
		LastName:  buyer.LastName,
		Emails:    BuyerEmailsFromDB(emails),
	}
}

type GetBuyerRequest struct {
	BuyerPk int64
}

type GetBuyerResponse struct {
	Buyer *Buyer
}

func (u *BuyerServer) GetBuyer(ctx context.Context, req *GetBuyerRequest) (
	resp *GetBuyerResponse, err error) {

	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		buyer, err := tx.Get_Buyer_By_Pk(ctx, database.Buyer_Pk(req.BuyerPk))
		if err != nil {
			return err
		}

		emails, err := tx.All_BuyerEmail_By_BuyerPk(ctx, database.BuyerEmail_BuyerPk(req.BuyerPk))
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

type LogInRequest struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

func (u *BuyerServer) BuyerLogIn(ctx context.Context, req *LogInRequest) (
	resp *database.BuyerSession, err error) {

	var email *database.BuyerEmail
	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		email, err = tx.Find_BuyerEmail_By_Address(ctx,
			database.BuyerEmail_Address(strings.ToLower(req.Email)))
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

type Product struct {
	Id             string  `json:"id"`
	Price          float32 `json:"price"`
	Discount       float32 `json:"discount"`
	DiscountActive bool    `json:"discountActive"`
	Sku            string  `json:"sku"`
	GoogleBucketId string  `json:"googleBucketId"`
	NumInStock     int     `json:"numInStock"`
	Description    string  `json:"description"`
}

type ProductResponse struct {
	Products  []*Product `json:"products"`
	PageToken string     `json:"pageToken"`
}

type ProductRequest struct {
	PageToken string `json:"pageToken"`
}

func ProductsFromDB(db_products []*database.Product) []*Product {

	products := []*Product{}
	for _, p := range db_products {
		products = append(products, &Product{
			Id:             p.Id,
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

func (u *BuyerServer) BuyerProducts(ctx context.Context, req *ProductRequest) (
	resp *ProductResponse, err error) {

	product_response := &ProductResponse{}
	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {

		db_products, ctoken, err := tx.Paged_Product_By_ProductActive_Equal_True_And_LadybugApproved_Equal_True_And_NumInStock_Not_Number(
			ctx, productRequestLimit, req.PageToken)
		if err != nil {
			return err
		}

		product_response.Products = ProductsFromDB(db_products)
		product_response.PageToken = ctoken

		return nil
	})
	if err != nil {
		return nil, err
	}

	return product_response, nil
}

//hashPassword takes a string, creates a hash using that string and returns the hash in a string
//format. One should note that GenerateFromPassword uses base64 encoding in it's logic.
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errs.Wrap(err)
	}
	return string(hash), err
	//return base64.URLEncoding.EncodeToString(hash), nil
}

//comparePasswordHash - this funtion takes the unhashed version of the password and the hash to
//compare it against. returns an error if there was an internal problem or if the hashed and
//unhashed password do not match
func comparePasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return errs.New("email or password does not match")
	}
	return nil
}
