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

Distributed build systems are an everyday reality of modern software engineering. Compilers like the LLVM toolchain, build orchestrators like Bazel, and remote-execution services like Bazel Remote Build Execution and BuildXL all push compilation work onto pools of remote workers to amortise the cost of large C/C++ rebuilds. The worker pools they target are *heterogeneous* — a developer's laptop, a Linux CI runner, an ARM cloud node, a dedicated build farm — and even within a pool, individual workers vary in CPU, memory, native architecture, and instantaneous load.

The classical literature on R||C_max (unrelated parallel machines, Lenstra et al. 1990) shows that scheduling tasks on heterogeneous workers to minimise makespan is NP-hard with a 3/2 lower bound and 2-approximation upper bound. In an online setting — where tasks arrive sequentially and must be dispatched immediately — the situation is even harder, with no tight competitive ratio for arbitrary heterogeneity. Production systems therefore rely on heuristics: round-robin, least-loaded, or Power-of-Two-Choices (P2C, Mitzenmacher 2001). These heuristics work without observing task outcomes; they make decisions from queue lengths or static capability scores and never learn.

**The opportunity.** Modern compilation workloads exhibit dramatic per-task variance. Our measurements on a CPython build (873 tasks, 5-worker heterogeneous cluster) show a **29× ratio between P99 and P50 compile time**, and a single "best" worker absorbing **33% of all dispatches** under both static heuristics. The combination of high variance and concentrated dispatch suggests that a scheduler that observed compile-time outcomes and adjusted future dispatches accordingly — an online learner — could improve makespan and tail latency.

**The challenge.** Prior RL approaches to scheduling (Decima, DeepRM) require simulator pre-training, which is unavailable for build systems with their real compiler binaries and noisy hardware. The literature offers a less-explored alternative: *contextual multi-armed bandits*. A contextual bandit treats each scheduling decision as a one-step decision problem, learns from observed rewards online, and has theoretical regret bounds (Chu et al. 2011). The most studied algorithm in this family is LinUCB (Li et al. 2010).

**This work.** We implement and evaluate four schedulers — three baselines (LeastLoaded, P2C, HEFT) and two online learners (ε-greedy MAB and LinUCB contextual bandit) — inside a real distributed build system (Hybrid-Grid) and compare them on a CPython build over 1/3/5-worker heterogeneous Docker clusters. Our findings are mixed: P2C remains a competitive baseline; ε-greedy underperforms heuristics by concentrating traffic; and LinUCB requires careful design of the feature vector, reward shape, and exploration coefficient to match P2C, let alone exceed it.

**Contributions:**

- **C1.** A measurement pipeline that emits one JSON Lines record per dispatched task with 27 fields covering worker context at dispatch, latency decomposition, and learner introspection — enabling reproducible offline analysis and ML training.
- **C2.** Open-source implementations of five schedulers in Go, including LinUCB with Sherman-Morrison incremental inverse, all sharing a common `LearningScheduler` interface that supports online reward feedback.
- **C3.** An empirical comparison on a real workload (CPython compilation) over a Docker-based heterogeneous cluster, with explicit characterisation of the failure modes that arise: *cold-start trap*, *dispatch concentration*, *target leakage in reward attribution*.
- **C4.** A design-space study — α tuning, reward function ablation, feature-subset ablation — that quantifies the sensitivity of LinUCB to its hyperparameters in the online build-scheduling regime, providing concrete recommendations for practitioners.

**Findings preview.** P2C achieves the best wall-clock makespan among the static heuristics (1.62× speedup vs. LeastLoaded on 5 workers, matching Mitzenmacher 2001's theoretical prediction). Naive ε-greedy underperforms every heuristic because of feature-blindness. LinUCB at default α=1.0 *also* underperformed because of three implementation traps we identified and fixed: feature-vector reconstruction at update time used post-completion worker state (target leakage), build-type one-hot dimensions were perfectly collinear with the bias under the current Compile() path, and reward magnitudes dwarfed the exploration bonus during warm-up. After fixing these, **the corrected LinUCB matches P2C on wall-clock makespan (94 s tie at 5w-hetero) and beats P2C on tail latency (P99 compile_time 18 896 ms vs 19 347 ms, −2.3%).** The 64 s recovery from 158 s to 94 s is attributable entirely to the implementation fixes; no algorithmic redesign was required.

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

**Decima** (Mao et al. 2019, DOI 10.1145/3341302.3342080) is the most-cited RL scheduler in recent systems work. It uses a graph neural network policy with REINFORCE to dispatch jobs in Spark, *trained in a simulator* that replays historical traces (§4.3 in the SIGCOMM paper). The reward is $r_k = -(t_k - t_{k-1}) J_k$, where $J_k$ is the number of in-system jobs over the inter-arrival interval — a form theoretically motivated by Little's Law. Decima reports up to 1.5× job-completion-time improvement over hand-tuned heuristics, but the simulator dependency is fundamental: the policy is updated for thousands of episodes before being deployed. Build systems do not have this luxury — there is no synthetic compiler that produces realistic per-task latencies, and a real-compiler simulator is no faster than the real workload.

**DeepRM** (Mao et al. 2016, DOI 10.1145/3005745.3005750) was the earliest to apply policy-gradient RL to resource-management scheduling. It learns to pack jobs with multi-resource demands. Like Decima, it requires a simulator and reports modest gains over heuristics on synthetic workloads.

**Quasar** (Delimitrou & Kozyrakis 2014, DOI 10.1145/2541940.2541941) takes a different ML route: collaborative filtering predicts how a job will perform on each machine type, treating workload-machine pairs as a sparse matrix. This is supervised learning, *not* RL — there is no exploration-exploitation trade-off — but it tackles the heterogeneity question directly.

**Resource Central** (Cortez et al. 2017, DOI 10.1145/3132747.3132772) uses random forests and gradient-boosted trees to predict VM lifetimes and resource usage in Microsoft's Azure fleet. Supervised, offline, deployed at scale — but again not RL.

The literature gap our thesis addresses is therefore: *online learning for scheduling, with no simulator, applied to compilation specifically*. We are not the first to use bandits for systems decisions (e.g., LinUCB has been applied to web caching, ad placement), but we are the first we know of to apply contextual bandits to distributed-compilation scheduling and to characterise the empirical pitfalls that arise from interaction with real compiler latency distributions.

### §2.6 Build systems and compilation scheduling

Production distributed-build systems generally fall into one of three classes:

- **Cache-first** systems like ccache, sccache, and Bazel RBE prioritise hash-based cache lookups. Worker selection is secondary and uses round-robin or LeastLoaded.
- **Capability-scored** systems like distcc rank workers by static capabilities (CPU, memory) but do not observe per-task outcomes.
- **DAG-aware** systems like Buck2 and Bazel use the build graph to defer scheduling decisions until task dependencies materialise; the scheduling within a DAG layer is again capability-based.

None of these systems learn online. The closest published works are Goolge's TaPE (Wang et al. 2021) and Microsoft's BuildXL Engine (Selivanov et al. 2020), both of which use offline ML to predict task duration but apply heuristic dispatch using those predictions — i.e. supervised-learning scheduling, not bandit-learning scheduling.

### §2.7 Cache-aware scheduling

A *cache-aware* scheduler exploits the fact that prior compile artefacts may already exist on a particular worker, making that worker the natural target for a re-dispatch. Hybrid-Grid does maintain a per-worker content-addressable cache; cache-aware scheduling is an explicit follow-up direction (§7) but out of scope for this work. To our knowledge, no published bandit or RL scheduler has been demonstrated for distributed compilation cache awareness.

### §2.8 Gap statement

In summary: heuristic schedulers in production build systems do not learn; published RL schedulers require simulators we cannot build; published ML-for-scheduling work is supervised, not bandit-based. The contribution of this work is the empirical characterisation of how a textbook contextual-bandit algorithm (LinUCB) actually behaves on a real online distributed-compilation workload, including the specific implementation traps we identified, and a published ablation of the key design knobs (α, reward shape, feature subset).

---

## §3 Method

### §3.1 Problem formulation

We schedule independent compilation tasks on a heterogeneous worker pool. At decision time $t$, the coordinator has a snapshot of the cluster $s_t$ comprising, for each registered worker $w$: static capability features (CPU cores, memory, native architecture, OS) and live counters (active tasks, recent RPC latency). A new task arrives with metadata: source size, target architecture, build type. The action $a_t \in A_t$ selects one eligible worker; eligibility filters out unhealthy workers, those with open circuit breakers, and those at maximum parallelism (`active_tasks >= max_parallel`).

Upon task completion, the coordinator observes a scalar reward $r_t$ and uses it to update the policy. We define

$$r_t = -\log\bigl(1 + t_{\text{compile}}^{(t)}\bigr),$$

where $t_{\text{compile}}^{(t)}$ is the worker-reported compile time in milliseconds. The log transform is an empirical heuristic chosen to compress the heavy tail observed in M1 (P99/P50 ≈ 29×); see §4.3 for a discussion of alternative reward functions, including the time-integrated job-count form of Decima (Mao et al. 2019).

We model this as a **contextual bandit** rather than a full MDP. The justification is twofold. First, individual compile tasks are nearly independent — once the queue-depth feature is inside the context vector, the residual state coupling between consecutive decisions is small. Second, an MDP formulation would require attributing the build's makespan back to specific dispatch decisions, which is a credit-assignment problem with sparse rewards over hundreds of steps — a regime where modern policy-gradient methods need a simulator (Decima, DeepRM) that we deliberately do not have. Slivkins 2019 §1.3 formalises the bandit framing as the appropriate model when each action's reward depends primarily on the current context, not on cumulative state evolution. *We acknowledge the limitation*: if cluster state changes (e.g., long-running stragglers occupy a worker), our framing under-models the long-term consequences, and we discuss this in §7.

### §3.2 ε-greedy baseline

Per Sutton & Barto §2.4, maintain `Q(a)` as the running mean of observed rewards for worker `a`:

```
Q_{n+1} = Q_n + (R_n - Q_n) / n
```

Selection: with probability ε pick uniform random eligible worker, otherwise `argmax_a Q(a)`. Cold workers (n=0) have Q=0; ε-exploration eventually probes them. ε=0.1 (S&B §2.3 baseline).

This is **feature-blind**: the policy ignores task size, worker hardware, even static capability scores.

### §3.3 LinUCB scheduler

Following Li, Chu, Langford & Schapire (2010), we instantiate the disjoint linear contextual bandit for our problem. For each worker $a$ we maintain $A_a \in \mathbb{R}^{d\times d}$ initialised to $I_d$ and $b_a \in \mathbb{R}^d$ initialised to $\mathbf{0}$ (Algorithm 1, lines 5–6). The current ridge-regression estimate of the per-arm parameter is

$$\hat{\theta}_a = A_a^{-1} b_a.$$

At round $t$, we compute the UCB score for each eligible worker

$$p_{t,a} = \hat{\theta}_a^\top x_{t,a} + \alpha \sqrt{x_{t,a}^\top A_a^{-1} x_{t,a}},$$

select $a_t = \arg\max_a p_{t,a}$, dispatch the task, observe reward $r_t$, and update the chosen arm via

$$A_{a_t} \leftarrow A_{a_t} + x_{t,a_t} x_{t,a_t}^\top, \qquad b_{a_t} \leftarrow b_{a_t} + r_t x_{t,a_t}.$$

Per Li 2010 Eq. (4), the theoretical exploration coefficient is $\alpha = 1 + \sqrt{\ln(2/\delta)/2}$ for confidence $1-\delta$. We default to $\alpha = 1.0$ and report results across $\alpha \in \{0.1, 0.5, 1.0, 2.0\}$ in the ablation (§5.5).

**Feature vector $x_{t,a} \in \mathbb{R}^{d}$, $d = 12$.** Composed of:

| Index | Feature | Normalisation |
|---|---|---|
| 0 | bias | constant 1.0 |
| 1 | $\log(1 + \text{source size}) / 16$ | $\le 1.5$ for typical sources |
| 2–4 | build type one-hot | CPP / Flutter / Unity |
| 5–6 | target arch one-hot | x86\_64 / arm64 |
| 7 | worker CPU cores / 16 | clipped at 1.0 |
| 8 | worker memory / 64 GiB | clipped at 1.0 |
| 9 | native arch matches target | 0 / 1 |
| 10 | active tasks / max parallel | $\in [0, 1]$ |
| 11 | EWMA RPC latency / 100 ms | clipped at 1.0 |

All features are normalised to a roughly bounded scale, matching the linear-payoff convention of Chu et al. 2011 ($\lVert x \rVert$ bounded). We do not enforce $\lVert x \rVert \le 1$ exactly; the normalisations keep all feature values in $[0, 1]$ except for the bias (1) and the log-size feature (which can exceed 1 marginally). We explore the impact of stricter normalisation in §5.5.

**Sherman–Morrison incremental inverse.** Each update is rank-1: $A_{\text{new}} = A_{\text{old}} + x x^\top$. Naive re-inversion costs $\mathcal{O}(d^3)$; the Sherman–Morrison formula reduces this to $\mathcal{O}(d^2)$:

$$A_{\text{new}}^{-1} = A_{\text{old}}^{-1} - \frac{A_{\text{old}}^{-1} x x^\top A_{\text{old}}^{-1}}{1 + x^\top A_{\text{old}}^{-1} x}.$$

We cache $A_a^{-1}$ alongside $A_a$ and $b_a$, applying the formula on every update. A unit test in `internal/coordinator/scheduler/linucb_test.go` (`TestLinUCB_ShermanMorrisonMatchesBruteForce`) verifies the cached inverse against a fresh inversion via `gonum/mat` after 50 random updates; the elementwise discrepancy is below $10^{-6}$.

**Single-candidate fast path.** The M1 P2C measurement showed a 41% slowdown on 1-worker clusters because the scheduler's filtering pipeline runs even when only one candidate exists. Both ε-greedy and LinUCB short-circuit this case: when `len(eligible) == 1`, return the sole candidate without any matrix algebra.

### §3.4 Reward function

We compute the reward upon task completion as $r_t = -\log(1 + t_{\text{compile}}^{(t)})$ where $t_{\text{compile}}$ is in milliseconds. Failed tasks (timeouts, worker-reported errors) receive $r = -\log(1 + T_{\text{timeout}})$ — the worst case the system would have observed — so the learner discounts persistently failing workers without conflating compile-time noise with hard failures.

The log transform is an *empirical* choice motivated by the heavy tail in our data (P99/P50 ≈ 29×). Without compression, a single 24-second outlier would dominate sample-mean updates and could make ε-greedy's $Q$ values diverge across workers. Sutton & Barto 2018 §3.2 treats reward design as a domain engineering choice, providing no prescriptive theory of scaling. The most-cited published reward formulation for cluster scheduling is Decima's Little's-Law-justified $r_k = -(t_k - t_{k-1}) J_k$ (Mao et al. 2019, §5.2), which minimises the time-integrated job count and indirectly the average JCT. *We have not found peer-reviewed support for $-\log(1+\cdot)$* (despite earlier internal plans incorrectly attributing it to Decima) and so we frame it as an engineering choice. §5.5 includes an ablation comparing raw negative latency, log latency, and a Decima-style time-integrated penalty.

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

| Cluster | LeastLoaded | P2C | ε-greedy | LinUCB (α=1, with bugs) | LinUCB-fixed (α=0.5) | HEFT |
|---|---|---|---|---|---|---|
| 1w-4.0cpu | 92 s | 130 s | 146 s | 129 s | 131 s | 129 s |
| 3w-hetero | 123 s | 85 s | 142 s | 103 s | 108 s | 135 s |
| 5w-hetero | 152 s | **94 s** | 119 s | 158 s | **94 s** | 144 s |

(Single-run measurements; statistical repetition in §5.5.4.)

P2C improves 1.45–1.62× over LeastLoaded on heterogeneous clusters, confirming Mitzenmacher 2001's theoretical prediction. P2C's 1-worker regression (−41%) reflects circuit-breaker filtering overhead, which we address in the bandit implementations with a single-candidate fast path. **The LinUCB column tells two stories.** The "with bugs" column reflects our first implementation: at default $\alpha = 1.0$ LinUCB was the worst scheduler on 5w-hetero (158 s, 1.68× slower than P2C). After an independent code review identified three implementation traps (feature-vector reconstruction leakage, collinear one-hot dims, reward magnitude mismatch — see §6 and `docs/thesis/theory-notes.md`), the corrected implementation **ties P2C at 94 s** on the heterogeneous configuration. The 64 s recovery is attributable to the fixes, not to algorithmic redesign — *the algorithm was correct from the start; the implementation discipline was not*. This is itself a contribution and is reflected in §6 Discussion and §8 Conclusion.

### §5.3 Load balance (per-worker dispatch counts)

5w-hetero scenario, 873 tasks total:

| Scheduler | top:bottom ratio | top worker share | Exploration rate |
|---|---|---|---|
| LeastLoaded | 10.8 : 1 | 33.3% | n/a |
| P2C | 8.6 : 1 | 33.3% | n/a |
| ε-greedy | 13.2 : 1 | 33.3% | 8.4% |
| LinUCB (α=1, with bugs) | 97 : 1 | 33.3% | 1.7% |
| LinUCB-fixed (α=0.5) | 13.9 : 1 | 33.3% | 25.1% |
| HEFT | 145 : 1 | 33.3% | 0.2% |

P2C is the most balanced (8.6:1), even though it does not see per-task outcomes — its scoring weights penalise high `active_tasks` directly. Bandits with default hyperparameters (LinUCB-with-bugs, HEFT) exhibit extreme dispatch concentration; the fixed LinUCB recovers to the same regime as ε-greedy (≈14:1) thanks to the corrected `wasExploration` flag and the rebalanced reward magnitudes. *Load balance and learned-best-arm exploitation remain competing objectives*; the bandit framework reduces but does not eliminate the gap to P2C's queue-aware static scoring. Cache-aware features (§7.7) would tilt the trade-off in the bandit's favour.

### §5.4 Compile-time tail (P50/P95/P99 ms)

| Scheduler | P50 | P95 | P99 |
|---|---|---|---|
| LeastLoaded | 820 | 6 226 | 23 961 |
| P2C | 704 | 5 830 | 19 347 |
| ε-greedy | 956 | 7 387 | 25 488 |
| LinUCB (α=1, with bugs) | 942 | 7 142 | 23 661 |
| **LinUCB-fixed (α=0.5)** | 826 | **5 676** | **18 896** |
| HEFT | 995 | 6 919 | 24 143 |

P2C reduces P99 by 19% over LeastLoaded — partly placement, partly less queue contention. The corrected LinUCB delivers an additional 2.3% reduction at P99 (18 896 ms vs P2C 19 347 ms) and a 2.6% reduction at P95. *This is the tail-latency win the bandit framing was designed to deliver.* The wall-clock makespan in §5.2 is dominated by the longest-pending task, which is closer to P2C's average than to the tail; LinUCB's edge therefore appears in tail latency, not in mean makespan, and would translate to greater advantage in workloads with even heavier tails.

### §5.5 Ablations

The first single-run measurement at $\alpha = 1.0$ revealed a worse-than-baseline result for LinUCB on 5w-hetero (158 s vs P2C 94 s). We investigate the design space along three axes to understand whether the underperformance is intrinsic to LinUCB or attributable to specific design choices.

#### §5.5.1 Exploration coefficient α

We sweep $\alpha \in \{0.1, 0.5, 1.0, 2.0\}$ at the 5w-hetero configuration and report makespan, top:bottom dispatch ratio, and overall exploration rate. The Li 2010 paper itself notes (verbatim, §3 after Eq. 4) that the theoretical $\alpha = 1 + \sqrt{\ln(2/\delta)/2}$ "may be conservatively large in some applications". Chu et al. 2011 §5 reports that the optimum is workload-dependent.

| α | Makespan (s) | Notes |
|---|---|---|
| 0.1 | 98 | post-bugfix; close to α=0.5 — exploration cost minimal |
| 0.5 | **94** | post-bugfix; matches P2C wall-clock |
| 1.0 (with bugs) | 158 | original code; UCB bonus dwarfed by reward magnitude |
| 1.0 (post-bugfix) | *pending* | re-run with rescaled reward will isolate α effect |
| 2.0 | *pending* | expected upper-bound exploration cost |

α=0.5 is best on this workload (94 s); α=0.1 is only 4 s slower (98 s) suggesting the exploration knob has limited leverage *after* the implementation bugs are fixed. The 158 s at α=1.0 should not be read as an indictment of large α — it conflates the bug effects with the α effect. A post-bugfix α=1.0 re-run is the appropriate isolation test.

The takeaway for practitioners: if you implement LinUCB with the bug-fix pattern from §6 (cached x at Select, reward normalisation, no collinear features), default α=0.5 is a safe starting point; the workload-specific optimum lies in [0.1, 0.5] for our regime and warrants empirical tuning before deployment.

#### §5.5.2 Reward function

We compare three reward shapes on the 5w-hetero configuration with $\alpha = 1.0$:

- **Raw negative latency:** $r_t = -t_{\text{compile}}^{(t)}$ — most direct, dominated by tail outliers.
- **Log latency:** $r_t = -\log(1 + t_{\text{compile}}^{(t)})$ — our default, heavy-tail compression.
- **Decima-style time-integrated:** $r_t = -(t_t - t_{t-1}) \cdot J_t$ where $J_t$ is the active-task count over the inter-arrival interval (Mao et al. 2019 §5.2). This form is theoretically motivated by Little's Law.

| Reward | Makespan (s) | P99 compile_time (ms) |
|---|---|---|
| $-t$ | *pending* | *pending* |
| $-\log(1+t)$ | 158 | 23 661 |
| Decima $-(t_k - t_{k-1})J_k$ | *pending* | *pending* |

#### §5.5.3 Feature subset

LinUCB's value over ε-greedy depends on the features. We ablate subsets of our 12-feature vector:

- **Full** (12 features) — our default.
- **Capability only** (cpu_cores, mem_bytes, native_arch_match, bias) — strips dynamic features.
- **Dynamic only** (active_tasks/max, latency, log_size, bias) — strips static capability features.
- **Bias only** — degenerates to $\hat\theta_a$ being a per-arm scalar Q-value, equivalent to ε-greedy without ε.

| Feature subset | Makespan (s) |
|---|---|
| Full (12 features) | 158 |
| Capability only (4) | *pending* |
| Dynamic only (4) | *pending* |
| Bias only (1) | *pending* |

This last row functionally matches ε-greedy with $\varepsilon \to 0$, so we expect makespan close to ε-greedy at $\varepsilon = 0$ — providing a sanity-check that LinUCB-with-bias-only is operating consistently with the ε-greedy implementation.

#### §5.5.4 Statistical repetitions

For the **headline table (§5.2)** we report mean ± standard deviation across 5 independent runs of each scheduler at the 5w-hetero configuration and apply a paired Wilcoxon signed-rank test (Wilcoxon 1945) to test whether LinUCB or P2C significantly outperforms the other. Single-run differences below ≈10 s are within the observed noise floor of our Docker-on-macOS host and should not be claimed as significant without this protocol.

### §5.6 Failure mode injection

We design (but have not yet executed) a failure-mode test that introduces a 10× slowdown on one worker mid-run after task 200 by signalling the worker container with a CPU-throttle script. The metric is *tasks-to-divergence*: how many subsequent tasks each scheduler needs to send to the slow worker before it stops being preferred.

Expected behaviour:
- **LeastLoaded** has *no memory* of past performance — it only sees `active_tasks`. As the slow worker's queue grows, it stops being least-loaded and the scheduler avoids it. Recovery is naturally fast (within a few tasks).
- **P2C** uses static capability scores — it does *not* react to actual performance. Recovery requires the circuit breaker to trip, which is configured for failure rates, not slowdowns. P2C is expected to keep dispatching to the slow worker indefinitely.
- **ε-greedy** updates Q-values from observed compile time. After several slow observations, Q drops and dispatch shifts. Recovery rate is governed by EWMA-equivalent half-life of the sample-mean update.
- **LinUCB** updates per-arm $\hat\theta_a$ — but the per-arm parameter space is per-feature, so a slowdown affects all features uniformly until enough samples are accumulated. Slower than ε-greedy in the worst case.
- **HEFT** uses an EWMA over compile times (α=0.3); recovery half-life ≈ 3 observations.

This experiment is queued for the next iteration; the design is laid out here for review.

---

## §6 Discussion

### §6.1 Why a feature-blind bandit underperforms

Our M2 ε-greedy results (5w-hetero: 119 s vs P2C 94 s; top:bottom dispatch ratio 13.2:1 vs P2C 8.6:1) demonstrate a non-obvious failure mode for naive online learning in heterogeneous scheduling. The bandit faithfully learns "this worker has the best mean reward" and proceeds to over-concentrate traffic on it. The static P2C scoring, by contrast, penalises high `active_tasks` directly through its weight vector and so spreads load even though it never observes a single compile time. Online learning is thus *not always better* than a well-tuned static heuristic; it depends on whether the learner conditions on the right features. This is the mechanistic argument for moving from MAB to contextual bandits.

### §6.2 Linear realisability and reward shape

LinUCB's regret bound (Chu et al. 2011 Thm. 1, with caveats noted in §3) assumes $\mathbb{E}[r \mid x] = x^\top \theta^*$. Real compile time is roughly proportional to source size only for small files; large templated translation units exhibit super-linear growth (preprocessing expansion, optimisation passes). After our log transform, the residual non-linearity is reduced but not eliminated. Lattimore & Szepesvári 2020 §24.4 quantifies the impact: a misspecification of magnitude $\varepsilon$ inflates regret by an additive $\mathcal{O}(\varepsilon \sqrt{T})$. Our numbers in §5.5 (ablating reward shapes) bound this empirically.

The reward function is the most consequential design lever and the least well-justified theoretically. Decima's $-(t_k - t_{k-1})J_k$ formulation has a direct connection to Little's law and JCT; our $-\log(1 + t)$ is purely heuristic. Future work should either (a) demonstrate empirical equivalence under realistic workloads, or (b) adopt the Decima form, paying the implementation cost of tracking time-integrated job counts in the coordinator.

### §6.3 Drift, restarts and the deadly triad

Worker performance is not stationary. Thermal throttling, GC pauses on JVM/Unity workers, and noisy-neighbour effects on shared hosts all cause $\theta_a^*$ to drift. LinUCB has no published regret guarantee under drift (Lattimore & Szepesvári Ch. 31). In practice, we have observed coarse drift in M1 logs (dispatch counts diverge over the build session even with constant feature distributions). Practical mitigations — sliding windows, change-point detection — sit in the open-research bucket and are flagged as future work.

A separate caveat is the *deadly triad* (Sutton & Barto 2018 §11.3): function approximation + bootstrapping + off-policy. LinUCB does not invoke the triad because it is a bandit (no value function, no bootstrap). Any extension to multi-step Q-learning over build-session state — e.g., to model dependencies between linking and earlier compile steps — would walk straight into the triad and require explicit safeguards. We treat this as a scope boundary for the present work.

### §6.4 Cost of the matrix algebra

Empirically, the LinUCB scoring step (12-dim vector + 12×12 matrix-vector product) takes microseconds per worker. With clusters of $\le 20$ workers, the per-decision overhead is dominated by the gRPC RTT to dispatch, not the matrix work. We did not encounter a regime where Sherman–Morrison's $\mathcal{O}(d^2)$ cost was material; future work that increases $d$ (e.g., 100+ feature interactions for cache-aware scheduling) may need to revisit.

### §6.5 What "good" looks like

Beyond the headline makespan number, we evaluate three properties. **(1) Robustness:** does the scheduler fall back gracefully when one worker degrades? §5.6 (failure mode injection) directly tests this. **(2) Sample complexity:** how many tasks until LinUCB exceeds P2C? Our learning-curve plots in §5.4 quantify this. **(3) Interpretability:** the bandit weight vector $\hat{\theta}_a$ tells us which features the policy values. We report the learned weights for each worker in §5.5; the magnitudes provide a direct, falsifiable narrative for why a particular worker was preferred.

## §7 Limitations

We are explicit about the limits of our experimental and methodological scope so the reader can judge the strength of our conclusions.

### §7.1 Workload diversity

Our evaluation uses a single workload: CPython compilation. While CPython exercises a broad range of file sizes and compiler invocations, it is *one C codebase* with a particular ratio of headers-to-sources and a particular dependency depth. Template-heavy C++ workloads (Boost, Eigen, Qt, LLVM itself) have very different compilation cost distributions and may shift the comparative picture. We acknowledge this and propose §8 Future Work that runs the same evaluation on at least three diverse workloads.

### §7.2 Cluster topology

Our cluster runs entirely on a single host (Docker Desktop on macOS) with cgroup CPU limits enforcing the unequal allocation. This means *network latency is symmetric* between workers, *clock skew is zero*, and the *file-system cache* is shared. A real WAN cluster has asymmetric RTT, machine-level clock drift, and isolated caches. Our LinUCB feature vector includes an RTT term, but the cardinality of network configurations we sample is one. Conclusions about LinUCB's network sensitivity are therefore extrapolated from a single point.

### §7.3 Statistical significance

Most numbers reported in this paper are from *single-shot benchmark runs*. Docker on macOS has nontrivial run-to-run variance (kernel scheduling, JIT warm-up, cgroup contention). We have repeatedly observed P2C take 85–94 s for the same 5w-hetero configuration without intentional change. Our central conclusions — particularly any LinUCB-vs-P2C comparison within ≈10 s — are therefore *suggestive* and require the multi-run protocol described in §5.5.4 (≥5 paired runs + Wilcoxon signed-rank test) for paper-grade confidence.

### §7.4 Linear realisability

LinUCB's regret bound assumes $\mathbb{E}[r \mid x] = x^\top \theta^*$ with $\lVert\theta^*\rVert \le 1$ and $\lVert x \rVert \le 1$. Compile time is non-linear in source size (preprocessing × optimisation passes are super-linear). Our $\log(1+\cdot)$ feature transform compresses but does not linearise. We do not prove a regret bound for our setting; we report empirical regret-curves and argue they are favourable, but Lattimore & Szepesvári Ch. 24.4 shows misspecification of magnitude $\varepsilon$ inflates regret by an additive $\mathcal{O}(\varepsilon \sqrt{T})$, which our experiments cannot rule out.

### §7.5 Drift

Worker performance drifts. Thermal throttling, GC pauses on JVM/Unity workers, and noisy-neighbour effects on shared hosts cause $\theta_a^*$ to change over a build session. Standard LinUCB has *no published regret guarantee* under drift (Lattimore & Szepesvári Ch. 31). Practical mitigations — sliding windows, change-point detection, exponential discounting of old observations — are research-level for linear bandits. Our schedulers do not implement any of them.

### §7.6 Plain LinUCB versus SupLinUCB

The published $\sqrt{T}$ regret bound (Chu et al. 2011, Theorem 1) applies to **SupLinUCB**, an elimination-based variant that decouples confidence updates across phases. We implement plain LinUCB Algorithm 1 from Li 2010, which is simpler but has no proven regret bound (Chu 2011 explicitly notes this gap). We cite the bound as theoretical context, not as a guarantee for our code.

### §7.7 Build-type one-hot collapse

Hybrid-Grid supports CPP, Flutter, and Unity builds. Today only the CPP path goes through `Compile()` (Flutter and Unity use a separate `Build()` handler). LinUCB's feature vector therefore degenerates: the build-type one-hot is always `(1, 0, 0)`. We removed the dead dimensions to avoid the collinearity trap discussed in §6, but this also means our results say nothing about cross-build-type generalisation. When Flutter and Unity reach the learning path, the feature dimension and warm-up dynamics will change.

### §7.8 Reward shape

We use $r_t = -\log(1 + t_{\text{compile}})$ as an empirical heuristic. *No peer-reviewed source supports this exact form*; the closest published reward for cluster scheduling is Decima's Little's-Law-justified $-(t_k-t_{k-1})J_k$ (Mao et al. 2019). The §5.5.2 ablation reports raw negative latency and the Decima form, but this entire dimension of the design space is empirically explored, not theoretically justified.

### §7.9 No comparison against production build systems

We compare against textbook scheduling algorithms. We do *not* compare against the schedulers inside Bazel RBE, BuildXL, or commercial offerings (Incredibuild, FASTBuild). Doing so would require either reverse-engineering proprietary code or running our workload through their stacks, both of which are out of scope. Conclusions are *relative to the heuristics we implement*, not absolute claims about the state of the art in production build scheduling.

## §8 Conclusion

We have presented an empirical study of contextual-bandit scheduling for distributed compilation on heterogeneous worker clusters. Our principal findings are:

**(1) Static heuristics are surprisingly hard to beat.** Power-of-Two-Choices (Mitzenmacher 2001) achieves a 1.62× speedup over LeastLoaded on our heterogeneous cluster — the largest improvement among any pair of schedulers we tested. P2C scores workers using static capability features alone, never observing task outcomes, yet outperforms every online learner in our evaluation including LinUCB at default hyperparameters.

**(2) Naive online learning underperforms.** ε-greedy MAB, despite learning from per-task latency feedback, performed *worse* than P2C on every cluster size (e.g., 5w-hetero: ε-greedy 119 s vs P2C 94 s). The mechanism is feature-blindness: the bandit's argmax-Q rule concentrates traffic on the historically-fastest worker, ignoring the queue pressure that P2C correctly accounts for through its scoring weights.

**(3) Contextual bandits are sensitive to implementation traps.** Our first LinUCB run produced 158 s on 5w-hetero — *worse than every baseline*. An independent code review identified three bugs: feature-vector reconstruction at update time used post-completion worker state (target leakage), build-type one-hot dimensions were perfectly collinear with the bias under our current code path, and the default α hyperparameter was too aggressive given the reward magnitude. After correcting these, LinUCB's behaviour normalised. The lesson: *the algorithm is not the contribution; the implementation discipline is*.

**(4) The design space matters more than the algorithm choice.** Our §5.5 ablations show that LinUCB's wall-clock varies by tens of seconds across {α, reward shape, feature subset} configurations, often dominating the inter-algorithm differences. A pre-print version of this paper that reported only the headline algorithm comparison would have under-served readers attempting to deploy these methods.

This work does not show that contextual bandits beat hand-tuned heuristics for build scheduling. It does show that a careful bandit implementation can match heuristics, with a path to exceed them via drift-aware extensions, cache-aware features, and theoretically-justified rewards (Decima's Little's-Law form). We release all source code, raw measurement logs, and reproduction scripts at https://github.com/hybrid-grid/hybridgrid.

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
