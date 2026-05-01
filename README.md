# go-movies-api

A production-ready REST API for movie recommendations built in Go.

## Overview

This project demonstrates a clean, layered backend architecture for a movie search API. Movies are returned ordered by rating and can be filtered by director, genre, release year, duration, and rating.

## Stack

- Go 1.25
- PostgreSQL
- Docker and Docker Compose
- go-playground/validator for input validation
- lib/pq for PostgreSQL connectivity
- Prometheus for metrics
- OpenTelemetry with Jaeger for distributed tracing
- testcontainers-go for integration tests

## Architecture

The backend is organized into small, focused layers:

```
HTTP -> Middleware -> Handler -> Validator -> Repository -> Database
```

- **Middleware** handles CORS, Prometheus metrics collection, and OpenTelemetry trace propagation
- **Handler** orchestrates request and response logic, creates child spans for tracing
- **Validator** parses and validates all query parameters before they reach the handler
- **Repository** is the only layer that talks to PostgreSQL, records query duration metrics and database spans
- **main.go** is the composition root that wires everything together

Configuration is loaded from environment variables. Required database credentials are validated at startup. The service shuts down gracefully on SIGTERM, draining in-flight requests before exiting.

## Getting Started

Make sure you have Docker and Docker Compose installed.

```bash
git clone https://github.com/charbelhanna96/go-movies-api
cd go-movies-api
docker compose up --build
```

- API: http://localhost:8080
- Swagger UI: http://localhost:8081
- Jaeger UI: http://localhost:16686
- Prometheus metrics: http://localhost:8080/metrics

## API

### GET /api/v1/movies

Returns movies ordered by rating descending. All parameters are optional and can be combined.

| Parameter | Type | Description |
|---|---|---|
| directors | string | Comma-separated director IDs |
| genres | string | Comma-separated genre IDs |
| min-year | integer | Minimum release year (1888-2100) |
| max-year | integer | Maximum release year (1888-2100) |
| min-duration | integer | Minimum duration in minutes |
| max-duration | integer | Maximum duration in minutes |
| min-rating | float | Minimum rating (0.0-5.0) |
| max-rating | float | Maximum rating (0.0-5.0) |
| limit | integer | Maximum number of results (1-1000) |

Example requests:

```bash
# All movies
curl http://localhost:8080/api/v1/movies

# Movies with rating above 4.5
curl "http://localhost:8080/api/v1/movies?min-rating=4.5"

# Director 1 movies released after 2000
curl "http://localhost:8080/api/v1/movies?directors=1&min-year=2000"

# Top 5 movies under 2 hours
curl "http://localhost:8080/api/v1/movies?max-duration=120&limit=5"

# Action or SciFi movies (genre IDs 1 and 5)
curl "http://localhost:8080/api/v1/movies?genres=1,5"
```

### GET /health

Returns the health status of the service including database connectivity.

```bash
curl http://localhost:8080/health
```

### GET /metrics

Exposes Prometheus metrics for scraping.

Metrics exposed:

- `http_requests_total` - total HTTP requests by method, path, and status code
- `http_request_duration_seconds` - HTTP request duration histogram by method and path
- `movies_returned_per_request` - histogram of movies returned per search request
- `db_query_duration_seconds` - database query duration histogram by operation

```bash
curl http://localhost:8080/metrics
```

## Observability

### Metrics

The service exposes Prometheus metrics at `/metrics`. Metrics are recorded at three levels:

- HTTP middleware records request count and duration for every request
- Handler records the number of movies returned per search
- Repository records database query duration per operation

### Tracing

The service uses OpenTelemetry to emit traces to Jaeger. Every request produces a trace with the following span hierarchy:

```
GET /api/v1/movies
├── ParseMovieFilters
└── GetMovies
    └── db.QueryMovies
```

Each span includes relevant attributes such as filter counts, movies returned, and HTTP metadata. Error spans are marked with status codes and descriptions. Trace context is propagated via standard HTTP headers, enabling end-to-end tracing across multiple services.

Open the Jaeger UI at http://localhost:16686 to inspect traces.

## Testing

Unit tests cover the validator and handler layers. Integration tests cover the repository layer using real PostgreSQL via testcontainers.

```bash
# Unit tests
go test ./internal/validator/... ./internal/handler/... -v

# Integration tests (requires Docker)
go test ./internal/repository/... -v -timeout 120s -tags integration
```

## Security

- All inputs are validated server-side before reaching the database
- SQL injection is prevented through parameterized queries
- CORS is configured with an origin allowlist
- Database credentials are required at startup and never hardcoded
- HTTP timeouts are configured on read, write, and idle connections
- Handler timeouts prevent long-running requests from blocking resources

## Dataset

The database is seeded with 22 movies across 10 genres directed by 10 directors.