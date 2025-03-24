package services_test

import (
	"math"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/matryer/is"

	"gofit/internal/models"
	"gofit/internal/repository"
	"gofit/internal/services"
	"gofit/internal/testutils"
	"gofit/pkg/apperrors"
	"gofit/pkg/config"
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

	t.Run("valid user", func(t *testing.T) {
		email := "testUserServiceRegisterUser@test.com"
		password := config.TestingPassword
		err := us.RegisterUser(email, password)
		is.NoErr(err)

		// User actually exists in db
		var user models.User
		var defaultUUID uuid.UUID
		us.UserRepo.DB.First(&user, "email = ?", email)
		is.True(user.ID != defaultUUID)
	})
}

func TestUserService_LoginUser(t *testing.T) {
	is := is.New(t)

	testDB := testutils.TestDBSetup()
	tx := testDB.Begin()
	defer tx.Rollback()

	userRepo := repository.NewUserRepository(tx)
	us := services.NewUserService(userRepo)

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

	// Create test user
	email := "testUserServiceLoginUser@test.com"
	password := config.TestingPassword
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

	t.Run("expected token claims", func(t *testing.T) {
		token, err := us.LoginUser(email, password)
		is.NoErr(err)

		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
			return []byte(os.Getenv(config.JwtCookieName)), nil
		})
		is.NoErr(err)
		is.True(parsedToken.Valid)

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		is.True(ok)

		user, err := userRepo.LookupUser(email)
		is.NoErr(err)

		is.Equal(claims["sub"], user.ID.String())

		exp, ok := claims["exp"].(float64)
		is.True(ok)
		expectedExp := float64(time.Now().Unix() + config.TokenExpiration)
		// Account for 5 second expiry difference
		is.True(math.Abs(exp-expectedExp) < 5)
	})
}

func TestUserService_GetUserProfile(t *testing.T) {
	is := is.New(t)

	testDB := testutils.TestDBSetup()
	tx := testDB.Begin()
	defer tx.Rollback()

	userRepo := repository.NewUserRepository(tx)
	us := services.NewUserService(userRepo)

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
		randomUUID, err := uuid.NewRandom()
		is.NoErr(err)
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
