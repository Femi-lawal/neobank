package tracing

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Config holds the tracing configuration
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string // e.g., "localhost:4317"
	Enabled        bool
}

// DefaultConfig returns a default tracing configuration
func DefaultConfig(serviceName string) Config {
	endpoint := "localhost:4317"
	if envEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); envEndpoint != "" {
		endpoint = envEndpoint
	}

	return Config{
		ServiceName:    serviceName,
		ServiceVersion: "1.0.0",
		Environment:    "development",
		OTLPEndpoint:   endpoint,
		Enabled:        true,
	}
}

// TracerProvider is a wrapper around the OpenTelemetry tracer provider
type TracerProvider struct {
	provider *sdktrace.TracerProvider
	tracer   trace.Tracer
	config   Config
}

// InitTracing initializes OpenTelemetry tracing
func InitTracing(ctx context.Context, cfg Config) (*TracerProvider, error) {
	if !cfg.Enabled {
		slog.Info("Tracing disabled")
		return &TracerProvider{config: cfg}, nil
	}

	// Create OTLP exporter
	conn, err := grpc.NewClient(cfg.OTLPEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		slog.Warn("Failed to connect to OTLP collector", "error", err)
		return &TracerProvider{config: cfg}, nil
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		slog.Warn("Failed to create OTLP exporter", "error", err)
		return &TracerProvider{config: cfg}, nil
	}

	// Create resource with service info
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			"", // Use empty schema URL to allow merge with Default() which has newer schema
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			attribute.String("environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create tracer provider
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global tracer provider and propagator
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer := provider.Tracer(cfg.ServiceName)

	slog.Info("Tracing initialized",
		"service", cfg.ServiceName,
		"endpoint", cfg.OTLPEndpoint,
	)

	return &TracerProvider{
		provider: provider,
		tracer:   tracer,
		config:   cfg,
	}, nil
}

// Shutdown gracefully shuts down the tracer provider
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	if tp.provider == nil {
		return nil
	}
	return tp.provider.Shutdown(ctx)
}

// Tracer returns the tracer for creating spans
func (tp *TracerProvider) Tracer() trace.Tracer {
	if tp.tracer == nil {
		return otel.Tracer(tp.config.ServiceName)
	}
	return tp.tracer
}

// SpanFromContext returns the current span from context
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// StartSpan starts a new span with the given name
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer("").Start(ctx, name, opts...)
}

// StartSpanWithTracer starts a new span using the provided tracer
func (tp *TracerProvider) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tp.Tracer().Start(ctx, name, opts...)
}

// AddEvent adds an event to the current span
func AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SetAttributes sets attributes on the current span
func SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// RecordError records an error on the current span
func RecordError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
}

// Common attribute keys for banking operations
var (
	AttrUserID    = attribute.Key("user.id")
	AttrAccountID = attribute.Key("account.id")
	AttrAmount    = attribute.Key("transaction.amount")
	AttrCurrency  = attribute.Key("transaction.currency")
	AttrPaymentID = attribute.Key("payment.id")
	AttrCardID    = attribute.Key("card.id")
	AttrProductID = attribute.Key("product.id")
	AttrOperation = attribute.Key("operation.type")
	AttrStatus    = attribute.Key("operation.status")
	AttrErrorCode = attribute.Key("error.code")
)
