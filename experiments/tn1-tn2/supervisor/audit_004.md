SEVERITY: WARN

- WARN: Host-side `nproc` drift in `meta.json`: `E1-j6/meta.json:39` records `nproc: 28`, while `E1-j7/meta.json:39` records `nproc: 36`; Docker capacity stayed constant at `docker_NCPU: 72` and `docker_MemTotal: 33592029184` in both files (`meta.json:40-41`).

- WARN: Builder saturation threshold exceeded in `E1-j6/stats_hybrid-linucb_round15.csv`: rows 27 and 49 show builder CPU `196.26%` and `199.28%`, so 2/25 builder samples = 8.0% were >= 95% of the 200% quota. Coordinator saturation in that cell was 0/25.

- WARN: Repeated ignored stdout exceptions appear in runner logs, though not as aborts: e.g. `E1-j6/runner.log:450-454` and `E1-j7/runner.log:558-562` show `Exception ignored ... stdout` / `OSError: [Errno 22] Invalid argument`, followed by task-log copy and `completed ...`.

- Integrity OK for completed `E1-j6`: `results.csv` has 80 rows = 20 blocks x 4 arms x 1 build x 1 client, no duplicate cells, `order_log.csv` has 20 blocks, and all 80 `tasks_*.jsonl` files are non-empty with exactly 295 `task_completed`, zero `success=false`, and continuous `dispatch_seq` 1-295.

- `E1-j7` is still in progress, not a warning by itself: audit snapshot shows `results.csv` has 70 rows, complete blocks 1-17 plus partial block 18 (`leastloaded`, `sed` done), 70 non-empty task logs, and current runner tail at `cell hybrid-linucb round 18 attempt 1`.

- No retry-rule excess found: `E1-j6/failures.csv` and `E1-j7/failures.csv` contain only the header (`arm,round,attempt,reason`), no failure rows; runner logs show no `attempt 2`, no campaign resume, and `campaign.log` has `E1-j6 runner exit 0`.

- Registered config fields match for existing scenarios: both metas have arms `leastloaded sed hybrid-linucb hybrid-linucb-d-g098`, `CLIENTS=1`, `SESSION_BUILDS=1`, `DRIFT=none`, `BUILDER_CPUS=2.0`, and git commit `1b8339d8f2be1e3e707272e2eefab752e355a137`, matching the `campaign.log` start commit prefix `1b8339d`.

- Preflights OK: `E1-j6/preflight_r1.txt` and `E1-j7/preflight_r1.txt` both show `STRICT=1`/`EXPECT_COMMIT=1b8339d` on line 1, commit match on line 5, no known noisy apps on line 17, High performance plan on lines 18-19, and `preflight: no FAIL` on line 24.

- Measurement shape otherwise OK: no makespan outliers by the registered rule in current `E1-j6` or `E1-j7` data, no monotone block-level makespan drift detected, and no arm is always first/last in `order_log.csv` (`E1-j6` largest last-position count is `hybrid-linucb-d-g098` 8/20; `E1-j7` partial is 8/18).