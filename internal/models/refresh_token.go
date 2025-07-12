package models

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshToken struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
	Revoked   bool
}

type RefreshTokenRepository struct {
	db *pgxpool.Pool
}

func NewRefreshTokenRepository(db *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) CreateRefreshToken(ctx context.Context, userID uuid.UUID, ttl time.Duration) (*RefreshToken, error) {
	token := &RefreshToken{
		UserID:    userID,
		Token:     uuid.New().String(),
		ExpiresAt: time.Now().Add(ttl),
		Revoked:   false,
	}

	q := `
		INSERT INTO refresh_tokens (user_id, token, expires_at, revoked)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(ctx, q, token.UserID, token.Token, token.ExpiresAt, token.Revoked)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (r *RefreshTokenRepository) GetRefreshToken(ctx context.Context, tokenString string) (*RefreshToken, error) {
	var token RefreshToken

	q := `
		SELECT id, created_at, user_id, token, expires_at, revoked
		FROM refresh_tokens
		WHERE token = $1
	`

	err := r.db.QueryRow(ctx, q, tokenString).Scan(
		&token.ID,
		&token.CreatedAt,
		&token.UserID,
		&token.Token,
		&token.ExpiresAt,
		&token.Revoked,
	)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *RefreshTokenRepository) RevokeRefreshToken(ctx context.Context, tokenString string) error {
	q := `
		UPDATE refresh_tokens
		SET revoked = true
		WHERE token = $1
	`

	_, err := r.db.Exec(ctx, q, tokenString)

	return err
}
