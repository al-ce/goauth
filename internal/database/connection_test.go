package database_test

import (
	"testing"

	"github.com/matryer/is"

	"gofit/internal/testutils"
)

func TestConnectToDB(t *testing.T) {
	is := is.New(t)
	testDB := testutils.TestDBSetup()
	t.Run("connects", func(t *testing.T) {
		is.True(testDB != nil)
	})
}
