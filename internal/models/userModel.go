package models

import (
	"net/mail"
	"time"

	"github.com/google/uuid"
	passwordvalidator "github.com/wagslane/go-password-validator"
	"golang.org/x/crypto/bcrypt"

	"github.com/al-ce/goauth/pkg/apperrors"
	"github.com/al-ce/goauth/pkg/config"
)

// User represents a user in the `users` table.
type User struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Email               string     `gorm:"type:varchar(255);not null;unique"`
	Password            string     `gorm:"type:text;not null"`
	LastLogin           *time.Time `gorm:"type:timestamp"`
	FailedLoginAttempts int        `gorm:"type:integer;default:0"`
	AccountLocked       bool       `gorm:"type:boolean;default:false"`
	AccountLockedUntil  *time.Time `gorm:"type:timestamp"`
}

// NewUser creates a new User value from an email and password.
func NewUser(email string, password string) (*User, error) {
	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Validate email
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, apperrors.ErrEmailFormat
	}

	if len(email) > 254 {
		return nil, apperrors.ErrEmailMaxLength
	}

	// Enforce minimum password complexity
	if err = passwordvalidator.Validate(password, config.MinEntropyBits); err != nil {
		return nil, apperrors.ErrPasswordComplexity
	}

	return &User{Email: email, Password: string(hash)}, err
}
