package services

import (
	"net/mail"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	passwordvalidator "github.com/wagslane/go-password-validator"
	"golang.org/x/crypto/bcrypt"

	"gofit/internal/models"
	"gofit/internal/repository"
	"gofit/pkg/apperrors"
	"gofit/pkg/config"
)

type UserService struct {
	UserRepo    *repository.UserRepository
	SessionRepo *repository.SessionRepository
}

func NewUserService(ur *repository.UserRepository, sr *repository.SessionRepository) *UserService {
	return &UserService{
		UserRepo:    ur,
		SessionRepo: sr,
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
	var err error
	if email == "" {
		err := apperrors.ErrEmailIsEmpty
		return "", err
	}
	if password == "" {
		err := apperrors.ErrPasswordIsEmpty
		return "", err
	}

	user, err := us.UserRepo.LookupUser(email)
	if err != nil {
		return "", err
	}

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

	// Create a session
	session := &models.Session{
		UserID:    user.ID,
		Token:     tokenString,
		ExpiresAt: time.Now().Add(time.Duration(config.TokenExpiration) * time.Second),
	}

	if err := us.SessionRepo.CreateSession(session); err != nil {
		return "", err
	}

	// Update last login time
	requestData := map[string]any{"last_login": time.Now()}
	if err := us.UpdateUser(user.ID.String(), requestData); err != nil {
		return "", err
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

func (us *UserService) UpdateUser(userID string, request map[string]any) error {
	if userID == "" {
		return apperrors.ErrUserIdEmpty
	}

	if password, ok := request["password"].(string); ok && password != "" {
		const minEntropyBits = 64
		if err := passwordvalidator.Validate(password, minEntropyBits); err != nil {
			return err
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		request["password"] = string(hashedPassword)
	}

	if email, ok := request["email"].(string); ok && email != "" {
		if _, err := mail.ParseAddress(email); err != nil {
			return err
		}
		if len(email) > 254 {
			return apperrors.ErrEmailMaxLength
		}
	}

	return us.UserRepo.UpdateUser(userID, request)
}
