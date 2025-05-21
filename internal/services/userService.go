package services

import (
	"net/mail"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	passwordvalidator "github.com/wagslane/go-password-validator"
	"golang.org/x/crypto/bcrypt"

	"godiscauth/internal/models"
	"godiscauth/internal/repository"
	"godiscauth/pkg/apperrors"
	"godiscauth/pkg/config"
)

// UserService is a struct that contains the repositories needed for user-related operations
type UserService struct {
	UserRepo    *repository.UserRepository
	SessionRepo *repository.SessionRepository
}

// NewUserService returns a value of type UserService
func NewUserService(ur *repository.UserRepository, sr *repository.SessionRepository) (*UserService, error) {
	if ur == nil {
		return nil, apperrors.ErrUserRepoIsNil
	}
	if sr == nil {
		return nil, apperrors.ErrSessionRepoIsNil
	}
	return &UserService{
		UserRepo:    ur,
		SessionRepo: sr,
	}, nil
}

// RegisterUser mediates the new User value creation and the insertion of a user into the database
func (us *UserService) RegisterUser(email, password string) error {
	// Check for empty fields
	if email == "" {
		return apperrors.ErrEmailIsEmpty
	}
	if password == "" {
		return apperrors.ErrPasswordIsEmpty
	}

	// Check if user exists before attempting registration
	// If user was found, we have duplicate user
	user, _ := us.UserRepo.GetUserByEmail(email)
	if user != nil {
		return apperrors.ErrDuplicateEmail
	}

	user, err := models.NewUser(email, password)
	if err != nil {
		return err
	}
	return us.UserRepo.RegisterUser(user)
}

// LoginUser authenticates a registered user and creates an associated session
func (us *UserService) LoginUser(email, password string) (string, error) {
	// Check for empty fields
	var err error
	if email == "" {
		err := apperrors.ErrEmailIsEmpty
		return "", err
	}
	if password == "" {
		err := apperrors.ErrPasswordIsEmpty
		return "", err
	}

	// Check if user exists
	user, err := us.UserRepo.GetUserByEmail(email)
	if err != nil {
		return "", err
	}

	// Deny login if account is locked
	if user.AccountLocked {
		return "", apperrors.ErrAccountIsLocked
	}

	// Return ErrInvalidLogin on bad password
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		// Increment failed login attempts
		err = us.UserRepo.IncrementFailedLogins(user.ID.String())
		if err != nil {
			return "", err
		}
		// Lock account on too many failed attempts
		if user.FailedLoginAttempts == config.MaxLoginAttempts-1 {
			err = us.UserRepo.LockAccount(user.ID.String())
			if err != nil {
				return "", err
			}
		}
		return "", apperrors.ErrInvalidLogin
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv(config.JwtCookieName)))
	if err != nil {
		return "", apperrors.ErrTokenGeneration
	}

	// Create session with expiration time
	expiresAt := time.Now().Add(time.Duration(config.TokenExpiration) * time.Second)
	session, err := models.NewSession(user.ID, tokenString, expiresAt)

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

// Logout invalidates a token by deleting its corresponding session
	return us.SessionRepo.DeleteSessionByToken(token)
}

func (us *UserService) LogoutEverywhere(userID string) error {
	if userID == "" {
		return apperrors.ErrUserIdEmpty
	}
	return us.SessionRepo.DeleteSessionByUserID(userID)
}

func (us *UserService) GetUserProfile(userID string) (*models.UserProfile, error) {
	if userID == "" {
		return nil, apperrors.ErrUserIdEmpty
	}
	user, err := us.UserRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Why not just return a User object with every other field set to nil?
	// The User object contains sensitive information like password hash.
	// Rather than trust ourselves to never expose that, we create a new struct
	userProfile := &models.UserProfile{
		Email:     user.Email,
		LastLogin: user.LastLogin,
	}
	return userProfile, nil
}

// UpdateUser mediates the query for an update API request and the update of a user in the database
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
		if err := passwordvalidator.Validate(password, config.MinEntropyBits); err != nil {
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

// RotateSession generates a new JWT token for the user and invalidates the old one
// RotateSession creates a new session and replaces the old one
func (us *UserService) RotateSession(oldToken string) (string, error) {
	// Check session exists
	oldSession, err := us.SessionRepo.GetUnexpiredSessionByToken(oldToken)
	if err != nil {
		return "", err
	}

	// Generate new JWT token with same claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": oldSession.UserID.String(),
		"exp": time.Now().Add(time.Duration(config.TokenExpiration) * time.Second).Unix(),
		"jti": uuid.New().String(), // JWT ID - making the token unique
	})
	newToken, err := token.SignedString([]byte(os.Getenv(config.JwtCookieName)))
	if err != nil {
		return "", apperrors.ErrTokenGeneration
	}

	// Create new session with the new token and expiration time
	expiresAt := time.Now().Add(time.Duration(config.TokenExpiration) * time.Second)
	newSession, err := models.NewSession(oldSession.UserID, newToken, expiresAt)
	if err != nil {
		return "", err
	}

	// Use the existing database connection/transaction from the repository
	db := us.SessionRepo.DB

	// Insert new session into the database
	if err := db.Create(newSession).Error; err != nil {
		return "", err
	}

	// Delete old session
	if err := db.Where("token = ?", oldToken).Delete(&models.Session{}).Error; err != nil {
		return "", err
	}

	return newToken, nil
}
