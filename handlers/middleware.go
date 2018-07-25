package handlers

import (
	"fmt"
	"net/http"

	"ladybug/database"
)

type authMiddleware struct {
	db *database.DB
}

func (a *authMiddleware) CheckVendorSessionCookie(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		cookie, err := req.Cookie("vendor_session")
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusUnauthorized)
			return
		}

		pk_row, err := a.db.Get_VendorSession_VendorPk_By_Id(req.Context(),
			database.VendorSession_Id(cookie.Value))
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusUnauthorized)
			return
		}

		c := req.Context()
		req = req.WithContext(WithBuyerPk(c, pk_row.VendorPk))

		handler.ServeHTTP(w, req)
	})
}

func (a *authMiddleware) CheckBuyerSessionCookie(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		cookie, err := req.Cookie("buyer_session")
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusUnauthorized)
			return
		}

		pk_row, err := a.db.Get_BuyerSession_BuyerPk_By_Id(req.Context(),
			database.BuyerSession_Id(cookie.Value))
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusUnauthorized)
			return
		}

		c := req.Context()
		req = req.WithContext(WithBuyerPk(c, pk_row.BuyerPk))

		handler.ServeHTTP(w, req)
	})
}
