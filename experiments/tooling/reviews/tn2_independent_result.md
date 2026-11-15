The independent recomputation is in [recompute.py](/D:/bench-work/review/tn2_independent/recompute.py), with [results.csv](/D:/bench-work/review/tn2_independent/results.csv) and the per-arm checks in [README.md](/D:/bench-work/review/tn2_independent/README.md). All 70 computed comparison rows match the analyzer’s saved output.

*Medians and paired differences are treatment − baseline; makespan is in seconds.*

| Test | n | Treatment median | Baseline median | Median paired difference | p raw | p Holm | r RB |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| H2.1a: time to adapt, E2-perm | 20 | 45.5 | 50.0 | 2.5 | 0.793527 | 1.000000 | −0.068 |
| H2.1b: post-drift share, E2-perm | 20 | 0.1900 | 0.2046 | −0.0094 | 0.409098 | 0.818195 | −0.219 |
| H2.2: recovery share, E2-trans | 20 | 0.2699 | 0.2699 | −0.0027 | 0.985435 | 1.000000 | −0.010 |
| H2.3a: makespan, E2-perm | 20 | 188.468 | 189.464 | −2.254 | 0.036234 | 0.217403 | −0.533 |
| H2.3b: makespan, E2-onoff | 20 | 183.797 | 184.630 | −1.725 | 0.058258 | 0.233032 | −0.486 |

None of the five comparisons reaches 0.05 after Holm adjustment. All 620 cells have 590 successful tasks. In six E2-onoff cells, JSONL contains a terminal `drift_on` at dispatch count 590 that the event CSV omits; it occurs after the final task and does not affect these tests. One E2-perm time-to-adapt cell is censored.