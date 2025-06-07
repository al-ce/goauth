package services_test

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/matryer/is"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/al-ce/goauth/internal/models"
	"github.com/al-ce/goauth/internal/repository"
	"github.com/al-ce/goauth/internal/services"
	"github.com/al-ce/goauth/internal/testutils"
	"github.com/al-ce/goauth/pkg/apperrors"
	"github.com/al-ce/goauth/pkg/config"
)

// TestUserService_NewUserService tests the creation of a new UserService
func TestUserService_NewUserService(t *testing.T) {
	is := is.New(t)
	testDB := testutils.TestDBSetup()

	t.Run("returns err with nil user repo", func(t *testing.T) {
		tx := testDB.Begin()
		defer tx.Rollback()

		sr, err := repository.NewSessionRepository(tx)
		is.Equal(err, nil)

		userService, err := services.NewUserService(nil, sr)
		is.Equal(userService, nil)
		is.Equal(err, apperrors.ErrUserRepoIsNil)
	})

	t.Run("returns err with nil session repo", func(t *testing.T) {
		tx := testDB.Begin()
		defer tx.Rollback()

		ur, err := repository.NewUserRepository(tx)
		is.Equal(err, nil)

		userService, err := services.NewUserService(ur, nil)
		is.Equal(userService, nil)
		is.Equal(err, apperrors.ErrSessionRepoIsNil)
	})

	t.Run("creates user service", func(t *testing.T) {
		userService := setupUserService(t)
		is.True(userService != nil)
	})
}

// TestUserService_RegisterUser tests that the user service can register a user
func TestUserService_RegisterUser(t *testing.T) {
	is := is.New(t)


	// Error on empty email
	t.Run("empty email", func(t *testing.T) {
		us := setupUserService(t)
		err := us.RegisterUser("", "password")
		is.Equal(err, apperrors.ErrEmailIsEmpty)
	})

	// Error on empty password
	t.Run("empty password", func(t *testing.T) {
		us := setupUserService(t)
		err := us.RegisterUser("some@test.com", "")
		is.Equal(err, apperrors.ErrPasswordIsEmpty)
	})

	// Can register with a valid User value
	t.Run("valid user", func(t *testing.T) {
		us := setupUserService(t)
		email := "testUserServiceRegisterUser@test.com"
		err := us.RegisterUser(email, testutils.TestingPassword)
		is.NoErr(err)

		// Check user actually exists in db
		var user models.User
		result := us.UserRepo.DB.First(&user, "email = ?", email)
		is.NoErr(result.Error)
	})

	// Cannot register a duplicate user
	t.Run("duplicate user", func(t *testing.T) {
		us := setupUserService(t)
		email := "testUserServiceRegisterUser1@test.com"
		err := us.RegisterUser(email, testutils.TestingPassword)
		is.NoErr(err)
		err = us.RegisterUser(email, testutils.TestingPassword)
		is.Equal(err, apperrors.ErrDuplicateEmail)
	})

}

func TestUserService_GetUserProfile(t *testing.T) {
	is := is.New(t)

	us := setupUserService(t)

	// Create test user
	email := "testUserServiceGetUserProfile@test.com"
	user := &models.User{
		Email: email, Password: testutils.TestingPassword,
	}
	err := us.UserRepo.RegisterUser(user)
	is.NoErr(err)

	t.Run("empty userID", func(t *testing.T) {
		userProfile, err := us.GetUserProfile("")
		is.Equal(err, apperrors.ErrUserIdEmpty)
		is.Equal(userProfile, nil)
	})

	t.Run("non-existent user", func(t *testing.T) {
		randomUUID := uuid.New()
		userProfile, err := us.GetUserProfile(randomUUID.String())
		is.True(err == gorm.ErrRecordNotFound)
		is.True(userProfile == nil)
	})

	t.Run("existing user", func(t *testing.T) {
		userProfile, err := us.GetUserProfile(user.ID.String())
		is.NoErr(err)
		is.Equal(userProfile.Email, user.Email)
	})
}

// TestUserService_UpdateUser test user updates into the database
func TestUserService_UpdateUser(t *testing.T) {
	is := is.New(t)

	// Error on empty user ID
	t.Run("empty user ID", func(t *testing.T) {
		us := setupUserService(t)
		err := us.UpdateUser("", map[string]any{})
		is.Equal(err, apperrors.ErrUserIdEmpty)
	})

	// Error on non-existent user
	t.Run("non-existent user", func(t *testing.T) {
		us := setupUserService(t)
		err := us.UpdateUser("doesNotExist", map[string]any{})
		is.Equal(err, apperrors.ErrUserNotFound)
	})

	// Can update user
	t.Run("can update user", func(t *testing.T) {
		// Register a user to update
		email := "testUpdateUser@test.com"
		us := setupUserService(t)

		err := us.RegisterUser(email, testutils.TestingPassword)
		is.NoErr(err)

		// Update user
		referenceTime := time.Now().UTC().Add(time.Hour * 24).UTC().Truncate(time.Second)
		var user models.User
		result := us.UserRepo.DB.First(&user, "email = ?", email)
		is.NoErr(result.Error)
		err = us.UpdateUser(user.ID.String(), map[string]any{
			"email":                 "newUserName@test.com",
			"password":              "new" + testutils.TestingPassword,
			"last_login":            referenceTime,
			"failed_login_attempts": 99,
			"account_locked":        true,
			"account_locked_until":  referenceTime,
		})
		is.NoErr(err)

		// Get updated user and check for updated fields
		updatedUser, err := us.UserRepo.GetUserByID(user.ID.String())
		is.NoErr(err)
		t.Run("updates email", func(t *testing.T) {
			is.Equal(updatedUser.Email, "newUserName@test.com")
		})
		t.Run("updates password", func(t *testing.T) {
			err := bcrypt.CompareHashAndPassword(
				[]byte(updatedUser.Password),
				[]byte("new"+testutils.TestingPassword),
			)
			is.NoErr(err)
		})
		t.Run("updates last_login", func(t *testing.T) {
			is.Equal(updatedUser.LastLogin, &referenceTime)
		})
		t.Run("updates failed_login_attempts", func(t *testing.T) {
			is.Equal(updatedUser.FailedLoginAttempts, 99)
		})
		t.Run("updates account_locked", func(t *testing.T) {
			is.True(updatedUser.AccountLocked)
		})
		t.Run("updates account_locked_until", func(t *testing.T) {
			is.Equal(updatedUser.LastLogin, &referenceTime)
		})
	})

	// Cannot update to an existing email
	t.Run("cannot update to existing email", func(t *testing.T) {
		email := "testUpdateExistingUser@test.com"
		us := setupUserService(t)

		err := us.RegisterUser(email, testutils.TestingPassword)
		is.NoErr(err)
		err = us.RegisterUser(email, testutils.TestingPassword)
		is.Equal(err, apperrors.ErrDuplicateEmail)
	})
}

// TestUserService_LoginUser tests that a user can be logged in, generating a
// session cookie and creating a session in the database
func TestUserService_LoginUser(t *testing.T) {
	is := is.New(t)

	email := "testUserServiceLoginUser@test.com"

	t.Run("non existing user", func(t *testing.T) {
		us := setupUserService(t)
		_, err := us.LoginUser("doesNotExist@test.com", "password")
		is.Equal(err, apperrors.ErrUserNotFound)
	})

	t.Run("empty email", func(t *testing.T) {
		us := setupUserService(t)
		_, err := us.LoginUser("", "password")
		is.Equal(err, apperrors.ErrEmailIsEmpty)
	})

	t.Run("empty password", func(t *testing.T) {
		us := setupUserService(t)
		_, err := us.LoginUser("some@test.com", "")
		is.Equal(err, apperrors.ErrPasswordIsEmpty)
	})

	t.Run("invalid password", func(t *testing.T) {
		us := setupUserService(t)
		err := us.RegisterUser(email, testutils.TestingPassword)
		is.NoErr(err)
		_, err = us.LoginUser(email, "thisIsNotThePassword")
		is.Equal(err, apperrors.ErrInvalidLogin)
	})

	t.Run("valid login", func(t *testing.T) {
		us := setupUserService(t)
		err := us.RegisterUser(email, testutils.TestingPassword)
		is.NoErr(err)
		token, err := us.LoginUser(email, testutils.TestingPassword)
		is.NoErr(err)
		is.True(token != "")
	})

	t.Run("deny locked account login", func(t *testing.T) {
		us := setupUserService(t)

		// Register user
		err := us.RegisterUser(email, testutils.TestingPassword)
		is.NoErr(err)

		// Get registered user
		user, err := us.UserRepo.GetUserByEmail(email)
		is.NoErr(err)

		// Lock account
		us.UserRepo.LockAccount(user.ID.String())

		// Attempt immediate locked-account login
		_, err = us.LoginUser(email, testutils.TestingPassword)
		is.Equal(err, apperrors.ErrAccountIsLocked)

		// Set account_locked_until to before now and attempt login
		// (should unlock account)
		result := us.UserRepo.DB.Model(&models.User{}).Where("id = ?", user.ID).
			Updates(map[string]any{
				"account_locked_until": time.Now().UTC().Add(-time.Hour),
			})
		is.NoErr(result.Error)

		// Attempt login, expect unlocked
		_, err = us.LoginUser(email, testutils.TestingPassword)
		is.NoErr(err)

	})

	t.Run("locks account after max attempts", func(t *testing.T) {
		us := setupUserService(t)
		// Register user
		err := us.RegisterUser(email, testutils.TestingPassword)
		is.NoErr(err)

		// Get registered user
		user, err := us.UserRepo.GetUserByEmail(email)
		is.NoErr(err)

		// Manually set failed attempts to max-1 (locked on next attempt)
		us.UserRepo.DB.Model(&models.User{}).Where("id = ?", user.ID).
			Updates(map[string]any{"failed_login_attempts": config.MaxLoginAttempts - 1})

		// Fail a login attempt
		us.LoginUser(email, "thisIsNotThePassword")
		// Attempt subsequent login, expecting locked account
		_, err = us.LoginUser(email, testutils.TestingPassword)
		is.Equal(err, apperrors.ErrAccountIsLocked)
	})
}

// TestUserService_Logout checks that a token is no longer valid after Logout is called
func TestUserService_Logout(t *testing.T) {
	is := is.New(t)

	us := setupUserService(t)
	t.Run("error on empty token string", func(t *testing.T) {
		err := us.Logout("")
		is.Equal(err, apperrors.ErrSessionIdIsEmpty)
	})

	t.Run("invalidates a token", func(t *testing.T) {
		// Register and login a user
		email := "testUserServiceLogout@test.com"
		err := us.RegisterUser(email, testutils.TestingPassword)
		is.NoErr(err)
		token, err := us.LoginUser(email, testutils.TestingPassword)
		is.NoErr(err)
		is.NoErr(err)

		// Logout a user
		err = us.Logout(token)

		// Get session ID from token
		parts := strings.Split(token, ".")
		is.True(len(parts) == 2)
		sessionID := parts[0]
		parsedID, err := uuid.Parse(sessionID)
		is.NoErr(err)

		// Check that corresponding session no longer exists in database
		session, err := us.SessionRepo.GetUnexpiredSessionByID(parsedID)
		is.Equal(session, nil)
		is.Equal(err, gorm.ErrRecordNotFound)
	})
}

// TestUserService_Logout checks that all tokens associated with a user ares no longer valid after
// LogoutEverywhere is called
func TestUserService_LogoutEverywhere(t *testing.T) {
	is := is.New(t)

	us := setupUserService(t)
	t.Run("error on empty userID string", func(t *testing.T) {
		err := us.LogoutEverywhere("")
		is.Equal(err, apperrors.ErrUserIdEmpty)
	})

	t.Run("invalidates all user's tokens", func(t *testing.T) {
		// Register a user
		email := "testUserServiceLogout@test.com"
		err := us.RegisterUser(email, testutils.TestingPassword)
		is.NoErr(err)
		// Get user ID
		user, err := us.UserRepo.GetUserByEmail(email)
		is.NoErr(err)
		// Login user multiple times
		tokens := []string{}
		for range 10 {
			token, err := us.LoginUser(email, testutils.TestingPassword)
			is.NoErr(err)
			tokens = append(tokens, token)
		}

		// Invalidate all user's tokens
		err = us.LogoutEverywhere(user.ID.String())

		// Check that corresponding session no longer exists in database
		for _, token := range tokens {
			// Get session ID from token
			parts := strings.Split(token, ".")
			is.True(len(parts) == 2)
			sessionID := parts[0]
			parsedID, err := uuid.Parse(sessionID)
			is.NoErr(err)
			// Confirm session is gone
			session, err := us.SessionRepo.GetUnexpiredSessionByID(parsedID)
			is.Equal(session, nil)
			is.Equal(err, gorm.ErrRecordNotFound)
		}
	})
}

func TestUserService_PermanentlyDeleteUser(t *testing.T) {
	is := is.New(t)

	us := setupUserService(t)

	t.Run("empty userID", func(t *testing.T) {
		err := us.PermanentlyDeleteUser("")
		is.Equal(err, apperrors.ErrUserIdEmpty)
	})

	t.Run("non-existent user", func(t *testing.T) {
		randomUUID := uuid.New()
		err := us.PermanentlyDeleteUser(randomUUID.String())
		is.True(err == apperrors.ErrUserNotFound)
	})

	t.Run("existing user", func(t *testing.T) {
		// Register a test user
		email := "testUserServicePermanentlyDeleteUser@test.com"
		err := us.RegisterUser(email, testutils.TestingPassword)
		is.NoErr(err)
		// Get user ID
		user, err := us.UserRepo.GetUserByEmail(email)
		is.NoErr(err)
		// Delete user
		err = us.PermanentlyDeleteUser(user.ID.String())
		is.NoErr(err)
		// Confirm user no longer exists in db
		user, err = us.UserRepo.GetUserByEmail(email)
		is.Equal(user, nil)
		is.Equal(err, gorm.ErrRecordNotFound)
	})
}

func setupUserService(t *testing.T) *services.UserService {
	t.Helper()

	testDB := testutils.TestDBSetup()
	tx := testDB.Begin()
	t.Cleanup(func() { tx.Rollback() })

	ur, err := repository.NewUserRepository(tx)
	if err != nil {
		t.Fatalf("failed to create user repository: %v", err)
	}
	sr, err := repository.NewSessionRepository(tx)
	if err != nil {
		t.Fatalf("failed to create session repository: %v", err)
	}
	us, err := services.NewUserService(ur, sr)
	if err != nil {
		t.Fatalf("failed to create user service: %v", err)
	}
	return us
}
