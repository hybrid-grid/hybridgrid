You are a skeptical senior reviewer of an experimental-methods plan for a master's/bachelor thesis benchmark. READ-ONLY: do not modify any file, do not run Docker, benchmarks, git commit or anything that writes. You may read files and run read-only commands (cat, grep, python reading files).

Files:
- Plan under review (Vietnamese): D:\Đồ án tốt nghiệp\hybridgrid\docs\thesis\ke-hoach-thi-nghiem-2.md
- Frozen pre-registration (authoritative, Vietnamese; cannot be changed): D:\Đồ án tốt nghiệp\hybridgrid\docs\thesis\hypotheses.md (sections 2, 2b, 3 are Experiment 2)
- Runner: D:\bench-work\hybridgrid\scripts\bench_exp.sh ; analyzer: D:\bench-work\hybridgrid\scripts\analyze_exp.py ; orchestrator: D:\bench-work\campaign.sh ; preflight: D:\bench-work\preflight.sh ; supervisor: D:\bench-work\supervisor.sh + supervisor_prompt.md
- Experiment 1 outputs (finished, for reference on formats/timing): D:\bench-work\out\official\E1-*\ and campaign.log ; drift pilot: D:\bench-work\out\pilot-drift\trans_s2\

Task: critique the plan hard. Find:
1. Factual errors in the plan versus the code and the pre-registration (wrong metric names, wrong family sizes, wrong cell counts/time estimates, wrong description of how drift/session/analysis works). Verify against code, cite file:line.
2. Places where the plan would (even subtly) deviate from the pre-registration or introduce researcher degrees of freedom (e.g. decision rules for "supported" that are not in hypotheses.md, extra analyses that could be mistaken for confirmatory, peeking).
3. Operational risks that could waste the ~27-39 h run or corrupt data: bugs in campaign.sh/supervisor.sh/bench_exp.sh paths exercised only by E2 (permanent, onoff, -arm, -g095/-g099, -l labels, SESSION_BUILDS=2 resume, re-analysis of E1 when campaign passes through E1 dirs, stress-ng leaks, drift_on timing, Windows sleep), and whether the proposed smoke test actually covers them. Check the claim about supervisor.sh finished() being triggered by the old 'STOP_AFTER reached' line.
4. Statistical concerns specific to E2 that the plan should state up front (censored time_to_adapt ranked n+1 in Wilcoxon, many ties, Holm family of 10 for H2.3, metric defined only for some cells e.g. target_after_recovery_share NaN, whether --strict drops blocks when a metric is NaN, onoff phases at build boundary).
5. Anything missing that a thesis committee would ask for.

Do NOT propose changing the frozen design (arms, gamma, lambda, DRIFT_AT, hypotheses); if you think the design itself is flawed, say so only as a "limitation to report", clearly separated.

Output (English, concise): sections BLOCKER / MAJOR / MINOR / LIMITATIONS-TO-REPORT, each finding with evidence (file:line or numbers) and a concrete fix to the PLAN. Max ~25 findings total. End with a one-line verdict: READY / READY AFTER FIXES / NOT READY.
