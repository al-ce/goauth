package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	passwordvalidator "github.com/wagslane/go-password-validator"
	"gorm.io/gorm"

	_ "godiscauth/docs"
	"godiscauth/internal/database"
	"godiscauth/internal/jobs"
	"godiscauth/internal/server"
	"godiscauth/pkg/config"
	"godiscauth/pkg/logger"
)

// main is the entry point for the auth service. It sets up the logger,
// connects to the database, starts the API server, and start any background jobs
func main() {
	checkSessionKey()

	logger.SetupLogger()

	db := connectDB()

	startAPIServer(db)

	quit := makeQuitListener()

	wg, cancel := startJobs(db)

	// Block until quit signal
	<-quit
	log.Info().Msg("Server is shutting down...")

	// Close context done channel, signaling jobs in waitgroup to finish
	cancel()
	waitForJobs(wg)

	log.Info().Msg("Server exited")
}

// Ensure session key is complex enough for encryption
func checkSessionKey() {
	if err := passwordvalidator.Validate(os.Getenv(config.SessionKey), config.MinEntropyBits); err != nil {
		log.Fatal().Err(err).Msg("Session secret is not complex enough")
	}
}

// Connect and migrate DB
func connectDB() *gorm.DB {
	db, err := database.NewDB()
	if err != nil {
		log.Fatal().Err(err).Msg("Error connecting to database")
	}
	err = database.Migrate(db)
	if err != nil {
		log.Fatal().Err(err).Msg("Error migrating database")
	}
	log.Info().Msg(fmt.Sprintf("Connected to postgres database"))
	return db
}

// Start API Server
func startAPIServer(db *gorm.DB) {
	log.Info().
		Str(config.AuthServerPort, os.Getenv(config.AuthServerPort)).
		Msg("Starting server")
	apiServer, err := server.NewAPIServer(db)
	if err != nil {
		log.Fatal().Err(err).Msg("Error initializing server")
	}
	go apiServer.Run()
}


// Start background jobs as goroutines
func startJobs(db *gorm.DB) (*sync.WaitGroup, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	go jobs.StartJobs(ctx, &wg, db)
	return &wg, cancel
}

// Make a channel to listen for a quit signal
func makeQuitListener() chan os.Signal {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	return quit
}

// Wait with timeout for goroutines to finish, or forcequit
func waitForJobs(wg *sync.WaitGroup) {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info().Msg("graceful shutdown")
	case <-time.After(10 * time.Second):
		log.Info().Msg("forcequitting")
	}
}
