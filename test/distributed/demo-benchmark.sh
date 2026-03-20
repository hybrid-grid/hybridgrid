#!/bin/bash
# Multi-machine distributed benchmark demo
# Run on Mac (coordinator machine) after all workers are connected
#
# 3 phases:
#   1. Local build (baseline, no hybridgrid)
#   2. Distributed build (cold cache)
#   3. Distributed build (warm cache)

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BINARIES_DIR="$SCRIPT_DIR/binaries/mac"
HGBUILD="${BINARIES_DIR}/hgbuild"
RESULTS_DIR="/tmp/hg-demo-results"

GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
NC='\033[0m'

log() { echo -e "${BLUE}[demo]${NC} $1"; }
success() { echo -e "${GREEN}[demo]${NC} $1"; }
error() { echo -e "${RED}[demo]${NC} $1"; }
warn() { echo -e "${YELLOW}[demo]${NC} $1"; }

mkdir -p "$RESULTS_DIR"

# ============================================
# Check prerequisites
# ============================================
check_prereqs() {
    log "Checking prerequisites..."

    if ! curl -s http://localhost:8080/api/v1/workers > /dev/null 2>&1; then
        error "Coordinator not running at localhost:8080"
        exit 1
    fi
    success "Coordinator: OK"

    local api_response
    api_response=$(curl -s http://localhost:8080/api/v1/workers 2>/dev/null)
    local worker_count
    worker_count=$(echo "$api_response" | python3 -c "
import sys, json
data = json.load(sys.stdin)
workers = data.get('workers', data) if isinstance(data, dict) else data
print(len(workers) if isinstance(workers, list) else 0)
" 2>/dev/null || echo "0")

    if [ "$worker_count" -lt 1 ]; then
        error "No workers connected!"
        exit 1
    fi
    success "Workers connected: $worker_count"

    echo "$api_response" | python3 -c "
import sys, json
data = json.load(sys.stdin)
workers = data.get('workers', data) if isinstance(data, dict) else data
if not isinstance(workers, list): workers = []
for w in workers:
    wid = w.get('id', 'unknown')
    arch = w.get('architecture', w.get('arch', 'unknown'))
    cores = w.get('cpu_cores', '?')
    mem = w.get('memory_gb', 0)
    addr = w.get('address', '?')
    healthy = 'OK' if w.get('healthy') else 'DOWN'
    print(f'  [{healthy}] {wid}  addr={addr}  arch={arch}  cores={cores}  ram={mem:.1f}GB')
" 2>/dev/null || true

    if [ ! -x "$HGBUILD" ]; then
        error "hgbuild not found at $HGBUILD. Run ./demo-setup.sh first"
        exit 1
    fi
    success "hgbuild: OK"
    echo ""
}

# ============================================
# Clone test project
# ============================================
setup_project() {
    local name=$1
    local repo=$2
    local dir="/tmp/$name"

    if [ ! -d "$dir" ]; then
        log "Cloning $name..."
        git clone --depth=1 "$repo" "$dir"
    else
        log "$name already cloned at $dir"
    fi

    if [ "$name" = "cpython" ] && [ ! -f "$dir/Makefile" ]; then
        log "Configuring CPython..."
        cd "$dir" && ./configure --disable-test-modules --disable-perf-trampoline 2>&1 | tail -3
    fi
}

# ============================================
# Run a timed build (verbose)
# ============================================
run_build() {
    local dir=$1
    local cmd=$2
    local logfile=$3

    cd "$dir"
    log "Cleaning previous build..."
    make clean > /dev/null 2>&1 || true

    log "Running: $cmd"
    local start=$(python3 -c "import time; print(time.time())")

    eval "$cmd" 2>&1 | tee "$logfile"

    local end=$(python3 -c "import time; print(time.time())")
    local elapsed=$(python3 -c "print(f'{float($end) - float($start):.1f}')")

    local total_lines=$(wc -l < "$logfile" 2>/dev/null | tr -d ' ')
    success "Done: ${elapsed}s ($total_lines output lines)"
    echo ""

    # Write elapsed to a known file for the caller
    echo "$elapsed" > "$RESULTS_DIR/.last_elapsed"
}

# ============================================
# Benchmark
# ============================================
benchmark_project() {
    local name=$1
    local dir="/tmp/$name"
    local jobs_local=8
    local jobs_dist=12

    echo ""
    log "${BOLD}========================================${NC}"
    log "${BOLD}  Benchmarking: $name${NC}"
    log "${BOLD}========================================${NC}"
    echo ""

    # ── Phase 1: Local baseline ──
    log "${BOLD}Phase 1: Local build (make -j$jobs_local)${NC}"
    log "This uses ONLY this Mac's CPU, no distribution."
    rm -rf ~/.hybridgrid/cache/* 2>/dev/null || true
    run_build "$dir" "make -j$jobs_local" "$RESULTS_DIR/local.log"
    local time_local=$(cat "$RESULTS_DIR/.last_elapsed")

    # ── Phase 2: Distributed (cold cache) ──
    log "${BOLD}Phase 2: Distributed build - COLD cache (hgbuild make -j$jobs_dist)${NC}"
    log "Cache cleared. Tasks distributed to all workers over network."
    rm -rf ~/.hybridgrid/cache/* 2>/dev/null || true
    run_build "$dir" "$HGBUILD make -j$jobs_dist" "$RESULTS_DIR/distributed.log"
    local time_dist=$(cat "$RESULTS_DIR/.last_elapsed")

    # ── Phase 3: Distributed (warm cache) ──
    log "${BOLD}Phase 3: Distributed build - WARM cache (hgbuild make -j$jobs_dist)${NC}"
    log "Cache is warm from Phase 2. Should hit local cache for most files."
    # Do NOT clear cache
    cd "$dir" && make clean > /dev/null 2>&1 || true
    log "Running: $HGBUILD make -j$jobs_dist"
    local cache_start=$(python3 -c "import time; print(time.time())")
    eval "$HGBUILD make -j$jobs_dist" 2>&1 | tee "$RESULTS_DIR/cached.log"
    local cache_end=$(python3 -c "import time; print(time.time())")
    local time_cached=$(python3 -c "print(f'{float($cache_end) - float($cache_start):.1f}')")
    local cache_lines=$(wc -l < "$RESULTS_DIR/cached.log" 2>/dev/null | tr -d ' ')
    success "Done: ${time_cached}s ($cache_lines output lines)"

    # ── Results ──
    local speedup_dist=$(python3 -c "print(f'{float($time_local) / float($time_dist):.2f}')")
    local speedup_cache=$(python3 -c "print(f'{float($time_local) / float($time_cached):.2f}')")

    echo ""
    echo -e "${BOLD}╔══════════════════════════════════════════════╗${NC}"
    echo -e "${BOLD}║              RESULTS: $name${NC}"
    echo -e "${BOLD}╠══════════════════════════════════════════════╣${NC}"
    echo -e "║  Local (make -j$jobs_local):          ${BOLD}${time_local}s${NC}"
    echo -e "║  Distributed (cold cache):  ${BOLD}${time_dist}s${NC}  → ${GREEN}${speedup_dist}x speedup${NC}"
    echo -e "║  Distributed (warm cache):  ${BOLD}${time_cached}s${NC}  → ${GREEN}${speedup_cache}x speedup${NC}"
    echo -e "${BOLD}╚══════════════════════════════════════════════╝${NC}"
    echo ""

    echo "$name,$time_local,$time_dist,$time_cached,$speedup_dist,$speedup_cache" >> "$RESULTS_DIR/results.csv"
}

# ============================================
# Main
# ============================================
main() {
    echo ""
    echo -e "${BOLD}╔══════════════════════════════════════════╗${NC}"
    echo -e "${BOLD}║   Hybrid-Grid Distributed Build Demo     ║${NC}"
    echo -e "${BOLD}║   Mac + Raspi5 + Windows PC              ║${NC}"
    echo -e "${BOLD}╚══════════════════════════════════════════╝${NC}"
    echo ""

    check_prereqs

    echo "project,local,dist_cold,dist_cached,speedup_dist,speedup_cache" > "$RESULTS_DIR/results.csv"

    case "${1:-redis}" in
        redis)
            setup_project "redis" "https://github.com/redis/redis"
            benchmark_project "redis"
            ;;
        cpython)
            setup_project "cpython" "https://github.com/python/cpython"
            benchmark_project "cpython"
            ;;
        both)
            setup_project "redis" "https://github.com/redis/redis"
            setup_project "cpython" "https://github.com/python/cpython"
            benchmark_project "redis"
            benchmark_project "cpython"
            ;;
        *)
            error "Usage: $0 [redis|cpython|both]"
            exit 1
            ;;
    esac

    success "Results saved to: $RESULTS_DIR/results.csv"
    success "Dashboard: http://localhost:8080"
}

main "$@"
