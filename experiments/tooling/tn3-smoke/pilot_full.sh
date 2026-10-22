#!/bin/bash
# TN3 full pilot (pilot data only, never official):
#  1. syn LeastLoaded cell with reduced weight -> build time and P90 threshold
#  2. one block of all 7 arms per workload, 2-build sessions (TN3 protocol)
. /d/bench-work/env.sh
cd "/d/Đồ án tốt nghiệp/hybridgrid" || exit 1
P=/d/bench-work/out/pilot-tn3
SYN="SYN_MEDIAN_UNITS=120 SYN_MAX_UNITS=4500"
ARMS7="leastloaded sed p90 sed-p90 linucb hybrid-linucb hybrid-linucb-d-arm-g098"
common() { env JOBS=8 CLIENTS=1 BUILDER_CPUS=2.0 BUILDER_MEM=2G "$@" bash scripts/bench_exp.sh; }
p90_of() { python3 - "$1" <<'PY'
import json, sys, numpy as np
T=[json.loads(l) for l in open(sys.argv[1]) if 'task_completed' in l]
print(int(np.percentile([t['preprocessed_size_bytes'] for t in T], 90, method='linear')))
PY
}
echo "$(date '+%F %T') step 1: syn LL"
rm -rf "$P/syn_ll"
common $SYN WORKLOAD=syn ARMS=leastloaded REPS=1 SESSION_BUILDS=1 OUT_DIR="$P/syn_ll" > "$P/syn_ll.log" 2>&1 || { echo "syn LL failed"; tail -5 "$P/syn_ll.log"; exit 1; }
cat "$P/syn_ll/results.csv"
SYN_P90=$(p90_of "$P/syn_ll/tasks_leastloaded_round1.jsonl"); ABSL_P90=2356237
echo "syn P90=$SYN_P90 absl P90=$ABSL_P90"
for w in absl syn; do
  thr=$ABSL_P90; [[ $w == syn ]] && thr=$SYN_P90
  echo "$(date '+%F %T') step 2: $w block of 7 arms (P90_THRESHOLD=$thr)"
  rm -rf "$P/block_$w"
  common $SYN WORKLOAD=$w ARMS="$ARMS7" P90_THRESHOLD=$thr REPS=1 SESSION_BUILDS=2 OUT_DIR="$P/block_$w" > "$P/block_$w.log" 2>&1
  echo "$(date '+%F %T') $w rc=$?"; tail -n 2 "$P/block_$w.log"; cat "$P/block_$w/results.csv"; tail -n +2 "$P/block_$w/failures.csv"
done
echo "PILOT DONE"
