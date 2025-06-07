package handlers_test

import (
	"io"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/al-ce/goauth/internal/testutils"
)

type UserCredentialsRequest struct {
	Email string
	Password string
}

func TestMain(m *testing.M) {
	testutils.TestEnvSetup()

	gin.SetMode(gin.TestMode)
	gin.DefaultWriter = io.Discard
	os.Exit(m.Run())
}
