package repository

import (
	"errors"

	"github.com/google/uuid"
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

func (r *UserRepository) LookupUser(email string) (*models.User, error) {
	var user models.User
	r.db.First(&user, "email = ?", email)

	defaultUUID := uuid.UUID{}
	if user.ID == defaultUUID {
		return nil, errors.New("User not found")
	}
	return &user, nil
}
