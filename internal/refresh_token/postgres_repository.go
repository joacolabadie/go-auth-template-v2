package refreshtoken

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joacolabadie/go-auth-template-v2/internal/utils"
)

type PostgresRefreshTokenRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRefreshTokenRepository(db *pgxpool.Pool) *PostgresRefreshTokenRepository {
	return &PostgresRefreshTokenRepository{db: db}
}

func (r *PostgresRefreshTokenRepository) CreateRefreshToken(ctx context.Context, userID uuid.UUID, ttl time.Duration) (*RefreshToken, error) {
	rawToken := uuid.New().String()
	hashedToken := utils.HashToken(rawToken)

	token := &RefreshToken{
		UserID:    userID,
		Token:     rawToken,
		ExpiresAt: time.Now().Add(ttl),
		Revoked:   false,
	}

	q := `
		INSERT INTO refresh_tokens (user_id, token, expires_at, revoked)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(ctx, q, token.UserID, hashedToken, token.ExpiresAt, token.Revoked)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (r *PostgresRefreshTokenRepository) GetRefreshToken(ctx context.Context, rawToken string) (*RefreshToken, error) {
	var token RefreshToken

	hashedToken := utils.HashToken(rawToken)

	q := `
		SELECT id, created_at, user_id, token, expires_at, revoked
		FROM refresh_tokens
		WHERE token = $1
	`

	err := r.db.QueryRow(ctx, q, hashedToken).Scan(
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

	token.Token = rawToken

	return &token, nil
}

func (r *PostgresRefreshTokenRepository) RevokeRefreshToken(ctx context.Context, rawToken string) error {
	hashedToken := utils.HashToken(rawToken)

	q := `
		UPDATE refresh_tokens
		SET revoked = true
		WHERE token = $1
	`

	_, err := r.db.Exec(ctx, q, hashedToken)

	return err
}
