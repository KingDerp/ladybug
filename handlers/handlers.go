package handlers

import (
	"fmt"
	"net/http"

	"ladybug/database"
	"ladybug/server"
)

type Handler struct {
	http.Handler
}

type authMiddleware struct {
	db *database.DB
}

func NewHandler(db *database.DB) *Handler {

	a := &authMiddleware{db: db}
	us := server.NewUserServer(db)
	u := newUserHandler(us)

	vs := server.NewVendorServer(db)
	v := newVendorHandler(vs)

	mux := http.NewServeMux()

	//user endpoints
	mux.Handle("/", a.CheckUserSessionCookie(http.HandlerFunc(rootHandler)))
	mux.Handle("/user/login", http.HandlerFunc(u.userLogInHandler))
	mux.Handle("/user/sign-up", http.HandlerFunc(u.userSignUpHandler))
	mux.Handle("/user", a.CheckUserSessionCookie(http.HandlerFunc(u.userHandler)))

	//vendor endpoints
	mux.Handle("/vendor/sign-up", http.HandlerFunc(v.vendorSignUpHandler))
	mux.Handle("/vendor/product", a.CheckVendorSessionCookie(
		http.HandlerFunc(v.vendorProductHandler)))

	return &Handler{Handler: mux}
}

func (a *authMiddleware) CheckUserSessionCookie(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		cookie, err := req.Cookie("session")
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusUnauthorized)
			return
		}

		pk_row, err := a.db.Get_Session_UserPk_By_Id(req.Context(),
			database.Session_Id(cookie.Value))
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusUnauthorized)
			return
		}

		c := req.Context()
		req = req.WithContext(WithUserPk(c, pk_row.UserPk))

		handler.ServeHTTP(w, req)
	})
}

func (a *authMiddleware) CheckVendorSessionCookie(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		cookie, err := req.Cookie("session")
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
		req = req.WithContext(WithUserPk(c, pk_row.VendorPk))

		handler.ServeHTTP(w, req)
	})
}
