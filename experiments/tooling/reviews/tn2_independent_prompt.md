You are the Codex WORKER. Task: an INDEPENDENT recomputation of the confirmatory tests of Experiment 2, to cross-check the project's analyzer. Do NOT import, copy or run D:\bench-work\hybridgrid\scripts\analyze_exp.py; write your own code from the definitions below. Read-only on all data. Write ONLY inside D:\bench-work\review\tn2_independent\ (create it): your script(s), results CSV and a short README. Use python via the venv: D:\bench-work\venv\Scripts\python.exe (numpy, scipy, pandas available). Do not run docker or anything else.

Data: D:\bench-work\out\official\E2-perm, E2-trans, E2-onoff, E2-lambda. Per scenario: results.csv (arm,round,order_pos,build_idx,client,elapsed_s,cell_makespan_s), tasks_<arm>_round<k>.jsonl (one JSON per line; records with "event":"task_completed" have dispatch_seq, worker_address, success; records with "event":"injected_event" have kind (drift_on/drift_off/drift_note), dispatch_count, detail), events_<arm>_round<k>.csv. Pre-registration: D:\Đồ án tốt nghiệp\hybridgrid\docs\thesis\hypotheses.md sections 0 and 2 (Vietnamese). Target worker: "worker-5:50051".

Definitions (from the pre-registration):
- Cell = (arm, round). Session makespan = sum over build_idx 1..2 of cell_makespan_s.
- onset = dispatch_count of the FIRST drift_on event + 1. post = completed tasks with dispatch_seq >= onset; pre_all = dispatch_seq in [1, onset).
- drift_post_share = fraction of post tasks whose worker_address == target.
- drift_pre_all_share likewise on pre_all.
- time_to_adapt: walk post tasks in dispatch_seq order; first count c >= 20 such that the share of target among the last 20 post tasks (positions c-20..c-1) <= 0.5 * drift_pre_all_share; value = c; if never, censored, rank value = len(post) + 1; NA if drift_pre_all_share < 0.05.
- target_after_recovery_share (E2-trans only): requires exactly one drift_on and one drift_off whose detail does not start with "cell_end"; recovered tasks = dispatch_seq >= (dispatch_count of that drift_off + 1); share of target among them.
- Paired test per comparison: within each round, difference treatment - baseline over the 20 rounds; two-sided Wilcoxon signed-rank (scipy.stats.wilcoxon, zero_method="wilcox"; exact when <=25 nonzero and no ties, else approx); matched-pairs rank-biserial r (negative = treatment smaller); median paired difference.
- Holm within family. Families: (A) "D vs HG" family per scenario E2-perm/E2-trans/E2-onoff: treatments hybrid-linucb-d-g095, -g098, -g099, hybrid-linucb-d-arm-g098 vs baseline hybrid-linucb (4 comparisons per metric). (B) default family per scenario: treatments = all arms whose label starts with hybrid-linucb (5 in E2-perm/trans/onoff) x baselines leastloaded and sed (10 comparisons per metric).

Compute and write results.csv with columns scenario,family,metric,treatment,baseline,n,treatment_median,baseline_median,median_diff,p_raw,p_holm,r_rb for at least:
- E2-perm family A: time_to_adapt (rank values), drift_post_share.
- E2-trans family A: target_after_recovery_share.
- E2-perm and E2-onoff family B: makespan.
Also report: per scenario and arm, number of cells, number of cells with 590 successful tasks, censored time_to_adapt count (E2-perm), and first drift_on dispatch_count min/max.
Final answer: a compact table of the five confirmatory rows (H2.1a: E2-perm time_to_adapt d-g098 vs hybrid-linucb; H2.1b: E2-perm drift_post_share same pair; H2.2: E2-trans target_after_recovery_share same pair; H2.3a/b: makespan d-g098 vs leastloaded in E2-perm / E2-onoff) with n, medians, median_diff, p_raw, p_holm, r_rb, plus any data anomaly you found.
