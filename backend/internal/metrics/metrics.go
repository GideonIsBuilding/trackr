package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// All Trackr metrics are defined here and registered automatically via promauto.
// promauto registers with the default registry so no manual registration needed.

// ── HTTP metrics ──────────────────────────────────────────────────────────────

// HTTPRequestsTotal counts every request by method, route, and status code.
// Use this to calculate error rate and traffic volume.
var HTTPRequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "trackr_http_requests_total",
		Help: "Total number of HTTP requests",
	},
	[]string{"method", "route", "status"},
)

// HTTPRequestDuration measures how long each request takes, bucketed by route.
// Use this to calculate latency percentiles (p50, p95, p99).
var HTTPRequestDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "trackr_http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets, // .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10
	},
	[]string{"method", "route"},
)

// HTTPRequestsInFlight tracks how many requests are being handled right now.
// Use this to detect traffic spikes and saturation.
var HTTPRequestsInFlight = promauto.NewGauge(
	prometheus.GaugeOpts{
		Name: "trackr_http_requests_in_flight",
		Help: "Number of HTTP requests currently being processed",
	},
)

// ── Business metrics ──────────────────────────────────────────────────────────

// ApplicationsCreatedTotal counts every new application logged.
// Business signal: is the product being used?
var ApplicationsCreatedTotal = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "trackr_applications_created_total",
		Help: "Total number of job applications logged",
	},
)

// ApplicationStatusUpdatesTotal counts status transitions by which status they moved to.
// Helps understand how applications progress through the funnel.
var ApplicationStatusUpdatesTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "trackr_application_status_updates_total",
		Help: "Total number of application status updates",
	},
	[]string{"to_status"},
)

// UserRegistrationsTotal counts new user sign-ups.
var UserRegistrationsTotal = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "trackr_user_registrations_total",
		Help: "Total number of user registrations",
	},
)

// ── Reminder engine metrics ───────────────────────────────────────────────────

// ReminderScansTotal counts how many times the engine has run its scan loop.
var ReminderScansTotal = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "trackr_reminder_scans_total",
		Help: "Total number of reminder engine scan cycles",
	},
)

// RemindersFiredTotal counts how many reminder notifications were actually sent.
var RemindersFiredTotal = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "trackr_reminders_fired_total",
		Help: "Total number of reminder notifications sent",
	},
)

// ReminderScanDuration measures how long each scan cycle takes.
// A slow scan means the DB query is getting expensive.
var ReminderScanDuration = promauto.NewHistogram(
	prometheus.HistogramOpts{
		Name:    "trackr_reminder_scan_duration_seconds",
		Help:    "Duration of each reminder engine scan cycle",
		Buckets: []float64{.001, .005, .01, .05, .1, .5, 1, 2},
	},
)

// ── AI extraction metrics ─────────────────────────────────────────────────────

// ExtractionRequestsTotal counts calls to the /api/extract endpoint,
// labelled by whether the quick title parser handled it or Gemini was needed.
var ExtractionRequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "trackr_extraction_requests_total",
		Help: "Total number of job data extraction requests",
	},
	[]string{"method"}, // "quick_parse" or "gemini"
)

// ExtractionDuration measures how long each extraction takes.
// Quick parses should be <1ms; Gemini calls 1-5s.
var ExtractionDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "trackr_extraction_duration_seconds",
		Help:    "Duration of job data extraction requests",
		Buckets: []float64{.001, .01, .1, .5, 1, 2, 5, 10, 15},
	},
	[]string{"method"},
)

// ExtractionErrorsTotal counts failed extraction attempts.
var ExtractionErrorsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "trackr_extraction_errors_total",
		Help: "Total number of failed extraction requests",
	},
	[]string{"stage"}, // "jina" or "gemini"
)

// ── Database metrics ──────────────────────────────────────────────────────────

// DBQueryDuration measures how long DB queries take, labelled by operation.
var DBQueryDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "trackr_db_query_duration_seconds",
		Help:    "Database query duration in seconds",
		Buckets: []float64{.001, .005, .01, .05, .1, .25, .5, 1},
	},
	[]string{"operation"},
)
