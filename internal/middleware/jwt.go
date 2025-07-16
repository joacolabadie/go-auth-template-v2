package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/joacolabadie/go-auth-template-v2/internal/auth"
	"github.com/labstack/echo/v4"
)

func JWTMiddleware(authService *auth.Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("access_token")
			if err != nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Missing access_token cookie"})
			}

			tokenString := cookie.Value

			claims, err := authService.ValidateToken(tokenString)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid or expired token"})
			}

			sub, ok := claims["sub"].(string)
			if !ok || sub == "" {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token claims"})
			}

			userID, err := uuid.Parse(sub)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid user ID in token"})
			}

			c.Set("userID", userID)

			return next(c)
		}
	}
}
