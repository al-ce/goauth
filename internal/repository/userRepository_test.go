package repository_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/matryer/is"

	"godiscauth/internal/models"
	"godiscauth/internal/repository"
	"godiscauth/internal/testutils"
	"godiscauth/pkg/apperrors"
	"godiscauth/pkg/config"
)

func TestUserRepository_RegisterUser(t *testing.T) {
	testDB := testutils.TestDBSetup()
	is := is.New(t)
	tx := testDB.Begin()
	defer tx.Rollback()

	ur := repository.NewUserRepository(tx)

	t.Run("fails on nil user", func(t *testing.T) {
		err := ur.RegisterUser(nil)
		is.Equal(err, apperrors.ErrUserIsNil)
	})

	t.Run("fails on missing email", func(t *testing.T) {
		user := &models.User{
			Password: "password",
		}
		err := ur.RegisterUser(user)
		is.Equal(err, apperrors.ErrEmailIsEmpty)
	})

	t.Run("fails on missing password", func(t *testing.T) {
		user := &models.User{
			Email: "testRegisterUser@test.com",
		}
		err := ur.RegisterUser(user)
		is.Equal(err, apperrors.ErrPasswordIsEmpty)
	})

	t.Run("creates user", func(t *testing.T) {
		user := &models.User{
			Email:    "testRegisterUser@test.com",
			Password: "password",
		}
		err := ur.RegisterUser(user)
		is.NoErr(err)

		var dbUser models.User
		tx.First(&dbUser, "ID = ?", user.ID)

		is.True(dbUser.ID != uuid.UUID{})
		is.Equal(dbUser.Email, "testRegisterUser@test.com")
		is.Equal(dbUser.Password, "password")
		is.Equal(dbUser.LastLogin, nil)
		is.Equal(dbUser.FailedLoginAttempts, 0)
		is.True(!dbUser.AccountLocked)
		is.Equal(dbUser.AccountLockedUntil, nil)
	})

	t.Run("fails on duplicate email", func(t *testing.T) {
		user := &models.User{
			Email:    "testDuplicateRegisterUser@test.com",
			Password: "password",
		}
		err := ur.RegisterUser(user)
		is.NoErr(err)

		err = ur.RegisterUser(user)
		is.Equal(err, apperrors.ErrDuplicateEmail)
	})
}

func TestUserRepository_LookupUser(t *testing.T) {
	testDB := testutils.TestDBSetup()
	is := is.New(t)
	tx := testDB.Begin()
	defer tx.Rollback()

	ur := repository.NewUserRepository(tx)

	// Register a user to look up
	user := &models.User{
		Email:    "testLookupUser@test.com",
		Password: "password",
	}
	err := ur.RegisterUser(user)
	is.NoErr(err)

	t.Run("non-existing user", func(t *testing.T) {
		dbUser, err := ur.GetUserByEmail("doesNotExist@test.com")
		is.Equal(dbUser, nil)
		is.Equal(err, apperrors.ErrUserNotFound)
	})
	t.Run("existing user", func(t *testing.T) {
		dbUser, err := ur.GetUserByEmail(user.Email)
		is.NoErr(err)

		is.True(dbUser.ID != uuid.UUID{})
		is.Equal(dbUser.Email, "testLookupUser@test.com")
		is.Equal(dbUser.Password, "password")
	})
}

func TestUserRepository_GetUserByID(t *testing.T) {
	testDB := testutils.TestDBSetup()
	is := is.New(t)
	tx := testDB.Begin()
	defer tx.Rollback()

	ur := repository.NewUserRepository(tx)

	email := "testGetUserByID@test.com"
	password := "password"

	// Register a user to look up
	user := &models.User{
		Email:    email,
		Password: password,
	}
	err := ur.RegisterUser(user)
	is.NoErr(err)

	t.Run("non-existing user", func(t *testing.T) {
		randUUID := uuid.New()

		dbUser, err := ur.GetUserByID(randUUID.String())
		is.Equal(dbUser, nil)
		is.Equal(err, apperrors.ErrUserNotFound)
	})

	t.Run("existing user", func(t *testing.T) {
		dbUser, err := ur.GetUserByID(user.ID.String())
		is.NoErr(err)

		is.True(dbUser.ID != uuid.UUID{})
		is.Equal(dbUser.Email, email)
		is.Equal(dbUser.Password, password)
		is.Equal(dbUser.ID, user.ID)
	})
}

func TestUserRepository_PermanentlyDeleteUser(t *testing.T) {
	testDB := testutils.TestDBSetup()
	is := is.New(t)
	tx := testDB.Begin()
	defer tx.Rollback()

	ur := repository.NewUserRepository(tx)

	email := "testPermanentlyDeleteUser@test.com"
	password := "password"

	// Register a user to delete
	user := &models.User{
		Email:    email,
		Password: password,
	}
	err := ur.RegisterUser(user)
	is.NoErr(err)

	t.Run("non-existing user", func(t *testing.T) {
		randUUID := uuid.New()
		rowsAffected, err := ur.PermanentlyDeleteUser(randUUID.String())
		is.Equal(rowsAffected, int64(0))
		is.NoErr(err)
	})

	t.Run("existing user", func(t *testing.T) {
		rowsAffected, err := ur.PermanentlyDeleteUser(user.ID.String())
		is.Equal(rowsAffected, int64(1))
		is.NoErr(err)
	})
}

func TestUserRepository_IncrementFailedLogins(t *testing.T) {
	testDB := testutils.TestDBSetup()

	email := "testHandleFailedLogin@test.com"
	password := "password"
	user := &models.User{
		Email:    email,
		Password: password,
	}

	t.Run("fails on non-existent user", func(t *testing.T) {
		is := is.New(t)
		tx := testDB.Begin()
		defer tx.Rollback()
		ur := repository.NewUserRepository(tx)
		err := ur.RegisterUser(user)
		is.NoErr(err)

		err = ur.IncrementFailedLogins(uuid.New().String())
		is.Equal(err, apperrors.ErrUserNotFound)
	})

	t.Run("increments FailedLoginAttempts", func(t *testing.T) {
		is := is.New(t)
		tx := testDB.Begin()
		defer tx.Rollback()
		ur := repository.NewUserRepository(tx)
		err := ur.RegisterUser(user)
		is.NoErr(err)

		err = ur.IncrementFailedLogins(user.ID.String())
		is.NoErr(err)
		user, err = ur.GetUserByEmail(user.Email)
		is.NoErr(err)
		is.Equal(user.FailedLoginAttempts, 1)
	})
}

func TestUserRepository_LockAccount(t *testing.T) {
	testDB := testutils.TestDBSetup()

	email := "testLockAccount@test.com"
	password := "password"
	user := &models.User{
		Email:    email,
		Password: password,
	}

	t.Run("locks on existing user", func(t *testing.T) {
		is := is.New(t)
		tx := testDB.Begin()
		defer tx.Rollback()
		ur := repository.NewUserRepository(tx)
		err := ur.RegisterUser(user)
		is.NoErr(err)

		err = ur.LockAccount(user.ID.String())
		user, err = ur.GetUserByEmail(user.Email)
		is.NoErr(err)
		is.True(user.AccountLocked)
	})

	t.Run("fails on non-existent user", func(t *testing.T) {
		is := is.New(t)
		tx := testDB.Begin()
		defer tx.Rollback()
		ur := repository.NewUserRepository(tx)

		err := ur.LockAccount(uuid.New().String())
		is.Equal(err, apperrors.ErrUserNotFound)
	})
}
