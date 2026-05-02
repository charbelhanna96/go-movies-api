package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Registry is a custom Prometheus registry that only exposes my metrics.
// It excludes default Go runtime and process metrics.
var Registry = prometheus.NewRegistry()

var factory = promauto.With(Registry)

var (
	RequestCount = factory.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	RequestDuration = factory.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	MoviesReturned = factory.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "movies_returned_per_request",
			Help:    "Number of movies returned per request.",
			Buckets: []float64{1, 5, 10, 20, 50, 100},
		},
	)

	DatabaseQueryDuration = factory.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
)