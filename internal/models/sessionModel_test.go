package models_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/matryer/is"

	"godiscauth/internal/models"
	"godiscauth/internal/testutils"
	"godiscauth/pkg/apperrors"
)

func TestSessionModel_NewSession(t *testing.T) {
	is := is.New(t)

	t.Run("Create a new session", func(t *testing.T) {
		_, err := models.NewSession(uuid.New(), "test-token", time.Now().Add(24*time.Hour))
		is.NoErr(err)
	})

	t.Run("fails when user ID is empty", func(t *testing.T) {
		_, err := models.NewSession(uuid.Nil, "test-token", time.Now().Add(24*time.Hour))
		is.Equal(err, apperrors.ErrUserIdEmpty)
	})

	t.Run("fails when token is empty", func(t *testing.T) {
		_, err := models.NewSession(uuid.New(), "", time.Now().Add(24*time.Hour))
		is.Equal(err, apperrors.ErrTokenIsEmpty)
	})
	t.Run("fails when expiration time is empty", func(t *testing.T) {
		_, err := models.NewSession(uuid.New(), "test-token", time.Time{})
		is.Equal(err, apperrors.ErrExpiresAtIsEmpty)
	})
}

func TestSessionModel_CascadeToSessions(t *testing.T) {
	testDB := testutils.TestDBSetup()
	is := is.New(t)

	t.Run("user sessions are deleted when user is deleted", func(t *testing.T) {
		tx := testDB.Begin()
		defer tx.Rollback()

		sr := repository.NewSessionRepository(tx)

		user, err := h.CreateTestUser(tx, "testCascadeDeleteSessions@test.com")
		is.NoErr(err)

		// Create test sessions
		for range 3 {
			session, err := models.NewSession(
				user.ID,
				uuid.New().String(),
				time.Now().Add(1*time.Hour),
			)
			is.NoErr(err)
			err = sr.CreateSession(session)
			is.NoErr(err)
		}

		// Check sessions are created
		var count int64
		tx.Model(&models.Session{}).Where("user_id = ?", user.ID).Count(&count)
		is.Equal(count, int64(3))

		// Delete the user
		result := tx.Unscoped().Where("id = ?", user.ID).Delete(&models.User{})
		is.Equal(result.RowsAffected, int64(1))

		// Expect all sessions deleted
		tx.Model(&models.Session{}).Where("user_id = ?", user.ID).Count(&count)
		is.Equal(count, int64(0))
	})
}
