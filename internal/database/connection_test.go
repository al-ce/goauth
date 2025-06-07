package database_test

import (
	"testing"

	"github.com/matryer/is"

	"github.com/al-ce/goauth/internal/database"
)

// TestConnectToDB tests the connection to the database.
func TestConnectToDB(t *testing.T) {
	is := is.New(t)
	t.Run("connects", func(t *testing.T) {
		testDB, err := database.NewDB()
		is.NoErr(err)
		is.True(testDB != nil)
	})
}
