package main

import (
	"os"

	"github.com/rs/zerolog/log"

	"gofit/initializers"
)

func init() {
	initializers.SetupLogger()
	initializers.LoadEnvVariables()
}

func main() {
	log.Info().Str("PORT", os.Getenv("PORT")).Msg("Starting server")
}
