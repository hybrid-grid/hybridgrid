# Hybrid-Grid Build

A distributed multi-platform build system for C/C++, Flutter, Unity, and more.

Hybrid-Grid distributes compilation tasks across multiple machines on your LAN (via mDNS auto-discovery) or WAN, dramatically reducing build times for large projects.

## v0.3.0 Release Status

### ✅ Production Ready Features
| Feature | Status | Notes |
|---------|--------|-------|
| **C/C++ Distributed Compilation** | ✅ Production Ready | Cross-compilation across Mac/Linux/Windows |
| **MSVC Flag Translation** | ✅ Working | GCC/Clang to MSVC flag mapping |
| **mDNS Auto-Discovery** | ✅ Working | Zero-config LAN discovery |
| **Local Cache** | ✅ Working | ~10x speedup on cache hits |
| **Web Dashboard** | ✅ Working | Real-time worker/task/cache stats + full capabilities |
| **`hgbuild make/ninja`** | ✅ Working | Wraps build tools with distributed CC |
| **`hgbuild cc/c++`** | ✅ Working | Drop-in gcc/g++ replacement |
| **`hgbuild graph`** | ✅ Working | Build dependency visualization |
| **Local Fallback** | ✅ Working | Auto-fallback when coordinator unavailable; `--no-fallback` to disable |
| **P2C Scheduler** | ✅ Working | Smart worker selection with scoring |
| **Circuit Breaker** | ✅ Working | Per-worker fault tolerance |
| **Docker Cross-Compile** | ✅ Working | dockcross integration |
| **Colored CLI Output** | ✅ Working | Progress bars and status tags |
| **Prometheus Metrics** | ✅ Production Ready | 12/12 custom `hybridgrid_*` metrics on coordinator + workers |
| **OpenTelemetry Tracing** | ✅ Production Ready | CLI flags on both binaries, Jaeger/Zipkin compatible |
| **TLS/mTLS** | ✅ Production Ready | Full CLI flags on both binaries, certificate-based auth |
| **Worker Capabilities API** | ✅ Working | `/api/v1/workers` exposes compilers, arch, build types, Docker |
| **Health Endpoints** | ✅ Working | `/health` on coordinator (`:8080`) and workers (`:9090`) |

### ⏳ Planned Features
| Feature | Status | Notes |
|---------|--------|-------|
| Flutter builds | ❌ Planned | v0.4.0 |
| Unity builds | ❌ Planned | v0.4.0 |
| Rust/Go/Node builds | ❌ Planned | v0.4.0 |
| WAN Registry | ❌ Planned | Currently LAN-only |
| Config Validation | ❌ Planned | Runtime config checks |

### What's New in v0.3.0
- **12/12 Prometheus metrics**: All metrics now instrumented — added `fallbacks_total`, `active_tasks`, `network_transfer_bytes`, `worker_latency_ms`, `circuit_state`
- **OpenTelemetry CLI flags**: `--tracing-enable`, `--tracing-endpoint`, `--tracing-service-name` on coordinator and worker
- **TLS/mTLS CLI flags**: `--tls-cert`, `--tls-key`, `--tls-ca`, `--tls-require-client-cert` on coordinator and worker
- **`--no-fallback` flag**: Fail fast when coordinator unavailable instead of silently compiling locally
- **Worker capabilities API**: `/api/v1/workers` now returns compilers, architectures, build types, Docker availability, version
- **Worker metrics**: Workers expose `/metrics` and `/health` at `:9090`
- **Health endpoints**: Both binaries expose `/health` for Docker/K8s healthchecks
- **Stress test fix**: Exit codes now correctly propagated through `make`/`ninja` wrappers
- **Full changelog**: [v0.2.4...v0.3.0](https://github.com/hybrid-grid/hybridgrid/compare/v0.2.4...v0.3.0)

### Tested Configurations
- **macOS** (ARM64/x86_64) → Coordinator + Worker ✅
- **Linux** (ARM64/x86_64) → Coordinator + Worker ✅
- **Windows** (x86_64/ARM64) → Worker ✅
- **Raspberry Pi** (ARM64) → Worker ✅
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

- **[Quick Start Guide](docs/quick-start.md)** - Get running in 5 minutes
- **[Feature Guide](docs/feature-guide.md)** - Comprehensive feature documentation with testing
- [System Architecture](docs/system-architecture.md) - Design and components
- [Configuration Reference](docs/configuration.md) - All config options
- [API Documentation](docs/api.md) - gRPC and REST APIs
- [Troubleshooting](docs/troubleshooting.md) - Common issues and solutions

## Monitoring

### Web Dashboard

Access at `http://coordinator:8080/` for:
- Real-time worker status with full capabilities (compilers, architectures, build types, Docker)
- Task statistics and latency histograms
- Cache hit rates
- Circuit breaker states

### Health Endpoints

```bash
curl http://coordinator:8080/health   # → "OK"
curl http://worker:9090/health        # → "OK"
```

Suitable for Docker/Kubernetes healthchecks.

### Prometheus Metrics

Scrape endpoints:
- Coordinator: `http://coordinator:8080/metrics`
- Workers: `http://worker-1:9090/metrics`, `http://worker-2:9090/metrics`

Example Prometheus scrape config:
```yaml
scrape_configs:
  - job_name: 'hg-coord'
    static_configs:
      - targets: ['coordinator:8080']
  - job_name: 'hg-workers'
    static_configs:
      - targets: ['worker-1:9090', 'worker-2:9090']
```

All 12 metrics (prefix: `hybridgrid_`):

| Metric | Type | Description |
|--------|------|-------------|
| `tasks_total` | Counter | Total compilation tasks by status/type/worker |
| `task_duration_seconds` | Histogram | End-to-end compilation latency |
| `queue_time_seconds` | Histogram | Time waiting in queue before dispatch |
| `cache_hits_total` | Counter | Cache hit count |
| `cache_misses_total` | Counter | Cache miss count |
| `workers_total` | Gauge | Connected workers by state and source |
| `queue_depth` | Gauge | Tasks currently waiting |
| `active_tasks` | Gauge | Tasks in-flight per worker |
| `fallbacks_total` | Counter | Local fallback compilations by reason |
| `network_transfer_bytes` | Histogram | Upload/download bytes per task |
| `worker_latency_ms` | Histogram | gRPC round-trip latency per worker |
| `circuit_state` | Gauge | Circuit breaker state (0=closed, 1=half-open, 2=open) |

### Worker Capabilities API

```bash
curl http://coordinator:8080/api/v1/workers | jq
```

Returns full worker capabilities including:
- `compilers` — detected compilers (gcc, clang, cl.exe, …)
- `architectures` — supported target architectures
- `build_types` — supported build types (C++, Rust, Go, Flutter, …)
- `docker_available` — whether Docker cross-compilation is available
- `max_parallel_tasks`, `cpu_cores`, `memory_gb`, `os`, `version`

### OpenTelemetry Tracing

Both coordinator and worker support OTLP gRPC export (Jaeger, Zipkin, etc.):

```bash
# Start Jaeger (or any OpenTelemetry collector)
docker run -d -p 16686:16686 -p 4317:4317 jaegertracing/all-in-one

# Start coordinator with tracing enabled
hg-coord serve \
  --tracing-enable \
  --tracing-endpoint=localhost:4317 \
  --tracing-service-name=hg-coord \
  --tracing-insecure

# Start worker with tracing enabled
hg-worker serve --coordinator=localhost:9000 \
  --tracing-enable \
  --tracing-endpoint=localhost:4317 \
  --tracing-service-name=hg-worker \
  --tracing-insecure

# View traces at http://localhost:16686
```

### TLS / mTLS

```bash
# Generate self-signed certs (for testing)
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes

# Start coordinator with TLS
hg-coord serve \
  --tls-cert=cert.pem \
  --tls-key=key.pem

# Start coordinator with mTLS (requires client certs)
hg-coord serve \
  --tls-cert=cert.pem \
  --tls-key=key.pem \
  --tls-ca=ca.pem \
  --tls-require-client-cert

# Worker connecting with TLS
hg-worker serve \
  --coordinator=coordinator:9000 \
  --tls-cert=client.pem \
  --tls-key=client-key.pem \
  --tls-ca=ca.pem
```

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
