package repository_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/matryer/is"
	"gorm.io/gorm"

	"godiscauth/internal/models"
	"godiscauth/internal/repository"
	"godiscauth/internal/testutils"
	"godiscauth/pkg/apperrors"
)

func TestSessionRepository_CreateSession(t *testing.T) {
	testDB := testutils.TestDBSetup()
	is := is.New(t)
	t.Run("fails on nil session", func(t *testing.T) {
		tx := testDB.Begin()
		defer tx.Rollback()
		sr := repository.NewSessionRepository(tx)

		err := sr.CreateSession(nil)
		is.Equal(err, apperrors.ErrSessionIsNil)
	})

	t.Run("creates session", func(t *testing.T) {
		tx := testDB.Begin()
		defer tx.Rollback()
		sr := repository.NewSessionRepository(tx)

		user, err := h.CreateTestUser(tx, "testCreateSession@test.com")
		is.NoErr(err)

		session, err := models.NewSession(user.ID, uuid.New().String(), time.Now().Add(1*time.Hour))
		is.NoErr(err)

		err = sr.CreateSession(session)
		is.NoErr(err)
	})

	t.Run("fails on duplicate session", func(t *testing.T) {
		tx := testDB.Begin()
		defer tx.Rollback()
		sr := repository.NewSessionRepository(tx)

		user, err := h.CreateTestUser(tx, "testCreateSession@test.com")
		is.NoErr(err)

		tokenStr := uuid.New().String()

		sessionOne, err := models.NewSession(user.ID, tokenStr, time.Now().Add(1*time.Hour))
		is.NoErr(err)

		err = sr.CreateSession(sessionOne)
		is.NoErr(err)

		sessionTwo, err := models.NewSession(
			user.ID,
			tokenStr,
			time.Now().Add(1*time.Hour),
		)
		is.NoErr(err)

		err = sr.CreateSession(sessionTwo)
		is.Equal(err, apperrors.ErrSessionAlreadyExists)
	})
}

func TestSessionRepository_GetSessionByToken(t *testing.T) {
	testDB := testutils.TestDBSetup()
	is := is.New(t)

	t.Run("fails on empty token", func(t *testing.T) {
		tx := testDB.Begin()
		defer tx.Rollback()
		sr := repository.NewSessionRepository(tx)

		session, err := sr.GetUnexpiredSessionByToken("")
		is.Equal(session, nil)
		is.Equal(err, apperrors.ErrTokenIsEmpty)
	})

	t.Run("fails on non-existing token", func(t *testing.T) {
		tx := testDB.Begin()
		defer tx.Rollback()
		sr := repository.NewSessionRepository(tx)

		session, err := sr.GetUnexpiredSessionByToken(uuid.New().String())
		is.Equal(session, nil)
		is.Equal(err, gorm.ErrRecordNotFound)
	})

	t.Run("retrieves session by token", func(t *testing.T) {
		tx := testDB.Begin()
		defer tx.Rollback()
		sr := repository.NewSessionRepository(tx)

		user, err := h.CreateTestUser(tx, "testGetSession@test.com")
		is.NoErr(err)

		session, err := models.NewSession(user.ID, uuid.New().String(), time.Now().Add(1*time.Hour))
		is.NoErr(err)

		err = sr.CreateSession(session)
		is.NoErr(err)

		retrievedSession, err := sr.GetUnexpiredSessionByToken(session.Token)
		is.NoErr(err)
		is.Equal(retrievedSession.UserID, session.UserID)
		is.Equal(retrievedSession.Token, session.Token)
	})
}

func TestSessionRepository_DeleteSessionByToken(t *testing.T) {
	testDB := testutils.TestDBSetup()
	is := is.New(t)
	t.Run("deletes session by token", func(t *testing.T) {
		tx := testDB.Begin()
		defer tx.Rollback()
		sr := repository.NewSessionRepository(tx)

		user, err := h.CreateTestUser(tx, "testDeleteSessionByToken@test.com")
		is.NoErr(err)

		session, err := models.NewSession(user.ID, uuid.New().String(), time.Now().Add(1*time.Hour))
		is.NoErr(err)

		err = sr.CreateSession(session)
		is.NoErr(err)

		err = sr.DeleteSessionByToken(session.Token)
		is.NoErr(err)

		retrievedSession, err := sr.GetUnexpiredSessionByToken(session.Token)
		is.Equal(retrievedSession, nil)
		is.Equal(err, gorm.ErrRecordNotFound)
	})

	t.Run("fails on non-existing token", func(t *testing.T) {
		tx := testDB.Begin()
		defer tx.Rollback()
		sr := repository.NewSessionRepository(tx)

		err := sr.DeleteSessionByToken(uuid.New().String())
		is.Equal(err, gorm.ErrRecordNotFound)
	})

	t.Run("fails on empty token", func(t *testing.T) {
		tx := testDB.Begin()
		defer tx.Rollback()
		sr := repository.NewSessionRepository(tx)

		err := sr.DeleteSessionByToken("")
		is.Equal(err, apperrors.ErrTokenIsEmpty)
	})
}

func TestSessionRepository_DeleteSessionByUserID(t *testing.T) {
	testDB := testutils.TestDBSetup()
	is := is.New(t)
	tx := testDB.Begin()
	defer tx.Rollback()

	sr := repository.NewSessionRepository(tx)

	t.Run("deletes session by user ID", func(t *testing.T) {
		user, err := h.CreateTestUser(tx, "testDeleteSession1@test.com")
		is.NoErr(err)

		sessionOne, err := models.NewSession(user.ID, uuid.New().String(), time.Now().Add(1*time.Hour))
		is.NoErr(err)
		err = sr.CreateSession(sessionOne)
		is.NoErr(err)

		sessionTwo, err := models.NewSession(user.ID, uuid.New().String(), time.Now().Add(1*time.Hour))
		is.NoErr(err)
		err = sr.CreateSession(sessionTwo)
		is.NoErr(err)

		err = sr.DeleteSessionByUserID(user.ID.String())
		is.NoErr(err)

		for _, session := range []*models.Session{sessionOne, sessionTwo} {
			retrievedSession, err := sr.GetUnexpiredSessionByToken(session.Token)
			if session.UserID == user.ID {
				is.Equal(retrievedSession, nil)
				is.Equal(err, gorm.ErrRecordNotFound)
			} else {
				is.NoErr(err)
				is.Equal(retrievedSession.UserID, user.ID)
			}
		}
	})

	t.Run("fails on non-existing user ID", func(t *testing.T) {
		err := sr.DeleteSessionByUserID(uuid.New().String())
		is.Equal(err, gorm.ErrRecordNotFound)
	})

	t.Run("fails on empty user ID", func(t *testing.T) {
		err := sr.DeleteSessionByUserID("")
		is.Equal(err, apperrors.ErrUserIdEmpty)
	})
}

func TestSessionRepository_StartSessionCleanup(t *testing.T) {
	testDB := testutils.TestDBSetup()
	is := is.New(t)
	tx := testDB.Begin()
	defer tx.Rollback()
	sr := repository.NewSessionRepository(tx)
	user, err := h.CreateTestUser(tx, "testSessionCleanup@test.com")
	is.NoErr(err)

	// Create two sessions, one that will expire before cleanup starts,
	// another that won't expire and shouldn't be cleaned up
	sessionToClean, err := models.NewSession(user.ID, uuid.New().String(), time.Now().Add(-time.Hour))
	is.NoErr(err)
	err = sr.CreateSession(sessionToClean)
	is.NoErr(err)

	sessionToRemain, err := models.NewSession(user.ID, uuid.New().String(), time.Now().Add(time.Hour))
	is.NoErr(err)
	err = sr.CreateSession(sessionToRemain)
	is.NoErr(err)

	// Start cleanup with a short interval
	ctx := t.Context()
	cleanupComplete, err := sr.StartSessionCleanup(time.Millisecond, ctx)
	is.NoErr(err)

	select {
	case <-cleanupComplete:
	case <-time.After(2 * time.Second):
		t.Fatal("cleanup took too long, aborting")
	}

	var cleanedSession models.Session
	err = tx.Where("id = ?", sessionToClean.ID).First(&cleanedSession).Error
	is.True(errors.Is(err, gorm.ErrRecordNotFound)) // Should not find the cleaned session

	var remainingSession models.Session
	err = tx.Where("id = ?", sessionToRemain.ID).First(&remainingSession).Error
	is.NoErr(err)
}
