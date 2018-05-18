package handlers

import (
	"fmt"
	"net/http"
)

type Handler struct {
	http.Handler
}

func NewHandler() *Handler {
	mux := http.NewServeMux()
	mux.Handle("/",
		Wrap(
			http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				// The "/" pattern matches everything, so we need to check
				// that we're at the root here.
				if req.URL.Path != "/" {
					http.NotFound(w, req)
					return
				}
				fmt.Fprintf(w, "Welcome to the home page!")
			}),
		),
	)

	return &Handler{Handler: mux}
}

func Wrap(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("session")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		cookie = cookie

		handler.ServeHTTP(w, r)
	})
}
