package tracing

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Errors
var (
	ErrEndpointRequired  = errors.New("tracing endpoint required when enabled")
	ErrInvalidSampleRate = errors.New("sample rate must be between 0 and 1")
	ErrTracerNotInit     = errors.New("tracer not initialized")
)

// TracerProvider wraps the OpenTelemetry tracer provider.
type TracerProvider struct {
	provider *sdktrace.TracerProvider
	tracer   trace.Tracer
	cfg      Config
}

// globalProvider holds the global tracer provider instance.
var (
	globalProvider *TracerProvider
	globalMu       sync.RWMutex
)

// Init initializes the global tracer provider.
func Init(ctx context.Context, cfg Config) (*TracerProvider, error) {
	if !cfg.Enable {
		log.Debug().Msg("Tracing disabled")
		return nil, nil
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid tracing config: %w", err)
	}

	// Create OTLP exporter
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
		otlptracegrpc.WithTimeout(cfg.Timeout),
	}

	if cfg.Insecure {
		opts = append(opts, otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	if len(cfg.Headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(cfg.Headers))
	}

	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource with service info
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion("0.1.0"),
			attribute.String("environment", "production"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create sampler based on sample rate
	var sampler sdktrace.Sampler
	if cfg.SampleRate >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if cfg.SampleRate <= 0 {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(cfg.SampleRate)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, sdktrace.WithMaxExportBatchSize(cfg.BatchSize)),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global tracer provider and propagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	provider := &TracerProvider{
		provider: tp,
		tracer:   tp.Tracer(cfg.ServiceName),
		cfg:      cfg,
	}

	globalMu.Lock()
	globalProvider = provider
	globalMu.Unlock()

	log.Info().
		Str("endpoint", cfg.Endpoint).
		Str("service", cfg.ServiceName).
		Float64("sample_rate", cfg.SampleRate).
		Msg("Tracing initialized")

	return provider, nil
}

// Shutdown shuts down the tracer provider.
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	if tp == nil || tp.provider == nil {
		return nil
	}
	return tp.provider.Shutdown(ctx)
}

// Tracer returns the tracer.
func (tp *TracerProvider) Tracer() trace.Tracer {
	if tp == nil {
		return otel.Tracer("hybridgrid")
	}
	return tp.tracer
}

// GetTracer returns the global tracer.
func GetTracer() trace.Tracer {
	globalMu.RLock()
	p := globalProvider
	globalMu.RUnlock()
	if p == nil {
		return otel.Tracer("hybridgrid")
	}
	return p.tracer
}

// IsEnabled returns whether tracing is enabled.
func IsEnabled() bool {
	globalMu.RLock()
	p := globalProvider
	globalMu.RUnlock()
	return p != nil && p.cfg.Enable
}

// StartSpan starts a new span.
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return GetTracer().Start(ctx, name, opts...)
}

// SpanFromContext returns the current span from context.
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// AddEvent adds an event to the current span.
func AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SetAttributes sets attributes on the current span.
func SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// RecordError records an error on the current span.
func RecordError(ctx context.Context, err error, opts ...trace.EventOption) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err, opts...)
}

// Common attribute keys for hybridgrid.
var (
	AttrTaskID       = attribute.Key("hybridgrid.task_id")
	AttrCompiler     = attribute.Key("hybridgrid.compiler")
	AttrSourceFile   = attribute.Key("hybridgrid.source_file")
	AttrWorkerID     = attribute.Key("hybridgrid.worker_id")
	AttrTargetArch   = attribute.Key("hybridgrid.target_arch")
	AttrCacheHit     = attribute.Key("hybridgrid.cache_hit")
	AttrFallback     = attribute.Key("hybridgrid.fallback")
	AttrExitCode     = attribute.Key("hybridgrid.exit_code")
	AttrDurationMs   = attribute.Key("hybridgrid.duration_ms")
	AttrQueueTimeMs  = attribute.Key("hybridgrid.queue_time_ms")
	AttrCompileMs    = attribute.Key("hybridgrid.compile_time_ms")
	AttrSourceSize   = attribute.Key("hybridgrid.source_size")
	AttrObjectSize   = attribute.Key("hybridgrid.object_size")
)
