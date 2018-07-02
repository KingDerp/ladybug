package database

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"io"
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
	Pk       int64
	FullName string
}

type Email struct {
	Pk         int64
	UserPk     int64
	Address    string
	CreateDate time.Time
	SaltedHash string
}

type Session struct {
	Pk         int64
	UserPk     int64
	Id         string
	CreateDate time.Time
}

type UserUpdateFields struct {
	Pk       int64  `json:"-"`
	FullName string `json:"fullName"`
}

func (db *DB) UpdateUser(ctx context.Context, u *UserUpdateFields) error {

	_, err := db.raw.Exec(`UPDATE users SET fullName = $1 WHERE pk = $2`,
		u.FullName, u.Pk)
	if err != nil {
		return errs.Wrap(err)
	}

	return nil
}

type EmailUpdateFields struct {
	Pk         int64  `json:"-"`
	Email      string `json:"email"`
	SaltedHash string `json:"-"`
}

func (db *DB) UpdateEmail(ctx context.Context, e *EmailUpdateFields) error {

	_, err := db.raw.Exec(`UPDATE emails SET email = $1, saltedHash = $2 WHERE pk = $3`,
		e.Email, e.SaltedHash, e.Pk)
	if err != nil {
		return errs.Wrap(err)
	}

	return nil
}

func (db *DB) GetUserByPk(ctx context.Context, user_pk int64) (user *User,
	err error) {

	user = &User{}

	err = db.raw.QueryRow(`SELECT userPk, fullName FROM user WHERE userPk = $1`, user_pk).
		Scan(&user.Pk, &user.FullName)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	return user, nil
}

func (db *DB) GetUserPkBySessionId(ctx context.Context, session_id string) (
	pk int64, err error) {

	err = db.raw.QueryRow(`SELECT userPk FROM sessions WHERE sessions.id = $1`,
		session_id).Scan(&pk)
	if err != nil {
		return 0, errs.Wrap(err)
	}

	return pk, nil
}

func (db *DB) CreateSession(ctx context.Context, user_pk int64) (out *Session, err error) {

	b := make([]byte, 32)

	_, err = io.ReadFull(rand.Reader, b)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	id := base64.URLEncoding.EncodeToString(b)

	out = &Session{}
	now := time.Now()
	err = db.raw.QueryRow(`INSERT INTO sessions (userPk, id, createDate, lastUsedDate) VALUES (
	  $1,$2,$3,$4) RETURNING userPk, id, createDate`, user_pk, id, now, now).Scan(
		&out.UserPk, &out.Id, &out.CreateDate)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	return out, nil
}

func (db *DB) CreateUserWithEmail(ctx context.Context, full_name, address,
	salted_hash string) (user *User, err error) {

	user = &User{}
	err = db.raw.QueryRow(`INSERT INTO users (fullName) VALUES ($1) RETURNING pk, fullName`,
		full_name).Scan(&user.Pk, &user.FullName)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	_, err = db.raw.Exec(`INSERT INTO emails (userPk, address, createDate, saltedHash) VALUES 
	($1, $2, $3, $4)`, user.Pk, address, time.Now().UTC(), salted_hash)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	return user, nil
}

func (db *DB) GetEmailsByUserPk(pk int64) (out []*Email, err error) {

	rows, err := db.raw.Query(`SELECT emails.Pk, emails.userPk, emails.address emails.saltedHash 
	FROM emails WHERE emails.userPk = $1`, pk)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	defer rows.Close()

	for rows.Next() {
		email := &Email{}
		err = rows.Scan(email.Pk, &email.UserPk, &email.Address,
			&email.SaltedHash)
		if err != nil {
			return nil, errs.Wrap(err)
		}

		out = append(out, email)
	}

	if err := rows.Err(); err != nil {
		return nil, errs.Wrap(err)
	}

	return out, nil
}

func (db *DB) GetEmailByAddress(address string) (out *Email, err error) {
	err = db.raw.QueryRow(`SELECT emails.Pk, emails.userPk, emails.address emails.saltedHash 
	FROM emails WHERE emails.address = $1`, address).Scan(&out.Pk, &out.UserPk, &out.Address,
		&out.SaltedHash)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	return out, nil
}

//func (db *DB) CreateEmail()

//TODO(mac): make an email table that maps back to a user (multuple emails allowed)
//TODO(mac): make a salted password hash
//table emails should have user_pk, password hash, and a salt
