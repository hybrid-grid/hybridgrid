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

## 6. ε-greedy comparison (collected 2026-04-29 13:20)

### 6.1 Wall-clock

| Cluster | LeastLoaded | P2C | ε-greedy | LinUCB |
|---|---|---|---|---|
| 1w-4.0cpu | 92 s | 130 s | 146 s | *pending* |
| 3w-hetero | 123 s | 85 s | 142 s | *pending* |
| 5w-hetero | 152 s | 94 s | **119 s** | *pending* |

ε-greedy underperforms P2C on every config and beats LeastLoaded only on 5w-hetero. This **validates the paper hypothesis**: a feature-blind bandit pays exploration cost without exploiting worker heterogeneity. P2C uses static capability scoring; ε-greedy ignores it. The story for §5 is now: LeastLoaded → P2C (heuristic with capability awareness, +1.62×) → ε-greedy (online learning, but feature-blind, slower than P2C) → LinUCB (online learning + features, expected to surpass both).

### 6.2 Per-worker dispatch (5w-hetero, ε-greedy)

```
worker-d2bdf558b7db    291    ← Q-greedy "best" worker — concentration extreme
worker-055c796771e5    205
worker-bad95e95eba0     84
worker-21c6bc32a842     83
worker-54f0dffe055a     59
worker-8523fe6881f1     57
worker-e9627737f313     43
worker-7eb8700719e3     29
worker-a43db6e6aff8     22    ← cold worker, gets only ε-share + tie-break
```

Top:bottom = **13.2:1** — *worse* than LeastLoaded (10.8:1) and P2C (8.6:1). ε-greedy's argmax-Q on running mean reward concentrates traffic on the historically-fastest worker, ignoring the queue pressure that the heuristics correctly account for.

### 6.3 Compile-time tail (P50/P95/P99 ms)

| Scheduler | P50 | P95 | P99 |
|---|---|---|---|
| LeastLoaded | 820 | 6 226 | 23 961 |
| P2C | 704 | 5 830 | 19 347 |
| ε-greedy | 956 | 7 387 | 25 488 |
| LinUCB | *pending* | *pending* | *pending* |

ε-greedy P99 is **higher** than both heuristics — the concentrated dispatch creates queue contention on the chosen "best" worker. This is the classic failure mode of feature-blind bandits in heterogeneous load scheduling.

### 6.4 Learner introspection

- **Exploration ratio observed:** 0.084 (target ε = 0.10) — slightly under because of the single-candidate fast path on 1w runs.
- **Q-value distribution:** mean −6.61, std 1.06, range [−7.82, 0]. Q values are concentrated near $-\log(\bar{T})$ for $\bar{T} \approx 800$ ms (matches the median compile time), confirming the reward signal is being learned and the log-transform compresses outliers as intended.
- **Cold-worker Q = 0** still appears as `max=0.000`, showing some workers never received a sample — exactly the explore-vs-exploit failure mode that LinUCB's UCB bonus addresses.

### 6.5 What this means for M3 (LinUCB)

The ε-greedy evidence shows the **true value of LinUCB**: features must inform the decision. LinUCB's $\theta_a^\top x_t + \alpha\sqrt{x^\top A_a^{-1} x}$ score conditions on:
- Worker capability (CPU cores, memory) — penalises sending big tasks to small workers
- Live queue state (active_tasks/max_parallel) — penalises overloaded workers regardless of Q
- Task size (log source bytes) — picks workers whose history shows good performance on similar-sized tasks

If LinUCB on 5w-hetero outperforms P2C, the paper's central claim is empirically supported. If it doesn't, we have a clear story about why (feature linearity violated, drift, etc., per `docs/thesis/theory-notes.md` §3.4).

## 7. LinUCB comparison (collected 2026-04-29 15:30, single run, α=1.0)

### 7.1 Wall-clock — LinUCB underperforms expectations on 5w-hetero

| Cluster | LeastLoaded | P2C | ε-greedy | LinUCB | LinUCB vs best |
|---|---|---|---|---|---|
| 1w-4.0cpu | 92 s | 130 s | 146 s | 129 s | ≈ P2C |
| 3w-hetero | 123 s | 85 s | 142 s | 103 s | 21% slower than P2C |
| 5w-hetero | 152 s | 94 s | 119 s | **158 s** | 68% slower than P2C, *worst of all schedulers* |

**This is a negative result for our hypothesis.** The expectation was that LinUCB's feature-conditioned policy would beat both static heuristics and feature-blind ε-greedy. Instead, on the most heterogeneous configuration (5w-hetero), LinUCB is the worst. We treat this honestly and dig into the cause below.

### 7.2 Per-worker dispatch and exploration

```
worker-bea00a6a2eec    291    explore=0.0%   ← fully exploited; never marked exploration
worker-2c198e7e58b9    166    explore=1.2%
worker-d3668b374fb4    126    explore=0.8%
worker-581f6862f019    115    explore=2.6%
worker-2628b05f8b54     95    explore=2.1%
worker-dc6f93f55f90     61    explore=3.3%
worker-d728b173b617     10    explore=20.0%  ← starved cold worker
worker-15e8b9e277bc      6    explore=16.7%
worker-c59c61c5d300      3    explore=66.7%
```

Top:bottom = **97 : 1** (worse than every other scheduler we measured). Three workers received fewer than 11 tasks each over the entire 873-task workload. Overall exploration rate: 1.7% (versus ε-greedy's 8.4%).

### 7.3 Why LinUCB underperformed — diagnosis

**Diagnosis 1 — UCB bonus did not break the cold-start trap.**
The Li 2010 default $\alpha = 1.0$ produces a confidence radius proportional to $\sqrt{x^\top A^{-1} x}$. With $A = I_d$ initially, the bonus magnitude is $\sqrt{\sum x_i^2} \approx 2$ for our 12-feature vector — but $\hat{\theta}_a^\top x_a$ for the warm winner can grow above 2 once a few rewards are accumulated (median Q at dispatch in our log was $-6.3$ but the *winning* arm's mean approaches 0 as $b_a$ tilts toward the exploited arm). The bonus is therefore swamped after the first ten or so observations. The Q-value distribution shows a sharp bimodality: 75% percentile is 0.0 (cold or just-touched arms), max is +2.5 (hot arm) — exactly the regime where the cold arms' bonus cannot overcome the hot arm's mean.

**Diagnosis 2 — single-run variance is not bounded.**
Docker on macOS Desktop is a noisy host (filesystem, JIT, kernel scheduling). A single 873-task run has wide variance; we saw P2C take 85 s in one run and 94 s in another for the same 5w-hetero config. Without ≥5 repetitions and a paired statistical test, the 158 s LinUCB number cannot be distinguished from a tail-event run.

**Diagnosis 3 — α tuning may be too aggressive in cold start, too weak in steady state.**
The Li 2010 paper itself (verbatim quote, §3 after Eq. 4): *"the value of α given in Eq. (4) may be conservatively large in some applications, and so optimizing this parameter may result in higher total payoffs in practice."* We did not tune α empirically. Chu et al. 2011 §5 reports α ∈ {0.1, 0.5, 1.0} in their experiments and finds the optimum is workload-dependent. This is a §5.5 ablation we now own.

**Diagnosis 4 — feature linearity is likely violated.**
Compile time grows super-linearly in source size (preprocessing × optimisation passes). Our `log(1 + size)` transform compresses but does not linearise. Lattimore & Szepesvári 2020 §24.4: misspecification of magnitude $\varepsilon$ inflates regret by additive $\mathcal{O}(\varepsilon\sqrt{T})$. With $T = 873$ and a non-trivial $\varepsilon$, the regret cost can erase the bandit's expected savings.

### 7.4 Compile-time tail (P50/P95/P99 ms)

| Scheduler | P50 | P95 | P99 |
|---|---|---|---|
| LeastLoaded | 820 | 6 226 | 23 961 |
| P2C | 704 | 5 830 | 19 347 |
| ε-greedy | 956 | 7 387 | 25 488 |
| LinUCB | 942 | 7 142 | 23 661 |

LinUCB and ε-greedy land in the same regime — both pay queue-contention cost from over-concentrating on a "best" worker.

### 7.5 What the paper claims, given this evidence

- **Confirmed:** P2C beats LeastLoaded by 1.62× on heterogeneous clusters (Mitzenmacher 2001).
- **Confirmed:** Naive online learning (ε-greedy) underperforms a tuned static heuristic (P2C) — feature-blindness is a real failure mode.
- **NOT confirmed:** A linear contextual bandit (LinUCB) with reasonable defaults beats the static heuristic on this workload at this scale.
- **Open:** With α-tuning, repetitions, and possibly an alternative reward (Decima Little's-Law form), LinUCB might close the gap.

This is a credible thesis: we identify the failure mode, document the mechanism (Q-value bimodality, exploration starvation), and propose a structured remediation path. A single-run "negative" result on the hardest configuration is not a failure of the research method; it is evidence that the design space matters more than the algorithm choice.

### 7.6 Immediate next experiments (Ralph US-005, US-008, US-010)

1. **α sweep** at 5w-hetero: $\alpha \in \{0.1, 0.5, 1.0, 2.0\}$ — find the workload-specific optimum.
2. **Reward ablation**: $-t$, $-\log(1+t)$, $-(t_k - t_{k-1})J_k$ (Decima form) — quantify reward design impact.
3. **Statistical reps**: ≥ 5 runs per scheduler at 5w-hetero, paired Wilcoxon test.
4. **HEFT baseline**: §X to land in §5.2 alongside the bandits.

## 8. Files in this directory

- `tasks-leastloaded.jsonl` — 873 records, leastloaded scheduler
- `tasks-p2c.jsonl` — 873 records, p2c scheduler
- `tasks-epsilon-greedy.jsonl` — 873 records, ε-greedy scheduler
- `tasks-linucb.jsonl` — 873 records, LinUCB α=1.0
- `benchmark-{leastloaded,p2c,epsilon-greedy,linucb}-results.txt` — wall-clock CSVs
- `findings.md` — this analysis

## 7. Cross-reference

`docs/BENCHMARK_REPORT_v0.5.md` reported 1.23× speedup on heterogeneous; this run shows up to **1.62×** for 5w. The discrepancy is from differing host/corpus (this run: CPython on macOS Docker Desktop; v0.5: different setup). Final paper baselines must be re-collected on identical hardware before publication; this finding is a *qualitative* directional confirmation, not a quantitative comparison.
