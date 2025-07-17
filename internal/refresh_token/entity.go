package refreshtoken

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
	Revoked   bool
}
