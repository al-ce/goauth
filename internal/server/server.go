package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type APIServer struct {
	DB     *gorm.DB
	Router *gin.Engine
}

func NewAPIServer(db *gorm.DB) *APIServer {
	return &APIServer{
		DB:     db,
		Router: gin.Default(),
	}
}

func (s *APIServer) Run() {
	r := s.Router
	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })

	r.POST("/register", RegisterUser)
}
