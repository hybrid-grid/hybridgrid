I verified that the measured-cell counts (140/140/140/200), session-makespan summation, and stated Holm family sizes of 4 and 10 match the code. The following issues remain.

## BLOCKER

1. **The plan invents confirmatory decision rules.** “Supported” requires both Holm significance and the directions of both median and rank-biserial effect; H2.1 also gains an unregistered “partially supported” category (`ke-hoach-thi-nghiem-2.md:48-52`). The preregistration specifies two-sided Wilcoxon, Holm, effect estimates, and directional hypotheses, but no such classification rule (`hypotheses.md:41-44,145-151`).  
   **Plan fix:** remove categorical supported/partial/not-supported decisions, or explicitly label them post-hoc. Report each registered hypothesis component with adjusted p-value, paired effect, CI, and observed direction.

2. **The manual stop procedure can leave an orphan runner writing into official data.** The campaign is launched via `nohup ... &` (`ke-hoach...:135-140`), so Ctrl-C at the prompt will not stop it. Killing only the campaign PID is unsafe: `campaign.sh` synchronously starts the runner but has no child/process-group termination logic (`campaign.sh:119-120,147-150`); cleanup runs only if the runner itself receives the signal (`bench_exp.sh:240-242`).  
   **Plan fix:** replace `Ctrl-C/kill campaign` with a tested shutdown procedure that records campaign and runner/process-group PIDs, signals the runner, waits for cleanup, verifies no `hgexp` containers or `stress-ng`, archives the interrupted block, and only then resumes.

3. **Retry accounting is neither persistent nor fully auditable.** `resumes=0` is reset on every campaign invocation (`campaign.sh:138`), so restarts can exceed the preregistered two block reruns. Worse, a resume at round 1 invokes `ROUND_START=1`, which truncates `results.csv`, `order_log.csv`, and `failures.csv` (`bench_exp.sh:198-201`), erasing failure history contrary to `hypotheses.md:194-203`. Automatic campaign retry can do this before an operator can archive anything (`campaign.sh:152-155`).  
   **Plan fix:** make this a known tool defect requiring resolution before the official run—either a pinned, documented bug-fix commit or an equally robust persistent retry ledger/archive mechanism. Add an intentional round-1 abort/resume smoke test.

4. **`--strict` does not protect confirmatory metrics from NaN or reduced sample sizes.** It fails only for structurally dropped blocks (`analyze_exp.py:479-483,616`). Pairwise tests silently remove NaNs (`analyze_exp.py:503-507`). `target_after_recovery_share` exists only with exactly one non-cleanup on/off pair (`analyze_exp.py:286-293`), yet the runner requires only onset and 100 post-onset completions (`bench_exp.sh:585-588`). Likewise, `time_to_adapt_rank` is NaN when the reference share is below 0.05 (`analyze_exp.py:180-182`). Thus H2.1/H2.2 can return `rc=0` with fewer than 20 pairs.  
   **Plan fix:** require a pre-inference audit of `n=20` for every confirmatory pair, exact transient event structure, non-NaN metrics, and expected family cardinality. If missingness still occurs without meeting a frozen exclusion criterion, report the actual reduced n or “not evaluable”; do not invent a new exclusion/rerun rule.

5. **The one-time E2-lambda decision is not enforced.** `PHASE2_APPROVED` content is ignored; only file existence is checked (`campaign.sh:132-134`). Skipping lambda depends solely on supplying `STOP_AFTER` to every invocation (`campaign.sh:117,159`). An omitted environment variable on resume silently reverses a recorded SKIP decision.  
   **Plan fix:** add a mandatory launch/resume checklist or wrapper that reads the recorded RUN/SKIP decision and rejects inconsistent `STOP_AFTER`; record the exact canonical resume command.

## MAJOR

6. **The smoke test omits the branch that produces H2.2’s confirmatory metric.** Its real runs cover permanent/onoff/permanent, not transient (`ke-hoach...:85-93`). The old transient pilot used `BUILDER_CPUS=1.0`, commit `07088e…`, not the official configuration (`out/pilot-drift/trans_s2/meta.json:8-9,17,27`).  
   **Plan fix:** add an official-configuration transient smoke and verify exactly one non-cleanup `drift_on`/`drift_off` pair plus a non-NaN `target_after_recovery_share`.

7. **Several smoke assertions are wrong or too weak.** `--strict` is useful with `REPS=1`; it does not require 20 blocks (`analyze_exp.py:594,616`). An onoff cell need not end with a `cell_end` off event if it finishes during an off phase (`bench_exp.sh:219-233,269-275`). The plan’s `[120,160]` smoke threshold also differs from the frozen/runner acceptance interval `[101,160]` (`hypotheses.md:186-192`; `bench_exp.sh:585-588`).  
   **Plan fix:** use `--strict`; make terminal-event expectations scenario-specific; distinguish the registered acceptance range from a diagnostic expectation that triggering normally occurs at or after 120.

8. **The smoke does not exercise the highest-risk orchestration paths.** It does not test campaign traversal over completed E1 directories, phase gating, persisted RUN/SKIP behavior, `SESSION_BUILDS=2` resume, retry exhaustion, supervisor termination, or safe interruption. Dry-running labels only checks parsing, not these paths (`bench_exp.sh:178-191`).  
   **Plan fix:** add disposable-output controller tests and one intentional two-build interruption/resume. Also run the existing label/analysis contract tests; label tests already assert special `-arm`, `-g`, `-l0/-l025/-l100` resolution (`scripts/tests/test_bench_exp_labels.sh:35-44`).

9. **The supervisor stale-line diagnosis is correct, but the fix is not implemented or tested.** Current `finished()` scans the entire historical log (`supervisor.sh:14`). The old `STOP_AFTER reached` therefore causes one audit, an immediate second audit without sleeping, then exit (`supervisor.sh:26-31`).  
   **Plan fix:** make the log-offset change a hard pre-launch prerequisite and regression-test it with an old terminal line plus a newly appended terminal line. Apply the offset to all four terminal patterns, not only `STOP_AFTER`.

10. **There is no actual alert delivery or automatic stop.** The supervisor only writes files and sleeps (`supervisor.sh:18-31`); its prompt explicitly prohibits stopping processes (`supervisor_prompt.md:1,11`). Therefore “on ALERT, stop immediately” is not operationally guaranteed (`ke-hoach...:157`).  
    **Plan fix:** name the human watcher and notification mechanism, maximum acknowledgement delay, and safe-stop procedure. Do not describe the supervisor itself as enforcing a stop.

11. **The timing interval is over-precise and the reserved window too short.** The 222–232 s estimate comes from one mismatched transient/leastloaded pilot, whereas the preregistration records approximately ±15% uncertainty (`hypotheses.md:163-165`). A combined 37 h expectation implies roughly 31.5–42.6 h before interruption/recovery contingency, not a firm ≤40 h.  
    **Plan fix:** reserve at least the preregistered upper range plus recovery margin and present the 8.6–9.0 h rows as point estimates, not credible bounds.

12. **The monitoring rule permits within-E2 peeking.** The plan allows scheduler comparisons once each directory finishes (`ke-hoach...:149-151`). That can influence discretionary handling of later failures, alerts, analyses, or narrative even if arms remain fixed.  
    **Plan fix:** blind treatment-effect outputs until all three confirmatory E2 scenarios finish; during execution inspect only integrity/configuration/event fields. The lambda RUN/SKIP decision must already be recorded.

13. **`drift_build_boundary_phases` is easy to misinterpret.** It reports a phase that *spans* a build-start boundary, not a phase transition that began during the inter-build gap (`analyze_exp.py:262-277`). Such entries are expected because the runner freezes the phase timer but preserves stress state during the gap (`bench_exp.sh:304-316,516-521`).  
    **Plan fix:** tell the supervisor not to warn merely because this field is nonempty. To test “no transition in the gap,” compare raw event timestamps with build-1 end/build-2 start timestamps.

14. **The manipulation check and rolling plots retain substantial analytic freedom.** Aggregation unit, pre/post windows, arm pooling, handling of zero/small files, and inferential treatment are unspecified (`ke-hoach...:170-175`).  
    **Plan fix:** either specify those choices before opening E2 outcomes or place these analyses in a separately labeled exploratory appendix with no confirmatory language.

15. **Official E1 analyses will really be overwritten.** Completed scenarios still pass through `analyze()` (`campaign.sh:143,158`), and the analyzer rewrites all three output files (`analyze_exp.py:609-614`). A fixed bootstrap seed alone does not constitute provenance protection.  
    **Plan fix:** make pre-run SHA-256 manifests and backups mandatory for all E1 `analysis*` directories/logs, compare hashes afterward, and hash raw E2 artifacts when each scenario closes.

## MINOR

16. **Several metric names are inaccurate shorthand.** The actual fields are `drift_post_early_share`, `drift_post_late_share`, `drift_on_phases_share`, `drift_off_phases_share`, and `last_finisher_address` (`analyze_exp.py:31-35,472-473`), not `drift_post_early/late_share`, `drift_on/off_phases_share`, or “worker chạy task cuối” (`ke-hoach...:166-169`).  
    **Plan fix:** use exact column names in the analysis checklist.

17. **Smoke accounting is misstated.** Every `ROUND_START=1` invocation runs one discarded warm-up cell (`bench_exp.sh:598-607`), and each such cell has two builds. Three real smoke invocations therefore mean three measured cells plus three warm-up cells—12 builds total—not “3 cells + 3 warm-up builds” (`ke-hoach...:66`).  
    **Plan fix:** correct the wording while retaining the approximately 0.4 h estimate.

18. **Supervisor audit arithmetic is wrong.** A 90-minute interval over 27 h is about 18 scheduled audits, not 27 (`ke-hoach...:106`; `supervisor.sh:8,31`).  
    **Plan fix:** state approximately 18–20 audits for the first three variants and about eight more for lambda.

19. **The smoke instructions need executable Git-Bash paths.** `OUT_DIR=D:\...` as shell syntax is unsafe because backslashes are consumed unless carefully quoted (`ke-hoach...:85`).  
    **Plan fix:** provide exact commands using quoted `/d/bench-work/out/pilot-e2/...` paths and all official environment variables.

## LIMITATIONS-TO-REPORT

20. **Time-to-adapt is not a survival analysis.** Censored cells become `n_post+1` (`analyze_exp.py:185-188,528-530`); different onset counts can produce different censoring caps, and many equal values force approximate Wilcoxon inference (`analyze_exp.py:94-102`).  
    **Plan fix:** disclose censoring proportions by arm, ties/zero differences, actual paired n, and that estimates are an ordinal capped proxy.

21. **H2.3 has limited power.** Each scenario’s makespan family includes five treatments × two baselines, so the registered comparison pays a Holm correction over ten tests (`analyze_exp.py:486-512`), with only 20 blocks.  
    **Plan fix:** elevate this from an operational risk to the statistical-limitations section and state that no post-outcome relaxation of the family is allowed.

22. **On/off exposure is only approximately controlled.** Ten-second phases are implemented by 0.2 s polling and Docker operations, paused during the build gap while stress state remains active (`bench_exp.sh:255-275,292-316`). Phase metrics pool decisions, so phase sample sizes are unequal and scheduler-dependent.  
    **Plan fix:** report observed phase durations, event dispatch counts, task counts per phase, and boundary-spanning phases.

23. **H2.4 does not obtain direct lambda-level tests from the default analyzer.** With eight HG/HG-D treatment labels and two baselines, default analysis creates 16 treatment-versus-LL/SED comparisons per metric—not pairwise contrasts among lambda levels (`campaign.sh:35`; `analyze_exp.py:490-510`).  
    **Plan fix:** state explicitly that the registered lambda claim rests on descriptive medians; generated baseline comparisons remain exploratory and are not a trend test.

24. **Generalizability and temporal confounding remain limited.** Scenarios run in a fixed order (`hypotheses.md:160-162`), the same seeded arm-order mechanism is reused, one host/workload is used, and E2’s persistent two-build coordinator differs from E1/paper protocol (`hypotheses.md:111-120,208-228`).  
    **Plan fix:** list these plainly as design limitations without attempting post-hoc correction.

**Verdict: NOT READY.**