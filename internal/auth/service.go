package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/joacolabadie/go-auth-template-v2/internal/models"
	"github.com/joacolabadie/go-auth-template-v2/internal/utils"
)

var (
	ErrEmailInUse         = errors.New("email already in use")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrExpiredToken       = errors.New("token has expired")
)

type AuthService struct {
	userRepo       *models.UserRepository
	jwtSecret      []byte
	accessTokenTTL time.Duration
}

func NewAuthService(userRepo *models.UserRepository, jwtSecret string, accessTokenTTL time.Duration) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		jwtSecret:      []byte(jwtSecret),
		accessTokenTTL: accessTokenTTL,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*models.User, error) {
	_, err := s.userRepo.GetUserByEmail(ctx, email)
	if err == nil {
		return nil, ErrEmailInUse
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.CreateUser(ctx, email, hashedPassword)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if !utils.ComparePasswords(user.PasswordHash, password) {
		return "", ErrInvalidCredentials
	}

	token, err := s.generateAccessToken(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) generateAccessToken(id uuid.UUID) (string, error) {
	expirationTime := time.Now().Add(s.accessTokenTTL)

	claims := jwt.MapClaims{
		"sub": id.String(),
		"exp": expirationTime.Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *AuthService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
