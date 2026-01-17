#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "\n${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}\n"
}

print_step() {
    echo -e "${YELLOW}▶ $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Configuration
CPYTHON_VERSION="v3.12.0"
CPYTHON_DIR="/workspace/cpython"
RESULTS_FILE="/workspace/results.txt"

print_header "HybridGrid Stress Test - CPython ${CPYTHON_VERSION}"

# Step 1: Clone CPython if not exists
if [ ! -d "$CPYTHON_DIR" ]; then
    print_step "Cloning CPython ${CPYTHON_VERSION}..."
    git clone --depth=1 --branch=${CPYTHON_VERSION} https://github.com/python/cpython.git ${CPYTHON_DIR}
    print_success "CPython cloned"
else
    print_success "CPython already exists"
fi

cd ${CPYTHON_DIR}

# Step 2: Configure (only once)
if [ ! -f "Makefile" ]; then
    print_step "Configuring CPython..."
    ./configure --disable-test-modules 2>&1 | tail -5
    print_success "Configuration complete"
else
    print_success "Already configured"
fi

# Use explicit coordinator address in Docker
COORDINATOR="${HG_COORDINATOR:-coordinator:9000}"
export HG_COORDINATOR="$COORDINATOR"

# Step 3: Check coordinator status
print_step "Checking coordinator status at ${COORDINATOR}..."
if hgbuild --coordinator=${COORDINATOR} status 2>/dev/null; then
    print_success "Coordinator is available"
    DISTRIBUTED_AVAILABLE=true
else
    print_error "Coordinator not available - will only run local test"
    DISTRIBUTED_AVAILABLE=false
fi

# Step 4: Count source files
SOURCE_COUNT=$(find . -name "*.c" | wc -l | tr -d ' ')
print_step "Found ${SOURCE_COUNT} C source files"

# Step 5: Clean build
print_step "Cleaning previous build..."
make clean 2>/dev/null || true

# Step 6: Run LOCAL build (baseline)
print_header "Test 1: Local Build (baseline)"
print_step "Building with local gcc -j4..."

LOCAL_START=$(date +%s.%N)
make -j4 2>&1 | tail -10
LOCAL_END=$(date +%s.%N)
LOCAL_TIME=$(echo "$LOCAL_END - $LOCAL_START" | bc)

print_success "Local build completed in ${LOCAL_TIME}s"

# Step 7: Clean for distributed build
print_step "Cleaning for distributed build..."
make clean

# Step 8: Run DISTRIBUTED build
if [ "$DISTRIBUTED_AVAILABLE" = true ]; then
    print_header "Test 2: Distributed Build (hgbuild)"
    print_step "Building with hgbuild make -j8..."

    DIST_START=$(date +%s.%N)
    hgbuild --coordinator=${COORDINATOR} -v make -j8 2>&1 | tail -20
    DIST_END=$(date +%s.%N)
    DIST_TIME=$(echo "$DIST_END - $DIST_START" | bc)

    print_success "Distributed build completed in ${DIST_TIME}s"

    # Step 9: Calculate speedup
    SPEEDUP=$(echo "scale=2; $LOCAL_TIME / $DIST_TIME" | bc)

    print_header "Results Summary"
    echo -e "Source files:      ${SOURCE_COUNT}"
    echo -e "Local build:       ${LOCAL_TIME}s"
    echo -e "Distributed build: ${DIST_TIME}s"
    echo -e "Speedup:           ${SPEEDUP}x"

    # Save results
    cat > ${RESULTS_FILE} << EOF
HybridGrid Stress Test Results
==============================
Date: $(date)
Project: CPython ${CPYTHON_VERSION}
Source files: ${SOURCE_COUNT}

Local build (make -j4):       ${LOCAL_TIME}s
Distributed build (hgbuild):  ${DIST_TIME}s
Speedup:                      ${SPEEDUP}x
EOF

    print_success "Results saved to ${RESULTS_FILE}"
else
    print_header "Results Summary (Local Only)"
    echo -e "Source files:      ${SOURCE_COUNT}"
    echo -e "Local build:       ${LOCAL_TIME}s"
    echo -e "Distributed:       N/A (coordinator not available)"
fi

# Step 10: Cache test (second distributed build)
if [ "$DISTRIBUTED_AVAILABLE" = true ]; then
    print_header "Test 3: Cache Hit Test"
    print_step "Rebuilding to test cache..."

    # Don't clean - should get cache hits
    touch Modules/main.c  # Touch one file to trigger partial rebuild

    CACHE_START=$(date +%s.%N)
    hgbuild --coordinator=${COORDINATOR} -v make -j8 2>&1 | grep -E '\[cache\]|\[remote\]|\[local\]' | head -20
    CACHE_END=$(date +%s.%N)
    CACHE_TIME=$(echo "$CACHE_END - $CACHE_START" | bc)

    print_success "Cache rebuild completed in ${CACHE_TIME}s"
fi

print_header "Test Complete!"
