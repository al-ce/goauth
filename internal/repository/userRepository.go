package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"godiscauth/internal/models"
	"godiscauth/pkg/apperrors"
	"godiscauth/pkg/config"
)

// UserRepository represents the entry point into the database for managing the `users` table
type UserRepository struct {
	DB *gorm.DB
}

// NewUserRepository returns a value for the UserRepository struct
func NewUserRepository(db *gorm.DB) (*UserRepository, error) {
	if db == nil {
		return nil, apperrors.ErrDatabaseIsNil
	}
	return &UserRepository{DB: db}, nil
}

// RegisterUser inserts a new user into the `users` table
func (ur *UserRepository) RegisterUser(u *models.User) error {
	// Validate user
	if u == nil {
		return apperrors.ErrUserIsNil
	}
	if u.Email == "" {
		return apperrors.ErrEmailIsEmpty
	}
	if u.Password == "" {
		return apperrors.ErrPasswordIsEmpty
	}

	err := ur.DB.Create(u).Error
	return err
}

// GetUserByEmail gets a user in the database by email
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	if email == "" {
		return nil, apperrors.ErrEmailIsEmpty
	}

	var user models.User

	result := r.DB.First(&user, "email = ?", email)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

// GetUserByID gets a user in the database by user ID (uuid as string)
func (r *UserRepository) GetUserByID(userID string) (*models.User, error) {
	if userID == "" {
		return nil, apperrors.ErrUserIdEmpty
	}

	var user models.User
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	result := r.DB.First(&user, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

// PermanentlyDeleteUser removes existing users from the database by ID
func (r *UserRepository) PermanentlyDeleteUser(userID string) (int64, error) {
	if userID == "" {
		return 0, apperrors.ErrUserIdEmpty
	}
	result := r.DB.Unscoped().Where("id = ?", userID).Delete(&models.User{})
	return result.RowsAffected, result.Error
}

// UpdateUser updates a user in the database by usedID, expecting a decoded request to pass updated fields
func (r *UserRepository) UpdateUser(userID string, request map[string]any) error {
	if userID == "" {
		return apperrors.ErrUserIdEmpty
	}
	var exists bool
	r.DB.Model(&models.User{}).Select("1").Where("id = ?", userID).First(&exists)
	if !exists {
		return apperrors.ErrUserNotFound
	}

	result := r.DB.Model(&models.User{}).Where("id = ?", userID).Updates(request)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return apperrors.ErrCouldNotUpdateUser
	}

	return nil
}

// IncrementFailedLogins increments failed login attempts and locks account every
// `config.MaxLoginAttempts` failed attempts.
func (r *UserRepository) IncrementFailedLogins(userID string) error {
	// Validate user ID
	if userID == "" {
		return apperrors.ErrUserIdEmpty
	}

	// Lookup user
	user, err := r.GetUserByID(userID)
	if err != nil {
		return apperrors.ErrUserNotFound
	}

	// Increment login attempts
	result := r.DB.Model(&models.User{}).Where("id = ?", user.ID).
		Updates(map[string]any{"failed_login_attempts": user.FailedLoginAttempts + 1})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		// Unsure under what circumstances this could happen, but handle the case anyway
		return apperrors.ErrCouldNotIncrementFailedLogins
	}

	return nil
}

// LockAccount locks a user account until the time spec'd in `config`
func (r *UserRepository) LockAccount(userID string) error {
	// Validate user ID
	if userID == "" {
		return apperrors.ErrUserIdEmpty
	}

	result := r.DB.Model(&models.User{}).Where("id = ?", userID).
		Updates(map[string]any{
			"account_locked":       true,
			"account_locked_until": time.Now().Add(config.AccountLockoutLength * time.Minute),
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrUserNotFound
	}

	return nil
}
