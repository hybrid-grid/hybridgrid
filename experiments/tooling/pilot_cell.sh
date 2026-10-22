#!/bin/bash
# Exploratory single cell (NOT an official result): Exp 1 shape, 1 client, -j8,
# one arm. Switches the power plan for the run and always restores it.
. /d/bench-work/env.sh
OUT=$BENCH_OUT/pilot-cell1
mkdir -p "$OUT"
LOG=$OUT/run.log
restore() { powershell.exe -NoProfile -File 'D:\bench-work\power.ps1' off >> "$LOG" 2>&1; echo "$(date +%T) power restored" >> "$LOG"; }
trap restore EXIT
powershell.exe -NoProfile -File 'D:\bench-work\power.ps1' on >> "$LOG" 2>&1
echo "$(date +%T) START" >> "$LOG"
bash /d/bench-work/preflight.sh "$OUT/preflight.txt" > /dev/null 2>&1
cd /d/bench-work/hybridgrid || exit 1
REPS=1 ARMS="leastloaded" JOBS=8 CLIENTS=1 OUT_DIR="$OUT" bash scripts/bench_exp.sh 2>&1 | while IFS= read -r l; do printf '%s %s\n' "$(date +%T)" "$l"; done >> "$LOG"
echo "$(date +%T) EXIT=${PIPESTATUS[0]}" >> "$LOG"
