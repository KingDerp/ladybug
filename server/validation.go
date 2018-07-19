package server

import (
	"fmt"
	"regexp"
	"unicode"

	"github.com/zeebo/errs"
)

type validateFunc func(r rune) bool

type passwordPolicy struct {
	description string
	validate    func(string) bool
}

//TODO(mac) this seems like it should be it's own package
const (
	minPasswordLen       = 8
	maxPasswordLen       = 50
	maxExecutiveContacts = 12
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

	//taken from https://stackoverflow.com/questions/201323/how-to-validate-an-email-address-using-a-regular-expression
	emailRegex = regexp.MustCompile(`(?:[a-z0-9!#$%&'*+/=?^_{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])`)
)

//validateVendorSignUpRequest verifies the required data for regisering as a vendor. returns an
//error is there is a problem or incoming data does not match the requirement
func validateVendorSignUpRequest(vsr *VendorSignUpRequest) error {

	if len(vsr.ExecutiveContacts) > maxExecutiveContacts {
		return errs.New("only a max of %d contacts are allowed", maxExecutiveContacts)
	}

	for _, e := range vsr.ExecutiveContacts {
		if err := validateFullName(e.FirstName, e.LastName); err != nil {
			return err
		}

		if err := validatePassword(e.Password); err != nil {
			return err
		}

		//TODO(mac): for production I'll have to figure out how to send email validations with links
		//etc.
		if err := validateEmail(e.Email); err != nil {
			return err
		}
	}

	//TODO(mac): eventually I'll have to validate addresses as well
	if err := validateAddress(vsr.BillingAddress); err != nil {
		return err
	}

	if !addressIsEmpty(vsr.ShippingAddress) {
		if err := validateAddress(vsr.ShippingAddress); err != nil {
			return err
		}
	}

	return nil

}

//validateSignUpRequest is an internal function that is used to verify only the required data in a sign up
//request. For example a billing address is required a shipping address is not.
func validateSignUpRequest(sur *SignUpRequest) error {

	//validate buyer name
	if err := validateFullName(sur.FirstName, sur.LastName); err != nil {
		return err
	}

	//validate password
	if err := validatePassword(sur.Password); err != nil {
		return err
	}

	//validate email
	//TODO(mac): for production I'll have to figure out how to send email validations with links
	//etc.
	if err := validateEmail(sur.Email); err != nil {
		return err
	}

	//TODO(mac): eventually I'll have to validate addresses as well
	if err := validateAddress(sur.BillingAddress); err != nil {
		return err
	}

	if !addressIsEmpty(sur.ShippingAddress) {
		if err := validateAddress(sur.ShippingAddress); err != nil {
			return err
		}
	}

	return nil
}

func checkMinPasswordPolicy(pw string) bool {
	return len(pw) >= minPasswordLen
}

func checkMaxPasswordPolicy(pw string) bool {
	return len(pw) <= maxPasswordLen
}

func addressIsEmpty(a *Address) bool {
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

//for now the billing address fields just cannot be empty.
//TODO(mac): research adress validator services and use them here
func validateAddress(a *Address) error {
	if a == nil {
		return errs.New("no address was submitted")
	}

	if a.StreetAddress == "" ||
		a.City == "" ||
		a.State == "" {
		return errs.New("city, state, or street fields are blank for billing address")
	}

	if a.Zip == 0 {
		return errs.New("you must provide a billing zip code")
	}

	return nil
}

func validateEmail(email string) error {
	if emailRegex.MatchString(email) {
		return nil
	}

	return errs.New("%s is not a valid email address", email)
}

func validatePassword(pw string) error {
	for _, p := range passwordPolicies {
		if !p.validate(pw) {
			return errs.New(p.description)
		}
	}
	return nil
}

func validateFullName(first_name, last_name string) error {
	if err := validateName(first_name); err != nil {
		return err
	}

	if err := validateName(last_name); err != nil {
		return err
	}

	return nil
}

func validateName(name string) error {
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
