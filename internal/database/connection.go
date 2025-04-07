package database

import (
	"os"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBConfig struct {
	Host       string
	Port       string
	User       string
	Password   string
	Name       string
	ServerPort string
}

func NewDB() *gorm.DB {
	var db *gorm.DB
	dsn := os.Getenv("DB")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("Error connecting to database")
	}
	return db
}
