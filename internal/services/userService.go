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

func (us *UserService) RegisterUser(email, password string) error {
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
	return us.ur.RegisterUser(user)
}

func (us *UserService) LoginUser(email, password string) (string, error) {
	if email == "" {
		return "", apperrors.ErrEmailIsEmpty
	}
	if password == "" {
		return "", apperrors.ErrPasswordIsEmpty
	}

	// Lookup user exists
	_, err := us.ur.LookupUser(email)
	if err != nil {
		return "", err
	}

	// Compare password with hash

	// Generate a jwt token

	// Return token string

	return "", nil
}
