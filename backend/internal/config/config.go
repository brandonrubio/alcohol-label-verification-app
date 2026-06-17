package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr              string
	DatabaseURL       string
	NeonAuthJWKSURL   string
	GeminiAPIKey      string
	GeminiModel       string
	AllowedOrigins    []string
	MaxUploadBytes    int64
	RequestTimeout    time.Duration
	AIProvider        string
	BatchConcurrency  int
	SkipAuthInDev     bool
}

func Load() (Config, error) {
	cfg := Config{
		Addr:             envOr("ADDR", ":8080"),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		NeonAuthJWKSURL:  os.Getenv("NEON_AUTH_JWKS_URL"),
		GeminiAPIKey:     os.Getenv("GEMINI_API_KEY"),
		GeminiModel:      envOr("GEMINI_MODEL", "gemini-2.5-flash"),
		AIProvider:       envOr("AI_PROVIDER", "gemini"),
		BatchConcurrency: envIntOr("BATCH_CONCURRENCY", 4),
		SkipAuthInDev:    envOr("SKIP_AUTH_IN_DEV", "false") == "true",
	}

	origins := envOr("ALLOWED_ORIGINS", "http://localhost:5173")
	cfg.AllowedOrigins = splitCSV(origins)

	maxUpload, err := strconv.ParseInt(envOr("MAX_UPLOAD_BYTES", "10485760"), 10, 64)
	if err != nil {
		return Config{}, fmt.Errorf("parse MAX_UPLOAD_BYTES: %w", err)
	}
	cfg.MaxUploadBytes = maxUpload

	timeoutSeconds, err := strconv.Atoi(envOr("REQUEST_TIMEOUT_SECONDS", "30"))
	if err != nil {
		return Config{}, fmt.Errorf("parse REQUEST_TIMEOUT_SECONDS: %w", err)
	}
	cfg.RequestTimeout = time.Duration(timeoutSeconds) * time.Second

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.NeonAuthJWKSURL == "" && !cfg.SkipAuthInDev {
		return Config{}, fmt.Errorf("NEON_AUTH_JWKS_URL is required unless SKIP_AUTH_IN_DEV=true")
	}

	if cfg.AIProvider == "gemini" && cfg.GeminiAPIKey == "" {
		return Config{}, fmt.Errorf("GEMINI_API_KEY is required when AI_PROVIDER=gemini")
	}

	return cfg, nil
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envIntOr(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
