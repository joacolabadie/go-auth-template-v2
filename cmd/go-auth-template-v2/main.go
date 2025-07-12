package main

import (
	"log"

	"github.com/joacolabadie/go-auth-template-v2/internal/auth"
	"github.com/joacolabadie/go-auth-template-v2/internal/config"
	"github.com/joacolabadie/go-auth-template-v2/internal/database"
	"github.com/joacolabadie/go-auth-template-v2/internal/handlers"
	mw "github.com/joacolabadie/go-auth-template-v2/internal/middleware"
	"github.com/joacolabadie/go-auth-template-v2/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dbPool, err := database.ConnectDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	e := echo.New()

	// CORS middleware configuration
	e.Use(middleware.CORS())

	// Rate limiting middleware configuration
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(20))))

	// Create repositories
	userRepo := models.NewUserRepository(dbPool)
	refreshTokenRepo := models.NewRefreshTokenRepository(dbPool)

	// Create services
	authService := auth.NewAuthService(userRepo, refreshTokenRepo, cfg.JWT.Secret, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)

	// Create handlers
	authHandler := handlers.NewAuthHandler(authService, cfg.Environment)
	userHandler := handlers.NewUserHandler(userRepo)

	// Public routes
	e.POST("/api/auth/register", authHandler.Register)
	e.POST("/api/auth/login", authHandler.Login)
	e.POST("/api/auth/refresh", authHandler.RefreshToken)

	// Protected routes
	e.GET("/api/user/profile", userHandler.Profile, mw.AuthMiddleware(authService))

	log.Printf("Starting server on port %s...", cfg.Server.Port)

	if err := e.Start(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
