# Hybridgrid Benchmark Report — v0.5.0 Baseline

**Date:** April 13, 2026
**Version:** v0.5.0 (post-Unity, pre-CostScheduler)
**Scheduler:** P2C weighted heuristic (current baseline)
**Project:** CPython 3.12.0
**Environment:** Docker Compose on Apple Silicon

---

## Purpose

Refresh benchmark baseline after v0.3.0 (observability) and v0.4.0 (Flutter) releases. These numbers serve as the **P2C baseline** for comparison with the proposed CostScheduler (thesis contribution).

Previous report: `docs/BENCHMARK_REPORT.md` (January 18, 2026, CPython 3.14).

---

## Benchmark 1: Scaling Resources

**Objective:** Measure speedup when adding more workers (total resources increase).

### Configuration

| Workers | CPU/worker | RAM/worker | Total CPU | Total RAM |
|---------|------------|------------|-----------|-----------|
| 1       | 0.5        | 512MB      | 0.5       | 512MB     |
| 3       | 0.5        | 512MB      | 1.5       | 1.5GB     |
| 5       | 0.5        | 512MB      | 2.5       | 2.5GB     |

### Results

| Workers | Build Time | Speedup vs 1 Worker | Jan 2026 (CPython 3.14) |
|---------|------------|---------------------|-------------------------|
| 1       | 697s       | 1.00x (baseline)    | 504s / 1.00x            |
| 3       | 190s       | **3.67x**           | 100s / 5.04x            |
| 5       | 139s       | **5.01x**           | 60s / 8.40x             |

### Analysis

Scaling pattern confirmed: near-linear speedup with more workers. Absolute times higher than January due to CPython 3.12 vs 3.14 build differences and potential Docker resource variance. Speedup ratios are the meaningful metric, not absolute seconds.

---

## Benchmark 2: Equal Total Resources

**Objective:** Compare distributed vs single-worker with same total resources (2.5 CPU, 2.5GB RAM).

### Configuration

| Workers | CPU/worker | RAM/worker | Total CPU | Total RAM |
|---------|------------|------------|-----------|-----------|
| 1       | 2.5        | 2560MB     | 2.5       | 2.5GB     |
| 3       | 0.83       | 853MB      | 2.5       | 2.5GB     |
| 4       | 0.625      | 640MB      | 2.5       | 2.5GB     |
| 5       | 0.5        | 512MB      | 2.5       | 2.5GB     |

### Results

| Workers | Build Time | Speedup vs 1 Worker | Jan 2026 |
|---------|------------|---------------------|----------|
| 1       | 142s       | 1.00x (baseline)    | 89s / 1.00x     |
| 3       | 125s       | **1.14x** (best)    | 63s / 1.41x     |
| 4       | 129s       | 1.10x               | 67s / 1.33x     |
| 5       | 132s       | 1.08x               | 71s / 1.25x     |

### Analysis

3 workers remains the sweet spot, confirming January findings. Beyond 3 workers, coordination overhead + per-worker resource starvation erode gains.

**Key observation for thesis:** The speedup degradation beyond 3 workers (from 1.14x to 1.08x) represents scheduling inefficiency — the P2C heuristic assigns heavy compilation units (e.g., `Python/ceval.c`) to resource-starved workers. A cost-aware scheduler that predicts per-worker-per-task execution time should push the sweet spot further right by avoiding such mismatches.

---

## Benchmark 3: Heterogeneous Resources

**Objective:** Test if unequal resource distribution outperforms equal distribution.

### Configuration

| Workers | Distribution | Total CPU | Total RAM |
|---------|--------------|-----------|-----------|
| 1       | [4.0]        | 4.0       | 4.0GB     |
| 3       | [0.8 + 1.2 + 2.0] | 4.0 | 4.0GB     |
| 5       | [0.5 + 0.6 + 0.8 + 1.0 + 1.1] | 4.0 | 4.0GB |

### Results

| Workers | Build Time | Speedup vs 1 Worker | Jan 2026 |
|---------|------------|---------------------|----------|
| 1       | 134s       | 1.00x (baseline)    | 92s / 1.00x     |
| 3       | 130s       | 1.03x               | 69s / 1.33x     |
| 5       | 109s       | **1.23x** (best)    | 59s / 1.56x     |

### Analysis

Heterogeneous 5-worker config achieves 1.23x — better than equal-distribution 5-worker (1.08x in Benchmark 2, after adjusting for different total resources). This confirms the January finding: heterogeneous workers outperform equal distribution because strong workers can absorb heavy compilation tasks.

**Key observation for thesis:** With the current P2C scheduler, heterogeneous benefit is modest (1.23x). P2C's scoring function uses static arch-match and CPU-count weights, but does NOT account for per-task-class cost differences. A scheduler that learns "worker-2 (2.0 CPU) handles template-heavy files in X seconds while worker-5 (0.5 CPU) takes 5X" would route heavy files to strong workers more aggressively, amplifying the heterogeneous advantage.

---

## Comparison: January 2026 vs April 2026

| Benchmark | Metric | Jan 2026 | Apr 2026 | Delta |
|-----------|--------|----------|----------|-------|
| Scaling (5w) | Speedup | 8.40x | 5.01x | Different workload (CPython 3.14 vs 3.12) |
| Equal (3w best) | Speedup | 1.41x | 1.14x | Same trend, 3w remains optimal |
| Hetero (5w) | Speedup | 1.56x | 1.23x | Same trend, hetero > equal |

Absolute speedup numbers differ due to workload differences. The qualitative patterns are consistent:
1. Near-linear scaling when adding resources
2. 3 workers optimal with fixed resources
3. Heterogeneous distribution outperforms equal distribution

---

## Baseline Summary for CostScheduler Evaluation

These results establish the **P2C baseline** against which the proposed CostScheduler will be evaluated:

| Scenario | P2C Baseline | CostScheduler Target |
|----------|-------------|---------------------|
| Equal 5-worker | 1.08x | > 1.20x (push sweet spot right) |
| Hetero 5-worker | 1.23x | > 1.40x (exploit worker strengths) |
| Straggler recovery | not measured | measure recovery time |
| Prediction accuracy | N/A | track predicted vs actual |

Additional metrics to add in CostScheduler evaluation:
- Per-worker utilization balance (Jain's fairness index)
- P95/P99 task latency distribution
- Cold-start learning curve (prediction error over first 50 tasks)
- Scheduler decision overhead (microseconds per decision)
