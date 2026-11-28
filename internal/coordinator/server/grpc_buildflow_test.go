package server

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

func TestBuildSucceeded(t *testing.T) {
	tests := []struct {
		name string
		resp *pb.BuildResponse
		err  error
		want bool
	}{
		{"completed", &pb.BuildResponse{Status: pb.TaskStatus_STATUS_COMPLETED}, nil, true},
		{"failed status", &pb.BuildResponse{Status: pb.TaskStatus_STATUS_FAILED}, nil, false},
		{"transport error", nil, errors.New("boom"), false},
		{"error with response", &pb.BuildResponse{Status: pb.TaskStatus_STATUS_COMPLETED}, errors.New("boom"), false},
		{"nil response", nil, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildSucceeded(tt.resp, tt.err); got != tt.want {
				t.Errorf("buildSucceeded() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildCompletionEvent(t *testing.T) {
	meta := buildTaskMeta{taskID: "t1", buildID: "b1", buildType: "flutter"}
	started := time.UnixMilli(1_000)
	completed := time.UnixMilli(3_500)

	tests := []struct {
		name         string
		resp         *pb.BuildResponse
		err          error
		wantStatus   string
		wantExit     int32
		wantErrMsg   string
		wantCompiled int64
	}{
		{
			name:         "success",
			resp:         &pb.BuildResponse{Status: pb.TaskStatus_STATUS_COMPLETED, ExitCode: 0, BuildTimeMs: 42},
			wantStatus:   "completed",
			wantCompiled: 42,
		},
		{
			name:         "worker reported failure",
			resp:         &pb.BuildResponse{Status: pb.TaskStatus_STATUS_FAILED, ExitCode: 7, Stderr: "bad", BuildTimeMs: 5},
			wantStatus:   "failed",
			wantExit:     7,
			wantErrMsg:   "bad",
			wantCompiled: 5,
		},
		{
			name:       "transport error",
			err:        errors.New("connection reset"),
			wantStatus: "failed",
			wantExit:   1,
			wantErrMsg: "connection reset",
		},
		{
			name:       "transport error wins over response",
			resp:       &pb.BuildResponse{Status: pb.TaskStatus_STATUS_FAILED, ExitCode: 7, Stderr: "bad"},
			err:        errors.New("connection reset"),
			wantStatus: "failed",
			wantExit:   1,
			wantErrMsg: "connection reset",
		},
		{
			name:       "nil response and no error",
			wantStatus: "failed",
			wantExit:   1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := buildCompletionEvent(meta, "w1", started, completed, 250*time.Millisecond, tt.resp, tt.err)
			if ev.Status != tt.wantStatus || ev.ExitCode != tt.wantExit || ev.ErrorMessage != tt.wantErrMsg {
				t.Errorf("got status=%q exit=%d err=%q, want status=%q exit=%d err=%q",
					ev.Status, ev.ExitCode, ev.ErrorMessage, tt.wantStatus, tt.wantExit, tt.wantErrMsg)
			}
			if ev.CompileTimeMs != tt.wantCompiled {
				t.Errorf("CompileTimeMs = %d, want %d", ev.CompileTimeMs, tt.wantCompiled)
			}
			if ev.ID != "t1" || ev.BuildID != "b1" || ev.BuildType != "flutter" || ev.WorkerID != "w1" {
				t.Errorf("identity fields wrong: %+v", ev)
			}
			if ev.StartedAtMs != 1_000 || ev.CompletedAtMs != 3_500 || ev.DurationMs != 2_500 || ev.QueueTimeMs != 250 {
				t.Errorf("timing fields wrong: %+v", ev)
			}
			if ev.FromCache {
				t.Error("FromCache = true, want false")
			}
		})
	}
}

func TestArchiveCacheEntryResponse(t *testing.T) {
	src := &pb.BuildResponse{
		Stdout:       "out",
		Stderr:       "err",
		Artifacts:    []byte{1, 2, 3},
		ArtifactList: []*pb.ArtifactInfo{{Name: "a", Path: "p", SizeBytes: 3, Checksum: "c"}},
		BuildTimeMs:  99,
	}
	entry := newArchiveCacheEntry(src)
	got := entry.response()

	if got.Status != pb.TaskStatus_STATUS_COMPLETED || got.ExitCode != 0 || !got.FromCache {
		t.Errorf("unexpected status fields: %+v", got)
	}
	if got.Stdout != "out" || got.Stderr != "err" || got.BuildTimeMs != 99 || len(got.Artifacts) != 3 || len(got.ArtifactList) != 1 {
		t.Errorf("unexpected content: %+v", got)
	}

	// Mutating the response must not reach the stored entry.
	got.Artifacts[0] = 9
	got.ArtifactList[0].Name = "mutated"
	again := entry.response()
	if again.Artifacts[0] != 1 || again.ArtifactList[0].Name != "a" {
		t.Errorf("cached entry was aliased by a returned response: %+v", again)
	}

	// An entry with no artifacts yields empty, not nil-panicking, copies.
	empty := newArchiveCacheEntry(&pb.BuildResponse{}).response()
	if len(empty.Artifacts) != 0 || empty.ArtifactList != nil {
		t.Errorf("empty entry response = %+v", empty)
	}
}

func TestArchiveForwardRequest(t *testing.T) {
	req := &pb.BuildRequest{
		TaskId:         "t",
		BuildId:        "b",
		SourceHash:     "h",
		SourceArchive:  []byte{1},
		BuildType:      pb.BuildType_BUILD_TYPE_FLUTTER,
		TargetPlatform: pb.TargetPlatform_PLATFORM_ANDROID,
		TimeoutSeconds: 30,
		CustomBuilder:  "must-not-forward",
	}
	got := archiveForwardRequest(req)
	if got.TaskId != "t" || got.BuildId != "b" || got.SourceHash != "h" || len(got.SourceArchive) != 1 ||
		got.BuildType != req.BuildType || got.TargetPlatform != req.TargetPlatform || got.TimeoutSeconds != 30 {
		t.Errorf("forwarded fields wrong: %+v", got)
	}
	if got.CustomBuilder != "" {
		t.Errorf("CustomBuilder forwarded: %q", got.CustomBuilder)
	}
}

func TestWorkerErrorResponse(t *testing.T) {
	got := workerErrorResponse(errors.New("boom"))
	if got.Status != pb.TaskStatus_STATUS_FAILED || got.ExitCode != 1 || got.Stderr != "worker error: boom" {
		t.Errorf("workerErrorResponse = %+v", got)
	}
}

func TestRecordCacheServed(t *testing.T) {
	tests := []struct {
		name     string
		notifier bool
	}{
		{"with notifier", true},
		{"without notifier", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{}
			n := &mockEventNotifier{}
			if tt.notifier {
				s.eventNotifier = n
			}
			var perType int64
			meta := buildTaskMeta{taskID: "t", buildID: "b", buildType: "unity"}

			s.recordCacheServed(meta, time.Now(), 123, &perType)

			if s.cacheHits != 1 || s.totalTasks != 1 || s.successTasks != 1 || perType != 1 {
				t.Errorf("counters: hits=%d total=%d success=%d perType=%d",
					s.cacheHits, s.totalTasks, s.successTasks, perType)
			}
			if !tt.notifier {
				return
			}
			if len(n.started) != 1 || len(n.completed) != 1 {
				t.Fatalf("events: started=%d completed=%d", len(n.started), len(n.completed))
			}
			c := n.completed[0]
			if c.Status != "completed" || !c.FromCache || c.DurationMs != 123 || c.CompileTimeMs != 123 ||
				c.WorkerID != "" || c.BuildID != "b" || c.BuildType != "unity" {
				t.Errorf("completed event: %+v", c)
			}
			if n.started[0].Status != "running" || n.started[0].WorkerID != "" {
				t.Errorf("started event: %+v", n.started[0])
			}
		})
	}
}

func TestNoWorkerResponse(t *testing.T) {
	s := &Server{}
	n := &mockEventNotifier{}
	s.eventNotifier = n
	meta := buildTaskMeta{taskID: "t", buildID: "b", buildType: "flutter"}

	s.countNoWorker()
	resp := s.noWorkerResponse(meta, time.Now(), errors.New("none"))

	if atomic.LoadInt64(&s.totalTasks) != 1 || atomic.LoadInt64(&s.failedTasks) != 1 {
		t.Errorf("counters: total=%d failed=%d", s.totalTasks, s.failedTasks)
	}
	if resp.Status != pb.TaskStatus_STATUS_FAILED || resp.ExitCode != 1 || resp.Stderr != "no worker available: none" {
		t.Errorf("response: %+v", resp)
	}
	if len(n.started) != 1 || len(n.completed) != 1 {
		t.Fatalf("events: started=%d completed=%d", len(n.started), len(n.completed))
	}
	c := n.completed[0]
	if c.Status != "failed" || c.ExitCode != 1 || c.ErrorMessage != "no worker available: none" || c.FromCache {
		t.Errorf("completed event: %+v", c)
	}

	// No notifier: must not panic.
	(&Server{}).noWorkerResponse(meta, time.Now(), errors.New("none"))
}

func TestBeginWorkerTask(t *testing.T) {
	s := &Server{}
	var builds int64

	release := s.beginWorkerTask("w1", &builds)
	if s.activeTasks != 1 || s.totalTasks != 1 || builds != 1 {
		t.Errorf("after begin: active=%d total=%d builds=%d", s.activeTasks, s.totalTasks, builds)
	}
	v, ok := s.activeTasksByWorker.Load("w1")
	if !ok || atomic.LoadInt64(v.(*int64)) != 1 {
		t.Fatalf("per-worker count not 1: ok=%v", ok)
	}

	release()
	if s.activeTasks != 0 || s.totalTasks != 1 || builds != 1 {
		t.Errorf("after release: active=%d total=%d builds=%d", s.activeTasks, s.totalTasks, builds)
	}
	if atomic.LoadInt64(v.(*int64)) != 0 {
		t.Errorf("per-worker count after release = %d", atomic.LoadInt64(v.(*int64)))
	}
}
