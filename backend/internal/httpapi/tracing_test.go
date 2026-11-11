package httpapi_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/soumik1987/sami-calculator/backend/internal/calculator"
	"github.com/soumik1987/sami-calculator/backend/internal/httpapi"
)

var traceIDPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

func tracedServer() (http.Handler, *tracetest.SpanRecorder) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	h := httpapi.NewHandler(calculator.Evaluate, slog.New(slog.NewTextHandler(io.Discard, nil)))
	return httpapi.Telemetry(tp, metricnoop.NewMeterProvider())(h), recorder
}

func TestTelemetryReturnsTraceID(t *testing.T) {
	h, recorder := tracedServer()

	rec := do(t, h, http.MethodPost, "/api/v1/calculations", `{"expression":"1+1"}`)

	got := rec.Header().Get("X-Trace-ID")
	if !traceIDPattern.MatchString(got) {
		t.Fatalf("got X-Trace-ID %q, want 32 hex characters", got)
	}
	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	if spans[0].Name() != "POST /api/v1/calculations" {
		t.Errorf("got span name %q", spans[0].Name())
	}
	if spans[0].SpanContext().TraceID().String() != got {
		t.Error("header and span have different trace IDs")
	}
}

func TestTelemetryContinuesIncomingTrace(t *testing.T) {
	h, _ := tracedServer()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculations", strings.NewReader(`{"expression":"1+1"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Trace-ID"); got != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Errorf("got trace ID %q, want the caller's", got)
	}
}

func TestTelemetryRecordsErrorType(t *testing.T) {
	h, recorder := tracedServer()

	do(t, h, http.MethodPost, "/api/v1/calculations", `{"expression":"1/0"}`)

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	want := attribute.String("error.type", "division_by_zero")
	for _, a := range spans[0].Attributes() {
		if a == want {
			return
		}
	}
	t.Errorf("span attributes %v do not include %v", spans[0].Attributes(), want)
}

func TestTelemetrySkipsHealthChecks(t *testing.T) {
	h, recorder := tracedServer()

	rec := do(t, h, http.MethodGet, "/healthz", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d", rec.Code)
	}
	if n := len(recorder.Ended()); n != 0 {
		t.Errorf("got %d spans for a health check, want 0", n)
	}
}
