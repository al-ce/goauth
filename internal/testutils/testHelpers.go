package testutils

import (
	"os"

	"gorm.io/gorm"

	"gofit/internal/database"
)

func TestEnvSetup() {
	os.Setenv("PORT", "3001")
	os.Setenv("DB", "host=localhost user=gofit_test password=gofit_test dbname=gofit_test port=5432 sslmode=disable TimeZone=UTC")
}

func TestDBSetup() *gorm.DB {
	db := database.NewDB()

	database.Migrate(db)
	return db
}
