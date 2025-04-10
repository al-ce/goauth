package testutils

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"goauth/internal/database"
	"goauth/internal/models"
	"goauth/pkg/config"
)

func TestEnvSetup() {
	os.Setenv("PORT", "3001")
	os.Setenv("DB", "host=localhost user=goauth_test password=goauth_test dbname=goauth_test port=5432 sslmode=disable TimeZone=UTC")

	zerolog.SetGlobalLevel(zerolog.Disabled)

	gin.SetMode(gin.TestMode)
	gin.DefaultWriter = io.Discard
}

func TestDBSetup() *gorm.DB {
	gormLogger := logger.New(
		log.New(io.Discard, "", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(postgres.Open(os.Getenv("DB")), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	database.Migrate(db)
	return db
}

func MakeRequest(router *gin.Engine, method, path string, body any) (*httptest.ResponseRecorder, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		method,
		path,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, err
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	return rr, nil
}

// UserHandler_RegisterUser is a helper to register a user with a valid email and password
// for userHandler methods. For userRepository testing, ur.RegisterUser() suffices.
func UserHandler_RegisterUser(db *gorm.DB, user *models.User) error {
	_, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user1, err := models.NewUser(user.Email, user.Password)
	if err != nil {
		return err
	}
	return db.Create(user1).Error
}

// Helper function to create a test user
func CreateTestUser(db *gorm.DB, userName string) (*models.User, error) {
	user, err := models.NewUser(userName, config.TestingPassword)
	if err != nil {
		return nil, err
	}

	if err := db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}
