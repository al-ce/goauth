package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"godiscauth/internal/models"
	"godiscauth/pkg/apperrors"
)

// SessionRepository represents the entry point into the database for managing
// the `sessions` table
type SessionRepository struct {
	DB *gorm.DB
}

// NewSessionRepository returns a value for the SessionRepository struct
func NewSessionRepository(db *gorm.DB) (*SessionRepository, error) {
	if db == nil {
		return nil, apperrors.ErrDatabaseIsNil
	}
	return &SessionRepository{DB: db}, nil
}

// CreateSession inserts a new session into the `sessions` table
func (sr *SessionRepository) CreateSession(session *models.Session) error {
	if session == nil {
		return apperrors.ErrSessionIsNil
	}
	// Lookup existing session by ID
	var existingSession models.Session
	result := sr.DB.Where("id = ?", session.ID).First(&existingSession)
	if result.Error == nil {
		// Session already exists
		return apperrors.ErrSessionAlreadyExists
	} else if result.Error != gorm.ErrRecordNotFound { // RecordNotFound is what we want to prevent duplicates
		// Some other error
		return result.Error
	}
	return sr.DB.Create(session).Error
}

// GetUnexpiredSessionByID retrieves a session from the database by sessionID, but ignores any expired sessions
func (sr *SessionRepository) GetUnexpiredSessionByID(sessionID uuid.UUID) (*models.Session, error) {
	if sessionID == uuid.Nil {
		return nil, apperrors.ErrSessionIdIsEmpty
	}
	var session models.Session
	result := sr.DB.Where("id = ? AND expires_at > ?", sessionID, time.Now().UTC()).First(&session)
	if result.Error != nil {
		return nil, result.Error
	}
	return &session, nil
}

// DeleteSessionByID deletes a single session from the database by sessionID
func (sr *SessionRepository) DeleteSessionByID(sessionID uuid.UUID) error {
	if sessionID == uuid.Nil {
		return apperrors.ErrSessionIdIsEmpty
	}
	result := sr.DB.Where("id = ?", sessionID).Delete(&models.Session{})
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// DeleteSessionsByUserID deletes all sessions associated with a userID from the database
func (sr *SessionRepository) DeleteSessionsByUserID(userID string) error {
	if userID == "" {
		return apperrors.ErrUserIdEmpty
	}
	result := sr.DB.Where("user_id = ?", userID).Delete(&models.Session{})
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}
