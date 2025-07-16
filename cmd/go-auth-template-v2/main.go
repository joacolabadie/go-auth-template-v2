package main

import (
	"log"

	"github.com/joacolabadie/go-auth-template-v2/internal/auth"
	"github.com/joacolabadie/go-auth-template-v2/internal/config"
	"github.com/joacolabadie/go-auth-template-v2/internal/database"
	refreshtoken "github.com/joacolabadie/go-auth-template-v2/internal/refresh_token"
	"github.com/joacolabadie/go-auth-template-v2/internal/routes"
	"github.com/joacolabadie/go-auth-template-v2/internal/user"
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
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"*"},
		AllowHeaders: []string{"*"},
	}))

	// Rate limiting middleware configuration
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(20))))

	// Create repositories
	userRepo := user.NewPostgresUserRepository(dbPool)
	refreshTokenRepo := refreshtoken.NewPostgresRefreshTokenRepository(dbPool)

	// Create services
	authService := auth.NewService(userRepo, refreshTokenRepo, cfg.JWT.Secret, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)

	// Create handlers
	authHandler := auth.NewHandler(authService, cfg.Environment)
	userHandler := user.NewHandler(userRepo)

	// Register routes
	routes.RegisterRoutes(e, authService, authHandler, userHandler)

	log.Printf("Starting server on port %s...", cfg.Server.Port)

	if err := e.Start(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
