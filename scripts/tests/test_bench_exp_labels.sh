#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
RUNNER="$ROOT/scripts/bench_exp.sh"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
cat > "$TMP/docker" <<'SH'
#!/usr/bin/env bash
printf 'docker was called\n' >> "$DOCKER_CALL_LOG"
exit 73
SH
chmod +x "$TMP/docker"
export DOCKER_CALL_LOG="$TMP/docker-calls"
export PATH="$TMP:$PATH"

run_valid() {
    env -u DRIFT_AT DRY_RUN=1 REPS=2 OUT_DIR="$TMP/out" ARMS="$1" bash "$RUNNER" > "$TMP/stdout" 2> "$TMP/stderr" || {
        cat "$TMP/stderr" >&2
        exit 1
    }
}
run_invalid() {
    local arms=$1 expected=$2 drift_at=${3:-120}
    if DRY_RUN=1 REPS=1 DRIFT_AT="$drift_at" ARMS="$arms" bash "$RUNNER" > "$TMP/stdout" 2> "$TMP/stderr"; then
        printf 'expected rejection of ARMS=%s DRIFT_AT=%s\n' "$arms" "$drift_at" >&2
        exit 1
    fi
    grep -Fq -- "$expected" "$TMP/stderr" || { cat "$TMP/stderr" >&2; exit 1; }
}
assert_output() {
    grep -Fq -- "$1" "$TMP/stdout" || { printf 'missing: %s\n' "$1" >&2; cat "$TMP/stdout" >&2; exit 1; }
}

P90_THRESHOLD=1024 run_valid 'leastloaded sed p90 sed-p90 hybrid-linucb-l0 hybrid-linucb-d-arm-g095-l025 hybrid-linucb-d-g098-l100'
assert_output 'arm leastloaded: hg-coord --scheduler=leastloaded'
assert_output 'arm p90: hg-coord --scheduler=p90'
assert_output 'arm sed-p90: hg-coord --scheduler=sed-p90'
assert_output '--size-threshold=1024 --task-log=/tmp/tasks.jsonl'
assert_output '--load-penalty=0.00 --task-log=/tmp/tasks.jsonl'
assert_output 'arm hybrid-linucb-d-arm-g095-l025: hg-coord --scheduler=hybrid-linucb-d'
assert_output '--load-penalty=0.25 --discount=0.95 --discount-mode=arm --task-log=/tmp/tasks.jsonl'
assert_output '--load-penalty=1.00 --discount=0.98 --discount-mode=global --task-log=/tmp/tasks.jsonl'
assert_output 'round 1:'
assert_output 'round 2:'
assert_output 'docker-compose-exp.yml'
assert_output 'configure: cd /workspace/cpython && ./configure --disable-test-modules --disable-perf-trampoline'
assert_output 'build: cd /workspace/cpython && hgbuild make -j"$1"'
assert_output 'meta keys: WORKLOAD ABSL_REF ABSL_COMMIT'
assert_output 'BUILDER_MEM P90_THRESHOLD'
if grep -F 'arm leastloaded:' "$TMP/stdout" | grep -Fq -- '--size-threshold'; then
    printf 'leastloaded received a P90 threshold\n' >&2
    exit 1
fi
[[ ! -e $DOCKER_CALL_LOG ]] || { printf 'DRY_RUN called docker\n' >&2; exit 1; }

CLIENTS=2 DRY_RUN=1 REPS=1 OUT_DIR="$TMP/out" ARMS=leastloaded bash "$RUNNER" > "$TMP/stdout" 2> "$TMP/stderr"
[[ $(grep -Fc -- '- gate:/gate' "$ROOT/test/stress/docker-compose-exp.yml") == 2 ]] || {
    printf 'both builders must mount the shared gate volume\n' >&2
    exit 1
}
grep -Fq '  gate:' "$ROOT/test/stress/docker-compose-exp.yml" || {
    printf 'shared gate volume missing\n' >&2
    exit 1
}
[[ ! -e $DOCKER_CALL_LOG ]] || { printf 'DRY_RUN called docker\n' >&2; exit 1; }

run_invalid 'hybrid-linucb-g095' 'require hybrid-linucb-d'
run_invalid 'hybrid-linucb-arm' 'require hybrid-linucb-d'
run_invalid 'leastloaded-l025' '-l requires'
run_invalid 'hybrid-linucb-d-z123' 'invalid arm'
run_invalid 'hybrid-linucb-d-g098-arm' 'invalid arm'
run_invalid 'hybrid-linucb-d-l501' 'load penalty exceeds 5.0'
run_invalid 'hybrid-linucb-d-g101' 'discount must be in (0, 1]'
run_invalid 'hybrid-linucb-d-g000' 'discount must be in (0, 1]'
run_invalid 'leastloaded' 'DRIFT_AT must be >= 101' 100
run_invalid 'p90' 'P90_THRESHOLD is required'
run_invalid 'sed-p90' 'P90_THRESHOLD is required'
if P90_THRESHOLD=0 DRY_RUN=1 REPS=1 ARMS=p90 bash "$RUNNER" > "$TMP/stdout" 2> "$TMP/stderr"; then
    printf 'expected rejection of P90_THRESHOLD=0\n' >&2
    exit 1
fi
grep -Fq 'P90_THRESHOLD must be a positive integer' "$TMP/stderr" || { cat "$TMP/stderr" >&2; exit 1; }
printf 'bench_exp label tests passed\n'
