//go:generate dbx.v1 golang -d postgres -d sqlite3 -p database ladybug.dbx .
//go:generate dbx.v1 schema -d postgres -d sqlite3 ladybug.dbx .

package database

import (
	"context"

	"github.com/sirupsen/logrus"
)

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
