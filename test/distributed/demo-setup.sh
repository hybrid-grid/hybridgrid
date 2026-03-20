#!/bin/bash
# Demo setup: Cross-compile hybridgrid for all platforms
# Coordinator: macOS (this machine)
# Workers: Raspberry Pi 5 (linux/arm64), Windows PC (windows/amd64)

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
OUTPUT_DIR="$SCRIPT_DIR/binaries"

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${BLUE}[setup]${NC} $1"; }
success() { echo -e "${GREEN}[setup]${NC} $1"; }
warn() { echo -e "${YELLOW}[setup]${NC} $1"; }

cd "$PROJECT_ROOT"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "v0.2.1-demo")
LDFLAGS="-X main.version=$VERSION"

mkdir -p "$OUTPUT_DIR"/{mac,raspi,windows}

# ============================================
# 1. macOS (coordinator + hgbuild)
# ============================================
log "Building for macOS (this machine)..."
GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o "$OUTPUT_DIR/mac/hg-coord" ./cmd/hg-coord
GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o "$OUTPUT_DIR/mac/hg-worker" ./cmd/hg-worker
GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o "$OUTPUT_DIR/mac/hgbuild" ./cmd/hgbuild
success "macOS binaries ready"

# ============================================
# 2. Raspberry Pi 5 (linux/arm64 worker)
# ============================================
log "Cross-compiling for Raspberry Pi 5 (linux/arm64)..."
GOOS=linux GOARCH=arm64 go build -ldflags "$LDFLAGS" -o "$OUTPUT_DIR/raspi/hg-worker" ./cmd/hg-worker
GOOS=linux GOARCH=arm64 go build -ldflags "$LDFLAGS" -o "$OUTPUT_DIR/raspi/hgbuild" ./cmd/hgbuild
success "Raspberry Pi binaries ready"

# ============================================
# 3. Windows PC (windows/amd64 worker)
# ============================================
log "Cross-compiling for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "$OUTPUT_DIR/windows/hg-worker.exe" ./cmd/hg-worker
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "$OUTPUT_DIR/windows/hgbuild.exe" ./cmd/hgbuild
success "Windows binaries ready"

# ============================================
# Summary
# ============================================
echo ""
success "============================================"
success "  All binaries built successfully!"
success "============================================"
echo ""
echo "📁 $OUTPUT_DIR/"
echo "├── mac/"
echo "│   ├── hg-coord      (coordinator)"
echo "│   ├── hg-worker     (optional local worker)"
echo "│   └── hgbuild       (build client)"
echo "├── raspi/"
echo "│   ├── hg-worker     (worker)"
echo "│   └── hgbuild       (optional client)"
echo "└── windows/"
echo "    ├── hg-worker.exe (worker)"
echo "    └── hgbuild.exe   (optional client)"
echo ""
echo "=========================================="
echo "  NEXT STEPS"
echo "=========================================="
echo ""
echo "1. START COORDINATOR (this Mac):"
echo "   $OUTPUT_DIR/mac/hg-coord serve --grpc-port=9000 --http-port=8080"
echo ""
echo "2. COPY & START WORKER on Raspberry Pi 5:"
echo "   scp $OUTPUT_DIR/raspi/hg-worker pi@<RASPI_IP>:~/"
echo "   ssh pi@<RASPI_IP> './hg-worker serve --coordinator=<MAC_IP>:9000'"
echo ""
echo "3. COPY & START WORKER on Windows PC:"
echo "   # Copy windows/hg-worker.exe to Windows PC"
echo "   # Run: hg-worker.exe serve --coordinator=<MAC_IP>:9000"
echo ""
echo "4. CLONE TEST PROJECT (on this Mac):"
echo "   git clone --depth=1 https://github.com/redis/redis /tmp/redis"
echo "   # OR for bigger test:"
echo "   git clone --depth=1 https://github.com/python/cpython /tmp/cpython"
echo ""
echo "5. RUN BENCHMARK:"
echo "   cd /tmp/redis && make clean"
echo "   # Baseline:"
echo "   time make -j4"
echo "   # Distributed:"
echo "   make clean && time $OUTPUT_DIR/mac/hgbuild make -j6"
echo ""
echo "6. DASHBOARD: http://localhost:8080"
