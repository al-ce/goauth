package handlers_test

import (
	"io"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"gofit/internal/testutils"
)

func TestMain(m *testing.M) {
	testutils.TestEnvSetup()

	gin.SetMode(gin.TestMode)
	gin.DefaultWriter = io.Discard
	os.Exit(m.Run())
}
