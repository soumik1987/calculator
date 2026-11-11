package httpapi_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	metricnoop "go.opentelemetry.io/otel/metric/noop"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/soumik1987/sami-calculator/backend/internal/httpapi"
)

func TestSecurityHeaders(t *testing.T) {
	rec := httptest.NewRecorder()

	httpapi.SecurityHeaders(okHandler).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	want := map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"Content-Security-Policy": "default-src 'none'; frame-ancestors 'none'",
		"Cache-Control":           "no-store",
		"Referrer-Policy":         "no-referrer",
	}
	for header, value := range want {
		if got := rec.Header().Get(header); got != value {
			t.Errorf("%s: got %q, want %q", header, got, value)
		}
	}
}

func TestRecoverReturnsInternalError(t *testing.T) {
	panicking := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") })
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	rec := httptest.NewRecorder()

	httpapi.Recover(logger)(panicking).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got status %d, want 500", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"internal_error"`) {
		t.Errorf("got body %s", rec.Body)
	}
	if strings.Contains(rec.Body.String(), "boom") {
		t.Error("panic value leaked to client")
	}
}

func TestLoggerRecordsRequest(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	teapot := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	httpapi.Logger(logger)(teapot).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/v1/calculations", nil))

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("log is not JSON: %v: %s", err, buf.String())
	}
	if entry["method"] != "POST" || entry["path"] != "/api/v1/calculations" || entry["status"] != float64(http.StatusTeapot) {
		t.Errorf("unexpected log entry: %v", entry)
	}
	if _, ok := entry["duration"]; !ok {
		t.Error("log entry has no duration")
	}
}

func TestLoggerDefaultsStatusToOK(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	writesBody := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	httpapi.Logger(logger)(writesBody).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if !strings.Contains(buf.String(), `"status":200`) {
		t.Errorf("got log %s, want status 200", buf.String())
	}
}

func TestLoggerIncludesTraceID(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	tp := sdktrace.NewTracerProvider()
	h := httpapi.Telemetry(tp, metricnoop.NewMeterProvider())(httpapi.Logger(logger)(okHandler))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/calculations", nil))

	want := `"traceId":"` + rec.Header().Get("X-Trace-ID") + `"`
	if !strings.Contains(buf.String(), want) {
		t.Errorf("got log %s, want %s", buf.String(), want)
	}
}
