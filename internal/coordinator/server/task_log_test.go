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
	assert.Equal(t, "p2c", rec.Scheduler)
	assert.Equal(t, "worker-X", rec.WorkerID)
	assert.Equal(t, "mdns", rec.WorkerDiscoverySource)
	assert.Equal(t, int32(8), rec.WorkerCPUCores)
	assert.False(t, rec.Success)
	// PreprocessedSource length is the only source data sent.
	assert.Equal(t, len("int main() { return 0; }"), rec.PreprocessedSizeBytes)
	assert.Equal(t, len("int main() { return 0; }"), rec.SourceSizeBytes)
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
