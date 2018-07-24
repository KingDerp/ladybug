package validate

import (
	"fmt"
	"regexp"
	"unicode"

	"github.com/zeebo/errs"
)

type Address struct {
	StreetAddress string `json:"streetAddress"`
	City          string `json:"city"`
	State         string `json:"state"`
	Zip           int    `json"zip"`
}

type validateFunc func(r rune) bool

type passwordPolicy struct {
	description string
	validate    func(string) bool
}

const (
	MinPasswordLen = 8
	MaxPasswordLen = 50
)

var (
	SpecialChars     = []rune(`~!@#$%^&*()><./,*-+;:`)
	passwordPolicies = []passwordPolicy{
		{
			description: "Password must not be empty",
			validate:    lenNotZero,
		},
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
				SpecialChars),
			validate: checkForRune(func(r rune) bool { return runeIn(r, SpecialChars) }),
		},
		{
			description: fmt.Sprintf("Password must be a maximum of %d characters", MaxPasswordLen),
			validate:    checkMaxPasswordPolicy,
		},
		{
			description: fmt.Sprintf("Password must be a minimum of %d characters", MinPasswordLen),
			validate:    checkMinPasswordPolicy,
		},
	}

	//taken from https://stackoverflow.com/questions/201323/how-to-validate-an-email-address-using-a-regular-expression
	emailRegex = regexp.MustCompile(`(?:[a-z0-9!#$%&'*+/=?^_{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])`)
)

func checkMinPasswordPolicy(pw string) bool {
	return len(pw) >= MinPasswordLen
}

func checkMaxPasswordPolicy(pw string) bool {
	return len(pw) <= MaxPasswordLen
}

func AddressIsEmpty(a *Address) bool {
	if a == nil {
		return true
	}

	if a.StreetAddress == "" &&
		a.City == "" &&
		a.State == "" &&
		a.Zip == 0 {
		return true
	}

	return false
}

func CheckAddress(a *Address) error {
	if a == nil {
		return errs.New("no address was submitted")
	}

	if a.StreetAddress == "" ||
		a.City == "" ||
		a.State == "" ||
		a.Zip == 0 {
		return errs.New("city, state, street, or zip fields are blank for billing address")
	}

	return nil
}

func CheckEmail(email string) error {
	if len(email) <= 0 {
		return errs.New("email address cannot be empty")
	}

	if emailRegex.MatchString(email) {
		return nil
	}

	return errs.New("%s is not a valid email address", email)
}

func CheckPassword(pw string) error {
	for _, p := range passwordPolicies {
		if !p.validate(pw) {
			return errs.New(p.description)
		}
	}
	return nil
}

func CheckFullName(first_name, last_name string) error {
	if err := checkName(first_name); err != nil {
		return err
	}

	if err := checkName(last_name); err != nil {
		return err
	}

	return nil
}

func checkName(name string) error {
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

func lenNotZero(s string) bool {
	return len(s) > 0
}
