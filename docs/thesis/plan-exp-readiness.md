# Plan: experiment readiness for the advisor's Experiments 1 and 2

Status: v2 after Codex plan review (lead: Claude, worker: Codex gpt-6-sol, reviewer: Codex gpt-5.5).
Branch `feat/exp-readiness` (on top of `feat/hg-linucb-d` @ c48f0f2).

Goal: after this plan, Experiments 1 (intermediate load) and 2 (hidden drift)
of the advisor's guide can be run as 20 randomised blocks with 4+ schedulers,
with raw JSONL logs, event timestamps in the task log, and one analysis
command that prints the Table-3 style result (median, paired delta, Holm p,
matched-pairs rank-biserial). Nothing in this plan runs the confirmatory
experiments; `hypotheses.md` is written after the pilot (section 8) and
committed before any confirmatory run.

## 0. What the advisor requires (source of truth)

General rules: >= 4 schedulers (LeastLoaded, SED, HG-LinUCB, HG-LinUCB-D);
randomised blocks, 20 blocks; `hypotheses.md` written before running and never
edited after results; every injected event after dispatch 100; timestamp of
every event in the task log; 1-2 "real-world analogue" sentences per scenario;
result table: median, delta, Holm p, paired matched-pairs rank-biserial.

Exp 1: CPython at -j6, -j7, -j8, -j9; plus one scenario with 2 concurrent
builds from 2 clients. Metrics: makespan; skip-idle rate (paper Table 5
definition: a decision sent to a busy worker while at least one idle worker
existed); rate of decisions that chose the strongest idle worker.

Exp 2: at about dispatch 120 run `stress-ng --cpu 1` INSIDE the container of
the 1.1-CPU worker (never `docker update --cpus`). Variants: permanent; 30 s
then recover; on/off oscillation. HG-LinUCB-D with gamma in {0.95, 0.98,
0.99}; lambda sweep {0, 0.25, 0.5, 1.0}. Metrics: makespan; share of tasks
sent to the slowed worker before/after the event; decisions needed for that
share to drop (time-to-adapt); which worker ran the last task to finish.

## 1. Problems found in the exploratory prep (REPORT.md, Codex review)

R1 `benchmark_rigorous.sh`: `set -e` without `pipefail`; the build runs as
`... | tee`, so a failing `docker compose exec`/`hgbuild` can be recorded as a
valid sample (only `make ***` lines are grepped).
R2 registration barrier only warns when fewer than 5 workers register.
R3 no completeness checks (task count, failures, complete blocks).
R4 task log has no dispatch timestamp / dispatch sequence; analyzers infer
dispatch as `ts - total_duration_ms`, which includes coordinator queueing.
R5 worker IDs are random per run; no stable worker identity in the log
(address `worker-5:50051` is not logged).
R6 skip-idle is approximated by `worker_active_tasks_at_dispatch > 0` or by
interval reconstruction; "strongest idle worker" cannot be computed exactly
without a snapshot of all candidates at decision time.
R7 `internal/cli/build` (the path used by `hgbuild make` / `hgbuild cc`) sends
no BuildId, so two concurrent builds cannot be separated in the task log.
R8 stress-ng is not in the image.
R9 analyzers hardcode `hybrid-linucb` as the only treatment, use Cliff's delta
(unpaired), `analyze_idle_skip.py` hardcodes -j5.
R10 the runner overwrites the tracked `test/stress/docker-compose-hetero.yml`.
R11 no way to run scheduler variants (gamma, discount mode, lambda) side by
side as separate arms of one block design.

## 2. Design decisions (lead)

D1 New runner `scripts/bench_exp.sh`; `benchmark_rigorous.sh` only gets the
R1/R2 hardening so old results stay reproducible. The new runner generates its
compose file at `test/stress/docker-compose-exp.yml` (gitignored) and never
touches the tracked hetero compose file.
D2 Arm labels. An arm is one entry of `ARMS` (space separated). Grammar:
`<base>[-arm][-g<ddd>][-l<ddd>]` where base is any scheduler accepted by
hg-coord (`leastloaded sed hybrid-linucb hybrid-linucb-d ...`), `-arm` sets
`--discount-mode=arm`, `-g095` sets `--discount=0.95` (only valid with base
`hybrid-linucb-d`), `-l025` sets `--load-penalty=0.25` (`-l0` = 0, `-l100` = 1.0;
only valid with bases that take a load penalty: hybrid-linucb, hybrid-linucb-d).
Defaults when a suffix is absent: discount 0.98 global, load penalty 0.5.
Examples: `hybrid-linucb-d-g095`, `hybrid-linucb-d-arm-g098`,
`hybrid-linucb-l0`, `hybrid-linucb-d-g098-l100`. The parser rejects unknown
suffixes and invalid combinations before anything starts. File names use the
label verbatim: `tasks_<label>_round<N>.jsonl`.
D3 Coordinator restart per cell stays the default (learner cold start each
build, as in the paper). `SESSION_BUILDS=N` (default 1) runs N builds back to
back on one coordinator session within a cell (cache cleared and `make clean`
between them, coordinator and learner NOT restarted). Makespan is recorded per
build (`build_idx`) and per session. Rationale: a 295-TU light build lasts
about 37 s and dispatch 120 happens around 15 s in, so the 30 s transient
variant may never recover inside one build. Whether Exp 2 uses N=1, a heavier
workload, or N>1 is decided from the pilot (section 8) and frozen in
`hypotheses.md`; it is also advisor question Q8.
D4 Concurrent clients: `CLIENTS=1|2`. With 2, the compose file has `builder`
and `builder2` (`BUILDER_CPUS` each, default 1.0, own source volume `cpython-src2`), both builds
start together (barrier: both `make` processes launched before timing starts),
each with its own `HG_BUILD_ID` (`<label>-r<N>-c<k>`). Recorded: per-client
elapsed and cell makespan = max of the two.
D3b With SESSION_BUILDS>1 the drift is SESSION-level: DRIFT_AT counts
dispatches of the whole session, the drift fires once per session, and the
analyzer reports drift metrics on the session's dispatch sequence (per-build
makespans are still recorded).
D5 Drift injector (Exp 2) runs in the background during the timed build:
`DRIFT=none|permanent|transient|onoff`, `DRIFT_AT=120` (dispatch count),
`DRIFT_TARGET=worker-5`, `DRIFT_CPU=1` (stress-ng workers), `DRIFT_DURATION=30`
(transient), `DRIFT_ON=10 DRIFT_OFF=10` (onoff). It polls the coordinator's
dispatch counter (D6) every 200 ms, fires when counter >= DRIFT_AT, starts
`stress-ng --cpu $DRIFT_CPU` inside the target container with `docker compose
exec -d`, verifies it is running (`pgrep stress-ng`), and stops it with
`pkill stress-ng` (transient end, each onoff off-phase, and unconditionally at
cell end, also on error/EXIT). DRIFT_AT < 101 is rejected (advisor rule 4).
D6 Coordinator changes (Go, small, all additive):
  a. Dispatch counter = number of BOOKED dispatch attempts (incremented in
     `dispatchLoop` right after IncrementTasks). An attempt later rejected by
     the worker (ResourceExhausted -> unbook + retry) keeps its number, so
     sequence gaps are expected. Exposed as Prometheus counter
     `hybridgrid_dispatch_decisions_total` and as a tiny read-only JSON
     endpoint `GET /dispatch-count` -> `{"dispatches": N}` on the ops port.
     The runner polls it from the HOST through the mapped port 18080 (no
     `docker exec` per poll).
  b. Task log fields on `task_completed` records:
     `dispatch_seq` (int64, counter value of this task's FINAL booked
     attempt; 1-based), `dispatch_ts` (RFC3339Nano UTC, time of that booking),
     `received_ts` (Compile entry), `worker_address` (WorkerInfo.Address),
     and `candidates`: a CLUSTER snapshot taken by dispatchLoop on its own
     goroutine immediately before SelectWith for the final attempt: every
     registered worker that matches the build type/arch (and the client-OS
     filter when it applies), each `{"address","active","max_parallel",
     "cpu_millis","healthy"}`. This is deliberately NOT the scheduler's
     private candidate list (schedulers build it differently, e.g. P2C vs
     LinUCB); the metrics ask whether an idle worker existed in the cluster,
     which is exactly what this snapshot records. Divergence from the
     scheduler's own read is limited to completions landing in the
     sub-microsecond gap between the two reads; documented as a limitation.
  c. Event records: a second record type in the same JSONL file,
     `{"ts", "event": "injected_event", "kind": "drift_on|drift_off|...",
     "target": "worker-5:50051", "dispatch_count": N, "detail": "..."}`.
     Written by the coordinator through a new ops endpoint
     `POST /events` on the ops HTTP port (8080), protected by the existing
     ops auth token when one is configured; opshttp receives a narrow
     interface (`EventSink{ LogInjectedEvent(...) error }` and
     `DispatchCounter{ Dispatches() int64 }`) implemented by the coordinator
     server, no other coupling; body validated (kind from a
     fixed allow-list, target and detail length-limited). The runner posts
     an event right after the action (start/stop of stress-ng) succeeds,
     with the counter value it observed. Rationale: the advisor wants event
     timestamps in the task log; going through the coordinator's own logger
     keeps one writer per file (no interleaved appends from outside).
  All existing analyzers must ignore non-`task_completed` records (audit and
  fix: they currently read every line).
D7 `internal/cli/build` sends `BuildId: taskid.BuildSessionID()` in both
CompileRequest constructions (R7).
D8 Dockerfile.base installs `stress-ng`.
D9 Completeness/validity gates in the runner, per cell: build exit code 0
(direct, `pipefail`); all 5 workers registered (else abort the cell); task log
has >= `MIN_TASKS` `task_completed` records (default: count observed in the
warm-up, minus 0) and zero `success=false`; drift cells have at least one
`drift_on` event and its `dispatch_count >= 101`. A failed cell is retried once
(fresh cluster) and logged to `failures.csv`; a second failure aborts the run
(resume with ROUND_START).
D10 Metadata per run `meta.json`: git commit + dirty flag, image IDs, host
info (nproc, docker NCPU/MemTotal), ARMS with resolved flags, JOBS, CLIENTS,
SESSION_BUILDS, BUILDER_CPUS, CONFIGURE_FLAGS, CPYTHON_VERSION, DRIFT settings, seed, and per
worker `cpu.max` and `nproc` read from inside each container at start.
D11 New analyzer `scripts/analyze_exp.py` (pandas + scipy):
  - loads results.csv + tasks_*.jsonl (+ events), validates completeness
    (complete blocks only; prints what was dropped and why; `--strict` exits
    non-zero if anything is dropped);
  - per-cell metrics: makespan (cell; also per client / per build);
    skip-idle rate = decisions with chosen.active > 0 while some candidate had
    active == 0, over all decisions (and over decisions with an idle
    candidate); strongest-idle rate = decisions that chose the candidate
    with the highest cpu_millis among candidates with active == 0 (ties by
    cpu_millis count as correct), over decisions with an idle candidate;
    drift metrics using dispatch_seq and the first drift_on event: share of
    decisions to the target in the pre window [101, onset) and post window
    [onset, end) (and per on/off phase for onoff); time-to-adapt = number of
    decisions after onset until the rolling share to the target over a
    window of W=20 decisions first falls to <= T (default T = half of the
    arm's pre-onset share in that cell; both W and the factor are CLI
    parameters; reported as censored "> n" when it never happens and as NA
    when the pre-onset share is below 0.05, because then there is nothing
    to adapt away from); for paired tests a censored value is ranked as
    larger than every observed value (n_post + 1); last-finisher = worker_address of
    the task with the latest completion ts; tasks on target after recovery
    (transient) as a secondary metric;
  - paired comparisons per metric: `--treatments` (default every arm whose
    base is hybrid-linucb or hybrid-linucb-d) vs `--baselines` (default
    leastloaded, sed); for each pair over complete blocks: median of each,
    median paired difference (treatment - baseline) with bootstrap 95% CI,
    two-sided Wilcoxon signed-rank p (exact when n <= 25 and no ties,
    otherwise normal approximation; zero differences dropped
    (zero_method='wilcox') and their count reported; if every difference is
    zero: p = 1, r = 0), Holm over the family = all treatment x baseline
    pairs of ONE metric inside ONE scenario directory (one OUT_DIR = one
    scenario: fixed -j, clients, drift variant); arms of a gamma/lambda
    sweep run in that scenario enlarge the family and this is stated in the
    output; matched-pairs rank-biserial r = (W+ - W-)/(W+ + W-) over the
    non-zero differences (sign: negative = treatment smaller);
  - outputs: `analysis/summary.md` (Table-3 style), `analysis/cells.csv`
    (one row per cell with every metric), `analysis/pairs.csv`.
  Unit tests `scripts/tests/test_analyze_exp.py` (stdlib unittest, synthetic
  fixtures): rank-biserial and Holm against hand-computed values, skip-idle /
  strongest-idle on crafted candidate lists, time-to-adapt incl. censoring,
  incomplete-block dropping, event records ignored by task metrics.
D12 Existing analyzers: only make them skip non-`task_completed` records and
remove the hardcoded -j5 line; no other changes.

## 3. Phases (each ends with a gate run by the lead)

P1 Go (Codex worker): D6 a-c, D7, D8, plus `.gitignore` for
`test/stress/docker-compose-exp.yml`. Tests: dispatch counter increments only
on successful booking; task log record contains the new fields with correct
dispatch_seq ordering under concurrency; candidates snapshot content;
`POST /events` validation (method, auth, allow-list, size limit) and that the
record lands in the task log; BuildId sent from internal/cli/build. Gate:
`gofmt -l`, `go vet ./...`, `go test ./...` on Windows, `go test -race` for
the touched packages in golang:1.25-bookworm, golden trace still bit-exact.
P2 Runner (Codex worker): `scripts/bench_exp.sh` (D1-D5, D9, D10) + R1/R2
hardening in `benchmark_rigorous.sh`. Must pass `bash -n` and shellcheck if
available; a `DRY_RUN=1` mode prints the resolved arms, generated compose file
and planned cell order without touching Docker (used for label-parser tests:
`scripts/tests/test_bench_exp_labels.sh`).
P3 Analyzer (Codex worker): D11, D12 + unit tests. Gate: tests pass; run on
the B4 smoke data (old format, must degrade gracefully: metrics needing new
fields reported as unavailable, not crash).
P4 Integration pilot (lead): one block each of
  (a) Exp 1 single client -j8, ARMS="leastloaded sed hybrid-linucb hybrid-linucb-d-g098";
  (b) Exp 1 two clients -j8, same arms;
  (c) Exp 2 permanent, transient, onoff with the same arms, DRIFT_AT=120.
Checks: events present with dispatch_count >= 101, stress-ng visible in the
target and gone after the cell, completeness gates pass, analyzer produces all
metrics, slowed-worker share changes after onset for at least LeastLoaded
(which cannot see the slowdown) - a sanity check that the drift is real.
   Also: `docker stats` samples (1 s) of builder(s) and coordinator during
   the pilot cells; if a builder or the 0.25-CPU coordinator sits at its
   quota, raise BUILDER_CPUS / coordinator CPU before confirmatory runs.
   And a DRY_RUN listing of the full confirmatory arm matrix with an
   estimated wall time (cells x blocks x measured cell time), because the
   gamma/lambda sweeps multiply the run length.
P5 Review (Codex gpt-5.5) of the whole diff + pilot outputs; fixes; then
draft `hypotheses.md` for the user/advisor to approve (not committed by the
worker). `hypotheses.md` also carries, per scenario, the 1-2 sentence
real-world analogue (advisor rule 6), the frozen arm matrix, W/T for
time-to-adapt and the Holm family definition.

## 4. Out of scope / known limitations kept

Feature [8] (RPC latency) stays constant; reward stays compile-time only; host
is a hybrid P/E-core Windows machine (final runs preferably on a homogeneous
Linux host); Experiments 3-6 not covered.

## 5. Open questions for the advisor (not blocking implementation)

Q8 long session vs per-build restart for Exp 2 (D3 supports both).
Q-idle: "worker rảnh" read as active == 0 (strict idle, paper Table 5); a
free-slot variant (active < max_parallel) is reported as secondary.

## 6. Threats to validity to report (from review)

Docker Desktop on Windows with hybrid P/E cores; stress-ng competes inside
the target container's own CFS quota (it slows that worker, it does not take
host capacity, which is the intended "hidden" drift); the 0.25-CPU
coordinator and 1.0-CPU builders may become bottlenecks (measured in P4);
the 295-TU light workload may be too short for the transient variant (D3,
pilot decides); shared Docker volumes for two clients.

## 7. Review log

v1 -> v2 (Codex gpt-5.5 plan review): candidates defined as a cluster
snapshot rather than the scheduler's private list (blocker 1, resolved by
redefining the metric input, no scheduler refactor so the golden trace is
untouched); counter = booked attempts with expected gaps (blocker 2);
session-level drift (blocker 3); analogue sentences go to hypotheses.md
(blocker 4); host polling of `/dispatch-count` instead of docker exec at 5 Hz
(major 1, also adopts the simpler JSON endpoint suggested as minor 4); narrow
EventSink/DispatchCounter interfaces for opshttp (major 2); BUILDER_CPUS and
docker stats gate (major 3); Holm family per metric per scenario (major 4);
zero-difference rule (major 5); time-to-adapt floor and censoring rank
(major 6); arm-matrix dry run with time estimate (major 7).

v2 -> v3 (lead fix round 2): healthy candidates alone count as idle or
free; single-build and single-client cells omit redundant component metrics.
Drift analysis retains the advisor's pre/post shares and adds all-pre counts,
configurable time-to-adapt reference, candidate-capacity-adjusted excess
shares, early/late post shares, and explicit treatment of cell-end cleanup
as distinct from recovery. Short drift windows are flagged in the summary
without dropping the cell. Concurrent builders now time `hgbuild make` inside
their containers behind a shared volume barrier, using the container clocks
for per-client elapsed time and cell makespan. The runner fails and retries a
drift cell if the first onset is more than 40 dispatches late or fewer than
100 completed tasks follow it.

v3 -> v4 (lead fix round 3): session drift duration and on/off timers pause
between builds while cleanup and gate setup run; build_start notes record each
gate release. HGTIME records container hostnames and cells fail when client
clocks imply a duration inconsistent with their own task receipt-to-completion
span, or when concurrent clients report one hostname. Drift phases ignore
notes as transitions, flag build boundaries, and add pooled on/off decision
shares and counts. Recovery metrics apply only to a single on/off transient
cycle. Analysis now rejects cells with missing, extra, or duplicated
build/client results before retaining complete blocks.
