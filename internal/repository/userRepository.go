package repository

import (
	"errors"

	"gorm.io/gorm"

	"gofit/internal/models"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) RegisterUser(u *models.User) error {
	if u == nil {
		return errors.New("user is nil")
	}

	if u.Email == "" {
		return errors.New("email is empty")
	}
	if u.Password == "" {
		return errors.New("password is empty")
	}

	// TODO: validate email format and pw strength

	return r.db.Create(u).Error
}
