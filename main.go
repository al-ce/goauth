package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"gofit/initializers"
)

func init() {
	initializers.SetupLogger()
	initializers.LoadEnvVariables()
}

func main() {
	log.Info().Str("PORT", os.Getenv("PORT")).Msg("Starting server")

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ping": "pong"})
	})

	r.Run()
}
