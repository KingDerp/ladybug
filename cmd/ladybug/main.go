package main

import (
	"context"
	"flag"
	"ladybug/database"
	"ladybug/handlers"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/zeebo/errs"
)

var (
	addressFlag = flag.String(
		"address",
		":8080",
		"the address ladybug binds to")
)

func main() {
	flag.Parse()

	err := run(context.Background())
	if err != nil {
		logrus.Errorf("%+v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {

	db, err := database.Open("postgres",
		"postgres://localhost/ladybug?user=ladybug&password=something_stupid")
	if err != nil {
		return err
	}

	handler := handlers.NewHandler(db)

	logrus.Infof("server listening on address %s\n", *addressFlag)
	return errs.Wrap(http.ListenAndServe(*addressFlag, handler))
}
