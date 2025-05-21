package models_test

import (
	"testing"

	"github.com/matryer/is"
	"golang.org/x/crypto/bcrypt"

	"godiscauth/internal/models"
	"godiscauth/internal/testutils"
	"godiscauth/pkg/apperrors"
)

func TestNewUser(t *testing.T) {
	is := is.New(t)

	// Valid email and password should return nil err, non-nil user value
	validEmail := "test@newuser.com"
	validPassword := config.TestingPassword
	t.Run("valid email and password", func(t *testing.T) {
		user, err := models.NewUser(validEmail, validPassword)
		is.True(user != nil)
		is.NoErr(err)

		// User value holds passed email and password
		is.Equal(user.Email, validEmail)
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(validPassword))
		is.NoErr(err)
	})

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
