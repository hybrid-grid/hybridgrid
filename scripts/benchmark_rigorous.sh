#!/bin/bash
# Rigorous scheduler benchmark — randomized complete block design.
#
# Motivation: benchmark_statistical.sh runs all REPS of one scheduler
# back-to-back, so any slow drift in the host (thermal throttling,
# background load) is confounded with scheduler identity — the scheduler
# that happens to run last is measured on a hotter machine. This wrapper
# removes that confound and every other one we can control:
#
#   1. Randomized block design: each round runs ALL schedulers once, in a
#      per-round shuffled order (seeded, logged). Every scheduler sees the
#      full range of host conditions. Rounds become statistical blocks, so
#      the analyzer can use PAIRED tests (Wilcoxon signed-rank / Friedman)
#      which control for round-level nuisance and are more powerful.
#   2. Warm-up build discarded before recording — the first build pays cold
#      filesystem / Docker-layer cache costs that later builds don't.
#   3. Registration barrier: wait until all N workers have registered
#      instead of a fixed sleep, so no build starts on a partial cluster.
#   4. Cooldown before each timed build — lets the host settle to a
#      comparable thermal state.
#   5. Sub-second wall-clock timing (python time.time) around ONLY the
#      `hgbuild make` call, excluding make-clean / cache-clear / cooldown.
#   6. Cache cleared every run (inherited) — no compile-cache leakage.
#   7. Bandit coordinators are restarted every build, so learners cold-start
#      each time. This HANDICAPS the bandit (no cross-build learning); a win
#      under this constraint is conservative, i.e. strictly fair.
#
# Output: results.csv with a `round` column (scheduler,round,order_pos,
# workers,elapsed_s) plus tasks_<scheduler>_round<N>.jsonl per cell.
# Analyze with scripts/analyze_rigorous.py.
#
# Usage:
#   bash scripts/benchmark_rigorous.sh
#   REPS=12 COOLDOWN=10 bash scripts/benchmark_rigorous.sh
#
# Env: REPS, SCHEDULERS, OUT_DIR, ALPHA, COOLDOWN, SEED_BASE, JOBS.

set -e

REPS="${REPS:-10}"
# ROUND_START lets a killed run resume: ROUND_START=6 appends rounds 6..REPS
# to an existing OUT_DIR without a header rewrite or a second warm-up. Round
# seeds are round-indexed, so resumed rounds match a from-scratch run.
ROUND_START="${ROUND_START:-1}"
SCHEDULERS="${SCHEDULERS:-leastloaded p2c linucb hybrid-linucb}"
OUT_DIR="${OUT_DIR:-/tmp/bench_rig_$(date +%Y%m%d_%H%M%S)}"
ALPHA="${ALPHA:-0.5}"
COOLDOWN="${COOLDOWN:-8}"
SEED_BASE="${SEED_BASE:-1000}"
JOBS="${JOBS:-5}"
# CPython ./configure flags. Default = light workload (~293 TU/build, test
# modules disabled). Heavy regime: drop --disable-test-modules so all C
# extension modules compile, producing more translation units and genuine
# queue contention. Changing this reconfigures the persistent source volume.
CONFIGURE_FLAGS="${CONFIGURE_FLAGS:---disable-test-modules --disable-perf-trampoline}"

WRAPPER_DIR="$(cd "$(dirname "$0")" && pwd)"
STRESS_DIR="$WRAPPER_DIR/../test/stress"
COMPOSE="docker compose -f docker-compose-hetero.yml"

mkdir -p "$OUT_DIR"
# Fresh run writes headers; a resume (ROUND_START>1) appends to existing files.
if [ "$ROUND_START" -le 1 ] || [ ! -f "$OUT_DIR/results.csv" ]; then
    echo "scheduler,round,order_pos,workers,elapsed_s" > "$OUT_DIR/results.csv"
    echo "round,order" > "$OUT_DIR/order_log.csv"
fi

cd "$STRESS_DIR"
# Library mode: reuse compose generators + build helpers. The sourced
# file installs a `trap cleanup EXIT` that tears the cluster down when we
# exit — do not add another EXIT trap here.
HG_BENCH_LIB=1 source ./benchmark-heterogeneous.sh

now() { python3 -c 'import time; print(repr(time.time()))'; }

# Block until `need` workers have registered with the fresh coordinator,
# or until `timeout` seconds elapse. Logs reset on every `compose down`,
# so the grep count reflects only the current run.
wait_registration() {
    local need=$1 timeout=$2 waited=0 n=0
    while true; do
        n=$($COMPOSE logs coordinator 2>&1 | grep -c "Worker registered" || true)
        [ "$n" -ge "$need" ] && break
        if [ "$waited" -ge "$timeout" ]; then
            warn "only $n/$need workers registered after ${timeout}s — proceeding"
            break
        fi
        sleep 2; waited=$((waited + 2))
    done
    echo "$n"
}

# Bring up one scheduler on a fresh cluster and run one timed build.
# Args: scheduler round order_pos record(0|1). Prints elapsed seconds.
run_cell() {
    local sched=$1 round=$2 pos=$3 record=$4
    SCHEDULER="$sched"
    compute_sched_args
    generate_5_workers

    $COMPOSE down 2>/dev/null || true
    $COMPOSE up -d coordinator builder
    sleep 5
    # task-log volume outlives `down`; truncate so this cell's JSONL holds
    # only its own rows (logger opens O_APPEND, external truncate is safe).
    $COMPOSE exec -T coordinator sh -c 'truncate -s 0 /tmp/tasks.jsonl' || true

    clone_cpython
    # Reconfigure when CONFIGURE_FLAGS change (a stale Makefile from a prior
    # flag set would silently build the wrong workload). The marker records
    # the flags the current Makefile was generated with.
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

    start_workers 5
    local registered
    registered=$(wait_registration 5 90)
    log "cell [$sched r$round p$pos] workers registered: $registered"

    # Cooldown so every build starts from a comparable thermal state.
    sleep "$COOLDOWN"

    # make clean + cache clear are OUTSIDE the timed window.
    $COMPOSE exec -T builder bash -c 'cd /workspace/cpython && make clean 2>/dev/null || true'
    $COMPOSE exec -T builder bash -c 'rm -rf /root/.hybridgrid/cache/*'
    sleep 2

    local tag="${sched}-r${round}"
    local t0 t1 elapsed
    t0=$(now)
    $COMPOSE exec -T builder bash -c "cd /workspace/cpython && hgbuild make -j${JOBS} 2>&1" \
        2>&1 | tee "/tmp/build_rig_${tag}.log" >/dev/null
    t1=$(now)
    elapsed=$(python3 -c "print(f'{$t1 - $t0:.2f}')")

    # A failed build must never contribute a sample.
    if grep -qE '^make(\[[0-9]+\])?: \*\*\*' "/tmp/build_rig_${tag}.log"; then
        echo "ERROR: build failed for $sched round $round (see /tmp/build_rig_${tag}.log)" >&2
        exit 1
    fi

    if [ "$record" = "1" ]; then
        $COMPOSE cp coordinator:/tmp/tasks.jsonl "$OUT_DIR/tasks_${sched}_round${round}.jsonl"
        echo "${sched},${round},${pos},5,${elapsed}" >> "$OUT_DIR/results.csv"
    fi
    echo "$elapsed"
}

log "=== RIGOROUS BENCHMARK ==="
log "schedulers: $SCHEDULERS  reps(rounds): $REPS  alpha: $ALPHA  cooldown: ${COOLDOWN}s  jobs: -j${JOBS}"
log "output: $OUT_DIR"

# Build images once against the current working tree.
SCHEDULER=leastloaded compute_sched_args
generate_5_workers
log "Building Docker images..."
$COMPOSE build

# --- Warm-up (discarded) -------------------------------------------------
# One full build to warm filesystem + Docker-layer caches and configure
# CPython, so no RECORDED build pays cold-start costs. Skipped on resume:
# the named cpython-src / build-cache volumes survive `compose down`.
if [ "$ROUND_START" -le 1 ]; then
    log "=== warm-up build (discarded) ==="
    run_cell "leastloaded" 0 0 0 >/dev/null
    success "warm-up complete"
else
    log "=== resume from round $ROUND_START (warm-up skipped) ==="
fi

# --- Randomized complete block design ------------------------------------
for round in $(seq "$ROUND_START" "$REPS"); do
    # Deterministic per-round shuffle (seeded, reproducible + logged).
    order=$(SEED_BASE="$SEED_BASE" python3 -c "
import os, random
random.seed(int(os.environ['SEED_BASE']) + $round)
a = '''$SCHEDULERS'''.split()
random.shuffle(a)
print(' '.join(a))
")
    log "=== round $round/$REPS  order: $order ==="
    echo "${round},${order// /|}" >> "$OUT_DIR/order_log.csv"

    pos=0
    for sched in $order; do
        pos=$((pos + 1))
        elapsed=$(run_cell "$sched" "$round" "$pos" 1)
        log "  [$sched] round $round: ${elapsed}s"
    done
done

success "All rounds complete. Results: $OUT_DIR/results.csv"
cat "$OUT_DIR/results.csv"

# Analysis (best-effort). Prefer the project venv (has scipy/pandas).
PYTHON_BIN=""
if [ -x "$WRAPPER_DIR/.venv-bench/bin/python" ]; then
    PYTHON_BIN="$WRAPPER_DIR/.venv-bench/bin/python"
elif command -v python3 >/dev/null 2>&1; then
    PYTHON_BIN=python3
fi
if [ -n "$PYTHON_BIN" ]; then
    "$PYTHON_BIN" "$WRAPPER_DIR/analyze_rigorous.py" "$OUT_DIR" || \
        warn "analysis failed; rerun: $PYTHON_BIN scripts/analyze_rigorous.py $OUT_DIR"
else
    warn "python not found; run later: python scripts/analyze_rigorous.py $OUT_DIR"
fi
