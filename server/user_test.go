package server

import (
	"testing"
)

func TestValidateUserIncompleteBillingAddress(t *testing.T) {
	sur := &SignUpRequest{
		FirstName:     "Joey",
		LastName:      "Fatone",
		Password:      "FatoneIsABoyBandIcon6^",
		Email:         "fatone@fatone.com",
		BillingStreet: "65 Florina ln.",
		BillingCity:   "Orlando",
		BillingState:  "FL",
		BillingZip:    98074,
	}

	err := validateSignUpRequest(sur)
	err = err

}
