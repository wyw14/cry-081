package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type HTTP struct {
	Address         string
	AllowedOrigin   string
	RequestTimeout  time.Duration
	ShutdownTimeout time.Duration
	RequestsPerMin  int
}

type Tokens struct {
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type Storage struct {
	Driver      string
	DatabaseURL string
}

type Config struct {
	Environment string
	HTTP        HTTP
	Tokens      Tokens
	Storage     Storage
}

type environment interface {
	LookupEnv(string) (string, bool)
}

type processEnvironment struct{}

func (processEnvironment) LookupEnv(key string) (string, bool) { return os.LookupEnv(key) }

func Load() (Config, error) {
	return read(processEnvironment{})
}

func read(source environment) (Config, error) {
	value := Config{
		Environment: textValue(source, "APP_ENV", "development"),
		HTTP: HTTP{
			Address:         textValue(source, "HTTP_ADDRESS", ":8080"),
			AllowedOrigin:   textValue(source, "ALLOWED_ORIGIN", "http://localhost:5173"),
			RequestTimeout:  10 * time.Second,
			ShutdownTimeout: 10 * time.Second,
			RequestsPerMin:  120,
		},
		Tokens: Tokens{AccessTTL: 15 * time.Minute, RefreshTTL: 30 * 24 * time.Hour},
		Storage: Storage{
			Driver:      strings.ToLower(textValue(source, "STORAGE_MODE", "memory")),
			DatabaseURL: textValue(source, "DATABASE_URL", "postgres://editorial:editorial@localhost:5432/editorial?sslmode=disable"),
		},
	}

	durations := []struct {
		key    string
		target *time.Duration
	}{
		{"ACCESS_TOKEN_TTL", &value.Tokens.AccessTTL},
		{"REFRESH_TOKEN_TTL", &value.Tokens.RefreshTTL},
		{"REQUEST_TIMEOUT", &value.HTTP.RequestTimeout},
		{"SHUTDOWN_TIMEOUT", &value.HTTP.ShutdownTimeout},
	}
	for _, item := range durations {
		raw, exists := source.LookupEnv(item.key)
		if !exists || strings.TrimSpace(raw) == "" {
			continue
		}
		parsed, err := time.ParseDuration(raw)
		if err != nil || parsed <= 0 {
			return Config{}, fmt.Errorf("%s must be a positive duration", item.key)
		}
		*item.target = parsed
	}

	if raw, exists := source.LookupEnv("RATE_LIMIT_PER_MINUTE"); exists && strings.TrimSpace(raw) != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			return Config{}, fmt.Errorf("RATE_LIMIT_PER_MINUTE must be a positive integer")
		}
		value.HTTP.RequestsPerMin = parsed
	}
	if value.Storage.Driver != "memory" && value.Storage.Driver != "postgres" {
		return Config{}, fmt.Errorf("STORAGE_MODE must be memory or postgres")
	}
	if value.Storage.Driver == "postgres" && strings.TrimSpace(value.Storage.DatabaseURL) == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required for postgres storage")
	}
	return value, nil
}

func textValue(source environment, key, fallback string) string {
	value, exists := source.LookupEnv(key)
	if !exists || strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}
