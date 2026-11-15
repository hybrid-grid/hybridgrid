#!/bin/bash
# Pause the campaign at the end of the E2-lambda block currently running:
# wait until every arm of block BLOCK has both build rows (exact key set) and a
# non-empty task log, then stop safely and snapshot the state outside
# out/official. LAST_PAUSE is written only when stop and snapshot both succeed.
# Usage: pause_at_block_end.sh <block>
set -u
. /d/bench-work/env.sh
O=$BENCH_OUT/official
S=$O/E2-lambda
P=/d/bench-work/pause_snapshots
BLOCK=${1:?block}
ARMS="leastloaded sed hybrid-linucb hybrid-linucb-l0 hybrid-linucb-l025 hybrid-linucb-l100 hybrid-linucb-d-g098 hybrid-linucb-d-g098-l0 hybrid-linucb-d-g098-l025 hybrid-linucb-d-g098-l100"
complete() {
    local a keys
    keys=$(awk -F, -v b="$BLOCK" 'NR > 1 && $2 == b { print $1 "," $4 "," $5 }' "$S/results.csv" | sort)
    [[ $(sort -u <<< "$keys" | grep -c .) == 20 && $(grep -c . <<< "$keys") == 20 ]] || return 1
    for a in $ARMS; do
        grep -qx "$a,1,1" <<< "$keys" && grep -qx "$a,2,1" <<< "$keys" || return 1
        [[ -s $S/tasks_${a}_round${BLOCK}.jsonl ]] || return 1
    done
}
echo "$(date '+%F %T') waiting for block $BLOCK to complete (10 arms x 2 builds)"
until complete; do
    [[ -e $O/campaign.pid ]] || { echo "campaign no longer running; abort"; exit 1; }
    sleep 3
done
echo "$(date '+%F %T') block $BLOCK complete; stopping"
bash /d/bench-work/stop_campaign.sh "pause requested by the user after block $BLOCK of E2-lambda (resume later after resume_check.sh)"
rc=$?
echo "stop rc=$rc"
(( rc == 0 )) || { echo "stop failed; not writing LAST_PAUSE"; exit 1; }
for i in $(seq 1 30); do [[ -e $O/campaign.pid ]] || break; sleep 1; done
[[ ! -e $O/campaign.pid ]] || { echo "campaign.pid still present; not writing LAST_PAUSE"; exit 1; }
TS=$(date '+%Y%m%d-%H%M%S')
mkdir -p "$P"
bash /d/bench-work/snapshot_state.sh "$P/pause_$TS" || { echo "snapshot failed; not writing LAST_PAUSE"; exit 1; }
printf '%s\n' "$P/pause_$TS" > "$P/LAST_PAUSE.tmp" && mv "$P/LAST_PAUSE.tmp" "$P/LAST_PAUSE"
echo "$(date '+%F %T') paused; snapshot $P/pause_$TS"
