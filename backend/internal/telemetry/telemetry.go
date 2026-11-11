// Package telemetry sets up OpenTelemetry tracing and metrics.
package telemetry

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const defaultServiceName = "calculator-api"

// Providers holds the tracer and meter providers to pass to instrumented code.
type Providers struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *sdkmetric.MeterProvider
}

// Setup creates the providers. Spans always get IDs, so logs and responses can carry a
// trace ID. Data leaves the process over OTLP/HTTP only when export is true; the exporters
// then read the standard OTEL_EXPORTER_OTLP_* variables, and OTEL_SERVICE_NAME overrides
// the service name.
func Setup(ctx context.Context, export bool) (*Providers, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(attribute.String("service.name", defaultServiceName)),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
	)
	if err != nil {
		return nil, err
	}

	traceOpts := []sdktrace.TracerProviderOption{sdktrace.WithResource(res)}
	metricOpts := []sdkmetric.Option{sdkmetric.WithResource(res)}
	if export {
		traceExporter, err := otlptracehttp.New(ctx)
		if err != nil {
			return nil, err
		}
		metricExporter, err := otlpmetrichttp.New(ctx)
		if err != nil {
			return nil, err
		}
		traceOpts = append(traceOpts, sdktrace.WithBatcher(traceExporter))
		metricOpts = append(metricOpts, sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)))
	}

	return &Providers{
		TracerProvider: sdktrace.NewTracerProvider(traceOpts...),
		MeterProvider:  sdkmetric.NewMeterProvider(metricOpts...),
	}, nil
}

// Shutdown flushes pending telemetry and stops the providers.
func (p *Providers) Shutdown(ctx context.Context) error {
	return errors.Join(p.TracerProvider.Shutdown(ctx), p.MeterProvider.Shutdown(ctx))
}
