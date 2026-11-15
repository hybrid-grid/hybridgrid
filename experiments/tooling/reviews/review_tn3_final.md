BLOCKER — none.

MAJOR

- [grpc.go:1136](<D:/Đồ án tốt nghiệp/hybridgrid/internal/coordinator/server/grpc.go:1136>): successful tasks longer than 120s receive reward below `-1`, while failures receive exactly `-1`. The 192.797s pilot task scores about `-1.041`, so LinUCB—whose implementation ignores the `success` argument—treats that success as worse than failure. Fix: retain the fixed 120000ms denominator, but set failure reward from `max(120000ms, RequestTimeout)`; at 120s this remains exactly `-1`, preserving Experiments 1–2. Add regression tests for both 120s invariance and failure ordering.

- [campaign_tn3.sh:35](<D:/bench-work/campaign_tn3.sh:35>): the frozen thresholds are not enforced; any positive `P90_SYN`/`P90_ABSL` is accepted. The tests explicitly demonstrate acceptance of `P90_SYN=123456` at [test_campaign_tn3.sh:107](<D:/bench-work/tests/test_campaign_tn3.sh:107>), contradicting registered `150189`. Fix: hard-code or require exact equality to `150189` and `2356237`, then update tests.

- [hypotheses-tn3.md:125](<D:/Đồ án tốt nghiệp/hybridgrid/docs/thesis/hypotheses-tn3.md:125>) does not define handling of missing confirmatory metric values. [analyze_exp.py:622](<D:/Đồ án tốt nghiệp/hybridgrid/scripts/analyze_exp.py:622>) silently performs pairwise NA deletion, potentially testing H3.x with fewer than 20 blocks despite the registration requiring 20. Fix: register that each confirmatory comparison requires 20 finite paired observations; otherwise rerun the block under the registered process or declare that hypothesis unavailable—no pairwise deletion or imputation. Make `--strict`/campaign enforce `n == 20`.

MINOR

- [hypotheses-tn3.md:20](<D:/Đồ án tốt nghiệp/hybridgrid/docs/thesis/hypotheses-tn3.md:20>): timeout evidence is stale/incomplete. The post-fix pilot maximum is 192.797s, and the relevant drift failure was SED-P90 2/2 before `HG_TIMEOUT`, then zero after. Replace the 113s/one-failure wording with the actual before/after evidence.

- [campaign_tn3.sh:198](<D:/bench-work/campaign_tn3.sh:198>): plain `linucb` is correctly excluded from default Family A, but the registered exploratory LinUCB-vs-HG comparison is never generated. Add an explicitly labeled exploratory analyzer invocation, or state that it will be produced post hoc with fixed arguments.

- [main.go:322](<D:/Đồ án tốt nghiệp/hybridgrid/cmd/hg-coord/main.go:322>): help says `--request-timeout` includes queueing, but the timeout begins only inside `forwardCompile` at [grpc.go:1297](<D:/Đồ án tốt nghiệp/hybridgrid/internal/coordinator/server/grpc.go:1297>). Change the help text to “worker RPC deadline”; queueing remains separately bounded by `dispatch-queue-timeout`.

- [ke-hoach-thi-nghiem-3.md:73](<D:/Đồ án tốt nghiệp/hybridgrid/docs/thesis/ke-hoach-thi-nghiem-3.md:73>) and [line 85](<D:/Đồ án tốt nghiệp/hybridgrid/docs/thesis/ke-hoach-thi-nghiem-3.md:85>) still say Abseil has 156 files; the runner, preregistration, and pilot establish 152. Update or mark the plan superseded. Also remove the stale old-branch comment at [campaign_tn3.sh:14](<D:/bench-work/campaign_tn3.sh:14>).

Verified: with `REQUEST_TIMEOUT`/`HG_TIMEOUT` unset, the coordinator remains at 120s, client at 2m, worker at 5m, and the success reward is numerically identical to Experiments 1–2. Family A is 2 treatments × 4 baselines = 8; Family B is 2 × 1 = 2; plain `linucb` belongs to neither default set. Registered metric names match analyzer output. Runner exclusion and block-rerun rules otherwise match the preregistration. Pilot-informed thresholds/timeouts and the TN2-informed H3.6 arm are disclosed; no additional outcome-dependent hypothesis choice was found.

NOT READY.