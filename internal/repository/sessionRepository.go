package repository

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"godiscauth/internal/models"
	"godiscauth/pkg/apperrors"
)

// SessionRepository represents the entry point into the database for managing
type SessionRepository struct {
	DB *gorm.DB
}

// NewSessionRepository returns a value for the SessionRepository struct
	return &SessionRepository{DB: db}
}

// CreateSession inserts a new session into the `sessions` table
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

// GetUnexpiredSessionByID retrieves a session from the database by sessionID, but ignores any expired sessions
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

// DeleteSessionByID deletes a single session from the database by sessionID
	if token == "" {
		return apperrors.ErrTokenIsEmpty
	}
	result := sr.DB.Where("token = ?", token).Delete(&models.Session{})
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// DeleteSessionsByUserID deletes all sessions associated with a userID from the database
	if userID == "" {
		return apperrors.ErrUserIdEmpty
	}
	result := sr.DB.Where("user_id = ?", userID).Delete(&models.Session{})
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// StartSessionCleanup starts a background goroutine to periodically clean up expired sessions
func (sr *SessionRepository) StartSessionCleanup(interval time.Duration, ctx context.Context) (chan struct{}, error) {
	// Notification channel
	cleanupComplete := make(chan struct{})

	// cf. https://gobyexample.com/tickers
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Info().Msg("session cleanup started")
				result := sr.DB.Where("expires_at < ?", time.Now()).Delete(&models.Session{})
				if result.Error != nil {
					log.Error().Err(result.Error).Msg("Error during session cleanup")
				} else {
					log.Info().Int64("deleted_count", result.RowsAffected).Msg("Session cleanup completed")
				}

				// Signal cleanup completion
				select {
				case cleanupComplete <- struct{}{}:
				default: // Do not block if no receiver is listening
				}

			case <-ctx.Done():
				log.Info().Msg("Stopping session cleanup routine")
				close(cleanupComplete)
				return
			}
		}
	}()

	log.Info().
		Dur("interval", interval).
		Msg("Started session cleanup routine")

	return cleanupComplete, nil
}
