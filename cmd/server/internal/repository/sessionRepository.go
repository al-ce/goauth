package repository

import (
	"time"

	"gorm.io/gorm"

	"gofit/internal/models"
)

type SessionRepository struct {
	DB *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{DB: db}
}

func (sr *SessionRepository) CreateSession(session *models.Session) error {
	return sr.DB.Create(session).Error
}

func (sr *SessionRepository) GetSessionByToken(token string) (*models.Session, error) {
	var session models.Session
	result := sr.DB.Where("token = ? AND expires_at > ?", token, time.Now()).First(&session)
	if result.Error != nil {
		return nil, result.Error
	}
	return &session, nil
}
