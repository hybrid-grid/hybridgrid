# Plan: LinUCB Scheduler — Milestone 2 (ε-greedy Bandit)

> **Pre-requisite:** M1 complete (`linucb-scheduler.md` — Definition of Done met).
>
> **Scope:** Implement the simplest possible learning scheduler — ε-greedy
> bandit — that exercises the M1 measurement pipeline end-to-end and
> establishes the paper's Method (§3) and Results (§5) skeleton.
>
> **Why ε-greedy first, not LinUCB?** Sutton & Barto 2018 §2.2 ("A Simple
> Bandit Algorithm") and §2.3 introduce ε-greedy as the entry point for
> all bandit learning. It has zero hyperparameters beyond ε itself, no
> matrix algebra, and serves as the canonical baseline that LinUCB is
> evaluated against (Li et al. 2010 §5 "Comparison with ε-greedy").
> Implementing it first means M3 (LinUCB) has a meaningful baseline to
> show improvement against — *the comparison is the contribution*.

---

## Bối cảnh

After M1, the coordinator can:
- Switch scheduler via `--scheduler` flag (`leastloaded`, `simple`, `p2c`)
- Emit per-task JSON Lines records via `--task-log`

M2 adds a **fourth** scheduler choice — `--scheduler=epsilon-greedy` — and
a feedback path so the scheduler updates its preferences from observed
compile times. This is the smallest possible learning agent.

### Algorithm (Sutton & Barto §2.4 "Incremental Implementation")

```
For each (workerID), maintain Q(workerID) ∈ ℝ — running mean of rewards.
For each (workerID), maintain N(workerID) ∈ ℕ — number of dispatches.

On Select():
    With probability ε: pick uniform random eligible worker (explore)
    Otherwise:           pick argmax Q(workerID) over eligible workers (exploit)

On RecordOutcome(workerID, reward):
    N(workerID) += 1
    Q(workerID) += (reward - Q(workerID)) / N(workerID)    # incremental mean
```

**Why incremental mean (not EMA):** Sutton & Barto §2.4 derives this as
the unbiased sample-mean update. EMA (§2.5) is for non-stationary
problems; we treat compile time per worker as quasi-stationary for now.
This decision is revisited in M3 if drift is observed.

**Reward function (open question — to decide empirically in M2):**
- Candidate A: `r = -compile_time_ms` (raw negative latency)
- Candidate B: `r = -log(1 + compile_time_ms)` (Decima §4.2 uses log slowdown)
- Candidate C: `r = -(compile_time_ms + queue_time_ms)` (account for
  scheduler-induced cost)

M2 ships **Candidate B** as default — log compresses the heavy tail
(typical compile times span 10ms–30s, see `docs/BENCHMARK_REPORT_v0.5.md`)
so a single 30-second outlier doesn't dominate Q updates. Other candidates
are wired but disabled, ablated in M3 evaluation.

### Eligibility filter (mirrors P2C)

```go
eligible := workers
    .filter(state != Unhealthy)
    .filter(activeTasks < maxParallel)
    .filter(circuitChecker.IsOpen() == false)
```

If no eligible workers, return `ErrNoMatchingWorkers` — same contract as
existing schedulers.

### Cold start

Workers with `N(workerID) == 0` are treated as having `Q = 0`. Combined
with ε-greedy exploration, all unseen workers are eventually probed. No
optimistic initialization (Sutton & Barto §2.6) — pure exploration is
adequate for clusters of <20 workers and avoids tuning a second knob.

### Hyperparameter

- `ε ∈ [0, 1]` — exploration rate. Default `0.1` (Sutton & Barto §2.3 fig 2.2 baseline).
- Annealing optional — *not* in M2; M3 may add `ε_t = ε_0 / sqrt(t)` if
  evaluation shows continued exploration cost outweighs benefit.

---

## Files

### New
| File | Purpose | LOC est |
|---|---|---|
| `internal/coordinator/scheduler/epsilon_greedy.go` | EpsilonGreedyScheduler struct, Select, RecordOutcome | ~150 |
| `internal/coordinator/scheduler/epsilon_greedy_test.go` | Convergence, exploration-exploitation balance, concurrency | ~200 |
| `internal/coordinator/scheduler/feedback.go` | Shared `LearningScheduler` interface for feedback hookup | ~30 |

### Modified
| File | Change |
|---|---|
| `internal/coordinator/server/grpc.go` | Wire feedback hook in Compile() success path; extend `newScheduler` factory with `epsilon-greedy` case |
| `cmd/hg-coord/main.go` | Add `--epsilon` flag (default 0.1); validate `epsilon-greedy` as valid `--scheduler` value |

### Schema additions to `TaskLogRecord` (M1 file `task_log.go`)
| Field | Why |
|---|---|
| `q_value_at_dispatch` (float64) | The learner's estimated Q(worker) at the moment of dispatch. Audit trail for M3 evaluation: did the chosen worker really have the best Q? |
| `was_exploration` (bool) | True when ε-greedy chose randomly. Lets us split evaluation into explore vs exploit windows. |

These fields are populated only when `--scheduler=epsilon-greedy`.
For non-learning schedulers they default to zero/false (and are filtered
out in pandas analysis).

---

## Interface

To avoid leaking learner internals into the existing `Scheduler` interface,
introduce a separate optional interface:

```go
// In scheduler/feedback.go
type LearningScheduler interface {
    Scheduler
    // RecordOutcome is called after a task completes. The reward is
    // computed by the caller (typically -log(compile_time_ms)).
    RecordOutcome(workerID string, reward float64, success bool)
    // SelectWithDispatchInfo returns the chosen worker plus the Q value
    // and whether the choice was exploratory. The third value is logged
    // into TaskLogRecord and never affects scheduling.
    SelectWithDispatchInfo(buildType pb.BuildType, arch pb.Architecture, clientOS string) (*registry.WorkerInfo, DispatchInfo, error)
}

type DispatchInfo struct {
    QValueAtDispatch float64
    WasExploration   bool
}
```

The Compile() handler does:

```go
worker, info, err := selectWith(s.scheduler, ...)
// ... dispatch ...
if learner, ok := s.scheduler.(scheduler.LearningScheduler); ok {
    reward := -math.Log(1 + float64(resp.CompilationTimeMs))
    learner.RecordOutcome(worker.ID, reward, success)
}
// Log record now includes info.QValueAtDispatch, info.WasExploration
```

`selectWith` is a small helper that prefers `SelectWithDispatchInfo` if
available, else falls back to `Select` with a zero `DispatchInfo`.

---

## Tasks

- [ ] **T2.1** Define `LearningScheduler` interface + `DispatchInfo` in `feedback.go`
- [ ] **T2.2** Implement `EpsilonGreedyScheduler` (Sutton & Barto §2.4 incremental mean)
- [ ] **T2.3** Unit test: convergence on synthetic 3-worker setup with known mean rewards (50 tasks → Q values within ε of true means)
- [ ] **T2.4** Unit test: exploration rate honored — over 10000 calls with ε=0.3, observed exploration count within ±2σ of 3000
- [ ] **T2.5** Unit test: concurrency safe (-race) under 16 goroutines × 100 selects
- [ ] **T2.6** Extend `newScheduler` factory + CLI validation for `epsilon-greedy`
- [ ] **T2.7** Add `--epsilon` flag (default 0.1, validated to [0,1])
- [ ] **T2.8** Wire `LearningScheduler` feedback in `grpc.go` Compile() (after `DecrementTasks`)
- [ ] **T2.9** Add `q_value_at_dispatch` and `was_exploration` to `TaskLogRecord`
- [ ] **T2.10** Integration test: server with `epsilon-greedy` scheduler, 1 worker registered, after N Compile failures the failing worker's Q drops
- [ ] **T2.11** Run benchmark heterogeneous × `epsilon-greedy` and compare to M1 P2C baseline
- [ ] **T2.12** Draft `docs/thesis/paper-skeleton.md` — Method §3 + Results §5 skeleton populated with M2 numbers
- [ ] **T2.13** Commit

---

## Verification

### Convergence test (synthetic, no Docker)

Three "workers" with hardcoded mean rewards `μ = [-1.0, -2.0, -3.0]`
(higher is better → worker 0 is best). Drive 200 selections, sampling
reward from `N(μ, 0.5²)`. Expected:

- After 200 selections: `Q[0] > Q[1] > Q[2]` (ranking matches)
- `|Q[i] - μ[i]| < 0.3` for each i (sample mean accuracy bound)
- Worker 0 selection ratio > 70% (exploitation works)

### Real benchmark

```bash
SCHEDULER=epsilon-greedy ./benchmark-heterogeneous.sh
# Compare wall-clock and per-task log against P2C baseline collected in M1
```

Hypothesis: ε-greedy ≥ LeastLoaded ≥ Simple, but ε-greedy ≤ P2C in
heterogeneous setting *because* ε-greedy ignores worker capability
features (CPU/memory) — this is exactly the gap LinUCB will close in M3.

This negative result is *deliberate*. The paper story:

> A naive bandit (ε-greedy) learns from rewards but not from observable
> worker features, so it pays an exploration cost without exploiting
> heterogeneity. LinUCB closes this gap by conditioning on a feature
> vector at decision time (§3.4).

---

## Open questions (deferred to M3)

1. **Reward function selection** — A vs B vs C above. Decide via offline
   replay on M2 logs.
2. **Annealing schedule** — does fixed ε=0.1 hurt asymptotic performance?
   Compare to `ε_t = 1/√t` empirically.
3. **State persistence** — currently in-memory. Build session restart
   loses learning. Acceptable for thesis; flagged for production.
4. **Per-class Q values** — should Q be `Q(worker, taskClass)` rather
   than `Q(worker)`? More expressive but exponentially more samples
   needed. M3 LinUCB sidesteps this with linear FA over features.

---

## References

| Citation | Used for |
|---|---|
| Sutton & Barto 2018 *Reinforcement Learning* §2.2-2.6 | ε-greedy algorithm, incremental mean, optimistic init discussion |
| Li et al. 2010 (LinUCB), WWW — DOI 10.1145/1772690.1772758 §5 | ε-greedy as canonical baseline for contextual bandits |
| Mao et al. 2019 (Decima), SIGCOMM — DOI 10.1145/3341302.3342080 §4.2 | Negative log slowdown as reward function precedent |
| `docs/BENCHMARK_REPORT_v0.5.md` (this repo) | Compile-time distribution justifying log reward |

---

## Definition of Done (M2)

✅ `hg-coord serve --scheduler=epsilon-greedy --epsilon=0.1 --task-log=...` runs
✅ TaskLogRecord includes `q_value_at_dispatch` and `was_exploration`
✅ Synthetic convergence test passes (Q ranking matches truth, mean error < 0.3)
✅ Concurrency-safe under -race (16 × 100 selects)
✅ Benchmark heterogeneous numbers collected for `epsilon-greedy`
✅ `docs/thesis/paper-skeleton.md` Method §3.1-3.3 (problem framing, ε-greedy
  algorithm, reward function) drafted with M2 numbers
✅ M3 plan (`linucb-scheduler-m3.md`) drafted, citing what M2 logs revealed
