package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/yourname/job-tracker/internal/metrics"
	"github.com/yourname/job-tracker/internal/middleware"
	"github.com/yourname/job-tracker/internal/model"
	"github.com/yourname/job-tracker/internal/store"
)

type ApplicationHandler struct {
	apps     *store.ApplicationStore
	contacts *store.ContactStore
}

func NewApplicationHandler(apps *store.ApplicationStore, contacts *store.ContactStore) *ApplicationHandler {
	return &ApplicationHandler{apps: apps, contacts: contacts}
}

func (h *ApplicationHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body struct {
		Company   string  `json:"company"`
		Role      string  `json:"role"`
		JobURL    *string `json:"job_url"`
		Location  *string `json:"location"`
		Source    string  `json:"source"`
		Notes     *string `json:"notes"`
		AppliedAt *string `json:"applied_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Company == "" || body.Role == "" {
		writeError(w, http.StatusBadRequest, "company and role are required")
		return
	}

	source := model.ApplicationSource(body.Source)
	if body.Source == "" {
		source = model.SourceOther
	}
	if !source.IsValid() {
		writeError(w, http.StatusBadRequest, "invalid source value")
		return
	}

	appliedAt := time.Now()
	if body.AppliedAt != nil {
		t, err := time.Parse("2006-01-02", *body.AppliedAt)
		if err != nil {
			writeError(w, http.StatusBadRequest, "applied_at must be in YYYY-MM-DD format")
			return
		}
		appliedAt = t
	}

	app, err := h.apps.Create(r.Context(), store.CreateApplicationInput{
		UserID:    userID,
		Company:   body.Company,
		Role:      body.Role,
		JobURL:    body.JobURL,
		Location:  body.Location,
		Source:    source,
		Notes:     body.Notes,
		AppliedAt: appliedAt,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create application")
		return
	}

	// Business metric — a new application was logged
	metrics.ApplicationsCreatedTotal.Inc()

	writeJSON(w, http.StatusCreated, app)
}

func (h *ApplicationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	statusFilter := model.ApplicationStatus(r.URL.Query().Get("status"))
	apps, err := h.apps.List(r.Context(), userID, statusFilter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch applications")
		return
	}

	if apps == nil {
		apps = []*model.Application{}
	}
	writeJSON(w, http.StatusOK, apps)
}

func (h *ApplicationHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	app, err := h.apps.GetByID(r.Context(), appID, userID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "application not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch application")
		return
	}

	writeJSON(w, http.StatusOK, app)
}

func (h *ApplicationHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
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

	var body struct {
		Status string  `json:"status"`
		Note   *string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	newStatus := model.ApplicationStatus(body.Status)
	if !newStatus.IsValid() {
		writeError(w, http.StatusBadRequest, "invalid status value")
		return
	}

	history, err := h.apps.UpdateStatus(r.Context(), appID, userID, newStatus, body.Note)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "application not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update status")
		return
	}

	// Business metric — track which status applications move to
	metrics.ApplicationStatusUpdatesTotal.WithLabelValues(body.Status).Inc()

	writeJSON(w, http.StatusOK, history)
}

func (h *ApplicationHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
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

	history, err := h.apps.GetStatusHistory(r.Context(), appID, userID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "application not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch history")
		return
	}

	if history == nil {
		history = []*model.StatusHistory{}
	}
	writeJSON(w, http.StatusOK, history)
}
