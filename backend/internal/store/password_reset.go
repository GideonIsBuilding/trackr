package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PasswordResetStore struct {
	db *pgxpool.Pool
}

func NewPasswordResetStore(db *pgxpool.Pool) *PasswordResetStore {
	return &PasswordResetStore{db: db}
}

// CreateToken generates a secure random token, stores its hash, and returns the raw token.
// The raw token is sent to the user; only the hash lives in the DB.
func (s *PasswordResetStore) CreateToken(ctx context.Context, userID uuid.UUID) (string, error) {
	// Generate 32 random bytes → 64 char hex string
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generating reset token: %w", err)
	}
	rawToken := hex.EncodeToString(raw)
	hash := hashToken(rawToken)

	// Invalidate any existing unused tokens for this user
	const invalidate = `
		UPDATE password_reset_tokens
		SET used_at = NOW()
		WHERE user_id = $1 AND used_at IS NULL`
	if _, err := s.db.Exec(ctx, invalidate, userID); err != nil {
		return "", fmt.Errorf("invalidating old tokens: %w", err)
	}

	const insert = `
		INSERT INTO password_reset_tokens (user_id, token_hash)
		VALUES ($1, $2)`
	if _, err := s.db.Exec(ctx, insert, userID, hash); err != nil {
		return "", fmt.Errorf("storing reset token: %w", err)
	}

	return rawToken, nil
}

// ValidateAndConsume checks the token is valid, unused, and not expired.
// If valid, it marks it used and returns the associated user ID.
func (s *PasswordResetStore) ValidateAndConsume(ctx context.Context, rawToken string) (uuid.UUID, error) {
	hash := hashToken(rawToken)

	const q = `
		SELECT id, user_id, expires_at, used_at
		FROM password_reset_tokens
		WHERE token_hash = $1`

	var (
		id        uuid.UUID
		userID    uuid.UUID
		expiresAt time.Time
		usedAt    *time.Time
	)

	err := s.db.QueryRow(ctx, q, hash).Scan(&id, &userID, &expiresAt, &usedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("looking up reset token: %w", err)
	}
	if usedAt != nil {
		return uuid.Nil, errors.New("token already used")
	}
	if time.Now().After(expiresAt) {
		return uuid.Nil, errors.New("token expired")
	}

	// Mark as used
	const consume = `UPDATE password_reset_tokens SET used_at = NOW() WHERE id = $1`
	if _, err := s.db.Exec(ctx, consume, id); err != nil {
		return uuid.Nil, fmt.Errorf("consuming token: %w", err)
	}

	return userID, nil
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
