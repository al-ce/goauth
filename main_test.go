package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/matryer/is"
)

var router *gin.Engine

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	gin.DefaultWriter = io.Discard
	router = setupRouter()
	os.Exit(m.Run())
}

func TestPingRoute(t *testing.T) {
	is := is.New(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	is.Equal(http.StatusOK, w.Code)
	is.Equal("pong", w.Body.String())
}
