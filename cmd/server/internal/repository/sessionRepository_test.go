package repository_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/matryer/is"
	"gorm.io/gorm"

	"gofit/internal/models"
	"gofit/internal/repository"
	"gofit/internal/testutils"
	"gofit/pkg/apperrors"
)

func TestSessionRepository_CreateSession(t *testing.T) {
	testDB := testutils.TestDBSetup()
	is := is.New(t)
	tx := testDB.Begin()
	defer tx.Rollback()

	sr := repository.NewSessionRepository(tx)

	t.Run("fails on nil session", func(t *testing.T) {
		err := sr.CreateSession(nil)
		is.Equal(err, apperrors.ErrSessionIsNil)
	})

	t.Run("creates session", func(t *testing.T) {
		session, err := models.NewSession(uuid.New(), uuid.New().String(), time.Now().Add(1*time.Hour))
		is.NoErr(err)

		err = sr.CreateSession(session)
		is.NoErr(err)
	})

	t.Run("fails on duplicate session", func(t *testing.T) {
		sessionOne, err := models.NewSession(uuid.New(), uuid.New().String(), time.Now().Add(1*time.Hour))
		is.NoErr(err)

		err = sr.CreateSession(sessionOne)
		is.NoErr(err)
		sessionTwo, err := models.NewSession(
			sessionOne.UserID,
			sessionOne.Token,
			time.Now().Add(1*time.Hour),
		)
		is.NoErr(err)

		err = sr.CreateSession(sessionTwo)
		is.Equal(err, apperrors.ErrSessionAlreadyExists)
	})

	t.Run("fails on nil session", func(t *testing.T) {
		err := sr.CreateSession(nil)
		is.Equal(err, apperrors.ErrSessionIsNil)
	})
}

func TestSessionRepository_GetSessionByToken(t *testing.T) {
	testDB := testutils.TestDBSetup()
	is := is.New(t)
	tx := testDB.Begin()
	defer tx.Rollback()

	sr := repository.NewSessionRepository(tx)

	t.Run("fails on empty token", func(t *testing.T) {
		session, err := sr.GetSessionByToken("")
		is.Equal(session, nil)
		is.Equal(err, apperrors.ErrTokenIsEmpty)
	})

	t.Run("fails on non-existing token", func(t *testing.T) {
		session, err := sr.GetSessionByToken(uuid.New().String())
		is.Equal(session, nil)
		is.Equal(err, gorm.ErrRecordNotFound)
	})

	t.Run("retrieves session by token", func(t *testing.T) {
		session, err := models.NewSession(uuid.New(), uuid.New().String(), time.Now().Add(1*time.Hour))
		is.NoErr(err)

		err = sr.CreateSession(session)
		is.NoErr(err)

		retrievedSession, err := sr.GetSessionByToken(session.Token)
		is.NoErr(err)
		is.Equal(retrievedSession.UserID, session.UserID)
		is.Equal(retrievedSession.Token, session.Token)
	})
}
