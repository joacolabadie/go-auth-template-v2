package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joacolabadie/go-auth-template-v2/internal/auth"
	"github.com/joacolabadie/go-auth-template-v2/internal/utils"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authService *auth.AuthService
	environment string
}

func NewAuthHandler(authService *auth.AuthService, environment string) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		environment: environment,
	}
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,gte=6"`
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Failed to parse request", "details": err.Error()})
	}

	if err := utils.Validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)

		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "Failed to validate request",
			"details": validationErrors.Error(),
		})
	}

	ctx := c.Request().Context()

	userID, err := h.authService.Register(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrEmailInUse) {
			return c.JSON(http.StatusConflict, echo.Map{
				"error": "A user with this email already exists",
			})
		} else {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error":   "Database error",
				"details": err.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "User registered successfully",
		"id":      userID,
	})
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,gte=6"`
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Failed to parse request", "details": err.Error()})
	}

	if err := utils.Validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)

		return c.JSON(http.StatusBadRequest, echo.Map{
			"error":   "Failed to validate request",
			"details": validationErrors.Error(),
		})
	}

	ctx := c.Request().Context()

	refreshTokenTTL := h.authService.RefreshTokenTTL()

	accessToken, refreshToken, err := h.authService.Login(ctx, req.Email, req.Password, refreshTokenTTL)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"error": "Invalid credentials",
			})
		} else {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error":   "Database error",
				"details": err.Error(),
			})
		}
	}

	isProd := h.environment == "production"

	ttl := h.authService.AccessTokenTTL()

	accessTokenCookie := &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		Expires:  time.Now().Add(ttl),
		MaxAge:   int(ttl.Seconds()),
		Secure:   isProd,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(accessTokenCookie)

	refreshTokenCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		Expires:  time.Now().Add(refreshTokenTTL),
		MaxAge:   int(refreshTokenTTL.Seconds()),
		Secure:   isProd,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(refreshTokenCookie)

	return c.JSON(http.StatusOK, echo.Map{
		"message": "User logged in successfully",
	})
}

func (h *AuthHandler) RefreshToken(c echo.Context) error {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Missing refresh_token cookie"})
	}

	refreshTokenString := cookie.Value

	ctx := c.Request().Context()

	token, err := h.authService.RefreshAccessToken(ctx, refreshTokenString)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidToken) || errors.Is(err, auth.ErrExpiredToken) {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"error": "Invalid or expired refresh token",
			})
		} else {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error": "Internal server error",
			})
		}
	}

	isProd := h.environment == "production"

	ttl := h.authService.AccessTokenTTL()

	accessTokenCookie := &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(ttl),
		MaxAge:   int(ttl.Seconds()),
		Secure:   isProd,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(accessTokenCookie)

	return c.JSON(http.StatusOK, echo.Map{
		"message": "Access token refreshed successfully",
	})
}

func (h *AuthHandler) Logout(c echo.Context) error {
	refreshTokenCookie, err := c.Cookie("refresh_token")
	if err != nil {
		auth.ClearCookies(c, h.environment)

		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "Logged out successfully"})
	}

	refreshToken := refreshTokenCookie.Value

	ctx := c.Request().Context()

	_ = h.authService.Logout(ctx, refreshToken)

	auth.ClearCookies(c, h.environment)

	return c.JSON(http.StatusUnauthorized, echo.Map{"message": "Logged out successfully"})

}
