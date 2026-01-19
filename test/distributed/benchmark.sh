#!/bin/bash
# Distributed benchmark for hybridgrid
# Run this on the Mac (coordinator + builder)

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
RESULTS_FILE="/tmp/distributed_benchmark_results.txt"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log() { echo -e "${BLUE}[benchmark]${NC} $1"; }
success() { echo -e "${GREEN}[benchmark]${NC} $1"; }

echo "# Hybridgrid Distributed Benchmark" > "$RESULTS_FILE"
echo "# Date: $(date)" >> "$RESULTS_FILE"
echo "# Machines: Mac (coordinator) + Windows (worker) + Raspi (worker)" >> "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"

# Check workers are connected
check_workers() {
    log "Checking connected workers..."
    curl -s http://localhost:8080/workers | head -20
    echo ""
}

# Benchmark with a project
run_benchmark() {
    local project_name=$1
    local project_dir=$2
    local runs=${3:-3}

    log "=== Benchmarking: $project_name ==="

    cd "$project_dir"

    # Baseline (local make)
    log "Running baseline (local make)..."
    local baseline_times=()
    for i in $(seq 1 $runs); do
        make clean >/dev/null 2>&1 || true
        local start=$(date +%s)
        make -j4 >/dev/null 2>&1
        local end=$(date +%s)
        local elapsed=$((end - start))
        baseline_times+=($elapsed)
        log "  Run $i: ${elapsed}s"
    done

    # Calculate average baseline
    local baseline_sum=0
    for t in "${baseline_times[@]}"; do
        baseline_sum=$((baseline_sum + t))
    done
    local baseline_avg=$((baseline_sum / runs))

    # Hybridgrid (distributed make)
    log "Running distributed (hgbuild make)..."
    local distributed_times=()
    for i in $(seq 1 $runs); do
        make clean >/dev/null 2>&1 || true
        local start=$(date +%s)
        hgbuild make -j6 >/dev/null 2>&1
        local end=$(date +%s)
        local elapsed=$((end - start))
        distributed_times+=($elapsed)
        log "  Run $i: ${elapsed}s"
    done

    # Calculate average distributed
    local distributed_sum=0
    for t in "${distributed_times[@]}"; do
        distributed_sum=$((distributed_sum + t))
    done
    local distributed_avg=$((distributed_sum / runs))

    # Calculate speedup
    local speedup=$(echo "scale=2; $baseline_avg / $distributed_avg" | bc)

    success "$project_name: baseline=${baseline_avg}s, distributed=${distributed_avg}s, speedup=${speedup}x"
    echo "$project_name,$baseline_avg,$distributed_avg,$speedup" >> "$RESULTS_FILE"
}

# Main
main() {
    log "Starting distributed benchmark..."
    echo ""

    check_workers

    # Clone test projects if not exist
    if [ ! -d /tmp/redis ]; then
        log "Cloning redis..."
        git clone --depth=1 https://github.com/redis/redis /tmp/redis
    fi

    # Run benchmarks
    run_benchmark "redis" "/tmp/redis" 3

    # Print results
    echo ""
    echo "=========================================="
    echo "         BENCHMARK RESULTS"
    echo "=========================================="
    cat "$RESULTS_FILE"
    echo "=========================================="
}

main "$@"
