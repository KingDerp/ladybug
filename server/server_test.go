package server

import (
	"context"
	"testing"

	"ladybug/database"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

type serverTest struct {
	t            *testing.T
	db           *database.DB
	BuyerServer  *BuyerServer
	VendorServer *VendorServer
}

//NOTE: just as a reminder while you are going through your tests create convience funtions that do
//things like create buyers etc
func newTest(t *testing.T) *serverTest {

	db, err := database.Open("sqlite3", "file:memdb1?mode=memory&cache=shared")
	require.NoError(t, err)

	//initialize database with schema
	_, err = db.Exec(db.Schema())
	require.NoError(t, err)

	buyer_server := NewBuyerServer(db)
	vendor_server := NewVendorServer(db)

	return &serverTest{
		t:            t,
		db:           db,
		BuyerServer:  buyer_server,
		VendorServer: vendor_server,
	}
}

func (h *serverTest) tearDown() {
	require.NoError(h.t, h.db.Close())
}

//createVendorInDB creates a vendor in the test database and returns the database struct for Vendor
//no other data is created for the vendor i.e. no email, phone, address etc.
func (h *serverTest) createVendorInDB(ctx context.Context) *database.Vendor {
	vendor, err := h.db.Create_Vendor(ctx,
		database.Vendor_Id(uuid.NewV4().String()),
		database.Vendor_Fein(uuid.NewV4().String()),
	)
	require.NoError(h.t, err)

	return vendor
}

type productOptions struct {
	Price           float32
	Discount        float32
	DiscountActive  bool
	Sku             string
	GoogleBucketId  string
	LadybugApproved bool
	ProductActive   bool
	NumInStock      int
	Description     string
}

//setDefaultOptions allows the caller of the function to pass in variables
func (p *productOptions) setDefaultProductOptions() {
	min := float32(1.25)
	max := float32(100.25)

	if p.Price == 0 {
		p.Price = randFloat(min, max)
	}

	if p.Discount == 0 {
		p.Price = randFloat(min, max)
	}

	if p.GoogleBucketId == "" {
		p.GoogleBucketId = "some_google_bucket_id"
	}

	if p.Description == "" {
		p.Description = "some_product_description"
	}
}

func (h *serverTest) createProductsInDB(ctx context.Context, n int, vendor_pk int64,
	options *productOptions) (products []*database.Product) {

	options.setDefaultProductOptions()
	products = []*database.Product{}

	for i := 0; i < n; i++ {
		p := h.createProductInDB(ctx, n, vendor_pk, options)
		products = append(products, p)
	}

	return products
}

func (h *serverTest) createProductInDB(ctx context.Context, n int, vendor_pk int64,
	options *productOptions) (product *database.Product) {
	p, err := h.db.Create_Product(ctx,
		database.Product_Id(uuid.NewV4().String()),
		database.Product_VendorPk(vendor_pk),
		database.Product_Price(options.Price),
		database.Product_Discount(options.Discount),
		database.Product_DiscountActive(options.DiscountActive),
		database.Product_Sku(uuid.NewV4().String()),
		database.Product_GoogleBucketId(options.GoogleBucketId),
		database.Product_LadybugApproved(options.LadybugApproved),
		database.Product_ProductActive(options.ProductActive),
		database.Product_NumInStock(options.NumInStock),
		database.Product_Description(options.Description),
	)
	require.NoError(h.t, err)

	return p
}

func (h *serverTest) createActiveAndApprovedProductsInStock(ctx context.Context, n int, vendor_pk int64) (
	products []*database.Product) {
	return h.createProductsInDB(ctx, n, vendor_pk,
		&productOptions{
			ProductActive:   true,
			LadybugApproved: true,
			NumInStock:      10,
		})
}

func (h *serverTest) createInactiveAndApprovedProductsInStock(ctx context.Context, n int, vendor_pk int64) (
	products []*database.Product) {
	return h.createProductsInDB(ctx, n, vendor_pk,
		&productOptions{
			ProductActive:   false,
			LadybugApproved: true,
			NumInStock:      10,
		})
}

func (h *serverTest) createActiveProductsNotApprovedInStock(ctx context.Context, n int, vendor_pk int64) (
	products []*database.Product) {
	return h.createProductsInDB(ctx, n, vendor_pk,
		&productOptions{
			ProductActive:   true,
			LadybugApproved: false,
			NumInStock:      10,
		})
}

func (h *serverTest) allProductsAreActive(products []*Product) bool {
	ctx := context.Background()
	for _, p := range products {

		db_product, err := h.db.Get_Product_By_Id(ctx, database.Product_Id(p.Id))
		require.NoError(h.t, err)

		if db_product.ProductActive == false {
			return false
		}
	}
	return true
}
