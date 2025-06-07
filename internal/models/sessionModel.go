package models

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/al-ce/goauth/pkg/apperrors"
	"github.com/al-ce/goauth/pkg/config"
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
		ExpiresAt: expiresAt.UTC(),
		CreatedAt: time.Now().UTC(),
	}, nil
}

// GenerateSessionID creates a new random session ID
func GenerateSessionID() (uuid.UUID, string, error) {
	sessionID := uuid.New()
	signature := createHMAC(sessionID.String())
	return sessionID, signature, nil
}

// ValidateSessionID verifies that a session ID matches its signature
func ValidateSessionID(sessionID uuid.UUID, signature string) bool {
	expectedSignature := createHMAC(sessionID.String())
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// createHMAC generates an HMAC signature for a session ID
// Ref: https://www.okta.com/identity-101/hmac/
func createHMAC(sessionID string) string {
	h := hmac.New(sha256.New, []byte(os.Getenv(config.SessionKey)))
	h.Write([]byte(sessionID))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
