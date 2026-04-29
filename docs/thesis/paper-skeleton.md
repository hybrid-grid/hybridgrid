# Paper Skeleton — Bandit Scheduling for Distributed Compilation

> **Working title:** Online Contextual-Bandit Scheduling for Distributed
> Compilation on Heterogeneous Clusters
>
> **Status:** Skeleton populated from M1 measurements (2026-04-29). M2
> (ε-greedy) and M3 (LinUCB) numbers fill the empty cells.
>
> **Audience:** SIGCOMM-style systems venue or thesis chapter.

---

## §1 Introduction

**Hook (from M1 evidence):** Distributed build systems target heterogeneous worker pools — laptops, CI runners, dedicated build farms — where per-worker compile time can vary by an order of magnitude for the same source. Existing schedulers (LeastLoaded, P2C) make decisions from instantaneous queue lengths and static capability scores; neither learns from the per-task latency it observes after dispatch. Our M1 measurement on a 5-worker heterogeneous CPython build shows P2C concentrates 33% of tasks on a single worker, and median compile time has a 29× P99/P50 ratio — the regime where learning from feedback is most valuable.

**Contributions (claimed):**
1. **C1.** A measurement pipeline (§4) that emits per-task structured records sufficient for offline ML training and online RL.
2. **C2.** An ε-greedy bandit scheduler (§3.2) demonstrating online learning from compile-time feedback in a real build system, without any pre-training simulator.
3. **C3.** A contextual-bandit (LinUCB) scheduler (§3.3) that exploits worker capability features to outperform feature-blind learners and static heuristics on heterogeneous clusters.
4. **C4.** Empirical evaluation (§5) on 1/3/5-worker CPython builds comparing four schedulers (LeastLoaded, P2C, ε-greedy, LinUCB) on wall-clock makespan, tail latency, and load balance.

---

## §2 Background and Related Work

### §2.1 R||C_max scheduling

Lenstra, Shmoys, Tardos 1990 (DOI 10.1007/BF01585745) — unrelated parallel-machine scheduling is NP-hard with a 3/2 lower bound and 2-approximation. Distributed compilation is a direct R||C_max instance: worker × file → compile time depends on the (unrelated) pair.

### §2.2 Power of Two Choices

Mitzenmacher 2001 (DOI 10.1109/71.963420) — picking 2 random servers and choosing the less loaded reduces max load from O(log n / log log n) to O(log log n) for homogeneous clusters. Sparrow (Ousterhout et al. 2013, DOI 10.1145/2517349.2522716) extends P2C with late binding. Our M1 results confirm P2C's heterogeneous benefit (1.62× over LeastLoaded on 5w-hetero) but reveal a 41% overhead penalty when only one candidate exists.

### §2.3 Multi-armed bandits

Sutton & Barto 2018 (RL textbook, 2nd ed.) §2.2-2.4 introduce ε-greedy with incremental sample-mean updates. Slivkins 2019 (*Introduction to Multi-Armed Bandits*, arXiv:1904.07272) provides regret bounds. Lattimore & Szepesvári 2020 (*Bandit Algorithms*, ch. 19) treats linear bandits.

### §2.4 Contextual bandits and LinUCB

Li et al. 2010 (DOI 10.1145/1772690.1772758) — LinUCB maintains per-arm `(A_a, b_a)` matrices, computes `θ_a = A_a⁻¹ b_a`, and selects `argmax_a θ_a^T x + α√(x^T A_a⁻¹ x)`. Sublinear regret O(√T) under linear payoff assumptions. Chu et al. 2011 — practical α tuning.

### §2.5 RL for systems scheduling

Decima (Mao et al. 2019, DOI 10.1145/3341302.3342080) and DeepRM (Mao et al. 2016, DOI 10.1145/3005745.3005750) apply policy-gradient RL to job scheduling but require simulator pre-training (Decima §4.3). Quasar (Delimitrou & Kozyrakis 2014, DOI 10.1145/2541940.2541941) and Resource Central (Cortez et al. 2017, DOI 10.1145/3132747.3132772) use supervised ML (collaborative filtering, random forests) — not RL.

### §2.6 Gap

No published RL scheduler for *distributed compilation* with strict online-only learning (no simulator). Heuristics in build systems (BuildXL, distcc, Bazel RBE) all use static capability scoring. Cache-aware scheduling for compilation is unaddressed in the bandit literature.

---

## §3 Method

### §3.1 Problem formulation

State `s_t` at decision time: cluster snapshot — for each worker, capability features (CPU, memory, native arch) and live counters (active tasks, recent latency). Action `a_t`: choose one eligible worker. Reward `r_t = -log(1 + compile_time_ms)` upon task completion (heavy-tail compression; precedent Decima §4.2). Horizon: build session ≈ 800-1500 tasks. Episodic, online, no simulator.

We model this as a **contextual bandit** rather than full MDP: each compile is approximately independent (low cross-task state coupling once queue depth is in the context vector). Slivkins 2019 §1.3 justifies the bandit framing when the agent's actions don't materially shift the future state distribution.

### §3.2 ε-greedy baseline

Per Sutton & Barto §2.4, maintain `Q(a)` as the running mean of observed rewards for worker `a`:

```
Q_{n+1} = Q_n + (R_n - Q_n) / n
```

Selection: with probability ε pick uniform random eligible worker, otherwise `argmax_a Q(a)`. Cold workers (n=0) have Q=0; ε-exploration eventually probes them. ε=0.1 (S&B §2.3 baseline).

This is **feature-blind**: the policy ignores task size, worker hardware, even static capability scores.

### §3.3 LinUCB scheduler [M3 — pending]

Per-worker `A_a ∈ ℝ^{d×d}, b_a ∈ ℝ^d`. Context vector `x_t ∈ ℝ^d` of d ≈ 12 features:
- Task: `log(source_size_bytes)`, build_type one-hot (3), target_arch one-hot (2)
- Worker: `cpu_cores/16`, `mem_bytes/64GB`, native_arch_match, `active_tasks/max_parallel`, `recent_rpc_latency/100ms`
- Interaction: `log(source_size) / cpu_cores`

Selection: `argmax_a θ_a^T x_t + α√(x_t^T A_a⁻¹ x_t)`, with `θ_a = A_a⁻¹ b_a`.

Update on `(a_t, r_t, x_t)`: `A_a ← A_a + x_t x_t^T`, `b_a ← b_a + r_t x_t`. α = 1.0 default (Li 2010 §3.2 prescription).

Sherman-Morrison incremental inverse keeps update cost O(d²).

### §3.4 Reward function

`r_t = -log(1 + compile_time_ms)` — log compresses the 29× P99/P50 spread we observed in M1. Failed tasks: `r = -log(1 + request_timeout_ms)` so the learner downweights consistently failing workers. Ablation in §5.5 compares to raw negative latency.

---

## §4 System

### §4.1 Hybrid-Grid architecture

Three components: (a) `hgbuild` — drop-in `make` shim that hashes preprocessor output and dispatches to coordinator; (b) `hg-coord` — gRPC coordinator with worker registry, scheduler, circuit breakers; (c) `hg-worker` — gRPC executor wrapping `gcc`/`clang`/etc. Workers self-register over mDNS or static `--coordinator` flag. All RPC is gRPC over HTTP/2 with TLS optional.

### §4.2 Scheduler integration

The `Scheduler` interface (`internal/coordinator/scheduler/scheduler.go`) exposes `Select(buildType, arch, clientOS) -> (Worker, error)`. A separate `LearningScheduler` interface (introduced in M2) extends this with `SelectWithDispatchInfo` (returning Q-value and exploration flag for logging) and `RecordOutcome(workerID, reward, success)`. The coordinator's `Compile()` handler uses a type assertion to feed back rewards only when a learner is configured; non-learning schedulers (LeastLoaded, P2C, Simple) are unchanged.

A factory in `internal/coordinator/server/grpc.go` selects the scheduler from `--scheduler={leastloaded,simple,p2c,epsilon-greedy}`. Scheduler-specific options (`--epsilon`, future `--alpha` for LinUCB) are routed through `Config` fields.

### §4.3 Per-task measurement pipeline

Every completed `Compile()` invocation emits one JSON Lines record via a `TaskLogger` (mutex-protected `io.Writer` wrapper, ~80 LOC). The 27-field schema captures:

- Identity: `task_id`, `build_type`, `scheduler`
- Worker context at dispatch: `worker_id`, `worker_cpu_cores`, `worker_mem_bytes`, `worker_active_tasks_at_dispatch`, `worker_max_parallel`, `worker_discovery_source`, `worker_native_arch`
- Task: `source_size_bytes`, `preprocessed_size_bytes`, `target_arch`, `client_os`
- Latency decomposition: `queue_time_ms`, `compile_time_ms`, `worker_rpc_latency_ms`, `total_duration_ms`
- Outcome: `success`, `exit_code`, `from_cache`
- Learner introspection: `q_value_at_dispatch`, `was_exploration` (zero for non-learners)

Files load directly into pandas with `pd.read_json(path, lines=True)` — no transformation step. Validated on 1746 records (873 leastloaded + 873 p2c) collected in M1; 100% schema conformance.

### §4.4 Reproducibility harness

`test/stress/benchmark-heterogeneous.sh` parameterizes the scheduler choice via `SCHEDULER=…` env var, generates a `docker-compose-hetero.yml` for 1/3/5-worker configurations with fixed total CPU (4.0) and unequal per-worker allocation, builds CPython, and reports wall-clock and per-task logs to a Docker volume that is extracted to host with `docker run --rm -v stress_task-logs:/data alpine cp ...`.

---

## §5 Evaluation

### §5.1 Setup

- **Workload:** CPython 3.x make build (≈ 870 compilation tasks per run)
- **Cluster:** Docker Compose with `cpus` cgroup limits per worker
- **Configurations:** 1w-4.0cpu (homogeneous), 3w-hetero (0.8/1.2/2.0 cpu), 5w-hetero (0.5/0.6/0.8/1.0/1.1 cpu)
- **Metrics:** wall-clock makespan, per-worker dispatch count, P50/P95/P99 compile time
- **Repetitions:** *TODO* — single-shot for M1 baseline; M3 evaluation needs ≥5 repetitions for variance estimation

### §5.2 Wall-clock makespan (Table 1)

| Cluster | LeastLoaded | P2C | ε-greedy | LinUCB |
|---|---|---|---|---|
| 1w-4.0cpu | 92 s | 130 s | *M2 pending* | *M3 pending* |
| 3w-hetero | 123 s | 85 s | *M2 pending* | *M3 pending* |
| 5w-hetero | 152 s | 94 s | *M2 pending* | *M3 pending* |

P2C improves 1.45-1.62× over LeastLoaded on heterogeneous clusters, confirming Mitzenmacher 2001's theoretical prediction. The 1-worker regression for P2C (-41%) is from circuit-breaker filtering overhead — addressed in M2 with the single-candidate fast path.

### §5.3 Load balance (per-worker dispatch counts)

5w-hetero scenario, 873 tasks total:

| Scheduler | top:bottom ratio | top worker share |
|---|---|---|
| LeastLoaded | 10.8 : 1 | 33.3% |
| P2C | 8.6 : 1 | 33.3% |
| ε-greedy | *M2 pending* | *M2 pending* |
| LinUCB | *M3 pending* | *M3 pending* |

Both static heuristics overload the strongest worker. Bandit methods that learn per-worker latency should reduce the dispatch skew without sacrificing makespan.

### §5.4 Compile-time tail (P50/P95/P99 ms)

| Scheduler | P50 | P95 | P99 |
|---|---|---|---|
| LeastLoaded | 820 | 6 226 | 23 961 |
| P2C | 704 | 5 830 | 19 347 |
| ε-greedy | *M2 pending* | *M2 pending* | *M2 pending* |
| LinUCB | *M3 pending* | *M3 pending* | *M3 pending* |

P2C reduces P99 by 19% over LeastLoaded — partly placement, partly less queue contention.

### §5.5 Ablations [M3 pending]

- Reward function: raw `-compile_time_ms` vs `-log(1+compile_time_ms)` vs queue-aware
- Feature subset: full LinUCB vs without worker-capability features (degenerates toward ε-greedy)
- ε schedule: fixed vs annealed `ε_t = 1/√t`
- α (LinUCB exploration scale): {0.1, 0.5, 1.0, 2.0}

### §5.6 Failure mode injection [M3 pending]

Inject one slow worker mid-run; measure recovery time (tasks-to-divergence-from-bad-arm) for each scheduler.

---

## §6 Discussion

[Filled after §5 numbers complete]

Likely points:
- LinUCB regret bound assumes linear payoff; real compile time is super-linear in source size — does this matter?
- Online learning without reset across builds — pros and cons, drift risk
- Cache-aware extension as natural follow-up (cache state is a sparse feature)

## §7 Limitations

- Single-host Docker simulation is bandwidth-symmetric; real WAN clusters have asymmetry our network feature doesn't model
- CPython is one workload; templates-heavy C++ (Boost, Eigen) will have different distribution and may shift findings
- No comparison against HEFT (Topcuoglu et al. 2002, DOI 10.1109/71.993206) — should be added before publication

## §8 Conclusion

[Filled last]

---

## Appendix: Reproduction

```bash
# Run any scheduler on heterogeneous cluster
SCHEDULER=epsilon-greedy ./test/stress/benchmark-heterogeneous.sh

# Extract per-task log from Docker volume
docker run --rm -v stress_task-logs:/data alpine cat /data/tasks.jsonl > tasks.jsonl

# Analyze
python3 -c "import pandas as pd; df = pd.read_json('tasks.jsonl', lines=True); print(df.describe())"
```

---

## TODO before submission

- [ ] Replace M2/M3 *pending* cells with measured numbers
- [ ] Add HEFT baseline (§5)
- [ ] ≥5 repetitions per config; report mean ± stddev
- [ ] Identical-hardware comparison with `docs/BENCHMARK_REPORT_v0.5.md` (currently divergent baselines)
- [ ] Statistical significance tests (paired t-test or Wilcoxon)
- [ ] Re-run with diverse workloads (templates-heavy C++, mixed C/CXX)
- [ ] Limitations section finalized after experiments
- [ ] §6 Discussion drafted
- [ ] Citations formatted (BibTeX)
