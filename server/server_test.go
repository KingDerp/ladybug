package server

import (
	"testing"

	"ladybug/database"

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
