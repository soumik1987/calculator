// Command server runs the calculator HTTP API.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/soumik1987/sami-calculator/backend/internal/calculator"
	"github.com/soumik1987/sami-calculator/backend/internal/config"
	"github.com/soumik1987/sami-calculator/backend/internal/httpapi"
	"github.com/soumik1987/sami-calculator/backend/internal/telemetry"
)

func main() {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	if err := run(cfg, logger); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, logger *slog.Logger) error {
	providers, err := telemetry.Setup(context.Background(), cfg.OTLPEndpoint != "")
	if err != nil {
		return err
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := providers.Shutdown(ctx); err != nil {
			logger.Error("telemetry shutdown failed", "error", err)
		}
	}()

	handler := httpapi.NewHandler(calculator.Evaluate, logger)
	handler = httpapi.CORS(cfg.AllowedOrigins)(handler)
	handler = httpapi.SecurityHeaders(handler)
	handler = httpapi.Recover(logger)(handler)
	handler = httpapi.Logger(logger)(handler)
	handler = httpapi.Telemetry(providers.TracerProvider, providers.MeterProvider)(handler)

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("listening", "addr", cfg.Addr, "allowedOrigins", cfg.AllowedOrigins, "logLevel", cfg.LogLevel.String(), "otlpEndpoint", cfg.OTLPEndpoint)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	logger.Info("shutting down")
	return srv.Shutdown(shutdownCtx)
}
