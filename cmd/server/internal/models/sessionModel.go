package models

import (
	"time"

	"github.com/google/uuid"

	"gofit/pkg/apperrors"
)

type Session struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Token     string    `gorm:"type:text;not null;unique"`
	ExpiresAt time.Time `gorm:"type:timestamp;not null"`
	CreatedAt time.Time `gorm:"type:timestamp;not null;default:now()"`
}

func NewSession(userID uuid.UUID, token string, expiresAt time.Time) (*Session, error) {
	if userID == uuid.Nil {
		return nil, apperrors.ErrUserIdEmpty
	}
	if token == "" {
		return nil, apperrors.ErrTokenIsEmpty
	}

	return &Session{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}, nil
}
