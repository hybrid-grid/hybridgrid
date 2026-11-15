#!/bin/bash
# Exploratory drift cells (NOT official results): transient drift, 1 client, -j8,
# one arm, with SESSION_BUILDS=1 and then 2. Power plan is switched for the run
# and always restored.
. /d/bench-work/env.sh
BASE=$BENCH_OUT/pilot-drift
mkdir -p "$BASE"
LOG=$BASE/run.log
restore() { powershell.exe -NoProfile -File 'D:\bench-work\power.ps1' off >> "$LOG" 2>&1; echo "$(date +%T) power restored" >> "$LOG"; }
trap restore EXIT
powershell.exe -NoProfile -File 'D:\bench-work\power.ps1' on >> "$LOG" 2>&1
cd /d/bench-work/hybridgrid || exit 1
for sb in 1 2; do
    OUT=$BASE/trans_s$sb
    mkdir -p "$OUT"
    echo "$(date +%T) START transient SESSION_BUILDS=$sb" >> "$LOG"
    bash /d/bench-work/preflight.sh "$OUT/preflight.txt" > /dev/null 2>&1
    REPS=1 ARMS="leastloaded" JOBS=8 CLIENTS=1 DRIFT=transient SESSION_BUILDS=$sb OUT_DIR="$OUT" bash scripts/bench_exp.sh 2>&1 | while IFS= read -r l; do printf '%s %s\n' "$(date +%T)" "$l"; done >> "$LOG"
    echo "$(date +%T) EXIT s$sb=${PIPESTATUS[0]}" >> "$LOG"
done
echo "$(date +%T) ALL DONE" >> "$LOG"
