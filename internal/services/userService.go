package services

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"gofit/internal/models"
	"gofit/internal/repository"
	"gofit/pkg/apperrors"
	"gofit/pkg/config"
)

type UserService struct {
	UserRepo *repository.UserRepository
}

func NewUserService(ur *repository.UserRepository) *UserService {
	return &UserService{
		UserRepo: ur,
	}
}

func (us *UserService) RegisterUser(email, password string) error {
	// Check for empty fields
	if email == "" {
		return apperrors.ErrEmailIsEmpty
	}
	if password == "" {
		return apperrors.ErrPasswordIsEmpty
	}

	user, err := models.NewUser(email, password)
	if err != nil {
		return err
	}
	return us.UserRepo.RegisterUser(user)
}

func (us *UserService) LoginUser(email, password string) (string, error) {
	if email == "" {
		return "", apperrors.ErrEmailIsEmpty
	}
	if password == "" {
		return "", apperrors.ErrPasswordIsEmpty
	}

	user, err := us.UserRepo.LookupUser(email)
	if err != nil {
		return "", err
	}

	// TODO: update last login

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", apperrors.ErrInvalidLogin
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv(config.JwtCookieName)))
	if err != nil {
		return "", apperrors.ErrTokenGeneration
	}

	return tokenString, nil
}

func (us *UserService) GetUserProfile(userID string) (*models.UserProfile, error) {
	if userID == "" {
		return nil, apperrors.ErrUserIdEmpty
	}
	user, err := us.UserRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	userProfile := &models.UserProfile{
		Email:     user.Email,
		LastLogin: user.LastLogin,
	}
	return userProfile, nil
}

func (us *UserService) PermanentlyDeleteUser(userID string) error {
	if userID == "" {
		return apperrors.ErrUserIdEmpty
	}
	rowsAffected, err := us.UserRepo.PermanentlyDeleteUser(userID)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return apperrors.ErrUserNotFound
	}
	return nil
}

func (us *UserService) UpdateUser(userID, newEmail, newPassword string) error {
	if userID == "" {
		return apperrors.ErrUserIdEmpty
	}
	user, err := us.UserRepo.GetUserByID(userID)
	if err != nil {
		return err
	}

	// "Ternary", only use new fields
	email, password := user.Email, user.Password
	if newEmail != "" {
		email = newEmail
	}
	if newPassword != "" {
		password = newPassword
	}

	// Use NewUser function for email/pw validation and hashing
	updatedUser, err := models.NewUser(email, password)
	if err != nil {
		return err
	}

	rowsAffected, err := us.UserRepo.UpdateUser(userID, updatedUser.Email, updatedUser.Password)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return apperrors.ErrFailedUserUpade
	}
	return nil
}
