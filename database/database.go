package database

import (
	"context"
	"database/sql"
	"time"

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

type Email struct {
	Pk         int64     `json:"-"`
	UserPk     int64     `json:"-"`
	FullEmail  string    `json:"fullEmail"`
	CreateDate time.Time `json:"-"`
	SaltedHash string    `json:"-"`
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

func (db *DB) CreateUserWithEmailNoReturn(full_name, full_email, salted_hash string) error {

	user := User{}
	err := db.raw.QueryRow(`INSERT INTO users (fullName) VALUES ($1)`, full_name).
		Scan(&user.Pk, &user.FullName)
	if err != nil {
		return errs.Wrap(err)
	}

	result, err := db.raw.Exec(`INSERT INTO emails (userPk, fullEmail, createDate, saltedHash) VALUES 
	($1, $2, $3, $4)`, user.Pk, full_email, time.Now().UTC(), salted_hash)
	if err != nil {
		return errs.Wrap(err)
	}

	num_rows, err := result.RowsAffected()
	if err != nil {
		return errs.Wrap(err)
	}
	if num_rows != 1 {
		return errs.New("expected 1 row to be affected got %d", num_rows)
	}

	return nil
}

//func (db *DB) CreateEmail()

//TODO(mac): make an email table that maps back to a user (multuple emails allowed)
//TODO(mac): make a salted password hash
//table emails should have user_pk, password hash, and a salt
