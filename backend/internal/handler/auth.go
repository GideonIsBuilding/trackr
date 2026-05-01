package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
	"unicode"

	"github.com/yourname/job-tracker/internal/service"
)

const sessionCookieName = "trackr_session"

// validatePassword enforces the minimum advisable password policy:
// at least 12 characters containing uppercase, lowercase, digit, and special character.
func validatePassword(p string) error {
	if len(p) < 12 {
		return errors.New("password must be at least 12 characters")
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, c := range p {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSpecial = true
		}
	}
	var missing []string
	if !hasUpper {
		missing = append(missing, "an uppercase letter")
	}
	if !hasLower {
		missing = append(missing, "a lowercase letter")
	}
	if !hasDigit {
		missing = append(missing, "a number")
	}
	if !hasSpecial {
		missing = append(missing, "a special character")
	}
	if len(missing) > 0 {
		return fmt.Errorf("password must contain %s", joinRequirements(missing))
	}
	return nil
}

func joinRequirements(items []string) string {
	switch len(items) {
	case 1:
		return items[0]
	case 2:
		return items[0] + " and " + items[1]
	default:
		return items[0] + ", " + joinRequirements(items[1:])
	}
}

type AuthHandler struct {
	auth   *service.AuthService
	expiry time.Duration
}

func NewAuthHandler(auth *service.AuthService, expiry time.Duration) *AuthHandler {
	return &AuthHandler{auth: auth, expiry: expiry}
}

// setSessionCookie writes an httpOnly cookie holding the JWT.
// SameSite=None + Secure lets the browser extension send it cross-site.
// Chrome treats localhost as a secure context, so Secure works over HTTP in dev.
func (h *AuthHandler) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   int(h.expiry.Seconds()),
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Timezone string `json:"timezone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Email == "" || body.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}
	if err := validatePassword(body.Password); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, token, err := h.auth.Register(r.Context(), body.Email, body.Password, body.Timezone)
	if errors.Is(err, service.ErrEmailTaken) {
		writeError(w, http.StatusConflict, "email address already registered")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "registration failed")
		return
	}

	h.setSessionCookie(w, token)
	writeJSON(w, http.StatusCreated, map[string]any{
		"user":  user,
		"token": token,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, token, err := h.auth.Login(r.Context(), body.Email, body.Password)
	if errors.Is(err, service.ErrInvalidCredentials) {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "login failed")
		return
	}

	h.setSessionCookie(w, token)
	writeJSON(w, http.StatusOK, map[string]any{
		"user":  user,
		"token": token,
	})
}

// Logout clears the session cookie so the browser discards it immediately.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   -1,
	})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// ExtensionToken is called by the browser extension with credentials:'include'.
// It reads the httpOnly session cookie set at login and returns the token + user,
// so the extension never needs its own login form.
func (h *AuthHandler) ExtensionToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	userID, err := h.auth.ParseToken(cookie.Value)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}

	user, err := h.auth.GetUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"token": cookie.Value,
		"user":  user,
	})
}
