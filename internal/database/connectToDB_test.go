package database

import (
	"os"
	"testing"

	"github.com/matryer/is"

	"gofit/internal/config"
)

func TestMain(m *testing.M) {
	config.LoadEnvVariables()
	ConnectToDB()

	os.Exit(m.Run())
}

func TestConnectToDB(t *testing.T) {
	is := is.New(t)
	t.Run("connects", func(t *testing.T) {
		is.True(DB != nil)
	})
}
