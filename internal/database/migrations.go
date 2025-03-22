package database

import "gofit/internal/models"

func Migrate() error {
	return DB.AutoMigrate(
		&models.User{},
	)
}
