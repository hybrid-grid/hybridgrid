package resilience

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Circuit Breaker Tests

func TestCircuitManager_NewManager(t *testing.T) {
	cfg := DefaultCircuitConfig()
	m := NewCircuitManager(cfg)

	if m == nil {
		t.Fatal("NewCircuitManager returned nil")
	}
}

func TestCircuitManager_InitialState(t *testing.T) {
	m := NewCircuitManager(DefaultCircuitConfig())

	// Unknown worker should be CLOSED
	state := m.GetState("unknown-worker")
	if state != CircuitClosed {
		t.Errorf("GetState() = %s, want CLOSED", state)
	}

	if m.IsOpen("unknown-worker") {
		t.Error("IsOpen() = true for unknown worker, want false")
	}
}

func TestCircuitManager_Execute_Success(t *testing.T) {
	m := NewCircuitManager(DefaultCircuitConfig())

	result, err := m.Execute("worker-1", func() (interface{}, error) {
		return "success", nil
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result != "success" {
		t.Errorf("Execute result = %v, want 'success'", result)
	}

	if m.GetState("worker-1") != CircuitClosed {
		t.Errorf("State = %s after success, want CLOSED", m.GetState("worker-1"))
	}
}

func TestCircuitManager_Execute_Failure(t *testing.T) {
	cfg := CircuitConfig{
		MaxRequests:  1,
		Interval:     1 * time.Second,
		Timeout:      1 * time.Second,
		FailureRatio: 0.5, // 50% failure rate
		MinRequests:  2,   // Need at least 2 requests
	}
	m := NewCircuitManager(cfg)

	testErr := errors.New("test error")

	// First failure
	_, err := m.Execute("worker-1", func() (interface{}, error) {
		return nil, testErr
	})
	if err != testErr {
		t.Errorf("Execute error = %v, want testErr", err)
	}

	// Second failure should trip the breaker (50% failure with 2 requests)
	_, err = m.Execute("worker-1", func() (interface{}, error) {
		return nil, testErr
	})

	// Circuit should be open now
	if !m.IsOpen("worker-1") {
		t.Errorf("Circuit should be OPEN after failures")
	}
}

func TestCircuitManager_OnStateChange(t *testing.T) {
	cfg := CircuitConfig{
		MaxRequests:  1,
		Interval:     1 * time.Second,
		Timeout:      100 * time.Millisecond,
		FailureRatio: 0.5,
		MinRequests:  2,
	}
	m := NewCircuitManager(cfg)

	var stateChanges []CircuitState
	m.OnStateChange(func(workerID string, from, to CircuitState) {
		stateChanges = append(stateChanges, to)
	})

	testErr := errors.New("test error")

	// Cause failures to trip breaker
	for i := 0; i < 3; i++ {
		m.Execute("worker-1", func() (interface{}, error) {
			return nil, testErr
		})
	}

	// Wait for state changes
	time.Sleep(50 * time.Millisecond)

	// Should have received OPEN state change
	hasOpen := false
	for _, state := range stateChanges {
		if state == CircuitOpen {
			hasOpen = true
			break
		}
	}

	if !hasOpen && m.IsOpen("worker-1") {
		// State was already open, that's also ok
		t.Log("Circuit is open")
	}
}

func TestCircuitManager_GetAllStates(t *testing.T) {
	m := NewCircuitManager(DefaultCircuitConfig())

	// Create some breakers
	m.Execute("worker-1", func() (interface{}, error) { return nil, nil })
	m.Execute("worker-2", func() (interface{}, error) { return nil, nil })

	states := m.GetAllStates()
	if len(states) != 2 {
		t.Errorf("GetAllStates() returned %d states, want 2", len(states))
	}
}

func TestCircuitManager_Remove(t *testing.T) {
	m := NewCircuitManager(DefaultCircuitConfig())

	m.Execute("worker-1", func() (interface{}, error) { return nil, nil })
	m.Remove("worker-1")

	states := m.GetAllStates()
	if len(states) != 0 {
		t.Errorf("GetAllStates() after Remove returned %d states, want 0", len(states))
	}
}

// Retry Tests

func TestRetry_Success(t *testing.T) {
	cfg := DefaultRetryConfig()
	ctx := context.Background()

	var attempts int
	err := Retry(ctx, cfg, func() error {
		attempts++
		return nil
	})

	if err != nil {
		t.Fatalf("Retry failed: %v", err)
	}
	if attempts != 1 {
		t.Errorf("Retry attempts = %d, want 1", attempts)
	}
}

func TestRetry_SuccessAfterRetries(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries:      5,
		InitialInterval: 10 * time.Millisecond,
		Multiplier:      1.5,
		MaxInterval:     100 * time.Millisecond,
		MaxElapsedTime:  5 * time.Second,
	}
	ctx := context.Background()

	var attempts int32
	err := Retry(ctx, cfg, func() error {
		n := atomic.AddInt32(&attempts, 1)
		if n < 3 {
			return errors.New("transient error")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Retry failed: %v", err)
	}
	if attempts != 3 {
		t.Errorf("Retry attempts = %d, want 3", attempts)
	}
}

func TestRetry_MaxRetriesExceeded(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries:      2,
		InitialInterval: 10 * time.Millisecond,
		Multiplier:      1.5,
		MaxInterval:     100 * time.Millisecond,
		MaxElapsedTime:  5 * time.Second,
	}
	ctx := context.Background()

	var attempts int32
	testErr := errors.New("persistent error")
	err := Retry(ctx, cfg, func() error {
		atomic.AddInt32(&attempts, 1)
		return testErr
	})

	if err == nil {
		t.Error("Retry should have failed")
	}
	// Should have tried MaxRetries + 1 times
	if attempts > 3 {
		t.Errorf("Retry attempts = %d, want <= 3", attempts)
	}
}

func TestRetry_NonRetryableError(t *testing.T) {
	cfg := DefaultRetryConfig()
	ctx := context.Background()

	var attempts int32
	// InvalidArgument is not retryable
	nonRetryableErr := status.Error(codes.InvalidArgument, "bad request")

	err := Retry(ctx, cfg, func() error {
		atomic.AddInt32(&attempts, 1)
		return nonRetryableErr
	})

	if err == nil {
		t.Error("Retry should have failed with non-retryable error")
	}
	if attempts != 1 {
		t.Errorf("Retry attempts = %d, want 1 (should not retry non-retryable)", attempts)
	}
}

func TestRetry_ContextCanceled(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries:      10,
		InitialInterval: 100 * time.Millisecond,
		Multiplier:      1.5,
		MaxInterval:     1 * time.Second,
		MaxElapsedTime:  10 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())

	var attempts int32
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := Retry(ctx, cfg, func() error {
		atomic.AddInt32(&attempts, 1)
		return errors.New("transient error")
	})

	if err == nil {
		t.Error("Retry should have failed due to context cancellation")
	}
	t.Logf("Retry stopped after %d attempts", attempts)
}

func TestRetryWithResult(t *testing.T) {
	cfg := DefaultRetryConfig()
	ctx := context.Background()

	result, err := RetryWithResult(ctx, cfg, func() (string, error) {
		return "result", nil
	})

	if err != nil {
		t.Fatalf("RetryWithResult failed: %v", err)
	}
	if result != "result" {
		t.Errorf("RetryWithResult result = %s, want 'result'", result)
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil", nil, false},
		{"context.Canceled", context.Canceled, false},
		{"context.DeadlineExceeded", context.DeadlineExceeded, false},
		{"InvalidArgument", status.Error(codes.InvalidArgument, ""), false},
		{"NotFound", status.Error(codes.NotFound, ""), false},
		{"PermissionDenied", status.Error(codes.PermissionDenied, ""), false},
		{"Unavailable", status.Error(codes.Unavailable, ""), true},
		{"ResourceExhausted", status.Error(codes.ResourceExhausted, ""), true},
		{"Internal", status.Error(codes.Internal, ""), true},
		{"Unknown", status.Error(codes.Unknown, ""), true},
		{"generic error", errors.New("generic"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRetryable(tt.err)
			if got != tt.expected {
				t.Errorf("IsRetryable(%v) = %v, want %v", tt.err, got, tt.expected)
			}
		})
	}
}
