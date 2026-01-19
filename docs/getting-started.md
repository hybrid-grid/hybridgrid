# Getting Started with Hybrid-Grid

This guide will help you set up Hybrid-Grid for distributed C/C++ compilation in under 5 minutes.

## Prerequisites

Before you begin, ensure you have:

- **Go 1.21+** (for building from source)
- **GCC/Clang** (for compilation on workers)
- **Network access** between machines (LAN or WAN)

## 5-Minute Quick Start

### Step 1: Install

Choose your installation method:

```bash
# Option A: Pre-built binaries (recommended)
curl -sSL https://raw.githubusercontent.com/hybrid-grid/homebrew-packages/main/install.sh | bash

# Option B: Homebrew (macOS/Linux)
brew tap hybrid-grid/packages
brew install hybridgrid

# Option C: Scoop (Windows)
scoop bucket add hybridgrid https://github.com/hybrid-grid/homebrew-packages
scoop install hybridgrid
```

### Step 2: Start Coordinator

On your main machine (coordinator):

```bash
hg-coord serve
```

You should see:
```
INFO  Coordinator starting
INFO  gRPC server listening on :9000
INFO  HTTP dashboard at http://localhost:8080
INFO  mDNS announcer started
```

### Step 3: Start Worker(s)

On each worker machine:

```bash
# Auto-discovers coordinator via mDNS
hg-worker serve
```

You should see:
```
INFO  Worker starting
INFO  Discovered coordinator at 192.168.1.100:9000
INFO  Connected to coordinator
INFO  Capabilities: [gcc g++ clang clang++]
```

### Step 4: Run Your First Distributed Build

From any machine on the network:

```bash
# Navigate to your C/C++ project
cd my-project

# Run make with distributed compilation
hgbuild make -j8
```

### Step 5: View Dashboard

Open your browser to `http://coordinator-ip:8080` to see:
- Connected workers
- Active/completed tasks
- Cache hit statistics
- Build performance metrics

## Installation Options

### Pre-built Binaries

Download from [GitHub Releases](https://github.com/hybrid-grid/hybridgrid/releases):

```bash
# Linux x86_64
curl -LO https://github.com/hybrid-grid/hybridgrid/releases/latest/download/hybridgrid_linux_amd64.tar.gz
tar xzf hybridgrid_linux_amd64.tar.gz
sudo mv hg-coord hg-worker hgbuild /usr/local/bin/

# macOS ARM64 (Apple Silicon)
curl -LO https://github.com/hybrid-grid/hybridgrid/releases/latest/download/hybridgrid_darwin_arm64.tar.gz
tar xzf hybridgrid_darwin_arm64.tar.gz
sudo mv hg-coord hg-worker hgbuild /usr/local/bin/

# Windows
# Download from releases page and add to PATH
```

### Homebrew (macOS/Linux)

```bash
brew tap hybrid-grid/packages
brew install hybridgrid
```

### Scoop (Windows)

```powershell
scoop bucket add hybridgrid https://github.com/hybrid-grid/homebrew-packages
scoop install hybridgrid
```

### Docker

```bash
# Pull images
docker pull ghcr.io/hybrid-grid/hg-coord:latest
docker pull ghcr.io/hybrid-grid/hg-worker:latest

# Run coordinator
docker run -d --name hg-coord \
  -p 9000:9000 -p 8080:8080 \
  ghcr.io/hybrid-grid/hg-coord:latest

# Run worker
docker run -d --name hg-worker \
  -e COORDINATOR=host.docker.internal:9000 \
  ghcr.io/hybrid-grid/hg-worker:latest
```

### Docker Compose

```bash
# Clone repository
git clone https://github.com/hybrid-grid/hybridgrid.git
cd hybridgrid

# Start 1 coordinator + 2 workers
docker compose up -d

# Scale workers
docker compose up -d --scale worker=4
```

### From Source

```bash
# Clone and build
git clone https://github.com/hybrid-grid/hybridgrid.git
cd hybridgrid
go build -o bin/ ./cmd/...

# Add to PATH
export PATH=$PATH:$(pwd)/bin
```

## Configuration

### Config File Location

Hybrid-Grid looks for configuration in:
1. `./hybridgrid.yaml` (current directory)
2. `~/.hybridgrid/config.yaml`
3. `/etc/hybridgrid/config.yaml`

### Create Default Config

```bash
hgbuild config init
```

### Example Configuration

```yaml
coordinator:
  grpc_port: 9000
  http_port: 8080
  heartbeat_ttl: 30s

worker:
  grpc_port: 50052
  max_concurrent_tasks: 4
  coordinator: "auto"  # Use mDNS discovery

client:
  coordinator_addr: ""  # Empty = auto-discover
  timeout: 2m
  fallback_enabled: true

cache:
  enabled: true
  dir: ~/.hybridgrid/cache
  max_size_gb: 10
  ttl_hours: 168  # 1 week

log:
  level: info
  format: console  # or json
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `HG_COORDINATOR` | Coordinator address | Auto-discover |
| `HG_CC` | C compiler | `gcc` |
| `HG_CXX` | C++ compiler | `g++` |
| `HG_VERBOSE` | Enable verbose output | `0` |

## Usage Examples

### Basic Make

```bash
# Standard make
hgbuild make

# Parallel build
hgbuild make -j8

# With target
hgbuild make clean all
```

### CMake Projects

```bash
mkdir build && cd build
cmake ..
hgbuild make -j8

# Or with Ninja
cmake -G Ninja ..
hgbuild ninja -j8
```

### Direct Compilation

```bash
# C file
hgbuild cc -c main.c -o main.o

# C++ with flags
hgbuild c++ -c main.cpp -o main.o -std=c++17 -O2 -Wall
```

### Verbose Output

```bash
hgbuild -v make -j4
```

Output shows source of each compilation:
```
[cache]  utils.c → utils.o (0.01s)
[remote] main.c → main.o (1.23s, worker-mac-1)
[local]  config.c → config.o (2.45s, fallback)
```

## Verify Installation

### Check Status

```bash
hgbuild status
```

Expected output:
```
Coordinator Status
──────────────────
Address:      192.168.1.100:9000
Status:       healthy ✓
Active Tasks: 0
Queued Tasks: 0
```

### List Workers

```bash
hgbuild workers
```

Expected output:
```
Workers: 3 total, 3 healthy

ID                   ARCH       CORES    STATUS
worker-mac-1         ARM64      8        healthy
worker-linux-1       X86_64     16       healthy
worker-pi-1          ARM64      4        healthy
```

### Cache Statistics

```bash
hgbuild cache stats
```

## Troubleshooting

### Worker Not Discovered

If mDNS discovery fails:

```bash
# Specify coordinator explicitly
hg-worker serve --coordinator=192.168.1.100:9000
```

### Compilation Falls Back to Local

Check coordinator connectivity:
```bash
hgbuild status
```

Enable verbose mode to see reason:
```bash
hgbuild -v make
```

### Cache Not Working

Check cache directory:
```bash
hgbuild cache stats
```

Clear if corrupted:
```bash
hgbuild cache clear
```

### Network Issues

If coordinator can't reach workers:
```bash
# Worker: Use explicit advertise address
hg-worker serve --advertise-address=192.168.1.50:50052
```

## Next Steps

- [Configuration Reference](configuration.md) - All config options
- [Architecture Overview](architecture.md) - System design
- [Architecture Diagrams](architecture-diagrams.md) - Visual diagrams
- [Troubleshooting](troubleshooting.md) - Common issues
- [API Documentation](api.md) - gRPC API reference

## Performance Tips

1. **Use parallel builds**: Always use `-j` flag with make/ninja
2. **Keep workers warm**: Persistent workers avoid startup overhead
3. **Local cache**: Enables instant rebuilds of unchanged files
4. **Network proximity**: Place workers on same LAN for best performance
5. **Match architectures**: Workers with same arch avoid cross-compilation overhead
