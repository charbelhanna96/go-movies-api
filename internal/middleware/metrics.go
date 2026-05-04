package middleware

import (
	"net/http"

	"github.com/charbelhanna96/go-movies-api/internal/metrics"
	commonmiddleware "github.com/charbelhanna96/go-movies-common/pkg/middleware"
)

// Metrics wraps the common metrics middleware with this service's metrics registry.
func Metrics(next http.Handler) http.Handler {
	return commonmiddleware.Metrics(
		metrics.RequestCount,
		metrics.RequestDuration,
		next,
	)
}
