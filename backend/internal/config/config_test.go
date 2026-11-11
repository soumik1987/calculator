package config_test

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/soumik1987/sami-calculator/backend/internal/config"
)

func env(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func TestLoadDefaults(t *testing.T) {
	got, err := config.Load(env(nil))
	if err != nil {
		t.Fatal(err)
	}

	want := config.Config{
		Addr:           ":8081",
		AllowedOrigins: []string{"http://localhost:8080"},
		LogLevel:       slog.LevelInfo,
		OTLPEndpoint:   "",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	got, err := config.Load(env(map[string]string{
		"PORT":                        "9000",
		"ALLOWED_ORIGINS":             "https://a.example, https://b.example",
		"LOG_LEVEL":                   "WARN",
		"OTEL_EXPORTER_OTLP_ENDPOINT": "http://collector:4318",
	}))
	if err != nil {
		t.Fatal(err)
	}

	want := config.Config{
		Addr:           ":9000",
		AllowedOrigins: []string{"https://a.example", "https://b.example"},
		LogLevel:       slog.LevelWarn,
		OTLPEndpoint:   "http://collector:4318",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLoadLogLevels(t *testing.T) {
	tests := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}
	for value, want := range tests {
		got, err := config.Load(env(map[string]string{"LOG_LEVEL": value}))
		if err != nil {
			t.Fatalf("LOG_LEVEL=%q: %v", value, err)
		}
		if got.LogLevel != want {
			t.Errorf("LOG_LEVEL=%q: got %v, want %v", value, got.LogLevel, want)
		}
	}
}

func TestLoadRejectsInvalidLogLevel(t *testing.T) {
	if _, err := config.Load(env(map[string]string{"LOG_LEVEL": "verbose"})); err == nil {
		t.Error("expected error for unknown log level")
	}
}
