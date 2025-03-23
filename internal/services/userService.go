package services

import (
	"gofit/internal/models"
	"gofit/internal/repository"
	"gofit/pkg/apperrors"
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
	if email == "" {
		return apperrors.ErrEmailIsEmpty
	}
	if password == "" {
		return apperrors.ErrPasswordIsEmpty
	}

	user, err := models.NewUser(email, password)
	if err != nil {
		return err
	}
	return s.ur.RegisterUser(user)
}
