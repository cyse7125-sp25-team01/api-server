package main

import (
	"context"
	"net/http" // Added this
	"os"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp" // Added this
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// InitTracer sets up the OpenTelemetry tracer
func InitTracer() (func(context.Context) error, error) {
	ctx := context.Background()

	// Get collector endpoint from environment variable
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "opentelemetry-collector.monitoring.svc.cluster.local:4317"
	}

	// Configure the exporter
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint),
	}

	// Use insecure connection if specified
	if os.Getenv("OTEL_INSECURE") == "true" {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	// Create the exporter
	exporter, err := otlptrace.New(ctx, otlptracegrpc.NewClient(opts...))
	if err != nil {
		return nil, err
	}

	// Get service info from environment
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "api-server"
	}

	serviceVersion := os.Getenv("SERVICE_VERSION")
	if serviceVersion == "" {
		serviceVersion = "0.0.1"
	}

	// Create a resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
		),
	)
	if err != nil {
		return nil, err
	}

	// Set up the trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)

	// Set up the propagator for distributed tracing
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Return a function to shut down the exporter
	return tp.Shutdown, nil
}

// InstrumentedHTTPClient returns an HTTP client instrumented with OpenTelemetry
func InstrumentedHTTPClient() *http.Client {
	return &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
}