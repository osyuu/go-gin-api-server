package server

import (
	"go-gin-api-server/config"
	"go-gin-api-server/internal/handler"
	"go-gin-api-server/internal/middleware"
	"go-gin-api-server/internal/repository"
	"go-gin-api-server/internal/service"
	"go-gin-api-server/pkg/logger"
	"go-gin-api-server/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// NewServer creates and configures a new Gin server
func NewServer(cfg *config.Config) *gin.Engine {
	// Set Gin mode based on environment
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Register custom validators
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		utils.RegisterCustomValidators(v)
	}

	// Create Gin router
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.GinZapMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"env":    cfg.Env,
		})
	})

	// Initialize repositories (using in-memory for now)
	userRepo := repository.NewUserRepository()
	authRepo := repository.NewAuthRepository()
	postRepo := repository.NewPostRepository()

	// Initialize JWT manager
	jwtMgr := utils.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessTokenExpiration)

	// Initialize services
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, authRepo, jwtMgr)
	postService := service.NewPostService(postRepo)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(authService)
	postHandler := handler.NewPostHandler(postService, logger.Log)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Register routes
	userHandler.RegisterRoutes(router)
	authHandler.RegisterRoutes(router)
	postHandler.RegisterRoutes(router)

	// Register protected routes
	userHandler.RegisterProtectedRoutes(router, authMiddleware)
	postHandler.RegisterProtectedRoutes(router, authMiddleware)

	return router
}
