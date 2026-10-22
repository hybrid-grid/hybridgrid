#!/bin/bash
# Exploratory cells for decision C (NOT official results):
#   2c_b10 : 2 clients, -j8, builder 1.0 CPU  (the Exp 1 two-client scenario)
#   1c_b20 : 1 client,  -j8, builder 2.0 CPU  (compare with pilot-cell1, builder 1.0)
# Power plan is switched for the run and always restored.
. /d/bench-work/env.sh
BASE=$BENCH_OUT/pilot-c
mkdir -p "$BASE"
LOG=$BASE/run.log
restore() { powershell.exe -NoProfile -File 'D:\bench-work\power.ps1' off >> "$LOG" 2>&1; echo "$(date +%T) power restored" >> "$LOG"; }
trap restore EXIT
powershell.exe -NoProfile -File 'D:\bench-work\power.ps1' on >> "$LOG" 2>&1
cd /d/bench-work/hybridgrid || exit 1
run() {
    local name=$1; shift
    local out=$BASE/$name
    mkdir -p "$out"
    echo "$(date +%T) START $name" >> "$LOG"
    bash /d/bench-work/preflight.sh "$out/preflight.txt" > /dev/null 2>&1
    env "$@" REPS=1 ARMS="leastloaded" JOBS=8 OUT_DIR="$out" bash scripts/bench_exp.sh 2>&1 | while IFS= read -r l; do printf '%s %s\n' "$(date +%T)" "$l"; done >> "$LOG"
    echo "$(date +%T) EXIT $name=${PIPESTATUS[0]}" >> "$LOG"
}
run 2c_b10 CLIENTS=2 BUILDER_CPUS=1.0
run 1c_b20 CLIENTS=1 BUILDER_CPUS=2.0
echo "$(date +%T) ALL DONE" >> "$LOG"
