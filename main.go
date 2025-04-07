package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"

	"gofit/internal/database"
	"gofit/internal/server"
	"gofit/pkg/logger"
)

func main() {
	logger.SetupLogger()

	if err := godotenv.Load(); err != nil {
		log.Fatal().Err(err).Msg("Error loading env file")
	}

	db := database.NewDB()
	database.Migrate(db)

	log.Info().Str("PORT", os.Getenv("PORT")).Msg("Starting server")

	apiServer := server.NewAPIServer(db)
	apiServer.Run()
}
