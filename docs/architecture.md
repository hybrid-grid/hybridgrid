# Architecture Overview

## System Components

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Hybrid-Grid Build                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────┐                      ┌─────────────────────────────────┐  │
│  │   hgbuild   │◄────── gRPC ────────►│          hg-coord               │  │
│  │    (CLI)    │                      │        (Coordinator)            │  │
│  └─────────────┘                      │                                 │  │
│                                       │  ┌───────────┐ ┌─────────────┐  │  │
│                                       │  │ Registry  │ │  Scheduler  │  │  │
│                                       │  │ (Workers) │ │    (P2C)    │  │  │
│                                       │  └───────────┘ └─────────────┘  │  │
│                                       │  ┌───────────┐ ┌─────────────┐  │  │
│                                       │  │  Circuit  │ │   Metrics   │  │  │
│                                       │  │ Breakers  │ │ (Prometheus)│  │  │
│                                       │  └───────────┘ └─────────────┘  │  │
│                                       │  ┌───────────┐ ┌─────────────┐  │  │
│                                       │  │ Dashboard │ │    mDNS     │  │  │
│                                       │  │(WebSocket)│ │  Discovery  │  │  │
│                                       │  └───────────┘ └─────────────┘  │  │
│                                       └───────────────┬─────────────────┘  │
│                                                       │                    │
│                         ┌─────────────────────────────┼─────────────────┐  │
│                         │                             │                 │  │
│                         ▼                             ▼                 ▼  │
│                  ┌─────────────┐              ┌─────────────┐    ┌─────────┐│
│                  │  hg-worker  │              │  hg-worker  │    │hg-worker││
│                  │             │              │             │    │         ││
│                  │ ┌─────────┐ │              │ ┌─────────┐ │    │┌───────┐││
│                  │ │ Native  │ │              │ │ Docker  │ │    ││ Both  │││
│                  │ │Executor │ │              │ │Executor │ │    ││       │││
│                  │ └─────────┘ │              │ └─────────┘ │    │└───────┘││
│                  │ ┌─────────┐ │              │ ┌─────────┐ │    │┌───────┐││
│                  │ │  Cache  │ │              │ │  Cache  │ │    ││ Cache │││
│                  │ └─────────┘ │              │ └─────────┘ │    │└───────┘││
│                  └─────────────┘              └─────────────┘    └─────────┘│
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Component Details

### hgbuild (CLI)

Entry point for build submissions. Acts as a drop-in replacement for `gcc`/`clang`.

**Responsibilities:**
- Parse compiler arguments
- Submit compilation tasks to coordinator
- Handle local fallback on failures
- Display build progress

### hg-coord (Coordinator)

Central orchestrator managing workers and task distribution.

**Subcomponents:**
| Component | Purpose |
|-----------|---------|
| Registry | Tracks connected workers, capabilities, health |
| Scheduler | P2C algorithm for optimal worker selection |
| Circuit Breakers | Per-worker failure isolation |
| Metrics | Prometheus counters, gauges, histograms |
| Dashboard | Real-time WebSocket-based web UI |
| mDNS Browser | Auto-discovers workers on LAN |

### hg-worker (Worker)

Executes compilation tasks on distributed machines.

**Execution Modes:**
- **Native** - Direct gcc/clang execution
- **Docker** - Cross-compilation via dockcross images

## Communication Protocol

### gRPC Services

```protobuf
service BuildService {
  // Worker registration
  rpc Handshake(HandshakeRequest) returns (HandshakeResponse);

  // Compile source file
  rpc Compile(CompileRequest) returns (CompileResponse);

  // Health check
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);

  // Get worker status
  rpc GetWorkerStatus(GetWorkerStatusRequest) returns (GetWorkerStatusResponse);
}
```

### Message Flow

```
1. Worker Startup
   Worker ──► Coordinator: Handshake(capabilities)
   Worker ◄── Coordinator: HandshakeResponse(worker_id)

2. Heartbeat (every 10s)
   Worker ──► Coordinator: HealthCheck
   Worker ◄── Coordinator: HealthCheckResponse

3. Compilation
   CLI ──► Coordinator: Compile(source, args, target)
   Coordinator: Select worker via P2C
   Coordinator ──► Worker: Compile(source, args)
   Worker: Execute gcc/clang
   Worker ──► Coordinator: CompileResponse(object, stdout, stderr)
   Coordinator ──► CLI: CompileResponse
```

## Scheduling Algorithm

### Power of Two Choices (P2C)

1. Randomly select 2 workers from eligible pool
2. Score each worker based on weighted factors
3. Choose worker with higher score

**Scoring Formula:**
```
score = 0
score += 50  if native_arch_match
score += 25  if cross_compile_capable
score += 10  * cpu_cores
score += 5   * ram_gb
score -= 15  * active_tasks
score -= 0.5 * latency_ms
score += 20  if lan_source
```

## Resilience Patterns

### Circuit Breaker States

```
     ┌───────────────────────────────────────┐
     │                                       │
     ▼                                       │
┌─────────┐  failure_rate > 60%  ┌──────────┐
│ CLOSED  │─────────────────────►│   OPEN   │
└────┬────┘                      └────┬─────┘
     │                                │
     │ success                        │ timeout (30s)
     │                                │
     ▼                                ▼
┌─────────┐  success            ┌──────────┐
│ CLOSED  │◄────────────────────│HALF_OPEN │
└─────────┘                     └────┬─────┘
                                     │
                                     │ failure
                                     ▼
                                ┌──────────┐
                                │   OPEN   │
                                └──────────┘
```

### Retry with Exponential Backoff

```
Attempt 1: immediate
Attempt 2: 100ms delay
Attempt 3: 200ms delay
Attempt 4: 400ms delay (give up after max_retries)
```

### Local Fallback

When all remote workers fail:
1. Check if fallback is enabled
2. Execute compilation locally
3. Report fallback metrics

## Caching

### Content-Addressable Store

```
Cache Key = xxhash64(compiler + args + source_hash)

~/.hybridgrid/cache/
├── ab/
│   └── ab1234567890abcd.o
├── cd/
│   └── cd9876543210fedc.o
└── ...
```

### Cache Flow

```
1. CLI computes cache key
2. Check local cache
3. If hit → return cached object
4. If miss → send to coordinator
5. Coordinator checks distributed cache
6. If miss → compile on worker
7. Store result in cache
8. Return object
```

## Directory Structure

```
hybridgrid/
├── cmd/
│   ├── hg-coord/          # Coordinator binary
│   ├── hg-worker/         # Worker binary
│   └── hgbuild/           # CLI binary
├── internal/
│   ├── cache/             # Content-addressable cache
│   ├── capability/        # Hardware detection
│   ├── cli/               # CLI components
│   ├── compiler/          # GCC/Clang parser
│   ├── config/            # Configuration loading
│   ├── coordinator/
│   │   ├── metrics/       # Prometheus metrics
│   │   ├── registry/      # Worker registry
│   │   ├── resilience/    # Circuit breaker, retry
│   │   ├── scheduler/     # P2C scheduler
│   │   └── server/        # gRPC server
│   ├── discovery/
│   │   └── mdns/          # mDNS discovery
│   ├── grpc/
│   │   ├── client/        # gRPC client
│   │   └── server/        # gRPC server utils
│   ├── observability/
│   │   ├── dashboard/     # Web dashboard
│   │   └── metrics/       # Metrics helpers
│   ├── security/
│   │   ├── auth/          # Token authentication
│   │   ├── tls/           # TLS configuration
│   │   └── validation/    # Input validation
│   └── worker/
│       ├── executor/      # Native/Docker executors
│       └── server/        # Worker gRPC server
├── proto/                 # Protobuf definitions
├── gen/go/                # Generated Go code
├── test/                  # Integration tests
└── docs/                  # Documentation
```
