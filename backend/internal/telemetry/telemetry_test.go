package telemetry_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/soumik1987/sami-calculator/backend/internal/telemetry"
)

func TestSetupWithoutExportStillCreatesTraceIDs(t *testing.T) {
	p, err := telemetry.Setup(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = p.Shutdown(context.Background()) }()

	_, span := p.TracerProvider.Tracer("test").Start(context.Background(), "work")
	span.End()

	if !span.SpanContext().TraceID().IsValid() {
		t.Error("span has no trace ID")
	}
}

func TestSetupExportsTracesOverOTLP(t *testing.T) {
	var mu sync.Mutex
	var paths []string
	collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.Path)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer collector.Close()
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", collector.URL)

	p, err := telemetry.Setup(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}
	_, span := p.TracerProvider.Tracer("test").Start(context.Background(), "work")
	span.End()
	if err := p.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}

	mu.Lock()
	defer mu.Unlock()
	for _, path := range paths {
		if path == "/v1/traces" {
			return
		}
	}
	t.Errorf("collector got %v, want a request to /v1/traces", paths)
}
