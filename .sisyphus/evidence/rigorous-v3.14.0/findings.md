# Rigorous Scheduler Benchmark — Randomized Block Design (CPython v3.14.0)

**Date:** 2026-07-09
**Harness:** `scripts/benchmark_rigorous.sh` (randomized complete block design)
**Analysis:** `scripts/analyze_rigorous.py` (Friedman + paired Wilcoxon signed-rank
+ Holm-Bonferroni + Cliff's delta + bootstrap 95% CI)
**Cluster:** 5w-hetero (0.5+0.6+0.8+1.0+1.1 = 4.0 CPU, unequal), `-j5`
**Reps:** 10 rounds; each round = all 4 schedulers once, in a seeded per-round
shuffled order; 1 discarded warm-up build before recording.

Two workloads:
- **light** — `--disable-test-modules --disable-perf-trampoline`, ~293 TU/build
- **heavy** — `--disable-perf-trampoline` only, ~371 TU/build (+27%)

---

## 0. Why this supersedes the earlier "statistical" run

The first pass (`.sisyphus/evidence/statistical-v3.14.0/`) ran all 10 reps of one
scheduler back-to-back (scheduler-major). That confounds scheduler identity with
slow host drift (thermal, background load): the scheduler measured last runs on a
differently-loaded machine. It reported hybrid-linucb beating LeastLoaded by 7.9%
(p=0.0001).

**That win did not survive a confound-controlled design.** Under randomized block
+ paired tests, hybrid-linucb and LeastLoaded are statistically tied. The earlier
result was an experimental artifact. This is the headline methodological lesson of
the study — a rigorous design caught a false positive produced by a careless one.

Confounds removed here: run-order/thermal (randomized block), cold FS/Docker cache
(discarded warm-up), partial-cluster starts (registration barrier), thermal state
(cooldown), second-quantised timing (sub-second wall clock), compile-cache leakage
(cache cleared per build). Bandits still cold-start every build (coordinator
restart) — a handicap that makes any bandit win conservative; see §3.

## 1. LIGHT workload (293 TU) — makespan

| Scheduler | median (s) | mean | std | 95% CI median |
|---|---|---|---|---|
| leastloaded | 40.24 | 40.38 | 1.02 | [39.69, 41.23] |
| **hybrid-linucb** | 40.52 | 40.56 | 1.06 | [39.82, 41.38] |
| p2c | 45.60 | 46.01 | 1.39 | [44.95, 47.28] |
| linucb | 47.48 | 47.11 | 1.45 | [46.33, 48.13] |

Friedman: chi2=24.60, **p=0.00002** (schedulers differ).

Paired Wilcoxon (hybrid < baseline), Holm-corrected:
| vs | med diff | p_holm | Cliff d | verdict |
|---|---|---|---|---|
| leastloaded | +0.09 s | 0.652 | +0.10 (negligible) | **tie** |
| p2c | −5.26 s | 0.003 | −1.00 (large) | **hybrid wins** |
| linucb | −6.80 s | 0.003 | −1.00 (large) | **hybrid wins** |

P99 compile time: Friedman p=0.169 (no detectable difference).

## 2. HEAVY workload (371 TU) — makespan

| Scheduler | median (s) | mean | std | 95% CI median |
|---|---|---|---|---|
| **leastloaded** | 48.65 | 51.36 | 6.78 | [46.37, 56.57] |
| hybrid-linucb | 50.64 | 55.42 | 16.60 | [47.34, 53.92] |
| p2c | 53.48 | 55.22 | 5.08 | [51.61, 60.01] |
| linucb | 60.20 | 59.45 | 6.35 | [53.50, 64.43] |

Friedman: chi2=10.58, **p=0.014** (schedulers differ, driven by linucb lagging).

Paired Wilcoxon (hybrid < baseline), Holm-corrected:
| vs | med diff | p_holm | Cliff d | verdict |
|---|---|---|---|---|
| leastloaded | +0.10 s | 0.820 | +0.25 (small) | **tie / slight loss** |
| p2c | −3.12 s | 0.158 | −0.36 (medium) | not significant |
| linucb | −7.61 s | 0.158 | −0.62 (large) | not significant |

P99 compile time: Friedman p=0.516 (no difference).

### 2.1 hybrid tail-risk (std 16.6)

hybrid-linucb's high std is one catastrophic outlier: **round 10 = 101.93 s**
(~2× normal; the other 9 rounds are 46–55 s). In that same round LeastLoaded was
53.15 s, so it is not pure system noise — the cold-started bandit made a genuinely
bad set of early exploration decisions. This demonstrates the cold-start bandit's
**tail risk**, matching thesis limitation #3 (no change-point / drift handling). The
median (robust to the outlier) is the fair summary: hybrid ≈ 2 s slower than
LeastLoaded.

## 3. Overall conclusion (honest, defensible)

| Comparison | Light 293 TU | Heavy 371 TU |
|---|---|---|
| hybrid vs **leastloaded** | tie (p=0.65) | tie / slight loss (p=0.82) |
| hybrid vs **p2c** | wins (p=0.003, δ=−1.0) | n.s. (p_holm=0.16) |
| hybrid vs **linucb** | wins (p=0.003, δ=−1.0) | n.s. (p_holm=0.16) |
| fastest | leastloaded ≈ hybrid | **leastloaded** |

**The bandit's theoretical advantage does not materialise at these workload scales.**
LeastLoaded (a simple heuristic) is at least as good as hybrid-linucb in both regimes
and is more stable (no tail risk). hybrid-linucb reliably beats only the weaker
learned/randomised baselines (pure LinUCB, P2C), and even that edge weakens under
heavier load plus multiple-comparison correction.

This is a negative result for "the bandit beats everything," but a high-value one:
the rigorous methodology (a) caught the earlier false positive and (b) localises
exactly when the bandit wins (vs weak baselines, light load) and loses (vs the best
heuristic; tail-risk under load).

### Leading hypothesis for why the bandit under-performs
Each build restarts the coordinator, so the bandit **cold-starts every build** and
never accumulates learning across the ~300–370 tasks. The decisive follow-up
experiment is a **persistent ("warm") bandit**: run N sequential builds against one
long-lived coordinator (the realistic production scenario) and measure the learning
curve. If build #k improves for the bandit while flat for LeastLoaded, that isolates
the value of learning; if not, the negative result is stronger still.

## 4. Reproduce

```bash
# light
OUT_DIR=/tmp/bench_rigorous REPS=10 bash scripts/benchmark_rigorous.sh
# heavy
CONFIGURE_FLAGS="--disable-perf-trampoline" \
  OUT_DIR=/tmp/bench_rigorous_heavy REPS=10 bash scripts/benchmark_rigorous.sh
# analyze either
scripts/.venv-bench/bin/python scripts/analyze_rigorous.py <OUT_DIR>
```

Artifacts: `light/` and `heavy/` each hold `results.csv`, `order_log.csv`,
`makespan_blocks.csv`, `makespan_rigorous.png`, and 40 × `tasks_<sched>_round<N>.jsonl`.
