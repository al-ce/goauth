package repository

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"goauth/internal/models"
	"goauth/pkg/apperrors"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (ur *UserRepository) RegisterUser(u *models.User) error {
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
	if err != nil && strings.Contains(err.Error(), `duplicate key value violates unique constraint "users_pkey"`) {
		return apperrors.ErrDuplicateEmail
	}
	return err
}

func (r *UserRepository) LookupUser(email string) (*models.User, error) {
	var user models.User

	result := r.DB.First(&user, "email = ?", email)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, result.Error
	}

	return &user, nil
}

func (r *UserRepository) GetUserByID(userID string) (*models.User, error) {
	var user models.User
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	result := r.DB.First(&user, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, result.Error
	}

	return &user, nil
}

func (r *UserRepository) PermanentlyDeleteUser(userID string) (int64, error) {
	result := r.DB.Unscoped().Where("id = ?", userID).Delete(&models.User{})
	return result.RowsAffected, result.Error
}

func (r *UserRepository) UpdateUser(userID string, request map[string]any) error {
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
		return apperrors.ErrNoChangesMade
	}

	return nil
}
