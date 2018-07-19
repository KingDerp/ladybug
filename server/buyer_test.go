package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateBuyerIncompleteBillingAddress(t *testing.T) {
	sur := &SignUpRequest{
		FirstName: "Joey",
		LastName:  "Fatone",
		Password:  "FatoneIsABoyBandIcon6^",
		Email:     "fatone@fatone.com",
		BillingAddress: &Address{
			StreetAddress: "65 Florina ln.",
			City:          "Orlando",
			State:         "FL",
			Zip:           98074,
		},
	}

	err := validateSignUpRequest(sur)
	require.NoError(t, err)
}

func TestCompareHash(t *testing.T) {
	//paswords match
	password := "Password1!"
	hash, err := hashPassword(password)
	require.NoError(t, err)

	err = comparePasswordHash(password, hash)
	require.NoError(t, err)

	//passwords do not match
	password = "Password2!"
	hash, err = hashPassword("password3@")
	require.NoError(t, err)

	err = comparePasswordHash(password, hash)
	require.EqualError(t, err, "email or password does not match")
}
