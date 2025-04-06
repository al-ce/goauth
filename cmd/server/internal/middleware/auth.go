package middleware

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"gofit/internal/repository"
	"gofit/pkg/config"
)

type AuthMiddleware struct {
	UserRepo    *repository.UserRepository
	SessionRepo *repository.SessionRepository
}

func NewAuthMiddleware(db *gorm.DB) *AuthMiddleware {
	return &AuthMiddleware{
		UserRepo: repository.NewUserRepository(db),
		SessionRepo: repository.NewSessionRepository(db),
	}
}

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

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Debug().Msg("Invalid token claims")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			log.Debug().Msg("Token expired")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		session, err := am.SessionRepo.GetSessionByToken(tokenString)
		if err != nil {
			log.Debug().Err(err).Msg("Session not found")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("userID", session.UserID.String())

		c.Next()
	}
}
