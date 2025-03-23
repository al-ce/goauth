package database_test

import (
	"os"
	"testing"

	"github.com/matryer/is"
	"gorm.io/gorm"

	"gofit/internal/testutils"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	testutils.TestEnvSetup()
	testDB = testutils.TestDBSetup()
	os.Exit(m.Run())
}

func TestConnectToDB(t *testing.T) {
	is := is.New(t)
	t.Run("connects", func(t *testing.T) {
		is.True(testDB != nil)
	})
}
