# AGENTS.md - Hybrid-Grid Build System

## Project Overview

Go 1.24 distributed build system using gRPC, protobuf, zerolog, Viper, and cobra.
Three binaries: `hgbuild` (CLI client), `hg-coord` (coordinator), `hg-worker` (worker node).

## Build / Run / Test Commands

```bash
# Build all binaries
make build

# Run coordinator / worker
make run-coord
make run-worker

# Run all tests with race detector
make test
# equivalent to: go test -v -race ./...

# Run tests for a single package
go test -v -race ./internal/cache/...
go test -v -race ./internal/coordinator/scheduler/...

# Run a single test by name
go test -v -race -run TestP2CScheduler_PrefersBetterWorker ./internal/coordinator/scheduler/...

# Run tests with coverage
make test-coverage
# generates coverage.out and coverage.html

# Run integration tests (requires INTEGRATION_TEST=1)
make test-integration
# equivalent to: INTEGRATION_TEST=1 go test -v ./test/integration/...

# Lint
make lint
# equivalent to: golangci-lint run

# Format check (CI enforces this)
gofmt -l .

# Security scan
gosec -exclude=G104,G109,G112,G115,G204,G301,G304,G306,G402 ./...
govulncheck ./...

# Regenerate protobuf code
make proto-gen

# Clean build artifacts
make clean
```

## Project Structure

```
cmd/                    # CLI entry points (hg-coord, hg-worker, hgbuild)
internal/               # Core application packages
  cache/                # Content-addressable compilation cache (xxhash)
  capability/           # Hardware/software capability detection
  cli/                  # CLI subcommands (build, fallback, output)
  compiler/             # Compiler argument parsing, preprocessing, MSVC flags
  config/               # Viper-based configuration
  coordinator/          # Coordinator: registry, scheduler, resilience, metrics
  discovery/            # mDNS/Zeroconf service discovery
  graph/                # Build dependency graph parsing and visualization
  grpc/                 # gRPC client/server wrappers
  observability/        # OpenTelemetry tracing, Prometheus metrics, dashboard
  platform/             # Platform-specific code
  security/             # Auth (token), TLS/mTLS, input validation
  worker/               # Worker: native, Docker, and MSVC executors
gen/go/                 # Generated protobuf Go code (do not edit)
proto/                  # Protobuf definitions (.proto files)
test/                   # Integration, load, chaos, and stress tests
configs/                # Configuration file templates
docs/                   # Documentation
scripts/                # Build/utility scripts
```

## Code Style Guidelines

### Imports

Three groups separated by blank lines, in this order:

```go
import (
    // 1. Standard library
    "context"
    "fmt"
    "time"

    // 2. External/third-party packages
    "github.com/rs/zerolog/log"
    "google.golang.org/grpc"

    // 3. Internal project packages
    pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
    "github.com/h3nr1-d14z/hybridgrid/internal/config"
)
```

Common import aliases:
- `pb` for protobuf: `pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"`
- `hgtls` for internal TLS to avoid collisions
- `otelcodes` for OpenTelemetry codes (avoids collision with gRPC `codes`)

### Formatting

- `gofmt` enforced in CI -- no deviations
- No `.editorconfig`; rely on `gofmt` and `golangci-lint`

### Naming Conventions

- **Packages**: lowercase, single word (`cache`, `scheduler`, `registry`)
- **Interfaces**: descriptive nouns (`Scheduler`, `Registry`, `CircuitChecker`)
- **Structs**: PascalCase (`SimpleScheduler`, `WorkerInfo`, `Config`)
- **Config types**: `XxxConfig` suffix (`CoordinatorConfig`, `CacheConfig`, `TLSConfig`)
- **Constructors**: `NewXxx` (`NewSimpleScheduler`, `NewStore`, `New`)
- **Sentinel errors**: `ErrXxx` (`ErrNoWorkers`, `ErrMaxRetriesExceeded`)
- **Test names**: `TestTypeName_BehaviorDescription` (`TestStore_TTLExpiration`, `TestP2CScheduler_SkipsOpenCircuit`)
- **Test helpers**: unexported functions at package level (`newTestRegistry`, `addCppWorker`)

### Types and Structs

- Use `mapstructure` tags for Viper config structs
- Configuration via nested `XxxConfig` structs composed in a top-level `Config`
- Interfaces defined where consumed, not where implemented
- Protobuf enums from `gen/go/hybridgrid/v1` used for build types and architectures

### Error Handling

1. **Wrap errors with context** using `fmt.Errorf` and `%w`:
   ```go
   return nil, fmt.Errorf("failed to connect: %w", err)
   ```
   Message pattern: `"failed to <action>: %w"` -- lowercase, descriptive action.

2. **Sentinel errors** as package-level `var` blocks:
   ```go
   var (
       ErrNoWorkers = errors.New("no workers available")
   )
   ```

3. **gRPC errors** use `status.Error` / `status.Errorf` with appropriate codes:
   ```go
   return nil, status.Error(codes.InvalidArgument, "task_id required")
   return nil, status.Errorf(codes.Internal, "failed to register: %v", err)
   ```
   Note: gRPC status uses `%v` (not `%w`) since it doesn't support wrapping.

4. **No custom error types** -- only sentinels, wrapped errors, and gRPC status errors.

5. **Check wrapped errors** with `errors.Is()`:
   ```go
   if errors.Is(err, context.Canceled) { ... }
   ```

### Testing

- **Framework**: stdlib `testing` package with manual assertions (no testify/assert in unit tests)
- **Assertions**: `t.Fatal`/`t.Fatalf` for critical failures; `t.Error`/`t.Errorf` for non-fatal
- **Table-driven tests** with `t.Run` subtests are the standard pattern
- **White-box testing**: test files use the same package (e.g., `package scheduler`, not `package scheduler_test`)
- **Mocking**: manual mock structs implementing interfaces (no mocking frameworks)
- **Temp files**: use `t.TempDir()` for directories, `os.CreateTemp` for individual files
- **Cleanup**: `defer` for resource cleanup; `t.TempDir()` auto-cleans

### Concurrency

- `sync.Mutex` / `sync.RWMutex` for shared state
- `sync/atomic` for counters and flags
- `context.Context` passed as first parameter for cancellation/timeout
- Circuit breaker pattern via `github.com/sony/gobreaker`
- Retry with exponential backoff via `github.com/cenkalti/backoff/v4`

### Logging

- `github.com/rs/zerolog` -- structured logging throughout
- Logger configured at startup: `zerolog.ConsoleWriter` for dev, JSON for production
- Use `log.Info().Str("key", val).Msg("message")` style (fluent API)

### Linting

golangci-lint with these linters enabled: `errcheck`, `govet`, `staticcheck`, `unused`, `gosimple`, `ineffassign`. `typecheck` is disabled. SA1019 (deprecation) is suppressed. Test files are excluded from `errcheck` and `staticcheck`. See `.golangci.yml` for full exclusion rules.

### Key Dependencies

| Package | Purpose |
|---------|---------|
| `google.golang.org/grpc` | gRPC transport |
| `google.golang.org/protobuf` | Protocol buffers |
| `github.com/spf13/cobra` | CLI framework |
| `github.com/spf13/viper` | Configuration |
| `github.com/rs/zerolog` | Structured logging |
| `github.com/sony/gobreaker` | Circuit breaker |
| `github.com/cenkalti/backoff/v4` | Exponential backoff |
| `github.com/cespare/xxhash/v2` | Content hashing for cache |
| `github.com/docker/docker` | Docker integration |
| `github.com/grandcat/zeroconf` | mDNS discovery |
| `github.com/prometheus/client_golang` | Prometheus metrics |
| `github.com/stretchr/testify` | Test assertions (integration tests) |
