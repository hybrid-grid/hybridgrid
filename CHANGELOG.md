# Changelog

All notable changes to Hybrid-Grid Build are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Build Sessions**: `HG_BUILD_ID` groups tasks into builds end-to-end (proto `build_id`, client env), giving the dashboard per-build history and detail pages
- **Dashboard Redesign** (Jenkins-inspired, fully offline — vendored Alpine.js and Chart.js): token-based theming, dark mode via `prefers-color-scheme`, build detail pages with per-task queue/compile timings, per-task console output view (bounded retention, `GET /api/v1/tasks/{id}/console`), duration and cache-hit charts, **Cluster Activity swimlanes** (per-worker lanes, occupied-slot sub-rows, queue/compile segments, click-through to builds), and a **Worker Throughput** chart (finished tasks per bucket per worker)
- **REST API**: `GET /api/v1/builds/{id}`, `GET /api/v1/tasks?limit=`, task console endpoint; timestamps as explicit `*_ms` fields (sub-second builds distinguishable)
- **Dispatch Queue Backpressure**: over-capacity compiles queue at the coordinator (bounded at 256 waiters, `--dispatch-queue-timeout`, default 30s, 0 disables) instead of failing the build with "no worker available"; the wait surfaces as real `queue_ms` in the API, dashboard, and task log
- **`/log-level` Authentication**: the runtime log-level endpoint on coordinator and worker requires the configured auth token (open when no token is set)

### Changed
- golangci-lint pinned to v2.13.2 in CI so lint findings are reproducible across upstream releases
- OpenTelemetry bumped to 1.41

### Fixed
- Build history marks a build truncated when the task ring evicts its oldest task (`buildTruncated`)
- Dashboard treats a missing `exit_code` as success (zero is omitted by JSON serialization)
- Coordinator task events report the coordinator-side queue wait (previously the worker's always-zero field)

## [v0.4.0] - 2026-03-20

Flutter Android distributed builds, on top of the v0.3.0 observability and transport work.

### Added
- **Flutter Builds**: distributed Flutter Android compilation via `hgbuild flutter build apk` and `hgbuild flutter build appbundle`
- **Docker-Based Flutter Workers**: pre-built `hybridgrid/flutter-android` image bundling the Android SDK and Gradle
- **Flutter Build Cache**: first builds always run on the worker (no cache available); identical rebuilds return cached artifacts

## [v0.3.0] - 2026-03-18

Observability completion and transport hardening: every planned metric instrumented, tracing and TLS surfaced as CLI flags, workers become first-class observables.

### Added
- **Complete Prometheus Metrics (12/12)**: instrumented `fallbacks_total`, `active_tasks`, `network_transfer_bytes`, `worker_latency_ms`, and `circuit_state` on coordinator and workers
- **OpenTelemetry CLI Flags**: `--tracing-enable`, `--tracing-endpoint`, `--tracing-service-name` on coordinator and worker
- **TLS/mTLS CLI Flags**: `--tls-cert`, `--tls-key`, `--tls-ca`, `--tls-require-client-cert` on coordinator and worker
- **`--no-fallback` Flag**: fail fast when the coordinator is unavailable instead of silently compiling locally
- **Worker Capabilities API**: `/api/v1/workers` returns compilers, architectures, build types, Docker availability, and version
- **Worker Metrics Server**: workers expose `/metrics` and `/health` at `:9090`
- **Health Endpoints**: both binaries expose `/health` for Docker/K8s healthchecks

### Fixed
- **Stress Test Exit Codes**: exit codes now correctly propagate through the `make`/`ninja` wrappers

## [v0.2.3] - 2026-03-15

Foundation hardening release focused on startup wiring, request tracing, logging, and test coverage.

### Added
- **Config Validation**: `Config.Validate()` and nested validation paths for coordinator, worker, cache, TLS, tracing, and logging settings
- **Request ID Tracing**: gRPC interceptors and server wiring for request ID propagation across coordinator and worker flows
- **Runtime Log-Level Handler**: reusable HTTP handler for `/log-level` runtime log changes
- **File-Based Logging**: `LogConfig.File` support with zerolog file output and explicit closer handling
- **Log Rotation**: Lumberjack-backed rotation with configurable size, backups, age, and compression
- **TracingConfig Expansion**: parity with observability tracing config, including service name, headers, timeout, and batch size

### Changed
- **OpenTelemetry Wiring**: `hg-coord`, `hg-worker`, and `hgbuild` now expose CLI tracing flags and initialize tracing at startup when enabled
- **TLS/mTLS Wiring**: coordinator, worker, and CLI now expose TLS flags; worker client startup no longer hardcodes insecure mode
- **CLI Verification**: `internal/cli/build` and `internal/cli/output` now have substantially broader automated coverage
- **Logging**: console/file output can be combined and rotated without changing existing default behavior

### Fixed
- Executor package coverage raised to 67.2%
- CLI build package coverage raised to 71.7%
- CLI output package coverage raised to 96.7%
- Capability package coverage improved to 25.5% on macOS, with platform-specific limits documented

## [v0.2.2] - 2025-02-20

Bug fixes, security hardening, and CI/dependency updates.

### Added
- OpenTelemetry tracing implementation (library, gRPC interceptors, per-RPC spans)
- TLS/mTLS support (cert loading, mTLS, token authentication)

### Fixed
- **15 bug fixes**: hardened worker execution, coordinator concurrency, cache TTL, mDNS discovery, dashboard WebSocket, graph renderer, compiler wrapper
- Security: cleared all `gosec` findings, fixed graph XSS vulnerability
- Portable root file operations for CI compatibility
- Safe graph embedding to prevent XSS attacks

### Changed
- **CI/Dependencies**: upgraded Go to 1.25.8, OpenTelemetry to v1.40.0 (resolved upstream CVEs)
- Enhanced coordinator concurrency handling

### Tested
- **macOS** (ARM64/x86_64) → Coordinator + Worker ✅
- **Linux** (ARM64/x86_64) → Coordinator + Worker ✅
- **Windows** (x86_64/ARM64) → Worker ✅
- **Raspberry Pi** (ARM64) → Worker ✅
- **100-file stress test** → 17s distributed, 3s cached ✅

## [v0.2.1] - 2025-02-15

Graph rendering improvements and XSS protection.

### Fixed
- Restore graph scanner error handling
- Add XSS protection to graph rendering

## [v0.2.0] - 2025-02-10

Production-ready Windows support and foundation stabilization.

### Added
- **Windows Support**: Full Windows compatibility for coordinator and worker (x86_64/ARM64)
- **Test Coverage**: Comprehensive Windows-specific test suite
- **CI Integration**: Windows testing in GitHub Actions pipeline

### Fixed
- Windows path validation (drive letter handling)
- Lint issues across codebase
- Formatting alignment with gofmt

### Changed
- Upgraded Go to 1.24 for dependency compatibility
- Enhanced CI pipeline with multiplatform testing

### Features (Existing)
- **C/C++ Distributed Compilation** - Cross-platform, cross-compilation capable
- **MSVC Flag Translation** - Automatic GCC/Clang to MSVC flag conversion
- **mDNS Auto-Discovery** - Zero-config LAN service discovery
- **Local Cache** - ~10x speedup on cache hits via content-addressable caching
- **Web Dashboard** - Real-time worker/task/cache statistics
- **Smart Scheduling** - P2C (Power of Two Choices) with capability matching
- **Circuit Breaker** - Per-worker fault tolerance
- **Docker Cross-Compile** - dockcross integration for heterogeneous builds
- **Colored CLI Output** - Visual build progress and status indicators
- **Prometheus Metrics** - Comprehensive observability

[Unreleased]: https://github.com/h3nr1-d14z/hybridgrid/compare/v0.4.0...HEAD
[v0.4.0]: https://github.com/h3nr1-d14z/hybridgrid/compare/v0.3.0...v0.4.0
[v0.3.0]: https://github.com/h3nr1-d14z/hybridgrid/compare/v0.2.3...v0.3.0
[v0.2.3]: https://github.com/h3nr1-d14z/hybridgrid/compare/v0.2.2...v0.2.3
[v0.2.2]: https://github.com/h3nr1-d14z/hybridgrid/compare/v0.2.1...v0.2.2
[v0.2.1]: https://github.com/h3nr1-d14z/hybridgrid/compare/v0.2.0...v0.2.1
[v0.2.0]: https://github.com/h3nr1-d14z/hybridgrid/releases/tag/v0.2.0
