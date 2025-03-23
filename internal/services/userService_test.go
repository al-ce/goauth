package services_test

import (
	"testing"

	"github.com/matryer/is"

	"gofit/internal/repository"
	"gofit/internal/services"
	"gofit/internal/testutils"
	"gofit/pkg/apperrors"
)

func TestUserService_RegisterUser(t *testing.T) {
	is := is.New(t)

	testDB := testutils.TestDBSetup()
	tx := testDB.Begin()
	defer tx.Rollback()

	userRepo := repository.NewUserRepository(tx)
	us := services.NewUserService(userRepo)

	t.Run("empty email", func(t *testing.T) {
		err := us.RegisterUser("", "password")
		is.Equal(err, apperrors.ErrEmailIsEmpty)
	})

	t.Run("empty password", func(t *testing.T) {
		err := us.RegisterUser("some@test.com", "")
		is.Equal(err, apperrors.ErrPasswordIsEmpty)
	})
}
