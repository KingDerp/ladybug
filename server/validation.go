package server

import (
	"fmt"
	"unicode"

	"github.com/zeebo/errs"
)

type validateFunc func(r rune) bool

type passwordPolicy struct {
	description string
	validate    func(string) bool
}

//TODO(mac) all of these functions need testing
const (
	minPasswordLen = 8
	maxPasswordLen = 50
)

var (
	specialChars     = []rune(`~!@#$%^&*()><./,*-+;:`)
	passwordPolicies = []passwordPolicy{
		{
			description: "Password must contain an upper case letter",
			validate:    checkForRune(unicode.IsUpper),
		},
		{
			description: "Password must contain a lower case letter",
			validate:    checkForRune(unicode.IsLower),
		},
		{
			description: "Password must contain a number",
			validate:    checkForRune(unicode.IsNumber),
		},
		{
			description: fmt.Sprintf("Password must contain a special character which includes: %s",
				specialChars),
			validate: checkForRune(func(r rune) bool { return runeIn(r, specialChars) }),
		},
		{
			description: fmt.Sprintf("Password must be a maximum of %d characters", maxPasswordLen),
			validate:    checkMaxPasswordPolicy,
		},
		{
			description: fmt.Sprintf("Password must be a minimum of %d characters", minPasswordLen),
			validate:    checkMinPasswordPolicy,
		},
	}
)

func validateUser(user *SignUpRequest) error {

	//validate user name
	if err := validateUserName(user.FullName); err != nil {
		return err
	}

	//validate password
	if err := validateUserPassword(user.Password); err != nil {
		return err
	}

	return nil
}

func checkMinPasswordPolicy(pw string) bool {
	return len(pw) >= minPasswordLen
}

func checkMaxPasswordPolicy(pw string) bool {
	return len(pw) <= maxPasswordLen
}

func validateUserPassword(pw string) error {
	for _, p := range passwordPolicies {
		if !p.validate(pw) {
			return errs.New(p.description)
		}
	}
	return nil
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

func checkForRune(fn func(r rune) bool) func(s string) bool {
	return func(s string) bool {
		for _, r := range s {
			if fn(r) {
				return true
			}
		}
		return false
	}
}

func runeIn(r rune, in []rune) bool {
	for _, candidate := range in {
		if candidate == r {
			return true
		}
	}
	return false
}
