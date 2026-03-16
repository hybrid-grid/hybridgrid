# v0.2.3 Foundation Hardening

## TL;DR

> **Quick Summary**: Solidify the Hybrid-Grid codebase by wiring up config validation, OTel tracing, TLS startup, log management, request ID tracing, and boosting test coverage — closing all "quick win" gaps before starting v0.3.0 feature work (Flutter/Unity/WAN).
>
> **Deliverables**:
> - Config validation with `Validate()` method and expanded `TracingConfig` (8 fields)
> - OTel tracing wired into all 3 binaries with CLI flags
> - TLS/mTLS wired into all 3 binaries with CLI flags, `Insecure: true` removed
> - gRPC request ID interceptor for correlation tracing
> - Runtime log-level endpoint on coordinator and worker
> - File-based logging with lumberjack rotation
> - Test coverage boost: executor→60%, capability→60%, cli/build→60%, cli/output→70%
> - CHANGELOG.md generation + Makefile target
> - CHECKLIST.md updated to reflect all completed items
>
> **Estimated Effort**: Medium (2-3 focused sessions)
> **Parallel Execution**: YES — 4 waves
> **Critical Path**: Task 1 → Task 3 → Task 4 → Task 5 → Task 13 (CHANGELOG) → Task 14 (CHECKLIST)

---

## Context

### Original Request
User chose "Foundation first, then v0.3.0" — close all quick-win gaps from CHECKLIST.md, release v0.2.3, then begin Flutter builds (v0.3.0 Phase 1).

### Interview Summary
**Key Discussions**:
- Config file loading (`--config`): CLI flags only for v0.2.3 — no config file refactoring
- Auth interceptor wiring: Explicitly deferred to v0.3.0 with TODO comment
- Log-level endpoint: Both coordinator and worker, extracted to `internal/logging` package
- Log rotation: Lumberjack dependency (cross-platform, in-process rotation)
- Test coverage targets: Per-package (executor→60%, capability→60%, cli/build→60%, cli/output→70%)
- CHANGELOG: One-time generation + Makefile target for future

**Research Findings**:
- `config.TracingConfig` has 4 fields but `tracing.Config` has 8 — must expand before OTel wiring
- `LogConfig.File` is dead code — must be made functional before adding lumberjack
- `internal/worker/server/grpc.go:85-103` has the pattern for conditional TLS/OTel wiring
- `cmd/hg-worker/main.go:171-183` has the pattern for HTTP endpoint additions
- Auth interceptor exists at `internal/security/auth/interceptor.go` but MUST NOT be wired in v0.2.3

### Metis Review
**Identified Gaps** (addressed):
- Config struct mismatch is blocking OTel wiring → Task 1 expands TracingConfig first
- LogConfig.File is dead code → Task 8 makes it functional before Task 9 adds lumberjack
- Auth interceptors must be explicitly excluded → Added to "Must NOT Have" guardrails
- Config file loading must not be refactored → Scope locked to CLI flags only
- Platform-specific test constraints → Test tasks exclude MSVC/Windows-only code paths

---

## Work Objectives

### Core Objective
Wire up all existing-but-unused infrastructure (OTel, TLS, logging, config validation) and boost test coverage to create a solid foundation for v0.3.0 feature development.

### Concrete Deliverables
- `internal/config/config.go` — Expanded `TracingConfig`, new `LogRotationConfig`, `Validate()` method
- `cmd/hg-coord/main.go` — OTel + TLS CLI flags and startup wiring
- `cmd/hg-worker/main.go` — OTel + TLS CLI flags, `Insecure: true` removed, log-level endpoint
- `cmd/hgbuild/main.go` — OTel + TLS CLI flags and startup wiring
- `internal/logging/` — New package: log-level HTTP handler, file writer setup
- `internal/grpc/interceptors/requestid.go` — Request ID interceptor
- Test files for executor, capability, cli/build, cli/output
- `CHANGELOG.md` — Release history from git tags
- `CHECKLIST.md` — Updated status

### Definition of Done
- [ ] `make test` passes with 0 failures
- [ ] `make lint` passes with 0 errors
- [ ] `gosec -exclude=G104,G109,G112,G115,G204,G301,G304,G306,G402 ./...` clean
- [ ] All 3 binaries accept `--tracing-enable`, `--tracing-endpoint`, `--tls-cert`, `--tls-key`, `--tls-ca` flags
- [ ] `hg-coord` and `hg-worker` expose `/log-level` HTTP endpoint
- [ ] CHANGELOG.md exists with v0.2.0–v0.2.3 entries
- [ ] Per-package coverage: executor≥60%, capability≥60%, cli/build≥60%, cli/output≥70%

### Must Have
- Config validation that catches invalid port ranges, empty required fields, conflicting options
- OTel tracing conditional on `--tracing-enable` flag (disabled by default)
- TLS conditional on `--tls-cert` + `--tls-key` flags (insecure by default, but NO hardcoded `Insecure: true`)
- Request IDs propagated via gRPC metadata
- Lumberjack log rotation with configurable max size, max backups, max age, compression
- Runtime log-level changes via HTTP PUT/POST to `/log-level`

### Must NOT Have (Guardrails)
- ❌ Auth interceptor wiring — explicitly deferred to v0.3.0 (TODO comment only)
- ❌ Config file loading refactoring — CLI flags only, no `--config` flag
- ❌ Flutter/Unity/WAN features — that's v0.3.0
- ❌ Build()/StreamBuild() implementation — stubs stay as-is
- ❌ MSVC/Windows-only test code on macOS — skip platform-specific paths
- ❌ Over-abstracted utility packages — keep changes focused and inline
- ❌ Excessive JSDoc/comments beyond what's needed for API clarity
- ❌ Changing existing working behavior — all changes are additive or fix dead code

---

## Verification Strategy (MANDATORY)

> **ZERO HUMAN INTERVENTION** — ALL verification is agent-executed. No exceptions.

### Test Decision
- **Infrastructure exists**: YES (Go stdlib `testing`, `make test`)
- **Automated tests**: Tests-after (write implementation, then add tests)
- **Framework**: Go stdlib `testing` with manual assertions (project convention)
- **Coverage tool**: `go test -coverprofile`

### QA Policy
Every task MUST include agent-executed QA scenarios.
Evidence saved to `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`.

- **CLI binaries**: Use Bash — Run binary with `--help`, verify flags listed, run with flags and check stderr/stdout
- **gRPC endpoints**: Use Bash — Start server, `grpcurl` or test client to exercise endpoints
- **HTTP endpoints**: Use Bash (curl) — Hit `/log-level`, assert response
- **Library/Module**: Use Bash — `go test -v -run TestSpecificName ./package/...`
- **Config validation**: Use Bash — `go test -v -run TestValidate ./internal/config/...`

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Start Immediately — foundation types + independent modules):
├── Task 1: Expand TracingConfig + LogRotationConfig in config.go [quick]
├── Task 2: Add Config.Validate() method [quick]
├── Task 6: Request ID interceptor [quick]
├── Task 7: Log-level HTTP handler package [quick]
├── Task 8: Make LogConfig.File functional (file writer) [quick]
└── Task 10: Test coverage — capability package [unspecified-high]

Wave 2 (After Wave 1 — binary wiring, depends on config + modules):
├── Task 3: OTel CLI flags + startup wiring (all 3 binaries) [unspecified-high]
├── Task 4: TLS CLI flags + startup wiring (all 3 binaries) [unspecified-high]
├── Task 5: Wire request ID interceptor into servers [quick]
├── Task 9: Add lumberjack rotation (depends on Task 8) [quick]
└── Task 11: Test coverage — executor package [unspecified-high]

Wave 3 (After Wave 2 — remaining tests + integration):
├── Task 12: Test coverage — cli/build + cli/output [unspecified-high]
└── Task 13: CHANGELOG.md generation [quick]

Wave 4 (After Wave 3 — documentation + final):
└── Task 14: Update CHECKLIST.md [quick]

Wave FINAL (After ALL tasks — independent review, 4 parallel):
├── Task F1: Plan compliance audit (oracle)
├── Task F2: Code quality review (unspecified-high)
├── Task F3: Real QA — build & run all 3 binaries with new flags (unspecified-high)
└── Task F4: Scope fidelity check (deep)

Critical Path: Task 1 → Task 3 → Task 4 → Task 13 → Task 14 → F1-F4
Parallel Speedup: ~60% faster than sequential
Max Concurrent: 6 (Wave 1)
```

### Dependency Matrix

| Task | Depends On | Blocks | Wave |
|------|-----------|--------|------|
| 1 | — | 2, 3, 4, 9 | 1 |
| 2 | 1 | 3, 4 | 1 (can start after T1 config shape) |
| 6 | — | 5 | 1 |
| 7 | — | 3, 4 | 1 |
| 8 | — | 9 | 1 |
| 10 | — | — | 1 |
| 3 | 1, 2, 7 | 12, 13 | 2 |
| 4 | 1, 2 | 12, 13 | 2 |
| 5 | 6 | — | 2 |
| 9 | 8, 1 | — | 2 |
| 11 | — | — | 2 |
| 12 | 3, 4 | 13 | 3 |
| 13 | all code tasks | 14 | 3 |
| 14 | 13 | F1-F4 | 4 |
| F1-F4 | 14 | — | FINAL |

### Agent Dispatch Summary

- **Wave 1**: **6 tasks** — T1→`quick`, T2→`quick`, T6→`quick`, T7→`quick`, T8→`quick`, T10→`unspecified-high`
- **Wave 2**: **5 tasks** — T3→`unspecified-high`, T4→`unspecified-high`, T5→`quick`, T9→`quick`, T11→`unspecified-high`
- **Wave 3**: **2 tasks** — T12→`unspecified-high`, T13→`quick`
- **Wave 4**: **1 task** — T14→`quick`
- **FINAL**: **4 tasks** — F1→`oracle`, F2→`unspecified-high`, F3→`unspecified-high`, F4→`deep`

---

## TODOs

- [x] 1. Expand TracingConfig and add LogRotationConfig to config.go

  **What to do**:
  - Add 4 missing fields to `TracingConfig`: `ServiceName string`, `Headers map[string]string`, `Timeout time.Duration`, `BatchSize int` — matching `tracing.Config` exactly
  - Add `LogRotationConfig` struct: `MaxSizeMB int`, `MaxBackups int`, `MaxAgeDays int`, `Compress bool` (all with `mapstructure` tags)
  - Add `Rotation LogRotationConfig` field to existing `LogConfig` struct
  - Update `DefaultConfig()` to include tracing defaults matching `tracing.DefaultConfig()`: ServiceName="hybridgrid", SampleRate=0.1, Insecure=true, Timeout=10s, BatchSize=512
  - Update `DefaultConfig()` to include log rotation defaults: MaxSizeMB=100, MaxBackups=3, MaxAgeDays=28, Compress=true
  - Update `setDefaults()` to register new viper defaults
  - Update `WriteExample()` example YAML to include the new fields as comments
  - Add a helper function `TracingToLibConfig(tc TracingConfig) tracing.Config` that converts `config.TracingConfig` → `tracing.Config` (mapping layer so callers can pass config values to `tracing.Init()`)

  **Must NOT do**:
  - Do NOT change the existing 4 fields of TracingConfig (Enable, Endpoint, SampleRate, Insecure) — only ADD new ones
  - Do NOT modify `tracing.Config` or `tracing.DefaultConfig()` — those are the source of truth
  - Do NOT add config file loading logic

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Single-file struct expansion, straightforward field additions
  - **Skills**: `[]`
  - **Skills Evaluated but Omitted**:
    - None needed — pure Go struct changes

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 2, 6, 7, 8, 10)
  - **Blocks**: Tasks 2, 3, 4, 9
  - **Blocked By**: None (can start immediately)

  **References**:

  **Pattern References** (existing code to follow):
  - `internal/config/config.go:48-53` — Current `TracingConfig` struct (4 fields to preserve, 4 to add)
  - `internal/config/config.go:92-97` — Current `LogConfig` struct (add `Rotation` field)
  - `internal/config/config.go:172-193` — `setDefaults()` function (add new defaults here)
  - `internal/config/config.go:196-250` — `WriteExample()` YAML template (add new fields as comments)

  **API/Type References** (contracts to implement against):
  - `internal/observability/tracing/config.go:8-32` — `tracing.Config` struct with all 8 fields — this is the TARGET shape
  - `internal/observability/tracing/config.go:35-45` — `tracing.DefaultConfig()` — copy these default values
  - `internal/security/tls/config.go:9-30` — `tls.Config` struct — reference for how `config.TLSConfig` already mirrors this pattern

  **WHY Each Reference Matters**:
  - `config.go:48-53` — Must preserve existing field names and tags exactly; add new fields in same style
  - `tracing/config.go:8-32` — The 8-field struct that `TracingConfig` must expand to match (ServiceName, Headers, Timeout, BatchSize are missing)
  - `tracing/config.go:35-45` — Default values to copy: ensures `config.DefaultConfig()` matches `tracing.DefaultConfig()`
  - `tls/config.go:9-30` — Shows how `config.TLSConfig` already mirrors `tls.Config` — follow same pattern for tracing

  **Acceptance Criteria**:
  - [ ] `config.TracingConfig` has exactly 8 fields matching `tracing.Config` field names
  - [ ] `config.LogConfig` has a `Rotation LogRotationConfig` field
  - [ ] `LogRotationConfig` has MaxSizeMB, MaxBackups, MaxAgeDays, Compress fields
  - [ ] `DefaultConfig()` sets all tracing defaults matching `tracing.DefaultConfig()`
  - [ ] `DefaultConfig()` sets rotation defaults: MaxSizeMB=100, MaxBackups=3, MaxAgeDays=28, Compress=true
  - [ ] `setDefaults()` registers all new fields
  - [ ] `WriteExample()` includes new tracing and rotation fields as comments
  - [ ] `TracingToLibConfig()` function exists and correctly maps all 8 fields
  - [ ] `go build ./...` succeeds

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: TracingConfig field parity with tracing.Config
    Tool: Bash
    Preconditions: Task 1 changes applied
    Steps:
      1. Run: go doc ./internal/config TracingConfig
      2. Run: go doc ./internal/observability/tracing Config
      3. Compare field names — both must have: Enable, Endpoint, ServiceName, SampleRate, Insecure, Headers, Timeout, BatchSize
    Expected Result: All 8 fields present in both structs with matching names
    Failure Indicators: Missing field name in TracingConfig output
    Evidence: .sisyphus/evidence/task-1-tracing-config-parity.txt

  Scenario: DefaultConfig returns correct tracing defaults
    Tool: Bash
    Preconditions: Task 1 changes applied
    Steps:
      1. Create a small Go test file: `go test -run TestDefaultConfigTracingDefaults ./internal/config/...`
         Test body: `cfg := DefaultConfig(); assert cfg.Tracing.ServiceName == "hybridgrid", cfg.Tracing.SampleRate == 0.1, cfg.Tracing.Insecure == true, cfg.Tracing.Timeout == 10*time.Second, cfg.Tracing.BatchSize == 512`
    Expected Result: All default values match tracing.DefaultConfig()
    Failure Indicators: Any assertion failure in test output
    Evidence: .sisyphus/evidence/task-1-default-config-test.txt

  Scenario: Build succeeds with expanded config
    Tool: Bash
    Preconditions: Task 1 changes applied
    Steps:
      1. Run: go build ./...
    Expected Result: Exit code 0, no compilation errors
    Failure Indicators: Any "cannot find", "undefined", or type mismatch errors
    Evidence: .sisyphus/evidence/task-1-build-success.txt
  ```

  **Commit**: YES
  - Message: `feat(config): expand TracingConfig and add LogRotationConfig`
  - Files: `internal/config/config.go`
  - Pre-commit: `go build ./...`

- [x] 2. Add Config.Validate() method with field constraints

  **What to do**:
  - Add a `Validate() error` method on `*Config` in `internal/config/config.go`
  - Validate coordinator: GRPCPort in 1-65535, HTTPPort in 1-65535, GRPCPort ≠ HTTPPort
  - Validate worker: Port in 1-65535, MaxParallel ≥ 0, Timeout > 0 if set, HeartbeatSec > 0 if set
  - Validate client: Timeout > 0 if set
  - Validate cache: MaxSize > 0 if enabled, TTLHours > 0 if enabled, Dir not empty if enabled
  - Validate log: Level is one of "debug", "info", "warn", "error", "fatal"; Format is one of "console", "json"
  - Validate log rotation: MaxSizeMB > 0, MaxBackups ≥ 0, MaxAgeDays > 0 (only if File is set)
  - Delegate to `TLS.Validate()` (already exists on `hgtls.Config`) — but since `config.TLSConfig` is a separate struct, add a similar `Validate()` on `config.TLSConfig` that checks: if Enabled, CertFile + KeyFile required; if RequireClientCert, ClientCA required
  - Delegate to `tracing.Config.Validate()` after converting via `TracingToLibConfig()` — OR add a `Validate()` on `config.TracingConfig` that checks: if Enable, Endpoint must not be empty, SampleRate 0-1
  - Return the FIRST error found (don't aggregate) with descriptive message: `"config: coordinator.grpc_port must be 1-65535, got %d"`
  - Write comprehensive tests in `internal/config/config_test.go` — table-driven tests covering: valid config, each invalid field, boundary values

  **Must NOT do**:
  - Do NOT use a validation library (no go-playground/validator) — manual validation matching project style
  - Do NOT aggregate errors — return first error
  - Do NOT change `Load()` to call `Validate()` automatically (callers wire this in main.go)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Single package, well-scoped validation logic with clear rules
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: YES (starts in Wave 1, but should wait for Task 1 to know final config shape)
  - **Parallel Group**: Wave 1 (can start once Task 1 config shape is committed)
  - **Blocks**: Tasks 3, 4
  - **Blocked By**: Task 1 (needs final TracingConfig and LogRotationConfig shape)

  **References**:

  **Pattern References** (existing code to follow):
  - `internal/security/tls/config.go:43-65` — `tls.Config.Validate()` — EXACT pattern to follow: return nil if disabled, check required fields, descriptive errors
  - `internal/observability/tracing/config.go:70-88` — `tracing.Config.Validate()` — similar pattern: return nil if disabled, check endpoint + sample rate

  **API/Type References** (contracts to implement against):
  - `internal/config/config.go:14-35` — Top-level `Config` struct with all nested configs
  - `internal/config/config.go:37-45` — `TLSConfig` (needs its own Validate)
  - `internal/config/config.go:48-53` — `TracingConfig` (needs its own Validate or delegate)

  **Test References** (testing patterns to follow):
  - No existing config tests — follow `internal/cache/store_test.go` or `internal/coordinator/scheduler/scheduler_test.go` for table-driven test style

  **WHY Each Reference Matters**:
  - `tls/config.go:43-65` — Gold standard for the validation pattern: "if !enabled, return nil; check required fields; return descriptive error"
  - `tracing/config.go:70-88` — Secondary example of same pattern, shows sample_rate bounds check
  - Config struct (lines 14-35) — Need to validate every nested config that has constraints

  **Acceptance Criteria**:
  - [ ] `Config.Validate()` method exists and returns error for invalid configs
  - [ ] Port range validation: 0 and 65536 both rejected
  - [ ] Log level validation: "trace" rejected, "info" accepted
  - [ ] Cache validation: MaxSize=0 with Enable=true rejected
  - [ ] TLS validation: Enabled=true without CertFile rejected
  - [ ] Tracing validation: Enable=true without Endpoint rejected
  - [ ] Valid default config passes validation: `DefaultConfig().Validate() == nil`
  - [ ] `config_test.go` has ≥15 test cases covering each validation rule
  - [ ] `go test ./internal/config/...` passes

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: DefaultConfig passes validation
    Tool: Bash
    Preconditions: Task 2 changes applied
    Steps:
      1. Run: go test -v -run TestValidate_DefaultConfig ./internal/config/...
    Expected Result: Test passes — DefaultConfig().Validate() returns nil
    Failure Indicators: Test failure or unexpected error from Validate()
    Evidence: .sisyphus/evidence/task-2-default-valid.txt

  Scenario: Invalid port rejected
    Tool: Bash
    Preconditions: Task 2 changes applied
    Steps:
      1. Run: go test -v -run TestValidate_InvalidPort ./internal/config/...
    Expected Result: Validate() returns error containing "grpc_port" and "1-65535"
    Failure Indicators: Test passes when it should fail, or error message doesn't mention field name
    Evidence: .sisyphus/evidence/task-2-invalid-port.txt

  Scenario: All validation tests pass
    Tool: Bash
    Preconditions: Task 2 changes applied
    Steps:
      1. Run: go test -v -count=1 ./internal/config/...
    Expected Result: All tests pass, ≥15 test cases run
    Failure Indicators: Any FAIL in output
    Evidence: .sisyphus/evidence/task-2-all-tests.txt
  ```

  **Commit**: YES
  - Message: `feat(config): add Validate() method with field constraints`
  - Files: `internal/config/config.go`, `internal/config/config_test.go`
  - Pre-commit: `go test ./internal/config/...`

- [x] 3. Wire OTel tracing with CLI flags in all 3 binaries

  **What to do**:
  - In `cmd/hg-coord/main.go`:
    - Add import for `tracing` package and `config` package
    - Add CLI flags to `serveCmd`: `--tracing-enable` (bool, default false), `--tracing-endpoint` (string, default "localhost:4317"), `--tracing-sample-rate` (float64, default 0.1), `--tracing-insecure` (bool, default true)
    - After config is built, construct `tracing.Config` using `config.TracingToLibConfig()` with CLI flag values overriding defaults
    - Call `tracing.Init(ctx, tracingCfg)` early in RunE, before server start
    - If tracing init succeeds, `defer tp.Shutdown(ctx)` for clean shutdown
    - Pass `tracingCfg` to `coordserver.Config{..., Tracing: tracingCfg}` so the server's conditional interceptor wiring (`if s.config.Tracing.Enable`) activates
    - Log tracing status on startup
  - In `cmd/hg-worker/main.go`:
    - Same CLI flags pattern as coordinator
    - Call `tracing.Init(ctx, tracingCfg)` early in RunE
    - `workerserver.Config` already has `Tracing tracing.Config` field — populate it from CLI flags
    - Worker server already checks `s.config.Tracing.Enable` at `grpc.go:100-103` — no changes needed in server code
  - In `cmd/hgbuild/main.go`:
    - Add `--tracing-enable` and `--tracing-endpoint` flags to root command (applies to all subcommands)
    - Call `tracing.Init()` in `PersistentPreRunE` so it initializes before any subcommand
    - Use `tracing.ClientConfig()` as base, override with CLI flags
    - Add tracing client interceptors to gRPC dial options when tracing is enabled

  **Must NOT do**:
  - Do NOT wire auth interceptor — deferred to v0.3.0
  - Do NOT add `--config` file loading — CLI flags only
  - Do NOT modify `internal/coordinator/server/grpc.go` or `internal/worker/server/grpc.go` — they already have conditional TLS/tracing wiring
  - Do NOT change tracing library code in `internal/observability/tracing/`

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Touches 3 binary files, requires careful flag registration and startup sequencing
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: NO (depends on Wave 1)
  - **Parallel Group**: Wave 2 (with Tasks 4, 5, 9, 11)
  - **Blocks**: Tasks 12, 13
  - **Blocked By**: Tasks 1, 2, 7

  **References**:

  **Pattern References** (existing code to follow):
  - `cmd/hg-coord/main.go:46-61` — Current `serveCmd.RunE` setup pattern (add tracing init here)
  - `cmd/hg-coord/main.go:130-134` — Current flag registration (add tracing flags in same style)
  - `cmd/hg-worker/main.go:54-61` — Worker flag reading pattern
  - `cmd/hg-worker/main.go:226-233` — Worker flag registration
  - `internal/coordinator/server/grpc.go:195-199` — Server-side conditional tracing wiring (already works, just needs config populated)
  - `internal/worker/server/grpc.go:99-103` — Worker-side conditional tracing wiring (same pattern)

  **API/Type References** (contracts to implement against):
  - `internal/observability/tracing/tracer.go:44` — `Init(ctx context.Context, cfg Config) (*TracerProvider, error)` — the function to call
  - `internal/observability/tracing/config.go:47-52` — `CoordinatorConfig()` — preset for coordinator
  - `internal/observability/tracing/config.go:55-59` — `WorkerConfig()` — preset for worker
  - `internal/observability/tracing/config.go:62-67` — `ClientConfig()` — preset for hgbuild
  - `internal/observability/tracing/grpc.go` — `ServerOptions()` and `ClientOptions()` — gRPC interceptors
  - `internal/coordinator/server/grpc.go:87` — `Config.Tracing tracing.Config` field
  - `internal/worker/server/grpc.go:29` — `Config.Tracing tracing.Config` field

  **WHY Each Reference Matters**:
  - `cmd/hg-coord/main.go:46-61` — Insert tracing init AFTER config setup, BEFORE `srv.Start()`
  - `tracing/tracer.go:44` — Exact function signature to call; returns TracerProvider for defer Shutdown
  - `tracing/config.go:47-67` — Pre-built configs for each binary; use as base, override with CLI flags
  - `server/grpc.go:195-199` — Proves no server-side changes needed; just populate Config.Tracing

  **Acceptance Criteria**:
  - [ ] `hg-coord serve --help` shows `--tracing-enable`, `--tracing-endpoint`, `--tracing-sample-rate`, `--tracing-insecure`
  - [ ] `hg-worker serve --help` shows same tracing flags
  - [ ] `hgbuild --help` shows `--tracing-enable`, `--tracing-endpoint`
  - [ ] `hg-coord serve --tracing-enable --tracing-endpoint=localhost:4317` logs "Tracing initialized" or "OpenTelemetry tracing enabled"
  - [ ] Without `--tracing-enable`, no tracing logs appear
  - [ ] `go build ./cmd/...` succeeds
  - [ ] `make test` passes

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: Coordinator help shows tracing flags
    Tool: Bash
    Preconditions: Task 3 changes applied, binaries built
    Steps:
      1. Run: go build -o /tmp/hg-coord ./cmd/hg-coord
      2. Run: /tmp/hg-coord serve --help
      3. Assert output contains: "--tracing-enable", "--tracing-endpoint"
    Expected Result: Both flags listed with descriptions
    Failure Indicators: Flags missing from help output
    Evidence: .sisyphus/evidence/task-3-coord-help.txt

  Scenario: Worker help shows tracing flags
    Tool: Bash
    Preconditions: Task 3 changes applied, binaries built
    Steps:
      1. Run: go build -o /tmp/hg-worker ./cmd/hg-worker
      2. Run: /tmp/hg-worker serve --help
      3. Assert output contains: "--tracing-enable", "--tracing-endpoint"
    Expected Result: Both flags listed
    Failure Indicators: Flags missing
    Evidence: .sisyphus/evidence/task-3-worker-help.txt

  Scenario: hgbuild help shows tracing flags
    Tool: Bash
    Preconditions: Task 3 changes applied, binaries built
    Steps:
      1. Run: go build -o /tmp/hgbuild ./cmd/hgbuild
      2. Run: /tmp/hgbuild --help
      3. Assert output contains: "--tracing-enable"
    Expected Result: Flag listed
    Failure Indicators: Flag missing
    Evidence: .sisyphus/evidence/task-3-hgbuild-help.txt

  Scenario: Tracing disabled by default (no flag)
    Tool: Bash
    Preconditions: Coordinator binary built
    Steps:
      1. Run: timeout 3 /tmp/hg-coord serve --grpc-port=19000 --http-port=18080 2>&1 || true
      2. Grep output for "tracing" or "Tracing"
    Expected Result: No "Tracing initialized" line appears (only if --tracing-enable is passed)
    Failure Indicators: Tracing init message without the flag
    Evidence: .sisyphus/evidence/task-3-tracing-disabled-default.txt
  ```

  **Commit**: YES
  - Message: `feat(cmd): wire OTel tracing with CLI flags in all binaries`
  - Files: `cmd/hg-coord/main.go`, `cmd/hg-worker/main.go`, `cmd/hgbuild/main.go`
  - Pre-commit: `make test`

- [x] 4. Wire TLS/mTLS with CLI flags, remove Insecure:true

  **What to do**:
  - In `cmd/hg-coord/main.go`:
    - Add CLI flags: `--tls-cert` (string), `--tls-key` (string), `--tls-ca` (string), `--tls-require-client-cert` (bool, default false)
    - If `--tls-cert` AND `--tls-key` are provided, set `TLS.Enabled = true` and populate config fields
    - Pass `hgtls.Config` to `coordserver.Config{..., TLS: tlsCfg}` — server already conditionally wires TLS at `grpc.go:181-193`
    - Log TLS status on startup (enabled/disabled, mTLS yes/no)
  - In `cmd/hg-worker/main.go`:
    - Same TLS flags as coordinator
    - Pass to `workerserver.Config{..., TLS: tlsCfg}` — server already wires at `grpc.go:85-97`
    - **Remove `Insecure: true` on line 110** in the `client.New()` call — replace with conditional: if TLS cert provided, use TLS credentials; otherwise use `insecure.NewCredentials()` (no hardcoded insecure)
    - Add `// TODO: Add TLS client credentials from --tls-* flags` comment if full client TLS is complex; at minimum remove the hardcoded `Insecure: true`
  - In `cmd/hgbuild/main.go`:
    - Add `--tls-cert`, `--tls-key`, `--tls-ca` flags to root command
    - When connecting to coordinator, if TLS flags provided, use TLS transport credentials
    - If no TLS flags, use insecure credentials (but not hardcoded `Insecure: true` — derive from absence of flags)

  **Must NOT do**:
  - Do NOT wire auth interceptor — deferred to v0.3.0
  - Do NOT modify `internal/security/tls/` library code — it's already complete
  - Do NOT modify `internal/coordinator/server/grpc.go` or `internal/worker/server/grpc.go`
  - Do NOT make TLS required by default — it's opt-in via CLI flags

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Touches 3 binary files, TLS wiring requires careful credential handling
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Task 3 in Wave 2, since they touch same files but different sections)
  - **Parallel Group**: Wave 2 (with Tasks 3, 5, 9, 11)
  - **Blocks**: Tasks 12, 13
  - **Blocked By**: Tasks 1, 2

  **References**:

  **Pattern References** (existing code to follow):
  - `internal/worker/server/grpc.go:84-97` — Server-side TLS wiring pattern (conditional on `s.config.TLS.Enabled`)
  - `internal/coordinator/server/grpc.go:180-193` — Same pattern in coordinator
  - `cmd/hg-worker/main.go:106-111` — The `Insecure: true` that needs removal (line 110)
  - `internal/grpc/client/client.go` — Client gRPC dial code (check how it uses `Insecure` field)

  **API/Type References** (contracts to implement against):
  - `internal/security/tls/config.go:9-30` — `tls.Config` struct — fields to populate from CLI flags
  - `internal/security/tls/loader.go` — `ServerCredentials(cfg Config)` and `ClientCredentials(cfg Config)` — functions to call
  - `internal/coordinator/server/grpc.go:86` — `Config.TLS hgtls.Config` field to populate
  - `internal/worker/server/grpc.go:28` — `Config.TLS hgtls.Config` field to populate

  **WHY Each Reference Matters**:
  - `worker/server/grpc.go:84-97` — Proves server-side TLS already works; just need to populate the config
  - `cmd/hg-worker/main.go:110` — The exact line with `Insecure: true` that must be removed
  - `tls/loader.go` — `ServerCredentials()` and `ClientCredentials()` are the functions to call for credential setup

  **Acceptance Criteria**:
  - [ ] `hg-coord serve --help` shows `--tls-cert`, `--tls-key`, `--tls-ca`, `--tls-require-client-cert`
  - [ ] `hg-worker serve --help` shows same TLS flags
  - [ ] `hgbuild --help` shows `--tls-cert`, `--tls-key`, `--tls-ca`
  - [ ] No hardcoded `Insecure: true` remains in `cmd/hg-worker/main.go`
  - [ ] `grep -r 'Insecure:.*true' cmd/` returns empty (no hardcoded insecure in any binary)
  - [ ] Without TLS flags, binaries still start normally (insecure by default via absence of flags)
  - [ ] `go build ./cmd/...` succeeds
  - [ ] `make test` passes

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: Coordinator help shows TLS flags
    Tool: Bash
    Preconditions: Task 4 changes applied, binary built
    Steps:
      1. Run: go build -o /tmp/hg-coord ./cmd/hg-coord
      2. Run: /tmp/hg-coord serve --help
      3. Assert output contains: "--tls-cert", "--tls-key", "--tls-ca"
    Expected Result: All 3 flags listed
    Failure Indicators: Any TLS flag missing from help output
    Evidence: .sisyphus/evidence/task-4-coord-tls-help.txt

  Scenario: No hardcoded Insecure:true in binaries
    Tool: Bash
    Preconditions: Task 4 changes applied
    Steps:
      1. Run: grep -rn 'Insecure:.*true' cmd/
    Expected Result: No matches (exit code 1, empty output)
    Failure Indicators: Any match found
    Evidence: .sisyphus/evidence/task-4-no-hardcoded-insecure.txt

  Scenario: Worker starts without TLS flags (insecure default)
    Tool: Bash
    Preconditions: Binary built
    Steps:
      1. Run: timeout 3 /tmp/hg-worker serve --coordinator=localhost:19999 --port=19001 2>&1 || true
      2. Check output does NOT contain "TLS enabled"
    Expected Result: No TLS log line — starts in insecure mode by default
    Failure Indicators: TLS-related error or unexpected TLS init
    Evidence: .sisyphus/evidence/task-4-worker-no-tls-default.txt
  ```

  **Commit**: YES
  - Message: `feat(cmd): wire TLS/mTLS with CLI flags, remove Insecure:true`
  - Files: `cmd/hg-coord/main.go`, `cmd/hg-worker/main.go`, `cmd/hgbuild/main.go`
  - Pre-commit: `make test`

- [x] 5. Wire request ID interceptor into coordinator and worker servers

  **What to do**:
  - In `cmd/hg-coord/main.go`:
    - Import the request ID interceptor package (created in Task 6)
    - Add the unary and stream server interceptors to the gRPC server options
    - Since coordinator creates its server via `coordserver.New(cfg)` and `coordserver.Start()` handles gRPC setup internally, you may need to either:
      (a) Add a `RequestIDInterceptor bool` field to `coordserver.Config` and wire it in `Start()`, OR
      (b) Pass interceptors via a new config field `AdditionalInterceptors []grpc.ServerOption`
    - Choose option (a) for simplicity — add `EnableRequestID bool` to `coordserver.Config`, default true, and wire in `Start()` alongside TLS/tracing interceptors
  - In `cmd/hg-worker/main.go`:
    - Same approach — add `EnableRequestID bool` to `workerserver.Config`, default true
    - Wire in `workerserver.Start()` alongside existing TLS/tracing interceptors
  - Add request ID to log context: after interceptor extracts/generates request ID, add it to zerolog context so all subsequent logs include it

  **Must NOT do**:
  - Do NOT add auth interceptor — deferred to v0.3.0
  - Do NOT modify the request ID interceptor itself (that's Task 6)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Small config additions and interceptor wiring in 2 server files + 2 main files
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 3, 4, 9, 11 in Wave 2)
  - **Parallel Group**: Wave 2
  - **Blocks**: None
  - **Blocked By**: Task 6 (interceptor must exist first)

  **References**:

  **Pattern References** (existing code to follow):
  - `internal/coordinator/server/grpc.go:195-199` — Where tracing interceptors are added (add request ID interceptor in same block)
  - `internal/worker/server/grpc.go:99-103` — Same pattern in worker server
  - `internal/coordinator/server/grpc.go:80-88` — `Config` struct (add `EnableRequestID bool` here)
  - `internal/worker/server/grpc.go:24-30` — Worker `Config` struct (add field here)

  **API/Type References** (contracts to implement against):
  - Task 6 output: `internal/grpc/interceptors/requestid.go` — The interceptor functions to call

  **WHY Each Reference Matters**:
  - `server/grpc.go:195-199` — Exact insertion point: after tracing interceptors, before `grpc.NewServer(opts...)`
  - Config structs — Need `EnableRequestID` field so wiring is conditional (matching TLS/tracing pattern)

  **Acceptance Criteria**:
  - [ ] `coordserver.Config` has `EnableRequestID bool` field
  - [ ] `workerserver.Config` has `EnableRequestID bool` field
  - [ ] `DefaultConfig()` in both packages sets `EnableRequestID: true`
  - [ ] Server `Start()` adds request ID interceptors when `EnableRequestID` is true
  - [ ] `make test` passes
  - [ ] `go build ./...` succeeds

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: Coordinator server config includes EnableRequestID
    Tool: Bash
    Preconditions: Task 5 changes applied
    Steps:
      1. Run: go doc ./internal/coordinator/server Config
      2. Assert output contains "EnableRequestID"
    Expected Result: Field listed in Config struct
    Failure Indicators: Field missing from output
    Evidence: .sisyphus/evidence/task-5-coord-config.txt

  Scenario: Build and tests pass with request ID wiring
    Tool: Bash
    Preconditions: Task 5 changes applied
    Steps:
      1. Run: go build ./...
      2. Run: make test
    Expected Result: Both succeed with exit code 0
    Failure Indicators: Compilation errors or test failures
    Evidence: .sisyphus/evidence/task-5-build-test.txt
  ```

  **Commit**: YES
  - Message: `feat(cmd): wire request ID interceptor into coordinator and worker`
  - Files: `internal/coordinator/server/grpc.go`, `internal/worker/server/grpc.go`, `cmd/hg-coord/main.go`, `cmd/hg-worker/main.go`
  - Pre-commit: `make test`

- [x] 6. Create request ID gRPC interceptor

  **What to do**:
  - Create new file `internal/grpc/interceptors/requestid.go`
  - Create package `interceptors` (new package — `internal/grpc/interceptors/`)
  - Implement unary server interceptor: `UnaryRequestIDInterceptor() grpc.UnaryServerInterceptor`
    - Extract request ID from incoming gRPC metadata key `x-request-id`
    - If not present, generate a new UUID (use `crypto/rand` or simple unique ID — no external dependency)
    - Add request ID to context (use a context key)
    - Add request ID to outgoing gRPC metadata (for propagation to downstream calls)
    - Log request ID with zerolog
  - Implement stream server interceptor: `StreamRequestIDInterceptor() grpc.StreamServerInterceptor`
    - Same logic: extract or generate, add to context
  - Implement client interceptors: `UnaryRequestIDClientInterceptor() grpc.UnaryClientInterceptor` and `StreamRequestIDClientInterceptor() grpc.StreamClientInterceptor`
    - Propagate request ID from context to outgoing metadata
  - Export context key getter: `RequestIDFromContext(ctx context.Context) string`
  - Write tests in `internal/grpc/interceptors/requestid_test.go`
    - Test: incoming request with x-request-id → preserved in context
    - Test: incoming request without x-request-id → UUID generated
    - Test: RequestIDFromContext returns correct value

  **Must NOT do**:
  - Do NOT add external UUID library — use `crypto/rand` + hex encoding or `fmt.Sprintf` with random bytes
  - Do NOT add auth interceptor in this package — deferred to v0.3.0 (add TODO comment: `// TODO(v0.3.0): Add auth interceptor`)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Single new package with clear interceptor pattern, well-scoped
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 7, 8, 10)
  - **Blocks**: Task 5
  - **Blocked By**: None (can start immediately)

  **References**:

  **Pattern References** (existing code to follow):
  - `internal/observability/tracing/grpc.go` — How tracing interceptors are structured (ServerOptions returns `[]grpc.ServerOption`)
  - `internal/security/auth/interceptor.go` — Auth interceptor pattern (metadata extraction from gRPC context)

  **API/Type References** (contracts to implement against):
  - `google.golang.org/grpc` — `UnaryServerInterceptor`, `StreamServerInterceptor` types
  - `google.golang.org/grpc/metadata` — `FromIncomingContext()`, `AppendToOutgoingContext()`

  **External References**:
  - gRPC-Go interceptor pattern: https://pkg.go.dev/google.golang.org/grpc#UnaryServerInterceptor

  **WHY Each Reference Matters**:
  - `tracing/grpc.go` — Shows how to package interceptors as `ServerOptions()` returning `[]grpc.ServerOption`
  - `auth/interceptor.go` — Shows gRPC metadata extraction pattern (`metadata.FromIncomingContext`)

  **Acceptance Criteria**:
  - [ ] `internal/grpc/interceptors/requestid.go` exists
  - [ ] `UnaryRequestIDInterceptor()` returns `grpc.UnaryServerInterceptor`
  - [ ] `StreamRequestIDInterceptor()` returns `grpc.StreamServerInterceptor`
  - [ ] `RequestIDFromContext(ctx)` returns string
  - [ ] Generated IDs are unique (test 100 IDs, all different)
  - [ ] Existing x-request-id from metadata is preserved
  - [ ] Tests pass: `go test -v ./internal/grpc/interceptors/...`

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: Request ID generated when not present
    Tool: Bash
    Preconditions: Task 6 changes applied
    Steps:
      1. Run: go test -v -run TestUnaryInterceptor_GeneratesID ./internal/grpc/interceptors/...
    Expected Result: Test passes, request ID is non-empty 32+ char hex string
    Failure Indicators: Empty ID or test failure
    Evidence: .sisyphus/evidence/task-6-generate-id.txt

  Scenario: Request ID preserved from metadata
    Tool: Bash
    Preconditions: Task 6 changes applied
    Steps:
      1. Run: go test -v -run TestUnaryInterceptor_PreservesID ./internal/grpc/interceptors/...
    Expected Result: Request ID matches the one sent in metadata
    Failure Indicators: ID mismatch or test failure
    Evidence: .sisyphus/evidence/task-6-preserve-id.txt

  Scenario: All interceptor tests pass
    Tool: Bash
    Preconditions: Task 6 changes applied
    Steps:
      1. Run: go test -v -count=1 ./internal/grpc/interceptors/...
    Expected Result: All tests pass
    Failure Indicators: Any FAIL
    Evidence: .sisyphus/evidence/task-6-all-tests.txt
  ```

  **Commit**: YES
  - Message: `feat(grpc): add request ID interceptor for correlation tracing`
  - Files: `internal/grpc/interceptors/requestid.go`, `internal/grpc/interceptors/requestid_test.go`
  - Pre-commit: `go test ./internal/grpc/interceptors/...`

- [x] 7. Create runtime log-level HTTP handler package

  **What to do**:
  - Create new package `internal/logging/` with file `handler.go`
  - Implement `NewLogLevelHandler() http.Handler` that supports:
    - `GET /log-level` → returns current log level as JSON: `{"level": "info"}`
    - `PUT /log-level` or `POST /log-level` → accepts JSON body `{"level": "debug"}` → changes zerolog global level
    - Validate level string: must be one of "trace", "debug", "info", "warn", "error", "fatal", "panic"
    - Return 400 with error message for invalid level
    - Return 200 with `{"level": "debug", "previous": "info"}` on success
  - Use `zerolog.SetGlobalLevel()` to change level at runtime
  - Log the level change itself: `log.Info().Str("from", prev).Str("to", new).Msg("Log level changed")`
  - Write tests in `internal/logging/handler_test.go`:
    - Test GET returns current level
    - Test PUT with valid level changes level
    - Test PUT with invalid level returns 400
    - Test POST also works (both methods supported)
    - Use `httptest.NewServer` for testing

  **Must NOT do**:
  - Do NOT add authentication to this endpoint in v0.2.3 — add TODO comment: `// TODO(v0.3.0): Add authentication for log-level endpoint`
  - Do NOT wire this into main.go files yet (that happens in Task 3/4 via the HTTP mux)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Single HTTP handler with clear API, well-scoped
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 6, 8, 10)
  - **Blocks**: Tasks 3, 4 (they wire this handler into binary HTTP mux)
  - **Blocked By**: None (can start immediately)

  **References**:

  **Pattern References** (existing code to follow):
  - `cmd/hg-worker/main.go:170-186` — HTTP mux setup pattern (this handler will be registered on this mux in a later task)
  - `internal/observability/dashboard/server.go:58` — How dashboard wraps HTTP mux (similar pattern)

  **API/Type References**:
  - `github.com/rs/zerolog` — `zerolog.SetGlobalLevel()`, `zerolog.ParseLevel()`, `zerolog.GlobalLevel()`

  **WHY Each Reference Matters**:
  - `cmd/hg-worker/main.go:170-186` — Shows the existing HTTP mux where this handler will be registered
  - zerolog API — `ParseLevel()` validates level strings; `SetGlobalLevel()` is the actual mutation

  **Acceptance Criteria**:
  - [ ] `internal/logging/handler.go` exists
  - [ ] GET returns JSON with current level
  - [ ] PUT/POST with valid level changes zerolog global level and returns 200
  - [ ] PUT/POST with invalid level returns 400 with error message
  - [ ] Response includes both new and previous level
  - [ ] Tests pass: `go test -v ./internal/logging/...`

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: GET returns current log level
    Tool: Bash
    Preconditions: Task 7 changes applied
    Steps:
      1. Run: go test -v -run TestLogLevelHandler_GET ./internal/logging/...
    Expected Result: Test passes, response body is valid JSON with "level" field
    Failure Indicators: Missing "level" key or non-JSON response
    Evidence: .sisyphus/evidence/task-7-get-level.txt

  Scenario: PUT changes log level
    Tool: Bash
    Preconditions: Task 7 changes applied
    Steps:
      1. Run: go test -v -run TestLogLevelHandler_PUT ./internal/logging/...
    Expected Result: Test passes, level changed, response includes "previous" and "level" fields
    Failure Indicators: Level not actually changed or response missing fields
    Evidence: .sisyphus/evidence/task-7-put-level.txt

  Scenario: Invalid level returns 400
    Tool: Bash
    Preconditions: Task 7 changes applied
    Steps:
      1. Run: go test -v -run TestLogLevelHandler_InvalidLevel ./internal/logging/...
    Expected Result: 400 status code returned with error message
    Failure Indicators: 200 returned for invalid level
    Evidence: .sisyphus/evidence/task-7-invalid-level.txt
  ```

  **Commit**: YES
  - Message: `feat(logging): add runtime log-level HTTP handler`
  - Files: `internal/logging/handler.go`, `internal/logging/handler_test.go`
  - Pre-commit: `go test ./internal/logging/...`

- [x] 8. Make LogConfig.File functional with zerolog file writer

  **What to do**:
  - Create `internal/logging/writer.go`
  - Implement `SetupFileWriter(filePath string) (io.WriteCloser, error)`:
    - Open file for append (create if not exists, mode 0644)
    - Return the file handle as WriteCloser
  - Implement `SetupLogger(cfg config.LogConfig) zerolog.Logger`:
    - If `cfg.File` is empty → use `zerolog.ConsoleWriter{Out: os.Stderr}` (current behavior)
    - If `cfg.File` is set → use `SetupFileWriter(cfg.File)` and write JSON to file
    - If `cfg.Format == "json"` → no ConsoleWriter wrapper, raw JSON output
    - If `cfg.Format == "console"` → ConsoleWriter (default, current behavior)
    - Support multi-output: console + file simultaneously using `zerolog.MultiLevelWriter`
    - Set log level from `cfg.Level` using `zerolog.ParseLevel()`
  - Write tests in `internal/logging/writer_test.go`:
    - Test file creation in temp dir
    - Test append behavior (second call appends, doesn't overwrite)
    - Test level parsing
    - Test multi-output (console + file)

  **Must NOT do**:
  - Do NOT add lumberjack here — that's Task 9
  - Do NOT modify existing main.go files — they will wire this in Tasks 3/4
  - Do NOT break existing behavior — if LogConfig.File is empty, behavior must be identical to current

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Small file I/O + zerolog configuration, clear scope
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 6, 7, 10)
  - **Blocks**: Task 9
  - **Blocked By**: None (can start immediately)

  **References**:

  **Pattern References** (existing code to follow):
  - `cmd/hg-coord/main.go:22-23` — Current logger setup: `zerolog.TimeFieldFormat = zerolog.TimeFormatUnix; log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})`
  - `cmd/hg-worker/main.go:30-31` — Same pattern in worker

  **API/Type References**:
  - `internal/config/config.go:92-97` — `LogConfig` struct with Level, Format, File fields
  - `github.com/rs/zerolog` — `zerolog.MultiLevelWriter()`, `zerolog.ParseLevel()`, `zerolog.New()`

  **WHY Each Reference Matters**:
  - `main.go:22-23` — This is the code that `SetupLogger()` replaces; must produce identical output when File is empty
  - `LogConfig` struct — Defines the input contract for `SetupLogger()`

  **Acceptance Criteria**:
  - [ ] `internal/logging/writer.go` exists
  - [ ] `SetupFileWriter()` creates file and returns WriteCloser
  - [ ] `SetupLogger()` with empty File returns console-writer logger (same as current)
  - [ ] `SetupLogger()` with File set writes logs to that file
  - [ ] File logs are in JSON format (for machine parsing)
  - [ ] Level setting works: "debug" shows debug logs, "error" hides info
  - [ ] Tests pass: `go test -v ./internal/logging/...`

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: File writer creates and writes to file
    Tool: Bash
    Preconditions: Task 8 changes applied
    Steps:
      1. Run: go test -v -run TestSetupFileWriter ./internal/logging/...
    Expected Result: File created in temp dir, content written, no error
    Failure Indicators: File not created or write error
    Evidence: .sisyphus/evidence/task-8-file-writer.txt

  Scenario: SetupLogger without file matches current behavior
    Tool: Bash
    Preconditions: Task 8 changes applied
    Steps:
      1. Run: go test -v -run TestSetupLogger_NoFile ./internal/logging/...
    Expected Result: Logger outputs to stderr with console format
    Failure Indicators: Output format differs from current zerolog ConsoleWriter
    Evidence: .sisyphus/evidence/task-8-no-file-default.txt

  Scenario: SetupLogger with file writes JSON to file
    Tool: Bash
    Preconditions: Task 8 changes applied
    Steps:
      1. Run: go test -v -run TestSetupLogger_WithFile ./internal/logging/...
    Expected Result: File contains valid JSON log lines
    Failure Indicators: File empty or non-JSON content
    Evidence: .sisyphus/evidence/task-8-file-json.txt
  ```

  **Commit**: YES
  - Message: `feat(logging): make LogConfig.File functional with zerolog file writer`
  - Files: `internal/logging/writer.go`, `internal/logging/writer_test.go`
  - Pre-commit: `go test ./internal/logging/...`

- [x] 9. Add lumberjack log rotation

  **What to do**:
  - Run `go get gopkg.in/natefinch/lumberjack.v2`
  - Create `internal/logging/rotation.go`
  - Implement `NewRotatingWriter(cfg config.LogRotationConfig, filePath string) io.WriteCloser`:
    - Create `&lumberjack.Logger{Filename: filePath, MaxSize: cfg.MaxSizeMB, MaxBackups: cfg.MaxBackups, MaxAge: cfg.MaxAgeDays, Compress: cfg.Compress}`
    - Return the lumberjack.Logger (it implements io.WriteCloser)
  - Update `SetupLogger()` (from Task 8) to use rotating writer when `LogConfig.File` is set AND `LogConfig.Rotation` has non-zero values:
    - If rotation configured → use `NewRotatingWriter()` instead of plain file writer
    - If rotation not configured (all zeros) → use plain file writer from Task 8
  - Write tests in `internal/logging/rotation_test.go`:
    - Test that rotating writer creates file
    - Test that MaxSize triggers rotation (write enough data to exceed MaxSize)
    - Test that Compress flag is passed through

  **Must NOT do**:
  - Do NOT modify lumberjack library code
  - Do NOT make rotation mandatory — it's opt-in based on config

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Small integration with well-documented library, clear API
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: NO (sequential after Task 8)
  - **Parallel Group**: Wave 2 (after Task 8 file writer is done)
  - **Blocks**: None
  - **Blocked By**: Tasks 1 (for LogRotationConfig), 8 (for SetupLogger base)

  **References**:

  **Pattern References** (existing code to follow):
  - Task 8 output: `internal/logging/writer.go` — `SetupLogger()` and `SetupFileWriter()` to extend

  **API/Type References**:
  - `internal/config/config.go` — `LogRotationConfig` struct (from Task 1)
  - `gopkg.in/natefinch/lumberjack.v2` — `lumberjack.Logger` struct

  **External References**:
  - Lumberjack docs: https://pkg.go.dev/gopkg.in/natefinch/lumberjack.v2

  **WHY Each Reference Matters**:
  - Task 8 writer.go — Must extend, not replace, the file writer logic
  - `LogRotationConfig` — Input contract from Task 1; fields map 1:1 to lumberjack fields

  **Acceptance Criteria**:
  - [ ] `go.mod` includes `gopkg.in/natefinch/lumberjack.v2`
  - [ ] `internal/logging/rotation.go` exists
  - [ ] `NewRotatingWriter()` returns lumberjack-backed WriteCloser
  - [ ] `SetupLogger()` uses rotating writer when rotation config is non-zero
  - [ ] Tests pass: `go test -v ./internal/logging/...`
  - [ ] `make lint` passes (no unused imports from lumberjack)

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: Rotating writer creates log file
    Tool: Bash
    Preconditions: Task 9 changes applied
    Steps:
      1. Run: go test -v -run TestNewRotatingWriter ./internal/logging/...
    Expected Result: File created, writes succeed
    Failure Indicators: File not created or write error
    Evidence: .sisyphus/evidence/task-9-rotating-writer.txt

  Scenario: Rotation config fields passed to lumberjack
    Tool: Bash
    Preconditions: Task 9 changes applied
    Steps:
      1. Run: go test -v -run TestRotatingWriter_Config ./internal/logging/...
    Expected Result: MaxSize, MaxBackups, MaxAge, Compress all match input config
    Failure Indicators: Field mismatch
    Evidence: .sisyphus/evidence/task-9-config-passthrough.txt

  Scenario: All logging tests still pass
    Tool: Bash
    Preconditions: Task 9 changes applied
    Steps:
      1. Run: go test -v -count=1 ./internal/logging/...
    Expected Result: All tests pass (including Task 7 and Task 8 tests)
    Failure Indicators: Any FAIL
    Evidence: .sisyphus/evidence/task-9-all-logging-tests.txt
  ```

  **Commit**: YES
  - Message: `feat(logging): add lumberjack log rotation`
  - Files: `go.mod`, `go.sum`, `internal/logging/rotation.go`, `internal/logging/rotation_test.go`, `internal/logging/writer.go`
  - Pre-commit: `go test ./internal/logging/...`

- [x] 10. Boost test coverage — capability package to ≥60%

  **What to do**:
  - Current coverage: ~24.8%
  - Target: ≥60%
  - Add tests to `internal/capability/` (likely `detect_test.go`)
  - Focus areas:
    - Test `Detect()` function with various system configurations
    - Test compiler detection paths (gcc, g++, clang, clang++ found/not found)
    - Test Docker availability detection
    - Test architecture detection
    - Test OS detection
    - Mock `exec.LookPath` or use test helpers that modify PATH to control which compilers are "found"
  - Use table-driven tests with `t.Run` subtests (project convention)
  - Use `t.TempDir()` for any temp file needs

  **Must NOT do**:
  - Do NOT test MSVC detection (platform-specific, not available on macOS)
  - Do NOT add external test frameworks — use stdlib `testing`
  - Do NOT modify the capability detection code itself — only add tests

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Requires understanding system-level detection code, creative mocking strategies
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 6, 7, 8)
  - **Blocks**: None
  - **Blocked By**: None (can start immediately)

  **References**:

  **Pattern References** (existing code to follow):
  - `internal/capability/detect.go` — The code under test (Detect function)
  - `internal/cache/store_test.go` — Table-driven test pattern to follow
  - `internal/coordinator/scheduler/scheduler_test.go` — Another test pattern example

  **Test References**:
  - `internal/capability/detect_test.go` — Existing tests (if any) to extend

  **WHY Each Reference Matters**:
  - `detect.go` — Must understand what Detect() does to write meaningful tests
  - Existing test files — Follow project's testing conventions (manual assertions, table-driven)

  **Acceptance Criteria**:
  - [ ] `go test -cover ./internal/capability/...` shows ≥60% coverage
  - [ ] All new tests pass
  - [ ] No MSVC/Windows-specific tests included
  - [ ] Tests are table-driven with `t.Run`

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: Capability coverage meets target
    Tool: Bash
    Preconditions: Task 10 changes applied
    Steps:
      1. Run: go test -v -cover ./internal/capability/... 2>&1
      2. Parse coverage percentage from output
    Expected Result: Coverage ≥ 60.0%
    Failure Indicators: Coverage below 60%
    Evidence: .sisyphus/evidence/task-10-capability-coverage.txt

  Scenario: All capability tests pass
    Tool: Bash
    Preconditions: Task 10 changes applied
    Steps:
      1. Run: go test -v -count=1 -race ./internal/capability/...
    Expected Result: All tests PASS, no race conditions
    Failure Indicators: Any FAIL or DATA RACE
    Evidence: .sisyphus/evidence/task-10-capability-tests.txt
  ```

  **Commit**: YES
  - Message: `test(capability): boost coverage to ≥60%`
  - Files: `internal/capability/*_test.go`
  - Pre-commit: `go test -cover ./internal/capability/...`

- [x] 11. Boost test coverage — executor package to ≥60%

  **What to do**:
  - Current coverage: ~24.2%
  - Target: ≥60%
  - Add tests to `internal/worker/executor/`
  - Focus areas:
    - Test `Manager` creation and configuration
    - Test native executor: command construction, argument handling, timeout behavior
    - Test Docker executor: image selection, volume mounting (mock Docker client)
    - Test executor selection logic (native vs Docker based on capabilities)
    - Test error handling paths: command not found, timeout, non-zero exit code
  - Use manual mock structs for Docker client (project convention — no mocking frameworks)
  - Skip tests that require actual Docker daemon with build tag or runtime check

  **Must NOT do**:
  - Do NOT test MSVC executor (platform-specific)
  - Do NOT require Docker daemon for unit tests — mock the Docker interface
  - Do NOT modify executor code — only add tests

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Complex executor code with Docker mocking, requires careful test design
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 3, 4, 5, 9)
  - **Blocks**: None
  - **Blocked By**: None (independent of other tasks)

  **References**:

  **Pattern References** (existing code to follow):
  - `internal/worker/executor/executor.go` — Manager and executor interfaces
  - `internal/worker/executor/native.go` — Native executor implementation
  - `internal/worker/executor/docker.go` — Docker executor implementation
  - `internal/worker/executor/executor_test.go` — Existing tests to extend

  **WHY Each Reference Matters**:
  - executor.go — Understand Manager API and executor interface for mocking
  - native.go — Target for native executor tests
  - docker.go — Target for Docker executor tests (need to mock Docker client interface)

  **Acceptance Criteria**:
  - [ ] `go test -cover ./internal/worker/executor/...` shows ≥60% coverage
  - [ ] All tests pass without Docker daemon
  - [ ] No MSVC executor tests
  - [ ] Tests are table-driven where applicable

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: Executor coverage meets target
    Tool: Bash
    Preconditions: Task 11 changes applied
    Steps:
      1. Run: go test -v -cover ./internal/worker/executor/... 2>&1
      2. Parse coverage percentage from output
    Expected Result: Coverage ≥ 60.0%
    Failure Indicators: Coverage below 60%
    Evidence: .sisyphus/evidence/task-11-executor-coverage.txt

  Scenario: All executor tests pass without Docker
    Tool: Bash
    Preconditions: Task 11 changes applied
    Steps:
      1. Run: go test -v -count=1 -race ./internal/worker/executor/...
    Expected Result: All tests PASS, no tests skipped due to Docker requirement
    Failure Indicators: Any FAIL or Docker-required skip
    Evidence: .sisyphus/evidence/task-11-executor-tests.txt
  ```

  **Commit**: YES
  - Message: `test(executor): boost coverage to ≥60%`
  - Files: `internal/worker/executor/*_test.go`
  - Pre-commit: `go test -cover ./internal/worker/executor/...`

- [x] 12. Boost test coverage — cli/build to ≥60% and cli/output to ≥70%

  **What to do**:
  - Current coverage: cli/build ~31.7%, cli/output ~43.4%
  - Targets: cli/build ≥60%, cli/output ≥70%
  - For `internal/cli/build/`:
    - Test command construction for `make`, `ninja`, `cc`, `c++` subcommands
    - Test compiler argument parsing (flags, includes, defines)
    - Test fallback behavior configuration
    - Test verbose mode output differences
  - For `internal/cli/output/`:
    - Test table formatting with various column widths
    - Test color code application
    - Test progress bar rendering
    - Test status tag formatting ([cache], [remote], [local])
    - Test edge cases: empty data, very long strings, unicode
  - Both packages: table-driven tests, stdlib `testing`

  **Must NOT do**:
  - Do NOT modify CLI code — only add tests
  - Do NOT test actual gRPC connections in unit tests — mock the client interface
  - Do NOT test hgbuild's main() — test the library functions it calls

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Two packages, significant test writing, requires understanding CLI internals
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Task 13)
  - **Blocks**: Task 13
  - **Blocked By**: Tasks 3, 4 (cli code may change with new flags)

  **References**:

  **Pattern References** (existing code to follow):
  - `internal/cli/build/` — Build command implementation
  - `internal/cli/output/` — Output formatting implementation
  - `internal/cli/build/build_test.go` — Existing tests to extend
  - `internal/cli/output/table_test.go` — Existing tests to extend

  **WHY Each Reference Matters**:
  - Existing test files — Extend, don't rewrite; follow established patterns

  **Acceptance Criteria**:
  - [ ] `go test -cover ./internal/cli/build/...` shows ≥60%
  - [ ] `go test -cover ./internal/cli/output/...` shows ≥70%
  - [ ] All tests pass
  - [ ] No gRPC connections in unit tests

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: cli/build coverage meets target
    Tool: Bash
    Preconditions: Task 12 changes applied
    Steps:
      1. Run: go test -v -cover ./internal/cli/build/... 2>&1
    Expected Result: Coverage ≥ 60.0%
    Failure Indicators: Coverage below 60%
    Evidence: .sisyphus/evidence/task-12-cli-build-coverage.txt

  Scenario: cli/output coverage meets target
    Tool: Bash
    Preconditions: Task 12 changes applied
    Steps:
      1. Run: go test -v -cover ./internal/cli/output/... 2>&1
    Expected Result: Coverage ≥ 70.0%
    Failure Indicators: Coverage below 70%
    Evidence: .sisyphus/evidence/task-12-cli-output-coverage.txt

  Scenario: All CLI tests pass with race detector
    Tool: Bash
    Preconditions: Task 12 changes applied
    Steps:
      1. Run: go test -v -race -count=1 ./internal/cli/build/... ./internal/cli/output/...
    Expected Result: All tests PASS, no DATA RACE
    Failure Indicators: Any FAIL or race condition
    Evidence: .sisyphus/evidence/task-12-all-cli-tests.txt
  ```

  **Commit**: YES
  - Message: `test(cli): boost cli/build to ≥60% and cli/output to ≥70%`
  - Files: `internal/cli/build/*_test.go`, `internal/cli/output/*_test.go`
  - Pre-commit: `go test -cover ./internal/cli/build/... ./internal/cli/output/...`

- [x] 13. Generate CHANGELOG.md with release history + Makefile target

  **What to do**:
  - Create `CHANGELOG.md` at project root with entries for v0.2.0, v0.2.1, v0.2.2, v0.2.3
  - Format: Keep a Changelog style (https://keepachangelog.com/)
  - For v0.2.0–v0.2.2: summarize from git tags and commit messages (`git log v0.2.0..v0.2.1 --oneline`)
  - For v0.2.3: list all changes made in this plan (config validation, OTel wiring, TLS wiring, request ID, logging, test coverage)
  - Add Makefile target `changelog`:
    - Script that generates changelog from git tags
    - Or: simple target that reminds to update manually: `@echo "Update CHANGELOG.md manually before release"`
    - Prefer: a small shell script in `scripts/changelog.sh` that extracts commit messages between tags
  - Add `make changelog` to the Makefile

  **Must NOT do**:
  - Do NOT add external changelog tools (no git-cliff, conventional-changelog, etc.)
  - Do NOT auto-generate without review — the script should produce a DRAFT that gets reviewed

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Markdown file + simple Makefile target, no code logic
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Task 12 in Wave 3)
  - **Parallel Group**: Wave 3
  - **Blocks**: Task 14
  - **Blocked By**: All code tasks (needs to summarize what was done)

  **References**:

  **Pattern References**:
  - `Makefile` — Existing targets to follow format
  - `README.md` "What's New in v0.2.2" section — Content source for v0.2.2 entry

  **External References**:
  - Keep a Changelog format: https://keepachangelog.com/en/1.1.0/

  **WHY Each Reference Matters**:
  - Makefile — Follow existing target naming and style
  - README.md — Has accurate v0.2.2 summary to copy into CHANGELOG

  **Acceptance Criteria**:
  - [ ] `CHANGELOG.md` exists at project root
  - [ ] Contains entries for v0.2.0, v0.2.1, v0.2.2, v0.2.3
  - [ ] Follows Keep a Changelog format (## [version] - date, ### Added/Changed/Fixed)
  - [ ] `make changelog` target exists and runs without error
  - [ ] v0.2.3 entry accurately lists all changes from this plan

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: CHANGELOG.md exists and has all versions
    Tool: Bash
    Preconditions: Task 13 changes applied
    Steps:
      1. Run: test -f CHANGELOG.md && echo "EXISTS" || echo "MISSING"
      2. Run: grep -c '## \[v0.2' CHANGELOG.md
    Expected Result: File exists, grep finds ≥4 version headers (v0.2.0, v0.2.1, v0.2.2, v0.2.3)
    Failure Indicators: File missing or fewer than 4 version entries
    Evidence: .sisyphus/evidence/task-13-changelog-exists.txt

  Scenario: Make changelog target works
    Tool: Bash
    Preconditions: Task 13 changes applied
    Steps:
      1. Run: make changelog
    Expected Result: Exit code 0, no errors
    Failure Indicators: make error or missing target
    Evidence: .sisyphus/evidence/task-13-make-changelog.txt
  ```

  **Commit**: YES
  - Message: `docs: generate CHANGELOG.md with release history`
  - Files: `CHANGELOG.md`, `Makefile`, `scripts/changelog.sh` (if created)
  - Pre-commit: `make changelog`

- [x] 14. Update CHECKLIST.md to reflect v0.2.3 completions

  **What to do**:
  - Read current `CHECKLIST.md`
  - Mark the following as completed:
    - Config validation
    - OTel/tracing startup wiring
    - TLS startup wiring
    - Request ID tracing
    - Runtime log level endpoint
    - File-based logging
    - Log rotation
  - Update coverage numbers in any coverage-tracking sections
  - Add any new items that emerged during implementation
  - Verify accuracy: every checkbox marked "done" must correspond to actual code

  **Must NOT do**:
  - Do NOT mark Flutter, Unity, or WAN features as done
  - Do NOT remove items — only update status

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Single markdown file update, straightforward
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: NO (must be last code task)
  - **Parallel Group**: Wave 4 (after all code changes)
  - **Blocks**: F1-F4 (final verification)
  - **Blocked By**: Task 13 (and transitively all code tasks)

  **References**:

  **Pattern References**:
  - `CHECKLIST.md` — The file to update

  **WHY Each Reference Matters**:
  - CHECKLIST.md — Must read current state to know what to mark as done

  **Acceptance Criteria**:
  - [ ] All v0.2.3 items marked as completed
  - [ ] No v0.3.0 items marked as completed
  - [ ] Coverage numbers updated
  - [ ] File is valid markdown

  **QA Scenarios (MANDATORY)**:

  ```
  Scenario: CHECKLIST reflects v0.2.3 completions
    Tool: Bash
    Preconditions: Task 14 changes applied
    Steps:
      1. Run: grep -c '\[x\]' CHECKLIST.md
      2. Run: grep -i 'config validation' CHECKLIST.md
      3. Run: grep -i 'tracing' CHECKLIST.md
    Expected Result: Config validation and tracing items marked [x]
    Failure Indicators: Items still marked [ ] after implementation
    Evidence: .sisyphus/evidence/task-14-checklist-updated.txt

  Scenario: v0.3.0 items NOT marked complete
    Tool: Bash
    Preconditions: Task 14 changes applied
    Steps:
      1. Run: grep -i 'flutter\|unity\|wan' CHECKLIST.md | grep '\[x\]'
    Expected Result: No matches (exit code 1) — v0.3.0 items still unchecked
    Failure Indicators: Any v0.3.0 item marked [x]
    Evidence: .sisyphus/evidence/task-14-no-v030-marked.txt
  ```

  **Commit**: YES
  - Message: `docs: update CHECKLIST.md to reflect v0.2.3 completions`
  - Files: `CHECKLIST.md`
  - Pre-commit: —

---

## Final Verification Wave (MANDATORY — after ALL implementation tasks)

> 4 review agents run in PARALLEL. ALL must APPROVE. Rejection → fix → re-run.

- [x] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read file, run binary with flags, run tests). For each "Must NOT Have": search codebase for forbidden patterns (auth interceptor wiring, `--config` flag, Flutter/Unity code) — reject with file:line if found. Check evidence files exist in `.sisyphus/evidence/`. Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [x] F2. **Code Quality Review** — `unspecified-high`
  Run `make test`, `make lint`, `gosec -exclude=G104,G109,G112,G115,G204,G301,G304,G306,G402 ./...`. Review all changed files for: `as any` equivalents, empty error handling (`_ = err`), `fmt.Println` in production code, commented-out code, unused imports. Check for over-abstraction and unnecessary packages.
  Output: `Build [PASS/FAIL] | Lint [PASS/FAIL] | Tests [N pass/N fail] | Gosec [PASS/FAIL] | VERDICT`

- [x] F3. **Real QA — Binary Smoke Test** — `unspecified-high`
  Build all 3 binaries (`make build`). Run each with `--help` and verify new flags appear. Start `hg-coord` with `--tracing-enable --tracing-endpoint=localhost:4317` and verify it starts (check stderr for tracing init log). Start `hg-worker` and `curl -X PUT localhost:{port}/log-level -d '{"level":"debug"}'` — verify 200 response. Test with invalid config values and verify `Validate()` catches them. Save evidence.
  Output: `Binaries [3/3 build] | Flags [N/N present] | Startup [N/N] | HTTP [N/N] | VERDICT`

- [x] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff (`git diff`). Verify 1:1 — everything in spec was built (no missing), nothing beyond spec was built (no creep). Check "Must NOT do" compliance: no auth wiring, no config file loading, no Flutter/Unity code. Detect cross-task contamination: Task N touching Task M's files. Flag unaccounted changes.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | Unaccounted [CLEAN/N files] | VERDICT`

---

## Commit Strategy

| After Task(s) | Commit Message | Files | Pre-commit Check |
|---------------|---------------|-------|-----------------|
| 1 | `feat(config): expand TracingConfig and add LogRotationConfig` | `internal/config/config.go` | `go build ./...` |
| 2 | `feat(config): add Validate() method with field constraints` | `internal/config/config.go`, `internal/config/config_test.go` | `go test ./internal/config/...` |
| 6 | `feat(grpc): add request ID interceptor for correlation tracing` | `internal/grpc/interceptors/requestid.go`, `internal/grpc/interceptors/requestid_test.go` | `go test ./internal/grpc/...` |
| 7 | `feat(logging): add runtime log-level HTTP handler` | `internal/logging/handler.go`, `internal/logging/handler_test.go` | `go test ./internal/logging/...` |
| 8 | `feat(logging): make LogConfig.File functional with zerolog file writer` | `internal/logging/writer.go`, `internal/logging/writer_test.go` | `go test ./internal/logging/...` |
| 9 | `feat(logging): add lumberjack log rotation` | `go.mod`, `go.sum`, `internal/logging/rotation.go`, `internal/logging/rotation_test.go` | `go test ./internal/logging/...` |
| 3 | `feat(cmd): wire OTel tracing with CLI flags in all binaries` | `cmd/hg-coord/main.go`, `cmd/hg-worker/main.go`, `cmd/hgbuild/main.go` | `make test` |
| 4 | `feat(cmd): wire TLS/mTLS with CLI flags, remove Insecure:true` | `cmd/hg-coord/main.go`, `cmd/hg-worker/main.go`, `cmd/hgbuild/main.go` | `make test` |
| 5 | `feat(cmd): wire request ID interceptor into coordinator and worker` | `cmd/hg-coord/main.go`, `cmd/hg-worker/main.go` | `make test` |
| 10 | `test(capability): boost coverage to ≥60%` | `internal/capability/*_test.go` | `go test -cover ./internal/capability/...` |
| 11 | `test(executor): boost coverage to ≥60%` | `internal/worker/executor/*_test.go` | `go test -cover ./internal/worker/executor/...` |
| 12 | `test(cli): boost cli/build to ≥60% and cli/output to ≥70%` | `internal/cli/build/*_test.go`, `internal/cli/output/*_test.go` | `go test -cover ./internal/cli/build/... ./internal/cli/output/...` |
| 13 | `docs: generate CHANGELOG.md with release history` | `CHANGELOG.md`, `Makefile` | `make changelog` (new target) |
| 14 | `docs: update CHECKLIST.md to reflect v0.2.3 completions` | `CHECKLIST.md` | — |

---

## Success Criteria

### Verification Commands
```bash
make build          # Expected: 3 binaries built, 0 errors
make test           # Expected: all tests pass, 0 failures
make lint           # Expected: 0 errors
go test -cover ./internal/capability/...       # Expected: ≥60%
go test -cover ./internal/worker/executor/...  # Expected: ≥60%
go test -cover ./internal/cli/build/...        # Expected: ≥60%
go test -cover ./internal/cli/output/...       # Expected: ≥70%
./bin/hg-coord --help  # Expected: --tracing-enable, --tls-cert flags visible
./bin/hg-worker --help # Expected: --tracing-enable, --tls-cert flags visible
./bin/hgbuild --help   # Expected: --tracing-enable, --tls-cert flags visible
gosec -exclude=G104,G109,G112,G115,G204,G301,G304,G306,G402 ./...  # Expected: clean
```

### Final Checklist
- [ ] All "Must Have" items present and functional
- [ ] All "Must NOT Have" items verified absent
- [ ] All tests pass (`make test`)
- [ ] Lint clean (`make lint`)
- [ ] Security scan clean (`gosec`)
- [ ] Per-package coverage targets met
- [ ] CHANGELOG.md exists with accurate history
- [ ] CHECKLIST.md reflects current state
- [ ] All `.sisyphus/evidence/` files captured
- [x] Ready for v0.2.3 tag and release
