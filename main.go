package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	passwordvalidator "github.com/wagslane/go-password-validator"

	"goauth/internal/database"
	"goauth/internal/server"
	"goauth/pkg/config"
	"goauth/pkg/logger"
)

func main() {
	logger.SetupLogger()

	if err := validateJWTSecret(); err != nil {
		log.Fatal().Err(err).Msg("JWT secret not strog enough")
	}

	db := database.NewDB()
	database.Migrate(db)

	log.Info().Str("PORT", os.Getenv("PORT")).Msg("Starting server")

	apiServer := server.NewAPIServer(db)
	apiServer.Run()
}

// validateJWTSecret ensures user auth won't use weak cookie name
func validateJWTSecret() error {
	secret := os.Getenv(config.JwtCookieName)
	if secret == "" {
		return errors.New("JWT secret is empty")
	}

	if err := passwordvalidator.Validate(secret, config.MinEntropyBits+16); err != nil {
		return fmt.Errorf("JWT secret not strong enough: %w", err)
	}

	return nil
}
