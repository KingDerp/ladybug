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
	s := server.NewUserServer(db)
	u := newUserHandler(s)

	mux := http.NewServeMux()

	mux.Handle("/", a.CheckSessionCookie(http.HandlerFunc(rootHandler)))
	mux.Handle("/login", http.HandlerFunc(u.userLogInHandler))
	mux.Handle("/user/sign-up", http.HandlerFunc(u.userSignUpHandler))
	mux.Handle("/user", a.CheckSessionCookie(http.HandlerFunc(u.userHandler)))

	return &Handler{Handler: mux}
}

func (a *authMiddleware) CheckSessionCookie(handler http.Handler) http.Handler {
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
