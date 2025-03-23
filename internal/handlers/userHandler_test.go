package handlers_test

import (
	"net/http"
	"testing"

	"github.com/matryer/is"

	"gofit/internal/server"
	"gofit/internal/testutils"
)

func TestUserHandler_RegisterUser(t *testing.T) {
	is := is.New(t)

	testDB := testutils.TestDBSetup()
	tx := testDB.Begin()
	defer tx.Rollback()

	server := server.NewAPIServer(tx)
	server.SetupRoutes()

	testCases := []struct {
		name     string
		req      Request
		expected int
	}{
		{
			name:     "valid request",
			req:      Request{Email: "some@test.com", Password: "password"},
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
