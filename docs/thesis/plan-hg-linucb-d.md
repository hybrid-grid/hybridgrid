# Plan v3: implement HG-LinUCB-D (discounted Hybrid-LinUCB), both discount clocks

Status: v3 = v2 plus the decision to implement AND run both discount modes (global clock "G" and per-arm "L"). v2 followed two independent Codex plan reviews (gpt-6-sol and gpt-5.5, both
"go-with-changes"/"no-go as written"; all BLOCKERs addressed below). Nothing is implemented yet.
Base branch `feat/sed-baseline` (tip `e114c78`), work branch `feat/hg-linucb-d`.

## 1. Goal and non-goals

Add the scheduler `hybrid-linucb-d` (HG-LinUCB-D, required by the advisor): the existing
Hybrid-LinUCB with a discount gamma on per-arm statistics so a worker whose speed changes while its quota does not can be
followed. Source: Russac, Vernade, Cappe, NeurIPS 2019 (D-LinUCB). Advisor's rule, verbatim:

    A <- gamma*A + x x^T + (1-gamma)*I ;  b <- gamma*b + r*x ;  A^{-1} recomputed (d = 12).

Non-goals of THIS task: fixing feature [8] (constant RPC latency), reward changes, other
schedulers, the experiment runner/analyzer, `hypotheses.md`, stress-ng. (Section 8 lists what
must be decided now even though it is implemented later.)

## 2. Pre-registered design decisions (cannot change after experiments start)

| # | Decision | Reason |
|---|---|---|
| D1 | **Decision clock.** `step` = the existing `totalDispatches` counter. It advances at the existing increment point, i.e. only when `SelectWithDispatchInfo` found eligible candidates (single-candidate fast path, warm-start dispatches and ResourceExhausted re-selections count; calls that return `ErrNoWorkers`/`ErrNoMatchingWorkers` do not, they are not decisions). `RecordOutcome` never advances it. Every arm ages with the clock whether or not it was pulled. | One round = one dispatch decision, as in Russac's time index, and consistent with the dispatch-clock text already sent to the advisor. Feedback arrives late, so the clock must not depend on completions. |
| D2 | **Delayed feedback is aged by decision time.** At `Select` store the dispatch step `s` of the task (new map `pendingStep`, keyed by TaskID, next to `pendingX`). When the outcome arrives at clock `t` its weight is `w = gamma^(t-s)`, clamped to `t-s >= 0`. If no step is stored (tests assign `pendingX` directly), `s = t`, weight 1. `pendingX`'s type is NOT changed (existing tests write to it). | Makes the state equal the weighted least-squares of Russac with decision-time weights, independent of completion order. |
| D3 | **Lazy exact catch-up.** Per arm store `lastStep`. Before an arm is read (`score`) or updated: `k = max(0, now - lastStep)`, `g = gamma^k`, `A <- I + g*(A-I)`, `b <- g*b`, `lastStep = max(lastStep, now)`, refresh `Ainv`, mark `theta` dirty. **`now` is computed while holding `s.mu`** as `now = max(atomic.LoadInt64(&totalDispatches), arm.lastStep)`, and that same `now` is used for the catch-up and for the weight `w = gamma^(max(now-s, 0))`, so a concurrent `Select` can never make the weight stale or non-monotone. Exact because `I` is the fixed point of `A <- gamma*A + (1-gamma)*I`. `score(workerID, x)` keeps its 2-argument signature (existing tests call it). A new arm starts at `A = I, b = 0, lastStep = now`. | O(1) bookkeeping per step; one d x d inverse per arm read after the clock moved (microseconds). |
| D4 | Update of the observed arm: catch up to `now`, then `A <- A + w*x x^T`, `b <- b + w*r*x`. Closed form (the test oracle): `A_t = I + sum_i gamma^(t-s_i) x_i x_i^T`, `b_t = sum_i gamma^(t-s_i) r_i x_i` over the observations that arrived by clock `t`, with `s_i` the decision step of observation i. With zero feedback lag this is exactly the advisor's recurrence on the decision clock. | Derivation: applying `F(A)=gamma*A+(1-gamma)I` k times gives `I+gamma^k(A-I)`. |
| D5 | Exploration bonus stays `alpha*sqrt(x^T A^{-1} x)` with the discounted A (advisor's simplification). Russac et al. score with `V^{-1} V~ V^{-1}` (V~ built with gamma^2) and a time-dependent confidence width; with the same coefficient our bonus is wider. State in code and paper: "advisor-specified discounted LinUCB variant; the bonus is heuristic and the paper's regret guarantee is not claimed". | Honest scope. |
| D6 | `gamma = 0` (unset) and `gamma = 1` mean "off": the current Sherman-Morrison path runs unchanged, bit-for-bit. Library level: values outside `(0,1]` are treated as off, never a panic. | Zero-value-is-off convention; no regression risk for `hybrid-linucb`. |
| D7 | Names: `hybrid-linucb-d` (warm-start/lambda from flags, default 100/0.5). Flag `--discount` (default 0.98, valid `(0,1]`, validated at startup only for this scheduler, ignored otherwise). **Runner label contract (implemented later):** results/task-file labels `hybrid-linucb-d-g095|g098|g099` map to `--scheduler=<base> --discount=<gamma>`; the coordinator accepts only the base name. Every experimental cell passes gamma explicitly; the 0.98 default is for ad-hoc runs. | Gamma sweeps would otherwise collide in `results.csv` and task-log filenames. |
| D8 | **Transactional update, Cholesky only.** For a known TaskID a shared helper `consumePending(taskID)` removes `pendingX` and `pendingStep` together exactly once (as the existing code already deletes `pendingX` before the update can fail). Then compute the candidate `A, b, Ainv` in temporaries and commit them with `count++`, `dirty`, `lastStep` only if every value is finite and `mat.Cholesky` succeeds on the symmetrised `A`; otherwise the sample is dropped and the arm left untouched. The invalid-reward (NaN/Inf) and unknown-TaskID early returns keep their current behaviour (no consumption), so both maps stay in step; a test asserts `len(pendingX) == len(pendingStep)`. A failed catch-up leaves the arm untouched, `lastStep` not advanced. No LU fallback (it would not preserve SPD). `Ainv` is symmetrised after inversion. | Never leave NaN/Inf in arm state; keeps behaviour of existing guards. |
| D9 | `A` (kept "for diagnostics") is updated in the discounted path. All mutation under `s.mu`; no new locks. | Consistency. |
| D10 | Log the scheduler parameters: new optional task-log field `scheduler_params` (string, `omitempty`, e.g. `alpha=0.5 warm_start=100 load_penalty=0.5 discount=0.95`), filled for the learning schedulers from `Config`. | Raw logs must be self-describing; cheap now, impossible to backfill. |
| D11 | Warm-start and load penalty are exactly as in `hybrid-linucb` (the learner updates, discounted, during warm-start; `selectWarmStart`'s `score` call also catches up). | HG-LinUCB-D differs from HG-LinUCB only by gamma. |
| D12 | **Two discount modes, both implemented and both run in the drift experiment.** Flag `--discount-mode global|arm`, default `global`. `global` (G) = D1-D4 above (every arm ages with the decision clock). `arm` (L) = the advisor's recurrence read literally: ONLY the arm that receives an accepted outcome is updated, eagerly: `A <- gamma*A + x x^T + (1-gamma)*I`, `b <- gamma*b + r*x`, then Cholesky inverse; no discount clock, no catch-up at score time, no `pendingStep`, no lag weighting (`totalDispatches` still advances as it does today; arm mode just never reads it). Closed form in `arm` mode over the arm's own n observations: `A_n = I + sum_{i=1..n} gamma^(n-i) x_i x_i^T`, `b_n = sum gamma^(n-i) r_i x_i`. Gamma off (0 or 1) is bit-identical to HG-LinUCB in BOTH modes. Why both: the advisor's text ("recompute A^{-1} at every update") is ambiguous between the two; the pilot (section 3) predicts L abandons a slowed worker fastest but may never return to it, while G returns but pays a permanent exploration tax. `scheduler_params` also records `discount_mode`. Runner labels: `hybrid-linucb-d-g098` = G, `hybrid-linucb-d-arm-g098` = L. Run plan: G at gamma 0.95/0.98/0.99, L at 0.98 only (+11% builds). Which comparison is primary is fixed in `hypotheses.md` after the advisor answers, before any run; G-vs-L is secondary. |

Behaviour of special paths (each has a test): single-candidate fast path advances the clock,
caches nothing, learns nothing; a re-selection with the same TaskID overwrites both
`pendingX` and `pendingStep` (existing overwrite semantics); an outcome without cache is dropped
without touching any arm; same worker ID after re-registration keeps its arm (the scheduler is
keyed by ID and unaware of registration) and the discount ages it.

Invariants: `A = A^T`, `A >= I`, `A*Ainv = I`, `theta = Ainv*b`, everything finite.

## 3. Pilot findings that shape the tests and the experiment (independent simulation, not the repo code)

A standalone simulation (3 arms, d = 2, alpha = 0.5, noise 0.05, 8 seeds) of undiscounted vs
discounted LinUCB under a permanent drop of arm 0's reward:

| history before drift | drop | share of arm 0 in the 150 steps after drift: gamma=1 / 0.99 / 0.98 / 0.95 |
|---|---|---|
| 120 decisions | -0.05 (realistic scale) | 0.41 / 0.33 / 0.31 / 0.32 |
| 120 | -0.50 | 0.09 / 0.07 / 0.07 / 0.11 |
| 300 | -0.50 | 0.15 / 0.08 / 0.08 / 0.11 |
| 1500 | -0.50 | **0.49 / 0.09 / 0.08 / 0.11** |
| 1500 | -0.08 | 0.75 / 0.27 / 0.27 / 0.29 |

Recovery after a 150-step slowdown (history 1500): share of arm 0 afterwards, steps +100..250:
gamma=1 0.89, 0.99 0.67, 0.98 0.55, 0.95 0.45.

Consequences (to carry into `hypotheses.md`, not fixed here): the benefit of discounting grows with
the history accumulated before the drift, so a 295-task build with drift at dispatch 120 leaves
little to forget and a long session (several builds, one coordinator) is needed; discounting costs a
permanent exploration tax (it keeps choosing the best arm only 45-65% of the time in steady state
against 97% undiscounted), so it can lose in the recovery and oscillating variants; gamma = 0.95
over-forgets. The reward spread across workers is about 0.07 (log-normalised) and a 45% slowdown is
about 0.05, so signal-to-noise is low in the real experiment.

## 4. Files

Modify: `internal/coordinator/scheduler/linucb.go` (config `Discount`, fields `discount`,
`pendingStep`, arm `lastStep`, catch-up helper, discounted branch in `RecordOutcome`, catch-up in
`score`); `internal/coordinator/server/grpc.go` (`Config.DiscountValue`, factory cases,
`scheduler_params` fill, "Valid:" comment); `internal/coordinator/server/task_log.go` (+ its test)
for D10; `internal/coordinator/server/scheduler_factory_test.go`; `cmd/hg-coord/main.go` (valid
names, error text, help, `--discount`, validation); `test/stress/benchmark-heterogeneous.sh` and
`scripts/benchmark_statistical.sh` (allow-lists, `DISCOUNT` and `DISCOUNT_MODE` env -> `--discount`, `--discount-mode`). `cmd/hg-coord/main.go` and `grpc.go` also carry `--discount-mode` / `Config.DiscountModeValue`. Keep that sentence consistent with D12. 
Create: `internal/coordinator/scheduler/linucb_discount_test.go` (worker);
`internal/coordinator/scheduler/linucb_trace_test.go` (trace driver `runTrace` + env-guarded generator)
and `internal/coordinator/scheduler/testdata/hybrid_linucb_golden.json` (both written by me in P0 and
committed BEFORE any production change, generated from the UNMODIFIED code).
Do not touch: `docs/thesis/*`, other schedulers, `gen/`, existing test files except the factory test
and the task-log test.

## 4b. API contract (names the tests rely on; P2 must implement exactly these)

```go
// linucb.go
type DiscountMode string
const (
	DiscountModeGlobal DiscountMode = "global" // G: every arm ages with the decision clock (default, also for "")
	DiscountModeArm    DiscountMode = "arm"    // L: only the arm that receives an outcome is updated
)

// LinUCBConfig gains:
//   Discount     float64      // gamma; 0 or 1 (or outside (0,1]) = off
//   DiscountMode DiscountMode // "" means global

// LinUCBScheduler gains fields: discount float64; discountMode DiscountMode;
//   pendingStep map[string]int64 // TaskID -> dispatch step (global mode only)
// linUCBArm gains: lastStep int64 // global mode: clock value A and b are aged to
// Existing fields keep their names and types: s.arms[id].{A,Ainv,b,theta,dirty,count},
// s.pendingX (map[string]*mat.VecDense), s.totalDispatches (the decision clock), s.mu, s.dim.

// Pure helpers (no receiver, never mutate their inputs):
// decay returns I + g*(A-I) and g*b for g = gamma^k (k >= 0).
func discountDecay(A *mat.Dense, b *mat.VecDense, gamma float64, k int64) (*mat.Dense, *mat.VecDense)
// commit returns A + w*x*x^T, b + w*r*x and the Cholesky inverse of the new A (symmetrised);
// ok is false when anything is non-finite or Cholesky fails, and then the other results must be ignored.
func discountUpdate(A *mat.Dense, b *mat.VecDense, x *mat.VecDense, reward, weight float64) (Anew *mat.Dense, bnew *mat.VecDense, Ainv *mat.Dense, ok bool)

// Scheduler-level observable behaviour:
//   global mode: RecordOutcome catches the arm up to now = max(totalDispatches, arm.lastStep), then adds
//     w*x*x^T with w = gamma^(max(now-s,0)), s from pendingStep (s = now if absent); score() catches up too.
//   arm mode: RecordOutcome applies A <- g*A + x*x^T + (1-g)*I, b <- g*b + r*x on that arm only; pendingStep
//     stays empty; lastStep is never used.
```

Tests may also use the P0 helpers in `linucb_trace_test.go`: `newTraceCluster`, `newTraceScheduler`,
`runTrace`, `sameTrace`, `compareTraces`, `traceResult`. Scope split: P1 (scheduler package) covers T1-T10,
T12-T15, T16 (scheduler-level part) and T17; the factory/CLI/task-log parts of T11 and T16 belong to P3.

## 5. Tests (stdlib `testing`, white-box, deterministic seeds, relative tolerances)

| ID | Test | Oracle |
|---|---|---|
| T1 | Golden: for gamma 0 and 1, in the modes "", global and arm, the 400-step P0 trace gives selections, exploration flags, Q values, dispatch count and `A, Ainv, b, count` identical to `testdata/hybrid_linucb_golden.json` produced by the unmodified code (bit-exact on amd64, relative 1e-12 on other architectures) AND bit-identical (`math.Float64bits`) to a scheduler built with Discount unset, on the same machine | exact bits |
| T3b | `max(now, lastStep)` rule: if `arm.lastStep` is ahead of the clock, `score` and `RecordOutcome` leave `lastStep` and the state unchanged (never move backward) | exact |
| T3c | Clock read inside the lock: with the arm's update pending and `s.mu` held by the test, a concurrent fast-path `Select` on another worker advances the clock; after release the update must use weight `gamma^(clock-s)`, not a stale one | exact |
| T2 | Closed form with lag: observations dispatched at steps `s_i` and completed at later steps (including reversed completion order) give `A, b` equal to D4's sums | max-norm error <= 1e-10 * max(1, ||A||) |
| T3 | Eager reference inside the test (advisor recurrence applied per decision step, observation added with weight `gamma^(t-s)` at arrival) equals the scheduler state | same |
| T4 | Idle arm: after k extra dispatches the idle arm reads `I + gamma^k (A-I)`, `gamma^k b`; reading twice without a new dispatch changes nothing; the clock advanced by RecordOutcome-only is impossible (clock unchanged) | same |
| T5 | Invariants after 2000 random steps for gamma in {0.5, 0.95, 0.99}: symmetric, eigenvalues >= 1 - 1e-9, `||A*Ainv - I||_max <= 1e-9 * max(1,||A||)`, `theta = Ainv*b`, finite | as stated |
| T6 | Guards: pending maps stay in step (`len(pendingX) == len(pendingStep)` after every operation of a Select/RecordOutcome-only trace); NaN/Inf reward dropped; unknown TaskID dropped; neither changes any arm or `count`; `RecordOutcome` never changes the clock; pure function `discountedUpdate` with NaN/Inf inputs returns not-ok and mutates nothing (transactional); huge k underflows `gamma^k` to 0 giving `A = I, b = 0` | exact |
| T7 | Fast path advances the clock, creates no `pendingX`/`pendingStep`; its later outcome is dropped; re-selection with the same TaskID overwrites both caches and the outcome uses the latest step | exact |
| T8 | Warm-start window: dispatch routed to least-loaded still caches x and step, still learns (discounted); `selectWarmStart`'s score catches up | exact |
| T9 | Race: 8 goroutines, 5 arms, **unique TaskIDs**, mix of Select/RecordOutcome; assert `count` sum equals the number of accepted outcomes and the clock equals the number of Select calls; run with `-race` in a Linux container | counts exact |
| T10 | Constructor/library: gamma outside `(0,1]` is off, no panic | exact |
| T11 | Factory: `hybrid-linucb-d` returns `*LinUCBScheduler`; default gamma 0.98; explicit value honoured; `hybrid-linucb` is off; `scheduler_params` string | exact |
| T12 | Drift adaptation, **predeclared, noise-free (deterministic)**: 3 arms, reward `m_a - 0.1*size_norm`, `m = (-0.30, -0.35, -0.40)`, arm 0 drops by 0.5 after decision 1500, alpha 0.5, warm-start 0, lambda 0, only `size` varies in the context, seeds 101..105 only change the size sequence, window = decisions 1501..1650. Assert for EVERY seed: share(gamma=1) >= 0.30 and share(gamma=0.95) <= share(gamma=1) - 0.25. (Noise-free pilot: 0.44 vs 0.11, every seed.) | thresholds fixed here |
| T13 | Recovery, **predeclared, noise-free**: same model, arm 0 slow for decisions 1501..1650 then recovers; for gamma = 0.98 the share of arm 0 in decisions 1751..1900 is >= 0.30 for every seed (pilot 0.54). Kills "decay only the observed arm" | thresholds fixed here |
| T14 | `arm` mode closed form: after n accepted outcomes for one arm `A, b` equal D12's sums (max-norm error <= 1e-10 * max(1,||A||)); an arm that receives no outcomes never changes (A = I, b = 0 after 1000 dispatches); eager reference equals scheduler state; invariants as T5 | as stated |
| T15 | `arm` mode guards: clock and `pendingStep` unused (map stays empty), NaN/Inf reward and unknown TaskID dropped without touching any arm, transactional update (non-finite candidate leaves the arm untouched), race test with unique TaskIDs (T9 analogue) | exact |
| T16 | Mode plumbing: `global` is the default; `arm` and `global` give different states on the same trace when gamma < 1 and identical states (bit-identical to the P0 golden) when gamma is 0 or 1; factory/CLI accept `--discount-mode`, reject other values for `hybrid-linucb-d` only; `scheduler_params` shows the mode | exact |
| T17 | `arm`-mode predeclared pilot behaviours, noise-free (pilot gamma 0.98: 0.05 / 0.03): permanent drift share(1501..1650) <= 0.15 for every seed, and gamma-off share >= 0.30. Recovery is REPORT-ONLY (the pilot says it may fail by design) | thresholds fixed here |

If the real code path misses T12/T13 thresholds the worker reports the numbers and stops; it does
not change them. I re-run the independent pilot harness against the real code to judge.

Mutants that must each fail at least one test (run by me): [`arm` mode] decay the arm at score time, decay all arms, update without the `(1-gamma)I` term, use the global clock; [both modes] mode flag ignored; [`global` mode] decay only the observed arm; omit the
`(1-gamma)I` term (A decays to 0); decay `b` but not `A` and vice versa; add `x x^T` before decay;
weight `1` instead of `gamma^(t-s)`; completion-time clock; `gamma` replaced by `1-gamma`; skip
catch-up in `score`; skip catch-up in `selectWarmStart`; Sherman-Morrison kept in discounted mode;
discounted path active when gamma = 1; clock advanced by RecordOutcome; non-transactional commit
on NaN.

## 6. Phases and gates (I run every gate; the worker never approves itself)

Worker: Codex CLI `gpt-6-sol`, `model_reasoning_effort=high`, `-s workspace-write`, in a dedicated
git worktree (fixed path) on branch `feat/hg-linucb-d`; no git commit/push/reset by the worker.
Reviewer: `gpt-5.5`, read-only (different model on purpose). `-race` needs cgo: I run it in
`golang:1.25-bookworm`. The worker cannot use the network: I run `go mod download` first.

| Phase | Work | Gate |
|---|---|---|
| P0 | I create the worktree/branch, `go mod download`, and prove the sandboxed worker can run `go vet` and `go test -run NONE` there. I also write the trace driver and generate the golden file from the unmodified code and commit both on the branch (the golden must exist before any production change; a test file that does not compile cannot generate it) | Environment proven (`GOFLAGS=-buildvcs=false`, worktree `.git` file works under the sandbox); the trace driver wraps `InMemoryRegistry` in a thin wrapper whose `ListByCapability` returns workers sorted by ID (the real registry iterates a map, so selections would otherwise vary); golden stores `math.Float64bits` as hex strings; golden regenerates identically twice; existing tests green |
| P1 tests | Worker writes `linucb_discount_test.go` only (T1 replays the P0 trace driver against the P0 golden). No production stubs. The test file is expected NOT to compile (only `undefined:` errors about the new fields/functions) | Gate: `go vet` fails only with `undefined:` errors; no stray temp files; reviewer (`gpt-5.5`) attacks the tests with wrong implementations; I apply accepted fixes |
| P2 implement | Worker implements section 4 for `linucb.go`; all of T1-T11 pass | vet, lint, existing tests unchanged and green, T1 bit-identical, my mutation run kills every mutant, then T12/T13 evaluated and reported |
| P3 wiring | `grpc.go`, `task_log.go`+test, factory test, `main.go`, scripts | binary starts with `--scheduler=hybrid-linucb-d --discount=0.95`; rejects `--discount=0` and `1.5` for `hybrid-linucb-d` only; `scheduler_params` appears in the log |
| P4 verification | me | race tests in Linux container, full tests of touched packages, lint, `gpt-5.5` review of the full diff, Docker smoke (5-worker cluster, `hybrid-linucb-d`, small build, non-zero arm counts, `scheduler` and `scheduler_params` in the task log), micro-benchmark of Select cost |

Rollback: everything stays on `feat/hg-linucb-d` until all gates pass.

## 7. Acceptance criteria (all required)

1. gofmt/vet/golangci-lint clean on touched packages.
2. Pre-existing tests unchanged and green; only the factory test and task-log test gained cases.
3. T1: gamma-off behaviour is bit-identical to the pre-change code.
4. Every mutant above is killed by a failing assertion (not a compile error).
5. `-race` clean for scheduler and server packages.
6. CLI and `scheduler_params` behave as in P3.
7. Reviewer report has no unresolved correctness finding.
8. Worker report lists files changed, confirms D1-D11 as implemented, T12/T13 numbers, deviations.

## 8. Decided now, implemented later (listed so they are not forgotten)

- Runner: label contract of D7; per-cell gamma recorded in `results.csv`, filenames and logs;
  a coordinator that stays up across a long session; stress-ng in the image, event timestamps,
  `cpu.stat` evidence; analyzer must treat `hybrid-linucb-d*` as treatments, not baselines, and
  pre-register the multiplicity family.
- Feature [8] (constant RPC latency) is deliberately LEFT AS IS for this round (decision log, section 9, item 4): report it as a known limitation in the paper.
- Report in the paper: dispatch-clock/decision-time weighting vs the advisor's literal recurrence,
  the heuristic bonus, the exploration tax, and the pilot table of section 3.

## 9. Decision log and open decisions

Decided (with the user):

1. Both discount clocks are implemented and both are run: G (`global`, primary) and L (`arm`, secondary,
   gamma 0.98 only). No question to the advisor is needed to proceed; if he prefers L as primary only the
   order of the comparisons in `hypotheses.md` changes.
2. `linucb-d` (plain discounted LinUCB) is OUT of this plan: it is not an advisor requirement.
   `scheduler_params` in the task log is IN.
3. The required fourth scheduler is HG-LinUCB-D (`hybrid-linucb-d`); G and L are its two variants.
4. **Feature [8] (recent RPC latency) stays as is (option C).** Today it is the constant 1.0 for every worker
   because the scheduler's own `LatencyTracker` is never fed (`linucb.go:605`, default 100 ms / 100). Reasons
   to leave it: (a) wiring real data would change HG-LinUCB's behaviour, and could hand the plain bandit a
   drift signal, confounding exactly the drift experiment; (b) the call latency includes compile time and
   saturates the `/100` cap, so a naive fix would still be near-constant; (c) all results so far were
   produced with it constant, so leaving it keeps experiments comparable with them; (d) the user's
   deployment is a LAN where latency between machines is small and stable, so the feature would carry
   little information anyway (strictly this reasoning applies to the physical-cluster experiment; the
   Docker experiments rely on (a)-(c)). Consequences: the P0 golden trace is generated once and never
   regenerated; the paper states this as a limitation (one of 12 context dimensions is constant); fixing
   it properly is a separate, later ablation that must be compared against this round.

5. **Lag weighting of D2 is KEPT** (weight `gamma^(t-s)` for late feedback, with `pendingStep`; decided with the user: the effect is small at -j5 but it is the faithful decision-time weighting). T2/T3/T7 keep their lag cases.

Still open: nothing blocks P0 or P1.

## 10. Outcome of the implementation (after P0-P4) and caveats to carry into `hypotheses.md`

Implemented on branch `feat/hg-linucb-d` (commits `1fcbec7`, `9487e58`, `159ff61`, `4d19cff` and the follow-up fix):
both discount modes, `--discount`, `--discount-mode`, `scheduler_params` in the task log. Verified: golden trace
bit-identical for discount off, 31 hand-made mutants each killed by an assertion, `-race` clean on Linux/Go 1.25,
Docker smoke (5 workers, 121/121 tasks, both modes), per-dispatch cost (task + outcome) 47 us off, 126 us global,
160 us arm.

Caveats found AFTER the plan was written (must appear in `hypotheses.md` and the paper):

1. **The hand-coded load penalty can mask the discount.** In the stationary smoke runs (CPython-like synthetic
   workload, -j7, 5 workers, warm-start 30, lambda 0.5) the per-worker task counts of `hybrid-linucb` and
   `hybrid-linucb-d` (gamma 0.95, global and arm) were nearly the same (15/18/24/29/35 vs 15/19/24/29/34 vs
   15/19/24/29/34): with lambda = 0.5 the term `lambda*LoadRatio` dominates the learned part, so a worker's
   learned speed hardly changes where tasks go. With lambda = 0 the runs differ (e.g. 16/19/26/30/30 for
   `hybrid-linucb`, 18/21/25/27/30 and 15/19/25/31/31 for D). Consequence: the drift experiment must report (and
   probably ablate) lambda; the advisor's lambda sweep {0, 0.25, 0.5, 1.0} is therefore not optional for
   interpreting D. A null result for D at lambda = 0.5 would not mean the discount does not work.
2. `pendingStep` has the same unbounded lifetime as `pendingX` when an outcome never arrives (global mode doubles
   the stale-entry footprint). Not a new class of leak; a TTL tied to the task timeout is a later improvement.
3. `scheduler_params` now also appears for `linucb` and `hybrid-linucb` (observability only; no scheduling change).
   For a requested but ineffective gamma (1, outside (0,1], NaN) the log records the request and
   `discount_effective=off`.
4. The benchmark script passes `--discount`/`--discount-mode` only for `hybrid-linucb-d`, so the command lines of
   all other schedulers are unchanged.
5. Everything in sections 3 (pilot: exploration tax, history length, over-forgetting of gamma = 0.95, low
   signal-to-noise) still applies.
