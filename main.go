package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/charbelhanna96/go-movies-api/internal/config"
	"github.com/charbelhanna96/go-movies-api/internal/db"
	"github.com/charbelhanna96/go-movies-api/internal/handler"
	"github.com/charbelhanna96/go-movies-api/internal/middleware"
	"github.com/charbelhanna96/go-movies-api/internal/repository"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg := config.Load()

	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		level = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})))

	if err := cfg.Validate(); err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	database, err := db.Connect(cfg.Database)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	movieRepo := repository.NewPostgresMovieRepository(database)

	healthHandler := handler.NewHealthHandler(database, cfg.HandlersConfig.HealthTimeout)
	moviesHandler := handler.NewMoviesHandler(movieRepo, cfg.HandlersConfig.MoviesTimeout)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler.Health)
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.HandleFunc("GET /api/v1/movies", moviesHandler.GetMovies)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      middleware.CORS(cfg.AllowedOrigins, middleware.Metrics(mux)),
		ReadTimeout:  cfg.HTTPConfig.ReadTimeout,
		WriteTimeout: cfg.HTTPConfig.WriteTimeout,
		IdleTimeout:  cfg.HTTPConfig.IdleTimeout,
	}

	go func() {
		slog.Info("backend service listening", "port", cfg.Port)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	slog.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.AppConfig.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}
