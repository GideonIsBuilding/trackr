package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserStoreAdditions holds extra methods. In your project, add UpdatePassword
// directly to the UserStore struct in users.go instead.
// This file shows the method signature to add.

// Add this method to UserStore in users.go:
func (s *UserStore) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	const q = `UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1`
	_, err := s.db.Exec(ctx, q, userID, passwordHash)
	if err != nil {
		return fmt.Errorf("updating password: %w", err)
	}
	return nil
}

// ApplicationStoreAdditions — add this method to ApplicationStore in applications.go:

type AppDeleter struct {
	db *pgxpool.Pool
}
