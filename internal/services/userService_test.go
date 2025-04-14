package services_test

import (
	"math"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/matryer/is"

	"goauth/internal/models"
	"goauth/internal/repository"
	"goauth/internal/services"
	"goauth/internal/testutils"
	"goauth/pkg/apperrors"
	"goauth/pkg/config"
)

func TestUserService_RegisterUser(t *testing.T) {
	is := is.New(t)

	testDB := testutils.TestDBSetup()
	tx := testDB.Begin()
	defer tx.Rollback()

	userRepo := repository.NewUserRepository(tx)
	sessionRepo := repository.NewSessionRepository(tx)
	us := services.NewUserService(userRepo, sessionRepo)

	t.Run("empty email", func(t *testing.T) {
		err := us.RegisterUser("", "password")
		is.Equal(err, apperrors.ErrEmailIsEmpty)
	})

	t.Run("empty password", func(t *testing.T) {
		err := us.RegisterUser("some@test.com", "")
		is.Equal(err, apperrors.ErrPasswordIsEmpty)
	})

	t.Run("valid user", func(t *testing.T) {
		email := "testUserServiceRegisterUser@test.com"
		password := config.TestingPassword
		err := us.RegisterUser(email, password)
		is.NoErr(err)

		// User actually exists in db
		var user models.User
		result := us.UserRepo.DB.First(&user, "email = ?", email)
		is.NoErr(result.Error)
	})
}

func TestUserService_LoginUser(t *testing.T) {
	is := is.New(t)

	testDB := testutils.TestDBSetup()

	// We need separate transactions for each test case due to unique session token constraints

	// Create test user data once
	email := "testUserServiceLoginUser@test.com"
	password := config.TestingPassword

	t.Run("non existing user", func(t *testing.T) {
		// Each test case gets its own transaction
		tx := testDB.Begin()
		defer tx.Rollback()

		userRepo := repository.NewUserRepository(tx)
		sessionRepo := repository.NewSessionRepository(tx)
		us := services.NewUserService(userRepo, sessionRepo)

		_, err := us.LoginUser("doesNotExist@test.com", "password")
		is.Equal(err, apperrors.ErrUserNotFound)
	})

	t.Run("empty email", func(t *testing.T) {
		tx := testDB.Begin()
		defer tx.Rollback()

		userRepo := repository.NewUserRepository(tx)
		sessionRepo := repository.NewSessionRepository(tx)
		us := services.NewUserService(userRepo, sessionRepo)

		_, err := us.LoginUser("", "password")
		is.Equal(err, apperrors.ErrEmailIsEmpty)
	})

	t.Run("empty password", func(t *testing.T) {
		tx := testDB.Begin()
		defer tx.Rollback()

		userRepo := repository.NewUserRepository(tx)
		sessionRepo := repository.NewSessionRepository(tx)
		us := services.NewUserService(userRepo, sessionRepo)

		_, err := us.LoginUser("some@test.com", "")
		is.Equal(err, apperrors.ErrPasswordIsEmpty)
	})

	// For the remaining tests that need a user, create a separate transaction and user
	t.Run("invalid password and valid login", func(t *testing.T) {
		tx := testDB.Begin()
		defer tx.Rollback()

		userRepo := repository.NewUserRepository(tx)
		sessionRepo := repository.NewSessionRepository(tx)
		us := services.NewUserService(userRepo, sessionRepo)

		// Create test user in this transaction
		err := us.RegisterUser(email, password)
		is.NoErr(err)

		t.Run("invalid password", func(t *testing.T) {
			_, err := us.LoginUser(email, "thisIsNotThePassword")
			is.Equal(err, apperrors.ErrInvalidLogin)
		})

		t.Run("valid login", func(t *testing.T) {
			token, err := us.LoginUser(email, password)
			is.NoErr(err)
			is.True(token != "")
		})
	})

	t.Run("expected token claims", func(t *testing.T) {
		tx := testDB.Begin()
		defer tx.Rollback()

		userRepo := repository.NewUserRepository(tx)
		sessionRepo := repository.NewSessionRepository(tx)
		us := services.NewUserService(userRepo, sessionRepo)

		testEmail := "tokenClaimsTest@example.com"
		err := us.RegisterUser(testEmail, password)
		is.NoErr(err)

		token, err := us.LoginUser(testEmail, password)
		is.NoErr(err)

		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
			return []byte(os.Getenv(config.JwtCookieName)), nil
		})
		is.NoErr(err)
		is.True(parsedToken.Valid)

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		is.True(ok)

		user, err := userRepo.GetUserByEmail(testEmail)
		is.NoErr(err)

		is.Equal(claims["sub"], user.ID.String())

		exp, ok := claims["exp"].(float64)
		is.True(ok)
		expectedExp := float64(time.Now().Unix() + config.TokenExpiration)
		// Account for 5 second expiry difference
		is.True(math.Abs(exp-expectedExp) < 5)

		// Verify session was created
		session, err := sessionRepo.GetUnexpiredSessionByToken(token)
		is.NoErr(err)
		is.Equal(session.UserID.String(), user.ID.String())
	})

	t.Run("deny locked account login", func(t *testing.T) {
		tx := testDB.Begin()
		defer tx.Rollback()

		userRepo := repository.NewUserRepository(tx)
		sessionRepo := repository.NewSessionRepository(tx)
		us := services.NewUserService(userRepo, sessionRepo)

		// Register user
		testEmail := "lockedUserLogin@example.com"
		err := us.RegisterUser(testEmail, password)
		is.NoErr(err)

		// Get registered user
		user, err := userRepo.GetUserByEmail(testEmail)
		is.NoErr(err)

		// Lock account
		us.UserRepo.LockAccount(user.ID.String())

		// Attempt locked-account login
		_, err = us.LoginUser(testEmail, password)
		is.Equal(err, apperrors.ErrAccountIsLocked)
	})

	t.Run("locks account after max attempts", func(t *testing.T) {
		tx := testDB.Begin()
		defer tx.Rollback()

		userRepo := repository.NewUserRepository(tx)
		sessionRepo := repository.NewSessionRepository(tx)
		us := services.NewUserService(userRepo, sessionRepo)

		// Register user
		testEmail := "lockoutTheUser@example.com"
		err := us.RegisterUser(testEmail, password)
		is.NoErr(err)

		// Get registered user
		user, err := userRepo.GetUserByEmail(testEmail)
		is.NoErr(err)

		// Manually set failed attempts to max-1
		tx.Model(&models.User{}).Where("id = ?", user.ID).
			Updates(map[string]any{"failed_login_attempts": config.MaxLoginAttempts - 1})

		// Fail a login attempt
		_, err = us.LoginUser(testEmail, "thisIsNotThePassword")
		is.Equal(err, apperrors.ErrInvalidLogin)

		// Attempt subsequent login, expecting locked account
		_, err = us.LoginUser(testEmail, password)
		is.Equal(err, apperrors.ErrAccountIsLocked)

	})
}

func TestUserService_GetUserProfile(t *testing.T) {
	is := is.New(t)

	testDB := testutils.TestDBSetup()
	tx := testDB.Begin()
	defer tx.Rollback()

	userRepo := repository.NewUserRepository(tx)
	sessionRepo := repository.NewSessionRepository(tx)
	us := services.NewUserService(userRepo, sessionRepo)

	// Create test user
	email := "testUserServiceGetUserProfile@test.com"
	password := config.TestingPassword
	user := &models.User{
		Email: email, Password: password,
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
		is.True(err == apperrors.ErrUserNotFound)
		is.True(userProfile == nil)
	})

	t.Run("existing user", func(t *testing.T) {
		userProfile, err := us.GetUserProfile(user.ID.String())
		is.NoErr(err)
		is.Equal(userProfile.Email, user.Email)
	})
}

func TestUserService_PermanentlyDeleteUser(t *testing.T) {
	is := is.New(t)

	testDB := testutils.TestDBSetup()
	tx := testDB.Begin()
	defer tx.Rollback()

	userRepo := repository.NewUserRepository(tx)
	sessionRepo := repository.NewSessionRepository(tx)
	us := services.NewUserService(userRepo, sessionRepo)

	// Create test user
	email := "testUserServicePermanentlyDeleteUser@test.com"
	password := config.TestingPassword
	user := &models.User{
		Email: email, Password: password,
	}
	err := us.UserRepo.RegisterUser(user)
	is.NoErr(err)

	t.Run("empty userID", func(t *testing.T) {
		err := us.PermanentlyDeleteUser("")
		is.Equal(err, apperrors.ErrUserIdEmpty)
	})

	t.Run("non-existent user", func(t *testing.T) {
		randomUUID := uuid.New()
		err = us.PermanentlyDeleteUser(randomUUID.String())
		is.True(err == apperrors.ErrUserNotFound)
	})

	t.Run("existing user", func(t *testing.T) {
		err := us.PermanentlyDeleteUser(user.ID.String())
		is.NoErr(err)
	})
}
