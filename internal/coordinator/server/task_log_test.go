package server

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/scheduler"
)

// loggerWithBuffer returns a TaskLogger writing to an in-memory buffer
// so tests can assert exact output.
func loggerWithBuffer() (*TaskLogger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	return &TaskLogger{w: buf}, buf
}

func sampleRecord() *TaskLogRecord {
	return &TaskLogRecord{
		TS:                          time.Date(2026, 4, 29, 10, 0, 0, 0, time.UTC),
		Event:                       "task_completed",
		TaskID:                      "task-1",
		BuildType:                   "cpp",
		BuildID:                     "build-session-1",
		Scheduler:                   "p2c",
		WorkerID:                    "worker-3",
		WorkerArch:                  "x86_64",
		WorkerNativeArch:            "x86_64",
		WorkerCPUCores:              8,
		WorkerMemBytes:              8 * 1024 * 1024 * 1024,
		WorkerActiveTasksAtDispatch: 2,
		WorkerMaxParallel:           8,
		WorkerDiscoverySource:       "mdns",
		TargetArch:                  "ARCH_X86_64",
		ClientOS:                    "linux",
		SourceSizeBytes:             524288,
		PreprocessedSizeBytes:       524288,
		RawSourceSizeBytes:          0,
		QueueTimeMs:                 5,
		CompileTimeMs:               1234,
		WorkerRPCLatencyMs:          1245,
		TotalDurationMs:             1250,
		Success:                     true,
		ExitCode:                    0,
		FromCache:                   false,
	}
}

func TestTaskLogger_EmitsJSONLines(t *testing.T) {
	l, buf := loggerWithBuffer()
	l.Log(sampleRecord())
	l.Log(sampleRecord())

	out := buf.String()
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	assert.Len(t, lines, 2, "expected one JSON object per line")

	var got TaskLogRecord
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &got))
	assert.Equal(t, "task-1", got.TaskID)
	assert.Equal(t, "p2c", got.Scheduler)
	assert.Equal(t, int32(2), got.WorkerActiveTasksAtDispatch)
}

func TestTaskLogger_NilSafe(t *testing.T) {
	var l *TaskLogger // nil receiver
	assert.NotPanics(t, func() {
		l.Log(sampleRecord())
		_ = l.Close()
	})
}

func TestTaskLogger_NilRecordSkipped(t *testing.T) {
	l, buf := loggerWithBuffer()
	l.Log(nil)
	assert.Empty(t, buf.String())
}

func TestTaskLogger_FileBackend(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.jsonl")
	l, err := NewTaskLogger(path)
	require.NoError(t, err)

	l.Log(sampleRecord())
	require.NoError(t, l.Close())

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"task_id":"task-1"`)
}

func TestTaskLogger_StdoutFallback(t *testing.T) {
	l, err := NewTaskLogger("")
	require.NoError(t, err)
	assert.Equal(t, os.Stdout, l.w)

	l2, err := NewTaskLogger("stdout")
	require.NoError(t, err)
	assert.Equal(t, os.Stdout, l2.w)
}

// TestCompile_EmitsTaskLogRecord verifies the Compile() hot path writes a
// well-formed TaskLogRecord through the configured TaskLogger. The worker is
// unreachable so the compile fails, but the log record must still be emitted
// with success=false and the correct scheduler/worker fields populated.
func TestCompile_EmitsTaskLogRecord(t *testing.T) {
	s, _, cleanup := setupTestServer(t, Config{
		Port:           0,
		HeartbeatTTL:   60 * time.Second,
		RequestTimeout: 1 * time.Second,
		SchedulerType:  "p2c",
	})
	defer cleanup()

	// Replace the default stdout-backed logger with a buffer-backed one.
	logger, buf := loggerWithBuffer()
	s.taskLogger = logger

	require.NoError(t, s.registry.Add(&registry.WorkerInfo{
		ID:      "worker-X",
		Address: "127.0.0.1:19998",
		Capabilities: &pb.WorkerCapabilities{
			NativeArch: pb.Architecture_ARCH_X86_64,
			CpuCores:   8,
			CpuMillis:  500, // e.g. a Docker --cpus=0.5 container reporting a host core count of 8
			Cpp:        &pb.CppCapability{Compilers: []string{"gcc"}},
		},
		MaxParallel:     4,
		DiscoverySource: "mdns",
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := s.Compile(ctx, &pb.CompileRequest{
		TaskId:             "log-task-1",
		PreprocessedSource: []byte("int main() { return 0; }"),
		Compiler:           "gcc",
		TargetArch:         pb.Architecture_ARCH_X86_64,
		BuildId:            "build-session-1",
	})
	require.NoError(t, err)
	assert.Equal(t, pb.TaskStatus_STATUS_FAILED, resp.Status, "expect failure since worker is unreachable")

	out := strings.TrimRight(buf.String(), "\n")
	require.NotEmpty(t, out, "TaskLogger should have emitted at least one record")

	lines := strings.Split(out, "\n")
	require.Len(t, lines, 1)

	var rec TaskLogRecord
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &rec))
	assert.Equal(t, "task_completed", rec.Event)
	assert.Equal(t, "log-task-1", rec.TaskID)
	assert.Equal(t, "build-session-1", rec.BuildID)
	assert.Equal(t, "p2c", rec.Scheduler)
	assert.Equal(t, "worker-X", rec.WorkerID)
	assert.Equal(t, "mdns", rec.WorkerDiscoverySource)
	assert.Equal(t, int32(8), rec.WorkerCPUCores)
	// Regression check for the cgroup-quota blindness (docs/thesis/
	// bao-cao-tien-do-260809.md §1.5): WorkerCPUCores
	// alone can't distinguish workers under a cgroup CPU quota (it stays
	// the host core count regardless of the quota), so WorkerCPUMillis
	// must independently reach the log record — proving the cgroup-aware
	// value flows from registry.WorkerInfo through to offline analysis,
	// not just internally to the scheduler.
	assert.Equal(t, int32(500), rec.WorkerCPUMillis)
	assert.False(t, rec.Success)
	// PreprocessedSource length is the only source data sent.
	assert.Equal(t, len("int main() { return 0; }"), rec.PreprocessedSizeBytes)
	assert.Equal(t, len("int main() { return 0; }"), rec.SourceSizeBytes)

	invalidResp, err := s.Compile(ctx, &pb.CompileRequest{
		TaskId:             "log-task-invalid-build-id",
		PreprocessedSource: []byte("int main() { return 0; }"),
		Compiler:           "gcc",
		TargetArch:         pb.Architecture_ARCH_X86_64,
		BuildId:            "invalid/build-id",
	})
	require.NoError(t, err)
	assert.Equal(t, pb.TaskStatus_STATUS_FAILED, invalidResp.Status, "expect failure since worker is unreachable")

	lines = strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	require.Len(t, lines, 2)
	var invalidRec TaskLogRecord
	require.NoError(t, json.Unmarshal([]byte(lines[1]), &invalidRec))
	assert.Empty(t, invalidRec.BuildID)
}

// TestTaskLogger_ConcurrentWrites verifies that JSON Lines stay well-formed
// when multiple goroutines log simultaneously. Each line must parse cleanly;
// any interleaving of partial writes would surface as a JSON unmarshal error.
func TestTaskLogger_ConcurrentWrites(t *testing.T) {
	l, buf := loggerWithBuffer()

	const writers = 16
	const perWriter = 100
	var wg sync.WaitGroup
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perWriter; j++ {
				l.Log(sampleRecord())
			}
		}()
	}
	wg.Wait()

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	assert.Len(t, lines, writers*perWriter)
	for _, line := range lines {
		var rec TaskLogRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("malformed JSON line under concurrency: %q (err=%v)", line, err)
		}
	}
}

func TestTaskLogRecord_SchedulerParamsJSON(t *testing.T) {
	rec := sampleRecord()
	oldJSON, err := json.Marshal(rec)
	require.NoError(t, err)
	assert.NotContains(t, string(oldJSON), "scheduler_params")
	assert.Contains(t, string(oldJSON), `"scheduler":"p2c"`)
	assert.Contains(t, string(oldJSON), `"worker_id":"worker-3"`)

	rec.SchedulerParams = "alpha=0.5 warm_start=100 load_penalty=0.5 discount=0.95 discount_mode=global"
	withParams, err := json.Marshal(rec)
	require.NoError(t, err)
	var fields map[string]any
	require.NoError(t, json.Unmarshal(withParams, &fields))
	assert.Equal(t, rec.SchedulerParams, fields["scheduler_params"])

	rec.SchedulerParams = ""
	again, err := json.Marshal(rec)
	require.NoError(t, err)
	assert.Equal(t, string(oldJSON), string(again), "empty optional field must preserve existing records")
}

func TestTaskLogRecord_LinUCBDiagnosticsJSON(t *testing.T) {
	rec := sampleRecord()
	without, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("marshal record without diagnostics: %v", err)
	}
	const legacy = `{"ts":"2026-04-29T10:00:00Z","event":"task_completed","dispatch_seq":0,"dispatch_ts":"0001-01-01T00:00:00Z","received_ts":"0001-01-01T00:00:00Z","worker_address":"","candidates":null,"task_id":"task-1","build_type":"cpp","build_id":"build-session-1","scheduler":"p2c","worker_id":"worker-3","worker_arch":"x86_64","worker_native_arch":"x86_64","worker_cpu_cores":8,"worker_cpu_millis":0,"worker_mem_bytes":8589934592,"worker_active_tasks_at_dispatch":2,"worker_max_parallel":8,"worker_discovery_source":"mdns","target_arch":"ARCH_X86_64","client_os":"linux","source_size_bytes":524288,"preprocessed_size_bytes":524288,"raw_source_size_bytes":0,"queue_time_ms":5,"compile_time_ms":1234,"worker_rpc_latency_ms":1245,"total_duration_ms":1250,"success":true,"exit_code":0,"from_cache":false,"q_value_at_dispatch":0,"was_exploration":false}`
	if string(without) != legacy {
		t.Fatalf("diagnostics off changed legacy JSON:\n got: %s\nwant: %s", without, legacy)
	}

	rec.LinUCBCandidates = []scheduler.CandidateDiag{{Address: "worker-3", Mean: -0.2, Bonus: 0.3, LoadPenalty: 0.1, P: 0, XAinvX: 0.25, ArmCount: 4, X: []float64{1, 0.5}}}
	rec.RewardOffset = -0.15
	with, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("marshal record with diagnostics: %v", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(with, &fields); err != nil {
		t.Fatalf("decode diagnostic record: %v", err)
	}
	if string(fields["reward_offset"]) != `-0.15` {
		t.Errorf("reward_offset = %s, want -0.15", fields["reward_offset"])
	}
	var candidates []map[string]any
	if err := json.Unmarshal(fields["linucb_candidates"], &candidates); err != nil {
		t.Fatalf("decode linucb_candidates: %v", err)
	}
	if len(candidates) != 1 {
		t.Fatalf("linucb_candidates length = %d, want 1", len(candidates))
	}
	for _, key := range []string{"address", "mean", "bonus", "load_penalty", "p", "x_ainv_x", "arm_count", "x"} {
		if _, ok := candidates[0][key]; !ok {
			t.Errorf("linucb_candidates[0] missing %q", key)
		}
	}
	if candidates[0]["address"] != "worker-3" {
		t.Errorf("address = %v, want worker-3", candidates[0]["address"])
	}
	if got, ok := candidates[0]["x"].([]any); !ok || len(got) != 2 || got[0] != float64(1) || got[1] != 0.5 {
		t.Errorf("x = %v, want [1 0.5]", candidates[0]["x"])
	}
}

func TestCompile_DiscountedSchedulerParamsInTaskLog(t *testing.T) {
	s, _, cleanup := setupTestServer(t, Config{
		Port:              0,
		HeartbeatTTL:      60 * time.Second,
		RequestTimeout:    1 * time.Second,
		SchedulerType:     "hybrid-linucb-d",
		DiscountValue:     0.95,
		DiscountModeValue: "global",
	})
	defer cleanup()

	logger, buf := loggerWithBuffer()
	s.taskLogger = logger
	require.NoError(t, s.registry.Add(&registry.WorkerInfo{
		ID:      "worker-params",
		Address: "127.0.0.1:19998",
		Capabilities: &pb.WorkerCapabilities{
			NativeArch: pb.Architecture_ARCH_X86_64,
			Cpp:        &pb.CppCapability{Compilers: []string{"gcc"}},
		},
		MaxParallel: 1,
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := s.Compile(ctx, &pb.CompileRequest{
		TaskId:             "params-task",
		PreprocessedSource: []byte("int main() { return 0; }"),
		Compiler:           "gcc",
		TargetArch:         pb.Architecture_ARCH_X86_64,
	})
	require.NoError(t, err)

	var rec TaskLogRecord
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &rec))
	assert.Equal(t, "hybrid-linucb-d", rec.Scheduler)
	assert.Equal(t, "alpha=0.5 warm_start=0 load_penalty=0 discount=0.95 discount_mode=global", rec.SchedulerParams)
}
