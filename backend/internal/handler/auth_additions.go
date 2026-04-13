package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/yourname/job-tracker/internal/middleware"
	"github.com/yourname/job-tracker/internal/service"
	"github.com/yourname/job-tracker/internal/store"
)

// --- Forgot / Reset password handlers ---
// Add these to AuthHandler (auth.go) or wire as a new PasswordResetHandler.

type PasswordResetHandler struct {
	svc *service.PasswordResetService
}

func NewPasswordResetHandler(svc *service.PasswordResetService) *PasswordResetHandler {
	return &PasswordResetHandler{svc: svc}
}

// POST /api/auth/forgot-password
// Body: { "email": "user@example.com" }
// Always returns 200 regardless of whether the email exists (prevents enumeration).
func (h *PasswordResetHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}

	// Fire and forget — we never reveal whether the email exists
	_ = h.svc.ForgotPassword(r.Context(), body.Email)

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "If that email is registered, you will receive a reset link shortly.",
	})
}

// POST /api/auth/reset-password
// Body: { "token": "...", "password": "newpassword" }
func (h *PasswordResetHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Token == "" || body.Password == "" {
		writeError(w, http.StatusBadRequest, "token and password are required")
		return
	}
	if len(body.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	err := h.svc.ResetPassword(r.Context(), body.Token, body.Password)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusBadRequest, "invalid or expired reset token")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Password updated successfully. Please sign in.",
	})
}

// --- Delete application handler ---
// Add this method to ApplicationHandler (applications.go)

// DELETE /api/applications/{id}
func (h *ApplicationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	appID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid application ID")
		return
	}

	if err := h.apps.Delete(r.Context(), appID, userID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "application not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete application")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
