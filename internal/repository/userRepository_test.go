package repository_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/matryer/is"
	"gorm.io/gorm"

	"github.com/al-ce/goauth/internal/models"
	"github.com/al-ce/goauth/internal/repository"
	"github.com/al-ce/goauth/internal/testutils"
	"github.com/al-ce/goauth/pkg/apperrors"
	"github.com/al-ce/goauth/pkg/config"
)

// TestUserRepository_NewUserRepository tests creation of UserRepository
// structs in the `repository` package
func TestUserRepository_NewUserRepository(t *testing.T) {
	is := is.New(t)

	testDB := testutils.TestDBSetup()

	t.Run("creates new user repo", func(t *testing.T) {
		ur, err := repository.NewUserRepository(testDB)
		is.True(ur != nil)
		is.NoErr(err)
	})

	t.Run("returns err with nil db", func(t *testing.T) {
		ur, err := repository.NewUserRepository(nil)
		is.Equal(ur, nil)
		is.Equal(err, apperrors.ErrDatabaseIsNil)
	})
}

// TestUserRepository_RegisterUser tests insertion of new users into the
// `users` table of the database
func TestUserRepository_RegisterUser(t *testing.T) {
	is := is.New(t)

	// Validates user before registration
	t.Run("fails on nil user", func(t *testing.T) {
		ur := setupUserRepository(t)

		err := ur.RegisterUser(nil)
		is.Equal(err, apperrors.ErrUserIsNil)
	})
	t.Run("fails on missing email", func(t *testing.T) {
		ur := setupUserRepository(t)

		user := &models.User{
			Password: "password",
		}
		err := ur.RegisterUser(user)
		is.Equal(err, apperrors.ErrEmailIsEmpty)
	})
	t.Run("fails on missing password", func(t *testing.T) {
		ur := setupUserRepository(t)

		user := &models.User{
			Email: "testRegisterUser@test.com",
		}
		err := ur.RegisterUser(user)
		is.Equal(err, apperrors.ErrPasswordIsEmpty)
	})

	// Inserts user into db with complete User value
	t.Run("registers user", func(t *testing.T) {
		ur := setupUserRepository(t)

		user := &models.User{
			Email:    "testRegisterUser@test.com",
			Password: "password",
		}
		err := ur.RegisterUser(user)
		is.NoErr(err)

		// Lookup expected user in the db
		var dbUser models.User
		ur.DB.First(&dbUser, "ID = ?", user.ID)

		// Sets given user values
		is.True(dbUser.ID != uuid.UUID{})
		is.Equal(dbUser.Email, "testRegisterUser@test.com")
		is.Equal(dbUser.Password, "password")
		// Sets default values
		is.Equal(dbUser.LastLogin, nil)
		is.Equal(dbUser.FailedLoginAttempts, 0)
		is.True(!dbUser.AccountLocked)
		is.Equal(dbUser.AccountLockedUntil, nil)
	})
}

// TestUserRepository_GetUserByEmail tests lookup of registered users in the database
func TestUserRepository_GetUserByEmail(t *testing.T) {
	is := is.New(t)

	// Error on empty email
	t.Run("empty email", func(t *testing.T) {
		ur := setupUserRepository(t)

		dbUser, err := ur.GetUserByEmail("")
		is.Equal(dbUser, nil)
		is.Equal(err, apperrors.ErrEmailIsEmpty)
	})

	// Error on non-existing user lookup
	t.Run("non-existing user", func(t *testing.T) {
		ur := setupUserRepository(t)

		dbUser, err := ur.GetUserByEmail("doesNotExist@test.com")
		is.Equal(dbUser, nil)
		is.Equal(err, gorm.ErrRecordNotFound)
	})

	// Success on existing-user lookup
	t.Run("existing user", func(t *testing.T) {
		ur := setupUserRepository(t)

		// Register a user to look up
		user := &models.User{
			Email:    "testGetUserByEmail@test.com",
			Password: "password",
		}
		err := ur.RegisterUser(user)
		is.NoErr(err)

		dbUser, err := ur.GetUserByEmail(user.Email)
		is.NoErr(err)

		is.True(dbUser.ID != uuid.UUID{})
		is.Equal(dbUser.Email, "testGetUserByEmail@test.com")
		is.Equal(dbUser.Password, "password")
	})
}

// TestUserRepository_LookupUser tests lookup of registered users in the database
func TestUserRepository_GetUserByID(t *testing.T) {
	is := is.New(t)

	ur := setupUserRepository(t)

	email := "testGetUserByID@test.com"

	// Register a user to look up
	user := &models.User{
		Email:    email,
		Password: testutils.TestingPassword,
	}
	err := ur.RegisterUser(user)
	is.NoErr(err)

	// Error on empty user ID
	t.Run("empty user ID", func(t *testing.T) {
		dbUser, err := ur.GetUserByID("")
		is.Equal(dbUser, nil)
		is.Equal(err, apperrors.ErrUserIdEmpty)
	})

	// Error on non-existing user lookup
	t.Run("non-existing user", func(t *testing.T) {
		randUUID := uuid.New()

		dbUser, err := ur.GetUserByID(randUUID.String())
		is.Equal(dbUser, nil)
		is.Equal(err, gorm.ErrRecordNotFound)
	})

	// Success on existing-user lookup
	t.Run("existing user", func(t *testing.T) {
		dbUser, err := ur.GetUserByID(user.ID.String())
		is.NoErr(err)

		is.True(dbUser.ID != uuid.UUID{})
		is.Equal(dbUser.Email, email)
		is.Equal(dbUser.Password, testutils.TestingPassword)
		is.Equal(dbUser.ID, user.ID)
	})
}

// TestUserRepository_PermanentlyDeleteUser tests deletion of existing users in database
func TestUserRepository_PermanentlyDeleteUser(t *testing.T) {
	is := is.New(t)

	ur := setupUserRepository(t)

	// Error on empty user ID
	t.Run("empty user ID", func(t *testing.T) {
		rowsAffected, err := ur.PermanentlyDeleteUser("")
		is.Equal(rowsAffected, int64(0))
		is.Equal(err, apperrors.ErrUserIdEmpty)
	})

	// Error on non-existing user lookup
	t.Run("non-existing user", func(t *testing.T) {
		randUUID := uuid.New()
		rowsAffected, err := ur.PermanentlyDeleteUser(randUUID.String())
		is.Equal(rowsAffected, int64(0))
		is.NoErr(err)
	})

	// Success on existing-user lookup
	t.Run("existing user", func(t *testing.T) {
		email := "testPermanentlyDeleteUser@test.com"

		// Register a user to delete
		user := &models.User{
			Email:    email,
			Password: testutils.TestingPassword,
		}
		err := ur.RegisterUser(user)
		is.NoErr(err)
		rowsAffected, err := ur.PermanentlyDeleteUser(user.ID.String())
		is.Equal(rowsAffected, int64(1))
		is.NoErr(err)
		// Confirm user is no longer in the database
		user, err = ur.GetUserByEmail(email)
		is.Equal(user, nil)
		is.Equal(err, gorm.ErrRecordNotFound)
	})
}

// TestUserRepository_UpdateUser test user updates into the database
func TestUserRepository_UpdateUser(t *testing.T) {
	is := is.New(t)

	// Error on empty user ID
	t.Run("empty user ID", func(t *testing.T) {
		ur := setupUserRepository(t)

		err := ur.UpdateUser("", map[string]any{})
		is.Equal(err, apperrors.ErrUserIdEmpty)
	})

	// Error on non-existent user
	t.Run("non-existent user", func(t *testing.T) {
		ur := setupUserRepository(t)

		err := ur.UpdateUser("doesNotExist", map[string]any{})
		is.Equal(err, apperrors.ErrUserNotFound)
	})

	// Can update user
	t.Run("can update user", func(t *testing.T) {
		ur := setupUserRepository(t)

		// Register a user to update
		email := "testUpdateUser@test.com"
		user := &models.User{
			Email:    email,
			Password: testutils.TestingPassword,
		}
		err := ur.RegisterUser(user)
		is.NoErr(err)

		// Update user
		referenceTime := time.Now().UTC().Add(time.Hour * 24).UTC().Truncate(time.Second)
		err = ur.UpdateUser(user.ID.String(), map[string]any{
			"email":                 "newUserName@test.com",
			"password":              "newpassword",
			"last_login":            referenceTime,
			"failed_login_attempts": 99,
			"account_locked":        true,
			"account_locked_until":  referenceTime,
		})
		is.NoErr(err)

		// Get updated user and check for updated fields
		user, err = ur.GetUserByID(user.ID.String())
		is.NoErr(err)
		t.Run("updates email", func(t *testing.T) {
			is.Equal(user.Email, "newUserName@test.com")
		})
		t.Run("updates password", func(t *testing.T) {
			is.Equal(user.Password, "newpassword")
		})
		t.Run("updates last_login", func(t *testing.T) {
			is.Equal(user.LastLogin, &referenceTime)
		})
		t.Run("updates failed_login_attempts", func(t *testing.T) {
			is.Equal(user.FailedLoginAttempts, 99)
		})
		t.Run("updates account_locked", func(t *testing.T) {
			is.True(user.AccountLocked)
		})
		t.Run("updates account_locked_until", func(t *testing.T) {
			is.Equal(user.LastLogin, &referenceTime)
		})
	})
}

func TestUserRepository_IncrementFailedLogins(t *testing.T) {
	is := is.New(t)

	// Error on empty user ID
	t.Run("empty user ID", func(t *testing.T) {
		ur := setupUserRepository(t)

		err := ur.IncrementFailedLogins("")
		is.Equal(err, apperrors.ErrUserIdEmpty)
	})

	// Error on non-existing user lookup
	t.Run("fails on non-existent user", func(t *testing.T) {
		ur := setupUserRepository(t)

		err := ur.IncrementFailedLogins(uuid.New().String())
		is.Equal(err, apperrors.ErrUserNotFound)
	})

	// Increments failed login attempts on successful lookup
	t.Run("increments FailedLoginAttempts", func(t *testing.T) {
		ur := setupUserRepository(t)

		// Register a user to fail logins with
		email := "testHandleFailedLogin@test.com"
		user := &models.User{
			Email:    email,
			Password: testutils.TestingPassword,
		}
		err := ur.RegisterUser(user)
		is.NoErr(err)

		// Increment login attempts
		for i := range 10 {
			err = ur.IncrementFailedLogins(user.ID.String())
			is.NoErr(err)

			user, err = ur.GetUserByEmail(user.Email)
			is.NoErr(err)
			is.Equal(user.FailedLoginAttempts, i+1)
		}
	})
}

func TestUserRepository_LockAccount(t *testing.T) {
	is := is.New(t)

	t.Run("locks on existing user", func(t *testing.T) {
		ur := setupUserRepository(t)

		// Register test user
		email := "testLockAccount@test.com"
		user := &models.User{
			Email:    email,
			Password: testutils.TestingPassword,
		}
		err := ur.RegisterUser(user)
		is.NoErr(err)

		beforeLock := time.Now().UTC()

		// Lock account
		err = ur.LockAccount(user.ID.String())
		is.NoErr(err)

		// Check lock status
		user, err = ur.GetUserByEmail(user.Email)
		is.NoErr(err)
		is.True(user.AccountLocked)

		// Account should be locked until AccountLockoutLength, within errorWindow
		lockDuration := user.AccountLockedUntil.Sub(beforeLock)
		errorWindow := 5 * time.Second
		minDuration := config.AccountLockoutLength - errorWindow
		maxDuration := config.AccountLockoutLength + errorWindow
		is.True(lockDuration >= minDuration && lockDuration <= maxDuration)
	})

	t.Run("fails on non-existent user", func(t *testing.T) {
		ur := setupUserRepository(t)

		err := ur.LockAccount(uuid.New().String())
		is.Equal(err, apperrors.ErrUserNotFound)
	})
}

func TestUserRepository_UnlockAccount(t *testing.T) {
	is := is.New(t)

	t.Run("unlocks locked user", func(t *testing.T) {
		ur := setupUserRepository(t)

		// Register test user
		email := "testUnlockAccount@test.com"
		user := &models.User{
			Email:    email,
			Password: testutils.TestingPassword,
		}
		err := ur.RegisterUser(user)
		is.NoErr(err)

		// Lock account
		err = ur.LockAccount(user.ID.String())
		is.NoErr(err)

		// Check lock status (expect locked)
		user, err = ur.GetUserByEmail(user.Email)
		is.NoErr(err)
		is.True(user.AccountLocked)

		// Unlock account
		err = ur.UnlockAccount(user.ID.String())
		is.NoErr(err)

		// Check lock status (expect unlocked)
		user, err = ur.GetUserByEmail(user.Email)
		is.NoErr(err)
		is.True(!user.AccountLocked)
		is.Equal(user.AccountLockedUntil, nil)
		is.Equal(user.FailedLoginAttempts, 0)
	})
}

// TestUserRepository_UnlockAllExpiredLocks tests that any locked accounts that
// are past the lock expiration date are locked by UnlockAllExpiredLocks
func TestUserRepository_UnlockAllExpiredLocks(t *testing.T) {
	is := is.New(t)

	t.Run("unlocks all locked users", func(t *testing.T) {
		ur := setupUserRepository(t)
		// Register test users
		var users []*models.User
		for i := range 10 {
			email := fmt.Sprintf("testUnlockAccount%d@test.com", i)
			user := &models.User{
				Email:    email,
				Password: testutils.TestingPassword,
			}
			err := ur.RegisterUser(user)
			is.NoErr(err)
			// Lock account manually, setting expiration date to a past time

			// Set expiration time to future (*1) for even iterations, past (*-1) for odd
			hourMultiplier := -1
			if i%2 == 0 {
				hourMultiplier = 1
			}
			expirationTime := time.Now().UTC().Add(time.Duration(hourMultiplier) * time.Hour)

			result := ur.DB.Model(&models.User{}).
				Where("id = ?", user.ID).
				Updates(map[string]any{
					"account_locked":       true,
					"account_locked_until": expirationTime,
				})
			is.NoErr(result.Error)
			users = append(users, user)

			// Check user was created and locked
			var gotUser models.User
			ur.DB.First(&gotUser, "email = ?", email)
			is.Equal(gotUser.AccountLocked, true)
		}

		// Unlock all locked accounts with past expiration
		affected, err := ur.UnlockAllExpiredLocks()
		is.NoErr(err)
		is.Equal(affected, int64(len(users))/2)

		// Check that all users with past expiration are now unlocked
		for i, u := range users {
			var user models.User
			ur.DB.First(&user, "id = ?", u.ID)
			// Even numbered users should be locked, odd unlocked
			is.Equal(user.AccountLocked, i%2 == 0)
		}
	})
}

func setupUserRepository(t *testing.T) (*repository.UserRepository) {
	t.Helper()

	testDB := testutils.TestDBSetup()
	tx := testDB.Begin()
	t.Cleanup(func() { tx.Rollback() })

	ur, err := repository.NewUserRepository(tx)
	if err != nil {
		t.Fatalf("failed to create user repository: %v", err)
	}
	return ur
}
