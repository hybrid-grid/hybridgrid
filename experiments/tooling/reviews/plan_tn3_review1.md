Read-only review completed; no files changed and Docker was not used.

Direct answers:

- `hgbuild ninja` does **not currently guarantee distributed CMake builds**. For a normally configured CMake tree it will usually execute the compiler already embedded in `build.ninja`; setting `CC/CXX` when Ninja starts is too late.
- `SourceSizeBytes` **is available before worker selection** for successfully preprocessed requests.
- Local preprocessing on the builder is a serious, currently unmeasured bottleneck risk.
- LinUCB **does have an implicit size×worker interaction** because it is disjoint and learns a separate size coefficient per worker. It has no shared cross-worker interaction model, however, and the size signal is capped.

## BLOCKER

- **CMake/Ninja will likely compile locally, making all scheduler comparisons meaningless.** The wrapper only sets `CC/CXX` when Ninja is launched ([main.go:1234](<D:/Đồ án tốt nghiệp/hybridgrid/cmd/hgbuild/main.go:1234>)); CMake has already frozen compiler commands during configuration. Its temporary wrappers are also deleted after the wrapping command exits ([main.go:1215](<D:/Đồ án tốt nghiệp/hybridgrid/cmd/hgbuild/main.go:1215>)). Worse, standard CMake/Ninja commands contain `-MD -MT <object> -MF <depfile>`; the parser treats the `-MT` object operand as an additional input ([parser.go:123](<D:/Đồ án tốt nghiệp/hybridgrid/internal/compiler/parser.go:123>), [parser.go:223](<D:/Đồ án tốt nghiệp/hybridgrid/internal/compiler/parser.go:223>)), so `IsDistributable` rejects the command because it sees more than one input ([parser.go:148](<D:/Đồ án tốt nghiệp/hybridgrid/internal/compiler/parser.go:148>)).  
  **Fix:** install persistent `hgcc`/`hgcxx` wrappers, configure CMake with those wrappers, correctly parse paired GCC options, and preserve/generate Ninja depfiles locally. Add a mandatory smoke gate: expected TU count equals coordinator task count, zero local compile/fallback, successful clean link.

- **The plan explicitly drops two advisor-required projects.** It schedules only abseil ([plan:34](<D:/Đồ án tốt nghiệp/hybridgrid/docs/thesis/ke-hoach-thi-nghiem-3.md:34>)) and proposes omitting protobuf and OpenCV ([plan:142](<D:/Đồ án tốt nghiệp/hybridgrid/docs/thesis/ke-hoach-thi-nghiem-3.md:142>)), contrary to abseil→protobuf→OpenCV.  
  **Fix:** define pinned commits, exact CMake options/targets, and 20-block scenarios for all three projects. Negative abseil results are not a reason to conditionally decide whether later required projects run.

## MAJOR

- **P90 cannot currently receive `TaskContext` through the ordinary scheduler interface.** Size is correctly constructed before selection ([grpc.go:866](<D:/Đồ án tốt nghiệp/hybridgrid/internal/coordinator/server/grpc.go:866>)), but `SelectWith` passes context only to `LearningScheduler`; normal schedulers receive only the old `Select` call ([feedback.go:87](<D:/Đồ án tốt nghiệp/hybridgrid/internal/coordinator/scheduler/feedback.go:87>)).  
  **Fix:** introduce a contextual non-learning scheduler interface, or explicitly implement and test a no-op feedback adapter for P90/SED-P90. Reject cells containing raw-mode fallback because their “size” is raw rather than preprocessed.

- **Builder preprocessing may erase scheduler differences.** Every wrapper preprocesses locally before contacting the coordinator ([build.go:184](<D:/Đồ án tốt nghiệp/hybridgrid/internal/cli/build/build.go:184>)). The existing campaign defaults to one builder CPU and 1 GiB RAM ([bench_exp.sh:15](<D:/Đồ án tốt nghiệp/hybridgrid/scripts/bench_exp.sh:15>), [bench_exp.sh:168](<D:/Đồ án tốt nghiệp/hybridgrid/scripts/bench_exp.sh:168>)), despite the plan assuming two CPUs. `PreprocessTime` exists but is not populated.  
  **Fix:** freeze builder CPU/RAM explicitly, log per-TU preprocessing duration, fallback and OOM counts, and pilot 2 versus 4 builder CPUs. Pre-register a bottleneck gate based on builder saturation, dispatch gaps, and worker idle time—not a subjective inspection.

- **LinUCB’s interaction exists, but its useful size range may collapse.** Separate per-worker arms and parameters provide worker-specific size slopes ([linucb.go:27](<D:/Đồ án tốt nghiệp/hybridgrid/internal/coordinator/scheduler/linucb.go:27>), [linucb.go:116](<D:/Đồ án tốt nghiệp/hybridgrid/internal/coordinator/scheduler/linucb.go:116>)). However, preprocessed size is capped at 4 MiB ([linucb.go:835](<D:/Đồ án tốt nghiệp/hybridgrid/internal/coordinator/scheduler/linucb.go:835>)); all larger heavy TUs become identical. The raw-size feature is zero in normal preprocessed mode.  
  **Fix:** pilot-report the fraction of TUs hitting the cap. Freeze a revised normalization before registration if the top tail saturates. Do not claim cross-worker generalization: each worker must independently observe enough small and large TUs.

- **No pure LinUCB arm, and warm-start dominates abseil.** The scenario rules call for pure LinUCB alongside the required four schedulers ([scenario:119](<D:/Đồ án tốt nghiệp/hybridgrid/docs/thesis/phan-tich-leastloaded-va-kich-ban.md:119>)); the plan omits it. The runner defaults to 100 LL warm-start dispatches ([bench_exp.sh:23](<D:/Đồ án tốt nghiệp/hybridgrid/scripts/bench_exp.sh:23>)), consuming roughly 64% of the claimed 156-TU abseil build.  
  **Fix:** add `linucb` with `N=0, λ=0`; preferably pre-register the proposed 2×2 `(N,λ)` ablation as secondary. For required HG arms, use a discarded training build followed by a measured clean build under the same coordinator, or justify why only ~56 post-warm-start dispatches are adequate.

- **The pilot gate is internally inconsistent and too weak.** It says one cell per workload but asks whether an event occurs in “most cells” ([plan:92](<D:/Đồ án tốt nghiệp/hybridgrid/docs/thesis/ke-hoach-thi-nghiem-3.md:92>), [plan:99](<D:/Đồ án tốt nghiệp/hybridgrid/docs/thesis/ke-hoach-thi-nghiem-3.md:99>)). It also lacks gates for distributed fraction, feature saturation, candidate choice, post-drift heavy TUs, fallback and OOM.  
  **Fix:** use several discarded LL pilots and require, before registration: 100% intended TUs remotely dispatched, zero fallback/failure/OOM, acceptable preprocessing share, adequate size–time correlation, enough top-tail dispatches with at least two eligible workers, and enough P90/top-5% TUs after actual drift onset.

- **The drift hypothesis is ambiguous and stronger than TN2 supports.** H3.5 says “HG-LinUCB (or HG-LinUCB-D)” ([plan:69](<D:/Đồ án tốt nghiệp/hybridgrid/docs/thesis/ke-hoach-thi-nghiem-3.md:69>)). TN2 found that LL self-adjusted and all schedulers sent roughly capacity share to the slowed worker ([TN2 report:34](<D:/Đồ án tốt nghiệp/hybridgrid/docs/thesis/bao-cao-tn2-261021.md:34>)); arm-local D was not significantly different from HG.  
  **Fix:** choose one primary drift treatment and fixed γ/mode before registration, state that selection used TN2, and require sufficient heavy TUs after the recorded—not nominal—drift onset.

- **Required tail time is not yet robustly recoverable from analysis inputs.** The runner records absolute start/end in build logs ([bench_exp.sh:479](<D:/Đồ án tốt nghiệp/hybridgrid/scripts/bench_exp.sh:479>)) but writes only elapsed duration to `results.csv` ([bench_exp.sh:543](<D:/Đồ án tốt nghiệp/hybridgrid/scripts/bench_exp.sh:543>)). The analyzer currently reports the last finisher, not `build_end − last_dispatch` ([analyze_exp.py:470](<D:/Đồ án tốt nghiệp/hybridgrid/scripts/analyze_exp.py:470>)).  
  **Fix:** store `start_ts` and `end_ts` in results, join by `build_id`, and define tail as `end_ts − max(dispatch_ts)` over successful intended TUs.

- **Pre-registration and multiplicity are underspecified.** H3.1–H3.3 span unspecified scenarios; H3.5 contains an “or”; directions and Holm families are absent. This conflicts with the scenario rules ([scenario:214](<D:/Đồ án tốt nghiệp/hybridgrid/docs/thesis/phan-tich-leastloaded-va-kich-ban.md:214>)). The analyzer always runs two-sided Wilcoxon ([analyze_exp.py:101](<D:/Đồ án tốt nghiệp/hybridgrid/scripts/analyze_exp.py:101>)) and builds automatic per-metric treatment×baseline Holm families ([analyze_exp.py:486](<D:/Đồ án tốt nghiệp/hybridgrid/scripts/analyze_exp.py:486>)).  
  **Fix:** pre-register exact scenario/contrast/metric triples, test direction, Holm families, minimum meaningful effect, exclusion/rerun rules, and fixed-sequence logic. Aggregate TU metrics to one value per build; never treat TUs as independent replicates.

- **Current runner/analyzer reject the new arms and metrics.** Their accepted scheduler lists contain neither `p90` nor `sed-p90` ([bench_exp.sh:67](<D:/Đồ án tốt nghiệp/hybridgrid/scripts/bench_exp.sh:67>), [analyze_exp.py:20](<D:/Đồ án tốt nghiệp/hybridgrid/scripts/analyze_exp.py:20>)); the metric registry lacks top-5% and tail-time outputs ([analyze_exp.py:25](<D:/Đồ án tốt nghiệp/hybridgrid/scripts/analyze_exp.py:25>)).  
  **Fix:** include arm-label parsing, metadata validation, metric tests, per-workload P90 values, and strict preflight checks in the implementation plan.

## MINOR

- **P90 provenance is not reproducible enough.** “Calculated from a pilot build” ([plan:46](<D:/Đồ án tốt nghiệp/hybridgrid/docs/thesis/ke-hoach-thi-nghiem-3.md:46>)) does not define workload-specific thresholds, percentile estimator, ties, strict `>` behavior, or the TU universe.  
  **Fix:** freeze one threshold per pinned workload, its exact byte value, quantile method, source manifest hash, and pilot artifact; reuse the synthetic threshold for synthetic+drift.

- **“Strongest worker” and top-5% need precise definitions.** Use registered `cpu_millis`, specify whether strength remains static during hidden drift, define tie handling, and report how often top-tail decisions had multiple eligible workers.

## Time estimate

The 27–30 hour arithmetic is internally correct only for the reduced 360-cell, six-arm plan ([plan:51](<D:/Đồ án tốt nghiệp/hybridgrid/docs/thesis/ke-hoach-thi-nghiem-3.md:51>), [plan:112](<D:/Đồ án tốt nghiệp/hybridgrid/docs/thesis/ke-hoach-thi-nghiem-3.md:112>)). It is not credible for the advisor-compliant campaign. Adding pure LinUCB and both omitted projects raises the campaign to roughly 700 cells; using the plan’s own 15–25 hours per added project gives approximately **60–85 machine-hours before reruns**. The 5–7 preparation days also omit the CMake wrapper/parser/depfile work. Re-estimate only after end-to-end smoke tests and report pilot median/P90 cell duration rather than a pre-pilot ±25% bound.

NOT READY