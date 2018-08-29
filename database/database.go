//go:generate dbx.v1 golang -d postgres -d sqlite3 -p database ladybug.dbx .
//go:generate dbx.v1 schema -d postgres -d sqlite3 ladybug.dbx .

package database

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/zeebo/errs"
)

func init() {
	WrapErr = func(e *Error) error {
		return errs.Wrap(e)
	}
	Logger = func(format string, args ...interface{}) {
		fmt.Printf(format+"\n", args...)
	}
}

func (db *DB) WithTx(ctx context.Context,
	fn func(context.Context, *Tx) error) (err error) {

	tx, err := db.Open(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
		} else {
			logrus.Error(err)
			tx.Rollback()
		}
	}()
	return fn(ctx, tx)
}

func IsConstraintViolationError(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == ErrorCode_ConstraintViolation
	}

	return false
}
