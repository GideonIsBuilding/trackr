package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/yourname/job-tracker/internal/middleware"
	"github.com/yourname/job-tracker/internal/model"
	"github.com/yourname/job-tracker/internal/store"
)

// AnalyticsHandler serves GET /api/analytics
type AnalyticsHandler struct {
	analytics *store.AnalyticsStore
}

func NewAnalyticsHandler(analytics *store.AnalyticsStore) *AnalyticsHandler {
	return &AnalyticsHandler{analytics: analytics}
}

func (h *AnalyticsHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	summary, err := h.analytics.GetSummary(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to compute analytics")
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

// ChecklistHandler serves PATCH /api/applications/{id}/checklist
type ChecklistHandler struct {
	apps *store.ApplicationStore
}

func NewChecklistHandler(apps *store.ApplicationStore) *ChecklistHandler {
	return &ChecklistHandler{apps: apps}
}

func (h *ChecklistHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	var checklist model.ApplicationChecklist
	if err := json.NewDecoder(r.Body).Decode(&checklist); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	app, err := h.apps.UpdateChecklist(r.Context(), appID, userID, checklist)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "application not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update checklist")
		return
	}

	writeJSON(w, http.StatusOK, app)
}
