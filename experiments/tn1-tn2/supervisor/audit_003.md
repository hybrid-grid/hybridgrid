SEVERITY: WARN

- `E1-j7` is still incomplete/in flight: `results.csv` has 18/80 expected cells, `order_log.csv` has 5/20 blocks, and 62 cells are not yet present. Evidence: `campaign.log` row 7 starts `E1-j7` with no exit line yet; `E1-j7/runner.log` row 865 is currently at `cell sed round 5 attempt 1`.

- Measurement quality warning: `E1-j6` has brief builder CPU saturation at or above 95% of the 2.0 CPU quota: 16/2151 builder samples = 0.74%. Max was `212.55%` in `E1-j6/stats_sed_round3.csv` row 55; also `199.28%` in `stats_hybrid-linucb_round15.csv` row 49 and `206.93%` in `stats_leastloaded_round18.csv` row 51. Coordinator saturation was 0/2155 samples; max coordinator was 15.82%.

- Completed scenario integrity looks clean for `E1-j6`: `results.csv` has 80/80 expected rows, duplicate `(arm, round, build_idx, client)` cells = 0, `tasks_*.jsonl` files = 80/80 and non-empty, every task file has 295 `task_completed`, `success=false` count = 0, and dispatch sequence sets are continuous `1..295`. `failures.csv` is header-only.

- `E1-j7` completed cells are clean so far but partial: 18 existing `tasks_*.jsonl` files are non-empty; each has 295 `task_completed`, `success=false` count = 0, and dispatch sequence sets are continuous `1..295`. `failures.csv` is header-only.

- No requested configuration drift found in existing `meta.json`: both scenarios use `ARMS=leastloaded sed hybrid-linucb hybrid-linucb-d-g098`, `CLIENTS=1`, `SESSION_BUILDS=1`, `DRIFT=none`, `BUILDER_CPUS=2.0`, and git commit `1b8339d8f2be1e3e707272e2eefab752e355a137`, matching `campaign.log` row 1 commit `1b8339d`.

- Preflight checks passed for both starts: `E1-j6/preflight_r1.txt` rows 1, 5, 17, 19, 24 show `STRICT=1`, expected commit matched, no noisy apps, performance power plan, and `preflight: no FAIL`; `E1-j7/preflight_r1.txt` has the same evidence at rows 1, 5, 17, 19, 24.

- No makespan outliers by the requested thresholds: `E1-j6` median = 92.488s, min/max = 87.674s/98.216s; `E1-j7` partial median = 88.839s, min/max = 85.005s/92.538s. None are `<0.7x` or `>1.5x` scenario median.

- No monotone time-of-day makespan drift detected: `E1-j6` round medians are not monotone; `E1-j7` partial round medians through round 5 are also not monotone.

- Block order is valid for `E1-j6`: `order_log.csv` has 20/20 rows and no malformed arm sets. First/last counts are not systematic: leastloaded first/last 5/6, sed 6/2, hybrid-linucb 6/4, hybrid-linucb-d-g098 3/8.

- Process log shows no retries/resumes/abort limit issue so far: `campaign.log` rows 2-6 show `E1-j6` started at block 1, runner exit 0, all 20 blocks complete, and analysis rc=0; no scenario has exceeded the preregistered resume limit of 2.