package initializers

import (
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func LoadEnvVariables() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading env file")
	}
}
