// Package handler contains HTTP handlers for the application.
package handler

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/charbelhanna96/go-movies-api/internal/web"
)

type HealthHandler struct {
	db      *sql.DB
	timeout time.Duration
}

func NewHealthHandler(db *sql.DB, timeout time.Duration) *HealthHandler {
	return &HealthHandler{db: db, timeout: timeout}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		slog.Warn("health check failed", "error", err)
		web.JSON(w, http.StatusServiceUnavailable, web.HealthResponse{Status: "unhealthy"})
		return
	}

	web.JSON(w, http.StatusOK, web.HealthResponse{Status: "healthy"})
}
