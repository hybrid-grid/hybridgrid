# Warm-Bandit Ablation — does persistent state let the bandit learn? (CPython v3.14.0)

**Date:** 2026-07-09
**Harness:** `scripts/benchmark_warm.sh` (persistent coordinator across sequential builds)
**Analysis:** `scripts/analyze_warm.py` (Spearman trend, paired Wilcoxon, bootstrap CI)
**Design:** per scheduler, one long-lived coordinator per session; K=8 sequential
builds per session (bandit accumulates in-memory state across them); SESSIONS=6
independent learning curves. Workload: light (~293 TU). Between builds: make clean +
cache clear + task-log truncate + cooldown, but NO coordinator restart.

Motivation: the rigorous benchmark restarts the coordinator per build, so the bandit
cold-starts every build. Leading hypothesis for the hybrid-vs-leastloaded tie was
that the bandit never accumulates learning. This ablation removes the restart. Code
trace (internal/coordinator/scheduler/linucb.go:87,95,84; server/grpc.go:257,211)
confirmed the bandit state is an in-memory singleton, created once at startup, never
reset — so keeping the process alive = a genuinely warm bandit.

---

## Result: hypothesis REFUTED — persistence does not help

| Test | hybrid-linucb | verdict |
|---|---|---|
| Learning trend (Spearman build_idx vs makespan) | rho=−0.079, p=0.59 | **flat — no learning** |
| build 1 vs build 8 (paired Wilcoxon, n=6) | median Δ=+0.21 s, p=0.42 | **no speedup** |
| WARM (build 8) vs leastloaded (paired, n=6) | median Δ=−0.19 s, p=0.50 | **still a tie** |

hybrid-linucb makespan by build index (mean): 43.8 → 43.2 → 43.3 → 43.3 → 44.1 →
44.3 → 43.5 → 43.4 s. Essentially a flat line across all 8 sequential builds.

Control check: leastloaded (stateless) shows a mild, non-significant downward drift
(rho=−0.281, p=0.053), attributable to host/filesystem warm-up over the first ~2
builds — NOT scheduler learning. The bandit does not even show that drift.

## Why: the bandit converges WITHIN a single build

Each build issues ~293 tasks; warm-start N=100, so after the first 100 tasks the
bandit is already exploiting, with ~193 more tasks to converge. With a small (9-dim)
feature space and a stationary workload (fixed workers, homogeneous compile tasks),
the per-arm A matrix saturates inside build 1. Builds 2–8 add no new information, so
cross-build persistence changes nothing.

## Conclusion for the whole investigation (three converging lines of evidence)

1. **Rigorous randomized-block** (`rigorous-v3.14.0/`): hybrid-linucb TIES leastloaded;
   the earlier "7.9% win, p=0.0001" was a run-order/thermal artifact.
2. **Heavy workload** (371 TU): tie persists; hybrid shows cold-start tail-risk.
3. **Warm-bandit ablation** (this): giving the bandit persistent state does NOT help —
   so the tie is **fundamental, not a cold-start artifact**.

Together these rule out the three obvious objections (experimental noise, too-light
load, insufficient learning). The honest thesis conclusion:

> On this compilation workload (homogeneous tasks, workers static within a build),
> **LeastLoaded is already near-optimal**; a contextual bandit converges to the same
> dispatch quality but cannot beat it — there is no hidden structure for the bandit to
> exploit that least-loaded ignores. The bandit reliably beats only weaker baselines
> (pure LinUCB, P2C) and only under light load.

The methodological contribution (rigorous design + warm ablation that falsifies its
own escape hatch) is stronger than a fragile positive result would have been.

## Where a bandit WOULD win (motivated future work)

The bandit needs a setting where least-loaded is suboptimal — i.e. hidden structure
correlated with the feature vector:
- **cache-affinity**: dispatch a file to the worker that already has it cached
  (per-(worker,file) cache-hit feature) — least-loaded ignores this entirely;
- non-stationary workers (thermal drift, contention) needing change-point adaptation;
- heterogeneous task costs correlated with observable features.

The cache-aware feature is the most promising and is the next ablation.

## Reproduce

```bash
BUILDS=8 SESSIONS=6 SCHEDULERS="hybrid-linucb leastloaded" \
  OUT_DIR=/tmp/bench_warm bash scripts/benchmark_warm.sh
scripts/.venv-bench/bin/python scripts/analyze_warm.py /tmp/bench_warm
```

Artifacts: `results.csv` (96 rows), `warm_by_build.csv`, `warm_learning_curve.png`,
96 × `tasks_<sched>_s<S>_b<K>.jsonl`.
