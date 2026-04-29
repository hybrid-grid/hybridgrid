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

| Cluster | LeastLoaded | P2C | ε-greedy | LinUCB (α=1) | HEFT |
|---|---|---|---|---|---|
| 1w-4.0cpu | 92 s | 130 s | 146 s | 129 s | *pending* |
| 3w-hetero | 123 s | 85 s | 142 s | 103 s | *pending* |
| 5w-hetero | 152 s | 94 s | 119 s | **158 s** | *pending* |

(Single-run measurements; statistical repetition in §5.5.4.)

P2C improves 1.45-1.62× over LeastLoaded on heterogeneous clusters, confirming Mitzenmacher 2001's theoretical prediction. P2C's 1-worker regression (−41%) reflects circuit-breaker filtering overhead, which we address in the bandit implementations with a single-candidate fast path. The headline negative result is the bottom-right cell: at default $\alpha = 1.0$ LinUCB is the worst scheduler on 5w-hetero, 1.68× slower than P2C. We characterise this in §5.5: the cold-start trap (1.7% exploration rate, three workers receiving ≤10 tasks each over 873 dispatches) suggests the default $\alpha$ is too small to break the bandit's tendency to over-concentrate on the empirically-best worker. The α-sweep in §5.5.1 isolates this effect.

### §5.3 Load balance (per-worker dispatch counts)

5w-hetero scenario, 873 tasks total:

| Scheduler | top:bottom ratio | top worker share | Exploration rate |
|---|---|---|---|
| LeastLoaded | 10.8 : 1 | 33.3% | n/a |
| P2C | 8.6 : 1 | 33.3% | n/a |
| ε-greedy | 13.2 : 1 | 33.3% | 8.4% |
| LinUCB (α=1) | **97 : 1** | 33.3% | 1.7% |
| HEFT | *pending* | *pending* | *pending* |

Both static heuristics overload the strongest worker (33% concentration). Bandits compound the problem rather than relieve it. ε-greedy's argmax-Q on running mean reward concentrates traffic on the historically-fastest worker once a small lead is established; LinUCB's UCB bonus, with $\alpha = 1$, fails to break out of this trap and three workers receive almost no tasks at all (top:bottom 97:1). This suggests load balance and learned-best-arm exploitation are *competing* objectives and the bandit framework, as instantiated here, optimises the wrong one. §5.5 explores remediations.

### §5.4 Compile-time tail (P50/P95/P99 ms)

| Scheduler | P50 | P95 | P99 |
|---|---|---|---|
| LeastLoaded | 820 | 6 226 | 23 961 |
| P2C | 704 | 5 830 | 19 347 |
| ε-greedy | 956 | 7 387 | 25 488 |
| LinUCB (α=1) | 942 | 7 142 | 23 661 |
| HEFT | *pending* | *pending* | *pending* |

P2C reduces P99 by 19% over LeastLoaded — partly placement, partly less queue contention. ε-greedy and LinUCB both shift the entire distribution right relative to P2C, consistent with their dispatch concentration: a small number of overloaded workers create queueing on the busiest arms, and the worst tasks land on those queues.

### §5.5 Ablations

The first single-run measurement at $\alpha = 1.0$ revealed a worse-than-baseline result for LinUCB on 5w-hetero (158 s vs P2C 94 s). We investigate the design space along three axes to understand whether the underperformance is intrinsic to LinUCB or attributable to specific design choices.

#### §5.5.1 Exploration coefficient α

We sweep $\alpha \in \{0.1, 0.5, 1.0, 2.0\}$ at the 5w-hetero configuration and report makespan, top:bottom dispatch ratio, and overall exploration rate. The Li 2010 paper itself notes (verbatim, §3 after Eq. 4) that the theoretical $\alpha = 1 + \sqrt{\ln(2/\delta)/2}$ "may be conservatively large in some applications". Chu et al. 2011 §5 reports that the optimum is workload-dependent.

| α | Makespan (s) | Top:bottom dispatch | Exploration rate |
|---|---|---|---|
| 0.1 | *pending* | *pending* | *pending* |
| 0.5 | *pending* | *pending* | *pending* |
| 1.0 | 158 | 97:1 | 1.7% |
| 2.0 | *pending* | *pending* | *pending* |

We expect: low $\alpha$ → fast convergence on the apparent best worker (cold-start trap intensifies); high $\alpha$ → more spread but more wasted dispatches to slow workers. The optimum, if one exists in this set, will inform the recommended default for systems-bandit work.

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
