package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/matryer/is"

	"github.com/al-ce/goauth/internal/server"
	"github.com/al-ce/goauth/internal/testutils"
)

// TestMain sets up the test environment for all tests in the `server_test` package.
func TestMain(m *testing.M) {
	testutils.TestEnvSetup()
	os.Exit(m.Run())
}

// TestPingRoute tests the `/ping` route of the API server.
func TestPingRoute(t *testing.T) {
	is := is.New(t)

	testutils.TestEnvSetup()

	testDB := testutils.TestDBSetup()
	server, err := server.NewAPIServer(testDB)
	is.NoErr(err)
	server.SetupRoutes()

	req, _ := http.NewRequest("GET", "/ping", nil)
	rr := httptest.NewRecorder()
	server.Router.ServeHTTP(rr, req)
	var response map[string]string
	err = json.NewDecoder(rr.Body).Decode(&response)

	is.Equal(http.StatusOK, rr.Code)
	is.Equal(response["message"], "pong")
}
