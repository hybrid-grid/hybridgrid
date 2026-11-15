---
title: Architecture
description: How a compile request actually flows through the system, and the package map.
---

## Request flow (C/C++)

```
hgbuild cc -c main.c
  │
  ├─ 1. Parse compiler args               (internal/compiler)
  ├─ 2. Check local cache (xxhash key)     (internal/cache) → hit? return immediately
  ├─ 3. Preprocess locally (gcc -E)
  ├─ 4. Send task to coordinator
  │       └─ P2C scheduler picks a worker  (internal/coordinator/scheduler)
  │           by capability + circuit state
  ├─ 5. Worker compiles (native/Docker/MSVC executor)
  ├─ 6. Coordinator/network failure? → fall back to local compile
  │       (unless --no-fallback)
  └─ 7. Cache the result locally
```

Flutter and Unity builds skip steps 2–3 (no local preprocessing makes sense
for a whole project) and instead archive+hash the project directory, sending
the whole thing as one `BuildRequest`.

## The gRPC contract

Everything goes through one `BuildService` (see `proto/hybridgrid/v1/build.proto`):

- `Handshake` — worker → coordinator registration, reports `WorkerCapabilities`
- `Compile` — the legacy unary C/C++ path (`hgbuild cc`/`make`)
- `Build` — generic path for Flutter/Unity/Rust/Go/Node (only C++/Flutter/Unity
  have a working executor)
- `StreamBuild` — client-streamed variant for projects over ~10MB
- `HealthCheck`, `GetWorkerStatus`, `GetWorkersForBuild`, `ReportCacheHit` — housekeeping

A separate `TelemetryService` (read-only, token-gated) exists purely for the
dashboard — the coordinator's build-plane gRPC and its observation plane are
different services on the same port.

## Scheduling & resilience

- **P2C (Power of Two Choices)** — the coordinator samples two candidate
  workers and picks the less loaded, rather than tracking global state for
  every worker on every decision.
- **Circuit breaker per worker** (`gobreaker`) — `CLOSED → OPEN` after
  repeated failures, `HALF_OPEN` to probe recovery. The dashboard's worker
  heartbeat trace visibly flatlines on `OPEN`.
- **Retry with exponential backoff** (`backoff/v4`) on top of the breaker.
- **Local fallback** — if the coordinator is unreachable at all, `hgbuild`
  compiles locally rather than failing the build outright.

## Package map

| Package | Responsibility |
|---|---|
| `internal/cache` | Content-addressable compile cache (xxhash) |
| `internal/capability` | Worker hardware/compiler/Unity/Flutter detection |
| `internal/cli` | `hgbuild` subcommands |
| `internal/compiler` | Compiler arg parsing, MSVC flag translation |
| `internal/coordinator/scheduler` | P2C scheduling |
| `internal/coordinator/resilience` | Circuit breaker + retry |
| `internal/coordinator/registry` | Worker registration/heartbeats |
| `internal/discovery` | mDNS/Zeroconf |
| `internal/worker/executor` | Native, Docker, MSVC, Flutter, Unity executors |
| `internal/telemetry` | Dashboard's data model + gRPC service |
| `internal/observability/ui` | Dashboard's REST/WS server (embeds the React build) |
| `internal/security` | Token auth, TLS/mTLS, input validation |

Protobuf sources live in `proto/hybridgrid/v1/`; generated Go code in
`gen/go/hybridgrid/v1` — regenerate with `make proto-gen`, never edit directly.
