package services

import (
	"golang.org/x/crypto/bcrypt"

	"gofit/internal/models"
	"gofit/internal/repository"
)

type UserService struct {
	ur *repository.UserRepository
}

func NewUserService(ur *repository.UserRepository) *UserService {
	return &UserService{
		ur: ur,
	}
}

func (s *UserService) RegisterUser(email, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := models.User{Email: email, Password: string(hash)}
	return s.ur.RegisterUser(&user)
}
