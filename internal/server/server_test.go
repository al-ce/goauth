package server_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/matryer/is"

	"gofit/internal/server"
	"gofit/internal/testutils"
)

func TestMain(m *testing.M) {
	testutils.TestEnvSetup()

	gin.SetMode(gin.TestMode)
	gin.DefaultWriter = io.Discard
	os.Exit(m.Run())
}

func TestPingRoute(t *testing.T) {
	is := is.New(t)

	testutils.TestEnvSetup()
	testDB := testutils.TestDBSetup()

	server := server.NewAPIServer(testDB)
	server.Run()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)

	server.Router.ServeHTTP(w, req)


	is.Equal(http.StatusOK, w.Code)
	is.Equal("pong", w.Body.String())
}
