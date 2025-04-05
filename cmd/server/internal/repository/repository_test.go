package repository_test

import (
	"os"
	"testing"

	"gofit/internal/testutils"
)

func TestMain(m *testing.M) {
	testutils.TestEnvSetup()

	os.Exit(m.Run())
}
