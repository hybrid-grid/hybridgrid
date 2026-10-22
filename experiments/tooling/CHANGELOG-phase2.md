# Phase 2 orchestration changes

Operational tooling changes only. The frozen runner, analyzer, preregistration, and official outputs were not edited. SHA-256 values are lowercase hexadecimal; original values come from `backup-pre-phase2/SHA256SUMS`.

| File | Plan item and reason | Before SHA-256 | After SHA-256 |
|---|---|---|---|
| `campaign.sh` | T1: persistent append-only abort ledger and fixed two-rerun cap. T2: archive partial block and warm-up files before relaunch. T3: parse and enforce the one-line phase decision. T4: check and write analysis completion marker. T5: PID files, live-instance guard, background runner, and manual-stop handling. | `96ddb68a14cc65a99c8be5285de50b9380e16d1f6554dc46498b3777414657eb` | `762f5e3c6752e0dff4b29e5c0e765e382509e33cef2607429810ef4b2cc7c3aa` |
| `supervisor.sh` | T6: new-log-line terminal detection, PID lock, `gpt-5.6-sol`, one final audit, and notification without campaign control. | `8bcbc748bf4eb174f640b5e84e14f63a3c29f133354d266d75d01b7faa671d53` | `e9173c0b22d301bb3aece79706c69964d108056e55fc7a195f0eb79f656a243e` |
| `supervisor_prompt.md` | T7: distinguish registered E2 exclusion criteria from diagnostics and preserve E2 result blinding. | `12305284165a499e35d0256feeebc629473ceae19fe5900162b8b7d862929c1c` | `56e82207761da7a9ee0d1cc151bc12463e92bf9d485e19e9e633d4de2006f8e3` |
| `stop_campaign.sh` | T5: request stop, signal runner, await campaign, force termination on timeout, tear down `hgexp`, and verify no project containers remain. | new file | `8d8c040638d7fbb42c55ea43f598e024fae889d6e0a42ddec1b7e214559f9506` |
| `analysis_marker.sh` | T4: atomically write and verify analyzer and output file hashes for all required analysis families. | new file | `6581740fbbbbdd4239cb49284e4252ae144db3f91b18d3e45c248b7e0ed8cf56` |
| `tests/test_campaign.sh` | T1–T5: temporary fake-runner cases for markers, phase gate, persistent reruns, exact archives, and manual stop. | new file | `a3d88e836c6c19e7a36d34e6a68e7e222a76615abbdd7186ac9170a3cd51f7aa` |
| `tests/test_supervisor.sh` | T6: temporary fake-Codex cases for old and new terminal lines, lock, final audit, and notification. | new file | `93105f4a66e90e4cb07ab2d89dea6102ed2bee37ac4b346d2a08b80d76babf0e` |
| `CHANGELOG-phase2.md` | T1–T7: records the operational diff and its file hashes. | new file | Self-hash is reported with the deliverable; embedding it here would change it. |

The plan calls for manually adding markers to the five official E1 directories after verifying their existing logs. That operation was not performed because the benchmark task explicitly forbids modifying `out/`.

Validation note: Git Bash could not start in the current sandbox (`CreateFileMapping ... Win32 error 5`), so `bash -n` and both Bash test scripts could not execute here. `shellcheck` is not installed. No real runner, Docker, Codex, or official campaign was invoked.

## Round 2

- `campaign.sh`: clear all inherited runner inputs after `env.sh` and log the reset (reviewer blocker: configuration can change between blocks). Record active attempts atomically, distinguish abandoned and manual interruptions from failed-cell reruns, and archive their partial files (reviewer blocker: unclean termination had no record). Freeze the first valid lambda decision in `PHASE2_DECISION` and reject edits (reviewer major: mutable decision). Refuse a fresh start when missing aggregate files coexist with later-round tasks or a populated archive (reviewer major: damaged directory). Preserve each prior block preflight report before replacing it (reviewer major: preflight evidence overwritten). Pass normalized analyzer arguments to marker checks (reviewer minor: missing command record).
- `analysis_marker.sh`: store and compare one normalized argument string per required analysis family (reviewer minor: analysis command changes did not invalidate the marker).
- `stop_campaign.sh`: check PID liveness and `/proc/<pid>/cmdline` identity, remove stale PID files, stop waiting when the campaign process exits, force-kill remaining `bench_exp.sh` processes on timeout, remove Compose orphans, and inspect all project containers (reviewer major: stale PID signaling, descendant cleanup, and incomplete Docker verification).
- `supervisor.sh`: continue numbering after the highest audit file, bound each Codex call to 2700 seconds, and reset the log offset after truncation (reviewer minor: restart and rotation handling).
- `tests/test_campaign.sh`: isolate the manual-stop archive case; cover inherited environment, abandoned attempts, immutable lambda decisions, damaged directories, archived preflight reports, changed analyzer arguments, and stale PIDs (reviewer major: stop test reused files; other round-two findings lacked regression coverage).
- `tests/test_supervisor.sh`: verify numbering after `audit_005.md`, the timeout wrapper, and log-truncation recovery (reviewer minor: restart and rotation handling).

The SHA-256 values in the table above were recomputed after the final writes to the six scripts and tests in this round. The changelog's own SHA-256 must be reported externally after its final write, since embedding that value here changes the file.

## Round 3 (Claude, 2026-10-17, after code review round 2 and real S5 test)

- campaign.sh: next_round no longer treats a non-empty archive/ as a damaged directory (review round 2 MAJOR: block-1 warm-up recovery would be refused); only later-round task logs with missing results.csv/meta.json stop the campaign.
- campaign.sh: current_block(): interruptions.csv, reruns.csv and .pending_archive record the first incomplete block at runner exit (S5 showed the runner start block was recorded).
- tests/test_campaign.sh: fake analyzer defaults to <out>/analysis like the real analyzer; archive-only directory test now expects block-1 recovery.
- supervisor_prompt.md: known harmless facts (d) on/off phases ~15 s and drift_on 125-132, (e) E1 directories already audited, (f) manual-stop interruptions.
- out/official: manifest_E1.sha256 (2672 files) and analysis.done markers for the five E1 directories, written after confirming the ten TN1 analysis runs were rc=0 (campaign.log 2026-10-10/11).

SHA-256 after round 3:

```
5b93a814fa36c3413b1dd796ad642f83bf7f59c80c50d85ef2a2b77d29fe486e *campaign.sh
e9173c0b22d301bb3aece79706c69964d108056e55fc7a195f0eb79f656a243e *supervisor.sh
4d8cb138c6eedf41e2bda7f2f158b2df7d644b90474148012b069688833758c6 *supervisor_prompt.md
8d8c040638d7fbb42c55ea43f598e024fae889d6e0a42ddec1b7e214559f9506 *stop_campaign.sh
6581740fbbbbdd4239cb49284e4252ae144db3f91b18d3e45c248b7e0ed8cf56 *analysis_marker.sh
6d0e3cc1521e46bdf3149a78fb6409da75a193a649748274478289388e91112b *tests/test_campaign.sh
93105f4a66e90e4cb07ab2d89dea6102ed2bee37ac4b346d2a08b80d76babf0e *tests/test_supervisor.sh
```

## TN3 (Experiment 3 campaign preparation, 2026-10-22)

- Added `campaign_tn3.sh` as a separate Experiment 3 orchestrator. It defaults to `out/official-tn3`, checks that required `TN3_REPO` matches `EXPECT_COMMIT`, requires positive `P90_SYN` and `P90_ABSL`, runs E3-syn, E3-absl, then E3-syn-drift, and pins all runner inputs. It retains the TN1-2 append-only abort ledger and two-rerun cap, partial-block archives, interruption and PID records, STOP_REQUESTED handling, power-plan switch and restore, STRICT preflight before every start/resume, configuration MISMATCH checks, and analysis.done markers. The two analysis families are the default analyzer and P90/SED-P90 versus LeastLoaded.
- Added `supervisor_prompt_tn3.md` for read-only, blinded TN3 audits, and `supervisor_tn3.sh`, a copy of the existing supervisor with only its default campaign and prompt paths changed.
- Added `tests/test_campaign_tn3.sh` with a sed-rewritten campaign copy, fake runner/analyzer/preflight/power command, and temporary data under `tmp-tests`. It covers fixed inputs and order, per-scenario thresholds, missing or mismatched required inputs, configuration mismatch, reruns and archives, STOP_REQUESTED, and both analysis markers. Bash tests were not run here because the worker cannot run Bash and the pilot is active; Claude will run them.
- Existing `preflight.sh` still requires branch `feat/exp-readiness` under STRICT mode. The frozen TN3 clone must use that local branch name for this unchanged preflight. Existing `stop_campaign.sh` identifies `campaign.sh` by process name, so use the TN3 campaign's STOP_REQUESTED file for a clean TN3 stop.
