package server

import (
	"context"
	"errors"
	"math"
	"sync/atomic"
	"testing"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

func TestCompileSuccess(t *testing.T) {
	tests := []struct {
		name string
		resp *pb.CompileResponse
		err  error
		want bool
	}{
		{name: "completed", resp: &pb.CompileResponse{Status: pb.TaskStatus_STATUS_COMPLETED}, want: true},
		{name: "failed status", resp: &pb.CompileResponse{Status: pb.TaskStatus_STATUS_FAILED}},
		{name: "error with completed response", resp: &pb.CompileResponse{Status: pb.TaskStatus_STATUS_COMPLETED}, err: errors.New("boom")},
		{name: "nil response", resp: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compileSuccess(tt.resp, tt.err); got != tt.want {
				t.Errorf("compileSuccess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompileReward(t *testing.T) {
	denom := math.Log1p(rewardNormalizationMs)
	tests := []struct {
		name    string
		success bool
		resp    *pb.CompileResponse
		timeout time.Duration
		want    float64
	}{
		{
			name:    "success scores negative log latency",
			success: true,
			resp:    &pb.CompileResponse{CompilationTimeMs: 500},
			timeout: 120 * time.Second,
			want:    -math.Log1p(500) / denom,
		},
		{
			name:    "failure scores the failure reward",
			resp:    &pb.CompileResponse{CompilationTimeMs: 500},
			timeout: 120 * time.Second,
			want:    -1,
		},
		{
			name:    "success without response falls back to failure reward",
			success: true,
			timeout: 600 * time.Second,
			want:    failureReward(600 * time.Second),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compileReward(tt.success, tt.resp, tt.timeout); got != tt.want {
				t.Errorf("compileReward() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueueForCapacity(t *testing.T) {
	newRun := func() *compileRun {
		return &compileRun{req: &pb.CompileRequest{TaskId: "t1"}}
	}

	t.Run("queue disabled", func(t *testing.T) {
		s := &Server{}
		st := newRun()
		if got := s.queueForCapacity(context.Background(), st, "enter"); got != queueDisabled {
			t.Errorf("got %v, want queueDisabled", got)
		}
		if st.queued {
			t.Error("run must not be marked queued when the queue is disabled")
		}
	})

	t.Run("first call enters the queue without waiting", func(t *testing.T) {
		s := &Server{config: Config{DispatchQueueTimeout: time.Minute}}
		st := newRun()
		if got := s.queueForCapacity(context.Background(), st, "enter"); got != queueRetry {
			t.Errorf("got %v, want queueRetry", got)
		}
		if !st.queued || st.queueDeadline.IsZero() {
			t.Errorf("run not marked queued: %+v", st)
		}
		if n := atomic.LoadInt64(&s.queuedCompiles); n != 1 {
			t.Errorf("queuedCompiles = %d, want 1", n)
		}
		s.leaveDispatchQueue()
	})

	t.Run("full queue is treated as disabled", func(t *testing.T) {
		s := &Server{config: Config{DispatchQueueTimeout: time.Minute}}
		atomic.StoreInt64(&s.queuedCompiles, maxQueuedCompiles)
		st := newRun()
		if got := s.queueForCapacity(context.Background(), st, "enter"); got != queueDisabled {
			t.Errorf("got %v, want queueDisabled", got)
		}
		if n := atomic.LoadInt64(&s.queuedCompiles); n != maxQueuedCompiles {
			t.Errorf("queuedCompiles = %d, want unchanged %d", n, maxQueuedCompiles)
		}
	})

	t.Run("queued run past its deadline times out", func(t *testing.T) {
		s := &Server{config: Config{DispatchQueueTimeout: time.Minute}}
		st := newRun()
		st.queued = true
		st.queueDeadline = time.Now().Add(-time.Second)
		if got := s.queueForCapacity(context.Background(), st, "enter"); got != queueTimedOut {
			t.Errorf("got %v, want queueTimedOut", got)
		}
	})
}

func TestCompletedEvent(t *testing.T) {
	started := time.UnixMilli(1_000_000)
	completed := started.Add(250 * time.Millisecond)
	base := func(resp *pb.CompileResponse, err error) *compileRun {
		return &compileRun{
			req:           &pb.CompileRequest{TaskId: "t1"},
			buildID:       "b1",
			worker:        &registry.WorkerInfo{ID: "w1"},
			taskStartTime: started,
			queueTime:     40 * time.Millisecond,
			resp:          resp,
			err:           err,
		}
	}

	tests := []struct {
		name    string
		st      *compileRun
		success bool
		status  string
		exit    int32
		errMsg  string
	}{
		{
			name:    "success keeps worker exit code",
			st:      base(&pb.CompileResponse{ExitCode: 0, CompilationTimeMs: 90}, nil),
			success: true,
			status:  "completed",
		},
		{
			name:   "forward error defaults exit code 1",
			st:     base(nil, errors.New("rpc down")),
			status: "failed",
			exit:   1,
			errMsg: "rpc down",
		},
		{
			name:   "compile failure reports worker exit code and stderr",
			st:     base(&pb.CompileResponse{ExitCode: 2, Stderr: "syntax error"}, nil),
			status: "failed",
			exit:   2,
			errMsg: "syntax error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := completedEvent(tt.st, tt.success, completed)
			if ev.Status != tt.status || ev.ExitCode != tt.exit || ev.ErrorMessage != tt.errMsg {
				t.Errorf("event = {Status:%q ExitCode:%d ErrorMessage:%q}, want {%q %d %q}",
					ev.Status, ev.ExitCode, ev.ErrorMessage, tt.status, tt.exit, tt.errMsg)
			}
			if ev.ID != "t1" || ev.BuildType != "cpp" || ev.BuildID != "b1" || ev.WorkerID != "w1" {
				t.Errorf("identity fields wrong: %+v", ev)
			}
			if ev.StartedAtMs != started.UnixMilli() || ev.CompletedAtMs != completed.UnixMilli() ||
				ev.DurationMs != 250 || ev.QueueTimeMs != 40 {
				t.Errorf("timing fields wrong: %+v", ev)
			}
		})
	}
}
