package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID  `json:"id"`
	CreatedAt    time.Time  `json:"created_at"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	LastLogin    *time.Time `json:"last_login"`
}
