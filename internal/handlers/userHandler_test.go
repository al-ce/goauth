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
	server := server.NewAPIServer(testDB)
	server.Run()

	t.Run("valid request", func(t *testing.T) {
		req, err := http.NewRequest(
			"POST",
			"/register",
			strings.NewReader(`{"email": "some@test.com", "password": "password"}`),
		)
		is.NoErr(err)

		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)

		is.Equal(rr.Code, http.StatusOK)
	})

	t.Run("invalid request no email", func(t *testing.T) {
		req, err := http.NewRequest(
			"POST",
			"/register",
			strings.NewReader(`{"password": "password"}`),
		)
		is.NoErr(err)

		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)

		is.Equal(rr.Code, http.StatusBadRequest)
	})

	t.Run("invalid request no password", func(t *testing.T) {
		req, err := http.NewRequest(
			"POST",
			"/register",
			strings.NewReader(`{"email": "some@test.com"}`),
		)
		is.NoErr(err)

		rr := httptest.NewRecorder()
		server.Router.ServeHTTP(rr, req)

		is.Equal(rr.Code, http.StatusBadRequest)
	})
}
