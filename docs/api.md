# API Documentation

> **Architecture split (v0.5):** the live cluster dashboard is now served by a standalone `hg-dashboard` binary. It exposes the REST API (`/api/v1/*`) and browser WebSocket (`/ws`) over HTTP, and pulls cluster state from the coordinator's gRPC `TelemetryService` (`:9000`). The coordinator itself keeps a small **ops HTTP server** (`/health`, `/metrics`, `/log-level`) on `--http-port` (default 8080). Reusable, stack-agnostic clients (including `hg-dashboard`) consume `TelemetryService` directly.

## gRPC Build API

### Service Definition

```protobuf
syntax = "proto3";
package hybridgrid.v1;

service BuildService {
  rpc Handshake(HandshakeRequest) returns (HandshakeResponse);
  rpc Compile(CompileRequest) returns (CompileResponse);
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
  rpc GetWorkerStatus(GetWorkerStatusRequest) returns (GetWorkerStatusResponse);
}
```

### Handshake

Register a worker with the coordinator.

**Request:**
```protobuf
message HandshakeRequest {
  string worker_id = 1;           // Unique worker identifier
  WorkerCapabilities capabilities = 2;
}

message WorkerCapabilities {
  repeated string architectures = 1;  // ["amd64", "arm64"]
  repeated string compilers = 2;      // ["gcc", "clang"]
  int32 cpu_cores = 3;
  int64 memory_bytes = 4;
  bool docker_available = 5;
}
```

**Response:**
```protobuf
message HandshakeResponse {
  bool accepted = 1;
  string assigned_id = 2;         // Coordinator-assigned ID
  string message = 3;
}
```

### Compile

Submit a compilation task.

**Request:**
```protobuf
message CompileRequest {
  string source_file = 1;         // Source filename
  bytes source_content = 2;       // Source file content
  string compiler = 3;            // "gcc" or "clang"
  repeated string args = 4;       // Compiler arguments
  string target_arch = 5;         // Target architecture
  string target_os = 6;           // Target OS
  map<string, bytes> includes = 7; // Header files
}
```

**Response:**
```protobuf
message CompileResponse {
  bool success = 1;
  bytes object_file = 2;          // Compiled object
  string stdout = 3;
  string stderr = 4;
  int32 exit_code = 5;
  int64 compile_time_ms = 6;
  string worker_id = 7;           // Which worker compiled
  bool from_cache = 8;            // Cache hit
}
```

### HealthCheck

Check worker/coordinator health.

**Request:**
```protobuf
message HealthCheckRequest {
  string worker_id = 1;           // Optional: specific worker
}
```

**Response:**
```protobuf
message HealthCheckResponse {
  bool healthy = 1;
  string status = 2;              // "ok", "degraded", "unhealthy"
  int32 active_tasks = 3;
  int64 uptime_seconds = 4;
}
```

### GetWorkerStatus

Get detailed worker status.

**Request:**
```protobuf
message GetWorkerStatusRequest {
  string worker_id = 1;           // Empty for all workers
}
```

**Response:**
```protobuf
message GetWorkerStatusResponse {
  repeated WorkerStatus workers = 1;
}

message WorkerStatus {
  string worker_id = 1;
  string address = 2;
  WorkerCapabilities capabilities = 3;
  int32 active_tasks = 4;
  int32 total_tasks = 5;
  int32 successful_tasks = 6;
  int64 avg_compile_time_ms = 7;
  string circuit_state = 8;       // "closed", "open", "half_open"
  string source = 9;              // "mdns", "wan", "static"
}
```

## gRPC TelemetryService (read-only observation plane)

**Location:** `proto/hybridgrid/v1/telemetry.proto`, registered on the coordinator's `:9000` gRPC server alongside `BuildService`. Read-only by design: no RPC here can alter build-plane state. Authentication mirrors the build plane: when the coordinator is configured with a token, every request must carry it as the `auth_token` field.

| RPC | Request | Response | Notes |
|-----|---------|----------|-------|
| `GetStats` | `GetTelemetryStatsRequest{auth_token}` | `GetTelemetryStatsResponse{stats}` | Cluster-wide aggregate stats (workers, tasks, cache, uptime) |
| `ListWorkers` | `ListTelemetryWorkersRequest{auth_token}` | `ListTelemetryWorkersResponse{workers[]}` | Registered workers as the coordinator sees them |
| `ListTasks` | `ListTelemetryTasksRequest{auth_token, limit?}` | `ListTelemetryTasksResponse{tasks[]}` | Recent tasks, newest first |
| `ListBuilds` | `ListTelemetryBuildsRequest{auth_token}` | `ListTelemetryBuildsResponse{builds[]}` | Logical build aggregates |
| `GetBuild` | `GetTelemetryBuildRequest{auth_token, build_id}` | `GetTelemetryBuildResponse{build, tasks[]}` | One build with retained tasks; `NOT_FOUND` when unknown |
| `GetConsole` | `GetTelemetryConsoleRequest{auth_token, task_id}` | `GetTelemetryConsoleResponse{console}` | Retained console output; `NOT_FOUND` when the task is unknown |
| `StreamEvents` | `StreamTelemetryEventsRequest{auth_token}` | `stream TelemetryEvent` | Live event stream: replays recent events on connect, then pushes new ones until the client disconnects. `TelemetryEvent{type, payload_json}`; types: `stats`, `task_started`, `task_completed`, `worker_added`, `worker_removed`, `ping` |

Clients (`hg-dashboard`) pass `--coordinator-token` as `auth_token` on every call. `StreamEvents` feeds the browser WebSocket fan-out: each `TelemetryEvent` is translated into a WS message `{"type","timestamp","data":<parsed payload_json>}`.

## Coordinator Ops HTTP API

Served by `hg-coord` on `--http-port` (default 8080). These are operator/metrics endpoints — **not** the dashboard.

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Liveness/readiness, returns `200 "ok"` |
| `/metrics` | GET | Prometheus metrics (see "Prometheus Metrics" below) |
| `/log-level` | GET/PUT | Inspect/change the runtime log level; requires the configured auth token (open when no token is set) |

## Dashboard HTTP API (served by `hg-dashboard`)

Served by the standalone `hg-dashboard` binary. By default it listens on `:8080`; in the test compose files it is published on the host as `:8081`. It serves the SPA, the REST API, and the browser WebSocket, all backed by a gRPC `TelemetryService` client to the coordinator (`--coordinator`, `:9000`).

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Web dashboard SPA |
| `/health` | GET | Health check (returns `200 "ok"`) |
| `/api/v1/stats` | GET | JSON stats for dashboard |
| `/api/v1/workers` | GET | Registered workers |
| `/api/v1/tasks?limit=N` | GET | Recent tasks |
| `/api/v1/builds` | GET | Logical build aggregates |
| `/api/v1/builds/{id}` | GET | One build + its retained tasks |
| `/api/v1/tasks/{id}/console` | GET | Retained console output for a task |
| `/api/v1/events` | GET | Retained events |
| `/ws` | WebSocket | Real-time updates (populated from `StreamEvents`) |

### GET /api/v1/stats

Returns current system statistics.

**Response:**
```json
{
  "workers": {
    "total": 4,
    "healthy": 3,
    "unhealthy": 1
  },
  "tasks": {
    "active": 12,
    "queued": 5,
    "total": 1523,
    "successful": 1498,
    "failed": 25
  },
  "cache": {
    "hits": 892,
    "misses": 631,
    "hit_rate": 0.586
  },
  "uptime_seconds": 86400
}
```

### WebSocket /ws (hg-dashboard)

Receives real-time events.

**Event Types:**
```json
// Worker connected
{
  "type": "worker_joined",
  "worker_id": "worker-1",
  "address": "192.168.1.10:50052",
  "timestamp": "2026-01-17T10:30:00Z"
}

// Worker disconnected
{
  "type": "worker_left",
  "worker_id": "worker-1",
  "reason": "heartbeat_timeout",
  "timestamp": "2026-01-17T10:35:00Z"
}

// Task completed
{
  "type": "task_completed",
  "task_id": "task-123",
  "worker_id": "worker-2",
  "success": true,
  "duration_ms": 234,
  "timestamp": "2026-01-17T10:30:05Z"
}

// Circuit state change
{
  "type": "circuit_state_changed",
  "worker_id": "worker-3",
  "old_state": "closed",
  "new_state": "open",
  "timestamp": "2026-01-17T10:31:00Z"
}

// Stats update (every 5s)
{
  "type": "stats_update",
  "stats": { /* same as /api/v1/stats */ },
}

## Prometheus Metrics

### Coordinator Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `hybridgrid_tasks_total` | Counter | Total compilation tasks |
| `hybridgrid_tasks_success_total` | Counter | Successful compilations |
| `hybridgrid_tasks_failed_total` | Counter | Failed compilations |
| `hybridgrid_active_tasks` | Gauge | Currently running tasks |
| `hybridgrid_queued_tasks` | Gauge | Tasks waiting for workers |
| `hybridgrid_task_duration_seconds` | Histogram | Compilation latency |
| `hybridgrid_workers_total` | Gauge | Connected workers |
| `hybridgrid_workers_healthy` | Gauge | Healthy workers |
| `hybridgrid_cache_hits_total` | Counter | Cache hits |
| `hybridgrid_cache_misses_total` | Counter | Cache misses |
| `hybridgrid_circuit_state` | Gauge | Circuit breaker state (0=closed, 1=half_open, 2=open) |

### Worker Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `hybridgrid_worker_tasks_total` | Counter | Tasks processed by this worker |
| `hybridgrid_worker_task_duration_seconds` | Histogram | Local compilation time |
| `hybridgrid_worker_active_tasks` | Gauge | Currently running tasks |
| `hybridgrid_worker_cpu_usage` | Gauge | CPU utilization |
| `hybridgrid_worker_memory_usage_bytes` | Gauge | Memory usage |

### Labels

Common labels across metrics:
- `worker_id` - Worker identifier
- `compiler` - Compiler used (gcc, clang)
- `target_arch` - Target architecture
- `status` - Task status (success, failed, timeout)

## Error Codes

### gRPC Status Codes

| Code | Meaning | Retry? |
|------|---------|--------|
| `OK` | Success | N/A |
| `INVALID_ARGUMENT` | Bad request | No |
| `NOT_FOUND` | Worker not found | No |
| `RESOURCE_EXHAUSTED` | Rate limited | Yes (backoff) |
| `UNAVAILABLE` | Service down | Yes |
| `DEADLINE_EXCEEDED` | Timeout | Yes |
| `INTERNAL` | Server error | Yes |

### Application Error Codes

Returned in `CompileResponse.stderr`:
```
HYBRIDGRID_ERR_001: Compiler not found
HYBRIDGRID_ERR_002: Source file too large
HYBRIDGRID_ERR_003: Unsupported target
HYBRIDGRID_ERR_004: Docker not available
HYBRIDGRID_ERR_005: Compilation timeout
```
