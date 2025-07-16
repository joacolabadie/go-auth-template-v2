package user

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, email, passwordHash string) (uuid.UUID, error) {
	user := &User{
		Email:        email,
		PasswordHash: passwordHash,
	}

	var id uuid.UUID

	q := `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id
	`

	err := r.db.QueryRow(ctx, q, user.Email, user.PasswordHash).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (r *PostgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	q := `
		SELECT id, created_at, email, password_hash, last_login
		FROM users
		WHERE email = $1
	`

	var user User
	var lastLogin pgtype.Timestamp

	err := r.db.QueryRow(ctx, q, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Email,
		&user.PasswordHash,
		&lastLogin,
	)
	if err != nil {
		return nil, err
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return &user, nil
}

func (r *PostgresUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	q := `
		SELECT id, created_at, email, password_hash, last_login
		FROM users
		WHERE id = $1
	`

	var user User
	var lastLogin pgtype.Timestamp

	err := r.db.QueryRow(ctx, q, id).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Email,
		&user.PasswordHash,
		&lastLogin,
	)
	if err != nil {
		return nil, err
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return &user, nil
}

func (r *PostgresUserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	q := `
		UPDATE users
		SET last_login = $1
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, q, time.Now(), id)

	return err
}
