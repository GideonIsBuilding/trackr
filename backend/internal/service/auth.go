package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yourname/job-tracker/internal/model"
	"github.com/yourname/job-tracker/internal/store"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email address already registered")
)

type AuthService struct {
	users     *store.UserStore
	jwtSecret string
	expiry    time.Duration
}

func NewAuthService(users *store.UserStore, jwtSecret string, expiry time.Duration) *AuthService {
	return &AuthService{users: users, jwtSecret: jwtSecret, expiry: expiry}
}

func (s *AuthService) Register(ctx context.Context, email, password, timezone string) (*model.User, string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("hashing password: %w", err)
	}

	if timezone == "" {
		timezone = "UTC"
	}

	user, err := s.users.Create(ctx, email, string(hash), timezone)
	if errors.Is(err, store.ErrConflict) {
		return nil, "", ErrEmailTaken
	}
	if err != nil {
		return nil, "", err
	}

	token, err := s.issueToken(user.ID)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*model.User, string, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, store.ErrNotFound) {
		return nil, "", ErrInvalidCredentials
	}
	if err != nil {
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	token, err := s.issueToken(user.ID)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

func (s *AuthService) issueToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.expiry)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}
	return signed, nil
}
