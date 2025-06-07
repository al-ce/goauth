package database

import (
	"os"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/al-ce/goauth/pkg/config"
)

// NewDB creates a new database connection using GORM and PostgreSQL.
func NewDB() (*gorm.DB, error) {
	var db *gorm.DB
	dsn := os.Getenv(config.DatabaseURL)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("Error connecting to database")
		return nil, err
	}
	return db, nil
}
