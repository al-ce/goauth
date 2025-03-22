package initializers

import (
	"os"
	"testing"

	"github.com/matryer/is"
)

func TestMain(m *testing.M) {
	LoadEnvVariables()
	ConnectToDB()

	os.Exit(m.Run())
}

func TestConnectToDB(t *testing.T) {
	is := is.New(t)
	t.Run("connects", func(t *testing.T) {
		is.True(DB != nil)
	})
}
