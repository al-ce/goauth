package testutils

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gofit/internal/database"
)

func TestEnvSetup() {
	os.Setenv("PORT", "3001")
	os.Setenv("DB", "host=localhost user=gofit_test password=gofit_test dbname=gofit_test port=5432 sslmode=disable TimeZone=UTC")

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
