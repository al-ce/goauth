package models

import (
	"time"

	"github.com/google/uuid"

	"godiscauth/pkg/apperrors"
	"godiscauth/pkg/config"
)

// Session represents a session in the `sessions` table
type Session struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	User      *User     `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE;"`
	ExpiresAt time.Time `gorm:"type:timestamp;not null"`
	CreatedAt time.Time `gorm:"type:timestamp;not null;default:now()"`
}

// NewSession creates a new Session value from a user id, a session id, and an expiration time
func NewSession(userID uuid.UUID, sessionID uuid.UUID, expiresAt time.Time) (*Session, error) {
	if userID == uuid.Nil {
		return nil, apperrors.ErrUserIdEmpty
	}
	if sessionID == uuid.Nil {
		return nil, apperrors.ErrSessionIdIsEmpty
	}
	if expiresAt.IsZero() {
		return nil, apperrors.ErrExpiresAtIsEmpty
	}

	return &Session{
		UserID:    userID,
		ID:        sessionID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}, nil
}
