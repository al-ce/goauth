package models

import (
	"time"

	"github.com/google/uuid"

	"goauth/pkg/apperrors"
)

type Session struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	User      *User     `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE;"`
	Token     string    `gorm:"type:text;not null;unique"`
	ExpiresAt time.Time `gorm:"type:timestamp;not null"`
	CreatedAt time.Time `gorm:"type:timestamp;not null;default:now()"`
}

// NewSession creates a new Session value from a user id, a token string, and
// an expiration time
func NewSession(userID uuid.UUID, token string, expiresAt time.Time) (*Session, error) {
	if userID == uuid.Nil {
		return nil, apperrors.ErrUserIdEmpty
	}
	if token == "" {
		return nil, apperrors.ErrTokenIsEmpty
	}
	if expiresAt.IsZero() {
		return nil, apperrors.ErrExpiresAtIsEmpty
	}

	return &Session{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}, nil
}
