You are a Codex WORKER (task C5 of Experiment 3). Repo D:\Đồ án tốt nghiệp\hybridgrid, branch feat/exp3 (do not switch, do not commit). Edit ONLY scripts/analyze_exp.py and scripts/tests/test_analyze_exp.py. Another worker edits Go code and scripts/bench_exp.sh in parallel: do not touch them. Use the venv Python conventions already used by the tests (stdlib unittest, numpy/scipy/pandas available). You cannot run Python reliably in your sandbox; Claude will run `python3 -m unittest scripts/tests/test_analyze_exp.py`.

Context: Experiment 3 (docs/thesis/ke-hoach-thi-nghiem-3.md sections 3.4 and 5) adds heavy-tailed C++ workloads and two new arm labels `p90` and `sed-p90` (bases accepted by the runner). results.csv now has two extra columns start_ts,end_ts (epoch seconds) per (arm,round,build_idx,client) row; older result files (Experiments 1-2) do not have them and must still analyze exactly as before (all existing tests must pass unchanged; existing outputs for an Experiment 1/2 directory must be byte-identical — add a regression test that runs the analyzer on a small fixture without the new columns and compares the produced pairs.csv/cells.csv columns and values with the current behaviour).

Implement:
1. Accept `p90` and `sed-p90` in BASES / parse_arm (no gamma/arm/penalty suffixes allowed on them). Default comparison families: treat p90 and sed-p90 as BASELINES alongside leastloaded and sed (so default families compare learning treatments against all four baselines); document this in the summary.
2. New per-cell metrics computed per build and aggregated per cell exactly like build_*_makespan_s (one value per build; for multi-build sessions report build_<k>_<metric> and, for the cell-level metric, the value of the LAST build (the measured build in the TN3 protocol) — make this rule explicit in the summary text):
   - top5_strong_share: among the top 5% of successful tasks of that build by preprocessed_size_bytes (ceil(0.05*n), ties by dispatch_seq), the share dispatched to the two workers with the largest worker_cpu_millis present in candidates (registered strength, static);
   - top5_mean_cpu_millis: mean worker_cpu_millis of those tasks;
   - top5_multi_choice_share: share of those decisions where at least two candidates had a free slot (active < max_parallel) at dispatch;
   - tail_s: end_ts of the build (max over clients) minus the maximum dispatch_ts (parsed ISO time from the task log) of successful tasks of that build; NaN when start_ts/end_ts are absent;
   - last_finisher_size_bytes and last_finisher_rank_pct (percentile rank of its preprocessed size within the build).
   Builds are identified by the task log build_id suffix -b<k>-c<client>; all clients of a build belong to the same build.
3. Add the new metrics to METRICS so they get paired tests; keep existing metric order first so existing outputs keep their column order (append new columns at the end).
4. Tests: synthetic fixtures for each metric (including ties and the 5% ceiling with small n), multi-build sessions, missing start/end columns, new arm labels accepted, invalid suffixes on p90 rejected, and the Experiment 1/2 regression described above.

Final answer: files changed, metric definitions as implemented, and test command.
