package server

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

// --- DefaultConfig ---

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, 50052, cfg.Port)
	assert.Equal(t, 4, cfg.MaxConcurrent)
	assert.Equal(t, 120*time.Second, cfg.DefaultTimeout)
}

// --- Server creation ---

func TestNew(t *testing.T) {
	cfg := Config{Port: 9000, MaxConcurrent: 8, DefaultTimeout: 60 * time.Second}
	s := New(cfg)

	require.NotNil(t, s)
	assert.Equal(t, 9000, s.config.Port)
	assert.Equal(t, 8, s.config.MaxConcurrent)
	assert.NotNil(t, s.capabilities)
	assert.NotNil(t, s.executor)
	assert.Equal(t, int32(8), s.capabilities.MaxParallelTasks)
}

func TestCapabilities(t *testing.T) {
	s := New(DefaultConfig())
	caps := s.Capabilities()

	require.NotNil(t, caps)
	assert.Equal(t, int32(4), caps.MaxParallelTasks)
}

func TestStop_NotStarted(t *testing.T) {
	s := New(Config{Port: 0, MaxConcurrent: 1})
	// Should not panic
	s.Stop()
}

// --- Handshake (worker does not handle) ---

func TestHandshake_Unimplemented(t *testing.T) {
	s := New(DefaultConfig())

	_, err := s.Handshake(context.Background(), &pb.HandshakeRequest{})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unimplemented, st.Code())
}

// --- Build (not implemented) ---

func TestBuild_Unimplemented(t *testing.T) {
	s := New(DefaultConfig())

	_, err := s.Build(context.Background(), &pb.BuildRequest{TaskId: "t1"})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unimplemented, st.Code())
}

// --- StreamBuild (not implemented) ---

func TestStreamBuild_Unimplemented(t *testing.T) {
	s := New(DefaultConfig())

	// StreamBuild takes a stream interface; calling with nil will cause a nil pointer.
	// We test the actual gRPC method returns Unimplemented via direct call.
	err := s.StreamBuild(nil)
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unimplemented, st.Code())
}

// --- GetWorkersForBuild (not applicable) ---

func TestGetWorkersForBuild_Unimplemented(t *testing.T) {
	s := New(DefaultConfig())

	_, err := s.GetWorkersForBuild(context.Background(), &pb.WorkersForBuildRequest{})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unimplemented, st.Code())
}

// --- Compile ---

func TestCompile_EmptyTaskId(t *testing.T) {
	s := New(DefaultConfig())

	_, err := s.Compile(context.Background(), &pb.CompileRequest{})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestCompile_ConcurrencyLimit(t *testing.T) {
	s := New(Config{Port: 0, MaxConcurrent: 1, DefaultTimeout: 5 * time.Second})

	// Manually set active tasks to the limit
	s.activeTasks = 1

	_, err := s.Compile(context.Background(), &pb.CompileRequest{
		TaskId:             "overflow-task",
		PreprocessedSource: []byte("int main() {}"),
		Compiler:           "gcc",
	})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.ResourceExhausted, st.Code())
	assert.Contains(t, st.Message(), "too many concurrent tasks")
}

func TestCompile_UsesCustomTimeout(t *testing.T) {
	s := New(Config{Port: 0, MaxConcurrent: 4, DefaultTimeout: 120 * time.Second})

	// We can't fully test timeout without a real executor, but we verify
	// the request flows through. The executor will fail since no compiler exists,
	// but we can at least verify the request is processed.
	resp, err := s.Compile(context.Background(), &pb.CompileRequest{
		TaskId:             "timeout-task",
		PreprocessedSource: []byte("int main() {}"),
		Compiler:           "gcc",
		TimeoutSeconds:     5,
	})
	// Should get a response (executor may fail but the flow completes)
	if err == nil {
		assert.NotNil(t, resp)
	}
}

func TestCompile_ExecutionFailure(t *testing.T) {
	s := New(Config{Port: 0, MaxConcurrent: 4, DefaultTimeout: 5 * time.Second})

	// Compile with invalid compiler - will attempt native execution
	resp, err := s.Compile(context.Background(), &pb.CompileRequest{
		TaskId:             "bad-compile",
		PreprocessedSource: []byte("not valid c++"),
		Compiler:           "nonexistent-compiler-xyz",
	})
	// The executor should return a result (not a gRPC error)
	if err == nil {
		assert.NotNil(t, resp)
		assert.Equal(t, pb.TaskStatus_STATUS_FAILED, resp.Status)
	}
}

func TestCompile_MetricsTracking(t *testing.T) {
	s := New(Config{Port: 0, MaxConcurrent: 4, DefaultTimeout: 5 * time.Second})

	// Initial state
	assert.Equal(t, int64(0), s.totalTasks)

	s.Compile(context.Background(), &pb.CompileRequest{
		TaskId:             "metrics-task",
		PreprocessedSource: []byte("int main() {}"),
		Compiler:           "gcc",
	})

	assert.Equal(t, int64(1), s.totalTasks)
	// Active tasks should be back to 0 after completion
	assert.Equal(t, int64(0), s.activeTasks)
}

// --- HealthCheck ---

func TestHealthCheck_Healthy(t *testing.T) {
	s := New(Config{Port: 0, MaxConcurrent: 4})

	resp, err := s.HealthCheck(context.Background(), &pb.HealthRequest{})
	require.NoError(t, err)
	assert.True(t, resp.Healthy)
	assert.Equal(t, int32(0), resp.ActiveTasks)
	assert.Equal(t, int32(0), resp.QueuedTasks)
}

func TestHealthCheck_Unhealthy_AtCapacity(t *testing.T) {
	s := New(Config{Port: 0, MaxConcurrent: 2})
	s.activeTasks = 2

	resp, err := s.HealthCheck(context.Background(), &pb.HealthRequest{})
	require.NoError(t, err)
	assert.False(t, resp.Healthy)
	assert.Equal(t, int32(2), resp.ActiveTasks)
}

func TestHealthCheck_ActiveTasksReported(t *testing.T) {
	s := New(Config{Port: 0, MaxConcurrent: 10})
	s.activeTasks = 5

	resp, err := s.HealthCheck(context.Background(), &pb.HealthRequest{})
	require.NoError(t, err)
	assert.True(t, resp.Healthy)
	assert.Equal(t, int32(5), resp.ActiveTasks)
}

// --- GetWorkerStatus ---

func TestGetWorkerStatus(t *testing.T) {
	s := New(Config{Port: 0, MaxConcurrent: 4})
	s.activeTasks = 2
	s.totalTasks = 10
	s.successTasks = 8
	s.totalTimeMs = 800

	resp, err := s.GetWorkerStatus(context.Background(), &pb.WorkerStatusRequest{})
	require.NoError(t, err)

	assert.Equal(t, int32(1), resp.TotalWorkers)
	assert.Equal(t, int32(1), resp.HealthyWorkers)
	require.Len(t, resp.Workers, 1)

	w := resp.Workers[0]
	assert.Equal(t, int32(2), w.ActiveTasks)
	assert.Equal(t, int64(10), w.TotalTasksCompleted)
	assert.Equal(t, float32(100), w.AvgLatencyMs) // 800ms / 8 success = 100ms
}

func TestGetWorkerStatus_ZeroSuccess(t *testing.T) {
	s := New(Config{Port: 0, MaxConcurrent: 4})
	s.totalTasks = 3
	s.successTasks = 0

	resp, err := s.GetWorkerStatus(context.Background(), &pb.WorkerStatusRequest{})
	require.NoError(t, err)
	require.Len(t, resp.Workers, 1)
	assert.Equal(t, float32(0), resp.Workers[0].AvgLatencyMs)
}

// --- Concurrent compile requests ---

func TestCompile_ConcurrentRequests(t *testing.T) {
	s := New(Config{Port: 0, MaxConcurrent: 10, DefaultTimeout: 5 * time.Second})

	var wg sync.WaitGroup
	results := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, err := s.Compile(context.Background(), &pb.CompileRequest{
				TaskId:             "concurrent-" + string(rune('0'+idx)),
				PreprocessedSource: []byte("int main() {}"),
				Compiler:           "gcc",
			})
			if err != nil {
				results <- err
			}
		}(i)
	}

	wg.Wait()
	close(results)

	// No ResourceExhausted errors since MaxConcurrent=10
	for err := range results {
		st, ok := status.FromError(err)
		if ok {
			assert.NotEqual(t, codes.ResourceExhausted, st.Code(),
				"Should not hit concurrency limit with MaxConcurrent=10")
		}
	}
}
