SEVERITY: OK

- Completed scenario integrity is clean: `E1-j6`, `E1-j7`, `E1-j8`, `E1-j9` each have `results.csv` 80/80 expected rows = 20 blocks x 4 arms x 1 build x 1 client; duplicates 0; missing 0.

- `E1-2c` is still in progress, not a validity warning: `D:\bench-work\out\official\E1-2c\results.csv` has 150/160 rows; complete blocks are 1-18, block 19 has `hybrid-linucb`, `hybrid-linucb-d-g098`, `sed`, and is missing `leastloaded`; block 20 is not present yet.

- Task log integrity is clean for all written cells: completed E1 scenarios have 80 non-empty `tasks_*_round*.jsonl` each; `E1-2c` has 75 non-empty written task logs. Every checked build has exactly 295 `task_completed` records; `E1-2c` two-client cells have 590 per arm-round = 295 per client build. `success=false` count is 0 and dispatch sequences are continuous.

- `failures.csv` is empty in all scenario directories checked: each contains only `arm,round,attempt,reason`; no retry rule from prereg section 3 was triggered or exceeded.

- Metadata matches prereg for all scenario dirs: `ARMS`, `JOBS`, `CLIENTS`, `SESSION_BUILDS`, `DRIFT`, and `BUILDER_CPUS=2.0` match expected E1 settings. All `meta.json` files have `git_commit=1b8339d8f2be1e3e707272e2eefab752e355a137`, matching `campaign.log` start commit `1b8339d`.

- Host Docker capacity is stable: all `meta.json` files report `docker_NCPU=72` and `docker_MemTotal=33592029184`. `nproc` varies 28/36, treated as the known harmless Git Bash process-view artifact.

- Preflight checks are clean for every start/resume seen: `preflight_r1.txt` exists for all five scenario dirs, with `STRICT=1`, `EXPECT_COMMIT=1b8339d`, commit match OK, High performance power plan OK, no known noisy apps, and no actual `FAIL` lines. No resume preflight beyond `r1` exists.

- Measurement saturation is below warning thresholds: builder samples at >=95% quota are `E1-j6` 16/2125, `E1-j7` 6/2030, `E1-j8` 20/1987, `E1-j9` 12/1949, `E1-2c` 12/6478; coordinator samples above 95% quota are 0 in every scenario.

- Makespan quality checks show no outliers: no written cell is >1.5x or <0.7x its scenario median. Medians were `E1-j6` 92.488s, `E1-j7` 87.816s, `E1-j8` 86.443s, `E1-j9` 84.831s, `E1-2c` 165.106s.

- No monotone time-of-day drift was found in block mean makespan for any scenario. Example evidence: `E1-j6` block means start 93.575s and end 92.578s; `E1-j9` starts 83.976s and ends 85.481s; neither series is monotone.

- Block order is not perfectly balanced but does not show a validity-threatening pattern by itself: completed E1 scenarios each have first counts `sed=6`, `leastloaded=5`, `hybrid-linucb=6`, `hybrid-linucb-d-g098=3`; last counts `leastloaded=6`, `hybrid-linucb-d-g098=8`, `hybrid-linucb=4`, `sed=2`.

- Campaign process log is clean so far: `campaign.log` shows fixed order `E1-j6 -> E1-j7 -> E1-j8 -> E1-j9 -> E1-2c`, runner exit 0 and analysis rc=0 for completed scenarios, no runner aborts, and no starts at block >1, so resumes used = 0 per scenario.

- Known harmless sampler shutdown messages appear in `runner.log` as expected: `OSError: [Errno 22] Invalid argument` lines occur, but they match the preregistered harmless docker-stats stdout condition and are not reflected in `results.csv` or task logs.