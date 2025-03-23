package repository

import (
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"gofit/internal/models"
	"gofit/pkg/apperrors"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) RegisterUser(u *models.User) error {
	if u == nil {
		return apperrors.ErrUserIsNil
	}

	if u.Email == "" {
		return apperrors.ErrEmailIsEmpty
	}
	if u.Password == "" {
		return apperrors.ErrPasswordIsEmpty
	}

	// TODO: validate email format and pw strength

	err := r.db.Create(u).Error
	if err != nil && strings.Contains(err.Error(), `duplicate key value violates unique constraint "users_pkey"`) {
		return apperrors.ErrDuplicateEmail
	}
	return err
}

func (r *UserRepository) LookupUser(email string) (*models.User, error) {
	var user models.User
	r.db.First(&user, "email = ?", email)

	defaultUUID := uuid.UUID{}
	if user.ID == defaultUUID {
		return nil, apperrors.ErrUserNotFound
	}
	return &user, nil
}
