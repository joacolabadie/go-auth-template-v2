package models

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID           uuid.UUID  `json:"id"`
	CreatedAt    time.Time  `json:"created_at"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	LastLogin    *time.Time `json:"last_login"`
}

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, email, passwordHash string) (*User, error) {
	user := &User{
		Email:        email,
		PasswordHash: passwordHash,
	}

	q := `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
	`

	_, err := r.db.Exec(ctx, q, user.Email, user.PasswordHash)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
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

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
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

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID, loginTime time.Time) error {
	q := `
		UPDATE users
		SET last_login = $1
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, q, loginTime, id)

	return err
}
