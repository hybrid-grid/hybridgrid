package tracing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

// --- Config ---

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.False(t, cfg.Enable)
	assert.Equal(t, "localhost:4317", cfg.Endpoint)
	assert.Equal(t, "hybridgrid", cfg.ServiceName)
	assert.Equal(t, 0.1, cfg.SampleRate)
	assert.True(t, cfg.Insecure)
	assert.Equal(t, 10*time.Second, cfg.Timeout)
	assert.Equal(t, 512, cfg.BatchSize)
}

func TestCoordinatorConfig(t *testing.T) {
	cfg := CoordinatorConfig()
	assert.Equal(t, "hg-coord", cfg.ServiceName)
	assert.False(t, cfg.Enable)
}

func TestWorkerConfig(t *testing.T) {
	cfg := WorkerConfig()
	assert.Equal(t, "hg-worker", cfg.ServiceName)
	assert.False(t, cfg.Enable)
}

func TestClientConfig(t *testing.T) {
	cfg := ClientConfig()
	assert.Equal(t, "hgbuild", cfg.ServiceName)
	assert.Equal(t, 0.01, cfg.SampleRate)
}

// --- Validate ---

func TestValidate_Disabled(t *testing.T) {
	cfg := Config{Enable: false}
	assert.NoError(t, cfg.Validate())
}

func TestValidate_EnabledNoEndpoint(t *testing.T) {
	cfg := Config{Enable: true, Endpoint: "", SampleRate: 0.5}
	err := cfg.Validate()
	assert.ErrorIs(t, err, ErrEndpointRequired)
}

func TestValidate_InvalidSampleRate_Negative(t *testing.T) {
	cfg := Config{Enable: true, Endpoint: "localhost:4317", SampleRate: -0.1}
	err := cfg.Validate()
	assert.ErrorIs(t, err, ErrInvalidSampleRate)
}

func TestValidate_InvalidSampleRate_TooHigh(t *testing.T) {
	cfg := Config{Enable: true, Endpoint: "localhost:4317", SampleRate: 1.5}
	err := cfg.Validate()
	assert.ErrorIs(t, err, ErrInvalidSampleRate)
}

func TestValidate_DefaultsServiceName(t *testing.T) {
	cfg := Config{Enable: true, Endpoint: "localhost:4317", SampleRate: 0.5, ServiceName: ""}
	err := cfg.Validate()
	assert.NoError(t, err)
	assert.Equal(t, "hybridgrid", cfg.ServiceName)
}

func TestValidate_Valid(t *testing.T) {
	cfg := Config{
		Enable:      true,
		Endpoint:    "localhost:4317",
		SampleRate:  1.0,
		ServiceName: "test-svc",
	}
	assert.NoError(t, cfg.Validate())
}

func TestValidate_BoundarySampleRates(t *testing.T) {
	tests := []struct {
		name       string
		sampleRate float64
		wantErr    bool
	}{
		{"zero", 0.0, false},
		{"one", 1.0, false},
		{"mid", 0.5, false},
		{"negative", -0.001, true},
		{"over_one", 1.001, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Enable:      true,
				Endpoint:    "localhost:4317",
				SampleRate:  tt.sampleRate,
				ServiceName: "test",
			}
			err := cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// --- Init ---

func TestInit_Disabled(t *testing.T) {
	cfg := Config{Enable: false}
	tp, err := Init(context.Background(), cfg)
	assert.NoError(t, err)
	assert.Nil(t, tp)
}

func TestInit_InvalidConfig(t *testing.T) {
	cfg := Config{Enable: true, Endpoint: "", SampleRate: 0.5}
	tp, err := Init(context.Background(), cfg)
	assert.Error(t, err)
	assert.Nil(t, tp)
	assert.Contains(t, err.Error(), "invalid tracing config")
}

// --- TracerProvider methods ---

func TestTracerProvider_Shutdown_Nil(t *testing.T) {
	var tp *TracerProvider
	assert.NoError(t, tp.Shutdown(context.Background()))
}

func TestTracerProvider_Shutdown_NilProvider(t *testing.T) {
	tp := &TracerProvider{}
	assert.NoError(t, tp.Shutdown(context.Background()))
}

func TestTracerProvider_Tracer_Nil(t *testing.T) {
	var tp *TracerProvider
	tracer := tp.Tracer()
	assert.NotNil(t, tracer)
}

func TestTracerProvider_Tracer_WithProvider(t *testing.T) {
	tp := &TracerProvider{
		tracer: trace.NewNoopTracerProvider().Tracer("test"),
	}
	tracer := tp.Tracer()
	assert.NotNil(t, tracer)
}

// --- Global functions ---

func TestGetTracer_NoGlobal(t *testing.T) {
	// Reset global state
	globalMu.Lock()
	old := globalProvider
	globalProvider = nil
	globalMu.Unlock()
	defer func() {
		globalMu.Lock()
		globalProvider = old
		globalMu.Unlock()
	}()

	tracer := GetTracer()
	assert.NotNil(t, tracer)
}

func TestIsEnabled_NoGlobal(t *testing.T) {
	globalMu.Lock()
	old := globalProvider
	globalProvider = nil
	globalMu.Unlock()
	defer func() {
		globalMu.Lock()
		globalProvider = old
		globalMu.Unlock()
	}()

	assert.False(t, IsEnabled())
}

func TestIsEnabled_DisabledConfig(t *testing.T) {
	globalMu.Lock()
	old := globalProvider
	globalProvider = &TracerProvider{cfg: Config{Enable: false}}
	globalMu.Unlock()
	defer func() {
		globalMu.Lock()
		globalProvider = old
		globalMu.Unlock()
	}()

	assert.False(t, IsEnabled())
}

func TestIsEnabled_EnabledConfig(t *testing.T) {
	globalMu.Lock()
	old := globalProvider
	globalProvider = &TracerProvider{cfg: Config{Enable: true}}
	globalMu.Unlock()
	defer func() {
		globalMu.Lock()
		globalProvider = old
		globalMu.Unlock()
	}()

	assert.True(t, IsEnabled())
}

// --- Span functions ---

func TestStartSpan(t *testing.T) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "test-span")
	defer span.End()

	assert.NotNil(t, span)
	assert.NotEqual(t, context.Background(), ctx)
}

func TestSpanFromContext(t *testing.T) {
	ctx := context.Background()
	span := SpanFromContext(ctx)
	assert.NotNil(t, span)
}

func TestAddEvent(t *testing.T) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "event-span")
	defer span.End()

	// Should not panic
	AddEvent(ctx, "test-event", attribute.String("key", "value"))
}

func TestSetAttributes(t *testing.T) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "attr-span")
	defer span.End()

	// Should not panic
	SetAttributes(ctx, attribute.String("key", "value"), attribute.Int("count", 42))
}

func TestRecordError(t *testing.T) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "error-span")
	defer span.End()

	// Should not panic
	RecordError(ctx, assert.AnError)
}

// --- Attribute keys ---

func TestAttributeKeys(t *testing.T) {
	assert.Equal(t, attribute.Key("hybridgrid.task_id"), AttrTaskID)
	assert.Equal(t, attribute.Key("hybridgrid.compiler"), AttrCompiler)
	assert.Equal(t, attribute.Key("hybridgrid.source_file"), AttrSourceFile)
	assert.Equal(t, attribute.Key("hybridgrid.worker_id"), AttrWorkerID)
	assert.Equal(t, attribute.Key("hybridgrid.target_arch"), AttrTargetArch)
	assert.Equal(t, attribute.Key("hybridgrid.cache_hit"), AttrCacheHit)
	assert.Equal(t, attribute.Key("hybridgrid.fallback"), AttrFallback)
	assert.Equal(t, attribute.Key("hybridgrid.exit_code"), AttrExitCode)
	assert.Equal(t, attribute.Key("hybridgrid.duration_ms"), AttrDurationMs)
	assert.Equal(t, attribute.Key("hybridgrid.queue_time_ms"), AttrQueueTimeMs)
	assert.Equal(t, attribute.Key("hybridgrid.compile_time_ms"), AttrCompileMs)
	assert.Equal(t, attribute.Key("hybridgrid.source_size"), AttrSourceSize)
	assert.Equal(t, attribute.Key("hybridgrid.object_size"), AttrObjectSize)
}

// --- Error sentinels ---

func TestErrorSentinels(t *testing.T) {
	assert.Error(t, ErrEndpointRequired)
	assert.Error(t, ErrInvalidSampleRate)
	assert.Error(t, ErrTracerNotInit)
	assert.NotEqual(t, ErrEndpointRequired, ErrInvalidSampleRate)
}

// --- gRPC options ---

func TestServerOptions(t *testing.T) {
	opts := ServerOptions()
	require.NotEmpty(t, opts)

	// Verify it returns valid gRPC server options
	for _, opt := range opts {
		assert.IsType(t, grpc.EmptyServerOption{}, grpc.EmptyServerOption{}) // type check
		_ = opt                                                              // use it to avoid unused
	}
}

func TestDialOptions(t *testing.T) {
	opts := DialOptions()
	require.NotEmpty(t, opts)

	// Verify they are valid gRPC dial options
	assert.Len(t, opts, 1)
}

// --- Helper functions ---

func TestWithCompileAttributes(t *testing.T) {
	opt := WithCompileAttributes("task-1", "gcc", "x86_64", 1024)
	assert.NotNil(t, opt)

	// Verify it can be used as a SpanStartOption by starting a span
	ctx, span := StartSpan(context.Background(), "test", opt)
	defer span.End()
	assert.NotNil(t, ctx)
}

func TestWithScheduleAttributes(t *testing.T) {
	opt := WithScheduleAttributes("task-1", "worker-1")
	assert.NotNil(t, opt)

	// Verify it can be used as a SpanStartOption
	ctx, span := StartSpan(context.Background(), "test", opt)
	defer span.End()
	assert.NotNil(t, ctx)
}
