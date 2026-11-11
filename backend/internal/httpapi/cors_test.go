package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/soumik1987/sami-calculator/backend/internal/httpapi"
)

var okHandler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func corsRequest(method, origin string) *http.Request {
	req := httptest.NewRequest(method, "/api/v1/calculations", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	return req
}

func TestCORSAllowedOrigin(t *testing.T) {
	h := httpapi.CORS([]string{"http://localhost:8080"})(okHandler)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, corsRequest(http.MethodGet, "http://localhost:8080"))

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:8080" {
		t.Errorf("got allow-origin %q", got)
	}
	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Errorf("got vary %q, want Origin", got)
	}
}

func TestCORSDisallowedOrigin(t *testing.T) {
	h := httpapi.CORS([]string{"http://localhost:8080"})(okHandler)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, corsRequest(http.MethodGet, "https://evil.example"))

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("got allow-origin %q, want none", got)
	}
}

func TestCORSWildcard(t *testing.T) {
	h := httpapi.CORS([]string{"*"})(okHandler)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, corsRequest(http.MethodGet, "https://any.example"))

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("got allow-origin %q, want *", got)
	}
}

func TestCORSPreflight(t *testing.T) {
	h := httpapi.CORS([]string{"http://localhost:8080"})(okHandler)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, corsRequest(http.MethodOptions, "http://localhost:8080"))

	if rec.Code != http.StatusNoContent {
		t.Errorf("got status %d, want 204", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("missing allow-methods header")
	}
	if rec.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Error("missing allow-headers header")
	}
}
