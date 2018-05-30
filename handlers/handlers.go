package handlers

import (
	"fmt"
	"net/http"

	"ladybug/database"
)

type Handler struct {
	http.Handler
}

type authMiddleware struct {
	db *database.DB
}

func NewHandler(db *database.DB) *Handler {

	a := &authMiddleware{db: db}

	mux := http.NewServeMux()

	mux.Handle("/", a.Wrap(http.HandlerFunc(rootHandler)))
	//TODO(mac): write a handler that will allow a user to log in with user name and password using
	//bcrypt to check the incoming password with the one in the db
	mux.Handle("/user", a.Wrap(http.HandlerFunc(a.userHandler)))

	/*
			mux.Handle("/",
				a.Wrap(
					http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						// The "/" pattern matches everything, so we need to check
						// that we're at the root here.
						if req.URL.Path != "/" {
							http.NotFound(w, req)
							return
						}
						user := GetUser(req.Context())

						b, err := json.Marshal(user)
						if err != nil {
							http.Error(w, "server error", http.StatusInternalServerError)
							return
						}

						h := w.Header()
						h.Set("Content-Type", "application/json")
						w.Write(b)
					}),
				),
			)

		mux.Handle("/user",
			a.Wrap(
				http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					user := GetUser(req.Context())

					if req.Method != "POST" {
						http.Error(w, "only post requests allowed to /user",
							http.StatusBadRequest)
						return
					}

					decoder := json.NewDecoder(req.Body)
					var incoming_user incomingUser

					err := decoder.Decode(&incoming_user)
					if err != nil {
						http.Error(w, "unable to parse json", http.StatusBadRequest)
						return
					}

					//TODO(mac): come up with more robust validations
					if incoming_user.userName == "" || incoming_user.password == "" {
						http.Error(w, "user name and password cannot be empty strings",
							http.StatusBadRequest)
						return
					}

					hash, err := bcrypt.GenerateFromPassword([]byte(incoming_user.password))
					if err != nil {
						http.Error(w, "server error", http.StatusInternalServerError)
						return
					}

					h := w.Header()
					h.Set("Content-Type", "application/json")
					w.Write(b)
				}),
			),
		)
	*/

	return &Handler{Handler: mux}
}

func (a *authMiddleware) Wrap(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		cookie, err := req.Cookie("session")
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusUnauthorized)
			//w.WriteHeader(http.StatusUnauthorized)
			return
		}

		user, err := a.db.GetUserBySessionId(req.Context(), cookie.Value)
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusUnauthorized)
			//w.WriteHeader(http.StatusUnauthorized)
			return
		}

		c := req.Context()
		req = req.WithContext(WithUser(c, user))

		handler.ServeHTTP(w, req)
	})
}
