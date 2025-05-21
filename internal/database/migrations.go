package database

import (
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"godiscauth/internal/models"
)

// Migrate automigrates the database according to the User and Session models
func Migrate(db *gorm.DB) error {
	// uuid extension
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error; err != nil {
		log.Fatal().Err(err).Msg("Error migrating database")
		return err
	}

	// make User migrations
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatal().Err(err).Msg("Error migrating User model")
		return err
	}

	// make Session migrations
	if err := db.AutoMigrate(&models.Session{}); err != nil {
		log.Fatal().Err(err).Msg("Error migrating Session model")
		return err
	}

	return nil
}
