package handlers_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/matryer/is"

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

	validEmail := "testRegisterUser@test.com"
	validPassword := "correcthorsebatterystaple" // strong password for validator

	// Test cases
	testCases := []struct {
		name     string
		req      Request
		expected int
	}{
		{
			name:     "valid request",
			req:      Request{Email: validEmail, Password: validPassword},
			expected: http.StatusOK,
		},
		{
			name:     "invalid request no email",
			req:      Request{Password: "password"},
			expected: http.StatusBadRequest,
		},
		{
			name:     "invalid request no password",
			req:      Request{Email: "some@test.com"},
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

	validEmail := "testUserHandlerLoginUser@test.com"
	validPassword := "correcthorsebatterystaple" // strong password for validator
	validEmail2 := "SECONDARYtestUserHandlerLoginUser@test.com"
	validPassword2 := "SECONDARYcorrecthorsebatterystaple"

	// Register first user
	rr, err := testutils.MakeRequest(
		server.Router,
		"POST",
		"/register",
		Request{Email: validEmail, Password: validPassword},
	)
	is.NoErr(err)
	is.Equal(rr.Code, http.StatusOK)

	// Register second user
	rr, err = testutils.MakeRequest(
		server.Router,
		"POST",
		"/register",
		Request{Email: validEmail2, Password: validPassword2},
	)
	is.NoErr(err)
	is.Equal(rr.Code, http.StatusOK)

	// Test cases
	testCases := []struct {
		name     string
		req      Request
		expected int
	}{
		{
			name:     "no email",
			req:      Request{Password: "password"},
			expected: http.StatusBadRequest,
		},
		{
			name:     "no password",
			req:      Request{Email: "some@test.com"},
			expected: http.StatusBadRequest,
		},
		{
			name:     "non-existent user",
			req:      Request{Email: "doesNotExist@test.com", Password: validPassword},
			expected: http.StatusBadRequest,
		},
		{
			name:     "incorrect password",
			req:      Request{Email: validEmail, Password: "thisIsNotThePassword"},
			expected: http.StatusBadRequest,
		},
		{
			name:     "existing password, mismatched existing user",
			req:      Request{Email: validEmail, Password: validPassword2},
			expected: http.StatusBadRequest,
		},
		{
			name:     "valid request",
			req:      Request{Email: validEmail, Password: validPassword},
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
