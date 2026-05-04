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
	"github.com/charbelhanna96/go-movies-api/internal/handler"
	"github.com/charbelhanna96/go-movies-api/internal/kafka"
	"github.com/charbelhanna96/go-movies-api/internal/metrics"
	"github.com/charbelhanna96/go-movies-api/internal/middleware"
	"github.com/charbelhanna96/go-movies-api/internal/repository"
	commondb "github.com/charbelhanna96/go-movies-common/pkg/db"
	commonmiddleware "github.com/charbelhanna96/go-movies-common/pkg/middleware"
	"github.com/charbelhanna96/go-movies-common/pkg/tracing"
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

	shutdownTracing, err := tracing.Setup(context.Background(), cfg.OtelEndpoint, "go-movies-api")
	if err != nil {
		slog.Error("failed to setup tracing", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := shutdownTracing(context.Background()); err != nil {
			slog.Error("failed to shutdown tracing", "error", err)
		}
	}()

	if err := cfg.Validate(); err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	database, err := commondb.Connect(cfg.Database)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := database.Close(); err != nil {
			slog.Error("failed to close database", "error", err)
		}
	}()

	kafkaProducer, err := kafka.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		slog.Error("failed to create kafka producer", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := kafkaProducer.Close(); err != nil {
			slog.Error("failed to close kafka producer", "error", err)
		}
	}()

	movieRepo := repository.NewPostgresMovieRepository(database)

	healthHandler := handler.NewHealthHandler(database, cfg.HandlersConfig.HealthTimeout)
	moviesHandler := handler.NewMoviesHandler(movieRepo, kafkaProducer, cfg.HandlersConfig.MoviesTimeout)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler.Health)
	mux.Handle("GET /metrics", promhttp.HandlerFor(
		metrics.Registry,
		promhttp.HandlerOpts{},
	))
	mux.HandleFunc("GET /api/v1/movies", moviesHandler.GetMovies)

	server := &http.Server{
		Addr: ":" + cfg.Port,
		Handler: commonmiddleware.CORS(
			cfg.AllowedOrigins,
			commonmiddleware.Tracing("go-movies-api", middleware.Metrics(mux)),
		),
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
