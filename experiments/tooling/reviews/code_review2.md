## Round-1 disposition

- RESOLVED — Runner inputs are cleared before each campaign and only registered values are passed (`campaign.sh:23-33,266-268`).
- ACCEPTED-AS-DECIDED — Abandoned and manual interruptions are recorded separately and deliberately excluded from the failed-cell rerun cap (`campaign.sh:44-54,242-248`).
- RESOLVED — The documented launch now writes exactly one approval line and logs the reason separately (`ke-hoach-thi-nghiem-2.md:166-169`).
- RESOLVED — The first lambda decision is frozen and later changes are rejected (`campaign.sh:59-83`).
- RESOLVED — `stop_campaign.sh` validates PID liveness and command identity before signaling (`stop_campaign.sh:15-37`).
- ACCEPTED-AS-DECIDED — TERM still lacks a dedicated process group, but the selected timeout policy kills matching runner shells and then the campaign (`stop_campaign.sh:38-53`).
- RESOLVED — Cleanup uses `--remove-orphans` and checks containers in every state (`stop_campaign.sh:56-69`).
- OPEN — The damaged-directory guard can misclassify legitimate block-1 recovery, described below (`campaign.sh:146-154`).
- RESOLVED — The applicable preflight report is copied before replacement (`campaign.sh:110-120`).
- ACCEPTED-AS-DECIDED — Stop tests remain fake-runner tests, but cases are isolated and expanded as requested (`tests/test_campaign.sh:182-222`).
- RESOLVED — Analysis markers now store and compare normalized arguments per family (`campaign.sh:289-300`; `analysis_marker.sh:14-29,64`).
- ACCEPTED-AS-DECIDED — Numbering, timeout, and shrink detection are fixed; same-size rotation/inode detection was not selected (`supervisor.sh:14-20,31-40,54-59`).
- RESOLVED — All recorded SHA-256 values match the current files (`CHANGELOG-phase2.md:7-13`).

## New round-2 bug

- MAJOR — Legitimate block-1 recovery can be permanently rejected. `next_round` treats any non-empty archive plus missing `meta.json` as damage (`campaign.sh:146-154`), while retries create that archive (`campaign.sh:112-122`). The real runner does not write `meta.json` until the warm-up succeeds (`hybridgrid/scripts/bench_exp.sh:598-607`). Thus, after one archived warm-up failure/interruption, another warm-up abort leaves a valid `.pending_archive` but the next launch stops as `MISMATCH` before the permitted rerun. This can abort the 40-hour campaign, although it preserves existing data. Fix by recording an initialized-scenario provenance marker after the first clean-directory check and allowing authenticated block-1 recovery, rather than treating every campaign-created archive as evidence of damage.

Fresh E2-perm, ordinary block-7/block-1 failed-cell retries, manual stop/resume, one reboot/resume, and immutable SKIP/RUN resume otherwise preserve completed blocks and do not mix partial data.

Claude’s reported results: `test_campaign` 21/21 and `test_supervisor` 5/5.

**Verdict: GO AFTER FIXES.**