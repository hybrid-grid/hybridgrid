#!/bin/bash
# Warm-bandit benchmark — does a persistent bandit learn across builds?
#
# The rigorous benchmark restarts the coordinator every build, so the bandit
# cold-starts each time and never accumulates learning across the ~300 tasks
# of a single build. That handicap is the leading hypothesis for why
# hybrid-linucb only ties LeastLoaded. This harness removes it.
#
# Design: for each scheduler, bring up ONE coordinator + worker set, then run
# K sequential builds against that SAME long-lived coordinator. The bandit
# state (in-memory, never reset — confirmed by code trace) accumulates across
# builds 1..K. We record makespan per build index, so a learning scheduler
# shows a downward curve while a stateless heuristic stays flat.
#
# Between builds we `make clean`, clear the compile cache, truncate the task
# log, and cool down — but we do NOT restart the coordinator (that would reset
# the bandit). This mirrors the realistic production scenario: a long-lived
# coordinator serving many sequential builds.
#
# R independent sessions per scheduler give R learning curves for averaging
# and paired build-1-vs-build-K tests.
#
# Output: results.csv (scheduler,session,build_idx,workers,elapsed_s) plus
# tasks_<sched>_s<S>_b<K>.jsonl. Analyze with scripts/analyze_warm.py.
#
# Usage:
#   bash scripts/benchmark_warm.sh
#   BUILDS=10 SESSIONS=5 SCHEDULERS="hybrid-linucb leastloaded" bash scripts/benchmark_warm.sh
#
# Env: BUILDS, SESSIONS, SCHEDULERS, OUT_DIR, ALPHA, COOLDOWN, JOBS, CONFIGURE_FLAGS.

set -e

BUILDS="${BUILDS:-8}"
SESSIONS="${SESSIONS:-5}"
# hybrid-linucb + linucb are the learners; leastloaded is the flat control that
# proves any downward trend is learning, not systematic host warm-up.
SCHEDULERS="${SCHEDULERS:-hybrid-linucb linucb leastloaded}"
OUT_DIR="${OUT_DIR:-/tmp/bench_warm_$(date +%Y%m%d_%H%M%S)}"
ALPHA="${ALPHA:-0.5}"
COOLDOWN="${COOLDOWN:-8}"
JOBS="${JOBS:-5}"
CONFIGURE_FLAGS="${CONFIGURE_FLAGS:---disable-test-modules --disable-perf-trampoline}"

WRAPPER_DIR="$(cd "$(dirname "$0")" && pwd)"
STRESS_DIR="$WRAPPER_DIR/../test/stress"
COMPOSE="docker compose -f docker-compose-hetero.yml"

mkdir -p "$OUT_DIR"
# Header written once. Re-invoking with the same OUT_DIR appends, so a run can
# be split per-scheduler across background tasks and resumed after an interrupt.
[ -f "$OUT_DIR/results.csv" ] || \
    echo "scheduler,session,build_idx,workers,elapsed_s" > "$OUT_DIR/results.csv"

# A (scheduler, session) is DONE if results.csv already holds all BUILDS rows
# for it — lets a re-invocation skip finished sessions and resume mid-run.
session_done() {
    local sched=$1 session=$2 have
    have=$(awk -F, -v s="$sched" -v n="$session" \
        '$1==s && $2==n {c++} END{print c+0}' "$OUT_DIR/results.csv")
    [ "$have" -ge "$BUILDS" ]
}

cd "$STRESS_DIR"
HG_BENCH_LIB=1 source ./benchmark-heterogeneous.sh

now() { python3 -c 'import time; print(repr(time.time()))'; }

wait_registration() {
    local need=$1 timeout=$2 waited=0 n=0
    while true; do
        n=$($COMPOSE logs coordinator 2>&1 | grep -c "Worker registered" || true)
        [ "$n" -ge "$need" ] && break
        if [ "$waited" -ge "$timeout" ]; then
            warn "only $n/$need workers registered after ${timeout}s — proceeding"; break
        fi
        sleep 2; waited=$((waited + 2))
    done
    echo "$n"
}

# Configure the persistent CPython source (reconfigure only when flags change).
ensure_configured() {
    $COMPOSE exec -T builder bash -c "
        cd /workspace/cpython
        want='${CONFIGURE_FLAGS}'
        have=\$(cat .hg-configure-flags 2>/dev/null || echo '')
        if [ ! -f Makefile ] || [ \"\$want\" != \"\$have\" ]; then
            make clean 2>/dev/null || true
            ./configure \$want 2>&1 | tail -3
            echo \"\$want\" > .hg-configure-flags
        fi
    "
}

# One timed build against the ALREADY-RUNNING cluster. Does NOT touch the
# coordinator, so the bandit keeps its accumulated state. Prints elapsed s.
timed_build() {
    local sched=$1 session=$2 bidx=$3
    sleep "$COOLDOWN"
    $COMPOSE exec -T builder bash -c 'cd /workspace/cpython && make clean 2>/dev/null || true'
    $COMPOSE exec -T builder bash -c 'rm -rf /root/.hybridgrid/cache/*'
    # Truncate the task log so this build's JSONL holds only its own rows. The
    # log is pure observability — truncating it does NOT reset the in-memory
    # bandit, so learning carries across builds.
    $COMPOSE exec -T coordinator sh -c 'truncate -s 0 /tmp/tasks.jsonl' || true
    sleep 2

    local tag="${sched}-s${session}-b${bidx}" t0 t1 elapsed
    t0=$(now)
    $COMPOSE exec -T builder bash -c "cd /workspace/cpython && hgbuild make -j${JOBS} 2>&1" \
        2>&1 | tee "/tmp/build_warm_${tag}.log" >/dev/null
    t1=$(now)
    elapsed=$(python3 -c "print(f'{$t1 - $t0:.2f}')")

    if grep -qE '^make(\[[0-9]+\])?: \*\*\*' "/tmp/build_warm_${tag}.log"; then
        echo "ERROR: build failed for $tag (see /tmp/build_warm_${tag}.log)" >&2
        exit 1
    fi
    $COMPOSE cp coordinator:/tmp/tasks.jsonl "$OUT_DIR/tasks_${sched}_s${session}_b${bidx}.jsonl"
    echo "${sched},${session},${bidx},5,${elapsed}" >> "$OUT_DIR/results.csv"
    echo "$elapsed"
}

log "=== WARM-BANDIT BENCHMARK ==="
log "schedulers: $SCHEDULERS  builds/session: $BUILDS  sessions: $SESSIONS  alpha: $ALPHA"
log "workload flags: $CONFIGURE_FLAGS  jobs: -j${JOBS}  output: $OUT_DIR"

SCHEDULER=leastloaded compute_sched_args
generate_5_workers
log "Building Docker images..."
$COMPOSE build

for sched in $SCHEDULERS; do
    SCHEDULER="$sched"
    compute_sched_args
    generate_5_workers
    log "############ scheduler: $sched ($SCHED_ARGS) ############"

    for session in $(seq 1 "$SESSIONS"); do
        if session_done "$sched" "$session"; then
            log "=== $sched session $session/$SESSIONS already complete — skip (resume) ==="
            continue
        fi
        # A partially-recorded session (interrupted mid-way) must redo the whole
        # curve — the bandit can't resume from stale in-memory state. Purge its
        # partial rows so the redo doesn't duplicate build indices.
        if awk -F, -v s="$sched" -v n="$session" '$1==s && $2==n{f=1} END{exit !f}' \
                "$OUT_DIR/results.csv"; then
            grep -v -E "^${sched},${session}," "$OUT_DIR/results.csv" > "$OUT_DIR/results.tmp" \
                && mv "$OUT_DIR/results.tmp" "$OUT_DIR/results.csv"
            log "purged partial rows for $sched session $session before redo"
        fi
        # Fresh cluster per session => bandit starts cold, then warms across
        # the K builds below. Sessions are independent learning curves.
        log "=== $sched session $session/$SESSIONS (fresh coordinator, cold bandit) ==="
        $COMPOSE down 2>/dev/null || true
        $COMPOSE up -d coordinator builder
        sleep 5
        clone_cpython
        ensure_configured
        start_workers 5
        registered=$(wait_registration 5 90)
        log "workers registered: $registered"

        for bidx in $(seq 1 "$BUILDS"); do
            elapsed=$(timed_build "$sched" "$session" "$bidx")
            log "  [$sched] session $session build $bidx/$BUILDS: ${elapsed}s"
        done
    done
done

success "All sessions complete. Results: $OUT_DIR/results.csv"
cat "$OUT_DIR/results.csv"

PYTHON_BIN=""
if [ -x "$WRAPPER_DIR/.venv-bench/bin/python" ]; then
    PYTHON_BIN="$WRAPPER_DIR/.venv-bench/bin/python"
elif command -v python3 >/dev/null 2>&1; then
    PYTHON_BIN=python3
fi
if [ -n "$PYTHON_BIN" ]; then
    "$PYTHON_BIN" "$WRAPPER_DIR/analyze_warm.py" "$OUT_DIR" || \
        warn "analysis failed; rerun: $PYTHON_BIN scripts/analyze_warm.py $OUT_DIR"
else
    warn "python not found; run later: python scripts/analyze_warm.py $OUT_DIR"
fi
