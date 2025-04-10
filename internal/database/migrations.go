package database

import (
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"goauth/internal/models"
)

func Migrate(db *gorm.DB) {
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error; err != nil {
		log.Fatal().Err(err).Msg("Error migrating database")
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatal().Err(err).Msg("Error migrating User model")
	}

	if err := db.AutoMigrate(&models.Session{}); err != nil {
		log.Fatal().Err(err).Msg("Error migrating Session model")
	}
}
