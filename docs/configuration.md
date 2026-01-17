# Configuration Reference

Hybrid-Grid uses YAML configuration files and environment variables.

## Config File Locations

1. `./hybridgrid.yaml` (current directory)
2. `~/.hybridgrid/config.yaml`
3. `/etc/hybridgrid/config.yaml`

## Environment Variables

All config options can be set via environment variables with `HG_` prefix:
- `HG_GRPC_PORT=9000`
- `HG_HTTP_PORT=8080`
- `HG_LOG_LEVEL=debug`

## Full Configuration

```yaml
# =============================================================================
# Coordinator Settings (hg-coord)
# =============================================================================
coordinator:
  # gRPC server port for worker connections
  grpc_port: 9000

  # HTTP server port for dashboard and metrics
  http_port: 8080

  # Time before marking inactive workers as dead
  heartbeat_ttl: 30s

  # Maximum concurrent tasks to accept
  max_concurrent_tasks: 1000

# =============================================================================
# Worker Settings (hg-worker)
# =============================================================================
worker:
  # gRPC server port
  grpc_port: 50052

  # HTTP server port for metrics
  http_port: 9090

  # Coordinator address to connect to
  coordinator: "localhost:9000"

  # Maximum parallel compilation tasks
  max_concurrent_tasks: 4

  # Compilation timeout
  task_timeout: 5m

  # Worker capabilities (auto-detected if not set)
  capabilities:
    architectures: ["amd64", "arm64"]
    compilers: ["gcc", "clang"]
    docker_available: true

# =============================================================================
# Cache Settings
# =============================================================================
cache:
  # Enable/disable caching
  enabled: true

  # Cache directory
  dir: ~/.hybridgrid/cache

  # Maximum cache size in GB
  max_size_gb: 10

  # Cache eviction policy: "lru" or "fifo"
  eviction_policy: lru

# =============================================================================
# Discovery Settings
# =============================================================================
discovery:
  # mDNS discovery (LAN)
  mdns:
    enabled: true
    service_name: "_hybridgrid._tcp"
    domain: "local."

  # WAN registry (future)
  wan:
    enabled: false
    registry_url: ""

# =============================================================================
# Scheduler Settings
# =============================================================================
scheduler:
  # Algorithm: "p2c" (Power of Two Choices) or "least-loaded"
  algorithm: p2c

  # P2C scoring weights
  weights:
    native_arch_match: 50
    cross_compile_capable: 25
    per_cpu_core: 10
    per_gb_ram: 5
    per_active_task: -15
    per_ms_latency: -0.5
    lan_source: 20

# =============================================================================
# Resilience Settings
# =============================================================================
resilience:
  # Circuit breaker
  circuit_breaker:
    # Requests before evaluating circuit state
    request_threshold: 10
    # Failure rate to open circuit (0.0-1.0)
    failure_rate_threshold: 0.6
    # Time circuit stays open before half-open
    open_timeout: 30s

  # Retry policy
  retry:
    max_retries: 3
    initial_interval: 100ms
    max_interval: 5s
    multiplier: 2.0

  # Local fallback
  fallback:
    enabled: true
    compilers: ["gcc", "clang"]

# =============================================================================
# TLS/Security Settings
# =============================================================================
tls:
  # Enable TLS for gRPC connections
  enabled: false

  # Server certificate
  cert_file: /etc/hybridgrid/server.crt

  # Server private key
  key_file: /etc/hybridgrid/server.key

  # CA certificate for mTLS
  ca_file: /etc/hybridgrid/ca.crt

  # Require client certificates (mTLS)
  require_client_cert: false

# =============================================================================
# Authentication Settings
# =============================================================================
auth:
  # Enable token-based authentication
  enabled: false

  # Shared secret token
  token: ""

  # Token file (alternative to inline token)
  token_file: /etc/hybridgrid/token

# =============================================================================
# Logging Settings
# =============================================================================
logging:
  # Log level: debug, info, warn, error
  level: info

  # Output format: json, console
  format: console

  # Log file (empty for stdout)
  file: ""

# =============================================================================
# Metrics Settings
# =============================================================================
metrics:
  # Enable Prometheus metrics
  enabled: true

  # Metrics endpoint path
  path: /metrics
```

## CLI Flags

All configuration can be overridden via command-line flags:

```bash
# Coordinator
hg-coord serve \
  --grpc-port=9000 \
  --http-port=8080 \
  --heartbeat-ttl=30s \
  --log-level=debug

# Worker
hg-worker serve \
  --grpc-port=50052 \
  --http-port=9090 \
  --coordinator=localhost:9000 \
  --max-tasks=4

# CLI
hgbuild build \
  --coordinator=localhost:9000 \
  --timeout=5m \
  -- gcc -c main.c -o main.o
```

## Precedence

Configuration is loaded in this order (later overrides earlier):
1. Default values
2. Config file
3. Environment variables
4. CLI flags
