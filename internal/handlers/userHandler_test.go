package handlers_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/matryer/is"

	"gofit/internal/models"
	"gofit/internal/server"
	"gofit/internal/testutils"
	"gofit/pkg/config"
)

func TestUserHandler_RegisterUser(t *testing.T) {
	is := is.New(t)

	testDB := testutils.TestDBSetup()
	tx := testDB.Begin()
	defer tx.Rollback()

	server := server.NewAPIServer(tx)
	server.SetupRoutes()

	email := "testRegisterUser@test.com"
	password := config.TestingPassword // strong password for validator

	// Test cases
	testCases := []struct {
		name     string
		req      UserCredentialsRequest
		expected int
	}{
		{
			name:     "valid request",
			req:      UserCredentialsRequest{Email: email, Password: password},
			expected: http.StatusOK,
		},
		{
			name:     "invalid request no email",
			req:      UserCredentialsRequest{Password: "password"},
			expected: http.StatusBadRequest,
		},
		{
			name:     "invalid request no password",
			req:      UserCredentialsRequest{Email: "some@test.com"},
			expected: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rr, err := testutils.MakeRequest(
				server.Router,
				"POST",
				"/register",
				tc.req,
			)
			is.NoErr(err)
			is.Equal(rr.Code, tc.expected)
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	is := is.New(t)

	testDB := testutils.TestDBSetup()
	tx := testDB.Begin()
	defer tx.Rollback()

	server := server.NewAPIServer(tx)
	server.SetupRoutes()

	// Register two test users directly to the DB
	email1 := "testUserHandlerLoginUser@test.com"
	password1 := config.TestingPassword // strong password for validator
	email2 := "SECONDARYtestUserHandlerLoginUser@test.com"
	password2 := "SECONDARY" + config.TestingPassword
	user1 := &models.User{
		Email:    email1,
		Password: password1,
	}
	user2 := &models.User{
		Email:    email2,
		Password: password2,
	}

	err := testutils.RegisterUser(tx, user1)
	is.NoErr(err)

	err = testutils.RegisterUser(tx, user2)
	is.NoErr(err)

	// Test cases
	testCases := []struct {
		name     string
		req      UserCredentialsRequest
		expected int
	}{
		{
			name:     "no email",
			req:      UserCredentialsRequest{Password: "password"},
			expected: http.StatusBadRequest,
		},
		{
			name:     "no password",
			req:      UserCredentialsRequest{Email: "some@test.com"},
			expected: http.StatusBadRequest,
		},
		{
			name:     "non-existent user",
			req:      UserCredentialsRequest{Email: "doesNotExist@test.com", Password: password1},
			expected: http.StatusBadRequest,
		},
		{
			name:     "incorrect password",
			req:      UserCredentialsRequest{Email: email1, Password: "thisIsNotThePassword"},
			expected: http.StatusBadRequest,
		},
		{
			name:     "existing password, mismatched existing user",
			req:      UserCredentialsRequest{Email: email1, Password: password2},
			expected: http.StatusBadRequest,
		},
		{
			name:     "valid request",
			req:      UserCredentialsRequest{Email: email1, Password: password1},
			expected: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rr, err := testutils.MakeRequest(
				server.Router,
				"POST",
				"/login",
				tc.req,
			)
			is.NoErr(err)
			is.Equal(rr.Code, tc.expected)

			// Check cookie is set on valid request, and is valid JWT
			if tc.name == "valid request" && tc.expected == http.StatusOK {
				cookies := rr.Result().Cookies()
				var jwtCookie *http.Cookie
				for _, cookie := range cookies {
					if cookie.Name == config.JwtCookieName {
						jwtCookie = cookie
						break
					}
				}
				is.True(jwtCookie != nil)

				tokenString := jwtCookie.Value

				parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
					return []byte(os.Getenv(config.JwtCookieName)), nil
				}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

				is.NoErr(err)
				is.True(parsedToken.Valid)
			}
		})
	}
}
