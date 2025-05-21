package repository_test

import (
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

// TestSessionRepository_NewSessionRepository tests creation of SessionRepository
// structs in the `repository` package
func TestSessionRepository_NewSessionRepository(t *testing.T) {
	is := is.New(t)

	testDB := testutils.TestDBSetup()

	t.Run("creates new session repo", func(t *testing.T) {
		sr, err := repository.NewSessionRepository(testDB)
		is.True(sr != nil)
		is.NoErr(err)
	})

	t.Run("returns err with nil db", func(t *testing.T) {
		tx := testDB.Begin()
		defer tx.Rollback()

		sr, err := repository.NewSessionRepository(nil)
		is.Equal(sr, nil)
		is.Equal(err, apperrors.ErrDatabaseIsNil)
	})
}

func TestSessionRepository_CreateSession(t *testing.T) {
	is := is.New(t)

	t.Run("fails on nil session", func(t *testing.T) {
		sr := setupSessionRepository(t)

		err := sr.CreateSession(nil)
		is.Equal(err, apperrors.ErrSessionIsNil)
	})

	t.Run("creates session", func(t *testing.T) {
		sr := setupSessionRepository(t)

		// Register test user
		user := &models.User{
			Email:    "testCreatesSession@test.com",
			Password: "password",
		}
		err := sr.DB.Create(user).Error
		is.NoErr(err)

		session, err := models.NewSession(user.ID, uuid.New(), time.Now().UTC().Add(1*time.Hour))
		is.NoErr(err)

		err = sr.CreateSession(session)
		is.NoErr(err)
	})

	t.Run("fails on duplicate session", func(t *testing.T) {
		sr := setupSessionRepository(t)

		// Register test user
		user := &models.User{
			Email:    "testCreatesSession@test.com",
			Password: "password",
		}
		err := sr.DB.Create(user).Error
		is.NoErr(err)

		sessionID := uuid.New()

		// Insert first session
		sessionOne, err := models.NewSession(user.ID, sessionID, time.Now().UTC().Add(1*time.Hour))
		is.NoErr(err)

		err = sr.CreateSession(sessionOne)
		is.NoErr(err)

		// Insert second session with same ID (expect error)
		sessionTwo, err := models.NewSession(
			user.ID,
			sessionID,
			time.Now().UTC().Add(1*time.Hour),
		)
		is.NoErr(err)

		err = sr.CreateSession(sessionTwo)
		is.Equal(err, apperrors.ErrSessionAlreadyExists)
	})
}

func TestSessionRepository_GetUnexpiredSessionByID(t *testing.T) {
	is := is.New(t)

	t.Run("fails on empty session id", func(t *testing.T) {
		sr := setupSessionRepository(t)

		session, err := sr.GetUnexpiredSessionByID(uuid.Nil)
		is.Equal(session, nil)
		is.Equal(err, apperrors.ErrSessionIdIsEmpty)
	})

	t.Run("fails on non-existing session", func(t *testing.T) {
		sr := setupSessionRepository(t)

		session, err := sr.GetUnexpiredSessionByID(uuid.New())
		is.Equal(session, nil)
		is.Equal(err, gorm.ErrRecordNotFound)
	})

	// Valid session in db can be retrieved
	t.Run("retrieves session by id", func(t *testing.T) {
		sr := setupSessionRepository(t)

		// Register test user
		user := &models.User{
			Email:    "testGetUnexpiredSessionByID@test.com",
			Password: "password",
		}
		err := sr.DB.Create(user).Error
		is.NoErr(err)

		// Insert associated session
		session, err := models.NewSession(user.ID, uuid.New(), time.Now().UTC().Add(1*time.Hour))
		is.NoErr(err)
		err = sr.CreateSession(session)
		is.NoErr(err)

		// Retrieved session has same values we passed in
		retrievedSession, err := sr.GetUnexpiredSessionByID(session.ID)
		is.NoErr(err)
		is.Equal(retrievedSession.UserID, session.UserID)
		is.Equal(retrievedSession.ID, session.ID)
	})
}

func TestSessionRepository_DeleteSessionByID(t *testing.T) {
	is := is.New(t)

	t.Run("fails on empty id", func(t *testing.T) {
		sr := setupSessionRepository(t)

		err := sr.DeleteSessionByID(uuid.Nil)
		is.Equal(err, apperrors.ErrSessionIdIsEmpty)
	})

	t.Run("fails on non-existing session", func(t *testing.T) {
		sr := setupSessionRepository(t)

		err := sr.DeleteSessionByID(uuid.New())
		is.Equal(err, gorm.ErrRecordNotFound)
	})

	t.Run("deletes session by id", func(t *testing.T) {
		sr := setupSessionRepository(t)

		// Register test user
		user := &models.User{
			Email:    "testDeleteSessionByID@test.com",
			Password: "password",
		}
		err := sr.DB.Create(user).Error
		is.NoErr(err)

		// Insert session to delete
		session, err := models.NewSession(user.ID, uuid.New(), time.Now().UTC().Add(1*time.Hour))
		is.NoErr(err)
		err = sr.CreateSession(session)
		is.NoErr(err)

		// Delete session
		err = sr.DeleteSessionByID(session.ID)
		is.NoErr(err)

		// Attempt to retrieve deleted session
		retrievedSession, err := sr.GetUnexpiredSessionByID(session.ID)
		is.Equal(retrievedSession, nil)
		is.Equal(err, gorm.ErrRecordNotFound)
	})
}

func TestSessionRepository_DeleteSessionsByUserID(t *testing.T) {
	is := is.New(t)

	t.Run("fails on non-existing user ID", func(t *testing.T) {
		sr := setupSessionRepository(t)

		err := sr.DeleteSessionsByUserID(uuid.New().String())
		is.Equal(err, gorm.ErrRecordNotFound)
	})

	t.Run("fails on empty user ID", func(t *testing.T) {
		sr := setupSessionRepository(t)

		err := sr.DeleteSessionsByUserID("")
		is.Equal(err, apperrors.ErrUserIdEmpty)
	})

	t.Run("deletes session by user ID", func(t *testing.T) {
		sr := setupSessionRepository(t)

		// Register test user
		user := &models.User{
			Email:    "testDeleteSessionsByUserID@test.com",
			Password: "password",
		}
		err := sr.DB.Create(user).Error
		is.NoErr(err)

		// Insert first session associated with user
		sessionOne, err := models.NewSession(user.ID, uuid.New(), time.Now().UTC().Add(1*time.Hour))
		is.NoErr(err)
		err = sr.CreateSession(sessionOne)
		is.NoErr(err)

		// Insert second session associated with user
		sessionTwo, err := models.NewSession(user.ID, uuid.New(), time.Now().UTC().Add(1*time.Hour))
		is.NoErr(err)
		err = sr.CreateSession(sessionTwo)
		is.NoErr(err)

		// Delete both sessions
		err = sr.DeleteSessionsByUserID(user.ID.String())
		is.NoErr(err)

		// Attempt to retrieve both sessions
		for _, session := range []*models.Session{sessionOne, sessionTwo} {
			retrievedSession, err := sr.GetUnexpiredSessionByID(session.ID)
			if session.UserID == user.ID {
				is.Equal(retrievedSession, nil)
				is.Equal(err, gorm.ErrRecordNotFound)
			} else {
				is.NoErr(err)
				is.Equal(retrievedSession.UserID, user.ID)
			}
		}
	})
}

func setupSessionRepository(t *testing.T) *repository.SessionRepository {
	t.Helper()

	testDB := testutils.TestDBSetup()
	tx := testDB.Begin()
	t.Cleanup(func() { tx.Rollback() })

	sr, err := repository.NewSessionRepository(tx)
	if err != nil {
		t.Fatalf("failed to create session repository: %v", err)
	}
	return sr
}
