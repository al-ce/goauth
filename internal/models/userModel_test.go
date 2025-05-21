package models_test

import (
	"testing"

	"github.com/matryer/is"
	"golang.org/x/crypto/bcrypt"

	"godiscauth/internal/models"
	"godiscauth/internal/testutils"
	"godiscauth/pkg/apperrors"
)

// TestNewUser tests new user creation in the `models` package.
func TestNewUser(t *testing.T) {
	is := is.New(t)

	// Valid email and password should return nil err, non-nil user value
	validEmail := "test@newuser.com"
	t.Run("valid email and password", func(t *testing.T) {
		user, err := models.NewUser(validEmail, testutils.TestingPassword)
		is.True(user != nil)
		is.NoErr(err)

		// User value holds passed email and password
		is.Equal(user.Email, validEmail)
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(testutils.TestingPassword))
		is.NoErr(err)

	})

	// Emails should not exceed 254 chars per RFC3696
	// Long emails should return error, nil user value
	t.Run("email exceeds 254 chars", func(t *testing.T) {
		exceeds254Chars := "example@"
		for range 256 {
			exceeds254Chars += "a"
		}
		exceeds254Chars += ".com"

		user, err := models.NewUser(exceeds254Chars, testutils.TestingPassword)
		is.Equal(user, nil)
		is.True(err == apperrors.ErrEmailMaxLength)
	})

	// Invalid email formats should return error, nil user value
	invalidFormats := map[string]string{
		"emptyString":     "",
		"missingAt":       "example.com",
		"multipleAts":     "at@at@at.com",
		"missingLocal":    "@at.com",
		"missingDomain":   "example@",
		"consecutiveDots": "example@at..com",
		"leadingDot":      ".example@at.com",
		"trailingDot":     "exmaple@at.com.",
		"space":           "example @at.com",
	}

	for name, email := range invalidFormats {
		t.Run(name, func(t *testing.T) {
			user, err := models.NewUser(email, testutils.TestingPassword)
			is.Equal(user, nil)
			is.True(err != nil)
		})
	}

	// Insufficiently complex passwords should return error, nil user value
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
