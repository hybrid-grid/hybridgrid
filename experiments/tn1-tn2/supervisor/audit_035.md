SEVERITY: OK

- At the 2026-10-21 00:24:52 +07 snapshot, E2-perm, E2-trans, and E2-onoff each have 20/20 complete blocks: 7 arms × 2 builds × 1 client × 20 = 280 result rows. E2-lambda is actively running block 15 after 14/20 complete blocks: 10 × 2 × 1 × 14 = 280 rows; incompleteness alone is not a warning. Evidence: each scenario’s [results.csv](</D:/bench-work/out/official/E2-perm/results.csv>) and [campaign.log](</D:/bench-work/out/official/campaign.log>).

- Across all completed E2 cells: no duplicate result keys; no missing or empty task logs; no malformed JSON; all 590 task completions per cell have `success=true`; dispatch sequences are continuous and unique; all build statuses are zero with valid terminal `HGTIME`; all cells confirm five registered workers. First `drift_on` counts are within [101,160], with at least 100 subsequent completions.

- One registered cell failure occurred during the live audit: E2-lambda `leastloaded`, block 15, attempt 1 registered 0/5 workers ([failures.csv line 2](</D:/bench-work/out/official/E2-lambda/failures.csv:2>), [runner.log line 5232](</D:/bench-work/out/official/E2-lambda/runner.log:5232>)). The permitted immediate retry registered all five workers ([runner.log line 5263](</D:/D:/bench-work/out/official/E2-lambda/runner.log:5263>)); no retry rule has been exceeded.

- Abort-driven block reruns are zero in every E2 scenario: no `reruns.csv` exists, which the campaign runner defines as count zero. E2-lambda has one documented operator pause/resume and one `manual stop` interruption, not a failure; it remains below the limit of two resumes. Evidence: [interruptions.csv](</D:/bench-work/out/official/E2-lambda/interruptions.csv>) and [campaign.log](</D:/bench-work/out/official/campaign.log:60>).

- Configuration matches preregistration in every `meta.json`: correct arms, jobs, clients, two E2 session builds, drift modes, `BUILDER_CPUS=2.0`, and clean commit `1b8339d8f2be1e3e707272e2eefab752e355a137`, matching campaign commit `1b8339d`. E2 Docker capacity is constant at 72 CPUs and 33,591,300,096 bytes RAM.

- Every E2 start/resume has the required preflight: r1 for all four scenarios and r15 for the E2-lambda resume. Each records `STRICT=1`, `EXPECT_COMMIT=1b8339d`, matching commit, High performance, no known noisy applications, and no `FAIL` lines. Evidence: [preflight_r15.txt](</D:/bench-work/out/official/E2-lambda/preflight_r15.txt>) and corresponding `preflight_r1.txt` files.

- E1 integrity is unchanged: all 2,672 entries in [manifest_E1.sha256](</D:/bench-work/out/official/manifest_E1.sha256>) exist and match SHA-256. The only five extras are `analysis.done` markers—one per E1 scenario—and there are no extra non-marker files.

- No makespan outliers crossed the registered thresholds. Scenario medians were 187.589 s (perm), 176.533 s (trans), 182.655 s (onoff), and 188.790 s (completed lambda cells); no cell was above 1.5× or below 0.7× its scenario median.

- Saturation is below warning thresholds: builder samples at ≥190% were 33/7,678 (0.43%), 49/7,273 (0.67%), 42/7,444 (0.56%), and 31/7,699 (0.40%); coordinator samples at ≥23.75% were 0 in every scenario. Evidence: each scenario’s `stats_*_round*.csv`.

- No monotone time drift was detected: block-median Spearman correlations were 0.268, 0.266, 0.271, and 0.090 for perm, trans, onoff, and current lambda data. No sequence was strictly monotonic.

- Block order is valid with no duplicate/malformed orders. In completed seven-arm scenarios, first-position counts range 1–4/20 and last-position counts 1–6/20; current lambda counts peak at 4 first and 3 last among 15 logged orders. No arm is systematically fixed first or last.

- Warm-up/cooldown and diagnostics are clean: every E2 scenario has exactly one discarded warm-up, valid zero-status/HGTIME logs, no round-0 result rows, and `COOLDOWN=8`; all E2-trans cells have exactly one non-`cell_end` on/off pair; no stress-ng cleanup leak or error marker was found.