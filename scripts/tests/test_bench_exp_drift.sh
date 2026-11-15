#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
RUNNER="$ROOT/scripts/bench_exp.sh"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

run_invalid() {
    local target=$1 expected=$2
    if DRY_RUN=1 REPS=1 ARMS=leastloaded DRIFT_TARGET="$target" bash "$RUNNER" > "$TMP/stdout" 2> "$TMP/stderr"; then
        printf 'expected rejection of DRIFT_TARGET=%s\n' "$target" >&2
        exit 1
    fi
    grep -Fq -- "$expected" "$TMP/stderr" || { cat "$TMP/stderr" >&2; exit 1; }
}

for target in '' ',worker-4' 'worker-4,' 'worker-4,,worker-5' 'worker-6' \
    'worker-4, worker-5' 'worker-4,worker-5 ' \
    'worker-1,worker-2,worker-3,worker-4,worker-5'; do
    run_invalid "$target" 'ERROR: DRIFT_TARGET must list 1 to 4 workers'
done
run_invalid 'worker-4,worker-4' "ERROR: DRIFT_TARGET contains duplicate worker 'worker-4'"

# Load only the runner's stress functions; no container or full cell is run.
source <(awk '/^stop_stress\(\) \{/ {copy=1} /^cleanup\(\) \{/ {copy=0} copy' "$RUNNER")
source <(awk '/^post_event\(\) \{/ {copy=1} /^inject_drift\(\) \{/ {copy=0} copy' "$RUNNER")

declare -A fake_up=()
DRIFT=permanent
DRIFT_CPU=3
DRIFT_STRESS_ARGS=
events_file="$TMP/events.csv"
cell_marker="$TMP/cell"
CALLS="$TMP/calls"
POSTS="$TMP/posts"
compose() {
    printf '%s\n' "$*" >> "$CALLS"
    local target action
    if [[ $2 == -d ]]; then
        target=$4
        [[ $target != ${FAIL_START:-} ]] || return 1
        fake_up[$target]=1
    else
        target=$3 action=$4
        case $action in
            pgrep) [[ ${fake_up[$target]:-0} == 1 && $target != ${FAIL_CONFIRM:-} ]] ;;
            pkill)
                [[ $target != ${FAIL_STOP:-} ]] || return 1
                fake_up[$target]=0
                ;;
        esac
    fi
}
curl() { cat >> "$POSTS"; printf '\n' >> "$POSTS"; }
dispatch_count() { printf '123\n'; }
now() { printf '9\n'; }
sleep() { :; }
reset_case() {
    : > "$CALLS"
    : > "$POSTS"
    : > "$events_file"
    rm -f "$cell_marker.error"
    fake_up=()
    FAIL_START= FAIL_CONFIRM= FAIL_STOP=
}
assert_calls() {
    printf '%s\n' "$@" > "$TMP/expected"
    diff -u "$TMP/expected" "$CALLS"
}
assert_events() {
    local expected=$1
    diff -u <(printf '%s\n' "$expected") "$events_file"
}

reset_case
DRIFT_TARGET=worker-5
drift_targets=(worker-5)
start_stress 101
stop_stress_phase 130
assert_calls \
    'exec -d -T worker-5 stress-ng --cpu 3' \
    'exec -T worker-5 pgrep -x stress-ng' \
    'exec -T worker-5 pgrep -x stress-ng' \
    'exec -T worker-5 pkill -x stress-ng'
assert_events $'9,drift_on,worker-5:50051,123,stress_started trigger=101\n9,drift_off,worker-5:50051,123,stress_stopped trigger=130'
grep -Fq '"target": "worker-5:50051"' "$POSTS"
[[ $(grep -c '"kind"' "$POSTS") == 2 ]]

reset_case
DRIFT_TARGET=worker-4,worker-5
drift_targets=(worker-4 worker-5)
start_stress 101
stop_stress_phase 130
assert_calls \
    'exec -d -T worker-4 stress-ng --cpu 3' \
    'exec -T worker-4 pgrep -x stress-ng' \
    'exec -d -T worker-5 stress-ng --cpu 3' \
    'exec -T worker-5 pgrep -x stress-ng' \
    'exec -T worker-4 pgrep -x stress-ng' \
    'exec -T worker-4 pkill -x stress-ng' \
    'exec -T worker-5 pgrep -x stress-ng' \
    'exec -T worker-5 pkill -x stress-ng'
assert_events $'9,drift_on,worker-4:50051|worker-5:50051,123,stress_started trigger=101\n9,drift_off,worker-4:50051|worker-5:50051,123,stress_stopped trigger=130'
[[ $(grep -c '"kind"' "$POSTS") == 2 ]]

reset_case
DRIFT=onoff
start_stress 101
stop_stress_phase 130
start_stress 140
stop_stress
assert_events $'9,drift_on,worker-4:50051|worker-5:50051,123,stress_started trigger=101\n9,drift_off,worker-4:50051|worker-5:50051,123,stress_stopped trigger=130\n9,drift_on,worker-4:50051|worker-5:50051,123,stress_started trigger=140\n9,drift_off,worker-4:50051|worker-5:50051,123,cell_end'
DRIFT=permanent

reset_case
FAIL_START=worker-5
if start_stress 101; then printf 'expected start failure\n' >&2; exit 1; fi
assert_calls \
    'exec -d -T worker-4 stress-ng --cpu 3' \
    'exec -T worker-4 pgrep -x stress-ng' \
    'exec -d -T worker-5 stress-ng --cpu 3' \
    'exec -T worker-4 pkill -x stress-ng' \
    'exec -T worker-5 pkill -x stress-ng'
[[ ${fake_up[worker-4]} == 0 && ! -s $events_file && ! -s $POSTS ]]

reset_case
FAIL_CONFIRM=worker-5
if start_stress 101; then printf 'expected confirmation failure\n' >&2; exit 1; fi
[[ $(grep -Fc 'exec -T worker-5 pgrep -x stress-ng' "$CALLS") == 10 ]]
[[ $(tail -n 2 "$CALLS") == $'exec -T worker-4 pkill -x stress-ng\nexec -T worker-5 pkill -x stress-ng' ]]
[[ ${fake_up[worker-4]} == 0 && ${fake_up[worker-5]} == 0 && ! -s $events_file ]]

reset_case
start_stress 101
FAIL_STOP=worker-4
if stop_stress_phase 130; then printf 'expected phase stop failure\n' >&2; exit 1; fi
[[ ${fake_up[worker-5]} == 0 && $(grep -Fc 'exec -T worker-5 pkill -x stress-ng' "$CALLS") == 1 ]]
[[ $(wc -l < "$events_file") == 1 ]]

reset_case
start_stress 101
FAIL_STOP=worker-4
if stop_stress; then printf 'expected cleanup failure\n' >&2; exit 1; fi
[[ ${fake_up[worker-5]} == 0 && $(grep -Fc 'exec -T worker-5 pkill -x stress-ng' "$CALLS") == 1 ]]
grep -Fxq 'stress-ng cleanup failed' "$cell_marker.error"

reset_case
DRIFT_TARGET=worker-4,worker-5
drift_targets=(worker-4 worker-5)
DRIFT_STRESS_ARGS='--vm 1 --vm-bytes 600M --vm-keep'
start_stress 101
DRIFT_STRESS_ARGS=
grep -Fxq 'exec -d -T worker-4 stress-ng --vm 1 --vm-bytes 600M --vm-keep' "$CALLS"
grep -Fxq 'exec -d -T worker-5 stress-ng --vm 1 --vm-bytes 600M --vm-keep' "$CALLS"
! grep -Fq -- '--cpu' "$CALLS"

printf 'bench_exp drift tests passed\n'
