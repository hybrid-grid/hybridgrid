#!/bin/bash
# Build hybridgrid binaries for all target platforms

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
BIN_DIR="$ROOT_DIR/bin"

mkdir -p "$BIN_DIR"

echo "Building hybridgrid binaries..."
echo "Root: $ROOT_DIR"
echo "Output: $BIN_DIR"
echo ""

# macOS arm64 (M1/M2)
echo "=== macOS arm64 ==="
GOOS=darwin GOARCH=arm64 go build -o "$BIN_DIR/hg-coord-darwin-arm64" "$ROOT_DIR/cmd/hg-coord"
GOOS=darwin GOARCH=arm64 go build -o "$BIN_DIR/hg-worker-darwin-arm64" "$ROOT_DIR/cmd/hg-worker"
GOOS=darwin GOARCH=arm64 go build -o "$BIN_DIR/hgbuild-darwin-arm64" "$ROOT_DIR/cmd/hgbuild"
echo "  ✓ hg-coord-darwin-arm64"
echo "  ✓ hg-worker-darwin-arm64"
echo "  ✓ hgbuild-darwin-arm64"

# Windows amd64
echo ""
echo "=== Windows amd64 ==="
GOOS=windows GOARCH=amd64 go build -o "$BIN_DIR/hg-worker-windows-amd64.exe" "$ROOT_DIR/cmd/hg-worker"
GOOS=windows GOARCH=amd64 go build -o "$BIN_DIR/hgbuild-windows-amd64.exe" "$ROOT_DIR/cmd/hgbuild"
echo "  ✓ hg-worker-windows-amd64.exe"
echo "  ✓ hgbuild-windows-amd64.exe"

# Linux arm64 (Raspberry Pi 5)
echo ""
echo "=== Linux arm64 (Raspberry Pi) ==="
GOOS=linux GOARCH=arm64 go build -o "$BIN_DIR/hg-worker-linux-arm64" "$ROOT_DIR/cmd/hg-worker"
GOOS=linux GOARCH=arm64 go build -o "$BIN_DIR/hgbuild-linux-arm64" "$ROOT_DIR/cmd/hgbuild"
echo "  ✓ hg-worker-linux-arm64"
echo "  ✓ hgbuild-linux-arm64"

# Linux amd64 (optional, for x86 Linux machines)
echo ""
echo "=== Linux amd64 ==="
GOOS=linux GOARCH=amd64 go build -o "$BIN_DIR/hg-worker-linux-amd64" "$ROOT_DIR/cmd/hg-worker"
GOOS=linux GOARCH=amd64 go build -o "$BIN_DIR/hgbuild-linux-amd64" "$ROOT_DIR/cmd/hgbuild"
echo "  ✓ hg-worker-linux-amd64"
echo "  ✓ hgbuild-linux-amd64"

echo ""
echo "=== Build complete ==="
echo ""
ls -lh "$BIN_DIR"

echo ""
echo "=== Deployment instructions ==="
echo ""
echo "1. COORDINATOR (Mac):"
echo "   $BIN_DIR/hg-coord-darwin-arm64 serve --mdns"
echo ""
echo "2. WINDOWS WORKER:"
echo "   Copy hg-worker-windows-amd64.exe to Windows machine"
echo "   Run: hg-worker-windows-amd64.exe serve --coordinator=<MAC_IP>:9000"
echo ""
echo "3. RASPBERRY PI WORKER:"
echo "   scp $BIN_DIR/hg-worker-linux-arm64 pi@<RASPI_IP>:~/"
echo "   ssh pi@<RASPI_IP> './hg-worker-linux-arm64 serve --coordinator=<MAC_IP>:9000'"
echo ""
