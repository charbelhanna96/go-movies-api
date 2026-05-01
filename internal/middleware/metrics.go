package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/charbelhanna96/go-movies-api/internal/metrics"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Metrics is a middleware that records Prometheus metrics for each request.
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(recorder.status)

		metrics.RequestCount.WithLabelValues(r.Method, r.URL.Path, status).Inc()
		metrics.RequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}
