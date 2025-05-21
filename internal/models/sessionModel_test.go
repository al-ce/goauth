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

// TestSessionModel_NewSession tests new Session creation in the `models`
// package
func TestSessionModel_NewSession(t *testing.T) {
	is := is.New(t)

	// Valid uuid and non-empty session id should return non-nil Session value, nil error
	t.Run("new valid session", func(t *testing.T) {
		session, err := models.NewSession(uuid.New(), uuid.New(), time.Now().UTC().Add(24*time.Hour))
		is.True(session != nil)
		is.NoErr(err)
	})

	t.Run("fails when user ID is empty", func(t *testing.T) {
		_, err := models.NewSession(uuid.Nil, uuid.New(), time.Now().UTC().Add(24*time.Hour))
		is.Equal(err, apperrors.ErrUserIdEmpty)
	})

	t.Run("fails when session is empty", func(t *testing.T) {
		_, err := models.NewSession(uuid.New(), uuid.Nil, time.Now().UTC().Add(24*time.Hour))
		is.Equal(err, apperrors.ErrSessionIdIsEmpty)
	})
	t.Run("fails when expiration time is empty", func(t *testing.T) {
		_, err := models.NewSession(uuid.New(), uuid.New(), time.Time{})
		is.Equal(err, apperrors.ErrExpiresAtIsEmpty)
	})
}

// TestSessionModel_CascadeToSessions tests that deleting a user in the
// database scrubs any associated sessions by OnDelete-Cascade
func TestSessionModel_CascadeToSessions(t *testing.T) {
	testDB := testutils.TestDBSetup()
	is := is.New(t)

	t.Run("user sessions are deleted when user is deleted", func(t *testing.T) {
		tx := testDB.Begin()
		defer tx.Rollback()

		// Create test user for sessions (required foreign key reference)
		userName := "testCascadeDeleteSessions@test.com"
		testUser, err := models.NewUser(userName, testutils.TestingPassword)
		is.NoErr(err)

		err = tx.Create(testUser).Error
		is.NoErr(err)

		// Create test sessions referencing the test user
		for range 3 {
			session, err := models.NewSession(
				testUser.ID,
				uuid.New(),
				time.Now().UTC().Add(1*time.Hour),
			)
			is.NoErr(err)
			err = tx.Create(session).Error
			is.NoErr(err)
		}

		// Check sessions are created
		var count int64
		tx.Model(&models.Session{}).Where("user_id = ?", testUser.ID).Count(&count)
		is.Equal(count, int64(3))

		// Delete the user
		result := tx.Unscoped().Where("id = ?", testUser.ID).Delete(&models.User{})
		is.Equal(result.RowsAffected, int64(1))

		// Expect all sessions deleted
		tx.Model(&models.Session{}).Where("user_id = ?", testUser.ID).Count(&count)
		is.Equal(count, int64(0))
	})
}
