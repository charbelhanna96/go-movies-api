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

## Architecture

The backend is organized into small, focused layers:

```
HTTP -> Handler -> Validator -> Repository -> Database
```

- **Middleware** handles CORS
- **Handler** orchestrates request and response logic
- **Validator** parses and validates query parameters
- **Repository** is the only layer that talks to PostgreSQL
- **main.go** is the composition root that wires everything together

Configuration is loaded from environment variables. Required database credentials are validated at startup. The service shuts down gracefully on SIGTERM, draining in-flight requests before exiting.

## Getting Started

Make sure you have Docker and Docker Compose installed.

```bash
git clone https://github.com/charbelhanna96/go-movies-api
cd go-movies-api
docker compose up --build
```

The API will be available at http://localhost:8080

Swagger UI will be available at http://localhost:8081

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
| limit | integer | Maximum number of results |

Example requests:

```bash
# All movies
curl http://localhost:8080/api/v1/movies

# Action movies with rating above 4.5
curl "http://localhost:8080/api/v1/movies?genres=1&min-rating=4.5"

# Christopher Nolan movies released after 2000
curl "http://localhost:8080/api/v1/movies?directors=1&min-year=2000"

# Top 5 movies under 2 hours
curl "http://localhost:8080/api/v1/movies?max-duration=120&limit=5"
```

### GET /health

Returns the health status of the service including database connectivity.

```bash
curl http://localhost:8080/health
```

### GET /metrics

Exposes Prometheus metrics for scraping. Intended for use with a Prometheus server or compatible monitoring tool.

Metrics exposed:

- `http_requests_total` - total HTTP requests by method, path, and status code
- `http_request_duration_seconds` - HTTP request duration histogram by method and path
- `movies_returned_per_request` - histogram of movies returned per search request
- `db_query_duration_seconds` - database query duration histogram by operation

```bash
curl http://localhost:8080/metrics
```

## Security

- All inputs are validated server-side before reaching the database
- SQL injection is prevented through parameterized queries
- CORS is configured with an allowlist
- Database credentials are required at startup and never hardcoded
- HTTP timeouts are configured to prevent hanging requests

## Dataset

The database is seeded with 22 movies across 10 genres directed by 10 directors including Christopher Nolan, Quentin Tarantino, Martin Scorsese, and others.