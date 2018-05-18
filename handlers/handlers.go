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
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// The "/" pattern matches everything, so we need to check
		// that we're at the root here.
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		fmt.Fprintf(w, "Welcome to the home page!")
	})

	return &Handler{Handler: mux}
}
