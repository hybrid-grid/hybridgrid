package telemetry

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

type fakeStatsProvider struct {
	stats   *Stats
	workers []*WorkerInfo
	console map[string][2]string
}

func (p *fakeStatsProvider) GetStats() *Stats {
	if p.stats == nil {
		return &Stats{Timestamp: time.Now().Unix()}
	}
	return p.stats
}

func (p *fakeStatsProvider) GetWorkers() []*WorkerInfo { return p.workers }

func (p *fakeStatsProvider) GetConsole(taskID string) (stdout, stderr string, truncated bool, ok bool) {
	out, found := p.console[taskID]
	if !found {
		return "", "", false, false
	}
	return out[0], out[1], false, true
}

// bareProvider deliberately lacks GetConsole so the Unavailable
// branch of GetConsole is reachable.
type bareProvider struct{}

func (bareProvider) GetStats() *Stats          { return &Stats{} }
func (bareProvider) GetWorkers() []*WorkerInfo { return nil }

// fakeStream collects sent events on a channel so the test can
// observe replay and live pushes without races.
type fakeStream struct {
	ctx  context.Context
	sent chan *pb.TelemetryEvent
}

func (f *fakeStream) Send(ev *pb.TelemetryEvent) error {
	f.sent <- ev
	return nil
}

func (f *fakeStream) Context() context.Context { return f.ctx }

func (f *fakeStream) RecvMsg(any) error            { return nil }
func (f *fakeStream) SendMsg(any) error            { return nil }
func (f *fakeStream) SetHeader(metadata.MD) error  { return nil }
func (f *fakeStream) SendHeader(metadata.MD) error { return nil }
func (f *fakeStream) SetTrailer(metadata.MD)       {}

func TestStore_BuildAggregation(t *testing.T) {
	s := NewStore()
	s.RecordTask(&TaskInfo{ID: "t1", BuildID: "b1", BuildType: "cpp", Status: "running", StartedAtMs: 1000})
	s.RecordTask(&TaskInfo{ID: "t2", BuildID: "b1", BuildType: "cpp", Status: "running", StartedAtMs: 2000})
	s.RecordTask(&TaskInfo{ID: "t1", BuildID: "b1", BuildType: "cpp", Status: "success", StartedAtMs: 1000, CompletedAtMs: 3000, FromCache: true})

	builds := s.Builds()
	if len(builds) != 1 {
		t.Fatalf("expected 1 build, got %d", len(builds))
	}
	b := builds[0]
	if b.ID != "b1" || b.TotalTasks != 2 || b.CompletedTasks != 1 || b.RunningTasks != 1 {
		t.Fatalf("unexpected aggregate: %+v", b)
	}
	if b.Status != "running" {
		t.Fatalf("build with a running task must report running, got %q", b.Status)
	}
	if b.FromCacheCount != 1 {
		t.Fatalf("expected 1 cache hit, got %d", b.FromCacheCount)
	}
	if b.FirstTaskAtMs != 1000 || b.LastTaskAtMs != 3000 {
		t.Fatalf("unexpected time bounds: %d..%d", b.FirstTaskAtMs, b.LastTaskAtMs)
	}

	// Completing the last task flips status to completed.
	s.RecordTask(&TaskInfo{ID: "t2", BuildID: "b1", BuildType: "cpp", Status: "failed", StartedAtMs: 2000, CompletedAtMs: 4000})
	if got := s.Builds()[0].Status; got != "failed" {
		t.Fatalf("build with a failed task must report failed, got %q", got)
	}
}

func TestStore_TaskEvictionMarksBuildTruncated(t *testing.T) {
	s := NewStore()
	s.maxTasksTotal = 2
	for i, id := range []string{"a", "b", "c"} {
		s.RecordTask(&TaskInfo{ID: id, BuildID: "b1", Status: "success", StartedAtMs: int64(i + 1)})
	}
	tasks := s.Tasks()
	if len(tasks) != 2 {
		t.Fatalf("expected retention cap of 2 tasks, got %d", len(tasks))
	}
	if tasks[0].ID != "c" || tasks[1].ID != "b" {
		t.Fatalf("expected newest-first [c b], got [%s %s]", tasks[0].ID, tasks[1].ID)
	}
	builds := s.Builds()
	if len(builds) != 1 || !builds[0].Truncated {
		t.Fatalf("evicted build must be marked truncated: %+v", builds)
	}
}

func TestStore_BuildDetailUnknownBuild(t *testing.T) {
	s := NewStore()
	if _, _, ok := s.BuildDetail("nope"); ok {
		t.Fatal("unknown build must report not-found")
	}
}

func TestService_AuthGatesEveryRPC(t *testing.T) {
	svc := NewService(NewStore(), &fakeStatsProvider{})
	svc.Token = "secret"

	if _, err := svc.GetStats(context.Background(), &pb.GetTelemetryStatsRequest{AuthToken: "wrong"}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("GetStats with bad token: got %v, want Unauthenticated", err)
	}
	if _, err := svc.ListWorkers(context.Background(), &pb.ListTelemetryWorkersRequest{}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("ListWorkers without token: got %v, want Unauthenticated", err)
	}
	resp, err := svc.GetStats(context.Background(), &pb.GetTelemetryStatsRequest{AuthToken: "secret"})
	if err != nil || resp.Stats == nil {
		t.Fatalf("GetStats with valid token failed: %v", err)
	}
}

func TestService_ListTasksLimitAndOrder(t *testing.T) {
	store := NewStore()
	svc := NewService(store, &fakeStatsProvider{})
	for i := 1; i <= 5; i++ {
		store.RecordTask(&TaskInfo{ID: string(rune('a' + i)), Status: "success", StartedAtMs: int64(i)})
	}
	resp, err := svc.ListTasks(context.Background(), &pb.ListTelemetryTasksRequest{Limit: 2})
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if len(resp.Tasks) != 2 || resp.Tasks[0].Id != "f" || resp.Tasks[1].Id != "e" {
		t.Fatalf("limit must bound to newest-first slice, got %+v", resp.Tasks)
	}
}

func TestService_GetBuildNotFound(t *testing.T) {
	svc := NewService(NewStore(), &fakeStatsProvider{})
	_, err := svc.GetBuild(context.Background(), &pb.GetTelemetryBuildRequest{BuildId: "ghost"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("got %v, want NotFound", err)
	}
}

func TestService_GetConsoleUnavailableAndFound(t *testing.T) {
	// Provider without console support reports Unavailable, not empty.
	svc := NewService(NewStore(), bareProvider{})
	if _, err := svc.GetConsole(context.Background(), &pb.GetTelemetryConsoleRequest{TaskId: "t"}); status.Code(err) != codes.Unavailable {
		t.Fatalf("got %v, want Unavailable", err)
	}

	withConsole := &fakeStatsProvider{console: map[string][2]string{"t": {"out", "err"}}}
	svc2 := NewService(NewStore(), withConsole)
	resp, err := svc2.GetConsole(context.Background(), &pb.GetTelemetryConsoleRequest{TaskId: "t"})
	if err != nil {
		t.Fatalf("GetConsole: %v", err)
	}
	if resp.Console.Stdout != "out" || resp.Console.Stderr != "err" {
		t.Fatalf("unexpected console payload: %+v", resp.Console)
	}
	if _, err := svc2.GetConsole(context.Background(), &pb.GetTelemetryConsoleRequest{TaskId: "ghost"}); status.Code(err) != codes.NotFound {
		t.Fatalf("unknown task: got %v, want NotFound", err)
	}
}
func TestService_StreamEventsReplaysThenPushes(t *testing.T) {
	store := NewStore()
	svc := NewService(store, &fakeStatsProvider{})
	store.RecordTask(&TaskInfo{ID: "old", Status: "success", StartedAtMs: 1})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream := &fakeStream{ctx: ctx, sent: make(chan *pb.TelemetryEvent, 16)}
	done := make(chan error, 1)
	go func() { done <- svc.StreamEvents(&pb.StreamTelemetryEventsRequest{}, stream) }()

	// Replay of the retained ring arrives first.
	ev := <-stream.sent
	if ev.Type != "task_completed" || ev.Timestamp == 0 {
		t.Fatalf("replayed event wrong: %+v", ev)
	}

	// Live push from the notifier path.
	onStart, _ := svc.CreateEventNotifier()
	onStart("live", "b1", "cpp", "running", "w1", 42)
	ev = <-stream.sent
	if ev.Type != "task_started" {
		t.Fatalf("live event type = %q, want task_started", ev.Type)
	}
	var task TaskInfo
	if err := json.Unmarshal([]byte(ev.PayloadJson), &task); err != nil || task.ID != "live" || task.StartedAtMs != 42 {
		t.Fatalf("live payload wrong: %q (%v)", ev.PayloadJson, err)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatalf("stream ended with %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("stream did not exit on client cancel")
	}
}
