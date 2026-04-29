# Plan: LinUCB Scheduler — Milestone 1 (Measurement Infrastructure)

> **Scope của file này**: Milestone 1 của roadmap LinUCB scheduler. KHÔNG cover toàn bộ implementation LinUCB — đó sẽ là plan riêng (`linucb-scheduler-m2.md`, `linucb-scheduler-m3.md`) sau khi M1 verified.
>
> **Nguyên tắc**: Mọi design choice phải có dẫn chứng (paper citation, code reference, hoặc đánh dấu rõ "needs empirical tuning"). Không suy diễn parameter không có cơ sở.

---

## Bối cảnh

Đề tài chốt hướng: **Contextual bandit (LinUCB) cho distributed compilation scheduling** — xem [docs/thesis/research-plan.md](../../docs/thesis/research-plan.md) và [docs/thesis/annotated-bibliography.md](../../docs/thesis/annotated-bibliography.md). Scheduler hiện tại trong production là `LeastLoadedScheduler` (hardcoded tại `internal/coordinator/server/grpc.go:177` — verified 2026-04-29). `P2CScheduler` đã implement (`internal/coordinator/scheduler/scheduler.go:162-329`) nhưng chưa được wire vào server.

Roadmap 3 milestones:

| Milestone | Phạm vi | Output | Tuần |
|---|---|---|---|
| **M1 (file này)** | Wire P2C + CLI flag + measurement pipeline | Baseline numbers + log infrastructure | 1 |
| M2 | Implement ε-greedy bandit minimal | Sườn paper section 3 (Method), section 5 (Results) | 2 |
| M3 | Upgrade ε-greedy → LinUCB đầy đủ | RL contribution chính, paper draft hoàn thiện | 3-4 |

Milestone 1 tập trung vào **infrastructure**, KHÔNG implement RL agent. Mục đích: pipeline đo lường hoạt động end-to-end để mọi iteration sau đều dùng chung.

### Tại sao bắt đầu từ infrastructure thay vì RL agent?

**Dẫn chứng:** Sculley et al. 2015 ("Hidden Technical Debt in Machine Learning Systems", NIPS 2015, Section 3) chỉ ra ML code chỉ chiếm phần nhỏ; phần lớn effort là data collection, monitoring, configuration. Ousterhout et al. 2013 (Sparrow, SOSP §6) dành 1 section cho measurement methodology trước khi present results.

**Practical:** Nếu code RL agent trước nhưng không có pipeline đo, không cách nào verify agent có học hay không. Pipeline chạy ngay với baseline (P2C) cũng cho ra numbers cần cho thesis (P2C performance current state).

---

## Current State (verified 2026-04-29)

### Code locations

| File | Line | Hiện trạng |
|---|---|---|
| `cmd/hg-coord/main.go:242-256` | Existing CLI flags definition | Chưa có `--scheduler` flag |
| `cmd/hg-coord/main.go:136-148` | `coordserver.Config` setup | Chưa truyền scheduler type |
| `internal/coordinator/server/grpc.go:86-94` | `Config` struct | Chưa có field `SchedulerType` |
| `internal/coordinator/server/grpc.go:175-178` | `New()` factory | Hardcoded `NewLeastLoadedScheduler(reg)` |
| `internal/coordinator/server/grpc.go:367-499` | `Compile()` handler | Hook point cho per-task logging |
| `internal/coordinator/server/grpc.go:478` | `DecrementTasks()` call | Sau line này có thể log per-task outcome |
| `internal/coordinator/scheduler/scheduler.go:162-329` | `P2CScheduler` đầy đủ | Có `ReportSuccess()` chưa từng được gọi từ grpc.go |
| `internal/coordinator/metrics/latency.go` | `LatencyTracker` (EWMA) | Reusable cho LinUCB |
| `test/stress/benchmark-heterogeneous.sh` | Benchmark script | Đã có, dùng làm baseline runner |

### Existing observability

- Prometheus metrics đã có (`internal/observability/metrics/`) — counters: `activeTasks`, `totalTasks`, `cacheHits`...
- Tracing via OTLP (`internal/observability/tracing/`) — span attributes per Compile() call
- Dashboard HTTP server (`internal/observability/dashboard/`) — events stream

### Gap cần đóng cho M1

- Per-task **structured log line** với đầy đủ features ML cần (worker_id, source_size, queue_time, compile_time, etc.) → hiện chỉ có Prometheus aggregates và OTLP traces, không phải per-task CSV/JSON
- Scheduler là pluggable nhưng không có CLI để switch
- P2C có nhưng không được test trong production benchmark

---

## Mục tiêu M1

1. ✅ `hg-coord serve --scheduler=p2c|leastloaded|simple` chạy được
2. ✅ Mỗi compile task ghi 1 dòng log JSON đầy đủ features cho RL training/evaluation
3. ✅ Run benchmark heterogeneous với cả 2 scheduler (LeastLoaded baseline + P2C) → có numbers đầu tiên cho thesis Section 5

**KHÔNG trong scope M1:**
- Implement bandit/RL agent (M2/M3)
- Cache-awareness features (Phase 2 sau)
- HEFT baseline (deferred — cần effort riêng)
- Persistence của log (in-memory + stdout đủ cho M1)

---

## Thiết kế chi tiết

### 1. CLI flag `--scheduler`

**File:** `cmd/hg-coord/main.go`

**Change:**
```go
// Thêm vào dòng 245 area (cùng nhóm với --no-mdns)
serveCmd.Flags().String("scheduler", "leastloaded",
    "Scheduler type: leastloaded, simple, p2c")

// Trong RunE (line ~70):
schedulerType, _ := cmd.Flags().GetString("scheduler")

// Validate (sau line 81):
validSchedulers := map[string]bool{"leastloaded": true, "simple": true, "p2c": true}
if !validSchedulers[schedulerType] {
    return fmt.Errorf("invalid scheduler %q; must be one of: leastloaded, simple, p2c", schedulerType)
}

// Trong cfg setup (line ~148):
cfg.SchedulerType = schedulerType
```

**Default value `leastloaded`:** giữ backward-compatible với production hiện tại (line 177).

### 2. `Config.SchedulerType` field

**File:** `internal/coordinator/server/grpc.go`

**Change tại line 86-94:**
```go
type Config struct {
    Port            int
    AuthToken       string
    HeartbeatTTL    time.Duration
    RequestTimeout  time.Duration
    TLS             hgtls.Config
    Tracing         tracing.Config
    EnableRequestID bool
    SchedulerType   string  // NEW: "leastloaded", "simple", "p2c"
}
```

**Default tại `DefaultConfig()` (line 97-104):**
```go
return Config{
    // ... existing ...
    SchedulerType: "leastloaded",  // NEW
}
```

### 3. Scheduler factory

**File:** `internal/coordinator/server/grpc.go`

**Change tại line 175-178:**
```go
func New(cfg Config) *Server {
    reg := registry.NewInMemoryRegistry(cfg.HeartbeatTTL)
    circuitMgr := resilience.NewCircuitManager(resilience.DefaultCircuitConfig())
    sched := newScheduler(cfg.SchedulerType, reg, circuitMgr)  // NEW factory call
    // ... rest unchanged ...
}

// Helper function added in same file (or new file scheduler_factory.go):
func newScheduler(typ string, reg registry.Registry, cm *resilience.CircuitManager) scheduler.Scheduler {
    switch typ {
    case "simple":
        return scheduler.NewSimpleScheduler(reg)
    case "p2c":
        return scheduler.NewP2CScheduler(scheduler.P2CConfig{
            Registry:       reg,
            CircuitChecker: cm,
        })
    case "leastloaded":
        fallthrough
    default:
        return scheduler.NewLeastLoadedScheduler(reg)
    }
}
```

**Lưu ý:** P2CConfig hiện không có `LatencyTracker` field set → `NewP2CScheduler` tự tạo default (`scheduler.go:183-186`). Acceptable cho M1.

**CircuitChecker interface** (`scheduler.go:170-172`):
```go
type CircuitChecker interface {
    IsOpen(workerID string) bool
}
```
`resilience.CircuitManager` cần expose method `IsOpen(workerID) bool` — **CẦN VERIFY**: kiểm tra interface compatibility trước khi code. Nếu thiếu, cần wrapper hoặc add method.

### 4. Per-task structured logging

**Mục đích:** Mỗi compile task → 1 JSON log line chứa đủ features cho:
- ML training (M3) — load CSV vào pandas/sklearn
- Evaluation (M2/M3) — compute makespan, P95 latency, fairness
- Debugging — trace per-task behavior

**Schema:**

```json
{
  "ts": "2026-04-29T10:30:45.123Z",
  "event": "task_completed",
  "task_id": "abc123",
  "build_type": "cpp",
  "scheduler": "p2c",
  "worker_id": "worker-3",
  "worker_arch": "x86_64",
  "worker_native_arch": "x86_64",
  "worker_cpu_cores": 8,
  "worker_mem_bytes": 8589934592,
  "worker_active_tasks_at_dispatch": 2,
  "worker_max_parallel": 8,
  "worker_discovery_source": "mdns",
  "target_arch": "x86_64",
  "client_os": "linux",
  "source_size_bytes": 524288,
  "preprocessed_size_bytes": 524288,
  "raw_source_size_bytes": 0,
  "queue_time_ms": 5,
  "compile_time_ms": 1234,
  "worker_rpc_latency_ms": 1245,
  "total_duration_ms": 1250,
  "success": true,
  "exit_code": 0,
  "from_cache": false
}
```

**Justification fields:**
- `worker_*` — features cho LinUCB context (cần cho M3)
- `source_size_bytes` — feature dominant (Cortez 2017 §4.2 chỉ ra workload size là feature top-1 cho VM lifetime prediction; tương tự cho compile time là hypothesis cần verify ở M2)
- `queue_time_ms` vs `compile_time_ms` — phân biệt scheduling-induced vs intrinsic latency (Dean & Barroso 2013, "The Tail at Scale", CACM 2013, §3 — tail latency phân tích cần break down theo component)
- `worker_rpc_latency_ms` — network overhead, dùng cho cost function feature
- `worker_active_tasks_at_dispatch` — load state at scheduling time (snapshot, không phải completion time)

**Implementation:**

Tạo file mới `internal/coordinator/server/task_log.go`:

```go
package server

import (
    "encoding/json"
    "io"
    "os"
    "sync"
    "time"

    pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
    "github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

// TaskLogger writes per-task structured records.
// Thread-safe; writes are JSON Lines (one record per line).
type TaskLogger struct {
    mu     sync.Mutex
    w      io.Writer
    closer io.Closer
}

// TaskLogRecord captures all features needed for offline analysis and ML training.
// Field names use snake_case to match Python pandas/sklearn convention.
type TaskLogRecord struct {
    TS                          time.Time `json:"ts"`
    Event                       string    `json:"event"`
    TaskID                      string    `json:"task_id"`
    BuildType                   string    `json:"build_type"`
    Scheduler                   string    `json:"scheduler"`
    WorkerID                    string    `json:"worker_id"`
    WorkerArch                  string    `json:"worker_arch"`
    WorkerNativeArch            string    `json:"worker_native_arch"`
    WorkerCPUCores              int32     `json:"worker_cpu_cores"`
    WorkerMemBytes              int64     `json:"worker_mem_bytes"`
    WorkerActiveTasksAtDispatch int32     `json:"worker_active_tasks_at_dispatch"`
    WorkerMaxParallel           int32     `json:"worker_max_parallel"`
    WorkerDiscoverySource       string    `json:"worker_discovery_source"`
    TargetArch                  string    `json:"target_arch"`
    ClientOS                    string    `json:"client_os"`
    SourceSizeBytes             int       `json:"source_size_bytes"`
    PreprocessedSizeBytes       int       `json:"preprocessed_size_bytes"`
    RawSourceSizeBytes          int       `json:"raw_source_size_bytes"`
    QueueTimeMs                 int64     `json:"queue_time_ms"`
    CompileTimeMs               int64     `json:"compile_time_ms"`
    WorkerRPCLatencyMs          int64     `json:"worker_rpc_latency_ms"`
    TotalDurationMs             int64     `json:"total_duration_ms"`
    Success                     bool      `json:"success"`
    ExitCode                    int32     `json:"exit_code"`
    FromCache                   bool      `json:"from_cache"`
}

func NewTaskLogger(path string) (*TaskLogger, error) {
    if path == "" || path == "stdout" {
        return &TaskLogger{w: os.Stdout}, nil
    }
    f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
    if err != nil {
        return nil, err
    }
    return &TaskLogger{w: f, closer: f}, nil
}

func (l *TaskLogger) Log(r *TaskLogRecord) {
    if l == nil || l.w == nil {
        return
    }
    l.mu.Lock()
    defer l.mu.Unlock()
    enc := json.NewEncoder(l.w)
    _ = enc.Encode(r)  // intentionally swallow error: logging must not crash hot path
}

func (l *TaskLogger) Close() error {
    if l == nil || l.closer == nil {
        return nil
    }
    return l.closer.Close()
}
```

**JSON Lines format chọn vì:** stdlib parsable, streaming-friendly, pandas đọc trực tiếp `pd.read_json(path, lines=True)`. Dẫn chứng: `jsonlines.org` spec; Apache Arrow / Spark đọc native.

### 5. Wire TaskLogger vào Compile()

**File:** `internal/coordinator/server/grpc.go`

**Add field tại line 127-156 `Server` struct:**
```go
type Server struct {
    // ... existing fields ...
    taskLogger *TaskLogger
}
```

**Initialize tại `New()` (line 175-220):**
```go
func New(cfg Config) *Server {
    // ... existing setup ...
    
    taskLogger, err := NewTaskLogger(cfg.TaskLogPath)
    if err != nil {
        log.Warn().Err(err).Msg("Failed to open task log; using stdout")
        taskLogger, _ = NewTaskLogger("")
    }
    
    return &Server{
        // ... existing ...
        taskLogger: taskLogger,
    }
}
```

**Add `Config.TaskLogPath`:**
```go
type Config struct {
    // ... existing ...
    SchedulerType   string
    TaskLogPath     string  // NEW: empty = stdout, else file path
}
```

**Add CLI flag in main.go:**
```go
serveCmd.Flags().String("task-log", "", "Path to per-task JSON log file (default: stdout)")
// ...
taskLogPath, _ := cmd.Flags().GetString("task-log")
cfg.TaskLogPath = taskLogPath
```

**Hook point trong `Compile()` — sau line 478 (`s.registry.DecrementTasks`):**

```go
s.registry.DecrementTasks(worker.ID, success, compileTime)

// NEW: emit per-task structured log
if s.taskLogger != nil {
    var workerCPUCores int32
    var workerMemBytes int64
    var workerNativeArch, workerArch string
    var workerMaxParallel int32
    if worker.Capabilities != nil {
        workerCPUCores = worker.Capabilities.CpuCores
        workerMemBytes = worker.Capabilities.MemoryBytes
        workerNativeArch = worker.Capabilities.NativeArch.String()
        workerArch = worker.Capabilities.NativeArch.String()  // alias for now
    }
    workerMaxParallel = worker.MaxParallel
    
    var compileTimeMs int64
    if resp != nil {
        compileTimeMs = resp.CompilationTimeMs
    }
    var exitCode int32
    if resp != nil {
        exitCode = resp.ExitCode
    }
    
    s.taskLogger.Log(&TaskLogRecord{
        TS:                          time.Now().UTC(),
        Event:                       "task_completed",
        TaskID:                      req.TaskId,
        BuildType:                   "cpp",
        Scheduler:                   s.config.SchedulerType,
        WorkerID:                    worker.ID,
        WorkerArch:                  workerArch,
        WorkerNativeArch:            workerNativeArch,
        WorkerCPUCores:              workerCPUCores,
        WorkerMemBytes:              workerMemBytes,
        WorkerActiveTasksAtDispatch: int32(activeTasksSnapshot),  // captured at line ~423 BEFORE dispatch
        WorkerMaxParallel:           workerMaxParallel,
        WorkerDiscoverySource:       worker.DiscoverySource,
        TargetArch:                  req.TargetArch.String(),
        ClientOS:                    req.ClientOs,
        SourceSizeBytes:             len(req.PreprocessedSource) + len(req.RawSource),
        PreprocessedSizeBytes:       len(req.PreprocessedSource),
        RawSourceSizeBytes:          len(req.RawSource),
        QueueTimeMs:                 queueTime.Milliseconds(),
        CompileTimeMs:               compileTimeMs,
        WorkerRPCLatencyMs:          workerLatency.Milliseconds(),
        TotalDurationMs:             totalDuration.Milliseconds(),
        Success:                     success,
        ExitCode:                    exitCode,
        FromCache:                   false,  // M1: cache hit not tracked here (cache is client-side)
    })
}
```

**LƯU Ý quan trọng:**
- `activeTasksSnapshot` cần capture **trước** `IncrementTasks()` (line 417) để có giá trị "lúc dispatch", không phải lúc complete. Cần thêm `activeAtDispatch := worker.ActiveTasks` ngay trước line 417.
- Cần verify `worker.ActiveTasks` thread-safety: `registry.WorkerInfo` có lock riêng hay snapshot copy? **CẦN ĐỌC `internal/coordinator/registry/registry.go`** trước khi code.
- Snapshot không atomic với scheduling decision (race window) nhưng acceptable cho M1 — log dùng cho offline analysis, không phải online RL.

### 6. Verification

#### Unit test
```bash
go test -race ./internal/coordinator/server/...
```
Phải tồn tại test cho `newScheduler` factory: 4 cases (simple/p2c/leastloaded/invalid).

#### E2E smoke test
```bash
make build
./bin/hg-coord serve --scheduler=p2c --task-log=/tmp/test.jsonl --grpc-port=9099 &
COORD_PID=$!
sleep 2
# Run minimal compile via existing test client (cần locate)
kill $COORD_PID
cat /tmp/test.jsonl | head -3   # verify JSON lines emit
```

#### Benchmark replay
```bash
cd test/stress
SCHEDULER=leastloaded ./benchmark-heterogeneous.sh  # baseline
SCHEDULER=p2c ./benchmark-heterogeneous.sh           # P2C numbers
```

**Cần modify `benchmark-heterogeneous.sh`:** thêm `--scheduler=$SCHEDULER` vào hg-coord serve command line. Currently line 38-42 (verified) hardcodes `command: hg-coord serve --grpc-port=9000 --http-port=8080`.

#### Output expected
- 2 JSON Lines logs ở `/tmp/benchmark_*.jsonl` chứa per-task records
- 2 wall-clock numbers từ benchmark output (tổng makespan)
- So sánh: P2C ≈ LeastLoaded trên homogeneous (theo Mitzenmacher 2001 P2C có lợi thế khi có heterogeneity), P2C khá hơn trên heterogeneous (theo current April 2026 baseline số 1.23x ghi trong `docs/BENCHMARK_REPORT_v0.5.md`)

---

## Files modified/added

### Modified (3 files)

| File | Lines changed | Risk |
|---|---|---|
| `cmd/hg-coord/main.go` | +6 (flags + cfg) | Low |
| `internal/coordinator/server/grpc.go` | +60 (Config fields, factory, log hook, snapshot capture) | Medium — Compile() là hot path |
| `test/stress/benchmark-heterogeneous.sh` | +5 (scheduler env var injection) | Low |

### Added (2 files)

| File | LOC | Test? |
|---|---|---|
| `internal/coordinator/server/task_log.go` | ~80 | Yes — `task_log_test.go` |
| `internal/coordinator/server/scheduler_factory_test.go` | ~60 | Yes |

### Touched (test files)

- `internal/coordinator/server/grpc_test.go` — verify scheduler is configurable

---

## Tasks (M1 work breakdown)

- [ ] **T1.1** Verify `resilience.CircuitManager.IsOpen(workerID) bool` exists; add wrapper if needed
- [ ] **T1.2** Verify `registry.WorkerInfo.ActiveTasks` thread-safety semantics (read existing tests)
- [ ] **T1.3** Add `Config.SchedulerType` + `Config.TaskLogPath` fields
- [ ] **T1.4** Implement `newScheduler` factory function
- [ ] **T1.5** Add CLI flags `--scheduler` and `--task-log`
- [ ] **T1.6** Implement `task_log.go` (TaskLogger, TaskLogRecord, NewTaskLogger)
- [ ] **T1.7** Wire TaskLogger in `New()` and Compile() hook
- [ ] **T1.8** Capture `activeAtDispatch` snapshot before IncrementTasks
- [ ] **T1.9** Unit tests: factory selection, TaskLogger concurrency
- [ ] **T1.10** Modify benchmark scripts to accept scheduler env var
- [ ] **T1.11** Run benchmark heterogeneous × 2 schedulers, capture logs
- [ ] **T1.12** Verify log schema with Python pandas: `pd.read_json('out.jsonl', lines=True)` parses without error
- [ ] **T1.13** Commit: `feat(coordinator): add --scheduler CLI flag and per-task structured logging`

---

## Open questions (cần research thêm trước M2)

Đây là những điều **không** giải quyết ở M1 mà ghi lại để tránh suy diễn:

1. **Reward function design** (Q-learning literature): nên dùng `-compile_time_ms`, `-log(compile_time_ms)`, hay `-(compile_time + queue_time)`? Cần đọc:
   - Sutton & Barto 2018 Ch.3.3 (return formulation)
   - Mao et al. 2019 (Decima) §4.2 — họ dùng negative job slowdown
   - **Action**: prototype 2-3 reward functions trên log data trước khi commit

2. **Feature normalization**: log-scale vs raw cho `source_size_bytes`? Bandit literature recommend gì?
   - Li et al. 2010 (LinUCB) original không address explicitly; assume features đã normalized
   - Lattimore & Szepesvári 2020 Ch.19 — feature scaling cho linear bandits
   - **Action**: empirical comparison, choose dựa trên feature correlation analysis (per research-plan Phase 3.3)

3. **LinUCB exploration parameter α**: paper Li 2010 §3.2 "α = 1 + √(ln(2/δ)/2)" với confidence δ. Practical recommendations từ Chu et al. 2011 (LinUCB extension): α ∈ [0.1, 2.0]. **KHÔNG commit value** ở M1, sẽ tune ở M3 với cross-validation.

4. **Online vs episodic**: build session là 1 episode (clear state mỗi build) hay continuous learning (state persist across builds)?
   - Trade-off: continuous → more sample, persistent learning. Episodic → safer, mỗi build độc lập.
   - **Action**: research Phase 3, default episodic ban đầu vì simpler.

5. **Cold start**: worker mới join cluster, không có history. Strategy?
   - Optimistic init (Sutton & Barto Ch.2.6): high initial Q to encourage exploration
   - Decay từ prior policy (P2C as default until N samples collected)
   - **Action**: M3 design decision, document chosen strategy với citation.

6. **CircuitChecker integration**: `P2CScheduler` filter qua circuit breaker (`scheduler.go:218`). LinUCB cũng cần? Nếu skip open circuits, lose training signal cho recovered workers. Need policy.

---

## Risks

| Risk | Impact | Mitigation |
|---|---|---|
| `worker.ActiveTasks` snapshot race với concurrent dispatch | Log có giá trị stale 1-2 tasks off | Acceptable cho M1 (offline analysis), refine M3 nếu cần atomic snapshot |
| TaskLogger I/O latency cản hot path | Compile() chậm thêm vài µs | JSON encoder + mutex << gRPC RTT (ms scale); benchmark verify |
| `CircuitChecker` interface mismatch | Build fail | T1.1 verify trước khi code |
| Benchmark scripts chạy lại cho 2 schedulers tốn thời gian | Wall-clock test 30+ phút | Run trong background while code other tasks |
| Feedback hooks introduce bugs trong production-critical Compile() | Regression | Full test suite + race detector trước commit |

---

## Verification trước khi cho M1 done

- [ ] `go build ./...` clean
- [ ] `go test -race ./...` all pass
- [ ] `golangci-lint run` clean (per project standard, kiểm tra `.golangci.yml`)
- [ ] E2E smoke test: 1 compile task → 1 log line correct schema
- [ ] Benchmark heterogeneous chạy với cả 3 schedulers (leastloaded, simple, p2c) không crash
- [ ] Log file parse bằng Python pandas không lỗi schema
- [ ] Benchmark numbers reproduce baseline (within ±10% noise) so với `docs/BENCHMARK_REPORT_v0.5.md` cho LeastLoaded
- [ ] Commit message rõ scope, link issue/research-plan
- [ ] Update `docs/thesis/research-notes/week-1-[name].md` với findings

---

## References

| Citation | Used for | Where in plan |
|---|---|---|
| Mitzenmacher 2001, IEEE TPDS — DOI 10.1109/71.963420 | P2C theoretical baseline | §1, §6 |
| Ousterhout et al. 2013 (Sparrow), SOSP — DOI 10.1145/2517349.2522716 | Measurement methodology precedent | §1 |
| Sculley et al. 2015, "Hidden Technical Debt in ML Systems", NIPS | Why infrastructure first | §1 |
| Cortez et al. 2017 (Resource Central), SOSP — DOI 10.1145/3132747.3132772 | Workload size as predictive feature | §4 schema rationale |
| Dean & Barroso 2013, "The Tail at Scale", CACM 56(2) — DOI 10.1145/2408776.2408794 | Tail latency decomposition | §4 schema rationale |
| Li et al. 2010 (LinUCB), WWW — DOI 10.1145/1772690.1772758 | LinUCB algorithm reference | Open Q3 |
| Chu et al. 2011 "Contextual Bandits with Linear Payoff Functions", AISTATS | LinUCB α tuning | Open Q3 |
| Sutton & Barto 2018 (RL textbook 2nd ed) | Reward design, optimistic init | Open Q1, Q5 |
| Lattimore & Szepesvári 2020 *Bandit Algorithms* | Feature scaling | Open Q2 |
| Mao et al. 2019 (Decima), SIGCOMM — DOI 10.1145/3341302.3342080 | Reward formulation precedent | Open Q1 |
| `docs/BENCHMARK_REPORT_v0.5.md` (this repo) | Current P2C baseline numbers | §6 verify |

---

## Definition of Done (M1)

✅ Có thể chạy `hg-coord serve --scheduler=p2c --task-log=/tmp/m1.jsonl`  
✅ File `/tmp/m1.jsonl` chứa 1 JSON line per task với schema đầy đủ  
✅ Pandas đọc được file không lỗi  
✅ Có 2 benchmark logs (leastloaded + p2c) trên heterogeneous cluster  
✅ Tests + race detector pass  
✅ Plan M2 (`linucb-scheduler-m2.md`) đã được draft dựa trên insights từ M1 logs

Sau khi DoD đạt, **M1 commit + push**, viết `linucb-scheduler-m2.md` (ε-greedy bandit), bắt đầu M2.
