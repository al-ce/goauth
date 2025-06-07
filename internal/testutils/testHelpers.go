package testutils

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/al-ce/goauth/internal/database"
	"github.com/al-ce/goauth/pkg/config"
)

const TestingPassword = "correcthorsebatterystaple"

// TestEnvSetup sets environment variables for the tests. The tests assume the
// relevant test database has been created and that a postgres service is
// running on port 5433 (e.g. a Docker container with port mapping 5433:5432)
// See `scripts/init_testing.sql` to create the testing database.
func TestEnvSetup() {
	os.Setenv(config.SessionKey, uuid.New().String())
	os.Setenv(config.AuthServerPort, "3001")

	// Use port 5433
	os.Setenv(config.DatabaseURL, "host=localhost user=goauth_test password=goauth_test dbname=goauth_test port=5433 sslmode=disable TimeZone=UTC")

	zerolog.SetGlobalLevel(zerolog.Disabled)

	gin.SetMode(gin.TestMode)
	gin.DefaultWriter = io.Discard
}

// TestDBSetup sets up a test database connection.
func TestDBSetup() *gorm.DB {
	// Silence GORM logs for testing
	gormLogger := logger.New(
		log.New(io.Discard, "", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// Connect to test DB
	db, err := gorm.Open(postgres.Open(os.Getenv(config.DatabaseURL)), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	err = database.Migrate(db)
	if err != nil {
		log.Fatalf("Error migrating database: %v", err)
	}
	return db
}
