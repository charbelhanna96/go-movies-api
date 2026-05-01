// Package config provides functionality to load application configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the application configuration loaded from environment variables
type Config struct {
	Port           string
	AllowedOrigins []string
	Database       DatabaseConfig
	HTTPConfig     HTTPConfig
	HandlersConfig HandlersConfig
	AppConfig      AppConfig
	LogLevel       string
	OtelEndpoint   string
}

type AppConfig struct {
	ShutdownTimeout time.Duration
}

type HTTPConfig struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            string
	Name            string
	User            string
	Password        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type HandlersConfig struct {
	HealthTimeout time.Duration
	MoviesTimeout time.Duration
}

// Load reads configuration from environment variables.
// If envFile is omitted, it attempts to load ".env".
// If envFile is an empty string, no env file is loaded.
// If envFile is provided, that file is loaded.
func Load(envFile ...string) Config {
	if len(envFile) == 0 {
		_ = godotenv.Load()
	} else if envFile[0] != "" {
		_ = godotenv.Overload(envFile[0])
	}

	return Config{
		Port:           getEnv("PORT", "8080"),
		AllowedOrigins: getEnvList("ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		OtelEndpoint:   getEnv("OTEL_ENDPOINT", "jaeger:4318"),
		AppConfig: AppConfig{
			ShutdownTimeout: time.Duration(getEnvInt("APP_SHUTDOWN_TIMEOUT_SEC", 15)) * time.Second,
		},
		Database: DatabaseConfig{
			Host:            getEnv("DATABASE_HOSTNAME", ""),
			Port:            getEnv("DATABASE_PORT", "5432"),
			Name:            getEnv("DATABASE_NAME", ""),
			User:            getEnv("DATABASE_USERNAME", ""),
			Password:        getEnv("DATABASE_PASSWORD", ""),
			MaxOpenConns:    getEnvInt("DATABASE_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DATABASE_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: time.Duration(getEnvInt("DATABASE_CONN_MAX_LIFETIME_MIN", 5)) * time.Minute,
			ConnMaxIdleTime: time.Duration(getEnvInt("DATABASE_CONN_MAX_IDLE_TIME_MIN", 1)) * time.Minute,
		},
		HTTPConfig: HTTPConfig{
			ReadTimeout:  time.Duration(getEnvInt("HTTP_READ_TIMEOUT_SEC", 10)) * time.Second,
			WriteTimeout: time.Duration(getEnvInt("HTTP_WRITE_TIMEOUT_SEC", 10)) * time.Second,
			IdleTimeout:  time.Duration(getEnvInt("HTTP_IDLE_TIMEOUT_SEC", 60)) * time.Second,
		},
		HandlersConfig: HandlersConfig{
			HealthTimeout: time.Duration(getEnvInt("HEALTH_HANDLER_TIMEOUT_SEC", 2)) * time.Second,
			MoviesTimeout: time.Duration(getEnvInt("MOVIES_HANDLER_TIMEOUT_SEC", 5)) * time.Second,
		},
	}
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvList(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	if len(result) == 0 {
		return fallback
	}

	return result
}

func getEnvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func (c Config) Validate() error {
	if c.Database.Password == "" {
		return fmt.Errorf("DATABASE_PASSWORD is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("DATABASE_USERNAME is required")
	}
	if c.Database.Host == "" {
		return fmt.Errorf("DATABASE_HOSTNAME is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("DATABASE_NAME is required")
	}
	return nil
}
