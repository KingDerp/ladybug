package main

import (
	"context"
	"fmt"
	"ladybug/handlers"
	"net/http"
	"os"

	"github.com/zeebo/errs"
)

func main() {
	err := run(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
	}
}

func run(ctx context.Context) error {
	handler := handlers.NewHandler()

	return errs.Wrap(http.ListenAndServe("localhost:8080", handler))
}
