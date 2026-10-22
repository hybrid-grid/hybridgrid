#!/bin/bash
# Request a clean stop, then tear down and verify the hgexp compose project.
set -u
. /d/bench-work/env.sh
CAMPAIGN=${CAMPAIGN:-$BENCH_OUT/official}
DOCKER=${DOCKER:-docker}
PROC_ROOT=${PROC_ROOT:-/proc}
if (( $# != 1 )) || [[ -z $1 || $1 == *$'\n'* || $1 == *$'\r'* ]]; then
    echo 'usage: stop_campaign.sh "<reason>" (one non-empty line)' >&2
    exit 2
fi
mkdir -p "$CAMPAIGN"
touch "$CAMPAIGN/STOP_REQUESTED" || exit 1
printf '%s MANUAL STOP %s\n' "$(date '+%F %T')" "$1" >> "$CAMPAIGN/campaign.log"
valid_pid() {
    local file=$1 expected=$2 value cmdline
    [[ -e $file ]] || return 1
    value=$(cat "$file" 2>/dev/null || true)
    if [[ ! $value =~ ^[0-9]+$ ]] || ! kill -0 "$value" 2>/dev/null || [[ ! -r $PROC_ROOT/$value/cmdline ]]; then
        rm -f "$file"
        return 1
    fi
    cmdline=$(tr '\0' ' ' < "$PROC_ROOT/$value/cmdline")
    if ! grep -Eq "(^|[[:space:]/])${expected}([[:space:]]|$)" <<< "$cmdline"; then
        rm -f "$file"
        return 1
    fi
    printf '%s\n' "$value"
}
runner=$(valid_pid "$CAMPAIGN/runner.pid" 'bench_exp\.sh' || true)
if [[ -n $runner ]]; then kill -TERM "$runner" 2>/dev/null || true; fi
deadline=$((SECONDS + 120))
campaign=$(valid_pid "$CAMPAIGN/campaign.pid" 'campaign\.sh' || true)
while [[ -n $campaign && SECONDS -lt deadline ]]; do
    sleep 1
    campaign=$(valid_pid "$CAMPAIGN/campaign.pid" 'campaign\.sh' || true)
done
if [[ -n $campaign ]]; then
    runner=$(valid_pid "$CAMPAIGN/runner.pid" 'bench_exp\.sh' || true)
    [[ -z $runner ]] || kill -9 "$runner" 2>/dev/null || true
    for proc in "$PROC_ROOT"/[0-9]*/cmdline; do
        [[ -r $proc ]] || continue
        descendant=${proc%/cmdline}; descendant=${descendant##*/}
        [[ $descendant =~ ^[0-9]+$ ]] || continue
        [[ $descendant == "$campaign" ]] && continue
        [[ -z $runner || $descendant != "$runner" ]] || continue
        kill -0 "$descendant" 2>/dev/null || continue
        cmdline=$(tr '\0' ' ' < "$proc")
        grep -Eq '(^|[[:space:]/])bench_exp\.sh([[:space:]]|$)' <<< "$cmdline" || continue
        kill -9 "$descendant" 2>/dev/null || true
    done
    kill -9 "$campaign" 2>/dev/null || true
    echo 'stop timeout: forced campaign and runner termination' >&2
fi
status=0
compose_file="$BENCH_REPO/test/stress/docker-compose-exp.yml"
if [[ -f $compose_file ]]; then
    "$DOCKER" compose -p hgexp -f "$compose_file" down --remove-orphans || status=1
else
    echo "FAIL: compose file missing: $compose_file" >&2
    status=1
fi
containers=$("$DOCKER" ps -aq --filter label=com.docker.compose.project=hgexp) || status=1
if [[ -z $containers && $status == 0 ]]; then
    echo 'OK: hgexp containers absent'
else
    echo "FAIL: hgexp cleanup incomplete (containers: ${containers:-<query failed or none>})" >&2
    status=1
fi
exit "$status"
