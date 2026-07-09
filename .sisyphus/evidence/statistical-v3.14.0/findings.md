# Statistical Benchmark — 4 schedulers × 10 reps (CPython v3.14.0)

**Date:** 2026-07-09
**Harness:** `scripts/benchmark_statistical.sh` (REPS=10, ALPHA=0.5)
**Cluster:** 5w-hetero (0.5+0.6+0.8+1.0+1.1 = 4.0 CPU, unequal)
**Workload:** CPython **v3.14.0** pinned tag, `./configure --disable-test-modules --disable-perf-trampoline`
**Analysis:** `scripts/analyze_benchmark.py` (Mann-Whitney U, one-sided `hybrid < baseline`)

---

## 0. Workload note — NOT comparable to the 873-task tables

This run uses the **pinned v3.14.0 tag**, which emits **293 compile tasks** per
build (~45 s makespan). The earlier thesis tables (`.sisyphus/evidence/m1/findings.md`,
`docs/thesis/*`) used a heavier snapshot with **873 tasks** (~152 s for LeastLoaded on
5w-hetero). **Absolute numbers here must not be merged into those tables** — present this
as an independent, lighter-workload measurement set. A heavier-workload confirmation run
(clone `main`) is tracked separately.

The low noise floor is a benefit of this run: per-scheduler makespan std is 0.85–2.07 s
across 10 reps, so the significance tests are well-powered despite the short makespan.

---

## 1. Makespan (wall-clock build time, seconds)

| Scheduler | n | mean | std | min | median | max |
|---|---|---|---|---|---|---|
| **hybrid-linucb** | 10 | **41.2** | 1.03 | 40 | 41.0 | 43 |
| leastloaded | 10 | 44.5 | 0.85 | 43 | 44.5 | 46 |
| p2c | 10 | 46.5 | 1.27 | 45 | 46.5 | 48 |
| linucb | 10 | 48.6 | 2.07 | 46 | 48.0 | 52 |

### Mann-Whitney U (one-sided, `hybrid-linucb < baseline`)

| vs baseline | U | p | median diff | verdict (α=0.05) |
|---|---|---|---|---|
| leastloaded | 0.5 | 0.0001 | −3.5 s | **SIGNIFICANT** |
| p2c | 0.0 | 0.0001 | −5.5 s | **SIGNIFICANT** |
| linucb | 0.0 | 0.0001 | −7.0 s | **SIGNIFICANT** |

hybrid-linucb is faster than leastloaded by **7.9%**, p2c by **11.4%**, linucb by **15.2%**.

## 2. Tail latency — per-run P99 compile time (ms)

| Scheduler | runs | P99 mean | P99 std |
|---|---|---|---|
| **hybrid-linucb** | 10 | **5 735** | 1 460 |
| p2c | 10 | 6 684 | 630 |
| leastloaded | 10 | 6 814 | 1 247 |
| linucb | 10 | 7 237 | 1 752 |

### Mann-Whitney U (one-sided, `hybrid-linucb < baseline`)

| vs baseline | U | p | median diff | verdict (α=0.05) |
|---|---|---|---|---|
| leastloaded | 29.0 | 0.0606 | −1 634 ms | not significant |
| p2c | 23.0 | 0.0226 | −1 535 ms | **SIGNIFICANT** |
| linucb | 21.0 | 0.0156 | −1 407 ms | **SIGNIFICANT** |

P99 has high per-run variance, so the vs-leastloaded gap (p=0.061) falls just short of
significance despite a −1.6 s median improvement. Report as a positive trend, not a claim.

## 3. Interpretation — hybrid design is robust at light load

The headline finding is a **regime inversion** relative to the 873-task tables:

- At light load (293 tasks, no sustained queue contention), **P2C loses its theoretical
  edge** (46.5 s, slower than LeastLoaded's 44.5 s) — the two-random-choice overhead is
  not amortised when no worker backs up.
- **Pure LinUCB is worst** (48.6 s) — it pays exploration cost with little contention to
  exploit.
- **hybrid-linucb wins anyway** (41.2 s, p=0.0001 vs all three). Its warm-start (N=100)
  plus load-penalty design avoids both the cold-start exploration tax and the P2C overhead
  trap.

This complements (does not replace) the heavy-load story: the hybrid scheduler is the only
one that stays ahead across *both* the light regime measured here and the contention-heavy
873-task regime measured earlier.

## 4. Honest limitations

1. Single-machine Docker cgroup cluster; no real network. Noise floor is low but not zero.
2. Lighter workload than the thesis's 873-task set — heavier confirmation run pending.
3. P99 vs leastloaded not significant (p=0.061).
4. One workload family (CPython). Template-heavy C++ (Boost/Eigen) may shift results.

## 5. Reproduce

```bash
OUT_DIR=/tmp/bench_stat_full REPS=10 \
  SCHEDULERS="leastloaded p2c linucb hybrid-linucb" ALPHA=0.5 \
  bash scripts/benchmark_statistical.sh
python scripts/analyze_benchmark.py /tmp/bench_stat_full
```

Raw artifacts in this directory: `results.csv` (40 rows), `summary.csv`, `p99_summary.csv`,
`makespan_boxplot.png`, `warmup_curve.png`, and 40 × `tasks_<scheduler>_run<N>.jsonl`.
