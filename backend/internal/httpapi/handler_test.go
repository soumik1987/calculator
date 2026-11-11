package httpapi_test

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/soumik1987/sami-calculator/backend/internal/calculator"
	"github.com/soumik1987/sami-calculator/backend/internal/httpapi"
)

func failing(err error) httpapi.Evaluator {
	return func(string) (float64, error) { return 0, err }
}

func newServer(evaluate httpapi.Evaluator) http.Handler {
	return httpapi.NewHandler(evaluate, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func do(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestCalculateSuccess(t *testing.T) {
	rec := do(t, newServer(calculator.Evaluate), http.MethodPost, "/api/v1/calculations", `{"expression":"(986x(4+8))+2"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("got content type %q", ct)
	}
	want := `{"expression":"(986x(4+8))+2","result":11834}`
	if got := strings.TrimSpace(rec.Body.String()); got != want {
		t.Errorf("got body %s, want %s", got, want)
	}
}

func TestCalculateErrors(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		evalErr     error
		wantStatus  int
		wantCode    httpapi.ErrorCode
		wantMessage string
	}{
		{"empty body", ``, nil, http.StatusBadRequest, httpapi.CodeInvalidRequest, "Request body is empty."},
		{"truncated json", `{"expression":`, nil, http.StatusBadRequest, httpapi.CodeInvalidRequest, "Request body is not valid JSON."},
		{"invalid json", `{expression: 1}`, nil, http.StatusBadRequest, httpapi.CodeInvalidRequest, "Request body is not valid JSON."},
		{"unknown field", `{"expression":"1","extra":1}`, nil, http.StatusBadRequest, httpapi.CodeInvalidRequest, `Unknown field "extra".`},
		{"wrong type", `{"expression":12}`, nil, http.StatusBadRequest, httpapi.CodeInvalidRequest, `Field "expression" must be a string.`},
		{"not an object", `["1+1"]`, nil, http.StatusBadRequest, httpapi.CodeInvalidRequest, "Request body must be a JSON object."},
		{"trailing data", `{"expression":"1"}{}`, nil, http.StatusBadRequest, httpapi.CodeInvalidRequest, "Request body must contain a single JSON object."},
		{"body too large", `{"expression":"` + strings.Repeat("1", 5000) + `"}`, nil, http.StatusRequestEntityTooLarge, httpapi.CodeRequestTooLarge, "Request body must be at most 4096 bytes."},
		{"syntax error", `{"expression":"2+3)"}`, nil, http.StatusBadRequest, httpapi.CodeInvalidExpression, `Unexpected ")" at position 4.`},
		{"empty expression", `{"expression":""}`, nil, http.StatusBadRequest, httpapi.CodeInvalidExpression, "Enter an expression."},
		{"division by zero", `{"expression":"1/(2-2)"}`, nil, http.StatusUnprocessableEntity, httpapi.CodeDivisionByZero, "Cannot divide by zero."},
		{"zero to a negative power", `{"expression":"0^-1"}`, nil, http.StatusUnprocessableEntity, httpapi.CodeDivisionByZero, ""},
		{"negative square root", `{"expression":"√(-4)"}`, nil, http.StatusUnprocessableEntity, httpapi.CodeNegativeSquareRoot, ""},
		{"logarithm of zero", `{"expression":"log(0)"}`, nil, http.StatusUnprocessableEntity, httpapi.CodeNonPositiveLogarithm, "Logarithm needs a number greater than zero."},
		{"undefined result", `{"expression":"(-8)^0.5"}`, nil, http.StatusUnprocessableEntity, httpapi.CodeUndefinedResult, ""},
		{"out of range", `{"expression":"10^1000"}`, nil, http.StatusUnprocessableEntity, httpapi.CodeOutOfRange, ""},
		{"unexpected failure", `{"expression":"1+1"}`, errors.New("boom"), http.StatusInternalServerError, httpapi.CodeInternal, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluate := httpapi.Evaluator(calculator.Evaluate)
			if tt.evalErr != nil {
				evaluate = failing(tt.evalErr)
			}
			rec := do(t, newServer(evaluate), http.MethodPost, "/api/v1/calculations", tt.body)

			if rec.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rec.Code, tt.wantStatus)
			}
			var body struct {
				Error struct {
					Code    httpapi.ErrorCode `json:"code"`
					Message string            `json:"message"`
				} `json:"error"`
			}
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("decode error body: %v", err)
			}
			if body.Error.Code != tt.wantCode {
				t.Errorf("got code %q, want %q", body.Error.Code, tt.wantCode)
			}
			if body.Error.Message == "" {
				t.Error("error message is empty")
			}
			if tt.wantMessage != "" && body.Error.Message != tt.wantMessage {
				t.Errorf("got message %q, want %q", body.Error.Message, tt.wantMessage)
			}
			if strings.Contains(body.Error.Message, "boom") {
				t.Error("internal error details leaked to client")
			}
		})
	}
}

func TestCalculateRequiresJSONContentType(t *testing.T) {
	for _, contentType := range []string{"", "text/plain", "application/x-www-form-urlencoded"} {
		t.Run(contentType, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/calculations", strings.NewReader(`{"expression":"1+1"}`))
			if contentType != "" {
				req.Header.Set("Content-Type", contentType)
			}
			rec := httptest.NewRecorder()

			newServer(calculator.Evaluate).ServeHTTP(rec, req)

			if rec.Code != http.StatusUnsupportedMediaType {
				t.Errorf("got status %d, want 415", rec.Code)
			}
		})
	}
}

func TestCalculateAcceptsJSONWithCharset(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculations", strings.NewReader(`{"expression":"1+1"}`))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	rec := httptest.NewRecorder()

	newServer(calculator.Evaluate).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got status %d, want 200", rec.Code)
	}
}

func TestHealth(t *testing.T) {
	rec := do(t, newServer(calculator.Evaluate), http.MethodGet, "/healthz", "")

	if rec.Code != http.StatusOK {
		t.Errorf("got status %d, want 200", rec.Code)
	}
}

func TestRemovedRoutes(t *testing.T) {
	for _, path := range []string{"/api/v1/operations", "/api/v1/evaluations"} {
		t.Run(path, func(t *testing.T) {
			rec := do(t, newServer(calculator.Evaluate), http.MethodGet, path, "")

			if rec.Code != http.StatusNotFound {
				t.Errorf("got status %d, want 404", rec.Code)
			}
		})
	}
}

func TestMethodNotAllowed(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			rec := do(t, newServer(calculator.Evaluate), method, "/api/v1/calculations", "")

			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("got status %d, want 405", rec.Code)
			}
		})
	}
}
