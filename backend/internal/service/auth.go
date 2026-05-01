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

// ParseToken validates a JWT and returns the subject (user ID string).
func (s *AuthService) ParseToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}
	sub, err := token.Claims.GetSubject()
	if err != nil {
		return "", fmt.Errorf("invalid claims")
	}
	return sub, nil
}

// GetUser looks up a user by their string UUID.
func (s *AuthService) GetUser(ctx context.Context, userID string) (*model.User, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}
	return s.users.GetByID(ctx, id)
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
