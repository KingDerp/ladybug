package server

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zeebo/errs"
)

func TestEmailValidation(t *testing.T) {

	email_tests := []struct {
		input    string
		expected error
	}{
		{"faulty_email", errs.New("%s is not a valid email address", "faulty_email")},
		{"faulty_email@test", errs.New("%s is not a valid email address", "faulty_email@test")},
		{"@test", errs.New("%s is not a valid email address", "@test")},
		{"@test.email", errs.New("%s is not a valid email address", "@test.email")},
		{"valid@test.email", nil},
		{"valid@test.marketing", nil},
		{"1@test.marketing", nil},
	}

	for _, tt := range email_tests {
		actual := validateEmail(tt.input)

		if tt.expected == nil {
			require.Equal(t, actual, tt.expected,
				fmt.Sprintf("validateEmail(%s): expected %#v, actual %#v",
					tt.input, tt.expected, actual),
			)
		} else {
			require.EqualError(t, actual, tt.expected.Error(),
				fmt.Sprintf("validateEmail(%s): expected %t, actual %t",
					tt.input, tt.expected, actual),
			)
		}
	}
}

func TestValidateName(t *testing.T) {
	name_tests := []struct {
		input    string
		expected error
	}{
		{"longer_than_50_characters_shouldn't_be_allowed_12345",
			errs.New("name cannot exceed 50 characters"),
		},
		{"", errs.New("name must not be empty")},
		{"this_should_be_a_valid_name", nil},
		{"Steve1", nil},
		{"Jobe", nil},
		{"123456", nil},
	}

	for _, tt := range name_tests {
		actual := validateName(tt.input)

		if tt.expected == nil {
			require.Equal(t, actual, tt.expected,
				fmt.Sprintf("validateName(%s): expected %#v, actual %#v",
					tt.input, tt.expected, actual),
			)
		} else {
			require.EqualError(t, actual, tt.expected.Error(),
				fmt.Sprintf("validateName(%s): expected %t, actual %t",
					tt.input, tt.expected, actual),
			)
		}
	}
}

func TestValidateUserPassword(t *testing.T) {
	password_tests := []struct {
		input    string
		expected error
	}{
		{"no_upper_case_letter", errs.New("Password must contain an upper case letter")},
		{"NO_LOWER_CASE_LETTER", errs.New("Password must contain a lower case letter")},
		{"HAS_no_Number", errs.New("Password must contain a number")},
		{"HAS_no_special_character_1", errs.New(fmt.Sprintf("Password must contain a special"+
			" character which includes: %s", specialChars))},
		{`HAS_invalid_char1}`, errs.New(fmt.Sprintf("Password must contain a special"+
			" character which includes: %s", specialChars))},
		{"$1Exceeds_max_count_length_of_50_characters_and_is_not_allowed",
			errs.New(fmt.Sprintf("Password must be a maximum of %d characters", maxPasswordLen))},
		{"$1Short",
			errs.New(fmt.Sprintf("Password must be a minimum of %d characters", minPasswordLen))},
		{"ValidPassword%!3", nil},
		{"LadybugRocks888)8", nil},
		{"ValidPassword%!3", nil},
		{"$5^ValidForSure", nil},
	}

	for _, tt := range password_tests {
		actual := validateUserPassword(tt.input)

		if tt.expected == nil {
			require.Equal(t, actual, tt.expected,
				fmt.Sprintf("validateUserPassword(%s): expected %#v, actual %#v",
					tt.input, tt.expected, actual),
			)
		} else {
			require.EqualError(t, actual, tt.expected.Error(),
				fmt.Sprintf("validateUserPassword(%s): expected %t, actual %t",
					tt.input, tt.expected, actual),
			)
		}
	}
}

func TestValidateIncompleteBillingAddress(t *testing.T) {
	sur := &SignUpRequest{
		BillingStreet: "",
		BillingCity:   "Orlando",
		BillingState:  "FL",
		BillingZip:    98074,
	}

	actual := validateBillingAddress(sur)
	expected := errs.New("city, state, or street fields are blank for billing address")
	require.EqualError(t, actual, expected.Error())
}

func TestValidateBillingZipIsZero(t *testing.T) {
	sur := &SignUpRequest{
		BillingStreet: "florida ln",
		BillingCity:   "Orlando",
		BillingState:  "FL",
		BillingZip:    0,
	}

	actual := validateBillingAddress(sur)
	expected := errs.New("you must provide a billing zip code")
	require.EqualError(t, actual, expected.Error())
}

func TestValidateBillingStreetMissing(t *testing.T) {
	sur := &SignUpRequest{
		BillingStreet: "",
		BillingCity:   "Orlando",
		BillingState:  "FL",
		BillingZip:    0,
	}

	actual := validateBillingAddress(sur)
	expected := errs.New("city, state, or street fields are blank for billing address")
	require.EqualError(t, actual, expected.Error())
}

func TestShippingAddressIsEmpty(t *testing.T) {
	actual := shippingAddressIsEmpty(&SignUpRequest{})
	expected := true
	require.Equal(t, actual, expected,
		fmt.Sprintf("shippingAddress(sur)  actual:%t expected:%t", actual, expected))
}

func TestShippingAddressIsNotEmpty(t *testing.T) {
	sur := &SignUpRequest{ShippingStreet: "not empty"}

	actual := shippingAddressIsEmpty(sur)
	expected := false
	require.Equal(t, actual, expected,
		fmt.Sprintf("shippingAddress(sur)  actual:%t expected:%t", actual, expected))
}
