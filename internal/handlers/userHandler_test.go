package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matryer/is"

	"gofit/internal/server"
	"gofit/internal/testutils"
)

func TestRegisterUser(t *testing.T) {
	is := is.New(t)

	testDB := testutils.TestDBSetup()

	makeRequest := func(body string) *httptest.ResponseRecorder {
		tx := testDB.Begin()
		defer tx.Rollback()

		server := server.NewAPIServer(tx)
		server.SetupRoutes()

		req, err := http.NewRequest(
			"POST",
			"/register",
			strings.NewReader(body),
		)
		is.NoErr(err)

		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)

		return rr
	}

	testCases := []struct {
		name     string
		body     string
		expected int
	}{
		{
			name:     "valid request",
			body:     `{"email": "some@test.com", "password": "password"}`,
			expected: http.StatusOK,
		},
		{
			name:     "invalid request no email",
			body:     `{"password": "password"}`,
			expected: http.StatusBadRequest,
		},
		{
			name:     "invalid request no password",
			body:     `{"email": "some@test.com"}`,
			expected: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rr := makeRequest(tc.body)
			is.Equal(rr.Code, tc.expected)
		})
	}
}
