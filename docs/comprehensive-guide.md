# Hybrid-Grid: Comprehensive Technical Documentation

> **Version:** 0.2.1
> **Last Updated:** 2026-01-21

A distributed multi-platform build system for C/C++ projects that dramatically reduces compilation times by distributing tasks across multiple machines.

---

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Installation & Quick Start](#2-installation--quick-start)
3. [How It Works](#3-how-it-works)
4. [Architecture Deep Dive](#4-architecture-deep-dive)
5. [Core Components](#5-core-components)
6. [Key Algorithms](#6-key-algorithms)
7. [Configuration Reference](#7-configuration-reference)
8. [Deployment Options](#8-deployment-options)
9. [Monitoring & Observability](#9-monitoring--observability)
10. [Cross-Platform Support](#10-cross-platform-support)
11. [Security](#11-security)
12. [Troubleshooting](#12-troubleshooting)
13. [Development Guide](#13-development-guide)

---

## 1. Project Overview

### What is Hybrid-Grid?

Hybrid-Grid is a **distributed build system** that accelerates C/C++ compilation by distributing compilation tasks across multiple machines on your local network (LAN) or wide area network (WAN). It provides:

- **Zero-configuration setup** via mDNS auto-discovery
- **Drop-in replacement** for gcc/g++/clang compilers
- **Content-addressable caching** with ~10x speedup on cache hits
- **Fault tolerance** with automatic local fallback
- **Real-time monitoring** via web dashboard

### Key Benefits

| Benefit | Description |
|---------|-------------|
| **Speed** | Distribute compilation across N workers = up to Nx faster builds |
| **Simplicity** | Just prefix your build command: `hgbuild make -j8` |
| **Zero Config** | Workers auto-discover coordinator via mDNS |
| **Resilience** | Circuit breakers + local fallback ensure builds complete |
| **Visibility** | Real-time web dashboard shows all workers and tasks |

### Production Ready Features (v0.2.1)

| Feature | Status |
|---------|--------|
| C/C++ Distributed Compilation | Production Ready |
| mDNS Auto-Discovery | Working |
| Local Cache | Working (~10x speedup) |
| Web Dashboard | Working |
| `hgbuild make/ninja` | Working |
| `hgbuild cc/c++` | Working |
| Local Fallback | Working |
| P2C Scheduler | Working |
| Circuit Breaker | Working |
| Docker Cross-Compile | Working |

### Tested Configurations

- **macOS** (ARM64/x86_64) - Coordinator + Worker
- **Linux** (ARM64/x86_64) - Coordinator + Worker
- **Windows** (x86_64/ARM64) - Worker
- **Raspberry Pi** (ARM64) - Worker
- **Stress Test**: 100 files in 17s distributed, 3s cached

---

## 2. Installation & Quick Start

### Pre-built Binaries

```bash
# Linux/macOS - Download and install
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

### Quick Start (3 Steps)

**Step 1: Start Coordinator** (on one machine)
```bash
hg-coord serve
```

**Step 2: Start Workers** (on each build machine)
```bash
hg-worker serve  # Auto-discovers coordinator via mDNS
```

**Step 3: Build Your Project**
```bash
hgbuild make -j8  # That's it!
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

---

## 3. How It Works

### High-Level Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                       COMPILATION FLOW                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. User runs: hgbuild make -j8                                │
│           │                                                     │
│           ▼                                                     │
│  2. hgbuild sets CC="hgbuild cc", CXX="hgbuild c++"           │
│           │                                                     │
│           ▼                                                     │
│  3. For each source file, make invokes hgbuild cc/c++          │
│           │                                                     │
│           ▼                                                     │
│  4. hgbuild:                                                    │
│      ├─ Parse compiler arguments                                │
│      ├─ Check local cache → Hit? Return immediately            │
│      ├─ Preprocess source locally (gcc -E)                     │
│      ├─ Send to coordinator                                     │
│      │    └─ Coordinator selects best worker                   │
│      │    └─ Worker compiles, returns object file              │
│      ├─ If coordinator down → compile locally (fallback)       │
│      └─ Store result in cache                                  │
│           │                                                     │
│           ▼                                                     │
│  5. Write .o file to disk                                      │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Detailed Compilation Pipeline

```
User: hgbuild cc -c main.c -o main.o

  ┌──────────────────────────────────────────────────────────────┐
  │ CLI (hgbuild)                                                │
  ├──────────────────────────────────────────────────────────────┤
  │ 1. Parse arguments                                           │
  │    → ParsedArgs{Compiler:"gcc", Input:["main.c"], ...}      │
  │                                                              │
  │ 2. Check distributable?                                      │
  │    → Yes: has -c flag + 1 source file                       │
  │                                                              │
  │ 3. Compute cache key                                         │
  │    → xxhash(gcc + flags + defines + source_hash)            │
  │                                                              │
  │ 4. Check local cache                                         │
  │    → Hit: Return cached object immediately (10x faster)     │
  │    → Miss: Continue to remote compilation                   │
  │                                                              │
  │ 5. Preprocess source                                         │
  │    → gcc -E main.c > preprocessed.i                         │
  │                                                              │
  │ 6. Create CompileRequest                                     │
  │    → {TaskID, Compiler, Args, PreprocessedSource, Arch}     │
  └──────────────────────────────────────────────────────────────┘
                              │
                              ▼ gRPC
  ┌──────────────────────────────────────────────────────────────┐
  │ Coordinator (hg-coord)                                       │
  ├──────────────────────────────────────────────────────────────┤
  │ 1. Receive CompileRequest                                    │
  │                                                              │
  │ 2. Select worker via Scheduler                               │
  │    → Filter by capability (C++, x86_64)                     │
  │    → Filter by OS (prefer same-OS)                          │
  │    → Select least-loaded worker                             │
  │                                                              │
  │ 3. Check circuit breaker                                     │
  │    → CLOSED: Proceed                                        │
  │    → OPEN: Try alternate worker                             │
  │                                                              │
  │ 4. Forward request to worker                                 │
  └──────────────────────────────────────────────────────────────┘
                              │
                              ▼ gRPC
  ┌──────────────────────────────────────────────────────────────┐
  │ Worker (hg-worker)                                           │
  ├──────────────────────────────────────────────────────────────┤
  │ 1. Receive CompileRequest                                    │
  │                                                              │
  │ 2. Select Executor                                           │
  │    → NativeExecutor: target_arch == native_arch             │
  │    → DockerExecutor: cross-compilation needed               │
  │    → MSVCExecutor: Windows with cl.exe                      │
  │                                                              │
  │ 3. Execute compilation                                       │
  │    → Create temp file with source                           │
  │    → Run: gcc -o main.o [args] temp.c                       │
  │    → Capture stdout/stderr/exit code                        │
  │                                                              │
  │ 4. Cache result locally                                      │
  │                                                              │
  │ 5. Return CompileResponse                                    │
  │    → {Status, ObjectFile, ExitCode, CompilationTime}        │
  └──────────────────────────────────────────────────────────────┘
                              │
                              ▼
  ┌──────────────────────────────────────────────────────────────┐
  │ CLI (hgbuild) - Continued                                    │
  ├──────────────────────────────────────────────────────────────┤
  │ 7. Receive CompileResponse                                   │
  │                                                              │
  │ 8. Write object file to disk                                 │
  │    → main.o                                                  │
  │                                                              │
  │ 9. Cache result                                              │
  │    → ~/.hybridgrid/cache/ab/ab1234567890abcd.o              │
  │                                                              │
  │ 10. Print status                                             │
  │     → "[remote] main.c → main.o (1.23s, worker-1)"          │
  └──────────────────────────────────────────────────────────────┘
```

### What Gets Distributed vs Local

| Operation | Location | Reason |
|-----------|----------|--------|
| Argument parsing | Local | Instant, no network needed |
| Cache lookup | Local | Sub-millisecond |
| Preprocessing (gcc -E) | Local | Needs local headers |
| Compilation | Remote | CPU-intensive, distributable |
| Linking | Local | Needs all .o files together |
| Assembly-only (-S) | Local | Usually small/fast |

---

## 4. Architecture Deep Dive

### System Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        HYBRID-GRID ARCHITECTURE                         │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────┐                                                        │
│  │   Client    │                                                        │
│  │  (hgbuild)  │  ◄─── Drop-in gcc/g++ replacement                     │
│  └──────┬──────┘                                                        │
│         │ gRPC :9000                                                    │
│         ▼                                                               │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                    Coordinator (hg-coord)                         │  │
│  ├──────────────────────────────────────────────────────────────────┤  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │  │
│  │  │   Registry  │  │  Scheduler  │  │   Circuit Breaker Mgr   │  │  │
│  │  │ (workers)   │  │ (selection) │  │   (fault tolerance)     │  │  │
│  │  └─────────────┘  └─────────────┘  └─────────────────────────┘  │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │  │
│  │  │   Metrics   │  │  Dashboard  │  │    mDNS Announcer       │  │  │
│  │  │ (Prometheus)│  │ (HTTP:8080) │  │   (auto-discovery)      │  │  │
│  │  └─────────────┘  └─────────────┘  └─────────────────────────┘  │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                              │                                          │
│       ┌──────────────────────┼──────────────────────┐                  │
│       │ gRPC                 │ gRPC                 │ gRPC             │
│       ▼                      ▼                      ▼                  │
│  ┌──────────┐          ┌──────────┐          ┌──────────┐             │
│  │ hg-worker│          │ hg-worker│          │ hg-worker│             │
│  │ (Node 1) │          │ (Node 2) │          │ (Node N) │             │
│  ├──────────┤          ├──────────┤          ├──────────┤             │
│  │Executors │          │Executors │          │Executors │             │
│  │• Native  │          │• Native  │          │• Native  │             │
│  │• Docker  │          │• Docker  │          │• Docker  │             │
│  │• MSVC    │          │• MSVC    │          │• MSVC    │             │
│  ├──────────┤          ├──────────┤          ├──────────┤             │
│  │  Cache   │          │  Cache   │          │  Cache   │             │
│  └──────────┘          └──────────┘          └──────────┘             │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility |
|-----------|----------------|
| **hgbuild** (CLI) | User interface, argument parsing, caching, fallback |
| **hg-coord** (Coordinator) | Worker registry, task scheduling, metrics, dashboard |
| **hg-worker** (Worker) | Task execution, local caching, capability reporting |

### Communication Protocol

All inter-component communication uses **gRPC** with Protocol Buffers:

```protobuf
service BuildService {
  // Worker registration
  rpc Handshake(HandshakeRequest) returns (HandshakeResponse);

  // Task submission
  rpc Compile(CompileRequest) returns (CompileResponse);

  // Large file streaming
  rpc StreamBuild(stream BuildChunk) returns (BuildResponse);

  // Health monitoring
  rpc HealthCheck(HealthRequest) returns (HealthResponse);

  // Status queries
  rpc GetWorkerStatus(WorkerStatusRequest) returns (WorkerStatusResponse);
}
```

---

## 5. Core Components

### 5.1 CLI (hgbuild)

**Location:** `cmd/hgbuild/`

The CLI acts as a drop-in replacement for gcc/g++/clang. It provides multiple modes:

#### Compiler Wrapper Mode
```bash
hgbuild cc -c main.c -o main.o      # Acts as gcc
hgbuild c++ -c main.cpp -o main.o   # Acts as g++
```

#### Build Tool Wrapper Mode
```bash
hgbuild make -j8     # Sets CC="hgbuild cc", then runs make
hgbuild ninja        # Sets CC="hgbuild cc", then runs ninja
```

#### Status/Management Mode
```bash
hgbuild status       # Show coordinator health
hgbuild workers      # List connected workers
hgbuild cache stats  # Show cache statistics
hgbuild cache clear  # Clear local cache
hgbuild graph        # Generate dependency graph
```

### 5.2 Coordinator (hg-coord)

**Location:** `cmd/hg-coord/`, `internal/coordinator/`

The coordinator is the central brain of the system:

```
┌─────────────────────────────────────────────────────────────┐
│                    COORDINATOR INTERNALS                     │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────┐                                        │
│  │  gRPC Server    │ ◄─── Handles Compile, Handshake, etc. │
│  └────────┬────────┘                                        │
│           │                                                 │
│           ▼                                                 │
│  ┌─────────────────┐                                        │
│  │    Registry     │ ◄─── In-memory worker store           │
│  │                 │      • Worker capabilities            │
│  │  WorkerInfo:    │      • Active task counts             │
│  │  • ID           │      • Health state                   │
│  │  • Address      │      • Last heartbeat                 │
│  │  • Capabilities │                                        │
│  │  • State        │                                        │
│  │  • Metrics      │                                        │
│  └────────┬────────┘                                        │
│           │                                                 │
│           ▼                                                 │
│  ┌─────────────────┐                                        │
│  │   Scheduler     │ ◄─── Selects best worker              │
│  │                 │      • Filter by capability           │
│  │  Algorithms:    │      • Filter by OS                   │
│  │  • LeastLoaded  │      • Sort by active tasks           │
│  │  • RoundRobin   │      • Return least loaded            │
│  └────────┬────────┘                                        │
│           │                                                 │
│           ▼                                                 │
│  ┌─────────────────┐                                        │
│  │ Circuit Breaker │ ◄─── Per-worker fault isolation       │
│  │                 │      • CLOSED → OPEN on 60% failures  │
│  │                 │      • OPEN for 60s timeout           │
│  │                 │      • HALF_OPEN for recovery test    │
│  └─────────────────┘                                        │
│                                                             │
│  ┌─────────────────┐  ┌─────────────────┐                  │
│  │    Metrics      │  │   Dashboard     │                  │
│  │  (Prometheus)   │  │  (HTTP + WS)    │                  │
│  └─────────────────┘  └─────────────────┘                  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

#### Registry

Stores worker information in-memory:

```go
type WorkerInfo struct {
    ID           string
    Address      string              // host:port
    Capabilities *pb.WorkerCapabilities
    State        WorkerState         // Idle, Busy, Unhealthy
    Metrics      WorkerMetrics       // ActiveTasks, AvgCompileTime
    LastHeartbeat time.Time
}
```

**TTL Cleanup:** Background goroutine removes workers without heartbeat after 60s.

#### Scheduler

Selects the optimal worker for each task:

```go
func (s *LeastLoadedScheduler) Select(buildType, arch, clientOS string) *WorkerInfo {
    // 1. Get workers with matching capability
    workers := s.registry.ListByCapability(buildType, arch)

    // 2. Filter by OS (prefer same-OS for native headers)
    if clientOS != "" {
        workers = filterByOS(workers, clientOS)
    }

    // 3. Exclude unhealthy workers
    workers = filterHealthy(workers)

    // 4. Sort by active tasks (ascending)
    sort.Slice(workers, func(i, j int) bool {
        return workers[i].ActiveTasks < workers[j].ActiveTasks
    })

    // 5. Return least loaded
    return workers[0]
}
```

### 5.3 Worker (hg-worker)

**Location:** `cmd/hg-worker/`, `internal/worker/`

Workers execute compilation tasks:

```
┌─────────────────────────────────────────────────────────────┐
│                      WORKER INTERNALS                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────┐                                        │
│  │  gRPC Server    │ ◄─── Receives CompileRequest          │
│  └────────┬────────┘                                        │
│           │                                                 │
│           ▼                                                 │
│  ┌─────────────────┐                                        │
│  │ Executor Manager│ ◄─── Selects appropriate executor     │
│  └────────┬────────┘                                        │
│           │                                                 │
│   ┌───────┼───────┬───────────────┐                        │
│   │       │       │               │                        │
│   ▼       ▼       ▼               ▼                        │
│ ┌─────┐ ┌─────┐ ┌─────┐    ┌─────────────┐                │
│ │Native│ │Docker│ │MSVC │    │ Capability  │                │
│ │Exec  │ │Exec  │ │Exec │    │ Detector    │                │
│ └──┬───┘ └──┬───┘ └──┬──┘    └─────────────┘                │
│    │        │        │                                      │
│    │        │        │       Detects:                       │
│    │        │        │       • CPU cores, memory           │
│    │        │        │       • Architecture (arm64, x86)   │
│    │        │        │       • Installed compilers         │
│    │        │        │       • Docker availability         │
│    ▼        ▼        ▼       • Cross-compile support       │
│  ┌─────────────────────┐                                    │
│  │   Local Compiler    │                                    │
│  │   (gcc/clang/cl)    │                                    │
│  └─────────────────────┘                                    │
│                                                             │
│  ┌─────────────────────┐                                    │
│  │     Local Cache     │ ◄─── Content-addressable store    │
│  │  ~/.hybridgrid/cache│                                    │
│  └─────────────────────┘                                    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

#### Executors

| Executor | When Used | How It Works |
|----------|-----------|--------------|
| **NativeExecutor** | Target arch = native arch | Direct gcc/clang execution |
| **DockerExecutor** | Cross-compilation | Uses dockcross images |
| **MSVCExecutor** | Windows with MSVC | Translates GCC flags to cl.exe |

#### Capability Detection

On startup, workers detect their capabilities:

```go
type WorkerCapabilities struct {
    Hostname    string
    CPUCores    int32
    MemoryMB    int64
    OS          string    // linux, darwin, windows
    Arch        string    // x86_64, arm64, armv7
    Docker      bool
    CppCapability *CppCapability  // Compilers, cross-compile support
}

type CppCapability struct {
    Compilers     []string  // ["gcc", "g++", "clang", "clang++"]
    CrossCompile  bool
    DockerImages  []string  // Available dockcross images
    MSVCVersion   string    // On Windows: "2022", "2019"
}
```

### 5.4 Cache System

**Location:** `internal/cache/`

Content-addressable cache with LRU eviction:

```
Cache Directory Structure:
~/.hybridgrid/cache/
├── ab/
│   └── ab1234567890abcd.o   ← Object files
├── cd/
│   └── cd9876543210efgh.o
└── ...

Key Generation:
┌─────────────────────────────────────────────────────┐
│  CacheKey = xxhash64(                               │
│    compiler        +    // "gcc"                    │
│    compiler_version +   // "11.4.0"                 │
│    target_arch     +    // "x86_64"                 │
│    sorted(flags)   +    // "-O2 -Wall"              │
│    sorted(defines) +    // "-DDEBUG"                │
│    source_hash          // xxhash(file_contents)   │
│  )                                                  │
└─────────────────────────────────────────────────────┘
```

**Cache Operations:**

| Operation | Description |
|-----------|-------------|
| `Get(key)` | Returns cached object if exists |
| `Put(key, data)` | Stores object, triggers LRU eviction if needed |
| `Delete(key)` | Removes specific entry |
| `Clear()` | Removes all cached objects |
| `Stats()` | Returns hit rate, size, entry count |

### 5.5 Discovery (mDNS)

**Location:** `internal/discovery/mdns/`

Zero-configuration network discovery:

```
Service Types:
• _hybridgrid-coord._tcp.local  → Coordinator
• _hybridgrid._tcp.local        → Worker

Discovery Flow:
┌──────────────┐     mDNS Query      ┌──────────────┐
│   hg-worker  │ ──────────────────► │  Network     │
│              │                     │  (multicast) │
└──────────────┘                     └──────┬───────┘
                                            │
                                            ▼
                                     ┌──────────────┐
                                     │  hg-coord    │
                                     │  responds    │
                                     └──────────────┘
```

**Fallback Chain:**
1. Command-line flag: `--coordinator=host:port`
2. Environment variable: `HG_COORDINATOR=host:port`
3. mDNS auto-discovery
4. Default: `localhost:9000`

---

## 6. Key Algorithms

### 6.1 Least-Loaded Scheduling

Selects workers with fewest active tasks:

```
Algorithm: LeastLoadedScheduler.Select()

Input: buildType, targetArch, clientOS
Output: WorkerInfo

1. workers = registry.ListByCapability(buildType, targetArch)
   // Filter workers that support C++ compilation for x86_64

2. IF clientOS specified:
     workers = workers.Filter(w => w.OS == clientOS)
   // Prefer same-OS for native header compatibility

3. workers = workers.Filter(w => w.State != UNHEALTHY)
   // Exclude unhealthy workers

4. SORT workers BY ActiveTasks ASC
   // Prioritize workers with fewer tasks

5. RETURN workers[0]
   // Return least loaded worker

Time Complexity: O(n log n) where n = worker count
```

### 6.2 Content-Addressable Caching

Uses xxhash for fast, deterministic cache keys:

```
Algorithm: CacheKey Generation

Input: CompileRequest
Output: 16-byte hex string

1. hasher = xxhash64.New()

2. hasher.Write(compiler)          // "gcc"
3. hasher.Write(compiler_version)  // "11.4.0"
4. hasher.Write(target_arch)       // "x86_64"

5. flags = SORT(compile_flags)
   FOR flag IN flags:
     hasher.Write(flag)            // "-O2", "-Wall"

6. defines = SORT(defines)
   FOR define IN defines:
     hasher.Write(define)          // "-DDEBUG"

7. source_hash = xxhash64(source_file_contents)
   hasher.Write(source_hash)

8. key = hex(hasher.Sum64())       // "ab1234567890abcd"

9. path = cache_dir + "/" + key[0:2] + "/" + key + ".o"
   // ~/.hybridgrid/cache/ab/ab1234567890abcd.o

RETURN key, path
```

**Why xxhash?**
- Non-cryptographic (speed > security for caching)
- 64-bit output (collision resistance sufficient)
- 5-10GB/s throughput on modern CPUs

### 6.3 Circuit Breaker Pattern

Per-worker fault isolation:

```
State Machine:

          ┌───────────────────────────────┐
          │                               │
          ▼                               │
     ┌─────────┐    failure_rate > 60%   │
     │ CLOSED  │ ────────────────────►   │
     │         │    (min 3 requests)      │
     └────┬────┘                          │
          │ success                       │
          │                               │
          ▼                               │
     ┌─────────┐                    ┌─────┴─────┐
     │ CLOSED  │ ◄──── success ───│ HALF_OPEN │
     └─────────┘                    └─────┬─────┘
                                          │
          ┌─────────┐   timeout 60s       │
          │  OPEN   │ ◄──────────────────┘
          │         │ ────────────────────►
          └─────────┘      failure
               ▲
               │
               └─ Rejects all requests
                  for 60 seconds

Configuration:
• MaxRequests: 3      (HALF_OPEN allows 3 test requests)
• Interval: 10s       (sliding window for CLOSED state)
• Timeout: 60s        (OPEN state duration)
• FailureRatio: 0.6   (60% failures trigger OPEN)
• MinRequests: 3      (need 3+ requests before checking ratio)
```

### 6.4 Local Fallback

Ensures builds complete even without coordinator:

```
Algorithm: LocalFallback

IF coordinator.unavailable OR timeout OR all_workers_busy:
    Log("Warning: coordinator not available, compiling locally")

    result = spawn_process(
        command: original_compiler,  // gcc
        args: original_args,         // -c main.c -o main.o
        timeout: 5 minutes
    )

    Print("[local] main.c → main.o (fallback)")
    RETURN result

ELSE:
    // Normal distributed compilation
    RETURN remote_compile(request)
```

---

## 7. Configuration Reference

### Configuration File

Create `~/.hybridgrid/config.yaml`:

```yaml
# Coordinator settings
coordinator:
  grpc_port: 9000         # gRPC server port
  http_port: 8080         # Dashboard/metrics port
  heartbeat_ttl: 60s      # Worker timeout

# Worker settings
worker:
  grpc_port: 50052        # Worker gRPC port
  http_port: 9090         # Worker metrics port
  max_concurrent: 4       # Max parallel tasks
  coordinator: "localhost:9000"  # Coordinator address

# Client settings
client:
  coordinator: "localhost:9000"
  timeout: 2m             # Compilation timeout
  fallback: true          # Enable local fallback

# Cache settings
cache:
  enabled: true
  dir: ~/.hybridgrid/cache
  max_size_gb: 10         # LRU eviction threshold
  ttl: 168h               # 1 week TTL

# Logging
log:
  level: info             # debug, info, warn, error
  format: text            # text or json

# TLS (optional)
tls:
  enabled: false
  cert_file: ""
  key_file: ""
  ca_file: ""

# Metrics
metrics:
  enabled: true
  path: /metrics
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `HG_COORDINATOR` | Coordinator address | `localhost:9000` |
| `HG_CC` | C compiler | `gcc` |
| `HG_CXX` | C++ compiler | `g++` |
| `HG_CACHE_DIR` | Cache directory | `~/.hybridgrid/cache` |
| `HG_LOG_LEVEL` | Log verbosity | `info` |

### Command-Line Flags

```bash
# Coordinator
hg-coord serve \
  --grpc-port=9000 \
  --http-port=8080 \
  --heartbeat-ttl=60s

# Worker
hg-worker serve \
  --coordinator=192.168.1.100:9000 \
  --port=50052 \
  --advertise-address=192.168.1.50:50052 \
  --max-parallel=4

# Client
hgbuild \
  --coordinator=192.168.1.100:9000 \
  --timeout=2m \
  --no-fallback \
  -v \
  make -j8
```

---

## 8. Deployment Options

### 8.1 Development (Single Machine)

```bash
# Terminal 1: Coordinator
hg-coord serve

# Terminal 2: Worker
hg-worker serve

# Terminal 3: Build
cd your-project
hgbuild make -j8
```

### 8.2 Docker Compose (Recommended)

```yaml
# docker-compose.yml
version: '3.8'

services:
  coordinator:
    image: ghcr.io/h3nr1-d14z/hybridgrid/hg-coord:latest
    ports:
      - "9000:9000"   # gRPC
      - "8080:8080"   # Dashboard
    environment:
      - HG_LOG_LEVEL=info

  worker:
    image: ghcr.io/h3nr1-d14z/hybridgrid/hg-worker:latest
    depends_on:
      - coordinator
    environment:
      - HG_COORDINATOR=coordinator:9000
    deploy:
      replicas: 4       # Scale workers
```

```bash
# Start
docker compose up -d

# Scale
docker compose up -d --scale worker=8

# View dashboard
open http://localhost:8080
```

### 8.3 LAN Deployment (mDNS)

```bash
# Machine 1 (Coordinator)
hg-coord serve
# Automatically announces via mDNS

# Machines 2-N (Workers)
hg-worker serve
# Automatically discovers coordinator

# Any machine (Client)
cd your-project
hgbuild make -j8
# Automatically discovers coordinator
```

### 8.4 Production Deployment

```bash
# 1. Start coordinator with explicit config
hg-coord serve \
  --grpc-port=9000 \
  --http-port=8080 \
  --heartbeat-ttl=60s

# 2. Start workers with explicit addresses
hg-worker serve \
  --coordinator=coord.internal:9000 \
  --advertise-address=$(hostname -I | awk '{print $1}'):50052

# 3. Configure client
export HG_COORDINATOR=coord.internal:9000
hgbuild make -j8
```

---

## 9. Monitoring & Observability

### 9.1 Web Dashboard

Access at `http://coordinator:8080/`

Features:
- Real-time worker status table
- Task completion timeline
- Cache hit rate graph
- Circuit breaker states
- WebSocket event stream

### 9.2 Prometheus Metrics

Scrape endpoints:
- Coordinator: `http://coordinator:8080/metrics`
- Workers: `http://worker:9090/metrics`

Key metrics:

```prometheus
# Task metrics
hybridgrid_tasks_total{status="success|failed"}
hybridgrid_active_tasks
hybridgrid_queued_tasks
hybridgrid_task_duration_seconds{quantile="0.5|0.9|0.99"}

# Cache metrics
hybridgrid_cache_hits_total
hybridgrid_cache_misses_total
hybridgrid_cache_size_bytes

# Worker metrics
hybridgrid_workers_total
hybridgrid_workers_healthy

# Circuit breaker
hybridgrid_circuit_state{worker="...",state="closed|open|half_open"}
```

### 9.3 Grafana Dashboard

Import dashboard for visualization:
- Task throughput over time
- Cache hit rate trends
- Worker load distribution
- Circuit breaker state changes

### 9.4 Logging

```bash
# Debug logging
HG_LOG_LEVEL=debug hg-coord serve

# JSON format (for log aggregation)
hg-coord serve --log-format=json
```

---

## 10. Cross-Platform Support

### Supported Architectures

| Architecture | Code | Notes |
|--------------|------|-------|
| x86_64 | `ARCH_X86_64` | Intel/AMD 64-bit |
| ARM64 | `ARCH_ARM64` | Apple Silicon, ARM servers |
| ARMv7 | `ARCH_ARMV7` | Raspberry Pi, embedded |

### Supported Operating Systems

| OS | Coordinator | Worker | Client |
|----|-------------|--------|--------|
| Linux | Yes | Yes | Yes |
| macOS | Yes | Yes | Yes |
| Windows | No | Yes | Yes |

### Cross-Compilation

```bash
# Native compilation (fastest)
# Client OS == Worker OS == Target OS
hgbuild cc -c main.c -o main.o

# Cross-compilation via Docker
# Uses dockcross images
hgbuild cc --target=linux-arm64 -c main.c -o main.o

# Android NDK
hgbuild cc --target=android-arm64 -c main.c -o main.o

# iOS SDK (macOS workers only)
hgbuild cc --target=ios-arm64 -c main.c -o main.o
```

### Docker Images for Cross-Compilation

| Target | Docker Image |
|--------|--------------|
| linux-arm64 | `dockcross/linux-arm64` |
| linux-armv7 | `dockcross/linux-armv7` |
| android-arm64 | `dockcross/android-arm64` |
| windows-x64 | `dockcross/windows-shared-x64` |

---

## 11. Security

### 11.1 Authentication

Token-based authentication:

```yaml
# config.yaml
auth:
  enabled: true
  token: "your-secret-token"
```

```bash
# Client
export HG_AUTH_TOKEN=your-secret-token
hgbuild make
```

### 11.2 TLS/mTLS

```yaml
# config.yaml
tls:
  enabled: true
  cert_file: /path/to/server.crt
  key_file: /path/to/server.key
  ca_file: /path/to/ca.crt  # For mTLS
```

### 11.3 Input Validation

- Path traversal prevention
- Maximum payload size limits
- Sanitized compiler arguments

---

## 12. Troubleshooting

### "no workers match requirements"

**Causes:**
1. Worker hasn't registered C++ capabilities
2. Worker's heartbeat expired (default 60s)
3. Architecture mismatch

**Solutions:**
```bash
# Check worker registration
hgbuild workers

# Check coordinator logs
tail -f /tmp/coord.log | grep worker

# Verify compilers on worker
which gcc g++ clang
```

### Worker not reachable

**Cause:** Coordinator can't connect back to worker

**Solution:**
```bash
# Use explicit advertise address
hg-worker serve \
  --coordinator=coord:9000 \
  --advertise-address=192.168.1.50:50052
```

### Compilation falls back to local

**Causes:**
- Coordinator not running
- Network issues
- All workers busy

**Diagnosis:**
```bash
# Verbose output
hgbuild -v make -j4

# Check coordinator
hgbuild status
```

### Cache not working

```bash
# Check permissions
ls -la ~/.hybridgrid/cache

# View cache stats
hgbuild cache stats

# Clear cache
hgbuild cache clear
```

---

## 13. Development Guide

### Building from Source

```bash
git clone https://github.com/h3nr1-d14z/hybridgrid.git
cd hybridgrid

# Build all binaries
go build -o bin/ ./cmd/...

# Run tests
go test -race ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Lint
golangci-lint run
```

### Project Structure

```
hybridgrid/
├── cmd/
│   ├── hg-coord/     # Coordinator entry point
│   ├── hg-worker/    # Worker entry point
│   └── hgbuild/      # CLI entry point
├── internal/
│   ├── cache/        # Content-addressable cache
│   ├── capability/   # Worker capability detection
│   ├── cli/          # CLI utilities
│   ├── compiler/     # Argument parser, preprocessor
│   ├── config/       # Configuration loading
│   ├── coordinator/  # Registry, scheduler, circuit breaker
│   ├── discovery/    # mDNS auto-discovery
│   ├── graph/        # Dependency visualization
│   ├── grpc/         # gRPC client/server
│   ├── observability/# Metrics, dashboard, tracing
│   ├── platform/     # Cross-platform utilities
│   ├── security/     # Auth, TLS, validation
│   └── worker/       # Executors, task handling
├── proto/            # Protocol Buffer definitions
├── gen/              # Generated gRPC code
├── docs/             # Documentation
└── test/             # Integration tests
```

### Generating Proto

```bash
protoc --go_out=. --go-grpc_out=. proto/build.proto
```

### Running Integration Tests

```bash
# Start test environment
docker compose -f docker-compose.test.yml up -d

# Run integration tests
go test -tags=integration ./test/...

# Cleanup
docker compose -f docker-compose.test.yml down
```

---

## Summary

Hybrid-Grid is a production-ready distributed build system that:

1. **Dramatically reduces build times** by distributing compilation across multiple machines
2. **Requires zero configuration** via mDNS auto-discovery
3. **Works as a drop-in replacement** for existing build tools
4. **Provides fault tolerance** via circuit breakers and local fallback
5. **Offers real-time visibility** through web dashboard and Prometheus metrics

For questions or issues, see:
- GitHub: https://github.com/h3nr1-d14z/hybridgrid
- Issues: https://github.com/h3nr1-d14z/hybridgrid/issues
