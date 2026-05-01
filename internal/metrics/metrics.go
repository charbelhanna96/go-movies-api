// Package metrics provides Prometheus metrics for the application.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics - recorded by middleware
	RequestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Business metric - recorded by handler
	MoviesReturned = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "movies_returned_per_request",
			Help:    "Number of movies returned per request.",
			Buckets: []float64{1, 5, 10, 20, 50, 100},
		},
	)

	// Database metric - recorded by repository
	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
)
