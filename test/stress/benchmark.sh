#!/bin/bash
# CPython build benchmark for hybridgrid
# Tests with 1, 3, and 5 workers

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() { echo -e "${BLUE}[benchmark]${NC} $1"; }
success() { echo -e "${GREEN}[benchmark]${NC} $1"; }
warn() { echo -e "${YELLOW}[benchmark]${NC} $1"; }
error() { echo -e "${RED}[benchmark]${NC} $1"; }

RESULTS_FILE="/tmp/benchmark_results.txt"
echo "# Hybridgrid CPython Build Benchmark" > "$RESULTS_FILE"
echo "# Date: $(date)" >> "$RESULTS_FILE"
echo "# Resource limits: 0.5 CPU, 512MB RAM per worker" >> "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"

cleanup() {
    log "Cleaning up..."
    docker compose down 2>/dev/null || true
}
trap cleanup EXIT

build_images() {
    log "Building Docker images (this may take a few minutes)..."
    docker compose build
}

stop_workers() {
    log "Stopping all workers..."
    docker compose stop worker-1 worker-2 worker-3 worker-4 worker-5 2>/dev/null || true
    docker compose rm -f worker-1 worker-2 worker-3 worker-4 worker-5 2>/dev/null || true
}

start_workers() {
    local count=$1
    log "Starting $count worker(s)..."

    for i in $(seq 1 $count); do
        docker compose up -d "worker-$i"
    done
    sleep 5

    # Verify workers registered
    log "Verifying workers registered..."
    sleep 2
    local registered=$(docker compose logs coordinator 2>&1 | grep "Worker registered" | wc -l || echo "0")
    log "Workers registered: $registered"
}

clone_cpython() {
    log "Cloning CPython in builder container..."
    docker compose exec -T builder bash -c '
        if [ ! -d /workspace/cpython ]; then
            git clone --depth=1 https://github.com/python/cpython.git /workspace/cpython
        else
            echo "CPython already cloned"
        fi
    '
}

configure_cpython() {
    log "Configuring CPython..."
    docker compose exec -T builder bash -c '
        cd /workspace/cpython
        # Always reconfigure to ensure correct flags
        ./configure --disable-test-modules --disable-perf-trampoline 2>&1 | tail -5
    '
}

run_build() {
    local workers=$1
    # Scale parallelism conservatively (1.5 tasks per worker) to avoid race conditions
    # Each worker has max_parallel=2, but we give headroom for scheduling latency
    local jobs=$(( (workers * 3) / 2 ))
    # Minimum of 2 jobs
    if [ "$jobs" -lt 2 ]; then
        jobs=2
    fi

    log "Running distributed build with $workers worker(s), -j$jobs..."

    # Clean previous build
    docker compose exec -T builder bash -c 'cd /workspace/cpython && make clean 2>/dev/null || true'

    # Clear cache for fair comparison
    docker compose exec -T builder bash -c 'rm -rf /root/.hybridgrid/cache/*'
    sleep 2

    local start_time=$(date +%s)

    # Run distributed build with hgbuild
    # Note: Don't use -v flag with make wrapper (DisableFlagParsing passes it to make)
    docker compose exec -T builder bash -c "
        cd /workspace/cpython
        hgbuild make -j$jobs 2>&1
    " 2>&1 | tee /tmp/build_${workers}w.log

    local end_time=$(date +%s)
    local elapsed=$((end_time - start_time))

    success "Build with $workers worker(s) completed in ${elapsed} seconds"
    echo "distributed,$workers,$elapsed" >> "$RESULTS_FILE"

    # Return just the number
    echo "$elapsed"
}

print_results() {
    echo ""
    echo "=========================================="
    echo "         BENCHMARK RESULTS"
    echo "=========================================="
    echo ""
    cat "$RESULTS_FILE"
    echo ""
    echo "=========================================="
}

# Main benchmark flow
main() {
    log "Starting CPython build benchmark..."
    log "Each test: coordinator (0.25 CPU, 256MB) + N workers (0.5 CPU, 512MB each)"
    echo ""

    # Build images first
    build_images

    # Start base services (coordinator + builder)
    log "Starting coordinator and builder..."
    docker compose up -d coordinator builder
    sleep 5

    # Clone and configure CPython
    clone_cpython
    configure_cpython

    # Test 1: 1 worker
    log ""
    log "=========================================="
    log "=== TEST 1: 1 Worker ==="
    log "=========================================="
    stop_workers
    start_workers 1
    TIME_1W=$(run_build 1 | tail -1)

    # Test 2: 3 workers
    log ""
    log "=========================================="
    log "=== TEST 2: 3 Workers ==="
    log "=========================================="
    stop_workers
    start_workers 3
    TIME_3W=$(run_build 3 | tail -1)

    # Test 3: 5 workers
    log ""
    log "=========================================="
    log "=== TEST 3: 5 Workers ==="
    log "=========================================="
    stop_workers
    start_workers 5
    TIME_5W=$(run_build 5 | tail -1)

    # Print final results
    print_results

    echo ""
    success "============================================"
    success "           BENCHMARK COMPLETE"
    success "============================================"
    echo ""
    echo "Results Summary:"
    echo "  1 worker:   ${TIME_1W} seconds"
    echo "  3 workers:  ${TIME_3W} seconds"
    echo "  5 workers:  ${TIME_5W} seconds"
    echo ""

    # Calculate speedup
    if [ -n "$TIME_1W" ] && [ "$TIME_1W" -gt 0 ] 2>/dev/null; then
        SPEEDUP_3=$(echo "scale=2; $TIME_1W / $TIME_3W" | bc 2>/dev/null || echo "N/A")
        SPEEDUP_5=$(echo "scale=2; $TIME_1W / $TIME_5W" | bc 2>/dev/null || echo "N/A")
        echo "Speedup:"
        echo "  3 workers vs 1: ${SPEEDUP_3}x"
        echo "  5 workers vs 1: ${SPEEDUP_5}x"
    fi
}

main "$@"
