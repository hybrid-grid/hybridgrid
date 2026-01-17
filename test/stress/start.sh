#!/bin/bash
set -e

cd "$(dirname "$0")"

echo "╔═══════════════════════════════════════════════════════════════╗"
echo "║       HybridGrid Stress Test - CPython Compilation            ║"
echo "╠═══════════════════════════════════════════════════════════════╣"
echo "║  Setup:                                                       ║"
echo "║    - 1 Coordinator (0.25 CPU, 256MB)                         ║"
echo "║    - 5 Workers (0.5 CPU each, 512MB each)                    ║"
echo "║    - 1 Builder (1 CPU, 1GB)                                  ║"
echo "║                                                               ║"
echo "║  Test: CPython v3.12.0 (~500 C files)                        ║"
echo "╚═══════════════════════════════════════════════════════════════╝"
echo ""

# Step 1: Build images
echo "▶ Building Docker images..."
docker compose build --progress=plain 2>&1 | tail -5

# Step 2: Start services
echo ""
echo "▶ Starting coordinator and workers..."
docker compose up -d

# Step 3: Wait for services
echo ""
echo "▶ Waiting for services to be ready..."
sleep 5

# Check worker count
echo ""
echo "▶ Checking worker status..."
docker compose exec -T builder hgbuild workers 2>/dev/null || echo "  (Workers still connecting...)"
sleep 3
docker compose exec -T builder hgbuild workers 2>/dev/null || true

# Step 4: Run the test
echo ""
echo "▶ Starting CPython compilation test..."
echo "  This will take ~10-15 minutes..."
echo ""

docker compose exec builder bash /workspace/run-test.sh

# Step 5: Show results
echo ""
echo "▶ Test complete! View dashboard at http://localhost:8080"
echo ""
echo "To stop: docker compose down"
echo "To view logs: docker compose logs -f"
