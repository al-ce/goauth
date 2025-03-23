package repository_test

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/matryer/is"
	"gorm.io/gorm"

	"gofit/internal/models"
	"gofit/internal/repository"
	"gofit/internal/testutils"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	testutils.TestEnvSetup()
	testDB = testutils.TestDBSetup()

	os.Exit(m.Run())
}

func TestCreateUser(t *testing.T) {
	is := is.New(t)

	t.Run("creates user", func(t *testing.T) {
		tx := testDB.Begin()
		defer tx.Rollback()

		ur := repository.NewUserRepository(tx)

		user := &models.User{
			Email:    "test@test.com",
			Password: "password",
		}
		err := ur.UserCreate(user)
		is.NoErr(err)

		var dbUser models.User
		tx.First(&dbUser, user.ID)

		is.True(dbUser.ID != uuid.UUID{})
		is.Equal(dbUser.Email, "test@test.com")
		is.Equal(dbUser.Password, "password")
		is.Equal(dbUser.LastLogin, nil)
		is.Equal(dbUser.FailedLoginAttempts, 0)
		is.True(!dbUser.AccountLocked)
		is.Equal(dbUser.AccountLockedUntil, nil)
	})
}
