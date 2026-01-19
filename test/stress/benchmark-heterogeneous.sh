#!/bin/bash
# Heterogeneous worker benchmark
# Tests 1, 3, 5 workers with UNEQUAL resource distribution
# Total resources fixed at 2.5 CPU, 2.5GB RAM

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

RESULTS_FILE="/tmp/benchmark_hetero_results.txt"
echo "# Hybridgrid Heterogeneous Worker Benchmark" > "$RESULTS_FILE"
echo "# Date: $(date)" >> "$RESULTS_FILE"
echo "# Total resources per test: 4.0 CPU, 4.0GB RAM" >> "$RESULTS_FILE"
echo "# Distribution: Unequal (some workers stronger than others)" >> "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"

cleanup() {
    log "Cleaning up..."
    docker compose -f docker-compose-hetero.yml down 2>/dev/null || true
}
trap cleanup EXIT

# Generate 1 worker config (4.0 CPU)
generate_1_worker() {
    cat > docker-compose-hetero.yml << 'EOF'
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

  worker-1:
    build:
      context: ../..
      dockerfile: test/stress/Dockerfile.base
    command: hg-worker serve --coordinator=coordinator:9000 --port=50051 --advertise-address=worker-1:50051 --max-parallel=8
    networks:
      - hgnet
    depends_on:
      - coordinator
    deploy:
      resources:
        limits:
          cpus: '4.0'
          memory: 4096M

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

# Generate 3 heterogeneous workers (0.8 + 1.2 + 2.0 = 4.0 CPU)
generate_3_workers() {
    cat > docker-compose-hetero.yml << 'EOF'
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

  # Light worker
  worker-1:
    build:
      context: ../..
      dockerfile: test/stress/Dockerfile.base
    command: hg-worker serve --coordinator=coordinator:9000 --port=50051 --advertise-address=worker-1:50051 --max-parallel=2
    networks:
      - hgnet
    depends_on:
      - coordinator
    deploy:
      resources:
        limits:
          cpus: '0.8'
          memory: 820M

  # Medium worker
  worker-2:
    build:
      context: ../..
      dockerfile: test/stress/Dockerfile.base
    command: hg-worker serve --coordinator=coordinator:9000 --port=50051 --advertise-address=worker-2:50051 --max-parallel=3
    networks:
      - hgnet
    depends_on:
      - coordinator
    deploy:
      resources:
        limits:
          cpus: '1.2'
          memory: 1228M

  # Heavy worker
  worker-3:
    build:
      context: ../..
      dockerfile: test/stress/Dockerfile.base
    command: hg-worker serve --coordinator=coordinator:9000 --port=50051 --advertise-address=worker-3:50051 --max-parallel=4
    networks:
      - hgnet
    depends_on:
      - coordinator
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 2048M

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

# Generate 5 heterogeneous workers (0.5 + 0.6 + 0.8 + 1.0 + 1.1 = 4.0 CPU)
generate_5_workers() {
    cat > docker-compose-hetero.yml << 'EOF'
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

  # Lightest worker
  worker-1:
    build:
      context: ../..
      dockerfile: test/stress/Dockerfile.base
    command: hg-worker serve --coordinator=coordinator:9000 --port=50051 --advertise-address=worker-1:50051 --max-parallel=1
    networks:
      - hgnet
    depends_on:
      - coordinator
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 512M

  # Light worker
  worker-2:
    build:
      context: ../..
      dockerfile: test/stress/Dockerfile.base
    command: hg-worker serve --coordinator=coordinator:9000 --port=50051 --advertise-address=worker-2:50051 --max-parallel=2
    networks:
      - hgnet
    depends_on:
      - coordinator
    deploy:
      resources:
        limits:
          cpus: '0.6'
          memory: 614M

  # Medium worker
  worker-3:
    build:
      context: ../..
      dockerfile: test/stress/Dockerfile.base
    command: hg-worker serve --coordinator=coordinator:9000 --port=50051 --advertise-address=worker-3:50051 --max-parallel=2
    networks:
      - hgnet
    depends_on:
      - coordinator
    deploy:
      resources:
        limits:
          cpus: '0.8'
          memory: 820M

  # Medium-heavy worker
  worker-4:
    build:
      context: ../..
      dockerfile: test/stress/Dockerfile.base
    command: hg-worker serve --coordinator=coordinator:9000 --port=50051 --advertise-address=worker-4:50051 --max-parallel=2
    networks:
      - hgnet
    depends_on:
      - coordinator
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 1024M

  # Heavy worker
  worker-5:
    build:
      context: ../..
      dockerfile: test/stress/Dockerfile.base
    command: hg-worker serve --coordinator=coordinator:9000 --port=50051 --advertise-address=worker-5:50051 --max-parallel=3
    networks:
      - hgnet
    depends_on:
      - coordinator
    deploy:
      resources:
        limits:
          cpus: '1.1'
          memory: 1126M

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
    log "Cloning CPython..."
    docker compose -f docker-compose-hetero.yml exec -T builder bash -c '
        if [ ! -d /workspace/cpython ]; then
            git clone --depth=1 https://github.com/python/cpython.git /workspace/cpython
        else
            echo "CPython already cloned"
        fi
    '
}

configure_cpython() {
    log "Configuring CPython..."
    docker compose -f docker-compose-hetero.yml exec -T builder bash -c '
        cd /workspace/cpython
        ./configure --disable-test-modules --disable-perf-trampoline 2>&1 | tail -5
    '
}

run_build() {
    local config_name=$1
    local jobs=$2

    log "Running build: $config_name with -j$jobs..."

    # Clean previous build
    docker compose -f docker-compose-hetero.yml exec -T builder bash -c 'cd /workspace/cpython && make clean 2>/dev/null || true'

    # Clear cache
    docker compose -f docker-compose-hetero.yml exec -T builder bash -c 'rm -rf /root/.hybridgrid/cache/*'
    sleep 2

    local start_time=$(date +%s)

    docker compose -f docker-compose-hetero.yml exec -T builder bash -c "
        cd /workspace/cpython
        hgbuild make -j$jobs 2>&1
    " 2>&1 | tee /tmp/build_hetero_${config_name}.log

    local end_time=$(date +%s)
    local elapsed=$((end_time - start_time))

    success "Build '$config_name' completed in ${elapsed} seconds"
    echo "$elapsed"
}

start_workers() {
    local count=$1
    log "Starting $count worker(s)..."
    for i in $(seq 1 $count); do
        docker compose -f docker-compose-hetero.yml up -d "worker-$i"
    done
    sleep 5
}

run_test() {
    local workers=$1
    local config_name=$2
    local generate_func=$3
    local jobs=$4

    log ""
    log "=========================================="
    log "=== TEST: $workers Workers - $config_name ==="
    log "=========================================="

    # Generate compose file
    $generate_func

    # Stop everything
    docker compose -f docker-compose-hetero.yml down 2>/dev/null || true

    # Start services
    log "Starting coordinator and builder..."
    docker compose -f docker-compose-hetero.yml up -d coordinator builder
    sleep 5

    # Clone/configure
    clone_cpython
    configure_cpython

    # Start workers
    start_workers $workers

    # Verify
    local registered=$(docker compose -f docker-compose-hetero.yml logs coordinator 2>&1 | grep "Worker registered" | wc -l || echo "0")
    log "Workers registered: $registered"

    # Run build
    local elapsed=$(run_build "$config_name" $jobs | tail -1)
    echo "$workers,$config_name,$elapsed" >> "$RESULTS_FILE"

    echo "$elapsed"
}

main() {
    log "Starting Heterogeneous Worker Benchmark"
    log "Total: 4.0 CPU, 4.0GB RAM per test (unequally distributed)"
    echo ""

    # Build images first
    log "Building Docker images..."
    generate_1_worker
    docker compose -f docker-compose-hetero.yml build

    # Test 1: 1 worker (4.0 CPU) - baseline
    # max_parallel=8, use -j6 for headroom
    TIME_1W=$(run_test 1 "1w-4.0cpu" "generate_1_worker" 6)

    # Test 2: 3 heterogeneous workers (0.8 + 1.2 + 2.0 = 4.0 CPU)
    # max_parallel=2+3+4=9, use -j6 for headroom
    TIME_3W=$(run_test 3 "3w-hetero" "generate_3_workers" 6)

    # Test 3: 5 heterogeneous workers (0.5 + 0.6 + 0.8 + 1.0 + 1.1 = 4.0 CPU)
    # max_parallel=1+2+2+2+3=10, use -j5 (very conservative due to weak workers)
    TIME_5W=$(run_test 5 "5w-hetero" "generate_5_workers" 5)

    # Print results
    echo ""
    echo "=========================================="
    echo "  HETEROGENEOUS BENCHMARK RESULTS"
    echo "=========================================="
    echo ""
    cat "$RESULTS_FILE"
    echo ""

    success "============================================"
    success "           BENCHMARK COMPLETE"
    success "============================================"
    echo ""
    echo "Configuration Details:"
    echo ""
    echo "  1 worker:  [4.0 CPU] = 4.0 total"
    echo "  3 workers: [0.8 + 1.2 + 2.0 CPU] = 4.0 total"
    echo "  5 workers: [0.5 + 0.6 + 0.8 + 1.0 + 1.1 CPU] = 4.0 total"
    echo ""
    echo "Results Summary:"
    echo "  1 worker:   ${TIME_1W} seconds"
    echo "  3 workers:  ${TIME_3W} seconds"
    echo "  5 workers:  ${TIME_5W} seconds"
    echo ""

    # Calculate speedup
    if [ -n "$TIME_1W" ] && [ "$TIME_1W" -gt 0 ] 2>/dev/null; then
        RATIO_3=$(echo "scale=2; $TIME_1W / $TIME_3W" | bc 2>/dev/null || echo "N/A")
        RATIO_5=$(echo "scale=2; $TIME_1W / $TIME_5W" | bc 2>/dev/null || echo "N/A")
        echo "Speedup vs 1 worker:"
        echo "  3 workers: ${RATIO_3}x"
        echo "  5 workers: ${RATIO_5}x"
        echo ""
        echo "Compare with EQUAL distribution benchmark to see"
        echo "if heterogeneous workers perform better!"
    fi
}

main "$@"
