package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"gofit/internal/config"
	"gofit/internal/database"
	"gofit/pkg/logger"
)

func init() {
	logger.SetupLogger()
	config.LoadEnvVariables()
	database.ConnectToDB()
	database.Migrate()
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	return router
}

func main() {
	log.Info().Str("PORT", os.Getenv("PORT")).Msg("Starting server")
	r := setupRouter()

	r.Run()
}
