package services

import (
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	passwordvalidator "github.com/wagslane/go-password-validator"
	"golang.org/x/crypto/bcrypt"

	"github.com/al-ce/goauth/internal/models"
	"github.com/al-ce/goauth/internal/repository"
	"github.com/al-ce/goauth/pkg/apperrors"
	"github.com/al-ce/goauth/pkg/config"
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
		return "", apperrors.ErrUserNotFound
	}

	// Deny login if account is locked
	if user.AccountLocked {
		// Unlock account if it is after lockout time
		if user.AccountLockedUntil == nil || time.Now().UTC().After(*user.AccountLockedUntil) {
			err := us.UserRepo.UnlockAccount(user.ID.String())
			if err != nil {
				return "", err
			}
		} else {
			return "", apperrors.ErrAccountIsLocked
		}
	}

	// Lock account on too many failed attempts
	if user.FailedLoginAttempts >= config.MaxLoginAttempts {
		err = us.UserRepo.LockAccount(user.ID.String())
		if err != nil {
			return "", err
		}
		return "", apperrors.ErrAccountIsLocked
	}


	// Validate password
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		// Increment failed login attempts
		err = us.UserRepo.IncrementFailedLogins(user.ID.String())
		if err != nil {
			return "", err
		}
		return "", apperrors.ErrInvalidLogin
	}

	// Generate session ID
	sessionID, signature, err := models.GenerateSessionID()
	if err != nil {
		return "", apperrors.ErrSessionIDGeneration
	}
	sessionToken := sessionID.String() + "." + signature

	// Create session with expiration time (use UTC)
	expiresAt := time.Now().UTC().Add(time.Duration(config.SessionExpiration) * time.Second)
	session, err := models.NewSession(user.ID, sessionID, expiresAt)

	if err := us.SessionRepo.CreateSession(session); err != nil {
		return "", err
	}

	// Update last login time
	requestData := map[string]any{"last_login": time.Now().UTC()}
	if err := us.UpdateUser(user.ID.String(), requestData); err != nil {
		return "", err
	}

	return sessionToken, nil
}

// Logout invalidates a token by deleting its corresponding session
func (us *UserService) Logout(sessionToken string) error {
	if sessionToken == "" {
		return apperrors.ErrSessionIdIsEmpty
	}
	// Split the session token
	parts := strings.Split(sessionToken, ".")
	if len(parts) != 2 {
		return apperrors.ErrInvalidTokenFormat
	}
	sessionID := parts[0]
	parsedID, err := uuid.Parse(sessionID)
	if err != nil {
		return apperrors.ErrInvalidTokenFormat
	}
	return us.SessionRepo.DeleteSessionByID(parsedID)
}

func (us *UserService) LogoutEverywhere(userID string) error {
	if userID == "" {
		return apperrors.ErrUserIdEmpty
	}
	return us.SessionRepo.DeleteSessionsByUserID(userID)
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
func (us *UserService) UpdateUser(userID string, request map[string]any) error {
	if userID == "" {
		return apperrors.ErrUserIdEmpty
	}

	// Check if user exists with this email before attempting update
	if email, ok := request["email"].(string); ok && email != "" {
		user, _ := us.UserRepo.GetUserByEmail(request["email"].(string))
		// If user was found, we have duplicate user
		if user != nil {
			return apperrors.ErrDuplicateEmail
		}
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

// PermanentlyDeleteUser removes the user from the database. This is a permanent operation rather
// than a "deletedAt" flag toggle.
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

// RotateSession generates a new session token for the user and invalidates the old one
// RotateSession creates a new session and replaces the old one
func (us *UserService) RotateSession(oldSessionID uuid.UUID) (string, error) {
	// Check session exists
	oldSession, err := us.SessionRepo.GetUnexpiredSessionByID(oldSessionID)
	if err != nil {
		return "", err
	}

	// Generate new session token with same claims
	newSessionID, signature, err := models.GenerateSessionID()
	if err != nil {
		return "", apperrors.ErrSessionIDGeneration
	}
	newSessionToken := newSessionID.String() + "." + signature

	// Create new session with the new token and expiration time
	expiresAt := time.Now().UTC().Add(time.Duration(config.SessionExpiration) * time.Second)
	newSession, err := models.NewSession(oldSession.UserID, newSessionID, expiresAt)
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
	if err := db.Where("id = ?", oldSessionID).Delete(&models.Session{}).Error; err != nil {
		return "", err
	}

	return newSessionToken, nil
}
