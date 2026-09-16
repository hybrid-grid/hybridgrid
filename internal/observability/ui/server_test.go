package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/telemetry"
)

const bufSize = 1024 * 1024

// fakeProvider is the only test StatsProvider implementation needed:
// the contract pins GetConsole's two failure paths via type
// assertions, so a bare variant (lacking GetConsole) reproduces the
// 503 absence case and a console variant reproduces the task 404 case.
type fakeProvider struct {
	stats   *telemetry.Stats
	workers []*telemetry.WorkerInfo
}

func (p *fakeProvider) GetStats() *telemetry.Stats {
	if p.stats == nil {
		return &telemetry.Stats{Timestamp: time.Now().Unix()}
	}
	return p.stats
}

func (p *fakeProvider) GetWorkers() []*telemetry.WorkerInfo { return p.workers }

// consoleProvider extends fakeProvider with retained stdout/stderr.
type consoleProvider struct {
	fakeProvider
	consoles map[string][3]any // task_id -> {stdout, stderr, truncated}
}

func (m *consoleProvider) GetConsole(taskID string) (stdout, stderr string, truncated bool, ok bool) {
	entry, found := m.consoles[taskID]
	if !found {
		return "", "", false, false
	}
	return entry[0].(string), entry[1].(string), entry[2].(bool), true
}

// setupServer binds a telemetry.Service over a bufconn gRPC server,
// dials it, and returns a ui Server wired against the gRPC client.
// The cleanup func tears down the connection and gRPC server; tests
// run the hub and HTTP mux in-line via httptest on s.server.Handler.
//
// The provider is shared between the telemetry store wiring and the
// test's expectations so byte-identical envelope assertions can be
// derived from the provider directly.
func setupServer(t *testing.T, provider telemetry.StatsProvider, store *telemetry.Store) (*Server, *grpc.ClientConn, *telemetry.Service, func()) {
	t.Helper()
	if store == nil {
		store = telemetry.NewStore()
	}
	teleSvc := telemetry.NewService(store, provider)

	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	pb.RegisterTelemetryServiceServer(srv, teleSvc)

	go func() {
		_ = srv.Serve(lis)
	}()

	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial bufconn: %v", err)
	}
	client := pb.NewTelemetryServiceClient(conn)

	uiServer := New(DefaultConfig(), client)
	go uiServer.hub.Run()
	uiServer.streamCtx, uiServer.streamCancel = context.WithCancel(context.Background())
	go uiServer.pumpStream(uiServer.streamCtx)

	cleanup := func() {
		if uiServer.streamCancel != nil {
			uiServer.streamCancel()
		}
		_ = conn.Close()
		srv.Stop()
	}
	return uiServer, conn, teleSvc, cleanup
}

// Do a slightly smarter unwrap on a Stats shape so tests can compare
// key fields the contract cares about without pinning timestamps.
func decodeJSON(t *testing.T, body *bytes.Buffer) map[string]interface{} {
	t.Helper()
	var out map[string]interface{}
	if err := json.Unmarshal(body.Bytes(), &out); err != nil {
		t.Fatalf("decode body: %v (raw=%q)", err, body.String())
	}
	return out
}

// doGET dispatches a GET through the bare handler in-process so
// assertions can read the response body directly without a TCP
// socket. Use this only for routes without path parameters; {id}
// routes need doGETRouted (the router populates path values only
// there).
func doGET(handler http.HandlerFunc, target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

// doGETRouted drives a GET through the full mux so path parameters
// ({id}) are populated by ServeMux's matcher. Required for /builds/{id}
// and /tasks/{id}/console.
func doGETRouted(s *Server, target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	s.routes().ServeHTTP(rec, req)
	return rec
}

// TestServer_StatsEnvelope pins the byte-identical contract for /api/v1/stats:
// the response body must equal the provider's Stats JSON (no wrapper).
func TestServer_StatsEnvelope(t *testing.T) {
	provider := &fakeProvider{stats: &telemetry.Stats{
		TotalTasks:     7,
		SuccessTasks:   5,
		CacheHits:      3,
		CacheHitRate:   0.5,
		TotalWorkers:   2,
		HealthyWorkers: 1,
		Timestamp:      123,
	}}
	s, _, _, cleanup := setupServer(t, provider, telemetry.NewStore())
	defer cleanup()

	rec := doGET(s.handleStats, "/api/v1/stats")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	want, _ := json.Marshal(provider.stats)
	// The only legal envelope is the Stats object itself. json.Marshal
	// emitted a trailing newline in both lines, so trim for clarity.
	if got := bytes.TrimSpace(rec.Body.Bytes()); !bytes.Equal(got, bytes.TrimSpace(want)) {
		t.Errorf("stats envelope drifted:\n got=%s\nwant=%s", got, bytes.TrimSpace(want))
	}
}

// TestServer_WorkersEnvelope guards both the workers wrapper shape and
// the count/timestamp keys the SPA depends on.
func TestServer_WorkersEnvelope(t *testing.T) {
	provider := &fakeProvider{workers: []*telemetry.WorkerInfo{
		{ID: "w1", Host: "h1", Healthy: true, CPUCores: 8},
		{ID: "w2", Host: "h2", Healthy: false, CPUCores: 4},
	}}
	s, _, _, cleanup := setupServer(t, provider, telemetry.NewStore())
	defer cleanup()

	rec := doGET(s.handleWorkers, "/api/v1/workers")
	body := decodeJSON(t, rec.Body)

	if count, _ := body["count"].(float64); count != 2 {
		t.Errorf("count = %v, want 2", body["count"])
	}
	if _, ok := body["timestamp"]; !ok {
		t.Errorf("missing timestamp key; body=%s", rec.Body.String())
	}
	workers, ok := body["workers"].([]interface{})
	if !ok || len(workers) != 2 {
		t.Fatalf("workers = %v, want 2-element array", body["workers"])
	}
	first := workers[0].(map[string]interface{})
	if first["id"] != "w1" {
		t.Errorf("first worker id = %v, want w1", first["id"])
	}
}

// TestServer_BuildsEnvelope hits ListBuilds: with one recorded build
// it must surface the same envelope keys (builds/count/timestamp) and
// pull the build_type from the recorded task.
func TestServer_BuildsEnvelope(t *testing.T) {
	store := telemetry.NewStore()
	store.RecordTask(&telemetry.TaskInfo{ID: "t1", BuildID: "b1", BuildType: "cpp", Status: "completed"})
	s, _, _, cleanup := setupServer(t, &fakeProvider{}, store)
	defer cleanup()

	rec := doGET(s.handleBuilds, "/api/v1/builds")
	body := decodeJSON(t, rec.Body)
	if count, _ := body["count"].(float64); count != 1 {
		t.Errorf("count = %v, want 1", body["count"])
	}
	if _, ok := body["timestamp"]; !ok {
		t.Errorf("missing timestamp key")
	}
	builds := body["builds"].([]interface{})
	b1 := builds[0].(map[string]interface{})
	if b1["id"] != "b1" {
		t.Errorf("build id = %v, want b1", b1["id"])
	}
	if b1["build_type"] != "cpp" {
		t.Errorf("build_type = %v, want cpp", b1["build_type"])
	}
}

// TestServer_TasksEnvelope covers Tasks newest-first ordering, the
// limit contract, and the count/timestamp keys.
func TestServer_TasksEnvelope(t *testing.T) {
	store := telemetry.NewStore()
	store.RecordTask(&telemetry.TaskInfo{ID: "t1", BuildID: "b1", Status: "completed"})
	store.RecordTask(&telemetry.TaskInfo{ID: "t2", BuildID: "b1", Status: "completed"})
	store.RecordTask(&telemetry.TaskInfo{ID: "t3", BuildID: "b1", Status: "completed"})

	s, _, _, cleanup := setupServer(t, &fakeProvider{}, store)
	defer cleanup()

	// Full window: latest first.
	rec := doGET(s.handleTasks, "/api/v1/tasks")
	body := decodeJSON(t, rec.Body)
	if count, _ := body["count"].(float64); count != 3 {
		t.Fatalf("count = %v, want 3", body["count"])
	}
	tasks := body["tasks"].([]interface{})
	if first := tasks[0].(map[string]interface{})["id"]; first != "t3" {
		t.Errorf("newest task id = %v, want t3", first)
	}

	// limit=N bounds to the newest N entries.
	rec = doGET(s.handleTasks, "/api/v1/tasks?limit=1")
	body = decodeJSON(t, rec.Body)
	if count, _ := body["count"].(float64); count != 1 {
		t.Fatalf("count = %v, want 1", body["count"])
	}
	tasks = body["tasks"].([]interface{})
	if first := tasks[0].(map[string]interface{})["id"]; first != "t3" {
		t.Errorf("bounded newest id = %v, want t3", first)
	}

	// limit=0 / unparseable keeps the full set.
	rec = doGET(s.handleTasks, "/api/v1/tasks?limit=0")
	body = decodeJSON(t, rec.Body)
	if count, _ := body["count"].(float64); count != 3 {
		t.Errorf("limit=0 count = %v, want 3 (no truncation)", body["count"])
	}
}

// TestServer_GetBuildNotFoundAndOk pins the byte-identical 404 body:
// exactly {"error":"build not found"}.
func TestServer_GetBuildNotFoundAndOk(t *testing.T) {
	store := telemetry.NewStore()
	store.RecordTask(&telemetry.TaskInfo{ID: "t1", BuildID: "b1", Status: "completed"})
	s, _, _, cleanup := setupServer(t, &fakeProvider{}, store)
	defer cleanup()

	// Unknown build: 404 with contract body.
	rec := doGETRouted(s, "/api/v1/builds/ghost")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unknown build status = %d, want 404", rec.Code)
	}
	want404 := `{"error":"build not found"}` + "\n"
	if got := rec.Body.String(); got != want404 {
		t.Errorf("404 body = %q, want %q", got, want404)
	}

	// Known build: 200 with {build, tasks}.
	rec = doGETRouted(s, "/api/v1/builds/b1")
	if rec.Code != http.StatusOK {
		t.Fatalf("known build status = %d, want 200", rec.Code)
	}
	body := decodeJSON(t, rec.Body)
	if _, ok := body["build"]; !ok {
		t.Errorf("missing build key")
	}
	tasks, _ := body["tasks"].([]interface{})
	if len(tasks) != 1 {
		t.Errorf("build tasks len = %d, want 1", len(tasks))
	}
}

// TestServer_Console503WhenLackingProvider exercises the 503 path: a
// bare StatsProvider (no GetConsole) on the server side surfaces as
// gRPC Unavailable, which the ui Server must translate to HTTP 503.
func TestServer_Console503WhenLackingProvider(t *testing.T) {
	provider := &fakeProvider{} // bare; does NOT implement ConsoleProvider
	s, _, _, cleanup := setupServer(t, provider, telemetry.NewStore())
	defer cleanup()

	rec := doGETRouted(s, "/api/v1/tasks/whatever/console")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("console w/o provider status = %d, want 503", rec.Code)
	}
}

// TestServer_Console404UnknownTask keeps the 404 path on a provider
// that does have console retention but has no entry for this task.
func TestServer_Console404UnknownTask(t *testing.T) {
	provider := &consoleProvider{
		fakeProvider: fakeProvider{},
		consoles:     map[string][3]any{},
	}
	s, _, _, cleanup := setupServer(t, provider, telemetry.NewStore())
	defer cleanup()

	rec := doGETRouted(s, "/api/v1/tasks/ghost/console")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unknown-task console status = %d, want 404", rec.Code)
	}
}

// TestServer_Console200 mirrors the success path: console retained,
// full JSON envelope with task_id/stdout/stderr/truncated.
func TestServer_Console200(t *testing.T) {
	provider := &consoleProvider{
		fakeProvider: fakeProvider{},
		consoles: map[string][3]any{
			"t1": {"hello-out", "hello-err", true},
		},
	}
	s, _, _, cleanup := setupServer(t, provider, telemetry.NewStore())
	defer cleanup()

	rec := doGETRouted(s, "/api/v1/tasks/t1/console")
	if rec.Code != http.StatusOK {
		t.Fatalf("console status = %d, want 200", rec.Code)
	}
	body := decodeJSON(t, rec.Body)
	if body["task_id"] != "t1" {
		t.Errorf("task_id = %v, want t1", body["task_id"])
	}
	if body["stdout"] != "hello-out" {
		t.Errorf("stdout = %v", body["stdout"])
	}
	if body["truncated"] != true {
		t.Errorf("truncated = %v, want true", body["truncated"])
	}
}

// TestServer_Health verifies the liveness probe answers 200 OK\n.
func TestServer_Health(t *testing.T) {
	s, _, _, cleanup := setupServer(t, &fakeProvider{}, telemetry.NewStore())
	defer cleanup()

	rec := doGET(func(w http.ResponseWriter, r *http.Request) {
		s.routes().ServeHTTP(w, r)
	}, "/health")

	if rec.Code != http.StatusOK {
		t.Fatalf("health status = %d, want 200", rec.Code)
	}
	if body := rec.Body.String(); body != "OK\n" {
		t.Errorf("health body = %q, want \"OK\\n\"", body)
	}
}

// TestServer_WebSocketReceivesStatsEvent drives the whole pipe: the
// telemetry service's RunStatsLoop broadcasts a stats event every 2s
// on a production coordinator; in this test we exercise the same path
// (minus the cadence) by racing a gRPC StreamEvents push through the
// ui server's pump, then reading the translated frame off the WS.
//
// The frame must be {"type":"stats","timestamp":T,"data":{...stats...}}.
func TestServer_WebSocketReceivesStatsEvent(t *testing.T) {
	provider := &fakeProvider{stats: &telemetry.Stats{TotalTasks: 42, Timestamp: 99}}
	teleSvcStore := telemetry.NewStore()
	s, _, teleSvc, cleanup := setupServer(t, provider, teleSvcStore)
	defer cleanup()

	ts := httptest.NewServer(s.routes())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial ws: %v", err)
	}
	defer ws.Close()

	// Wait for the hub to register the client (best-effort; the
	// register channel is processed synchronously in Run()).
	waitFor(t, time.Second, func() bool {
		return s.hub.ClientCount() == 1
	})

	// RunStatsLoop pulls the provider snapshot and pushes it to every
	// stream subscriber; here we cancel as soon as we observe one
	// translated frame so the test does not wait on the 2s cadence.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		teleSvc.RunStatsLoop(ctx)
		close(done)
	}()

	// Read frames until a translated stats frame arrives. The pump may
	// also deliver any replay frames first (empty here), then the live
	// push.
	ws.SetReadDeadline(time.Now().Add(6 * time.Second))
	var statsFrame map[string]interface{}
	for i := 0; i < 8 && statsFrame == nil; i++ {
		_, payload, err := ws.ReadMessage()
		if err != nil {
			cancel()
			<-done
			t.Fatalf("read ws message: %v", err)
		}
		for _, line := range bytes.Split(payload, []byte("\n")) {
			line = bytes.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			var msg map[string]interface{}
			if err := json.Unmarshal(line, &msg); err != nil {
				cancel()
				<-done
				t.Fatalf("decode ws frame: %v (line=%q)", err, line)
			}
			if msg["type"] == "stats" {
				statsFrame = msg
				break
			}
		}
	}
	cancel()
	<-done

	if statsFrame == nil {
		t.Fatal("never received a stats frame")
	}

	data, ok := statsFrame["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("stats frame data missing: %v", statsFrame)
	}
	if data["total_tasks"].(float64) != 42 {
		t.Errorf("translated stats total_tasks = %v, want 42", data["total_tasks"])
	}
	if _, ok := statsFrame["timestamp"]; !ok {
		t.Errorf("stats frame missing timestamp: %v", statsFrame)
	}
}

// TestServer_WebSocketReplaysTaskEventsOnConnect verifies the
// replay-on-connect path: push a task event before the client
// connects, then the new client must receive it immediately through
// StreamEvents' replay phase.
func TestServer_WebSocketReplaysTaskEventsOnConnect(t *testing.T) {
	store := telemetry.NewStore()
	// Record a task before connecting so the StreamEvents replay set
	// is non-empty.
	store.RecordTask(&telemetry.TaskInfo{
		ID:      "t-pre",
		BuildID: "bpre",
		Status:  "completed",
	})
	s, _, _, cleanup := setupServer(t, &fakeProvider{}, store)
	defer cleanup()

	ts := httptest.NewServer(s.routes())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial ws: %v", err)
	}
	defer ws.Close()

	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	// Expect the translated replay frame for the recorded task.
	var sawTask bool
	for i := 0; i < 4; i++ {
		_, payload, err := ws.ReadMessage()
		if err != nil {
			break
		}
		for _, line := range bytes.Split(payload, []byte("\n")) {
			line = bytes.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			var msg map[string]interface{}
			if json.Unmarshal(line, &msg) != nil {
				continue
			}
			if msg["type"] == "task_completed" {
				sawTask = true
			}
		}
	}
	if !sawTask {
		t.Error("never received a task replay frame on connect")
	}
}

// waitFor polls cond until it returns true or the deadline elapses.
func waitFor(t *testing.T, d time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("condition never met within %v", d)
}

// TestGeneratedClientSatisfiesInterface is a compile-time assertion:
// the generated TelemetryServiceClient must widen to the ui.Client
// interface. Without it the production wiring (cmd/hg-dashboard) is
// a runtime cast time bomb if the proto regen drifts.
func TestGeneratedClientSatisfiesInterface(t *testing.T) {
	var _ Client = pb.NewTelemetryServiceClient(nil)
}
