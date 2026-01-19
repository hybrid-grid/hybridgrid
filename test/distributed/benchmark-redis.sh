#!/bin/bash
# Benchmark: Compile Redis using hybridgrid distributed compilation
# Run this script from Windows to utilize the Windows worker

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
HGBUILD="${SCRIPT_DIR}/../../bin/hgbuild-windows-amd64.exe"
WORK_DIR="/tmp/hgbuild-benchmark"
REDIS_VERSION="7.2.4"
REDIS_URL="https://github.com/redis/redis/archive/refs/tags/${REDIS_VERSION}.tar.gz"

echo "=== Redis Distributed Compilation Benchmark ==="
echo "Date: $(date)"
echo "Worker: Windows (x86_64)"

# Check hgbuild
if [ ! -f "$HGBUILD" ]; then
    echo "Error: hgbuild-windows-amd64.exe not found"
    echo "For Linux/Mac, update HGBUILD path above"
    exit 1
fi

# Create work directory
rm -rf "$WORK_DIR"
mkdir -p "$WORK_DIR"
cd "$WORK_DIR"

# Download Redis
echo ""
echo "=== Downloading Redis ${REDIS_VERSION} ==="
curl -L -o redis.tar.gz "$REDIS_URL"
tar -xzf redis.tar.gz
cd "redis-${REDIS_VERSION}"

# Count C files
C_FILES=$(find src -name "*.c" | wc -l)
echo "Found $C_FILES C source files"

# Benchmark 1: Local compilation
echo ""
echo "=== Benchmark 1: Local Compilation (make -j8) ==="
make clean > /dev/null 2>&1 || true
START=$(date +%s)
make -j8 MALLOC=libc > /dev/null 2>&1
END=$(date +%s)
LOCAL_TIME=$((END - START))
echo "Local build time: ${LOCAL_TIME}s"

# Benchmark 2: Distributed compilation
echo ""
echo "=== Benchmark 2: Distributed Compilation (hgbuild make) ==="
make clean > /dev/null 2>&1 || true

# Set CC/CXX to use hgbuild wrapper
export CC="$HGBUILD cc"
export CXX="$HGBUILD c++"

START=$(date +%s)
# Use hgbuild make wrapper if available, or regular make with CC override
make -j8 MALLOC=libc > /dev/null 2>&1 || {
    echo "hgbuild make failed, trying with CC override"
    make -j8 MALLOC=libc CC="$HGBUILD cc" > /dev/null 2>&1
}
END=$(date +%s)
DIST_TIME=$((END - START))
echo "Distributed build time: ${DIST_TIME}s"

# Results
echo ""
echo "=== Results ==="
echo "Local:       ${LOCAL_TIME}s"
echo "Distributed: ${DIST_TIME}s"
if [ "$DIST_TIME" -lt "$LOCAL_TIME" ]; then
    SPEEDUP=$(echo "scale=2; $LOCAL_TIME / $DIST_TIME" | bc)
    echo "Speedup:     ${SPEEDUP}x"
else
    SLOWDOWN=$(echo "scale=2; $DIST_TIME / $LOCAL_TIME" | bc)
    echo "Slowdown:    ${SLOWDOWN}x (overhead too high for this project size)"
fi

# Cleanup
cd /
rm -rf "$WORK_DIR"

echo ""
echo "=== Benchmark Complete ==="
