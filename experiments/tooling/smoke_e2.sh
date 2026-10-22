#!/bin/bash
# Smoke test S1-S4 for Experiment 2 (plan section 3.2 item 3). Pilot data, never official.
. /d/bench-work/env.sh
cd "$BENCH_REPO" || exit 1
P=$BENCH_OUT/pilot-e2
mkdir -p "$P"
run() { # name drift arm
    local name=$1 drift=$2 arm=$3
    rm -rf "$P/$name"
    echo "$(date '+%F %T') START $name drift=$drift arm=$arm"
    ARMS="$arm" REPS=1 JOBS=8 CLIENTS=1 SESSION_BUILDS=2 DRIFT=$drift BUILDER_CPUS=2.0 \
        OUT_DIR="$P/$name" bash scripts/bench_exp.sh > "$P/$name.log" 2>&1
    echo "$(date '+%F %T') END $name rc=$?"
}
run S1 permanent hybrid-linucb-d-arm-g098
run S2 transient hybrid-linucb-d-g099
run S3 onoff hybrid-linucb-d-g095
run S4 permanent hybrid-linucb-d-g098-l025
echo "$(date '+%F %T') ALL DONE"
