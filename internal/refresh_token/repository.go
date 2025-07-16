package refreshtoken

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	CreateRefreshToken(ctx context.Context, userID uuid.UUID, ttl time.Duration) (*RefreshToken, error)
	GetRefreshToken(ctx context.Context, tokenString string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenString string) error
}
