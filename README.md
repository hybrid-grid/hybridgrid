# Hybrid-Grid Build

A distributed multi-platform build system for C/C++, Flutter, Unity, and more.

Hybrid-Grid distributes compilation tasks across multiple machines on your LAN (via mDNS auto-discovery) or WAN, dramatically reducing build times for large projects.

## v0.1.0 Release Status

### ✅ Working Features
| Feature | Status | Notes |
|---------|--------|-------|
| **C/C++ Distributed Compilation** | ✅ Production Ready | Cross-compilation across Mac/Linux/Windows |
| **mDNS Auto-Discovery** | ✅ Working | Zero-config LAN discovery |
| **Local Cache** | ✅ Working | ~10x speedup on cache hits |
| **Cache Hit Dashboard** | ✅ Working | Real-time stats reporting |
| **`hgbuild make`** | ✅ Working | Wraps make with distributed CC |
| **`hgbuild cc/c++`** | ✅ Working | Drop-in gcc/g++ replacement |
| **Local Fallback** | ✅ Working | Auto-fallback when coordinator unavailable |
| **Web Dashboard** | ✅ Working | Real-time worker/task/cache stats |
| **P2C Scheduler** | ✅ Working | Smart worker selection with scoring |
| **Circuit Breaker** | ✅ Working | Per-worker fault tolerance |
| **Docker Cross-Compile** | ✅ Working | dockcross integration |

### ⏳ Not Yet Implemented
| Feature | Status | Notes |
|---------|--------|-------|
| Flutter builds | ❌ Not Started | Planned for v0.2.0 |
| Unity builds | ❌ Not Started | Planned for v0.2.0 |
| Rust/Go/Node builds | ❌ Not Started | Planned for v0.3.0 |
| WAN Registry | ❌ Not Started | Currently LAN-only |
| Kubernetes/Helm | ❌ Not Started | Docker Compose available |
| TLS/mTLS | ⚠️ Config Only | Code exists but not tested in prod |

### Tested Configurations
- **macOS** (ARM64) → Coordinator + Worker ✅
- **Raspberry Pi** (ARM64) → Worker ✅
- **Windows WSL** (x86_64) → Worker ✅
- **100-file stress test** → 17s distributed, 3s cached ✅

## Features

- **Distributed Compilation** - Spread builds across multiple workers
- **Auto-Discovery** - mDNS-based worker discovery on LAN
- **Cross-Platform** - Support for Android, iOS, Web, Windows, Linux, macOS targets
- **Docker Integration** - Cross-compilation via dockcross images
- **Content-Addressable Cache** - Skip redundant compilations with xxhash-based caching
- **Smart Scheduling** - P2C (Power of Two Choices) algorithm with capability matching
- **Fault Tolerant** - Circuit breakers, retries with exponential backoff, local fallback
- **Observable** - Prometheus metrics, real-time web dashboard
- **Secure** - TLS/mTLS, token-based authentication

## Architecture

```
┌─────────────┐     gRPC      ┌─────────────┐
│   hgbuild   │──────────────►│  hg-coord   │
│    (CLI)    │               │(Coordinator)│
└─────────────┘               └──────┬──────┘
                                     │
                    ┌────────────────┼────────────────┐
                    │                │                │
                    ▼                ▼                ▼
             ┌──────────┐     ┌──────────┐     ┌──────────┐
             │hg-worker │     │hg-worker │     │hg-worker │
             │ (Node 1) │     │ (Node 2) │     │ (Node N) │
             └──────────┘     └──────────┘     └──────────┘
```

## Quick Start

### Zero-Config C/C++ Compilation

```bash
# On coordinator machine
hg-coord serve

# On worker machines (auto-discovers coordinator via mDNS)
hg-worker serve

# Compile your project - just prefix with 'hgbuild'
hgbuild make -j8            # Wraps make, distributes compilation
hgbuild ninja               # Wraps ninja
hgbuild cc -c main.c        # Direct gcc replacement
hgbuild c++ -c main.cpp     # Direct g++ replacement
```

### Using Docker Compose

```bash
# Start coordinator + 2 workers
docker compose up -d

# Scale to more workers
docker compose up -d --scale worker=4

# View dashboard
open http://localhost:8080
```

### Using Binaries

```bash
# Start coordinator
hg-coord serve --grpc-port=9000 --http-port=8080

# Start workers (on each machine)
hg-worker serve --coordinator=coordinator-host:9000

# If hostname resolution is problematic, use --advertise-address
hg-worker serve --coordinator=coordinator-host:9000 --advertise-address=192.168.1.50:50052

# Compile with distributed compilation
hgbuild make -j8
```

## C/C++ Compilation

### How It Works

```
hgbuild make
    │
    ├─► Sets CC="hgbuild cc", CXX="hgbuild c++"
    │
    └─► make invokes hgbuild cc/c++ for each file
            │
            ├─► 1. Parse compiler args
            ├─► 2. Check local cache → hit? return immediately
            ├─► 3. Preprocess locally (gcc -E)
            ├─► 4. Send to coordinator → worker compiles
            ├─► 5. If coordinator down → compile locally (fallback)
            └─► 6. Store result in cache
```

### Using with Make

```bash
# Simple usage
hgbuild make

# Parallel build
hgbuild make -j8

# With custom target
hgbuild make clean all

# Verbose output (shows [remote]/[local]/[cache] per file)
hgbuild -v make -j4
```

### Using with CMake

```bash
# Configure
mkdir build && cd build
cmake ..

# Build with hgbuild
hgbuild make -j8

# Or use ninja
cmake -G Ninja ..
hgbuild ninja
```

### Using with Autotools

```bash
./configure
hgbuild make -j8
```

### Direct Compiler Replacement

```bash
# Use as drop-in gcc/g++ replacement
hgbuild cc -c main.c -o main.o
hgbuild c++ -c main.cpp -o main.o -std=c++17

# All gcc/g++ flags work
hgbuild cc -O2 -Wall -I/usr/include -DDEBUG -c src.c
```

### Environment Variables

```bash
# Override coordinator address
export HG_COORDINATOR=192.168.1.100:9000

# Override compiler
export HG_CC=clang
export HG_CXX=clang++

# Then use normally
hgbuild make -j8
```

### Output Modes

With `-v` (verbose), you see the status per file:

```
[cache]  main.c → main.o (0.01s)
[remote] utils.c → utils.o (1.23s, worker-1)
[local]  config.c → config.o (2.45s, fallback)
```

### Local Fallback

When coordinator is unavailable, hgbuild automatically falls back to local compilation:

```
Warning: coordinator not available, compiling locally
```

To disable fallback (fail if no coordinator):
```bash
hgbuild --no-fallback make
```

## Installation

### Pre-built Binaries

Download from [Releases](https://github.com/h3nr1-d14z/hybridgrid/releases):

```bash
# Linux/macOS
curl -sSL https://github.com/h3nr1-d14z/hybridgrid/releases/latest/download/hybridgrid_$(uname -s | tr '[:upper:]' '[:lower:]')_$(uname -m).tar.gz | tar xz
sudo mv hg-* hgbuild /usr/local/bin/
```

### Docker Images

```bash
docker pull ghcr.io/h3nr1-d14z/hybridgrid/hg-coord:latest
docker pull ghcr.io/h3nr1-d14z/hybridgrid/hg-worker:latest
```

### From Source

```bash
git clone https://github.com/h3nr1-d14z/hybridgrid.git
cd hybridgrid
go build -o bin/ ./cmd/...
```

## Configuration

Create `~/.hybridgrid/config.yaml`:

```yaml
coordinator:
  grpc_port: 9000
  http_port: 8080
  heartbeat_ttl: 30s

worker:
  grpc_port: 50052
  http_port: 9090
  max_concurrent_tasks: 4
  coordinator: "localhost:9000"

cache:
  enabled: true
  dir: ~/.hybridgrid/cache
  max_size_gb: 10

tls:
  enabled: false
  cert_file: ""
  key_file: ""
  ca_file: ""
```

See [Configuration Reference](docs/configuration.md) for all options.

## Troubleshooting

### "no workers match requirements"

Workers may not be matching for C++ builds if:
1. Worker hasn't registered C++ capabilities (check gcc/g++/clang in PATH on worker)
2. Worker's heartbeat expired (coordinator TTL is 60s by default)
3. Architecture mismatch and Docker not available

Check coordinator logs:
```bash
# Should see: cpp_compilers=["gcc","g++","clang","clang++"]
tail -f /tmp/coord.log | grep -i worker
```

### Worker not reachable from coordinator

If coordinator can't connect back to worker:
```bash
# Use explicit advertise address
hg-worker serve --coordinator=coord:9000 --advertise-address=192.168.1.50:50052
```

### Compilation falls back to local

Reasons for local fallback:
- Coordinator not available (check `hg-coord serve` is running)
- Worker timeout (increase with `--timeout`)
- Network issues between coordinator and worker

Use verbose mode to diagnose:
```bash
hgbuild -v make -j4
```

### Cache not working

Check cache directory permissions:
```bash
ls -la ~/.hybridgrid/cache
```

Clear cache:
```bash
hgbuild cache clear
```

## Documentation

- [Installation Guide](docs/installation.md)
- [Configuration Reference](docs/configuration.md)
- [Architecture Overview](docs/architecture.md)
- [API Documentation](docs/api.md)
- [Troubleshooting](docs/troubleshooting.md)

## Monitoring

### Web Dashboard

Access at `http://coordinator:8080/` for:
- Real-time worker status
- Task statistics
- Cache hit rates
- Circuit breaker states

### Prometheus Metrics

Scrape endpoints:
- Coordinator: `http://coordinator:8080/metrics`
- Workers: `http://worker:9090/metrics`

Key metrics:
- `hybridgrid_tasks_total` - Total compilation tasks
- `hybridgrid_task_duration_seconds` - Compilation latency
- `hybridgrid_cache_hits_total` / `cache_misses_total` - Cache efficiency
- `hybridgrid_workers_total` - Connected workers
- `hybridgrid_circuit_state` - Circuit breaker status

## Development

```bash
# Run tests
go test -race ./...

# Run with coverage
go test -coverprofile=coverage.out ./...

# Lint
golangci-lint run

# Generate proto
protoc --go_out=. --go-grpc_out=. proto/build.proto
```

## License

MIT License - see [LICENSE](LICENSE) for details.
