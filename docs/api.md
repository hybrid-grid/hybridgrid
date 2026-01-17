# API Documentation

## gRPC API

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

## HTTP API

### Dashboard

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Web dashboard UI |
| `/health` | GET | Health check (returns "OK") |
| `/metrics` | GET | Prometheus metrics |
| `/api/stats` | GET | JSON stats for dashboard |
| `/ws` | WebSocket | Real-time updates |

### GET /api/stats

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

### WebSocket /ws

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
  "stats": { /* same as /api/stats */ },
  "timestamp": "2026-01-17T10:30:05Z"
}
```

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
