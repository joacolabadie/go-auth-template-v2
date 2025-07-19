package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	refreshtoken "github.com/joacolabadie/go-auth-template-v2/internal/refresh_token"
	"github.com/joacolabadie/go-auth-template-v2/internal/user"
	"github.com/joacolabadie/go-auth-template-v2/internal/utils"
)

type Service struct {
	userRepo         user.Repository
	refreshTokenRepo refreshtoken.Repository
	jwtSecret        []byte
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
}

func NewService(userRepo user.Repository, refreshTokenRepo refreshtoken.Repository, jwtSecret string, accessTokenTTL time.Duration, refreshTokenTTL time.Duration) *Service {
	return &Service{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtSecret:        []byte(jwtSecret),
		accessTokenTTL:   accessTokenTTL,
		refreshTokenTTL:  refreshTokenTTL,
	}
}

func (s *Service) Register(ctx context.Context, email, password string, refreshTokenTTL time.Duration) (uuid.UUID, string, string, error) {
	_, err := s.userRepo.GetUserByEmail(ctx, email)
	if err == nil {
		return uuid.Nil, "", "", ErrEmailInUse
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, "", "", err
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return uuid.Nil, "", "", err
	}

	id, err := s.userRepo.CreateUser(ctx, email, hashedPassword)
	if err != nil {
		return uuid.Nil, "", "", err
	}

	if err := s.userRepo.UpdateLastLogin(ctx, id); err != nil {
		return uuid.Nil, "", "", err
	}

	accessToken, err := s.generateAccessToken(id)
	if err != nil {
		return uuid.Nil, "", "", err
	}

	refreshToken, err := s.refreshTokenRepo.CreateRefreshToken(ctx, id, refreshTokenTTL)
	if err != nil {
		return uuid.Nil, "", "", err
	}

	return id, accessToken, refreshToken.Token, nil
}

func (s *Service) Login(ctx context.Context, email, password string, refreshTokenTTL time.Duration) (string, string, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", ErrInvalidCredentials
	}

	if !utils.ComparePasswords(user.PasswordHash, password) {
		return "", "", ErrInvalidCredentials
	}

	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		return "", "", err
	}

	accessToken, err := s.generateAccessToken(user.ID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.refreshTokenRepo.CreateRefreshToken(ctx, user.ID, refreshTokenTTL)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken.Token, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	return s.refreshTokenRepo.RevokeRefreshToken(ctx, refreshToken)
}

func (s *Service) generateAccessToken(id uuid.UUID) (string, error) {
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

func (s *Service) ValidateToken(tokenString string) (jwt.MapClaims, error) {
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

func (s *Service) RefreshAccessToken(ctx context.Context, refreshTokenString string) (string, string, error) {
	token, err := s.refreshTokenRepo.GetRefreshToken(ctx, refreshTokenString)
	if err != nil {
		return "", "", ErrInvalidToken
	}

	if token.Revoked {
		return "", "", ErrInvalidToken
	}

	if time.Now().After(token.ExpiresAt) {
		return "", "", ErrExpiredToken
	}

	err = s.refreshTokenRepo.RevokeRefreshToken(ctx, refreshTokenString)
	if err != nil {
		return "", "", err
	}

	user, err := s.userRepo.GetUserByID(ctx, token.UserID)
	if err != nil {
		return "", "", err
	}

	accessToken, err := s.generateAccessToken(user.ID)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := s.refreshTokenRepo.CreateRefreshToken(ctx, user.ID, s.refreshTokenTTL)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken.Token, nil
}

func (s *Service) AccessTokenTTL() time.Duration {
	return s.accessTokenTTL
}

func (s *Service) RefreshTokenTTL() time.Duration {
	return s.refreshTokenTTL
}
