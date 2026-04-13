package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/yourname/job-tracker/internal/middleware"
	"github.com/yourname/job-tracker/internal/store"
)

type ReminderHandler struct {
	apps      *store.ApplicationStore
	reminders *store.ReminderStore
}

func NewReminderHandler(apps *store.ApplicationStore, reminders *store.ReminderStore) *ReminderHandler {
	return &ReminderHandler{apps: apps, reminders: reminders}
}

// Configure sets the trigger_after_days for an application's reminder.
func (h *ReminderHandler) Configure(w http.ResponseWriter, r *http.Request) {
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

	// Verify the application belongs to this user
	if _, err := h.apps.GetByID(r.Context(), appID, userID); err != nil {
		writeError(w, http.StatusNotFound, "application not found")
		return
	}

	var body struct {
		TriggerAfterDays int `json:"trigger_after_days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.TriggerAfterDays < 1 {
		writeError(w, http.StatusBadRequest, "trigger_after_days must be at least 1")
		return
	}

	reminder, err := h.reminders.Upsert(r.Context(), appID, body.TriggerAfterDays)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to configure reminder")
		return
	}

	writeJSON(w, http.StatusOK, reminder)
}

// Snooze postpones the reminder for N days.
func (h *ReminderHandler) Snooze(w http.ResponseWriter, r *http.Request) {
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

	if _, err := h.apps.GetByID(r.Context(), appID, userID); err != nil {
		writeError(w, http.StatusNotFound, "application not found")
		return
	}

	var body struct {
		Days int `json:"days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Days < 1 || body.Days > 90 {
		writeError(w, http.StatusBadRequest, "days must be between 1 and 90")
		return
	}

	until := time.Now().AddDate(0, 0, body.Days)
	if err := h.reminders.Snooze(r.Context(), appID, until); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to snooze reminder")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"snoozed_until": until,
	})
}
