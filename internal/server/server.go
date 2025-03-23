package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gofit/internal/handlers"
	"gofit/internal/middleware"
	"gofit/internal/repository"
	"gofit/internal/services"
)

type APIServer struct {
	Router      *gin.Engine
	Handlers    *Handlers
	Middlewares *Middlewares
}

func NewAPIServer(db *gorm.DB) *APIServer {
	repos := initRepositories(db)
	services := initServices(repos)
	handlers := initHandlers(services)
	middlewares := initMiddlewares(db)

	server := &APIServer{
		Router:      gin.Default(),
		Handlers:    handlers,
		Middlewares: middlewares,
	}

	return server
}

func (s *APIServer) SetupRoutes() {
	r := s.Router
	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })
	r.POST("/register", s.Handlers.User.RegisterUser)
	r.POST("/login", s.Handlers.User.Login)

    protected := r.Group("")
    protected.Use(s.Middlewares.Auth.RequireAuth())
    {
        protected.GET("/profile", s.Handlers.User.GetProfile)
    }

	// TODO: admin group

	// admin := r.Group("/admin")
	// admin.Use(authMiddleware.RequireAuth(), authMiddleware.RequireAdmin())
	// {
	// 	admin.GET("/users", someHandler)
	// }

}

func (s *APIServer) Run() {
	s.SetupRoutes()
	s.Router.Run()
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

func initMiddlewares(db *gorm.DB) *Middlewares {
	return &Middlewares{
		Auth: middleware.NewAuthMiddleware(db),
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

type Middlewares struct {
	Auth *middleware.AuthMiddleware
}
