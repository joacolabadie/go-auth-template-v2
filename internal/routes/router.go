package routes

import (
	"github.com/joacolabadie/go-auth-template-v2/internal/auth"
	"github.com/joacolabadie/go-auth-template-v2/internal/handlers"
	"github.com/joacolabadie/go-auth-template-v2/internal/middleware"
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo, authService *auth.AuthService, authHandler *handlers.AuthHandler, userHandler *handlers.UserHandler) {
	// Public routes
	e.POST("/api/auth/register", authHandler.Register)
	e.POST("/api/auth/login", authHandler.Login)
	e.POST("/api/auth/refresh", authHandler.RefreshToken)

	// Protected routes
	e.GET("/api/user/profile", userHandler.Profile, middleware.AuthMiddleware(authService))
	e.POST("/api/auth/logout", authHandler.Logout, middleware.AuthMiddleware(authService))
}
