#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
RUNNER="$ROOT/scripts/bench_exp.sh"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

invalid() {
    local value=$1 expected=$2
    if DRY_RUN=1 ARMS=leastloaded REPS=1 DRIFT_AT=101 DRIFT_ON_MAX="$value" \
        bash "$RUNNER" > "$TMP/out" 2> "$TMP/err"; then
        printf 'expected DRIFT_ON_MAX=%s rejection\n' "$value" >&2; exit 1
    fi
    grep -Fq "$expected" "$TMP/err" || { cat "$TMP/err" >&2; exit 1; }
}
invalid '' 'DRIFT_ON_MAX must be a positive integer'
invalid 0 'DRIFT_ON_MAX must be a positive integer'
invalid -1 'DRIFT_ON_MAX must be a positive integer'
invalid abc 'DRIFT_ON_MAX must be a positive integer'
invalid 100 'DRIFT_ON_MAX must be >= DRIFT_AT'

# Exercise the actual default assignment and drift_on boundary expression.
DRIFT_AT=101
unset DRIFT_ON_MAX
source <(awk '/^DRIFT_ON_MAX="\$\{DRIFT_ON_MAX-/ {print; exit}' "$RUNNER")
[[ $DRIFT_ON_MAX == 141 ]]
eval "$(awk '/first_on <= DRIFT_ON_MAX/ {print "check_drift_max() {"; print "local first_on=$1"; print; print "}"; exit}' "$RUNNER")"
FAIL_REASON=
check_drift_max 141
if check_drift_max 142; then printf 'expected default drift limit rejection\n' >&2; exit 1; fi
[[ $FAIL_REASON == 'drift_on too late: dispatch_count=142 exceeds 141' ]]
DRIFT_ON_MAX=152
FAIL_REASON=
check_drift_max 152
if check_drift_max 153; then printf 'expected configured drift limit rejection\n' >&2; exit 1; fi
[[ $FAIL_REASON == 'drift_on too late: dispatch_count=153 exceeds 152' ]]
printf 'bench_exp drift_on max tests passed\n'
