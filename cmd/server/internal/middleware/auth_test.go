package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/matryer/is"

	"gofit/internal/middleware"
	"gofit/internal/testutils"
	"gofit/pkg/config"
)

func TestMiddlewareAuth_RequireAuth(t *testing.T) {
	is := is.New(t)

	testDB := testutils.TestDBSetup()
	tx := testDB.Begin()
	defer tx.Rollback()

	authMw := middleware.NewAuthMiddleware(tx)

	router := gin.New()

	// Create test handler and route
	expectedResp := "test handler called"
	testHandler := func(c *gin.Context) {
		c.String(http.StatusOK, expectedResp)
	}

	router.GET("/protected", authMw.RequireAuth(), testHandler)

	// Generate a test token
	randUUID, err := uuid.NewRandom()
	is.NoErr(err)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": randUUID.String(),
		"exp": time.Now().Unix() + config.TokenExpiration,
	})
	tokenString, err := token.SignedString([]byte(os.Getenv(config.JwtCookieName)))
	is.NoErr(err)

	t.Run("with valid token", func(t *testing.T) {
		reqWithAuth, err := http.NewRequest("GET", "/protected", nil)
		is.NoErr(err)

		reqWithAuth.AddCookie(&http.Cookie{
			Name:  config.JwtCookieName,
			Value: tokenString,
		})
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, reqWithAuth)

		is.Equal(http.StatusOK, rr.Code)
		is.Equal(expectedResp, rr.Body.String())
	})

	t.Run("without token", func(t *testing.T) {
		reqNoAuth, err := http.NewRequest("GET", "/protected", nil)
		is.NoErr(err)

		rrNoAuth := httptest.NewRecorder()
		router.ServeHTTP(rrNoAuth, reqNoAuth)

		is.Equal(http.StatusUnauthorized, rrNoAuth.Code)
	})
}
