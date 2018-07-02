package server

import (
	"context"
	"encoding/base64"
	"ladybug/database"

	"github.com/zeebo/errs"
	"golang.org/x/crypto/bcrypt"
)

type UserServer struct {
	db *database.DB
}

func NewUserServer(db *database.DB) *UserServer {
	return &UserServer{db: db}
}

type GetUserRequest struct {
	UserPk int64
}

type GetUserResponse struct {
	User *User
}

type Email struct {
	Address string
}

type User struct {
	FullName string   `json:"fullName"`
	Emails   []*Email `json:"emails"`
}

func EmailFromDB(email *database.Email) *Email {
	return &Email{
		Address: email.Address,
	}
}

func EmailsFromDB(emails []*database.Email) []*Email {
	out := []*Email{}
	for _, email := range emails {
		out = append(out, EmailFromDB(email))
	}
	return out

}

func UserFromDB(user *database.User, emails []*database.Email) *User {
	return &User{
		FullName: user.FullName,
		Emails:   EmailsFromDB(emails),
	}
}

func (u *UserServer) GetUser(ctx context.Context, req *GetUserRequest) (
	resp *GetUserResponse, err error) {

	user, err := u.db.GetUserByPk(ctx, req.UserPk)
	if err != nil {
		return nil, err
	}

	emails, err := u.db.GetEmailsByUserPk(req.UserPk)
	if err != nil {
		return nil, err
	}

	return &GetUserResponse{User: UserFromDB(user, emails)}, nil
}

type UpdateUserRequest struct {
	FullName        string `json:"fullName"`
	Email           string `json:"email"`
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"password"`
}

type UpdateUserResponse struct {
	FullName string `json:"fullName"`
	Email    string `json:"email"`
}

func (u *UserServer) UpdateUser(ctx context.Context, req *UpdateUserRequest) (
	resp *UpdateUserResponse, err error) {

	//TODO(mac): at what layer do I validate that these fields are not blank
	email, err := u.db.GetEmailByAddress(req.Email)
	if err != nil {
		return nil, err
	}

	hash, err := hashPassword(req.CurrentPassword)
	if err != nil {
		return nil, err
	}

	if err := compareHash(hash, email.SaltedHash); err != nil {
		return nil, err
	}

	user_updates := database.UserUpdateFields{Pk: email.UserPk, FullName: req.FullName}

	//TODO(mac): these two code blocks seems like a really good place for a transaction
	if err := u.db.UpdateUser(ctx, &user_updates); err != nil {
		return nil, err
	}

	email_updates := database.EmailUpdateFields{Pk: email.Pk, Email: email.Address, SaltedHash: hash}
	if err := u.db.UpdateEmail(ctx, &email_updates); err != nil {
		return nil, err
	}

	return &UpdateUserResponse{FullName: req.FullName, Email: email.Address}, nil
}

type SignUpRequest struct {
	FullName string `json:"fullName"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type SignUpResponse struct {
	Session *database.Session
}

func (u *UserServer) SignUp(ctx context.Context, req *SignUpRequest) (resp *SignUpResponse,
	err error) {

	//TODO(mac): need to validate email as well
	err = validateUser(req)
	if err != nil {
		return nil, err
	}

	hash, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user, err := u.db.CreateUserWithEmail(ctx, req.FullName, req.Email, hash)
	if err != nil {
		return nil, err
	}

	session, err := u.db.CreateSession(ctx, user.Pk)
	if err != nil {
		return nil, err
	}

	return &SignUpResponse{Session: session}, nil
}

type LogInRequest struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

//TODO(mac): how can I check for the type of errors that this func returns so I'm not returning
//erros that might give more information about my server than I want.
func (u *UserServer) LogIn(ctx context.Context, req *LogInRequest) (resp *database.Session,
	err error) {

	email, err := u.db.GetEmailByAddress(req.Email)
	if err != nil {
		return nil, err
	}

	hash, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	if err := compareHash(hash, email.SaltedHash); err != nil {
		return nil, err
	}

	if req.Email != email.Address {
		return nil, errs.New("password or email do not match")
	}

	session, err := u.db.CreateSession(ctx, email.UserPk)
	if err != nil {
		return nil, err
	}

	return session, nil
}

//TODO(mac): look at vim ctrl-p extentions for fuzzy file search
func hashPassword(pw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return "", errs.Wrap(err)
	}

	return base64.URLEncoding.EncodeToString(hash), nil
}

//TODO(mac): test bcrypt password compnrison to make sure that it works
func compareHash(a, b string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(a), []byte(b)); err != nil {
		return errs.New("email or password does not match")
	}
	return nil
}
