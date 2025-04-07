package models_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/matryer/is"

	"goauth/internal/models"
	"goauth/pkg/apperrors"
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
}
