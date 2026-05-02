// Package db provides functions for connecting to and interacting with the database.
package db

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/charbelhanna96/go-movies-api/internal/config"
	_ "github.com/lib/pq"
)

// Connect creates and verifies a PostgreSQL connection pool.
func Connect(conf config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		conf.Host,
		conf.Port,
		conf.Name,
		conf.User,
		conf.Password,
	)

	database, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database connection: %w", err)
	}

	database.SetMaxOpenConns(conf.MaxOpenConns)
	database.SetMaxIdleConns(conf.MaxIdleConns)
	database.SetConnMaxLifetime(conf.ConnMaxLifetime)
	database.SetConnMaxIdleTime(conf.ConnMaxIdleTime)

	if err := database.Ping(); err != nil {
		if err := database.Close(); err != nil {
			slog.Error("failed to close database connection", "error", err)
		}
		return nil, fmt.Errorf("ping database: %w", err)
	}

	slog.Info("database connection established", "host", conf.Host, "name", conf.Name)

	return database, nil
}
