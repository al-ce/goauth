package repository

import (
	"time"

	"gorm.io/gorm"

	"gofit/internal/models"
	"gofit/pkg/apperrors"
)

type SessionRepository struct {
	DB *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{DB: db}
}

func (sr *SessionRepository) CreateSession(session *models.Session) error {
	if session == nil {
		return apperrors.ErrSessionIsNil
	}

	// Lookup existing session by token
	var existingSession models.Session
	result := sr.DB.Where("token = ?", session.Token).First(&existingSession)
	if result.Error == nil {
		// Session already exists
		return apperrors.ErrSessionAlreadyExists
	} else if result.Error != gorm.ErrRecordNotFound {
		// Some other error
		return result.Error
	}
	return sr.DB.Create(session).Error
}

func (sr *SessionRepository) GetSessionByToken(token string) (*models.Session, error) {
	if token == "" {
		return nil, apperrors.ErrTokenIsEmpty
	}

	var session models.Session
	result := sr.DB.Where("token = ? AND expires_at > ?", token, time.Now()).First(&session)
	if result.Error != nil {
		return nil, result.Error
	}
	return &session, nil
}

func (sr *SessionRepository) DeleteSessionByToken(token string) error {
	result := sr.DB.Where("token = ?", token).Delete(&models.Session{})
	return result.Error
}

func (sr *SessionRepository) DeleteSessionByUserID(userID string) error {
	result := sr.DB.Where("user_id = ?", userID).Delete(&models.Session{})
	return result.Error
}
