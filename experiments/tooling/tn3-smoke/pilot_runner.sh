#!/bin/bash
# TN3 pilot: real runner, one LeastLoaded cell per new workload. Pilot data only.
. /d/bench-work/env.sh
cd "/d/Đồ án tốt nghiệp/hybridgrid" || exit 1
P=/d/bench-work/out/pilot-tn3
for w in absl syn; do
  rm -rf "$P/runner_$w"
  echo "$(date '+%F %T') START $w"
  WORKLOAD=$w SYN_MEDIAN_UNITS=200 SYN_MAX_UNITS=7500 ARMS=leastloaded REPS=1 SESSION_BUILDS=1 JOBS=8 \
    BUILDER_CPUS=2.0 BUILDER_MEM=2G OUT_DIR="$P/runner_$w" bash scripts/bench_exp.sh > "$P/runner_$w.log" 2>&1
  echo "$(date '+%F %T') END $w rc=$?"; tail -n 3 "$P/runner_$w.log"; cat "$P/runner_$w/results.csv" 2>/dev/null
done
