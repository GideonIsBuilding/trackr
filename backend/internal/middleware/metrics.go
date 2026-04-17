package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/yourname/job-tracker/internal/metrics"
)

// PrometheusMiddleware wraps every HTTP handler and records:
//   - request count (by method, route, status)
//   - request duration (by method, route)
//   - in-flight requests (gauge)
//
// It uses a responseWriter wrapper to capture the status code after the
// handler runs, since Go's http.ResponseWriter doesn't expose it natively.
func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Track in-flight requests
		metrics.HTTPRequestsInFlight.Inc()
		defer metrics.HTTPRequestsInFlight.Dec()

		// Wrap the ResponseWriter so we can capture the status code
		wrapped := &statusCapture{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()
		route := r.URL.Path
		method := r.Method
		status := fmt.Sprintf("%d", wrapped.status)

		metrics.HTTPRequestsTotal.WithLabelValues(method, route, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(method, route).Observe(duration)
	})
}

// statusCapture wraps http.ResponseWriter to intercept the written status code.
type statusCapture struct {
	http.ResponseWriter
	status int
}

func (s *statusCapture) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}
