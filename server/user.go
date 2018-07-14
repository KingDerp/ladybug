package server

import (
	"context"
	"ladybug/database"
	"strings"

	uuid "github.com/satori/go.uuid"
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
	FullName  string   `json:"fullName"`
	FirstName string   `jsons:"firstName"`
	LastName  string   `jsons:"lastName"`
	Emails    []*Email `json:"emails"`
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
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Emails:    EmailsFromDB(emails),
	}
}

func (u *UserServer) GetUser(ctx context.Context, req *GetUserRequest) (
	resp *GetUserResponse, err error) {

	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		user, err := tx.Get_User_By_Pk(ctx, database.User_Pk(req.UserPk))
		if err != nil {
			return err
		}

		emails, err := tx.All_Email_By_UserPk(ctx, database.Email_UserPk(req.UserPk))
		if err != nil {
			return err
		}

		resp = &GetUserResponse{User: UserFromDB(user, emails)}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type UpdateUserRequest struct {
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
	Email           string `json:"email"`
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"password"`
}

type UpdateUserResponse struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

func (u *UserServer) UpdateUser(ctx context.Context, req *UpdateUserRequest) (
	resp *UpdateUserResponse, err error) {

	//TODO(mac): at what layer do I validate that these fields are not blank
	var email *database.Email
	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		email, err = tx.Get_Email_By_Address(ctx, database.Email_Address(req.Email))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	hash, err := hashPassword(req.CurrentPassword)
	if err != nil {
		return nil, err
	}

	if err := comparePasswordHash(hash, email.SaltedHash); err != nil {
		return nil, err
	}

	user_updates := database.User_Update_Fields{
		FirstName: database.User_FirstName(req.FirstName),
		LastName:  database.User_LastName(req.LastName),
	}

	email_updates := database.Email_Update_Fields{
		Address:    database.Email_Address(req.Email),
		SaltedHash: database.Email_SaltedHash(hash),
	}

	var user *database.User
	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		user, err = tx.Update_User_By_Pk(ctx, database.User_Pk(email.UserPk), user_updates)
		if err != nil {
			return err
		}

		email, err = tx.Update_Email_By_Pk(ctx, database.Email_Pk(email.Pk), email_updates)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &UpdateUserResponse{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     email.Address,
	}, nil
}

type LogInRequest struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

//TODO(mac): how can I check for the type of errors that this func returns so I'm not returning
//erros that might give more information about my server than I want.
func (u *UserServer) LogIn(ctx context.Context, req *LogInRequest) (resp *database.Session,
	err error) {

	var email *database.Email
	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		email, err = tx.Find_Email_By_Address(ctx, database.Email_Address(
			strings.ToLower(req.Email)))
		if err != nil {
			return err
		}

		if email == nil {
			return errs.New("No email exists with that address")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := comparePasswordHash(req.Password, email.SaltedHash); err != nil {
		return nil, err
	}

	var session *database.Session
	err = u.db.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {

		session, err = tx.Create_Session(ctx, database.Session_UserPk(email.UserPk),
			database.Session_Id(uuid.NewV4().String()))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return session, nil
}

//HashPassword takes a string, creates a hash using that string and returns the hash in a string
//format. One should note that GenerateFromPassword uses base64 encoding in it's logic.
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errs.Wrap(err)
	}
	return string(hash), err
	//return base64.URLEncoding.EncodeToString(hash), nil
}

//ComparePasswordHash - this funtion takes the unhashed version of the password and the hash to
//compare it against. returns an error if there was an internal problem or if the hashed and
//unhashed password do not match
func comparePasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return errs.New("email or password does not match")
	}
	return nil
}
