package repository_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/matryer/is"

	"gofit/internal/models"
	"gofit/internal/repository"
	"gofit/internal/testutils"
)

func TestUserRepository_RegisterUser(t *testing.T) {
	testDB := testutils.TestDBSetup()
	is := is.New(t)

	t.Run("creates user", func(t *testing.T) {
		tx := testDB.Begin()
		defer tx.Rollback()

		ur := repository.NewUserRepository(tx)

		user := &models.User{
			Email:    "test@test.com",
			Password: "password",
		}
		err := ur.RegisterUser(user)
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
