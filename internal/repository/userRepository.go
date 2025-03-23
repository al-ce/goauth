package repository

import (
	"gorm.io/gorm"

	"gofit/internal/models"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) UserCreate(u *models.User) error {
	return r.db.Create(u).Error
}

func (r *UserRepository) UsersIndex() ([]models.User, error) {
	var users []models.User
	err := r.db.Find(&users).Error
	return users, err
}
