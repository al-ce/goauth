package repository_test

import (
	"os"
	"testing"

	"github.com/al-ce/goauth/internal/testutils"
)

func TestMain(m *testing.M) {
	testutils.TestEnvSetup()

	os.Exit(m.Run())
}
