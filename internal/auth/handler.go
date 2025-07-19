package auth

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/joacolabadie/go-auth-template-v2/internal/utils"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	service     *Service
	environment string
}

func NewHandler(service *Service, environment string) *Handler {
	return &Handler{
		service:     service,
		environment: environment,
	}
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,gte=6"`
}

func (h *Handler) Register(c echo.Context) error {
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

	refreshTokenTTL := h.service.RefreshTokenTTL()

	userID, accessToken, refreshToken, err := h.service.Register(ctx, req.Email, req.Password, refreshTokenTTL)
	if err != nil {
		if errors.Is(err, ErrEmailInUse) {
			return c.JSON(http.StatusConflict, echo.Map{
				"error": "A user with this email already exists",
			})
		} else {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error": "Internal server error",
			})
		}
	}

	accessTokenTTL := h.service.AccessTokenTTL()

	SetAuthCookies(c, h.environment, accessToken, refreshToken, accessTokenTTL, refreshTokenTTL)

	return c.JSON(http.StatusOK, echo.Map{
		"message": "User registered successfully",
		"user_id": userID,
	})
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,gte=6"`
}

func (h *Handler) Login(c echo.Context) error {
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

	refreshTokenTTL := h.service.RefreshTokenTTL()

	accessToken, refreshToken, err := h.service.Login(ctx, req.Email, req.Password, refreshTokenTTL)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"error": "Invalid credentials",
			})
		} else {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error": "Internal server error",
			})
		}
	}

	accessTokenTTL := h.service.AccessTokenTTL()

	SetAuthCookies(c, h.environment, accessToken, refreshToken, accessTokenTTL, refreshTokenTTL)

	return c.JSON(http.StatusOK, echo.Map{
		"message": "User logged in successfully",
	})
}

func (h *Handler) RefreshToken(c echo.Context) error {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Missing refresh_token cookie"})
	}

	refreshTokenString := cookie.Value

	ctx := c.Request().Context()

	accessToken, newRefreshToken, err := h.service.RefreshAccessToken(ctx, refreshTokenString)
	if err != nil {
		if errors.Is(err, ErrInvalidToken) || errors.Is(err, ErrExpiredToken) {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"error": "Invalid or expired refresh token",
			})
		} else {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error": "Internal server error",
			})
		}
	}

	accessTokenTTL := h.service.AccessTokenTTL()
	refreshTokenTTL := h.service.RefreshTokenTTL()

	SetAuthCookies(c, h.environment, accessToken, newRefreshToken, accessTokenTTL, refreshTokenTTL)

	return c.JSON(http.StatusOK, echo.Map{
		"message": "Access token refreshed successfully",
	})
}

func (h *Handler) Logout(c echo.Context) error {
	ctx := c.Request().Context()

	if refreshTokenCookie, err := c.Cookie("refresh_token"); err == nil {
		_ = h.service.Logout(ctx, refreshTokenCookie.Value)
	}

	ClearAuthCookies(c, h.environment)

	return c.JSON(http.StatusOK, echo.Map{"message": "User logged out successfully"})
}
