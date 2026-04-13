package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/yourname/job-tracker/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type PasswordResetService struct {
	users      *store.UserStore
	resetStore *store.PasswordResetStore
	email      *EmailService
}

func NewPasswordResetService(
	users *store.UserStore,
	resetStore *store.PasswordResetStore,
	email *EmailService,
) *PasswordResetService {
	return &PasswordResetService{users: users, resetStore: resetStore, email: email}
}

// ForgotPassword looks up the user and sends a reset email.
// Always returns nil to prevent email enumeration attacks —
// the caller should respond the same way whether the email exists or not.
func (s *PasswordResetService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, store.ErrNotFound) {
		return nil // silent — don't reveal whether email is registered
	}
	if err != nil {
		return fmt.Errorf("looking up user: %w", err)
	}

	token, err := s.resetStore.CreateToken(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("creating reset token: %w", err)
	}

	if err := s.email.SendPasswordReset(user.Email, token); err != nil {
		return fmt.Errorf("sending reset email: %w", err)
	}

	return nil
}

// ResetPassword validates the token and updates the user's password.
func (s *PasswordResetService) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	userID, err := s.resetStore.ValidateAndConsume(ctx, rawToken)
	if err != nil {
		return err // "token expired", "token already used", or ErrNotFound
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	if err := s.users.UpdatePassword(ctx, userID, string(hash)); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	return nil
}
