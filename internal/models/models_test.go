package models_test

import (
	"os"
	"testing"

	"github.com/al-ce/goauth/internal/testutils"
)

// TestMain sets up the test environment for all tests in the `models_test` package.
func TestMain(m *testing.M) {
	testutils.TestEnvSetup()

	os.Exit(m.Run())
}
