#!/bin/bash
# Statistical scheduler benchmark (docs/thesis/hybrid_linucb_proposal.md §4
# step 4): repeat the 5w-hetero CPython build REPS times per scheduler and
# aggregate makespans + per-task JSONL logs for significance testing.
#
# The 5w-hetero cluster is the only config where scheduler differences
# exceed the ±25s single-run noise floor (LeastLoaded 152s vs P2C 94s),
# so all repetitions run there. Each rep restarts the coordinator, which
# resets bandit arm state — samples are independent.
#
# Budget: REPS×|SCHEDULERS| CPython builds (873 tasks each, ~3-5 min/build
# plus one-time clone+configure) ≈ 2-4 hours for the default 10×4.
#
# Usage:
#   bash scripts/benchmark_statistical.sh                 # full run
#   REPS=1 SCHEDULERS="hybrid-linucb" bash scripts/benchmark_statistical.sh   # smoke
#
# Env overrides: REPS, SCHEDULERS (space-separated), OUT_DIR, ALPHA,
# WARM_START, LOAD_PENALTY.

set -e

REPS="${REPS:-10}"
SCHEDULERS="${SCHEDULERS:-leastloaded p2c linucb hybrid-linucb}"
OUT_DIR="${OUT_DIR:-/tmp/bench_stat_$(date +%Y%m%d_%H%M%S)}"
# α=0.5 for both linucb and hybrid-linucb (the α-sweep optimum); other
# schedulers ignore the flag, so one value keeps the comparison fair.
ALPHA="${ALPHA:-0.5}"

WRAPPER_DIR="$(cd "$(dirname "$0")" && pwd)"
STRESS_DIR="$WRAPPER_DIR/../test/stress"
COMPOSE="docker compose -f docker-compose-hetero.yml"

mkdir -p "$OUT_DIR"
echo "scheduler,run,workers,elapsed_s" > "$OUT_DIR/results.csv"
echo "[stat] output directory: $OUT_DIR"
echo "[stat] schedulers: $SCHEDULERS  reps: $REPS  alpha: $ALPHA"

# Source the harness in library mode: reuses compute_sched_args, the
# compose generators, clone_cpython, run_build and start_workers. Note
# the harness's top level runs at source time — it cds into test/stress
# and installs a `trap cleanup EXIT` that tears the cluster down when
# this wrapper exits. Do not install another EXIT trap here.
cd "$STRESS_DIR"
HG_BENCH_LIB=1 source ./benchmark-heterogeneous.sh

# Build images once, against the current working tree (the binaries
# under test live inside the image).
SCHEDULER=leastloaded compute_sched_args
generate_5_workers
log "Building Docker images..."
$COMPOSE build

for sched in $SCHEDULERS; do
    SCHEDULER="$sched"
    case "$SCHEDULER" in
        leastloaded|simple|p2c|epsilon-greedy|linucb|hybrid-linucb|heft) ;;
        *) echo "ERROR: invalid scheduler '$SCHEDULER'" >&2; exit 1 ;;
    esac
    # SCHED_ARGS is baked into the compose file, so recompute and
    # regenerate for every scheduler — never reuse a stale file.
    compute_sched_args
    generate_5_workers
    log "=== scheduler: $SCHEDULER ($SCHED_ARGS) ==="

    for run in $(seq 1 "$REPS"); do
        log "--- $SCHEDULER run $run/$REPS ---"
        $COMPOSE down 2>/dev/null || true
        $COMPOSE up -d coordinator builder
        sleep 5

        # The task-log volume outlives `compose down`; without this
        # truncate each run's JSONL would contain every prior run's
        # rows (the logger opens O_APPEND, so external truncate is safe).
        # sh -c keeps the container path out of the argument list, where
        # Git Bash on Windows would rewrite /tmp/... to a host path.
        $COMPOSE exec -T coordinator sh -c 'truncate -s 0 /tmp/tasks.jsonl' || true

        clone_cpython
        # configure_cpython reruns ./configure (~1.5 min) every call;
        # the Makefile survives `make clean` in the persistent volume,
        # so only the first run pays for it.
        $COMPOSE exec -T builder bash -c '
            cd /workspace/cpython
            if [ ! -f Makefile ]; then
                ./configure --disable-test-modules --disable-perf-trampoline 2>&1 | tail -5
            else
                echo "CPython already configured"
            fi
        '

        start_workers 5
        registered=$($COMPOSE logs coordinator 2>&1 | grep -c "Worker registered" || echo "0")
        log "workers registered: $registered"

        elapsed=$(run_build "5w-${SCHEDULER}-r${run}" 5 | tail -1)

        # A failed build must never contribute a makespan sample — a
        # partial build finishes fast and would silently skew the stats.
        buildlog="/tmp/build_hetero_5w-${SCHEDULER}-r${run}.log"
        if grep -qE '^make(\[[0-9]+\])?: \*\*\*' "$buildlog"; then
            echo "ERROR: build failed for $SCHEDULER run $run (see $buildlog)" >&2
            exit 1
        fi

        $COMPOSE cp coordinator:/tmp/tasks.jsonl "$OUT_DIR/tasks_${SCHEDULER}_run${run}.jsonl"
        lines=$(wc -l < "$OUT_DIR/tasks_${SCHEDULER}_run${run}.jsonl" | tr -d ' ')
        log "run $run: ${elapsed}s, ${lines} task records"
        echo "${SCHEDULER},${run},5,${elapsed}" >> "$OUT_DIR/results.csv"
    done
done

success "All runs complete. Results: $OUT_DIR/results.csv"
cat "$OUT_DIR/results.csv"

# Analysis is best-effort: the raw CSV/JSONL above are the durable
# artifacts, so a missing Python environment must not fail the run.
PYTHON_BIN=""
if command -v python3 >/dev/null 2>&1 && python3 -c '' >/dev/null 2>&1; then
    PYTHON_BIN=python3
elif command -v python >/dev/null 2>&1 && python -c '' >/dev/null 2>&1; then
    PYTHON_BIN=python
elif command -v py >/dev/null 2>&1; then
    PYTHON_BIN="py -3"
fi
if [ -n "$PYTHON_BIN" ]; then
    # Unquoted on purpose: PYTHON_BIN may be "py -3" (two words).
    $PYTHON_BIN "$WRAPPER_DIR/analyze_benchmark.py" "$OUT_DIR" || \
        warn "analysis failed; rerun later with: python scripts/analyze_benchmark.py $OUT_DIR"
else
    warn "python not found; run later: python scripts/analyze_benchmark.py $OUT_DIR"
fi
