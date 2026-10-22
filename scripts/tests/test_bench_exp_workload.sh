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
    env -u WORKLOAD -u BUILDER_MEM -u CONFIGURE_FLAGS -u DRIFT_AT \
        DRY_RUN=1 REPS=1 ROUND_START=1 ARMS=leastloaded COOLDOWN=8 SESSION_BUILDS=1 EST_CELL_S=60 \
        ALPHA=0.5 EPSILON=0.1 WARM_START=100 OUT_DIR="$TMP/out" "$@" \
        bash "$RUNNER" > "$TMP/stdout" 2> "$TMP/stderr" || {
        cat "$TMP/stderr" >&2
        exit 1
    }
    [[ ! -e $DOCKER_CALL_LOG ]] || { printf 'DRY_RUN called docker\n' >&2; exit 1; }
}
assert_output() {
    grep -Fq -- "$1" "$TMP/stdout" || { printf 'missing: %s\n' "$1" >&2; cat "$TMP/stdout" >&2; exit 1; }
}
run_invalid() {
    local name=$1 value=$2 expected=$3
    if env -u WORKLOAD DRY_RUN=1 REPS=1 ROUND_START=1 ARMS=leastloaded DRIFT_AT=120 \
        SYN_MEDIAN_UNITS=40 SYN_MAX_UNITS=1500 OUT_DIR="$TMP/out" \
        "$name=$value" bash "$RUNNER" > "$TMP/stdout" 2> "$TMP/stderr"; then
        printf 'expected rejection of %s=%s\n' "$name" "$value" >&2
        exit 1
    fi
    grep -Fq -- "$expected" "$TMP/stderr" || { cat "$TMP/stderr" >&2; exit 1; }
    [[ ! -e $DOCKER_CALL_LOG ]] || { printf 'invalid DRY_RUN called docker\n' >&2; exit 1; }
}

run_valid
# Snapshot of the complete DRY_RUN stdout from 1b8339d for these inputs,
# apart from the machine-specific compose path. New lines follow this prefix.
cat > "$TMP/exp12.snapshot" <<'SNAPSHOT'
compose: <COMPOSE>
arm leastloaded: hg-coord --scheduler=leastloaded --alpha=0.5 --epsilon=0.1 --warm-start=100 --load-penalty=0.5 --task-log=/tmp/tasks.jsonl
round 1: leastloaded
estimated wall time: 136 s (2 cells x 1 builds x 60 s, plus cooldown; excludes setup)
SNAPSHOT
sed -n '1,4p' "$TMP/stdout" | sed '1s|^compose: .*|compose: <COMPOSE>|' > "$TMP/actual.snapshot"
diff -u "$TMP/exp12.snapshot" "$TMP/actual.snapshot"
[[ $(wc -l < "$TMP/stdout") == 7 ]] || { printf 'unexpected CPython DRY_RUN lines\n' >&2; cat "$TMP/stdout" >&2; exit 1; }
assert_output 'configure: cd /workspace/cpython && ./configure --disable-test-modules --disable-perf-trampoline'
assert_output 'build: cd /workspace/cpython && hgbuild make -j"$1"'
assert_output 'meta keys: WORKLOAD ABSL_REF ABSL_COMMIT SYN_SEED SYN_FILES SYN_C_FRACTION SYN_ALPHA SYN_MEDIAN_UNITS SYN_MAX_UNITS SYN_HIDDEN SYN_MANIFEST_SHA256 BUILDER_MEM'
grep -Fq 'memory: 1G' "$ROOT/test/stress/docker-compose-exp.yml"

run_valid WORKLOAD=absl BUILDER_MEM=2G
assert_output 'configure: HG_COORDINATOR=127.0.0.1:1 CC=hgcc CXX=hgcxx cmake -S /workspace/abseil-cpp -B /workspace/abseil-cpp-build -G Ninja -DCMAKE_BUILD_TYPE=Release -DCMAKE_CXX_STANDARD=17 -DABSL_BUILD_TESTING=OFF -DABSL_PROPAGATE_CXX_STD=ON'
assert_output 'build: cd /workspace/abseil-cpp-build && hgbuild ninja --no-fallback -j"$1"'
grep -Fq 'memory: 2G' "$ROOT/test/stress/docker-compose-exp.yml"

run_valid WORKLOAD=syn SYN_SEED=7 SYN_FILES=12 SYN_C_FRACTION=0.25 SYN_ALPHA=1.8 SYN_MEDIAN_UNITS=10 SYN_MAX_UNITS=50 SYN_HIDDEN=0.1
assert_output 'generate: python3 scripts/gen_synthetic_cpp.py --out DIR --seed 7 --files 12 --c-fraction 0.25 --pareto-alpha 1.8 --median-units 10 --max-units 50 --hidden-cost-fraction 0.1 (host)'
assert_output 'configure: HG_COORDINATOR=127.0.0.1:1 CC=hgcc CXX=hgcxx cmake -S /workspace/syn -B /workspace/syn-build -G Ninja -DCMAKE_BUILD_TYPE=Release -DCMAKE_CXX_STANDARD=17'
assert_output 'build: cd /workspace/syn-build && hgbuild ninja --no-fallback -j"$1"'

run_invalid WORKLOAD unknown 'WORKLOAD must be cpython, absl, or syn'
run_invalid SYN_SEED -1 'SYN_SEED must be a nonnegative integer'
run_invalid SYN_FILES 0 'SYN_FILES must be a positive integer'
run_invalid SYN_C_FRACTION 1.1 'SYN_C_FRACTION and SYN_HIDDEN must be in [0, 1]'
run_invalid SYN_ALPHA 0 'SYN_C_FRACTION and SYN_HIDDEN must be in [0, 1], SYN_ALPHA must be > 0'
run_invalid SYN_MEDIAN_UNITS 0 'SYN_MEDIAN_UNITS must be a positive integer'
run_invalid SYN_MAX_UNITS 0 'SYN_MAX_UNITS must be a positive integer'
run_invalid SYN_MAX_UNITS 20 'SYN_MAX_UNITS must be >= SYN_MEDIAN_UNITS'
run_invalid SYN_HIDDEN 1.1 'SYN_C_FRACTION and SYN_HIDDEN must be in [0, 1]'
run_invalid BUILDER_MEM invalid 'invalid BUILDER_MEM'
printf 'bench_exp workload tests passed\n'
