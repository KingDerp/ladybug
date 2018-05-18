package database

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/zeebo/errs"
)

type DB struct {
	raw *sql.DB
}

func Open(url string) (out *DB, err error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	err = db.Ping()
	if err != nil {
		return nil, errs.Wrap(err)
	}

	return &DB{raw: db}, nil
}

type User struct {
	Pk       int64  `json:"-"`
	FullName string `json:"fullName"`
}

func (db *DB) GetUserBySessionId(ctx context.Context, session_id string) (
	out *User, err error) {

	user := &User{}

	err = db.raw.QueryRow(`SELECT users.pk, users.fullName FROM users 
	  JOIN sessions ON users.pk = sessions.user_pk WHERE sessions.id = $1`,
		session_id).Scan(&user.Pk, &user.FullName)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	return user, nil
}
