# Hybrid-Grid Quick Start Guide

Get up and running with distributed C/C++ compilation in 5 minutes.

---

## Installation

### macOS/Linux (Homebrew)
```bash
brew install hybrid-grid/packages/hybridgrid
```

### Windows (Scoop)
```powershell
scoop bucket add hybridgrid https://github.com/hybrid-grid/packages
scoop install hybridgrid
```

### Direct Install (curl)
```bash
curl -fsSL https://raw.githubusercontent.com/hybrid-grid/homebrew-packages/main/install.sh | bash
```

### Verify Installation
```bash
hgbuild version
# hybridgrid version 0.2.1
```

---

## 5-Minute Setup

### Step 1: Start Coordinator (Machine A)

```bash
hg-coord serve
```

Output:
```
INFO Starting coordinator on :9000 (gRPC), :8080 (HTTP)
INFO mDNS service announced: _hybridgrid._tcp.local.
INFO Dashboard available at http://localhost:8080
```

### Step 2: Start Worker(s) (Machine B, C, ...)

```bash
hg-worker serve
```

Output:
```
INFO Detected capabilities: 8 cores, 16GB RAM, gcc 13.2, clang 16.0
INFO Discovered coordinator at 192.168.1.100:9000
INFO Registered with coordinator, ready for tasks
```

### Step 3: Build Your Project

```bash
# Option A: Wrap make
hgbuild make -j8

# Option B: Use as compiler directly
hgbuild cc -c main.c -o main.o

# Option C: Wrap any build command
hgbuild wrap cmake --build . --parallel 8
```

---

## Usage Examples

### Basic Compilation

```bash
# Compile single file
hgbuild cc -c src/main.c -o build/main.o

# Compile with flags
hgbuild cc -O2 -Wall -std=c11 -c src/main.c -o build/main.o

# C++ compilation
hgbuild c++ -std=c++17 -O2 -c src/app.cpp -o build/app.o
```

### Build Tool Integration

```bash
# Make (parallel build)
hgbuild make -j$(nproc)

# Ninja
hgbuild ninja

# CMake
mkdir build && cd build
cmake ..
hgbuild wrap cmake --build . --parallel 8
```

### Check Status

```bash
# System status
hgbuild status

# List workers
hgbuild workers

# Detailed worker info
hgbuild workers -v
```

### Cache Management

```bash
# View cache stats
hgbuild cache stats

# Clear cache
hgbuild cache clear
```

### Generate Build Graph

```bash
# From Makefile
hgbuild graph -i Makefile -o build-graph.html

# Open in browser
open build-graph.html  # macOS
xdg-open build-graph.html  # Linux
```

---

## Configuration

### Config File Location

```bash
# Initialize config
hgbuild config init

# View current config
hgbuild config show
```

Config files are searched in order:
1. `./hybridgrid.yaml` (project-local)
2. `~/.hybridgrid/config.yaml` (user)
3. `/etc/hybridgrid/config.yaml` (system)

### Example Configuration

```yaml
# hybridgrid.yaml
coordinator:
  grpc_port: 9000
  http_port: 8080

worker:
  max_concurrent_tasks: 8
  task_timeout: 120s

cache:
  enabled: true
  max_size_gb: 10
  ttl: 168h

discovery:
  mdns: true
```

### Environment Variables

```bash
# Specify coordinator directly (skip mDNS)
export HG_COORDINATOR=192.168.1.100:9000

# Override compiler
export HG_CC=clang
export HG_CXX=clang++

# Enable verbose output
export HG_VERBOSE=1
```

---

## Understanding Output

### Status Tags

| Tag | Meaning |
|-----|---------|
| `[remote]` | Compiled on remote worker |
| `[cache]` | Retrieved from cache |
| `[local]` | Compiled locally (fallback) |

### Example Output

```bash
$ hgbuild make -j4
[remote] Compiling src/main.c → build/main.o (worker-1)
[cache]  Compiling src/utils.c → build/utils.o
[remote] Compiling src/network.c → build/network.o (worker-2)
[local]  Compiling src/platform.c → build/platform.o
Linking build/app
Build complete: 4 files, 2 remote, 1 cache, 1 local
```

---

## Dashboard

Access the web dashboard at `http://localhost:8080` (coordinator machine).

### Available Views

| Panel | Description |
|-------|-------------|
| **Workers** | Connected workers with CPU/memory usage |
| **Tasks** | Total, success, failed, queued counts |
| **Cache** | Hit rate with historical sparkline |
| **Recent** | Latest 10 compilation tasks |

### API Endpoints

```bash
# Get stats
curl http://localhost:8080/api/stats

# Get workers
curl http://localhost:8080/api/workers

# Prometheus metrics
curl http://localhost:8080/metrics
```

---

## Troubleshooting

### Coordinator Not Found

```bash
# Check if mDNS is working
hgbuild status

# Specify coordinator directly
HG_COORDINATOR=192.168.1.100:9000 hgbuild cc -c main.c -o main.o
```

### Worker Not Connecting

```bash
# Check coordinator is running
curl http://coordinator-ip:8080/api/stats

# Start worker with explicit coordinator
hg-worker serve --coordinator coordinator-ip:9000
```

### Falling Back to Local

Common causes:
- Coordinator unreachable
- All workers busy
- Circuit breaker open

Check with:
```bash
HG_VERBOSE=1 hgbuild cc -c main.c -o main.o
```

### Cache Issues

```bash
# Check cache directory
ls -la ~/.hybridgrid/cache/

# Clear and rebuild
hgbuild cache clear
hgbuild make -j4
```

---

## Next Steps

- **[Feature Guide](./feature-guide.md)** - Detailed feature documentation
- **[Architecture](./system-architecture.md)** - System design
- **[API Reference](./api-reference.md)** - gRPC/REST APIs
- **[Contributing](../CONTRIBUTING.md)** - Development guide

---

## Quick Reference Card

```bash
# Install
brew install hybrid-grid/packages/hybridgrid

# Start system
hg-coord serve                    # Coordinator
hg-worker serve                   # Worker(s)

# Build
hgbuild make -j8                  # Wrap make
hgbuild cc -c file.c -o file.o    # Direct compile

# Monitor
hgbuild status                    # System status
hgbuild workers -v                # Worker list
open http://localhost:8080        # Dashboard

# Cache
hgbuild cache stats               # View stats
hgbuild cache clear               # Clear cache

# Config
hgbuild config show               # View config
hgbuild config init               # Create config
```
