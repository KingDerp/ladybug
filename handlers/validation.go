package handlers

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/zeebo/errs"
)

type validateFunc func(r rune) bool

type passwordValidation struct {
	isValid    bool
	errMessage string
	validate   func(r rune) bool
}

//TODO(mac) all of these functions need testing
const (
	minPasswordLen = 8
)

var (
	specialChars = []string{"~", "!", "@", "#", "$", "%", "^", "&", "*", "(", ")", ">", "<", ".",
		"/", ",", "*", "-", "+", ";", ":"}
)

func initializeUserValidations() []passwordValidation {
	return []passwordValidation{
		passwordValidation{
			isValid:    false,
			errMessage: "Password must contain an upper case letter",
			validate:   unicode.IsUpper,
		},
		passwordValidation{
			isValid:    false,
			errMessage: "Password must contain a lower case letter",
			validate:   unicode.IsLower,
		},
		passwordValidation{
			isValid:    false,
			errMessage: "Password must contain a number",
			validate:   unicode.IsNumber,
		},
		passwordValidation{
			isValid: false,
			errMessage: fmt.Sprintf("Password must contain a special character which inludes: %s",
				strings.Join(specialChars, ", ")),
			validate: isSpecialChar,
		},
	}
}

func validateUser(user *incomingUser) error {

	//validate user name
	if err := validateUserName(user.fullName); err != nil {
		return err
	}

	//validate password
	if err := validateUserPassword(user.password); err != nil {
		return err
	}

	return nil
}

func validateUserPassword(pw string) error {

	//check the overall password before we get into the details
	if len(pw) < minPasswordLen {
		return errs.New("password must be at least 8 characters")
	}

	if len(pw) > 50 {
		return errs.New("password cannot exceed 50 characters")
	}

	//iterate through characters in string
	//match cases in userValidations
	iuv := initializeUserValidations()
	for _, r := range pw {
		for _, v := range iuv {
			v.isValid = v.validate(r)
		}
	}

	return hasAnyFalse(iuv)
}

func validateUserName(name string) error {
	switch {
	case name == "":
		return errs.New("name must not be empty")
	case len(name) > 50:
		return errs.New("name cannot exceed 50 characters")
	default:
		return nil
	}
}

func isSpecialChar(r rune) bool {
	special_chars_str := strings.Join(specialChars, "")
	for _, sc := range special_chars_str {
		if r == sc {
			return true
		}
	}
	return false
}

func hasAnyFalse(uv []passwordValidation) error {
	for _, v := range uv {
		if v.isValid == false {
			return errs.New(v.errMessage)
		}
	}
	return nil
}
