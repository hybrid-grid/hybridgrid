SEVERITY: OK

- No validity-threatening problems found in the current campaign outputs. Scenario dirs present now are `E1-j6`, `E1-j7`, `E1-j8`, `E1-j9`, `E1-2c`; `campaign.log` line 27 records `STOP_AFTER reached after E1-2c`, so absent E2 dirs are consistent with the phase-1 stop.

- Integrity OK: `results.csv` rows match expected blocks/arms/builds/clients: `E1-j6/j7/j8/j9` each `80 = 20x4x1x1`; `E1-2c` `160 = 20x4x1x2`. No duplicate `(arm, round, build_idx, client)` rows.

- Task logs OK: all five scenarios have 80 expected `tasks_<arm>_round<k>.jsonl` files, none empty. Across 400 task-log cells, every build has exactly 295 `task_completed` records: single-client cells have 295, `E1-2c` cells have 590 split as `[295,295]`.

- Task outcomes/sequences OK: computed from all task JSONL files, `success=false` count is 0, duplicate task IDs count is 0, and `dispatch_seq` is continuous per cell (`1..295` or `1..590`).

- Failures/retries OK: every `failures.csv` contains only header row `arm,round,attempt,reason`; no failed cells, no immediate retry evidence, and no scenario-level retry/resume exceeded section 3.

- Configuration OK: all `meta.json` files have `ARMS="leastloaded sed hybrid-linucb hybrid-linucb-d-g098"`, `SESSION_BUILDS=1`, `DRIFT=none`, `BUILDER_CPUS=2.0`, clean git state, and identical `git_commit=1b8339d8f2be1e3e707272e2eefab752e355a137`, matching `campaign.log` line 1 commit `1b8339d`. Jobs/clients match pre-registration: `j6/j7/j8/j9 = 6/7/8/9 clients=1`; `E1-2c jobs=8 clients=2`.

- Preflight OK for every start: each `preflight_r1.txt` line 1 has `STRICT=1 EXPECT_COMMIT=1b8339d`; line 5 says commit matches; line 10 says Docker `NCPU=72 Mem=31 GiB`; line 17 says no known noisy apps; line 19 says performance power plan; line 24 says `preflight: no FAIL`.

- Measurement quality OK: no makespan outliers by the registered threshold. Scenario medians/outlier counts: `E1-j6 92.488s/0`, `E1-j7 87.816s/0`, `E1-j8 86.443s/0`, `E1-j9 84.831s/0`, `E1-2c 165.140s/0`.

- Saturation OK: builder samples at >=190% were below 5% in every scenario: `E1-j6 16/2125=0.75%`, `E1-j7 6/2030=0.30%`, `E1-j8 20/1987=1.01%`, `E1-j9 12/1949=0.62%`, `E1-2c 12/6910=0.17%`. Coordinator samples at >=23.75% were `0` in all scenarios.

- Order/time trends OK: no monotone block-level makespan drift in any scenario. `order_log.csv` has 20 rows per scenario; no arm is always first/last. Position counts are the same across E1 dirs, e.g. first-position counts: `sed=6`, `hybrid-linucb=6`, `leastloaded=5`, `hybrid-linucb-d-g098=3`.

- Process OK: `campaign.log` lines 2,7,12,17,22 all start at block 1; lines 3,8,13,18,23 all `runner exit 0`; lines 4,9,14,19,24 all `all 20 blocks complete`. No abort/resume entries found.

- Known harmless facts observed: runner logs contain `Exception ignored ... OSError: [Errno 22]` counts of 6-12 per scenario, matching the documented sampler-stop artifact; `host.nproc` varies once (`E1-j7=36`, others `28`) while `docker_NCPU=72` and `docker_MemTotal=33592029184` stay constant.