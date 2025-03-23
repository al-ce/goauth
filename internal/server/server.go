package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gofit/internal/handlers"
	"gofit/internal/repository"
	"gofit/internal/services"
)

type APIServer struct {
	Router   *gin.Engine
	Handlers *Handlers
}

func NewAPIServer(db *gorm.DB) *APIServer {
	repos := initRepositories(db)
	services := initServices(repos)
	handlers := initHandlers(services)

	server := &APIServer{
		Router:   gin.Default(),
		Handlers: handlers,
	}

	return server
}

func (s *APIServer) Run() {
	r := s.Router
	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })

	r.POST("/register", s.Handlers.User.RegisterUser)
}

func initRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User: repository.NewUserRepository(db),
	}
}

func initServices(repos *Repositories) *Services {
	return &Services{
		User: services.NewUserService(repos.User),
	}
}

func initHandlers(services *Services) *Handlers {
	return &Handlers{
		User: handlers.NewUserHandler(services.User),
	}
}

type Repositories struct {
	User *repository.UserRepository
}

type Services struct {
	User *services.UserService
}

type Handlers struct {
	User *handlers.UserHandler
}
