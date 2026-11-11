package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"time"
)

// Middleware wraps a handler with extra behaviour.
type Middleware func(http.Handler) http.Handler

// SecurityHeaders sets response headers suited to a JSON-only API.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set(headerContentTypeOptions, "nosniff")
		h.Set(headerContentSecurityPolicy, "default-src 'none'; frame-ancestors 'none'")
		h.Set(headerCacheControl, "no-store")
		h.Set(headerReferrerPolicy, "no-referrer")
		next.ServeHTTP(w, r)
	})
}

// Recover turns a panic in next into a logged 500 response.
func Recover(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if v := recover(); v != nil {
					if err, ok := v.(error); ok && errors.Is(err, http.ErrAbortHandler) {
						panic(v)
					}
					logger.Error("panic", "value", v, "method", r.Method, "path", r.URL.Path, "traceId", traceID(r))
					writeError(w, http.StatusInternalServerError, CodeInternal, msgInternal)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// Logger logs one line per request with its method, path, status and duration.
func Logger(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"duration", time.Since(start).String(),
				"traceId", traceID(r),
			)
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}
