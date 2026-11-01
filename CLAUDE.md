# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Hybrid-Grid is a Go 1.25 distributed build system that spreads C/C++ (and Flutter Android) compilation
across worker machines discovered via mDNS on a LAN, using gRPC/protobuf for coordination. Four binaries:

- `hgbuild` — CLI client; wraps `make`/`ninja` or acts as a drop-in `cc`/`c++` replacement
- `hg-coord` — coordinator; schedules tasks to workers, exposes dashboard/metrics/health
- `hg-worker` — worker node; executes compiles natively or via Docker (dockcross/Flutter images)
- `hg-dashboard` — standalone web dashboard binary

## Build / Run / Test Commands

```bash
make build                   # builds the dashboard UI, then all four binaries into bin/
make build-ui                 # npm install && npm run build in internal/observability/ui/web
                               # (required before building/running hg-dashboard from source —
                               # its go:embed directive reads web/dist, which is gitignored)
make run-coord                # go run ./cmd/hg-coord serve
make run-worker                # go run ./cmd/hg-worker serve
make run-dashboard             # build-ui, then go run ./cmd/hg-dashboard serve

make test                     # go test -v -race ./...
go test -v -race ./internal/cache/...                                 # single package
go test -v -race -run TestP2CScheduler_PrefersBetterWorker ./internal/coordinator/scheduler/...  # single test

make test-coverage             # go test -coverprofile=coverage.out ./... ; go tool cover -html
make test-integration          # INTEGRATION_TEST=1 go test -v ./test/integration/...

make lint                      # golangci-lint run
gofmt -l .                     # format check (CI enforces this)

gosec -exclude=G104,G109,G112,G115,G204,G301,G304,G306,G402 ./...
govulncheck ./...

make proto-gen                 # regenerate gen/go/hybridgrid/v1 from proto/*.proto (do not edit generated code)
make clean
```

Other test suites live under `test/` (`chaos`, `distributed`, `e2e`, `load`, `stress`) — run individually
with `go test` against their package path; most require services to already be running (coordinator/workers)
or the `INTEGRATION_TEST=1` env var, similar to `test/integration`.

## Architecture

```
hgbuild (CLI) --gRPC--> hg-coord (Coordinator) --gRPC--> hg-worker (Node 1..N)
```

Compilation flow (`hgbuild make`/`cc`/`c++`):
1. Parse compiler args (`internal/compiler`)
2. Check local content-addressable cache (`internal/cache`, xxhash-keyed) — hit returns immediately
3. Preprocess locally (`gcc -E`)
4. Send task to coordinator, which schedules it to a worker via P2C (Power of Two Choices) scoring
   (`internal/coordinator/scheduler`) based on registered capabilities (`internal/capability`)
5. Worker compiles natively or in Docker (`internal/worker/executor`) and returns the artifact
6. On coordinator/network failure, `hgbuild` falls back to local compilation automatically unless
   `--no-fallback` is passed
7. Result is stored in the local cache

### Package map (`internal/`)

| Package | Responsibility |
|---|---|
| `cache` | Content-addressable compile cache (xxhash) |
| `capability` | Hardware/software/compiler capability detection on workers |
| `cli` | `hgbuild` subcommands (build, fallback, output formatting) |
| `compiler` | Compiler argument parsing/preprocessing, MSVC flag translation |
| `config` | Viper-based configuration (`XxxConfig` structs) |
| `coordinator/registry` | Worker registration/heartbeats |
| `coordinator/scheduler` | P2C scheduling, capability matching |
| `coordinator/resilience` | Circuit breaker (`gobreaker`) / retry (`backoff/v4`) integration |
| `coordinator/metrics` | Prometheus instrumentation |
| `coordinator/opshttp` | Coordinator HTTP endpoints (dashboard, `/health`, `/api/v1/workers`) |
| `coordinator/server` | gRPC server wiring |
| `discovery` | mDNS/Zeroconf service discovery |
| `graph` | Build dependency graph parsing/visualization (`hgbuild graph`) |
| `grpc` | gRPC client/server wrapper helpers |
| `logging` | zerolog setup |
| `observability` | OpenTelemetry tracing + Prometheus metrics wiring |
| `platform` | Platform-specific code |
| `security` | Token auth, TLS/mTLS, input validation |
| `telemetry` | Telemetry data types/collection |
| `worker/executor` | Native, Docker, and MSVC compile execution |
| `worker/server` | Worker gRPC/HTTP server wiring |

Protobuf sources are in `proto/hybridgrid/v1/*.proto` (`build.proto`, `telemetry.proto`); generated Go code
lives in `gen/go/hybridgrid/v1` — never edit generated files, run `make proto-gen` instead.

### Dashboard frontend (`internal/observability/ui/web`)

`hg-dashboard`'s browser UI is a React + TypeScript SPA (Vite, TanStack Router, TanStack Query, Zustand,
Tailwind v4, shadcn-style primitives) living entirely under `internal/observability/ui/web/`. It talks only
to the REST/WebSocket API in `internal/observability/ui/{api,websocket}.go` (`/api/v1/stats|workers|tasks|
builds|builds/{id}|tasks/{id}/console`, `/ws`) — never to the coordinator directly. `server.go` embeds the
built `web/dist` via `go:embed` and serves it with SPA fallback (any unknown path serves `index.html` since
routing is client-side). Run `npm run build` (or `make build-ui`) before building/running `hg-dashboard` —
`web/dist` is gitignored generated output, not committed source, and Go's `go:embed` fails at compile time
if it's missing.

## Code Style

Full conventions (import grouping, naming, error handling, testing, concurrency, logging) are documented in
`AGENTS.md` at the repo root — read it before making non-trivial changes. Highlights:

- Errors: wrap with `fmt.Errorf("failed to <action>: %w", err)`; sentinel errors as package-level `ErrXxx`
  vars; gRPC handlers return `status.Error`/`status.Errorf` (uses `%v`, not `%w`); no custom error types.
- Tests: stdlib `testing` only (no testify in unit tests — integration tests do use `testify`), table-driven
  with `t.Run`, white-box (`package foo`, not `foo_test`), manual mock structs, `t.TempDir()` for temp dirs.
- Config: Viper `XxxConfig` structs with `mapstructure` tags, composed into a top-level `Config`.
- Concurrency: `sync.Mutex`/`sync/atomic` for shared state, `context.Context` as first param, circuit
  breaker via `gobreaker`, retries via `backoff/v4`.
- Logging: `zerolog` fluent API (`log.Info().Str("key", val).Msg("message")`).
- golangci-lint enables `errcheck`, `govet`, `staticcheck`, `unused`, `gosimple`, `ineffassign`; `typecheck`
  disabled; test files excluded from `errcheck`/`staticcheck` (see `.golangci.yml`).