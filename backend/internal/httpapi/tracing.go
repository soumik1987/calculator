package httpapi

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

// Telemetry starts a server span per request, continuing a W3C traceparent sent by the
// caller, records HTTP metrics, and returns the trace ID in the X-Trace-ID header.
// Health checks are not traced.
func Telemetry(tp trace.TracerProvider, mp metric.MeterProvider) Middleware {
	return func(next http.Handler) http.Handler {
		withTraceID := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if sc := trace.SpanContextFromContext(r.Context()); sc.IsValid() {
				w.Header().Set(headerTraceID, sc.TraceID().String())
			}
			next.ServeHTTP(w, r)
		})
		return otelhttp.NewHandler(withTraceID, spanOperation,
			otelhttp.WithTracerProvider(tp),
			otelhttp.WithMeterProvider(mp),
			otelhttp.WithPropagators(propagation.TraceContext{}),
			otelhttp.WithFilter(func(r *http.Request) bool { return r.URL.Path != pathHealth }),
			otelhttp.WithSpanNameFormatter(spanName),
		)
	}
}

// spanName uses the route for known paths and only the method otherwise, so random
// URLs can't create unbounded span names.
func spanName(_ string, r *http.Request) string {
	if r.URL.Path == pathCalculations {
		return r.Method + " " + pathCalculations
	}
	return r.Method
}

func traceID(r *http.Request) string {
	if sc := trace.SpanContextFromContext(r.Context()); sc.IsValid() {
		return sc.TraceID().String()
	}
	return ""
}

func recordErrorType(r *http.Request, code ErrorCode) {
	trace.SpanFromContext(r.Context()).SetAttributes(semconv.ErrorTypeKey.String(string(code)))
}
