package middleware

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"goauth/internal/repository"
	"goauth/internal/services"
	"goauth/pkg/config"
)

type AuthMiddleware struct {
	UserRepo    *repository.UserRepository
	SessionRepo *repository.SessionRepository
}

func NewAuthMiddleware(db *gorm.DB) *AuthMiddleware {
	return &AuthMiddleware{
		UserRepo:    repository.NewUserRepository(db),
		SessionRepo: repository.NewSessionRepository(db),
	}
}

// RequireAuth is a middleware used to authorize users with JWT tokens from
// the cookie, checking if the session in the database matching the token is
// valid and not expired. The session is rotated if it is halfway expired.
func (am *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get cookie from request
		tokenString, err := c.Cookie(config.JwtCookieName)
		if err != nil {
			log.Debug().Err(err).Msg("No auth cookie found")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Decode and validate
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			return []byte(os.Getenv(config.JwtCookieName)), nil
		})

		if err != nil || !token.Valid {
			log.Debug().Err(err).Msg("Invalid token")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Get claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Debug().Msg("Invalid token claims")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Check if token is expired
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			log.Debug().Msg("Token expired")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Get session from database
		session, err := am.SessionRepo.GetSessionByToken(tokenString)
		if err != nil {
			log.Debug().Err(err).Msg("Session not found")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("userID", session.UserID.String())

		// Rotate session if halfway expired
		halfway := session.CreatedAt.Add(session.ExpiresAt.Sub(session.CreatedAt) / 2)
		if time.Now().After(halfway) {
			userService := services.NewUserService(am.UserRepo, am.SessionRepo)
			newToken, err := userService.RotateSession(tokenString)
			if err != nil {
				log.Debug().Err(err).Msg("Failed to rotate session")
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			} else {
				c.SetSameSite(http.SameSiteStrictMode)
				c.SetCookie(config.JwtCookieName, newToken, int(config.TokenExpiration), "", "", true, true)
			}
		}

		c.Next()
	}
}
