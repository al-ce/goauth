package database

import (
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"gofit/internal/models"
)

func Migrate(db *gorm.DB) error {
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error; err != nil {
		log.Fatal().Err(err).Msg("Error migrating database")
	}

	return db.AutoMigrate(
		&models.User{},
		&models.Session{},
	)
}
