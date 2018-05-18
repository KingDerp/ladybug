package main

import (
	"context"
	"ladybug/handlers"
	"net/http"
)

func main() {
}

func run(ctx context.Context) error {
	handler := handlers.NewHandler()

	return http.ListenAndServe("localhost:8080", handler)
}
