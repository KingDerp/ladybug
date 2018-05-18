package handlers

import (
	"context"
	"encoding/json"
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

type contextKey int

const (
	userContextKey contextKey = iota
)

func WithUser(ctx context.Context, user *database.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func GetUser(ctx context.Context) *database.User {
	user, _ := ctx.Value(userContextKey).(*database.User)
	return user
}
