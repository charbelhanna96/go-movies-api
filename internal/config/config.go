// Package config provides functionality to load application configuration from environment variables.
package config

import (
	"fmt"
	"time"

	commonconfig "github.com/charbelhanna96/go-movies-common/pkg/config"
	commondb "github.com/charbelhanna96/go-movies-common/pkg/db"
	"github.com/joho/godotenv"
)

// Config holds the application configuration loaded from environment variables
type Config struct {
	Port           string
	AllowedOrigins []string
	Database       commondb.DatabaseConfig
	HTTPConfig     HTTPConfig
	HandlersConfig HandlersConfig
	AppConfig      AppConfig
	LogLevel       string
	OtelEndpoint   string
	KafkaBrokers   []string
}

type AppConfig struct {
	ShutdownTimeout time.Duration
}

type HTTPConfig struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
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
		Port:           commonconfig.GetEnv("PORT", "8080"),
		AllowedOrigins: commonconfig.GetEnvList("ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		LogLevel:       commonconfig.GetEnv("LOG_LEVEL", "info"),
		OtelEndpoint:   commonconfig.GetEnv("OTEL_ENDPOINT", "jaeger:4318"),
		KafkaBrokers:   commonconfig.GetEnvList("KAFKA_BROKERS", []string{"kafka:9092"}),
		AppConfig: AppConfig{
			ShutdownTimeout: time.Duration(commonconfig.GetEnvInt("APP_SHUTDOWN_TIMEOUT_SEC", 15)) * time.Second,
		},
		Database: commondb.DatabaseConfig{
			Host:            commonconfig.GetEnv("DATABASE_HOSTNAME", ""),
			Port:            commonconfig.GetEnv("DATABASE_PORT", "5432"),
			Name:            commonconfig.GetEnv("DATABASE_NAME", ""),
			User:            commonconfig.GetEnv("DATABASE_USERNAME", ""),
			Password:        commonconfig.GetEnv("DATABASE_PASSWORD", ""),
			MaxOpenConns:    commonconfig.GetEnvInt("DATABASE_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    commonconfig.GetEnvInt("DATABASE_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: time.Duration(commonconfig.GetEnvInt("DATABASE_CONN_MAX_LIFETIME_MIN", 5)) * time.Minute,
			ConnMaxIdleTime: time.Duration(commonconfig.GetEnvInt("DATABASE_CONN_MAX_IDLE_TIME_MIN", 1)) * time.Minute,
		},
		HTTPConfig: HTTPConfig{
			ReadTimeout:  time.Duration(commonconfig.GetEnvInt("HTTP_READ_TIMEOUT_SEC", 10)) * time.Second,
			WriteTimeout: time.Duration(commonconfig.GetEnvInt("HTTP_WRITE_TIMEOUT_SEC", 10)) * time.Second,
			IdleTimeout:  time.Duration(commonconfig.GetEnvInt("HTTP_IDLE_TIMEOUT_SEC", 60)) * time.Second,
		},
		HandlersConfig: HandlersConfig{
			HealthTimeout: time.Duration(commonconfig.GetEnvInt("HEALTH_HANDLER_TIMEOUT_SEC", 2)) * time.Second,
			MoviesTimeout: time.Duration(commonconfig.GetEnvInt("MOVIES_HANDLER_TIMEOUT_SEC", 5)) * time.Second,
		},
	}
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
