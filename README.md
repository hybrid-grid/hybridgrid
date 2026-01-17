# Hybrid-Grid Build

A distributed multi-platform build system for C/C++, Flutter, Unity, and more.

Hybrid-Grid distributes compilation tasks across multiple machines on your LAN (via mDNS auto-discovery) or WAN, dramatically reducing build times for large projects.

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

### Using Docker Compose (Recommended)

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

# Submit a build
hgbuild build -c gcc main.c -o main
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
