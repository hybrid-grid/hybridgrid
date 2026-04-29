# M1 Findings — Measurement Pipeline Validation + LeastLoaded Baseline

**Date:** 2026-04-29
**Scope:** Validation that the M1 measurement infrastructure produces ML-ready data, plus first heterogeneous-cluster baseline numbers for the LeastLoaded scheduler.

---

## 1. Pipeline correctness

- 873 task records emitted across 3 cluster configs (1, 3, 5 workers)
- 25 fields per record — all populated, JSON Lines well-formed
- `pd.read_json('tasks-leastloaded.jsonl', lines=True)` parses without error
- 100% success rate (no failed compiles)
- Sample record:
  ```json
  {"ts":"2026-04-29T05:51:52.403...","scheduler":"leastloaded",
   "worker_id":"worker-e6091d59c146","worker_cpu_cores":8,
   "source_size_bytes":578471,"compile_time_ms":90,...}
  ```

## 2. Wall-clock baselines (LeastLoaded scheduler)

| Cluster | Workers | Build time |
|---|---|---|
| 1w-4.0cpu | 1 | 92 s |
| 3w-hetero | 3 | 123 s |
| 5w-hetero | 5 | 152 s |

**Observation:** more workers → longer makespan despite identical total CPU. Negative scaling indicates scheduling/network overhead dominates parallelism gains in this regime — exactly the failure mode that motivates a learned scheduler.

## 3. Empirical justifications for M2 design

### 3.1 Heavy tail confirms log reward (M2 §"Reward function")

| Metric | P50 | P95 | P99 |
|---|---|---|---|
| `compile_time_ms` | 820 | 6 226 | 23 961 |
| `source_size_bytes` | 969 818 | 1 363 202 | 2 342 631 |

P99/P50 ratio ≈ 29× for compile time. A 24-second outlier would dominate raw-time sample-mean updates. `r = -log(1 + compile_time_ms)` (M2 default) compresses this ratio to ~3×, justifying Decima's log-slowdown precedent.

### 3.2 Severe load imbalance — the LinUCB target

Per-worker dispatch counts under LeastLoaded across the 5-worker run:

```
worker-e6091d59c146    291      ← arm64, 8 cores  (overloaded)
worker-fa39165b895a    148
worker-c01674ea302a    102
worker-50f933f2103a     93
worker-b32a879251c6     59
worker-64e56d591b88     57
worker-70924164f829     55
worker-6e280c2b165d     41
worker-c1c3554c1ccd     27      ← (underloaded)
```

Top:bottom ratio = **10.8:1**.

LeastLoaded picks the worker with fewest active tasks at decision time, ignoring (a) historical compile-time-per-worker, (b) worker capability vector (CPU cores, memory, native arch). The first task on a worker reduces its eligibility for subsequent tasks regardless of whether that worker is actually fast. Accumulated over 873 tasks the assignment skews badly.

**This is the core gap the bandit-based scheduler addresses** — Q-values per worker (M2 ε-greedy) capture latency-per-worker observation, and feature-conditioned Q (M3 LinUCB) lets the policy generalize to unseen task/worker pairs.

## 4. Files in this directory

- `tasks-leastloaded.jsonl` — full task log, 873 records, 575 KB
- `benchmark-leastloaded-results.txt` — wall-clock CSV from benchmark script
- `findings.md` — this file

## 5. P2C comparison (collected 2026-04-29 13:08)

### 5.1 Wall-clock

| Cluster | LeastLoaded | P2C | Δ |
|---|---|---|---|
| 1w-4.0cpu | 92 s | 130 s | P2C **−41%** (overhead penalty when no choice) |
| 3w-hetero | 123 s | 85 s | P2C **+31%** (1.45× speedup) |
| 5w-hetero | 152 s | 94 s | P2C **+38%** (1.62× speedup) |

Sign of the gap matches Mitzenmacher 2001: P2C's benefit grows with cluster heterogeneity. On a 1-worker "cluster" P2C's circuit-breaker / candidate-filtering pipeline executes for nothing and adds 38 seconds of pure overhead.

### 5.2 Per-worker dispatch (5w-hetero, P2C)

```
worker-322c8c6e3e2e    291      ← same overload pattern at top (1 worker takes 33%)
worker-156dff9e0fa7    154
worker-bd9b85cec892     78
worker-03905ef0a12d     78
worker-63e9ef95e6c5     72
worker-c6fda69bf044     59
worker-6005403c5b8a     54
worker-2ff7e2983775     53
worker-28ed8cbfa37b     34
```

Top:bottom = **8.6:1** (LeastLoaded was 10.8:1). Middle/bottom workers receive more uniform load under P2C, but the top worker still dominates because P2C's "weighted score" favors high-CPU/high-memory workers regardless of their current queue depth at scheduling time.

### 5.3 Compile-time distribution

| Metric | LeastLoaded | P2C | Δ |
|---|---|---|---|
| P50 compile_time_ms | 820 | 704 | −14% |
| P95 compile_time_ms | 6 226 | 5 830 | −6% |
| P99 compile_time_ms | 23 961 | 19 347 | −19% |

P2C reduces both median and P99 — partly from better placement (heavy task to fast worker), partly from less per-worker queue pressure. Heavy tail still present but compressed.

### 5.4 What this means for M2/M3

The 33% top-worker dominance under P2C is **not** addressed by improved load tracking — P2C and LeastLoaded both lock onto the strongest worker via different heuristics. The bandit approach (M2/M3) needs to:
1. **Track per-worker compile time history** (M2 ε-greedy already does — Q value incorporates real latency)
2. **Penalize over-utilization** (M3 LinUCB feature: `active_tasks/max_parallel`)
3. **Reward well-placed tasks regardless of worker capability** — i.e. send small tasks to small workers if they're idle, freeing big workers for big tasks

The 41% overhead penalty on 1-worker P2C also suggests M2/M3 should fast-path single-candidate scheduling.

## 6. Files in this directory

- `tasks-leastloaded.jsonl` — 873 records, leastloaded scheduler
- `tasks-p2c.jsonl` — 873 records, p2c scheduler
- `benchmark-leastloaded-results.txt`, `benchmark-p2c-results.txt` — wall-clock CSVs
- `findings.md` — this analysis

## 7. Cross-reference

`docs/BENCHMARK_REPORT_v0.5.md` reported 1.23× speedup on heterogeneous; this run shows up to **1.62×** for 5w. The discrepancy is from differing host/corpus (this run: CPython on macOS Docker Desktop; v0.5: different setup). Final paper baselines must be re-collected on identical hardware before publication; this finding is a *qualitative* directional confirmation, not a quantitative comparison.
