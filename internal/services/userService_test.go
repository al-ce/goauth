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

	us := newTestUserService()

	t.Run("empty email", func(t *testing.T) {
		err := us.RegisterUser("", "password")
		is.Equal(err, apperrors.ErrEmailIsEmpty)
	})

	t.Run("empty password", func(t *testing.T) {
		err := us.RegisterUser("some@test.com", "")
		is.Equal(err, apperrors.ErrPasswordIsEmpty)
	})
}

func TestUserService_LoginUser(t *testing.T) {
	is := is.New(t)

	us := newTestUserService()

	t.Run("non existing user", func(t *testing.T) {
		_, err := us.LoginUser("doesNotExist@test.com", "password")
		is.Equal(err, apperrors.ErrUserNotFound)
	})

	t.Run("empty email", func(t *testing.T) {
		_, err := us.LoginUser("", "password")
		is.Equal(err, apperrors.ErrEmailIsEmpty)
	})

	t.Run("empty password", func(t *testing.T) {
		_, err := us.LoginUser("some@test.com", "")
		is.Equal(err, apperrors.ErrPasswordIsEmpty)
	})
}

func newTestUserService() *services.UserService {
	testDB := testutils.TestDBSetup()
	tx := testDB.Begin()
	defer tx.Rollback()

	userRepo := repository.NewUserRepository(tx)
	return services.NewUserService(userRepo)
}
