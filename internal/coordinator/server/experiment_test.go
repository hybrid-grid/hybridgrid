package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/scheduler"
)

func addExperimentWorker(t *testing.T, s *Server, id, address, osName string, cpu, maxParallel int32) {
	t.Helper()
	if err := s.registry.Add(&registry.WorkerInfo{
		ID: id, Address: address, MaxParallel: maxParallel,
		Capabilities: &pb.WorkerCapabilities{
			NativeArch: pb.Architecture_ARCH_X86_64, Os: osName, CpuMillis: cpu,
			Cpp: &pb.CppCapability{Compilers: []string{"gcc"}},
		},
	}); err != nil {
		t.Fatal(err)
	}
}

func TestDispatch_BookingCountAndCandidateSnapshot(t *testing.T) {
	s := New(Config{HeartbeatTTL: time.Minute, SchedulerType: "leastloaded"})
	defer s.Stop()
	buildType, arch := pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64
	if got := s.dispatchWithMetadata(buildType, arch, "linux", scheduler.TaskContext{}); got.err == nil {
		t.Fatal("expected no matching workers")
	}
	if s.Dispatches() != 0 {
		t.Fatal("failed selection counted as booking")
	}
	addExperimentWorker(t, s, "a", "worker-a:50051", "linux", 1100, 3)
	addExperimentWorker(t, s, "b", "worker-b:50051", "linux", 2200, 3)
	addExperimentWorker(t, s, "c", "worker-c:50051", "windows", 900, 3)
	addExperimentWorker(t, s, "d", "worker-d:50051", "linux", 700, 3)
	if err := s.registry.UpdateState("d", registry.WorkerStateUnhealthy); err != nil {
		t.Fatal(err)
	}
	if err := s.registry.IncrementTasks("a"); err != nil {
		t.Fatal(err)
	}
	got := s.dispatchWithMetadata(buildType, arch, "linux", scheduler.TaskContext{})
	if got.err != nil || got.seq != 1 || got.dispatchTS.IsZero() || s.Dispatches() != 1 {
		t.Fatalf("booking: %+v, count=%d", got, s.Dispatches())
	}
	byAddress := make(map[string]DispatchCandidate)
	for _, candidate := range got.candidates {
		byAddress[candidate.Address] = candidate
	}
	if len(byAddress) != 3 {
		t.Fatalf("OS-filtered candidates: %+v", byAddress)
	}
	for _, want := range []DispatchCandidate{
		{Address: "worker-a:50051", Active: 1, MaxParallel: 3, CPUMillis: 1100, Healthy: true},
		{Address: "worker-b:50051", Active: 0, MaxParallel: 3, CPUMillis: 2200, Healthy: true},
		{Address: "worker-d:50051", Active: 0, MaxParallel: 3, CPUMillis: 700, Healthy: false},
	} {
		if byAddress[want.Address] != want {
			t.Errorf("candidate %s = %+v, want %+v", want.Address, byAddress[want.Address], want)
		}
	}
}

func TestCompile_ConcurrentDispatchSequencesInLog(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.jsonl")
	workerAddress, stopWorker := setupTestWorker(t, nil)
	defer stopWorker()
	s, _, cleanup := setupTestServer(t, Config{
		HeartbeatTTL: time.Minute, RequestTimeout: 3 * time.Second,
		SchedulerType: "leastloaded", TaskLogPath: path,
	})
	defer cleanup()
	addExperimentWorker(t, s, "worker", workerAddress, "linux", 1100, 32)
	const count = 16
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			resp, err := s.Compile(context.Background(), &pb.CompileRequest{
				TaskId: fmt.Sprintf("task-%d", i), RawSource: []byte("int x;"),
				TargetArch: pb.Architecture_ARCH_X86_64,
			})
			if err != nil || resp.Status != pb.TaskStatus_STATUS_COMPLETED {
				t.Errorf("Compile(%d): response=%v error=%v", i, resp, err)
			}
		}(i)
	}
	wg.Wait()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var records []TaskLogRecord
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var rec TaskLogRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatal(err)
		}
		records = append(records, rec)
	}
	if len(records) != count || s.Dispatches() != count {
		t.Fatalf("records=%d, dispatches=%d", len(records), s.Dispatches())
	}
	sort.Slice(records, func(i, j int) bool { return records[i].DispatchSeq < records[j].DispatchSeq })
	for i, rec := range records {
		if rec.DispatchSeq != int64(i+1) || rec.DispatchTS.IsZero() || rec.ReceivedTS.IsZero() ||
			rec.WorkerAddress != workerAddress || len(rec.Candidates) != 1 {
			t.Errorf("record %d: %+v", i, rec)
		}
		if i > 0 && rec.DispatchTS.Before(records[i-1].DispatchTS) {
			t.Errorf("dispatch timestamp reversed between seq %d and %d", i, i+1)
		}
	}
}
