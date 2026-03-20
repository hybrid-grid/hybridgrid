# Hybrid-Grid Feature Guide

> Comprehensive documentation of all implemented features, how they work, and how to test them.
>
> **Last Updated:** 2026-01-21 | **Version:** 0.2.1

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Core Components](#core-components)
3. [CLI Commands](#cli-commands)
4. [Build System](#build-system)
5. [Caching](#caching)
6. [Scheduler](#scheduler)
7. [Fault Tolerance](#fault-tolerance)
8. [Dashboard & Monitoring](#dashboard--monitoring)
9. [Graph Visualization](#graph-visualization)
10. [Security](#security)
11. [Cross-Platform Support](#cross-platform-support)
12. [Testing Guide](#testing-guide)

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        Client Machine                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │   hgbuild   │  │  hgbuild cc │  │  hgbuild make/ninja     │ │
│  │   (CLI)     │  │  (wrapper)  │  │  (build tool wrapper)   │ │
│  └──────┬──────┘  └──────┬──────┘  └───────────┬─────────────┘ │
│         │                │                      │               │
│         └────────────────┼──────────────────────┘               │
│                          │                                      │
│                    Local Cache                                  │
└──────────────────────────┼──────────────────────────────────────┘
                           │ gRPC (port 9000)
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Coordinator (hg-coord)                     │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────────────┐  │
│  │ Registry │  │Scheduler │  │ Metrics  │  │   Dashboard    │  │
│  │          │  │  (P2C)   │  │(Prometheus)│ │  (HTTP :8080)  │  │
│  └──────────┘  └──────────┘  └──────────┘  └────────────────┘  │
│  ┌──────────┐  ┌──────────┐                                    │
│  │ Circuit  │  │  mDNS    │                                    │
│  │ Breakers │  │Announcer │                                    │
│  └──────────┘  └──────────┘                                    │
└──────────────────────────┬──────────────────────────────────────┘
                           │ gRPC (port 50052)
           ┌───────────────┼───────────────┐
           ▼               ▼               ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  hg-worker 1 │  │  hg-worker 2 │  │  hg-worker N │
│   (Linux)    │  │   (macOS)    │  │  (Windows)   │
│  ┌────────┐  │  │  ┌────────┐  │  │  ┌────────┐  │
│  │Native  │  │  │  │Docker  │  │  │  │ MSVC   │  │
│  │Executor│  │  │  │Executor│  │  │  │Executor│  │
│  └────────┘  │  │  └────────┘  │  │  └────────┘  │
└──────────────┘  └──────────────┘  └──────────────┘
```

---

## Core Components

### 1. hgbuild (CLI Client)

**Location:** `cmd/hgbuild/`

The primary user-facing tool for interacting with the distributed build system.

#### How It Works
1. Parses compiler arguments to extract source files, flags, and output targets
2. Checks local cache for previously compiled objects
3. If cache miss, sends compilation request to coordinator via gRPC
4. Falls back to local compilation if coordinator unavailable
5. Stores result in local cache

#### Key Features
- Drop-in replacement for gcc/g++/clang
- Build tool wrappers (make, ninja)
- Auto-discovery via mDNS
- Verbose output with status tags `[remote]`, `[cache]`, `[local]`

#### Testing
```bash
# Unit tests
go test -v ./cmd/hgbuild/...

# Manual testing - version
./hgbuild version

# Manual testing - status
./hgbuild status

# Manual testing - compile
./hgbuild cc -c main.c -o main.o

# Manual testing - with verbose output
./hgbuild cc -v -c main.c -o main.o
```

---

### 2. hg-coord (Coordinator)

**Location:** `cmd/hg-coord/`

Central orchestrator that manages workers and distributes tasks.

#### How It Works
1. Starts gRPC server (port 9000) and HTTP server (port 8080)
2. Announces service via mDNS for auto-discovery
3. Accepts worker registrations (Handshake RPC)
4. Receives compilation requests from clients
5. Schedules tasks to workers using P2C algorithm
6. Tracks worker health via heartbeats
7. Exposes metrics and dashboard

#### Key Features
- Worker registration and health tracking
- P2C (Power of Two Choices) scheduling
- Per-worker circuit breakers
- Real-time WebSocket dashboard
- Prometheus metrics endpoint

#### Configuration
```yaml
# hybridgrid.yaml
coordinator:
  grpc_port: 9000
  http_port: 8080
  heartbeat_ttl: 60s
  request_timeout: 120s
```

#### Testing
```bash
# Unit tests
go test -v ./internal/coordinator/...

# Integration test - start coordinator
./hg-coord serve --grpc-port 9000 --http-port 8080

# Check dashboard
curl http://localhost:8080/api/stats

# Check metrics
curl http://localhost:8080/metrics
```

---

### 3. hg-worker (Worker Agent)

**Location:** `cmd/hg-worker/`

Compilation agent that executes build tasks.

#### How It Works
1. Detects local system capabilities (CPU, memory, compilers)
2. Discovers coordinator via mDNS or connects to specified address
3. Registers with coordinator (Handshake)
4. Maintains heartbeat connection
5. Receives and executes compilation tasks
6. Returns compiled artifacts

#### Executor Types
| Executor | Use Case | Platforms |
|----------|----------|-----------|
| NativeExecutor | Direct gcc/clang execution | All |
| DockerExecutor | Cross-compilation via dockcross | Linux, macOS |
| MSVCExecutor | Windows MSVC compilation | Windows only |

#### Configuration
```yaml
# hybridgrid.yaml
worker:
  grpc_port: 50052
  http_port: 9090
  max_concurrent_tasks: 8  # defaults to CPU count
  task_timeout: 120s
```

#### Testing
```bash
# Unit tests
go test -v ./internal/worker/...

# Capability detection tests
go test -v ./internal/capability/...

# Manual testing - start worker
./hg-worker serve --coordinator localhost:9000

# Auto-discovery mode
./hg-worker serve
```

---

## CLI Commands

### Command Reference

| Command | Description | Example |
|---------|-------------|---------|
| `version` | Print version info | `hgbuild version` |
| `status` | Show coordinator/worker status | `hgbuild status` |
| `workers` | List available workers | `hgbuild workers -v` |
| `build` | Submit build job | `hgbuild build -f Makefile` |
| `config show` | Show current config | `hgbuild config show` |
| `config init` | Create config file | `hgbuild config init` |
| `cache stats` | Show cache statistics | `hgbuild cache stats` |
| `cache clear` | Clear local cache | `hgbuild cache clear` |
| `graph` | Generate dependency graph | `hgbuild graph -i Makefile -o graph.html` |
| `cc` | C compiler wrapper | `hgbuild cc -c main.c -o main.o` |
| `c++` | C++ compiler wrapper | `hgbuild c++ -c main.cpp -o main.o` |
| `make` | Wrap make command | `hgbuild make -j8` |
| `ninja` | Wrap ninja command | `hgbuild ninja` |
| `wrap` | Wrap any build command | `hgbuild wrap cmake --build .` |

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `HG_COORDINATOR` | Coordinator address | Auto-discover |
| `HG_CC` | C compiler override | `gcc` |
| `HG_CXX` | C++ compiler override | `g++` |
| `HG_VERBOSE` | Enable verbose output | `false` |

### Testing CLI Commands
```bash
# Test all CLI commands
go test -v ./cmd/hgbuild/...

# Test specific command
./hgbuild workers --help
./hgbuild config show
./hgbuild cache stats
```

---

## Build System

### Compiler Support

#### GCC/Clang
**Location:** `internal/compiler/parser.go`

Parses and validates compiler arguments:
- Input/output files
- Include directories (`-I`)
- Preprocessor definitions (`-D`)
- Optimization flags (`-O0` to `-O3`, `-Ofast`)
- C/C++ standards (`-std=c++17`, etc.)
- Warning flags (`-Wall`, `-Werror`)

```bash
# Test compiler parsing
go test -v ./internal/compiler/... -run TestParse
```

#### MSVC Translation
**Location:** `internal/compiler/msvc_flags.go`

Translates GCC/Clang flags to MSVC equivalents:

| GCC/Clang | MSVC |
|-----------|------|
| `-O2` | `/O2` |
| `-O3` | `/Ox` |
| `-Wall` | `/W4` |
| `-Werror` | `/WX` |
| `-std=c++17` | `/std:c++17` |
| `-I/path` | `/I/path` |
| `-DFOO=1` | `/DFOO=1` |
| `-g` | `/Zi` |
| `-fPIC` | (ignored) |

```bash
# Test MSVC translation
go test -v ./internal/compiler/... -run MSVC -cover
# Expected: 100% coverage
```

### Preprocessing
**Location:** `internal/compiler/preprocess.go`

Before sending to remote worker:
1. Run preprocessor locally to resolve includes
2. Send preprocessed source (no header dependencies)
3. Or bundle raw source + headers for cross-compilation

```bash
# Test preprocessing
go test -v ./internal/compiler/... -run Preprocess
```

---

## Caching

**Location:** `internal/cache/`

### How It Works
1. Generate cache key from: source hash + compiler + flags + architecture
2. Check if key exists in `~/.hybridgrid/cache/`
3. On hit: return cached object file
4. On miss: compile and store result

### Cache Key Generation
```go
// Key = xxhash(source + compiler + sorted_flags + arch)
key := cache.GenerateKey(source, compiler, flags, arch)
```

### Configuration
```yaml
cache:
  enabled: true
  dir: ~/.hybridgrid/cache
  max_size_gb: 10
  ttl: 168h  # 1 week
  eviction_policy: lru
```

### Cache Statistics
```bash
# View cache stats
./hgbuild cache stats

# Output:
# Cache Directory: /Users/x/.hybridgrid/cache
# Total Entries: 1,234
# Total Size: 2.3 GB
# Hit Rate: 67.8%
```

### Testing
```bash
# Unit tests
go test -v ./internal/cache/...

# Test cache key generation
go test -v ./internal/cache/... -run TestGenerateKey

# Test LRU eviction
go test -v ./internal/cache/... -run TestEviction

# Test Windows filename validation
go test -v ./internal/cache/... -run TestWindows
```

---

## Scheduler

**Location:** `internal/coordinator/scheduler/`

### P2C Algorithm (Power of Two Choices)

1. Randomly select 2 workers from available pool
2. Filter by required capabilities (arch, compiler)
3. Prefer same-OS workers (avoids header incompatibility)
4. Choose worker with fewer active tasks
5. If tie, use worker with lower latency (EWMA)

### Scoring Weights
```yaml
scheduler:
  weights:
    load: 0.4      # Active task count
    latency: 0.3   # Historical latency
    capability: 0.3 # Capability match score
```

### Testing
```bash
# Unit tests
go test -v ./internal/coordinator/scheduler/...

# Test P2C selection
go test -v ./internal/coordinator/scheduler/... -run TestP2C

# Test capability matching
go test -v ./internal/coordinator/scheduler/... -run TestCapability
```

---

## Fault Tolerance

**Location:** `internal/coordinator/resilience/`

### Circuit Breaker

Per-worker circuit breakers prevent cascading failures:

```
CLOSED ──(failures > 60%)──► OPEN ──(60s timeout)──► HALF_OPEN
   ▲                                                      │
   └──────────────(3 successes)───────────────────────────┘
```

| State | Behavior |
|-------|----------|
| CLOSED | Normal operation, track failure rate |
| OPEN | Reject all requests, wait timeout |
| HALF_OPEN | Allow 3 test requests |

### Configuration
```yaml
resilience:
  circuit_breaker:
    failure_threshold: 0.6  # 60%
    open_timeout: 60s
    half_open_requests: 3
    min_requests: 3  # Before calculating ratio
```

### Local Fallback

When coordinator unavailable:
1. Detect connection failure
2. Switch to local compilation
3. Mark output as `[local]`
4. Retry coordinator periodically

### Testing
```bash
# Unit tests
go test -v ./internal/coordinator/resilience/...

# Test circuit breaker states
go test -v ./internal/coordinator/resilience/... -run TestCircuitBreaker

# Test local fallback
go test -v ./internal/cli/fallback/...

# Chaos testing (simulated failures)
go test -v ./test/chaos/...
```

---

## Dashboard & Monitoring

**Location:** `internal/observability/`

### Dashboard Features

Access at `http://localhost:8080`

| Panel | Description |
|-------|-------------|
| Workers | Real-time worker status, CPU, memory |
| Tasks | Total, success, failed, queued counts |
| Cache | Hit rate with sparkline history |
| Recent Tasks | Latest 10 compilation tasks |

### Prometheus Metrics

Endpoint: `http://localhost:8080/metrics`

| Metric | Type | Description |
|--------|------|-------------|
| `hybridgrid_tasks_total` | Counter | Total tasks by status |
| `hybridgrid_cache_hits_total` | Counter | Cache hit count |
| `hybridgrid_cache_misses_total` | Counter | Cache miss count |
| `hybridgrid_workers_total` | Gauge | Connected workers |
| `hybridgrid_active_tasks` | Gauge | Currently running tasks |
| `hybridgrid_task_duration_seconds` | Histogram | Task execution time |
| `hybridgrid_circuit_state` | Gauge | Circuit breaker state per worker |

### WebSocket Events

Real-time updates via WebSocket at `ws://localhost:8080/ws`:
- Worker connect/disconnect
- Task start/complete/fail
- Stats updates (every 5s)

### Testing
```bash
# Unit tests
go test -v ./internal/observability/...

# Test dashboard API
curl http://localhost:8080/api/stats
curl http://localhost:8080/api/workers

# Test WebSocket (wscat required)
wscat -c ws://localhost:8080/ws
```

---

## Graph Visualization

**Location:** `internal/graph/`

### Supported Input Formats

| Format | Detection | Example |
|--------|-----------|---------|
| Makefile | File named `Makefile` or `*.mk` | `hgbuild graph -i Makefile` |
| compile_commands.json | CMake compilation database | `hgbuild graph -i compile_commands.json` |

### Output Formats

| Format | Flag | Description |
|--------|------|-------------|
| HTML | `--format html` | Interactive D3.js visualization |
| DOT | `--format dot` | Graphviz format |
| JSON | `--format json` | Raw dependency data |

### Makefile Parsing Features

- Standard rules (`target: prerequisites`)
- Pattern rules (`%.o: %.c`)
- Line continuations (`\` at end of line)
- Variable detection (limited)

### Security (XSS Protection)

All user-provided data is escaped before rendering:
```javascript
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
```

### Testing
```bash
# Unit tests
go test -v ./internal/graph/...

# Test Makefile parsing
go test -v ./internal/graph/... -run TestParseMakefile

# Test pattern rules
go test -v ./internal/graph/... -run TestPatternRule

# Test line continuation
go test -v ./internal/graph/... -run TestLineContinuation

# Test XSS protection
go test -v ./internal/graph/... -run TestEscapeHtml

# Manual testing
./hgbuild graph -i Makefile -o graph.html
open graph.html
```

---

## Security

**Location:** `internal/security/`

### Authentication

Token-based authentication for gRPC calls:

```yaml
security:
  auth:
    enabled: true
    token: "your-secret-token-here"  # Min 32 bytes
```

Features:
- Cryptographically secure token generation
- Constant-time comparison (timing attack prevention)
- gRPC interceptor enforcement

### TLS Configuration

```yaml
security:
  tls:
    enabled: true
    cert_file: /path/to/cert.pem
    key_file: /path/to/key.pem
    ca_file: /path/to/ca.pem  # For mTLS
```

### Input Validation

- Path traversal prevention in graph parser
- Windows reserved filename checking in cache
- Request parameter sanitization

### Testing
```bash
# Unit tests
go test -v ./internal/security/...

# Test token generation
go test -v ./internal/security/auth/... -run TestGenerate

# Test token validation
go test -v ./internal/security/auth/... -run TestValidate

# Test constant-time comparison
go test -v ./internal/security/auth/... -run TestConstantTime

# Test path validation
go test -v ./internal/security/validation/... -run TestPath
```

---

## Cross-Platform Support

### Platform Detection
**Location:** `internal/capability/detect.go`

| Platform | Memory Detection | Compiler Detection |
|----------|------------------|-------------------|
| Linux | `/proc/meminfo` | `which gcc` |
| macOS | `sysctl hw.memsize` | `which clang` |
| Windows | WMI query | Registry + `where cl.exe` |

### Cross-Compilation

Using dockcross Docker images:
```bash
# Linux ARM64 from x86
docker run --rm dockcross/linux-arm64 bash -c "gcc -c main.c -o main.o"
```

### MSVC Detection (Windows)
**Location:** `internal/capability/msvc.go`

Detects:
- Visual Studio versions (2022, 2019, 2017)
- Windows SDK
- Architecture (x64, x86, arm64)
- Environment setup script path

### Testing
```bash
# Capability detection
go test -v ./internal/capability/...

# Platform-specific tests
go test -v ./internal/platform/...

# MSVC detection (Windows only)
go test -v ./internal/capability/... -run MSVC
```

---

## Testing Guide

### Test Categories

| Category | Location | Command |
|----------|----------|---------|
| Unit | `*_test.go` | `go test ./...` |
| Integration | `test/integration/` | `go test ./test/integration/...` |
| Distributed | `test/distributed/` | `go test ./test/distributed/...` |
| Load | `test/load/` | `go test ./test/load/...` |
| Stress | `test/stress/` | `go test ./test/stress/...` |
| Chaos | `test/chaos/` | `go test ./test/chaos/...` |

### Running All Tests
```bash
# All unit tests
go test ./...

# With coverage
go test ./... -cover

# With coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Verbose output
go test -v ./...

# Specific package
go test -v ./internal/compiler/...

# Specific test
go test -v ./internal/cache/... -run TestEviction
```

### Integration Testing

```bash
# Start coordinator
./hg-coord serve &

# Start worker(s)
./hg-worker serve &

# Run integration tests
go test -v ./test/integration/...

# Manual integration test
echo 'int main() { return 0; }' > /tmp/test.c
./hgbuild cc -c /tmp/test.c -o /tmp/test.o
```

### Load Testing

```bash
# Run load tests
go test -v ./test/load/... -timeout 10m

# Custom concurrency
go test -v ./test/load/... -count=1 -parallel=100
```

### Coverage by Package

| Package | Target Coverage |
|---------|-----------------|
| `internal/compiler` | 90%+ |
| `internal/cache` | 85%+ |
| `internal/coordinator` | 80%+ |
| `internal/worker` | 80%+ |
| `internal/graph` | 85%+ |
| `internal/security` | 90%+ |

### Checking Coverage
```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out

# View coverage by function
go tool cover -func=coverage.out

# View detailed HTML report
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

---

## Quick Reference

### Start Full System
```bash
# Terminal 1: Coordinator
./hg-coord serve

# Terminal 2: Worker
./hg-worker serve

# Terminal 3: Build
./hgbuild make -j8
```

### Verify Installation
```bash
# Check versions
./hgbuild version
./hg-coord version
./hg-worker version

# Check connectivity
./hgbuild status
./hgbuild workers
```

### Troubleshooting
```bash
# Enable verbose logging
HG_VERBOSE=1 ./hgbuild cc -c main.c -o main.o

# Check coordinator logs
./hg-coord serve 2>&1 | tee coord.log

# Check worker logs
./hg-worker serve 2>&1 | tee worker.log

# Check metrics
curl http://localhost:8080/metrics | grep hybridgrid
```

---

## Appendix: File Locations

| Component | Location |
|-----------|----------|
| CLI | `cmd/hgbuild/` |
| Coordinator | `cmd/hg-coord/` |
| Worker | `cmd/hg-worker/` |
| Cache | `internal/cache/` |
| Compiler Parser | `internal/compiler/` |
| Configuration | `internal/config/` |
| Registry | `internal/coordinator/registry/` |
| Scheduler | `internal/coordinator/scheduler/` |
| Circuit Breaker | `internal/coordinator/resilience/` |
| mDNS Discovery | `internal/discovery/mdns/` |
| Graph Parser | `internal/graph/` |
| Dashboard | `internal/observability/dashboard/` |
| Metrics | `internal/observability/metrics/` |
| Authentication | `internal/security/auth/` |
| Worker Executor | `internal/worker/executor/` |
| Proto Definitions | `proto/hybridgrid/v1/` |
| Tests | `test/` |

---

*This document is auto-generated and maintained as part of the hybridgrid project.*
