I reviewed all requested files read-only. No files, containers, benchmarks, or Git state were changed.

References below: [V2](</D:/Đồ án tốt nghiệp/hybridgrid/docs/thesis/ke-hoach-thi-nghiem-2.md:1>), [preregistration](</D:/Đồ án tốt nghiệp/hybridgrid/docs/thesis/hypotheses.md:158>), [campaign](D:/bench-work/campaign.sh:23), [runner](D:/bench-work/hybridgrid/scripts/bench_exp.sh:198), [analyzer](D:/bench-work/hybridgrid/scripts/analyze_exp.py:584), [supervisor](D:/bench-work/supervisor.sh:14), [supervisor prompt](D:/bench-work/supervisor_prompt.md:1).

## A. Round-1 findings 1–24

1. **RESOLVED** — V2:54–60 removes confirmatory support/partial-support classifications and reports each registered component separately.
2. **PARTIAL** — T5 recognizes the orphan-runner problem, but specifies neither recorded process identities nor a race-free, bounded Windows/MSYS shutdown.
3. **PARTIAL** — T1/T2 add persistence and archives, but omit several overwritten artifacts, atomic accounting, retry-exhaustion tests, and manual stops are explicitly excluded from the cap.
4. **RESOLVED** — V2:178–181 requires `n=20`/NaN/event-structure audits and says reduced `n` or “not evaluable” must be reported without new exclusions.
5. **PARTIAL** — T3 correctly proposes content-based RUN/SKIP enforcement, but it remains unimplemented and its parsing/validation contract is underspecified.
6. **RESOLVED** — S2 is an official-configuration transient smoke that checks one non-cleanup pair and non-NaN recovery share.
7. **RESOLVED** — V2 now uses `--strict`, `[101,160]`, and scenario-specific terminal-event expectations.
8. **PARTIAL** — Controller, gating, interruption, and existing tests were added, but T1/T2 persistence/exhaustion are untested and S5 cannot run as described with the current controller.
9. **PARTIAL** — T6 and its stale-line regression test are specified, but the current supervisor still scans the entire historical log at supervisor:14.
10. **PARTIAL** — Notification and stopping are proposed, but automatic stopping on probabilistic Codex output is itself unsafe; current scripts still provide neither.
11. **PARTIAL** — Base estimates now show ±15%, but the 32 h/46 h reservations do not cover both the stated upper bound and maximum registered block reruns.
12. **RESOLVED** — V2:165–168 blinds scheduler comparisons until all three confirmatory scenarios finish, with lambda decided beforehand.
13. **RESOLVED** — T7 treats boundary-spanning phases as normal, while S3 checks raw timestamps for actual gap transitions.
14. **PARTIAL** — The new appendix is explicitly exploratory and substantially prespecified, but zero/nonpositive ratio denominators and precise cell-level aggregation remain undefined.
15. **PARTIAL** — An E1 manifest and T4 are proposed, but there is no backup, no per-scenario E2 raw hash, and T4’s directory-existence test is unsafe.
16. **RESOLVED** — V2:183–189 uses the analyzer’s exact field names.
17. **RESOLVED** — V2:69 correctly counts five measured plus five warm-up cells.
18. **RESOLVED** — V2:169 states approximately 18–20 audits plus eight for lambda.
19. **RESOLVED** — V2:123–124 supplies a quoted, executable `/d/...` Git Bash command.
20. **RESOLVED** — V2:201–203 reports censoring, unequal caps, ties, approximate Wilcoxon inference, and actual paired `n`.
21. **RESOLVED** — V2:204–205 identifies H2.3’s 10-test family and prohibits post-outcome narrowing.
22. **PARTIAL** — Phase durations, task counts, and boundary phases are promised, but reporting observed on/off event dispatch counts is still not explicit.
23. **RESOLVED** — V2:208–209 correctly limits H2.4 to descriptive medians and labels the 16 baseline comparisons non-trend analyses.
24. **PARTIAL** — Fixed scenario order, host/workload, and two-build protocol are disclosed, but reuse of the same seeded arm-order mechanism is omitted.

## B. New proposals

| Proposal | Assessment |
|---|---|
| **T1** | Directionally consistent with preregistration §3, provided it counts only the registered runner-abort reruns and hard-locks the cap at two. A mutable scalar counter and the still-overridable `MAX_RESUMES` create undercount/corruption and override risks. |
| **T2** | Preserving history is appropriate, but four files are insufficient: the runner also overwrites per-round task, event, stats, build-log, status, and temporary artifacts. The archive timestamp must avoid `:` on Windows. |
| **T3** | Correctly enforces the one-time §2b decision. It must parse one exact anchored `E2_LAMBDA=RUN|SKIP` entry, reject missing/duplicate/malformed entries, and never `source` the file. |
| **T4** | **Unsafe.** The analyzer creates `analysis/` before all outputs are written and before final return; a killed analysis can therefore leave a plausible but incomplete directory. E1 also requires both `analysis/` and `analysis_sed_vs_ll/`. |
| **T5** | Safe stopping is possible in Git Bash, but not as specified. `nohup` does not form a new process group, the launch records no PID, killing campaign Bash does not kill its child, MSYS signals do not reliably terminate native `docker.exe` descendants, and TERM cleanup can be delayed while Bash waits on a foreground native command. |
| **T6** | **Do not auto-stop on a Codex ALERT.** The prompt considers makespan outliers and time trends; allowing model output to interrupt later blocks creates an unregistered, potentially outcome-adjacent stopping mechanism. Manual stops are also excluded from T1, allowing unlimited model-triggered reruns. |
| **T7** | Mostly sound as diagnostic guidance. However, exact 590 tasks and exactly one transient pair are not registered exclusion criteria in §3; they must not become ALERT-triggered rerun rules. |

For Windows shutdown, the controlling mechanism must use recorded campaign/runner identities plus a stop flag preventing campaign restart, a bounded wait, and explicit `docker compose -p hgexp -f <exact-file> down`. Verification should use the Compose project label, not merely container-name matching. Blind process-tree killing is unsafe because MSYS and Windows-native PIDs/descendants are not interchangeable.

The smoke volume is proportionate, not excessive: four configuration cells plus one interruption/resume case is reasonable before a roughly 40-hour run. It is nevertheless insufficient operationally because it does not test T1/T2 persistence, third-abort exhaustion, archive completeness, or an ALERT path. S5 is also not executable as written: campaign:29 hard-codes `REPS=20`, and campaign:33–50 hard-codes scenario arms, conflicting with “`REPS=2`, one arm.”

The base timing arithmetic is reasonable. The reservation is not. Using V2’s own upper bounds and 40-minute block allowance:

- Three scenarios: `30 h + 1.5 h preparation + 6×40 min ≈ 35.5 h`, not 32 h.
- Including lambda: `44 h + 1.5 h + 8×40 min ≈ 50.8 h`, not 46 h.

Warm-up/setup overhead makes approximately 36 h and 51 h safer reservations.

## C. New issues

1. **BLOCKER — Codex-controlled auto-stop is an unregistered stopping rule.**  
   Evidence: V2:96,169–172; the prompt examines outliers/trends at supervisor-prompt:8, while preregistration §2b/§3 only defines completion and failure-driven reruns.  
   **Plan fix:** ALERT sends a notification and evidence only. Stop automatically only for deterministic, explicitly enumerated integrity/safety conditions independent of treatment results.

2. **BLOCKER — T5 cannot reliably identify or quiesce the process tree.**  
   Evidence: V2:154 launches with `nohup ... &` without recording `$!`; campaign:147–149 synchronously starts the runner; runner cleanup occurs only through its traps at runner:240–242.  
   **Plan fix:** add tested PID/start-identity files, a stop-request flag that suppresses campaign retry, bounded waits, explicit Compose teardown, and verification of both MSYS processes and Compose-project resources.

3. **MAJOR — T1 is not a durable audit ledger.**  
   Evidence: current `MAX_RESUMES` remains operator-overridable at campaign:28; V2:91 only proposes a scalar file.  
   **Plan fix:** hard-lock the official cap at two and atomically append cause, block, timestamp, old/new count, and runner exit code; test restart persistence and third-abort refusal.

4. **MAJOR — T2 loses aborted-block raw evidence.**  
   Evidence: T2 archives only four files, while runner:452–457 and 466 overwrite events, stats, build logs, and markers for the same arm/round.  
   **Plan fix:** archive every artifact belonging to the affected round before rerun and provide a consolidated failure/rerun inventory for reporting.

5. **MAJOR — T4 can accept partial or stale analysis.**  
   Evidence: analyzer:609 creates the directory before writing its three files at 611–614; campaign:99–112 requires additional scenario-specific analyses.  
   **Plan fix:** skip only after validating all required files, exact analyzer command/configuration, expected cell/pair/family counts, successful logs, and a completion manifest written last.

6. **MAJOR — The proposed orchestration test does not exercise the new retry machinery.**  
   Evidence: V2:103–107 tests T3/T4/T6; S5 tests a manual stop, not two automatic runner aborts, persistence across invocation, archive behavior, or exhaustion.  
   **Plan fix:** add a disposable controller fixture with deterministic fake runner exit codes and test round-1 truncation protection, persistent counts, two retries, and third-abort stop.

7. **MAJOR — S5 conflicts with the production controller interface.**  
   Evidence: V2:120 says `REPS=2`, one arm; campaign hard-codes 20 repetitions and scenario arm sets.  
   **Plan fix:** use a dedicated test fixture/controller copy; do not make official `REPS` or arm lists externally mutable merely to support the test.

8. **MAJOR — Supervisor termination does not cover the new manual-stop marker.**  
   Evidence: T5 writes `MANUAL STOP` at V2:95, but T6 retains “all four” current terminal patterns; supervisor:14 has no `MANUAL STOP`.  
   **Plan fix:** recognize the post-start manual-stop marker, prevent duplicate supervisor instances, and test that exactly one final audit occurs.

9. **MAJOR — T7 risks adding exclusion criteria.**  
   Evidence: preregistration:186–192 uses warm-up-relative task count and at least 100 post-onset completions, not exact 590 or exact transient structure.  
   **Plan fix:** explicitly classify these as diagnostics only; deviations may be reported but must not alone exclude/rerun a cell or trigger an automatic stop.

10. **MAJOR — Time contingency remains understated.**  
    Evidence: V2:67–79 reserves 32/46 h despite upper bounds of 30/44 h plus preparation and up to six/eight block reruns.  
    **Plan fix:** reserve approximately 36 h for the three scenarios or 51 h including lambda, or clearly separate base, uncertainty, and retry-contingency budgets.

11. **MAJOR — T3’s file parsing is underspecified.**  
    Evidence: V2:93 says only “read the decision”; the approval file contains additional free-form text at 149–151.  
    **Plan fix:** accept exactly one anchored decision record, reject ambiguity before any cell starts, log the parsed decision, and treat the file as data rather than shell code.

12. **MINOR — The new manipulation ratio lacks invalid-denominator handling.**  
    Evidence: V2:190–196 defines `compile_time_ms / preprocessed_size_bytes` without specifying zero, negative, or missing sizes.  
    **Plan fix:** prespecify validity checks and the exact cell-level aggregation; report excluded-task counts in this exploratory appendix.

**Verdict: NOT READY**