package repository

import (
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"gofit/internal/models"
	"gofit/pkg/apperrors"
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
	r.DB.First(&user, "email = ?", email)

	defaultUUID := uuid.UUID{}
	if user.ID == defaultUUID {
		return nil, apperrors.ErrUserNotFound
	}
	return &user, nil
}

func (r *UserRepository) GetUserByID(userID string) (*models.User, error) {
	var user models.User
	r.DB.First(&user, "ID = ?", userID)

	defaultUUID := uuid.UUID{}
	if user.ID == defaultUUID {
		return nil, apperrors.ErrUserNotFound
	}
	return &user, nil
}

func (r *UserRepository) PermanentlyDeleteUser(userID string) (int64, error) {
	result := r.DB.Unscoped().Where("id = ?", userID).Delete(&models.User{})
	return result.RowsAffected, result.Error
}
