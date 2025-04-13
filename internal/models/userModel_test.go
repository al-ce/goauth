package models_test

import (
	"testing"

	"github.com/matryer/is"

	"goauth/pkg/config"
	"goauth/internal/models"
)

func TestNewUser(t *testing.T) {
	is := is.New(t)

	validEmail := "test@newuser.com"
	validPassword := config.TestingPassword

	_, err := models.NewUser(validEmail, validPassword)
	is.NoErr(err)

	exceeds254 := "example@"
	for range 256 {
		exceeds254 += "a"
	}
	exceeds254 += ".com"

	invalidEmails := map[string]string{
		"emptyString":     "",
		"exceeds254":      exceeds254,
		"missingAt":       "example.com",
		"multipleAts":     "at@at@at.com",
		"missingLocal":    "@at.com",
		"missingDomain":   "example@",
		"consecutiveDots": "example@at..com",
		"leadingDot":      ".example@at.com",
		"trailingDot":     "exmaple@at.com.",
		"space":           "example @at.com",
	}

	for name, email := range invalidEmails {
		t.Run(name, func(t *testing.T) {
			_, err := models.NewUser(email, validPassword)
			is.True(err != nil)
		})
	}

	invalidPasswords := map[string]string{
		"emptyString": "",
		"tooShort":    "short",
		"sequential":  "12345678",
		"repeating":   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}

	for name, password := range invalidPasswords {
		t.Run(name, func(t *testing.T) {
			_, err := models.NewUser(validEmail, password)
			is.True(err != nil)
		})
	}
}
