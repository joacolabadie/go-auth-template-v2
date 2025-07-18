package user

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{
		repo: repo,
	}
}

type ProfileResponse struct {
	ID        string     `json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	Email     string     `json:"email"`
	LastLogin *time.Time `json:"last_login,omitempty"`
}

func (h *Handler) Profile(c echo.Context) error {
	userID, ok := c.Get("userID").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	user, err := h.repo.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "Failed to retrieve user profile",
		})
	}

	response := ProfileResponse{
		ID:        user.ID.String(),
		CreatedAt: user.CreatedAt,
		Email:     user.Email,
		LastLogin: user.LastLogin,
	}

	return c.JSON(http.StatusOK, response)
}
