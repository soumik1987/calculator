// Package config loads service settings from the environment.
package config

import (
	"fmt"
	"log/slog"
	"strings"
)

const (
	envPort           = "PORT"
	envAllowedOrigins = "ALLOWED_ORIGINS"
	envLogLevel       = "LOG_LEVEL"
	envOTLPEndpoint   = "OTEL_EXPORTER_OTLP_ENDPOINT"

	defaultPort           = "8081"
	defaultAllowedOrigins = "http://localhost:8080"
	defaultLogLevel       = slog.LevelInfo
)

// Config holds the service settings.
type Config struct {
	Addr           string
	AllowedOrigins []string
	LogLevel       slog.Level
	OTLPEndpoint   string
}

// Load reads configuration using getenv, typically os.Getenv.
func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		Addr:           ":" + valueOr(getenv(envPort), defaultPort),
		AllowedOrigins: splitList(valueOr(getenv(envAllowedOrigins), defaultAllowedOrigins)),
		LogLevel:       defaultLogLevel,
		OTLPEndpoint:   getenv(envOTLPEndpoint),
	}

	if raw := getenv(envLogLevel); raw != "" {
		if err := cfg.LogLevel.UnmarshalText([]byte(raw)); err != nil {
			return Config{}, fmt.Errorf("%s must be debug, info, warn or error, got %q", envLogLevel, raw)
		}
	}
	return cfg, nil
}

func valueOr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}
