package models

import (
	"net/mail"
	"time"

	"github.com/google/uuid"
	passwordvalidator "github.com/wagslane/go-password-validator"
	"golang.org/x/crypto/bcrypt"

	"goauth/pkg/apperrors"
	"goauth/pkg/config"
)

type User struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Email               string     `gorm:"type:varchar(255);not null;unique"`
	Password            string     `gorm:"type:text;not null"`
	LastLogin           *time.Time `gorm:"type:timestamp"`
	FailedLoginAttempts int        `gorm:"type:integer;default:0"`
	AccountLocked       bool       `gorm:"type:boolean;default:false"`
	AccountLockedUntil  *time.Time `gorm:"type:timestamp"`
}

func NewUser(email string, password string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, err
	}
	if len(email) > 254 {
		return nil, apperrors.ErrEmailMaxLength
	}

	if err = passwordvalidator.Validate(password, config.MinEntropyBits); err != nil {
		return nil, err
	}

	return &User{Email: email, Password: string(hash)}, err
}
