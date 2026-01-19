#!/bin/bash
# Fair benchmark: equal total resources across all tests
# 1 worker (1.5 CPU) vs 3 workers (0.5 each) vs 5 workers (0.3 each)

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

RESULTS_FILE="/tmp/benchmark_fair_results.txt"
echo "# Hybridgrid Fair Benchmark (Equal Total Resources)" > "$RESULTS_FILE"
echo "# Date: $(date)" >> "$RESULTS_FILE"
echo "# Total resources per test: 2.5 CPU, 2.5GB RAM" >> "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"

cleanup() {
    log "Cleaning up..."
    docker compose -f docker-compose-fair.yml down 2>/dev/null || true
}
trap cleanup EXIT

# Generate docker-compose for specific worker count and resources
generate_compose() {
    local workers=$1
    local cpu_per_worker=$2
    local mem_per_worker=$3
    local max_parallel=$4

    cat > docker-compose-fair.yml << EOF
services:
  coordinator:
    build:
      context: ../..
      dockerfile: test/stress/Dockerfile.base
    command: hg-coord serve --grpc-port=9000 --http-port=8080
    ports:
      - "9000:9000"
      - "8080:8080"
    networks:
      - hgnet
    deploy:
      resources:
        limits:
          cpus: '0.25'
          memory: 256M

EOF

    for i in $(seq 1 $workers); do
        cat >> docker-compose-fair.yml << EOF
  worker-$i:
    build:
      context: ../..
      dockerfile: test/stress/Dockerfile.base
    command: hg-worker serve --coordinator=coordinator:9000 --port=50051 --advertise-address=worker-$i:50051 --max-parallel=$max_parallel
    networks:
      - hgnet
    depends_on:
      - coordinator
    deploy:
      resources:
        limits:
          cpus: '$cpu_per_worker'
          memory: ${mem_per_worker}M

EOF
    done

    cat >> docker-compose-fair.yml << EOF
  builder:
    build:
      context: ../..
      dockerfile: test/stress/Dockerfile.base
    command: sleep infinity
    environment:
      - HG_COORDINATOR=coordinator:9000
    networks:
      - hgnet
    depends_on:
      - coordinator
    volumes:
      - build-cache:/root/.hybridgrid/cache
      - cpython-src:/workspace
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 1G
    stdin_open: true
    tty: true

networks:
  hgnet:
    driver: bridge

volumes:
  build-cache:
  cpython-src:
EOF
}

clone_cpython() {
    log "Cloning CPython in builder container..."
    docker compose -f docker-compose-fair.yml exec -T builder bash -c '
        if [ ! -d /workspace/cpython ]; then
            git clone --depth=1 https://github.com/python/cpython.git /workspace/cpython
        else
            echo "CPython already cloned"
        fi
    '
}

configure_cpython() {
    log "Configuring CPython..."
    docker compose -f docker-compose-fair.yml exec -T builder bash -c '
        cd /workspace/cpython
        ./configure --disable-test-modules --disable-perf-trampoline 2>&1 | tail -5
    '
}

run_build() {
    local workers=$1
    local jobs=$2

    log "Running distributed build with $workers worker(s), -j$jobs..."

    # Clean previous build
    docker compose -f docker-compose-fair.yml exec -T builder bash -c 'cd /workspace/cpython && make clean 2>/dev/null || true'

    # Clear cache for fair comparison
    docker compose -f docker-compose-fair.yml exec -T builder bash -c 'rm -rf /root/.hybridgrid/cache/*'
    sleep 2

    local start_time=$(date +%s)

    docker compose -f docker-compose-fair.yml exec -T builder bash -c "
        cd /workspace/cpython
        hgbuild make -j$jobs 2>&1
    " 2>&1 | tee /tmp/build_fair_${workers}w.log

    local end_time=$(date +%s)
    local elapsed=$((end_time - start_time))

    success "Build with $workers worker(s) completed in ${elapsed} seconds"
    echo "$elapsed"
}

run_test() {
    local workers=$1
    local cpu_per_worker=$2
    local mem_per_worker=$3
    local max_parallel=$4
    local jobs=$5

    log ""
    log "=========================================="
    log "=== TEST: $workers Worker(s) @ ${cpu_per_worker} CPU, ${mem_per_worker}MB each ==="
    log "=========================================="

    # Generate compose file
    generate_compose $workers $cpu_per_worker $mem_per_worker $max_parallel

    # Stop everything
    docker compose -f docker-compose-fair.yml down 2>/dev/null || true

    # Start services
    log "Starting coordinator and builder..."
    docker compose -f docker-compose-fair.yml up -d coordinator builder
    sleep 5

    # Clone/configure on first run
    clone_cpython
    configure_cpython

    # Start workers
    log "Starting $workers worker(s)..."
    for i in $(seq 1 $workers); do
        docker compose -f docker-compose-fair.yml up -d "worker-$i"
    done
    sleep 5

    # Verify workers
    local registered=$(docker compose -f docker-compose-fair.yml logs coordinator 2>&1 | grep "Worker registered" | wc -l || echo "0")
    log "Workers registered: $registered"

    # Run build
    local elapsed=$(run_build $workers $jobs | tail -1)
    echo "distributed,$workers,$cpu_per_worker,$mem_per_worker,$elapsed" >> "$RESULTS_FILE"

    echo "$elapsed"
}

main() {
    log "Starting Fair Benchmark (Equal Total Resources)"
    log "Total: 2.5 CPU, 2.5GB RAM per test"
    echo ""

    # Build base image first
    log "Building Docker images..."
    generate_compose 1 2.5 2560 6
    docker compose -f docker-compose-fair.yml build

    # Test 1: 1 worker with 2.5 CPU, 2560MB, max_parallel=6, -j6
    TIME_1W=$(run_test 1 2.5 2560 6 6)

    # Test 2: 3 workers with 0.83 CPU, 853MB each, max_parallel=2, -j6
    TIME_3W=$(run_test 3 0.83 853 2 6)

    # Test 3: 4 workers with 0.625 CPU, 640MB each, max_parallel=2, -j6
    TIME_4W=$(run_test 4 0.625 640 2 6)

    # Test 4: 5 workers with 0.5 CPU, 512MB each, max_parallel=2, -j6
    TIME_5W=$(run_test 5 0.5 512 2 6)

    # Print results
    echo ""
    echo "=========================================="
    echo "         FAIR BENCHMARK RESULTS"
    echo "=========================================="
    echo ""
    cat "$RESULTS_FILE"
    echo ""
    echo "=========================================="

    success "============================================"
    success "           BENCHMARK COMPLETE"
    success "============================================"
    echo ""
    echo "Results Summary (Equal Total Resources: 2.5 CPU, 2.5GB):"
    echo "  1 worker  (2.5 CPU):   ${TIME_1W} seconds"
    echo "  3 workers (0.83 CPU):  ${TIME_3W} seconds"
    echo "  4 workers (0.625 CPU): ${TIME_4W} seconds"
    echo "  5 workers (0.5 CPU):   ${TIME_5W} seconds"
    echo ""

    # Calculate speedup/slowdown
    if [ -n "$TIME_1W" ] && [ "$TIME_1W" -gt 0 ] 2>/dev/null; then
        RATIO_3=$(echo "scale=2; $TIME_1W / $TIME_3W" | bc 2>/dev/null || echo "N/A")
        RATIO_4=$(echo "scale=2; $TIME_1W / $TIME_4W" | bc 2>/dev/null || echo "N/A")
        RATIO_5=$(echo "scale=2; $TIME_1W / $TIME_5W" | bc 2>/dev/null || echo "N/A")
        echo "Relative Performance:"
        echo "  3 workers vs 1: ${RATIO_3}x"
        echo "  4 workers vs 1: ${RATIO_4}x"
        echo "  5 workers vs 1: ${RATIO_5}x"
        echo ""
        echo "Note: >1.0 means distributed is faster (parallelization wins)"
        echo "      <1.0 means single worker is faster (overhead loses)"
    fi
}

main "$@"
