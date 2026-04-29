# Plan: CostScheduler — Heterogeneous-aware Predictive Scheduling

## Context

Hybrid-Grid hiện dùng `LeastLoadedScheduler` (hardcoded tại `grpc.go:177`). P2CScheduler tồn tại nhưng chưa bao giờ được wire. Cả hai đều dùng static heuristic — không học từ dữ liệu thực thi, không phân biệt loại task, không tối ưu cho cluster heterogeneous.

Đây là đóng góp nghiên cứu chính của đồ án: thay thế heuristic bằng **predictive cost model** học từ lịch sử thực thi per-worker-per-task-class. Benchmark baseline (April 2026) cho thấy heuristic chỉ đạt 1.23x speedup trên heterogeneous cluster — CostScheduler target vượt con số này.

### Current Scheduler State (verified from code)
- `scheduler.go:162-320` — P2CScheduler với 7 scoring constants hardcoded
- `scheduler.go:25-26` — `Scheduler` interface: `Select(buildType, arch, clientOS) (*WorkerInfo, error)`
- `grpc.go:177` — hardcoded `NewLeastLoadedScheduler(reg)`
- `metrics/latency.go` — LatencyTracker EWMA (alpha=0.5, default 100ms)
- `metrics/ewma.go` — reusable EWMA struct: `NewEWMA(alpha)`, `Update()`, `Value()`, `IsInitialized()`
- `registry.go` — WorkerInfo has: ActiveTasks, TotalTasks, SuccessfulTasks, FailedTasks, AvgCompileTime, MaxParallel
- P2C `ReportSuccess()` exists nhưng **chưa bao giờ được gọi** từ grpc.go
- Không có CLI flag chọn scheduler

### Data Available at Scheduling Time
- buildType (CPP/Flutter/Unity), targetArch, clientOS
- Worker: CPU cores, memory_bytes, native_arch, docker_available, active_tasks, discovery_source, max_parallel, latency EWMA
- **KHÔNG** có: source size, file name, include count, compiler flags

### Data Available After Task Completion
- `CompilationTimeMs` (worker execution time)
- Queue time (coordinator measured)
- Worker RPC latency
- Success/failure, exit code
- Transfer sizes

---

## Approach: Predictive Cost Model

### Core Idea
Với mỗi cặp (worker, task), ước lượng tổng thời gian hoàn thành:

```
cost(w, t) = predicted_exec_time(w, class(t))
           + queue_wait(w)
           + network_overhead(w, t)
           + capacity_penalty(w)
```

Worker có min cost được chọn. Evaluate toàn bộ candidates (cluster size <20 nên O(N) negligible so với RPC overhead). Không cần P2C random sampling.

### 1. Task Classification

Classify tại scheduling time bằng thông tin available ở `Compile()` call site (`grpc.go:367`):

```go
type TaskClass struct {
    BuildType  pb.BuildType    // CPP, Flutter, Unity
    SizeBucket SizeBucket      // SMALL(<64KB), MEDIUM(64-512KB), LARGE(512KB-4MB), XLARGE(>4MB)
    Arch       pb.Architecture // x86_64, arm64
}

type SizeBucket int
const (
    SizeBucketSmall  SizeBucket = iota // preprocessed source < 64 KB
    SizeBucketMedium                    // 64 KB - 512 KB
    SizeBucketLarge                     // 512 KB - 4 MB
    SizeBucketXLarge                    // > 4 MB
)
```

**Tại sao 4 buckets?** Preprocessed C/C++ source sizes follow bimodal distribution:
- Simple .c files ít includes: <64KB
- Application files với STL: 64-512KB
- Template-heavy (Boost, Qt, Eigen): 512KB-4MB
- Auto-generated code: >4MB

Key format: `"CPP:MEDIUM:x86_64"` — ~84 classes tối đa, thực tế 4-12 classes active.

Source size = `len(req.PreprocessedSource)` hoặc `len(req.RawSource)`. Đã available tại Compile() call site nhưng chưa được truyền vào scheduler.Select() — cần threading qua TaskContext.

### 2. Interface Design

Mở rộng bằng interface mới, giữ nguyên `Scheduler` interface cũ:

```go
// Thêm vào scheduler.go

type TaskContext struct {
    SourceSizeBytes int    // len(preprocessedSource) or len(rawSource)
    Compiler        string // "gcc", "clang", etc.
    SourceFilename  string // for extension heuristics
    IncludeCount    int    // len(includeFiles)
}

type CostAwareScheduler interface {
    Scheduler  // backward compatible — Select() vẫn hoạt động
    SelectWithContext(buildType pb.BuildType, arch pb.Architecture, clientOS string, ctx TaskContext) (*registry.WorkerInfo, error)
    RecordOutcome(workerID string, taskClass TaskClass, execTimeMs float64, success bool)
}
```

**Backward compatibility:** CostScheduler.Select() delegates tới SelectWithContext() với zero TaskContext (default SizeBucketMedium). Flutter/Unity callers không cần sửa.

**Type assertion tại grpc.go:** `if cs, ok := s.scheduler.(CostAwareScheduler); ok { ... }` — scheduler không phải CostAwareScheduler → no-op.

### 3. PerfStore — Historical Performance Tracking

```go
type PerfStore struct {
    mu      sync.RWMutex
    entries map[string]*PerfEntry  // key: "workerID:CPP:MEDIUM:x86_64"
    alpha   float64                // EMA alpha = 0.3
    global  map[string]*EWMA       // per-taskClass global average across all workers
}

type PerfEntry struct {
    ExecTimeEMA *metrics.EWMA  // reuse existing metrics.EWMA (metrics/ewma.go)
    SampleCount int64
    LastUpdated time.Time
}
```

**Alpha = 0.3** (not 0.5 của LatencyTracker):
- RTT changes rapidly → alpha=0.5 responsive
- Compilation time cho cùng task class trên cùng worker ổn định hơn → alpha=0.3 smoother
- Half-life ~3 observations: `−1/ln(0.7) ≈ 2.8`
- Nghĩa là 3 tasks gần nhất dominate nhưng outliers không gây dao động lớn

**Methods:**
- `GetExecTime(workerID, taskClass) float64` — return EMA hoặc cold start estimate
- `GetWorkerAvg(workerID) float64` — average across all classes cho worker (dùng cho queue wait)
- `Record(workerID, taskClass, execTimeMs)` — update EMA + global stats
- `GetGlobalClassAvg(taskClass) float64` — cluster-wide average cho class

### 4. Cost Function

```go
func (s *CostScheduler) estimateCost(w *WorkerInfo, tc TaskClass, ctx TaskContext) float64 {
    // 1. Predicted execution time (EMA or cold start)
    execTime := s.perfStore.GetExecTime(w.ID, tc)

    // 2. Queue wait: each active task delays ~worker's average exec time
    queueWait := float64(w.ActiveTasks) * s.perfStore.GetWorkerAvg(w.ID)

    // 3. Network overhead: 2×RTT + source transfer time
    rtt := s.latencyTracker.Get(w.ID) * 2.0
    bwMBps := 100.0 // LAN default
    if w.DiscoverySource != "mdns" && w.DiscoverySource != "LAN" {
        bwMBps = 10.0 // WAN default
    }
    transferMs := float64(ctx.SourceSizeBytes) / (bwMBps * 1000.0)
    network := rtt + transferMs

    // 4. Capacity penalty (nonlinear near max parallel)
    maxP := float64(max(w.MaxParallel, 4))
    loadRatio := float64(w.ActiveTasks) / maxP
    capPenalty := 0.0
    if loadRatio >= 0.5 {
        capPenalty = (loadRatio - 0.5) * 2.0 * execTime
    }

    return execTime + queueWait + network + capPenalty
}
```

**Tại sao không có additive weights?**
- Tất cả terms có đơn vị tự nhiên: milliseconds
- `predicted_exec_time` = ms, `queue_wait` = ms, `network` = ms, `capacity_penalty` = ms
- Cộng trực tiếp → tổng cost = estimated total time in ms
- Không cần tune k1, k2, k3 như P2C's heuristic score
- Cost function tự adapt qua EMA học từ dữ liệu thực

### 5. Cold Start Strategy (3 levels)

Khi `PerfStore.GetExecTime()` chưa có history cho cặp (worker, class):

**Level 1 — Global class average:**
Workers khác có data cho class này → dùng global EMA, adjust theo hardware ratio:
```
estimate = globalClassAvg × (referenceSpeed / workerSpeed)
workerSpeed = cpuCores × memGB^0.25
```
CPU-dominated với diminishing returns từ memory.

**Level 2 — Worker cross-class:**
Worker có data cho class khác, không có cho class này → dùng worker's average, scale theo bucket ratio:
```
sizeRatios = {SMALL: 0.25, MEDIUM: 1.0, LARGE: 4.0, XLARGE: 12.0}
estimate = workerAvg × sizeRatios[targetBucket] / sizeRatios[avgBucket]
```

**Level 3 — Hardware prior (no history at all):**
```go
baseTimes = {SMALL: 200ms, MEDIUM: 800ms, LARGE: 3000ms, XLARGE: 10000ms}
coreScaling = 1.0 / sqrt(cores / 4.0)  // normalize to 4-core baseline
estimate = baseTimes[bucket] × coreScaling
```
`1/sqrt(cores/4)`: single-file compilation mostly single-threaded, nhưng nhiều cores = ít OS scheduling interference.

**Warm-up blending (SampleCount < 10):**
```
blended = (1 - n/10) × coldEstimate + (n/10) × emaValue
```
Tránh EMA bị dominated bởi 1-2 observations đầu tiên.

### 6. Feedback Loop Wiring

Hook vào `grpc.go` sau `registry.DecrementTasks()`:

**Compile() ~line 478:**
```go
s.registry.DecrementTasks(worker.ID, success, compileTime)
// NEW: feed back to cost scheduler
if cs, ok := s.scheduler.(scheduler.CostAwareScheduler); ok {
    sourceSize := len(req.PreprocessedSource) + len(req.RawSource)
    tc := scheduler.ClassifyTask(pb.BuildType_BUILD_TYPE_CPP, req.TargetArch, sourceSize)
    cs.RecordOutcome(worker.ID, tc, float64(resp.CompilationTimeMs), success)
}
```

**handleFlutterBuild() ~line 729:** same pattern, BUILD_TYPE_FLUTTER, source archive size.

**handleUnityBuild() ~line 905:** same pattern, BUILD_TYPE_UNITY, source archive size.

### 7. CLI Flag

```bash
hg-coord serve --scheduler=cost|p2c|leastloaded|simple
```

**`grpc.go:86` — Config struct:**
```go
type Config struct {
    // ... existing fields ...
    SchedulerType string  // NEW
}
```

**`grpc.go:175` — New() factory:**
```go
func New(cfg Config) *Server {
    reg := registry.NewInMemoryRegistry(cfg.HeartbeatTTL)
    circuitMgr := resilience.NewCircuitManager(resilience.DefaultCircuitConfig())
    sched := newScheduler(cfg.SchedulerType, reg, circuitMgr)  // NEW: factory
    // ...
}

func newScheduler(typ string, reg registry.Registry, cm *resilience.CircuitManager) scheduler.Scheduler {
    switch typ {
    case "simple":
        return scheduler.NewSimpleScheduler(reg)
    case "p2c":
        return scheduler.NewP2CScheduler(scheduler.P2CConfig{Registry: reg, CircuitChecker: cm})
    case "cost":
        return scheduler.NewCostScheduler(scheduler.CostSchedulerConfig{Registry: reg, CircuitChecker: cm})
    default:
        return scheduler.NewLeastLoadedScheduler(reg) // backward compatible default
    }
}
```

**`cmd/hg-coord/main.go`:** thêm flag `--scheduler` (string, default "leastloaded"), pass vào Config.

---

## Files

### New (6 files)
| File | Purpose | LOC estimate |
|---|---|---|
| `internal/coordinator/scheduler/task_class.go` | TaskClass, SizeBucket, ClassifyTask(), TaskContext | ~80 |
| `internal/coordinator/scheduler/perf_store.go` | PerfStore, PerfEntry, cold start 3 levels, warm-up blending | ~200 |
| `internal/coordinator/scheduler/cost_scheduler.go` | CostScheduler, CostSchedulerConfig, estimateCost(), SelectWithContext(), RecordOutcome() | ~250 |
| `internal/coordinator/scheduler/task_class_test.go` | Bucket boundaries, key format, edge cases | ~100 |
| `internal/coordinator/scheduler/perf_store_test.go` | EMA convergence, cold start levels, concurrency (-race), global stats | ~200 |
| `internal/coordinator/scheduler/cost_scheduler_test.go` | Min cost selection, learning after N tasks, arch penalty, circuit breaker, fallback | ~250 |

### Modified (3 files)
| File | Change | Delta |
|---|---|---|
| `internal/coordinator/scheduler/scheduler.go` | Add CostAwareScheduler interface, TaskContext struct | +20 lines |
| `internal/coordinator/server/grpc.go` | SchedulerType in Config, newScheduler factory, 3 feedback hooks, TaskContext threading | +50 lines |
| `cmd/hg-coord/main.go` | --scheduler flag, wire to Config | +10 lines |

### Reused (no changes needed)
- `internal/coordinator/metrics/ewma.go` — `EWMA{alpha, value, initialized}`, `NewEWMA()`, `Update()`, `Value()`, `IsInitialized()`
- `internal/coordinator/metrics/latency.go` — `LatencyTracker`, `Record()`, `Get()`
- `internal/coordinator/registry/registry.go` — `WorkerInfo`, `ListByCapability()`, `IncrementTasks()`, `DecrementTasks()`

---

## Phases

### Phase 1: Foundation (Week 1-2)
- [ ] `task_class.go` + `task_class_test.go`
  - TaskClass struct, SizeBucket enum (4 levels), ClassifyTask()
  - TaskContext struct
  - Tests: boundary values (0, 64KB-1, 64KB, 512KB-1, 512KB, 4MB-1, 4MB, 100MB), key format
- [ ] `perf_store.go` + `perf_store_test.go`
  - PerfStore with EMA (alpha=0.3), PerfEntry, GlobalClassStats
  - Cold start 3 levels, warm-up blending
  - Tests: EMA convergence (5 values → expected EMA), cold start fallback chain, concurrent access (-race), global stats update
- [ ] Add `CostAwareScheduler` interface + `TaskContext` to `scheduler.go`
- **Commit:** `feat(scheduler): add task classification and performance store`

### Phase 2: Core Scheduler (Week 2-3)
- [ ] `cost_scheduler.go` + `cost_scheduler_test.go`
  - CostScheduler struct, CostSchedulerConfig
  - estimateCost() — 4 components
  - SelectWithContext() — filter + min cost selection
  - Select() — delegate to SelectWithContext with zero context
  - RecordOutcome() — update PerfStore + LatencyTracker
  - Tests:
    - `TestCostScheduler_SelectsMinCost` — 3 workers, different hardware
    - `TestCostScheduler_PrefersIdleWorker` — loaded vs idle
    - `TestCostScheduler_LearnedPreference` — after recording outcomes, verify preference shifts
    - `TestCostScheduler_BackwardCompatible_Select` — base Select() works
    - `TestCostScheduler_SkipsOpenCircuit`
    - `TestCostScheduler_SkipsOverloaded`
    - `TestCostScheduler_NetworkAwareness_LAN_vs_WAN`
    - `TestCostScheduler_ColdStart`
    - `TestCostScheduler_ArchPenalty` — cross-compile worker penalized
    - `BenchmarkCostScheduler_Select_10Workers` — prove overhead < 1μs
- **Commit:** `feat(scheduler): implement CostScheduler with predictive cost model`

### Phase 3: Integration (Week 3-4)
- [ ] `grpc.go`: SchedulerType in Config, newScheduler() factory replacing hardcoded line 177
- [ ] `grpc.go`: Feedback hooks tại 3 vị trí (Compile, Flutter, Unity)
- [ ] `grpc.go`: Thread TaskContext from Compile() → SelectWithContext()
- [ ] `cmd/hg-coord/main.go`: --scheduler flag
- [ ] Verify: `go build ./... && go test -race ./...` — zero regressions
- **Commit:** `feat(coordinator): wire CostScheduler with CLI flag and feedback loop`

### Phase 4: Validation & Benchmark (Week 5-6)
- [ ] E2E test: start coordinator với `--scheduler=cost`, run distributed build
- [ ] Benchmark matrix: 4 schedulers × 3 cluster configs (scaling, equal, heterogeneous)
- [ ] Thu thập metrics: wall-clock, P95 task latency, worker utilization balance
- [ ] So sánh cold start vs warm (sau 100+ tasks)
- [ ] Document kết quả trong `docs/BENCHMARK_REPORT_cost_scheduler.md`
- **Commit:** `test(scheduler): add CostScheduler benchmarks and evaluation data`

---

## Verification

### Unit Tests
```bash
go test -race -v ./internal/coordinator/scheduler/...
```

### Full Build + Test
```bash
go build ./... && go test -race ./...
```

### E2E with Each Scheduler
```bash
hg-coord serve --scheduler=cost --grpc-port=9000
# In another terminal:
cd test/stress && ./benchmark-heterogeneous.sh
```

### Benchmark Matrix
```bash
for sched in simple leastloaded p2c cost; do
    hg-coord serve --scheduler=$sched --grpc-port=9000 &
    cd test/stress && ./benchmark-heterogeneous.sh
    # Collect /tmp/benchmark_hetero_results.txt
    kill %1
done
```

---

## Design Justifications (for thesis defense)

| Decision | Alternative | Justification |
|---|---|---|
| Task class by size bucket (4 levels) | Per-file tracking | Per-file unbounded + sparse; 84 classes fit memory, match compilation performance regimes |
| EMA alpha=0.3 | alpha=0.5 (existing), alpha=0.1 | Compile time ổn hơn RTT (α=0.5) nhưng cần responsive hơn SMA (α=0.1). Half-life ~3 tasks |
| All terms in milliseconds | Additive weights (P2C style) | Natural units → no weight tuning needed; model self-adapts through EMA |
| CostAwareScheduler interface | Break existing Scheduler interface | Zero-cost backward compat; type assertion = 1 branch; existing callers unchanged |
| In-memory PerfStore | Persistent storage (SQLite) | Profiles stabilize in ~20 tasks; build system restart acceptable for build cache |
| Hardware-based cold start | P2C fallback for first N tasks | Cost-comparable estimates from task 1; avoids mixing scheduling paradigms |
| Evaluate all candidates | P2C random 2-sample | Cluster <20 nodes; O(N) evaluation < 1μs vs ms-scale RPC; exact > approximate |
| 3-level cold start fallback | Single default value | Graceful degradation: global→worker→hardware; each level adds specificity |

---

## Research Hypotheses (to validate with benchmarks)

- **H1:** CostScheduler warm > P2C heuristic ≥15% makespan trên heterogeneous cluster
- **H2:** Gap widens với increasing heterogeneity (homogeneous → small gap, heterogeneous → large gap)
- **H3:** Prediction error giảm theo thời gian (learning curve visible in first 50 tasks)
- **H4:** CostScheduler ≈ P2C trên homogeneous cluster (no regression)
- **H5:** Straggler recovery nhanh hơn: inject slow worker → CostScheduler routes away faster

---

## Open Questions (for iterative improvement)

1. **Bandwidth estimation**: hardcoded 100MB/s LAN vs 10MB/s WAN. Should this be learned from actual transfer times?
2. **Queue wait model**: assumes active tasks finish in sequence. With parallel execution on multi-core workers, actual wait = `activeTasks / parallelSlots × avgTime`. Worth the complexity?
3. **Task class granularity**: 4 size buckets enough? Should compiler type (gcc vs clang) or language standard (C++17 vs C++20) be a classification axis?
4. **EMA decay**: should entries with no updates for >1 hour be discounted? Worker performance can change (thermal throttling, other workloads).
5. **Straggler detection**: should CostScheduler actively detect stragglers (actual time >> predicted) and re-route? Or leave to circuit breaker?
