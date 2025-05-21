package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"godiscauth/internal/models"
	"godiscauth/internal/repository"
	"godiscauth/internal/services"
	"godiscauth/pkg/apperrors"
	"godiscauth/pkg/config"
)

type AuthMiddleware struct {
	UserRepo    *repository.UserRepository
	SessionRepo *repository.SessionRepository
}

func NewAuthMiddleware(db *gorm.DB) (*AuthMiddleware, error) {
	if db == nil {
		return nil, apperrors.ErrDatabaseIsNil
	}
	ur, err := repository.NewUserRepository(db)
	if err != nil {
		return nil, err
	}
	sr, err := repository.NewSessionRepository(db)
	if err != nil {
		return nil, err
	}
	return &AuthMiddleware{
		UserRepo:    ur,
		SessionRepo: sr,
	}, nil
}

// RequireAuth is a middleware used to authorize users with session tokens from
// the cookie, checking if the session in the database matching the token is
// valid and not expired. The session is rotated if it is halfway expired.
func (am *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get cookie from request
		sessionToken, err := c.Cookie(config.SessionCookieName)
		if err != nil {
			log.Debug().Err(err).Msg("No auth cookie found")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Split the session token
		parts := strings.Split(sessionToken, ".")
		if len(parts) != 2 {
			log.Debug().Msg("Invalid token format")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		sessionID, signature := parts[0], parts[1]
		parsedID, err := uuid.Parse(sessionID)
		if err != nil {
			log.Debug().Msg("Invalid token format")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Verify the HMAC signature
		if !models.ValidateSessionID(parsedID, signature) {
			log.Debug().Msg("Invalid token signature")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Get session from database
		session, err := am.SessionRepo.GetUnexpiredSessionByID(parsedID)
		if err != nil {
			log.Debug().Err(err).Msg("Session not found")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Check if session is expired
		if time.Now().UTC().After(session.ExpiresAt) {
			log.Debug().Msg("Session expired")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("userID", session.UserID.String())

		// Rotate session if halfway expired
		halfway := session.CreatedAt.Add(session.ExpiresAt.Sub(session.CreatedAt) / 2)
		if time.Now().UTC().After(halfway) {
			userService, err := services.NewUserService(am.UserRepo, am.SessionRepo)
			if err != nil {
				log.Debug().Err(err).Msg("Failed to rotate session")
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			// Rotate session
			newSessionToken, err := userService.RotateSession(parsedID)
			if err != nil {
				log.Debug().Err(err).Msg("Failed to rotate session")
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			} else {
				c.SetSameSite(http.SameSiteStrictMode)
				c.SetCookie(config.SessionCookieName, newSessionToken, int(config.SessionExpiration), "", "", true, true)
			}
		}

		c.Next()
	}
}
