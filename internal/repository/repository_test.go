package repository_test

import (
	"os"
	"testing"

	"godiscauth/internal/testutils"
)

func TestMain(m *testing.M) {
	testutils.TestEnvSetup()

	os.Exit(m.Run())
}
