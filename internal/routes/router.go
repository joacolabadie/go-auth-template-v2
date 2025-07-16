package routes

import (
	"github.com/joacolabadie/go-auth-template-v2/internal/auth"
	"github.com/joacolabadie/go-auth-template-v2/internal/middleware"
	"github.com/joacolabadie/go-auth-template-v2/internal/user"
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo, authService *auth.Service, authHandler *auth.Handler, userHandler *user.Handler) {
	// Public routes
	e.POST("/api/auth/register", authHandler.Register)
	e.POST("/api/auth/login", authHandler.Login)
	e.POST("/api/auth/refresh", authHandler.RefreshToken)

	// Protected routes
	e.GET("/api/user/profile", userHandler.Profile, middleware.JWTMiddleware(authService))
	e.POST("/api/auth/logout", authHandler.Logout, middleware.JWTMiddleware(authService))
}
