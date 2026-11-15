# Independent Experiment 2 recomputation

Run: `D:\bench-work\venv\Scripts\python.exe recompute.py` from this directory.
Inputs are read only. The script reads `results.csv`, task JSONL, and event CSV
for all four E2 scenarios; it writes only `results.csv` and this README.
No project analyzer code is imported or run.

Cell makespan is the sum of the two build makespans. Dispatch metrics use
completed tasks ordered by dispatch sequence and event positions from task JSONL.
Time to adapt uses the first qualifying 20-task post-onset window; censored
cells take rank value `len(post)+1`, and pre-share below 0.05 gives NA.
Pairs are matched by round; differences are treatment minus baseline.
Two-sided Wilcoxon uses `zero_method=wilcox`, exact with <=25 nonzero
untied absolute differences, otherwise normal approximation, without
continuity correction. Matched-pairs rank-biserial r uses signed rank sums.
Holm adjustment is separate within each scenario, family, and metric.

## Cell checks

| Scenario | Arm | Cells | 590 successful | Censored TTA | First drift_on min-max |
| --- | --- | ---: | ---: | ---: | ---: |
| E2-lambda | hybrid-linucb | 20 | 20 | — | 128–131 |
| E2-lambda | hybrid-linucb-d-g098 | 20 | 20 | — | 128–130 |
| E2-lambda | hybrid-linucb-d-g098-l0 | 20 | 20 | — | 127–132 |
| E2-lambda | hybrid-linucb-d-g098-l025 | 20 | 20 | — | 128–130 |
| E2-lambda | hybrid-linucb-d-g098-l100 | 20 | 20 | — | 127–130 |
| E2-lambda | hybrid-linucb-l0 | 20 | 20 | — | 127–132 |
| E2-lambda | hybrid-linucb-l025 | 20 | 20 | — | 127–131 |
| E2-lambda | hybrid-linucb-l100 | 20 | 20 | — | 128–132 |
| E2-lambda | leastloaded | 20 | 20 | — | 127–131 |
| E2-lambda | sed | 20 | 20 | — | 128–130 |
| E2-onoff | hybrid-linucb | 20 | 20 | — | 127–132 |
| E2-onoff | hybrid-linucb-d-arm-g098 | 20 | 20 | — | 128–131 |
| E2-onoff | hybrid-linucb-d-g095 | 20 | 20 | — | 128–131 |
| E2-onoff | hybrid-linucb-d-g098 | 20 | 20 | — | 128–130 |
| E2-onoff | hybrid-linucb-d-g099 | 20 | 20 | — | 127–130 |
| E2-onoff | leastloaded | 20 | 20 | — | 128–130 |
| E2-onoff | sed | 20 | 20 | — | 128–130 |
| E2-perm | hybrid-linucb | 20 | 20 | 0 | 128–132 |
| E2-perm | hybrid-linucb-d-arm-g098 | 20 | 20 | 0 | 128–131 |
| E2-perm | hybrid-linucb-d-g095 | 20 | 20 | 1 | 127–130 |
| E2-perm | hybrid-linucb-d-g098 | 20 | 20 | 0 | 128–131 |
| E2-perm | hybrid-linucb-d-g099 | 20 | 20 | 0 | 128–131 |
| E2-perm | leastloaded | 20 | 20 | 0 | 127–130 |
| E2-perm | sed | 20 | 20 | 0 | 128–130 |
| E2-trans | hybrid-linucb | 20 | 20 | — | 127–131 |
| E2-trans | hybrid-linucb-d-arm-g098 | 20 | 20 | — | 128–132 |
| E2-trans | hybrid-linucb-d-g095 | 20 | 20 | — | 128–132 |
| E2-trans | hybrid-linucb-d-g098 | 20 | 20 | — | 127–131 |
| E2-trans | hybrid-linucb-d-g099 | 20 | 20 | — | 127–131 |
| E2-trans | leastloaded | 20 | 20 | — | 127–131 |
| E2-trans | sed | 20 | 20 | — | 128–130 |

## Data anomalies

- E2-onoff/hybrid-linucb/r8: JSONL and CSV events differ
- E2-onoff/hybrid-linucb/r9: JSONL and CSV events differ
- E2-onoff/hybrid-linucb/r18: JSONL and CSV events differ
- E2-onoff/hybrid-linucb/r19: JSONL and CSV events differ
- E2-onoff/hybrid-linucb-d-arm-g098/r15: JSONL and CSV events differ
- E2-onoff/hybrid-linucb-d-g099/r11: JSONL and CSV events differ

For these six E2-onoff cells, JSONL has an additional `drift_on`
at dispatch_count 590 after the final task; event CSV omits it.
The first onset and reported confirmatory metrics are unaffected.
