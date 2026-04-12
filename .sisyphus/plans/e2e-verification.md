# End-to-End Verification of Hybrid-Grid v0.2.3

## TL;DR

> **Quick Summary**: Verify that ALL functional features of Hybrid-Grid v0.2.3 work correctly end-to-end using a Docker Compose cluster on macOS. Tests cover compilation pipeline, cache, dashboard API, TLS/mTLS, OTel tracing, log-level endpoint, Prometheus metrics, local fallback, and stress testing with CPython.
>
> **Deliverables**:
> - E2E test Docker infrastructure (test-specific Dockerfile + docker-compose)
> - Small multi-file C test project for pipeline verification
> - Self-signed TLS certificates for security testing
> - Verification scripts per feature with evidence capture
> - CPython stress test execution with performance metrics
> - Findings report documenting any bugs/issues discovered
>
> **Estimated Effort**: Medium (2-3 focused sessions)
> **Parallel Execution**: YES — 4 waves
> **Critical Path**: Task 1 → Task 2 → Task 3 → Task 4 → Tasks 5-8 (parallel) → Task 9 → Task 10 → Task 11

---

## Context

### Original Request
User wants to fully test whether the current version (v0.2.3) with all features actually works end-to-end. Docker Compose on macOS (single machine) as the testing approach.

### Interview Summary
**Key Discussions**:
- Testing approach: Docker Compose cluster (coordinator + workers) on macOS
- Test workload: Small C project first (pipeline verification), then CPython stress test (scale)
- Feature coverage: Everything — compilation, cache, dashboard, TLS, OTel, log-level, metrics
- Build()/StreamBuild() stubs are explicitly OUT of scope

**Research Findings**:
- All-in-one Docker image (alpine:3.19) has NO gcc/g++ — workers can't compile. Must create test image with build tools.
- `--no-fallback` CLI flag documented in README but NOT implemented. Skip that test.
- Coordinator `/health` endpoint does NOT exist — docker-compose healthcheck is silently broken. Document as bug.
- mDNS does NOT work inside Docker bridge network. Use `--coordinator=coordinator:9000` flag instead.
- `hgbuild cc` calls `Compile()` RPC (not `Build()`), so it's fully functional.
- Workers auto-register via gRPC Handshake. No manual setup needed.
- Cache defaults to `~/.hybridgrid/cache`, not configurable via CLI flag.

### Metis Review
**Identified Gaps** (addressed):
- Worker Docker image lacks compilers → Task 1 creates test Dockerfile with build-essential
- mDNS untestable in Docker → Skipped; use explicit `--coordinator` flag
- No `--no-fallback` flag → Test fallback naturally by stopping coordinator
- Coordinator /health missing → Document as bug, use alternative healthcheck
- Cache isolation between test phases → Each phase uses separate cache dir
- Port conflicts between phases → `docker compose down -v` cleanup between phases
- TLS cert generation → Explicit task with openssl commands

---

## Work Objectives

### Core Objective
Prove that every functional feature of Hybrid-Grid v0.2.3 works correctly in a realistic multi-node Docker environment, capturing evidence for each verification.

### Concrete Deliverables
- `test/e2e/Dockerfile.worker` — Test worker image with gcc/g++/make
- `test/e2e/docker-compose.yml` — Test cluster config (coordinator + 2 workers)
- `test/e2e/docker-compose.tls.yml` — TLS-enabled cluster overlay
- `test/e2e/docker-compose.otel.yml` — OTel collector overlay
- `test/e2e/testdata/` — Multi-file C test project
- `test/e2e/certs/` — Self-signed TLS certificates (gitignored)
- `.sisyphus/evidence/` — Screenshots, logs, metrics, test outputs per task
- `.sisyphus/findings.md` — Bugs/issues discovered during verification

### Definition of Done
- [x] Docker cluster starts with coordinator + 2 workers (both healthy)
- [x] `hgbuild cc -c hello.c` compiles via coordinator → worker → returns object file
- [x] `hgbuild make -j4` builds a multi-file C project successfully
- [x] Second identical build shows cache hits in metrics/logs
- [x] Dashboard API returns worker list and task stats
- [x] TLS-secured compilation works with self-signed certs
- [x] OTel traces appear in collector after compilation
- [x] `/log-level` GET/PUT works on coordinator and worker
- [x] Prometheus metrics show correct task/cache counters
- [x] Local fallback compiles when coordinator is down
- [x] CPython stress test completes with distributed compilation

### Must Have
- Actual C/C++ compilation through the full pipeline (not just gRPC flow)
- Cache hit verification via metrics (not just "it worked again")
- TLS with mTLS mutual authentication (server + client certs)
- Evidence files for every verification (logs, curl output, screenshots)
- Clean Docker cleanup between test phases

### Must NOT Have (Guardrails)
- ❌ Do NOT fix bugs found during testing — only document in `.sisyphus/findings.md`
- ❌ Do NOT modify production code (`cmd/`, `internal/`, `gen/`) — only create test infrastructure
- ❌ Do NOT test `Build()` or `StreamBuild()` RPC stubs (v0.3.0 scope)
- ❌ Do NOT test mDNS inside Docker containers (known to not work with bridge networking)
- ❌ Do NOT test `--no-fallback` flag (doesn't exist in code)
- ❌ Do NOT create permanent test frameworks — keep scripts simple and disposable
- ❌ Do NOT build a PKI infrastructure — self-signed certs only (one CA, one server, one client)
- ❌ Do NOT tune/optimize the CPython stress test — run it AS-IS from `test/stress/`
- ❌ Do NOT add permanent test dependencies to go.mod
- ❌ Do NOT test cross-compilation (ARM→x86 etc.) — same-architecture only

---

## Verification Strategy (MANDATORY)

> **ZERO HUMAN INTERVENTION** — ALL verification is agent-executed. No exceptions.

### Test Decision
- **Infrastructure exists**: YES (Docker, Makefile, integration tests)
- **Automated tests**: None (this IS the verification — we're testing existing code)
- **Framework**: Shell scripts with curl + docker compose + hgbuild CLI

### QA Policy
Every task MUST include agent-executed QA scenarios.
Evidence saved to `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`.

- **Docker cluster**: Use Bash — docker compose commands, health checks, log inspection
- **Compilation**: Use Bash — hgbuild CLI commands, file existence checks
- **API/HTTP endpoints**: Use Bash (curl) — HTTP requests, JSON parsing
- **Metrics**: Use Bash (curl + grep) — Prometheus text format parsing
- **Dashboard**: Use Bash (curl) — API endpoint responses

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Foundation — must complete before anything else):
├── Task 1: Create E2E test Docker infrastructure [deep]
├── Task 2: Create multi-file C test project [quick]
└── Task 3: Generate self-signed TLS certificates [quick]

Wave 2 (Core Verification — after Wave 1):
├── Task 4: Verify Docker cluster startup + worker registration [unspecified-high]
└── Task 5: Verify compilation pipeline (hgbuild cc + make) [deep]

Wave 3 (Feature Verification — after Wave 2, MAX PARALLEL):
├── Task 6: Verify cache hit/miss behavior [unspecified-high]
├── Task 7: Verify dashboard API + Prometheus metrics [unspecified-high]
├── Task 8: Verify log-level endpoint [quick]
├── Task 9: Verify local fallback compilation [unspecified-high]
└── Task 10: Verify TLS/mTLS secured compilation [deep]

Wave 4 (Advanced Verification — after Wave 3, SEQUENTIAL due to port conflicts):
├── Task 11: Verify OTel tracing with collector [deep]
└── Task 12: Run CPython stress test [unspecified-high] (AFTER Task 11 — test/stress binds same ports 9000/8080)

Wave FINAL (Review — after ALL tasks):
├── Task F1: Plan compliance audit [oracle]
├── Task F2: Evidence completeness check [unspecified-high]
├── Task F3: Findings report compilation [writing]
└── Task F4: Scope fidelity check [deep]

Critical Path: Task 1 → Task 4 → Task 5 → Task 6 → Task 10 → Task 11 → Task 12 → F1-F4
Parallel Speedup: ~45% faster than sequential
Max Concurrent: 5 (Wave 3)
```

### Dependency Matrix

| Task | Depends On | Blocks |
|------|-----------|--------|
| 1 | — | 4, 5, 6, 7, 8, 9, 10, 11, 12 |
| 2 | — | 5 |
| 3 | — | 10 |
| 4 | 1 | 5, 6, 7, 8, 9, 10, 11, 12 |
| 5 | 1, 2, 4 | 6, 9 |
| 6 | 5 | — |
| 7 | 4 | — |
| 8 | 4 | — |
| 9 | 5 | — |
| 10 | 3, 4 | 11 |
| 11 | 10 | 12 |
| 12 | 4, 11 | — |

### Agent Dispatch Summary

- **Wave 1** (3 tasks): T1→`deep`, T2→`quick`, T3→`quick`
- **Wave 2** (2 tasks): T4→`unspecified-high`, T5→`deep`
- **Wave 3** (5 tasks): T6→`unspecified-high`, T7→`unspecified-high`, T8→`quick`, T9→`unspecified-high`, T10→`deep`
- **Wave 4** (2 tasks): T11→`deep`, T12→`unspecified-high`
- **FINAL** (4 tasks): F1→`oracle`, F2→`unspecified-high`, F3→`writing`, F4→`deep`

---

## TODOs

- [x] 1. Create E2E Test Docker Infrastructure

  **What to do**:
  - Create `test/e2e/Dockerfile.worker` based on `test/stress/Dockerfile.base` pattern — Debian bookworm-slim with `build-essential gcc g++ make` installed, copies all 3 binaries from Go build stage
  - Create `test/e2e/docker-compose.yml` with:
    - 1 coordinator (hg-coord): ports 9000 (gRPC), 8080 (HTTP/dashboard), command: `hg-coord serve --grpc-port=9000 --http-port=8080`
    - 2 workers (hg-worker): using the custom Dockerfile.worker:
      - worker-1: `hostname: worker-1`, command: `hg-worker serve --coordinator=coordinator:9000 --port=50052 --http-port=9090 --advertise-address=worker-1:50052 --max-parallel=4`, publish `9091:9090` (host port 9091 for worker-1 HTTP access from host)
      - worker-2: `hostname: worker-2`, command: `hg-worker serve --coordinator=coordinator:9000 --port=50052 --http-port=9090 --advertise-address=worker-2:50052 --max-parallel=4`, no host HTTP port (avoid conflict)
      (CRITICAL: worker uses `--port` NOT `--grpc-port` — see `cmd/hg-worker/main.go:328`)
      (CRITICAL: `--advertise-address` MUST match cert SANs in Task 3 — worker falls back to runtime hostname when omitted, see `cmd/hg-worker/main.go:218`, and coordinator stores that address for dials at `internal/coordinator/server/grpc.go:264`)
      (CRITICAL: `hostname: worker-1` in compose ensures the container hostname matches the SAN — without this, Docker assigns random container IDs as hostnames)
    - Bridge network `hg-e2e-net`
    - Healthcheck for workers: `curl -sf http://localhost:9090/health` (worker HAS /health)
      (CRITICAL: use `curl`, NOT `wget` — `test/stress/Dockerfile.base` only installs `curl`, not `wget`)
    - Healthcheck for coordinator: `curl -sf http://localhost:8080/metrics` (coordinator does NOT have /health — use /metrics instead)
      (CRITICAL: use `curl`, NOT `wget` — same reason as above)
    - Resource limits: 512MB memory, 0.5 CPU per container
  - Create `test/e2e/docker-compose.tls.yml` as override file — adds TLS flags: `--tls-cert=/certs/server.crt --tls-key=/certs/server.key --tls-ca=/certs/ca.crt --tls-require-client-cert`, mounts `./certs/` volume. **CRITICAL**: The `--tls-require-client-cert` flag MUST be included on both coordinator and workers to enforce mTLS — without it, the server accepts unauthenticated connections even with TLS enabled (see `cmd/hg-coord/main.go:244` and `internal/security/tls/loader.go:35`)
  - Create `test/e2e/docker-compose.otel.yml` as override file — adds Jaeger all-in-one container (port 16686 UI, 4317 OTLP gRPC), adds `--tracing-enable --tracing-endpoint=jaeger:4317` to coordinator and workers
  - Add `test/e2e/certs/` to `.gitignore`

  **Must NOT do**:
  - Do NOT modify the root `Dockerfile` or `docker-compose.yml`
  - Do NOT use the `all-in-one` image target (lacks compilers)
  - Do NOT set up host networking or macvlan for mDNS

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Docker infrastructure requires understanding the existing Dockerfile multi-stage build, stress test patterns, and careful port/network configuration
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 2, 3)
  - **Parallel Group**: Wave 1
  - **Blocks**: Tasks 4, 5, 6, 7, 8, 9, 10, 11, 12
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `Dockerfile:1-132` — Existing multi-stage build with 4 targets; reuse the Go build stages, adapt the final stage to include build tools
  - `test/stress/Dockerfile.base:1-30` — Debian-based image with build-essential; this IS the pattern to follow for test workers
  - `docker-compose.yml:1-70` — Root compose file; follow same structure but with custom image and /metrics healthcheck for coordinator
  - `test/stress/docker-compose.yml:1-50` — Stress test compose; shows resource limits and scaling pattern

  **API/Type References**:
  - `cmd/hg-coord/main.go:82-169` — Coordinator CLI flags for TLS (--tls-cert, --tls-key, --tls-ca)
  - `cmd/hg-worker/main.go:177-196` — Worker CLI flags for TLS
  - `cmd/hg-coord/main.go:103-133` — Coordinator CLI flags for tracing (--tracing-enable, --tracing-endpoint)

  **WHY Each Reference Matters**:
  - `Dockerfile` — You need to reuse the Go build stage to compile binaries, NOT build from scratch
  - `Dockerfile.base` — This proves the Debian+build-essential pattern works for real compilation
  - Root `docker-compose.yml` — Shows the correct port mappings and service names the CLI expects
  - Coordinator main.go — Exact flag names for TLS/OTel overlay compose files

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Docker images build successfully
    Tool: Bash
    Preconditions: Docker Desktop running on macOS
    Steps:
      1. Run `docker compose -f test/e2e/docker-compose.yml build`
      2. Check exit code is 0
      3. Run `docker images | grep hg-e2e` to verify images exist
    Expected Result: Build completes with exit code 0, images appear in docker images output
    Failure Indicators: Build fails with missing package, Dockerfile syntax error, or Go compilation error
    Evidence: .sisyphus/evidence/task-1-docker-build.log

  Scenario: Docker compose config validates
    Tool: Bash
    Preconditions: test/e2e/docker-compose.yml exists
    Steps:
      1. Run `docker compose -f test/e2e/docker-compose.yml config`
      2. Run `docker compose -f test/e2e/docker-compose.yml -f test/e2e/docker-compose.tls.yml config`
      3. Run `docker compose -f test/e2e/docker-compose.yml -f test/e2e/docker-compose.otel.yml config`
    Expected Result: All three commands output valid YAML and exit 0
    Failure Indicators: YAML parse error, undefined service reference, invalid port mapping
    Evidence: .sisyphus/evidence/task-1-compose-config.log
  ```

  **Commit**: YES
  - Message: `test(e2e): add Docker infrastructure for E2E verification`
  - Files: `test/e2e/Dockerfile.worker`, `test/e2e/docker-compose.yml`, `test/e2e/docker-compose.tls.yml`, `test/e2e/docker-compose.otel.yml`, `.gitignore`

- [x] 2. Create Multi-File C Test Project

  **What to do**:
  - Create `test/e2e/testdata/` directory with a small multi-file C project:
    - `main.c` — calls functions from math_utils and string_utils, prints results, returns 0
    - `math_utils.c` / `math_utils.h` — 2-3 simple math functions (add, multiply, factorial)
    - `string_utils.c` / `string_utils.h` — 2-3 simple string functions (strlen, reverse, concat)
    - `Makefile` — builds all .c files into .o files, links into `testapp` binary
      - `CC ?= gcc` to allow override
      - `all` target: compile all .c → .o, link to `testapp`
      - `clean` target: remove .o and testapp
      - Uses `-Wall -Wextra -O2` flags
  - The project must have at least 3 separate .c files (to test parallel distributed compilation with -j)
  - Include a deliberate compile-error file: `test/e2e/testdata/bad.c` (syntax error) for error-path testing

  **Must NOT do**:
  - Do NOT use complex build systems (CMake, autotools)
  - Do NOT include external dependencies
  - Do NOT make the project so large it takes >10s to compile

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Simple file creation, no complex logic
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 1, 3)
  - **Parallel Group**: Wave 1
  - **Blocks**: Task 5
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - `test/stress/run-test.sh:25-40` — Shows how CPython stress test structures its build command with hgbuild make

  **WHY Each Reference Matters**:
  - `run-test.sh` — Shows the exact `hgbuild make -j` invocation pattern your Makefile must be compatible with

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: C project compiles locally with gcc
    Tool: Bash
    Preconditions: gcc installed on macOS (xcode command line tools)
    Steps:
      1. Run `make -C test/e2e/testdata/ clean all`
      2. Check exit code is 0
      3. Run `test/e2e/testdata/testapp` and verify it outputs expected text
      4. Run `make -C test/e2e/testdata/ clean`
    Expected Result: Compiles without warnings (with -Wall -Wextra), testapp runs and exits 0
    Failure Indicators: Compilation errors, undefined references, segfault on run
    Evidence: .sisyphus/evidence/task-2-local-compile.log

  Scenario: Bad.c fails to compile with clear error
    Tool: Bash
    Preconditions: gcc installed
    Steps:
      1. Run `gcc -c test/e2e/testdata/bad.c -o /dev/null 2>&1`
      2. Check exit code is non-zero
    Expected Result: gcc returns exit code 1 with syntax error message
    Evidence: .sisyphus/evidence/task-2-bad-compile.log
  ```

  **Commit**: YES (groups with Task 3)
  - Message: `test(e2e): add C test project and TLS cert generation`
  - Files: `test/e2e/testdata/*`, `test/e2e/gen-certs.sh`

- [x] 3. Generate Self-Signed TLS Certificates

  **What to do**:
  - Create `test/e2e/gen-certs.sh` script that generates:
    - CA certificate + key (`ca.crt`, `ca.key`)
    - Server certificate + key signed by CA (`server.crt`, `server.key`) — SAN: `localhost`, `coordinator`, `worker-1`, `worker-2`
    - Client certificate + key signed by CA (`client.crt`, `client.key`) — for mTLS testing
    - All certs output to `test/e2e/certs/` directory
    - 365-day validity, RSA 2048-bit
  - Script must be idempotent (removes existing certs before generating)
  - Add `test/e2e/certs/` to `.gitignore` (if not done in Task 1)

  **Must NOT do**:
  - Do NOT build a PKI with intermediate CAs
  - Do NOT use ECDSA (keep it simple with RSA)
  - Do NOT generate CRL or OCSP configuration

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Standard openssl commands, well-known pattern
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 1, 2)
  - **Parallel Group**: Wave 1
  - **Blocks**: Task 10
  - **Blocked By**: None

  **References**:

  **API/Type References**:
  - `internal/security/tls/config.go` — TLSConfig struct shows what fields are expected: CertFile, KeyFile, CAFile, RequireClientCert
  - `cmd/hg-coord/main.go:82-100` — TLS flag names: `--tls-cert`, `--tls-key`, `--tls-ca`, `--tls-require-client-cert`

  **WHY Each Reference Matters**:
  - `config.go` — Tells you the cert/key/CA file paths the system expects
  - `main.go` — Exact CLI flag names for mounting certs in Docker compose

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Cert generation script produces valid certificates
    Tool: Bash
    Preconditions: openssl installed (standard on macOS)
    Steps:
      1. Run `bash test/e2e/gen-certs.sh`
      2. Verify files exist: `ls test/e2e/certs/{ca,server,client}.{crt,key}` — 6 files
      3. Verify CA cert: `openssl x509 -in test/e2e/certs/ca.crt -noout -subject` — shows CA subject
      4. Verify server cert signed by CA: `openssl verify -CAfile test/e2e/certs/ca.crt test/e2e/certs/server.crt` — returns OK
      5. Verify client cert signed by CA: `openssl verify -CAfile test/e2e/certs/ca.crt test/e2e/certs/client.crt` — returns OK
      6. Verify server cert SAN includes localhost: `openssl x509 -in test/e2e/certs/server.crt -noout -text | grep -A1 "Subject Alternative Name"` — shows localhost, coordinator
    Expected Result: 6 cert/key files, all valid, server cert has correct SANs
    Failure Indicators: openssl errors, missing files, verification failure
    Evidence: .sisyphus/evidence/task-3-cert-generation.log

  Scenario: Cert generation is idempotent
    Tool: Bash
    Steps:
      1. Run `bash test/e2e/gen-certs.sh` (second time)
      2. Verify same 6 files exist and are valid
    Expected Result: Script completes without error, existing certs replaced
    Evidence: .sisyphus/evidence/task-3-idempotent.log
  ```

  **Commit**: YES (groups with Task 2)
  - Message: `test(e2e): add C test project and TLS cert generation`
  - Files: `test/e2e/gen-certs.sh`

- [x] 4. Verify Docker Cluster Startup + Worker Registration

  **What to do**:
  - Start the E2E Docker cluster: `docker compose -f test/e2e/docker-compose.yml up -d --build`
  - Wait for containers to be healthy (poll healthchecks, max 60s)
  - Verify coordinator is running: `curl http://localhost:8080/metrics` returns Prometheus text
  - Verify workers registered: `curl http://localhost:8080/api/v1/workers` returns JSON with 2 workers
  - Verify hgbuild can connect: run `hgbuild status --coordinator=localhost:9000` (or equivalent)
  - Verify hgbuild workers list: run `hgbuild workers --coordinator=localhost:9000` shows 2 workers
  - Capture docker compose logs for all services
  - Document any issues in `.sisyphus/findings.md`

  **Must NOT do**:
  - Do NOT proceed if cluster doesn't start — this blocks all subsequent tasks
  - Do NOT modify production code to fix startup issues — document as findings

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Requires Docker operations, HTTP verification, log analysis, and judgment on health
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO (gate for all subsequent tasks)
  - **Parallel Group**: Wave 2 (sequential within wave)
  - **Blocks**: Tasks 5, 6, 7, 8, 9, 10, 11, 12
  - **Blocked By**: Task 1

  **References**:

  **Pattern References**:
  - `test/e2e/docker-compose.yml` — Created in Task 1; service names, port mappings
  - `docker-compose.yml:25-28` — Root compose healthcheck pattern (use /metrics not /health for coordinator)

  **API/Type References**:
  - `internal/observability/dashboard/server.go:69-71` — API endpoints: `/api/v1/stats`, `/api/v1/workers`, `/api/v1/events`
  - `cmd/hgbuild/main.go:205-237` — `hgbuild status` command implementation
  - `cmd/hgbuild/main.go:239-288` — `hgbuild workers` command implementation

  **WHY Each Reference Matters**:
  - `docker-compose.yml` — Exact service names and ports to verify
  - `dashboard/server.go` — The API endpoints to curl for verification
  - `hgbuild/main.go` — The CLI commands available for cluster inspection

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Docker cluster starts and all containers are healthy
    Tool: Bash
    Preconditions: Docker Desktop running, Task 1 complete (images built)
    Steps:
      1. Run `docker compose -f test/e2e/docker-compose.yml up -d --build`
      2. Wait for healthy (portable — no GNU timeout on macOS):
         ```bash
         for i in $(seq 1 30); do
           COUNT=$(docker compose -f test/e2e/docker-compose.yml ps | grep -c "healthy" || true)
           [ "$COUNT" -ge 3 ] && break
           sleep 2
         done
         ```
         (Polls every 2s for up to 60s. On macOS, GNU `timeout` is not available by default.)
      3. Run `docker compose -f test/e2e/docker-compose.yml ps` and verify all 3 containers show "healthy"
    Expected Result: 3 containers (coordinator, worker-1, worker-2) all in "healthy" state within 60s
    Failure Indicators: Container exits with error, healthcheck fails, timeout exceeded
    Evidence: .sisyphus/evidence/task-4-cluster-startup.log

  Scenario: Workers registered with coordinator
    Tool: Bash (curl)
    Preconditions: Cluster running and healthy
    Steps:
      1. Run `curl -s http://localhost:8080/api/v1/workers | python3 -m json.tool`
      2. Verify response is a JSON object with a `workers` array field (NOT a bare array — see `internal/observability/dashboard/api.go:92`)
      3. Verify `.workers` array has at least 2 entries, each containing `id`, `address`, `architecture`, `healthy` fields (see `WorkerInfo` struct at `api.go:26-41`)
      4. Verify response also contains `count` field >= 2 and `timestamp` field
      5. Run `curl -s http://localhost:8080/api/v1/stats | python3 -m json.tool`
      6. Verify stats contain `total_workers` and `healthy_workers` fields (NOT `active_workers` — see `Stats` struct at `api.go:10-23`)
      7. Verify `total_workers` >= 2 and `healthy_workers` >= 2
    Expected Result: /api/v1/workers returns `{"workers": [...], "count": 2, "timestamp": ...}` with 2 worker entries. /api/v1/stats returns `{"total_workers": 2, "healthy_workers": 2, ...}`.
    Failure Indicators: Empty workers array, missing `workers` key, count < 2, connection refused, 404 response
    Evidence: .sisyphus/evidence/task-4-worker-registration.json

  Scenario: hgbuild CLI connects to cluster
    Tool: Bash
    Preconditions: Cluster running, hgbuild binary built (make build)
    Steps:
      1. Run `make build` to ensure hgbuild binary exists
      2. Run `./bin/hgbuild status --coordinator=localhost:9000`
      3. Verify output shows coordinator status (active tasks, queued tasks)
      4. Run `./bin/hgbuild workers --coordinator=localhost:9000`
      5. Verify output lists 2 workers with capabilities
    Expected Result: status shows coordinator info, workers lists 2 workers with C/C++ capabilities
    Failure Indicators: Connection refused, timeout, empty worker list
    Evidence: .sisyphus/evidence/task-4-cli-connect.log
  ```

  **Commit**: NO (evidence only — committed in batch)

- [x] 5. Verify Compilation Pipeline (hgbuild cc + make)

  **What to do**:
  - **Single file compilation**: Copy `test/e2e/testdata/main.c` to a temp directory, compile with `HG_COORDINATOR=localhost:9000 ./bin/hgbuild cc -v -c main.c -o main.o`
  - **CRITICAL**: The `cc` and `c++` subcommands use `DisableFlagParsing: true` (`cmd/hgbuild/main.go:724`), which means cobra does NOT parse flags after the subcommand. `--coordinator`, `--tls-*`, and other hgbuild flags placed AFTER `cc` are stripped from gcc args by `filterHgbuildFlags()` but NOT set as variable values. Use either: (a) `HG_COORDINATOR` env var, or (b) `./bin/hgbuild --coordinator=localhost:9000 cc ...` with flags BEFORE the subcommand.
  - Verify `main.o` was created and is a valid object file
  - **Multi-file build**: Run `HG_COORDINATOR=localhost:9000 ./bin/hgbuild -v make -C test/e2e/testdata/ -j4` (the `make` subcommand also has `DisableFlagParsing: true`, so use env var)
  - Verify all .o files created and `testapp` binary links successfully
  - Verify compilation happened remotely (check coordinator logs for task assignment)
  - **Error path**: Compile `bad.c` and verify error message is returned
  - Capture verbose output with `-v` flag to see `[remote]`/`[cache]` tags
  - Leave the cluster running for subsequent tasks

  **Must NOT do**:
  - Do NOT stop/restart the Docker cluster
  - Do NOT modify the test C files

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Multi-step verification with pipeline tracing, log correlation, and error-path testing
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO (sequential after Task 4)
  - **Parallel Group**: Wave 2
  - **Blocks**: Tasks 6, 9
  - **Blocked By**: Tasks 1, 2, 4

  **References**:

  **Pattern References**:
  - `test/e2e/testdata/Makefile` — Created in Task 2; the Makefile to build
  - `test/stress/run-test.sh:25-40` — Shows `hgbuild make` invocation pattern
  - `internal/cli/build/build.go:186-200` — Build flow: preprocess → send to coordinator → receive object

  **API/Type References**:
  - `cmd/hgbuild/main.go:714-748` — `cc` and `c++` command definitions
  - `cmd/hgbuild/main.go:1081-1114` — `make` and `ninja` command definitions
  - `cmd/hgbuild/main.go:880` — Where `svc.Build(ctx, req)` is called

  **WHY Each Reference Matters**:
  - `build.go` — Understanding the compile flow helps debug if compilation fails
  - `main.go cc/make` — The exact CLI interface being tested

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Single file compiles via coordinator
    Tool: Bash
    Preconditions: Cluster running (Task 4), hgbuild built, test/e2e/testdata exists
    Steps:
      1. Create temp dir: `TMPDIR=$(mktemp -d)`
      2. Copy: `cp test/e2e/testdata/main.c test/e2e/testdata/math_utils.h test/e2e/testdata/string_utils.h $TMPDIR/`
      3. Run: `HG_COORDINATOR=localhost:9000 ./bin/hgbuild cc -v -c $TMPDIR/main.c -o $TMPDIR/main.o 2>&1`
         (IMPORTANT: use HG_COORDINATOR env var, NOT --coordinator after cc — see DisableFlagParsing note)
      4. Verify `$TMPDIR/main.o` exists: `test -f $TMPDIR/main.o`
      5. Verify it's a valid object file: `file $TMPDIR/main.o` contains "object" or "Mach-O"
      6. Verify verbose output contains `[remote]` or `[cache]` tag
    Expected Result: main.o created, valid object file, verbose output shows remote compilation
    Failure Indicators: "no workers available", connection refused, empty object file
    Evidence: .sisyphus/evidence/task-5-single-compile.log

  Scenario: Multi-file project builds with hgbuild make
    Tool: Bash
    Preconditions: Cluster running, test project exists
    Steps:
      1. Run `make -C test/e2e/testdata/ clean` to start fresh
      2. Run `HG_COORDINATOR=localhost:9000 ./bin/hgbuild -v make -C test/e2e/testdata/ -j4 2>&1`
         (IMPORTANT: use HG_COORDINATOR env var, NOT --coordinator after make)
      3. Verify all .o files exist: `ls test/e2e/testdata/*.o | wc -l` — should be 3 (main, math_utils, string_utils)
      4. Verify binary exists: `test -f test/e2e/testdata/testapp`
      5. Run the binary: `test/e2e/testdata/testapp` — exits 0
    Expected Result: All 3 .c files compiled, linked into testapp, testapp runs and exits 0
    Failure Indicators: Compilation error, linker error, binary crashes
    Evidence: .sisyphus/evidence/task-5-make-build.log

  Scenario: Compilation error is reported correctly
    Tool: Bash
    Steps:
      1. Run `HG_COORDINATOR=localhost:9000 ./bin/hgbuild cc -v -c test/e2e/testdata/bad.c -o /tmp/bad.o 2>&1`
      2. Verify exit code is non-zero
      3. Verify stderr/output contains syntax error message from compiler
    Expected Result: Exit code ≠ 0, error output includes gcc error message about syntax error
    Failure Indicators: Exit code 0 (false success), no error message, crash
    Evidence: .sisyphus/evidence/task-5-compile-error.log
  ```

  **Commit**: YES (groups with Tasks 6, 7, 8)
  - Message: `test(e2e): verify compilation pipeline + cache + dashboard`
  - Files: `.sisyphus/evidence/task-4-*`, `.sisyphus/evidence/task-5-*`

- [x] 6. Verify Cache Hit/Miss Behavior

  **What to do**:
  - After Task 5's compilation, the cache should contain entries
  - Scrape Prometheus metrics BEFORE second build: `curl http://localhost:8080/metrics | grep hybridgrid_cache`
  - Clean only the object files (not cache): `make -C test/e2e/testdata/ clean`
  - Rebuild: `HG_COORDINATOR=localhost:9000 ./bin/hgbuild -v make -C test/e2e/testdata/ -j4`
  - Scrape metrics AFTER second build — cache_hits should have increased
  - Verify verbose output shows `[cache]` tags instead of `[remote]`
  - Compare first-build vs second-build timing
  - Verify cache directory has files: `ls ~/.hybridgrid/cache/`

  **Must NOT do**:
  - Do NOT clear the cache between first and second build
  - Do NOT restart the cluster

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Requires metrics scraping, log analysis, and timing comparison
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 7, 8 — but only after Task 5)
  - **Parallel Group**: Wave 3
  - **Blocks**: None
  - **Blocked By**: Task 5

  **References**:

  **Pattern References**:
  - `internal/cache/store.go:104-136` — Cache Get() with TTL check; shows how cache hits work
  - `internal/cli/build/build.go:157-180` — Where cache is checked before remote compile

  **API/Type References**:
  - `internal/coordinator/metrics.go` or `internal/observability/` — Prometheus metric names (search for `cache_hits`, `cache_misses`)

  **WHY Each Reference Matters**:
  - `store.go` — Understanding cache key format helps debug cache misses
  - `build.go` — The cache is checked CLIENT-SIDE (in hgbuild), not server-side

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Second build hits cache
    Tool: Bash
    Preconditions: Task 5 completed (first build done, cache populated)
    Steps:
      1. Record pre-build metrics: `curl -s http://localhost:8080/metrics | grep hybridgrid_cache > /tmp/metrics-before.txt`
      2. Clean objects: `make -C test/e2e/testdata/ clean`
      3. Rebuild: `HG_COORDINATOR=localhost:9000 ./bin/hgbuild -v make -C test/e2e/testdata/ -j4 2>&1 | tee /tmp/cache-build.log`
      4. Record post-build metrics: `curl -s http://localhost:8080/metrics | grep hybridgrid_cache > /tmp/metrics-after.txt`
      5. Verify verbose output contains `[cache]` tags: `grep -c '\[cache\]' /tmp/cache-build.log`
      6. Verify cache directory has content: `ls ~/.hybridgrid/cache/ | wc -l`
    Expected Result: At least 1 `[cache]` tag in verbose output. Cache hits metric increased. Build time significantly faster than first build.
    Failure Indicators: All `[remote]` (no cache hits), metrics unchanged, cache dir empty
    Evidence: .sisyphus/evidence/task-6-cache-hits.log

  Scenario: Cache directory contains entries
    Tool: Bash
    Steps:
      1. Run `ls -la ~/.hybridgrid/cache/ | head -20`
      2. Verify at least 3 cache entries (one per .c file)
    Expected Result: Cache directory contains files corresponding to compiled sources
    Evidence: .sisyphus/evidence/task-6-cache-dir.log
  ```

  **Commit**: YES (groups with Tasks 4, 5, 7, 8)
  - Message: `test(e2e): verify compilation pipeline + cache + dashboard`

 - [x] 7. Verify Dashboard API + Prometheus Metrics

  **What to do**:
  - Test all dashboard API endpoints:
    - `GET /api/v1/stats` — verify JSON response with task counts
    - `GET /api/v1/workers` — verify JSON response with worker list and capabilities
    - `GET /api/v1/events` — verify JSON response (may be empty if no SSE)
  - Test Prometheus metrics endpoint:
    - `GET /metrics` — verify response contains `hybridgrid_` prefixed metrics
    - Verify specific metrics exist: `hybridgrid_tasks_total`, `hybridgrid_task_duration_seconds`, `hybridgrid_cache_hits_total`, `hybridgrid_cache_misses_total`, `hybridgrid_workers_total`
    - Verify task counts are > 0 (from Task 5 compilations)
  - Test WebSocket endpoint:
    - Use `curl` or `websocat` to connect to `ws://localhost:8080/ws` and verify connection accepted
  - Test static assets:
    - `GET /` — verify HTML dashboard page is returned

  **Must NOT do**:
  - Do NOT test dashboard UI rendering (no Playwright for embedded HTML)
  - Do NOT restart the cluster

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Multiple HTTP endpoints to verify with structured assertions
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 6, 8)
  - **Parallel Group**: Wave 3
  - **Blocks**: None
  - **Blocked By**: Task 4

  **References**:

  **Pattern References**:
  - `internal/observability/dashboard/server.go:60-78` — All registered routes and their handlers

  **API/Type References**:
  - `internal/observability/dashboard/server.go:81-110` — Stats handler returns StatsResponse struct
  - `internal/observability/dashboard/server.go:112-140` — Workers handler returns worker list
  - `internal/coordinator/metrics.go` — Prometheus metric definitions

  **WHY Each Reference Matters**:
  - `server.go` — Exact endpoint paths and expected response shapes
  - `metrics.go` — Exact Prometheus metric names to grep for

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Dashboard API returns valid stats
    Tool: Bash (curl)
    Preconditions: Cluster running, at least 1 compilation completed (Task 5)
    Steps:
      1. Run `curl -s http://localhost:8080/api/v1/stats | python3 -m json.tool`
      2. Verify JSON response is valid and contains fields from `Stats` struct (`api.go:10-23`):
         - `total_tasks` (int, should be > 0 after compilations)
         - `success_tasks`, `failed_tasks`, `active_tasks`, `queued_tasks` (ints)
         - `cache_hits`, `cache_misses` (ints)
         - `cache_hit_rate` (float)
         - `total_workers`, `healthy_workers` (ints, both >= 2)
         - `uptime_seconds`, `timestamp` (ints)
      3. Verify `total_tasks` > 0 (compilations from Task 5 should be counted)
      4. Run `curl -s http://localhost:8080/api/v1/workers | python3 -m json.tool`
      5. Verify response is a JSON object with `workers` array (NOT a bare array — see `api.go:92`)
      6. Verify `.workers` array has >= 2 entries, each with `id`, `address`, `healthy`, `architecture` fields
      7. Verify response has `count` >= 2
    Expected Result: Stats endpoint returns valid JSON with all `Stats` fields, `total_tasks` > 0. Workers endpoint returns `{"workers": [...], "count": 2, ...}`.
    Failure Indicators: 404, invalid JSON, empty response, `total_tasks` still 0, missing `workers` key
    Evidence: .sisyphus/evidence/task-7-dashboard-api.json

  Scenario: Prometheus metrics contain hybridgrid counters
    Tool: Bash (curl + grep)
    Preconditions: Cluster running with completed compilations
    Steps:
      1. Run `curl -s http://localhost:8080/metrics`
      2. Grep for `hybridgrid_tasks_total` — verify exists and value > 0
      3. Grep for `hybridgrid_workers_total` — verify exists
      4. Grep for `hybridgrid_cache` — verify cache metrics exist
      5. Save full metrics output
    Expected Result: At least 3 `hybridgrid_` prefixed metric families present, task counter > 0
    Failure Indicators: No hybridgrid_ metrics, all counters at 0
    Evidence: .sisyphus/evidence/task-7-prometheus-metrics.txt

  Scenario: Dashboard HTML page loads
    Tool: Bash (curl)
    Steps:
      1. Run `curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/`
      2. Verify HTTP status 200
      3. Run `curl -s http://localhost:8080/ | head -5`
      4. Verify response contains `<html` or `<!DOCTYPE`
    Expected Result: Root path returns HTTP 200 with HTML content
    Evidence: .sisyphus/evidence/task-7-dashboard-html.log
  ```

  **Commit**: YES (groups with Tasks 4, 5, 6, 8)

- [x] 8. Verify Log-Level Endpoint

  **What to do**:
  - Test coordinator log-level endpoint at `http://localhost:8080/log-level`:
    - `GET` — verify returns current level (should be "info" by default)
    - `PUT` with `{"level":"debug"}` — verify 200 response
    - `GET` again — verify returns "debug"
    - `PUT` with `{"level":"info"}` — restore to original
  - Test worker log-level endpoint at `http://localhost:9091/log-level` (worker-1's HTTP port 9090 published to host port 9091 — see Task 1 compose):
    - Same GET/PUT cycle
  - Test invalid level:
    - `PUT` with `{"level":"invalid"}` — verify error response (400)

  **Must NOT do**:
  - Do NOT leave log level at debug (restore to info after test)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Simple HTTP endpoint testing with curl
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 6, 7)
  - **Parallel Group**: Wave 3
  - **Blocks**: None
  - **Blocked By**: Task 4

  **References**:

  **API/Type References**:
  - `internal/logging/handler.go:62-116` — GET returns `{"level":"info"}`, PUT accepts `{"level":"debug"}`, validates against allowed levels

  **WHY Each Reference Matters**:
  - `handler.go` — Exact request/response format and valid level values

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Log level GET/PUT cycle on coordinator
    Tool: Bash (curl)
    Preconditions: Cluster running
    Steps:
      1. GET: `curl -s http://localhost:8080/log-level` — capture response
      2. Verify response contains "level" field (e.g., `{"level":"info"}`)
      3. PUT: `curl -s -X PUT -H 'Content-Type: application/json' -d '{"level":"debug"}' http://localhost:8080/log-level`
      4. Verify PUT returns 200
      5. GET: `curl -s http://localhost:8080/log-level` — verify returns `{"level":"debug"}`
      6. Restore: `curl -s -X PUT -H 'Content-Type: application/json' -d '{"level":"info"}' http://localhost:8080/log-level`
    Expected Result: GET returns JSON with level. PUT changes level. Second GET confirms change. Restore works.
    Failure Indicators: 404, 500, level doesn't change
    Evidence: .sisyphus/evidence/task-8-loglevel-coord.log

  Scenario: Invalid log level returns error
    Tool: Bash (curl)
    Steps:
      1. PUT: `curl -s -w '\n%{http_code}' -X PUT -H 'Content-Type: application/json' -d '{"level":"invalid"}' http://localhost:8080/log-level`
      2. Verify HTTP status is 400
    Expected Result: HTTP 400 Bad Request with error message
    Evidence: .sisyphus/evidence/task-8-loglevel-invalid.log

  Scenario: Log level GET/PUT cycle on worker
    Tool: Bash (curl)
    Preconditions: Cluster running, worker-1 HTTP port published on host port 9091 (see Task 1 compose)
    Steps:
      1. GET: `curl -s http://localhost:9091/log-level` — capture response
      2. Verify response contains "level" field (e.g., `{"level":"info"}`)
      3. PUT: `curl -s -X PUT -H 'Content-Type: application/json' -d '{"level":"debug"}' http://localhost:9091/log-level`
      4. Verify PUT returns 200
      5. GET: `curl -s http://localhost:9091/log-level` — verify returns `{"level":"debug"}`
      6. Restore: `curl -s -X PUT -H 'Content-Type: application/json' -d '{"level":"info"}' http://localhost:9091/log-level`
    Expected Result: Worker log level changes successfully via HTTP API on host port 9091
    Failure Indicators: Connection refused (port not published), 404 (endpoint not registered on worker)
    Evidence: .sisyphus/evidence/task-8-loglevel-worker.log
  ```

  **Commit**: YES (groups with Tasks 4, 5, 6, 7)

- [x] 9. Verify Local Fallback Compilation

  **What to do**:
  - Stop the Docker cluster: `docker compose -f test/e2e/docker-compose.yml down`
  - Verify coordinator is unreachable: `curl http://localhost:8080/metrics` should fail
  - Clean test project objects: `make -C test/e2e/testdata/ clean`
  - Attempt compilation with hgbuild: `HG_COORDINATOR=localhost:9000 ./bin/hgbuild -v cc -c test/e2e/testdata/main.c -o /tmp/fallback-main.o`
  - Verify compilation SUCCEEDS (local fallback kicked in)
  - Verify verbose output contains `[local]` or fallback indicator
  - Verify the object file is valid
  - Restart the Docker cluster for subsequent tasks: `docker compose -f test/e2e/docker-compose.yml up -d`
  - Wait for cluster to be healthy again

  **Must NOT do**:
  - Do NOT test `--no-fallback` (flag doesn't exist)
  - Do NOT modify production code

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Requires cluster lifecycle management, fallback verification, and cluster restart
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO (requires stopping/starting the cluster, conflicts with Tasks 6, 7, 8)
  - **Parallel Group**: Wave 3 (but must wait for Tasks 6, 7, 8 to finish since they need the running cluster)
  - **Blocks**: None
  - **Blocked By**: Tasks 5, 6, 7, 8

  **References**:

  **Pattern References**:
  - `internal/cli/build/build.go:50-58` — `FallbackEnabled: true` default; shows fallback is always on
  - `internal/cli/build/build.go:140-155` — Fallback logic: if remote compile fails, try local gcc

  **WHY Each Reference Matters**:
  - `build.go` — Understanding the fallback flow helps verify the right code path was taken

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Compilation falls back to local when coordinator is down
    Tool: Bash
    Preconditions: Cluster stopped, gcc installed locally
    Steps:
      1. Stop cluster: `docker compose -f test/e2e/docker-compose.yml down`
      2. Verify coordinator down: `curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/metrics 2>/dev/null || echo "UNREACHABLE"`
      3. Clean: `rm -f /tmp/fallback-main.o`
      4. Compile: `HG_COORDINATOR=localhost:9000 ./bin/hgbuild -v cc -c test/e2e/testdata/main.c -o /tmp/fallback-main.o 2>&1 | tee /tmp/fallback.log`
      5. Verify object created: `test -f /tmp/fallback-main.o && echo "EXISTS"`
      6. Verify fallback in output: `grep -i -E '(fallback|local|falling back)' /tmp/fallback.log`
    Expected Result: Compilation succeeds with local fallback. Output indicates fallback mode. Valid object file created.
    Failure Indicators: Compilation fails entirely, hangs forever, no fallback message
    Evidence: .sisyphus/evidence/task-9-local-fallback.log

  Scenario: Cluster restarts after fallback test
    Tool: Bash
    Steps:
      1. Start: `docker compose -f test/e2e/docker-compose.yml up -d`
      2. Wait for healthy: poll for 60s
      3. Verify: `curl -s http://localhost:8080/api/v1/workers | python3 -m json.tool`
    Expected Result: Cluster back to healthy with 2 workers registered
    Evidence: .sisyphus/evidence/task-9-cluster-restart.log
  ```

  **Commit**: YES (groups with Tasks 10, 11, 12)
  - Message: `test(e2e): verify fallback + TLS + OTel + stress test`

- [x] 10. Verify TLS/mTLS Secured Compilation

  **What to do**:
  - Generate TLS certs if not done: `bash test/e2e/gen-certs.sh`
  - Stop the plain cluster: `docker compose -f test/e2e/docker-compose.yml down`
  - Start cluster with TLS overlay: `docker compose -f test/e2e/docker-compose.yml -f test/e2e/docker-compose.tls.yml up -d --build`
  - Wait for healthy
  - Verify compilation works with TLS:
    - `./bin/hgbuild --coordinator=localhost:9000 --tls-cert=test/e2e/certs/client.crt --tls-key=test/e2e/certs/client.key --tls-ca=test/e2e/certs/ca.crt --insecure=false cc -v -c test/e2e/testdata/main.c -o /tmp/tls-main.o`
    - **CRITICAL**: ALL hgbuild flags (`--coordinator`, `--tls-*`, `--insecure`) MUST go BEFORE the `cc` subcommand. The `cc` command uses `DisableFlagParsing: true` so flags AFTER `cc` are passed to gcc and will break compilation. The `--insecure=false` flag is required because `insecure` defaults to `true` (`cmd/hgbuild/main.go:161`).
  - Verify compilation FAILS without TLS certs:
    - `./bin/hgbuild --coordinator=localhost:9000 --insecure=false cc -c test/e2e/testdata/main.c -o /tmp/notls-main.o` — should fail with TLS handshake error
  - Stop TLS cluster, restart plain cluster for subsequent tasks

  **Must NOT do**:
  - Do NOT create permanent TLS infrastructure
  - Do NOT test certificate rotation or expiry

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: TLS setup is complex — cert mounting, flag wiring, error path verification, cluster lifecycle
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO (requires cluster restart with TLS, conflicts with other cluster tasks)
  - **Parallel Group**: Wave 3 (after Tasks 6, 7, 8 complete)
  - **Blocks**: Task 11
  - **Blocked By**: Tasks 3, 4, 9

  **References**:

  **Pattern References**:
  - `test/e2e/docker-compose.tls.yml` — Created in Task 1; TLS overlay config
  - `test/e2e/gen-certs.sh` — Created in Task 3; cert generation

  **API/Type References**:
  - `internal/security/tls/config.go` — TLSConfig struct fields
  - `cmd/hg-coord/main.go:82-100` — TLS flags: --tls-cert, --tls-key, --tls-ca, --tls-require-client-cert
  - `cmd/hgbuild/main.go` — Search for TLS-related flags for hgbuild CLI

  **WHY Each Reference Matters**:
  - `config.go` — Required cert paths the system expects
  - `main.go` — Exact flag names for both server and client sides

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: TLS-secured compilation succeeds with valid certs
    Tool: Bash
    Preconditions: Certs generated (Task 3), TLS cluster running
    Steps:
      1. Stop plain cluster: `docker compose -f test/e2e/docker-compose.yml down`
      2. Start TLS cluster: `docker compose -f test/e2e/docker-compose.yml -f test/e2e/docker-compose.tls.yml up -d --build`
      3. Wait for healthy (60s timeout)
      4. Compile with TLS (flags BEFORE cc): `./bin/hgbuild --coordinator=localhost:9000 --tls-cert=test/e2e/certs/client.crt --tls-key=test/e2e/certs/client.key --tls-ca=test/e2e/certs/ca.crt --insecure=false cc -v -c test/e2e/testdata/main.c -o /tmp/tls-main.o 2>&1`
         (CRITICAL: all --tls-* and --coordinator flags go BEFORE `cc` due to DisableFlagParsing)
      5. Verify `/tmp/tls-main.o` exists and is valid object file
    Expected Result: Compilation succeeds over TLS, valid object file produced
    Failure Indicators: TLS handshake error, connection refused, cert validation failure
    Evidence: .sisyphus/evidence/task-10-tls-compile.log

  Scenario: mTLS enforcement verified — server rejects connections without client certificates
    Tool: Bash
    Preconditions: TLS cluster running with --tls-require-client-cert on coordinator (port 9000)
    Steps:
      1. Probe coordinator gRPC port with openssl WITHOUT client cert:
         `echo | openssl s_client -connect localhost:9000 -CAfile test/e2e/certs/ca.crt 2>&1 | tee /tmp/tls-no-client-cert.log`
         (This connects with TLS but presents NO client certificate)
      2. Verify the connection was rejected — look for handshake failure or alert:
         `grep -iE '(alert|handshake failure|error|certificate required|ssl_error|verify return)' /tmp/tls-no-client-cert.log`
      3. Now probe WITH valid client cert to confirm the server is actually running:
         `echo | openssl s_client -connect localhost:9000 -CAfile test/e2e/certs/ca.crt -cert test/e2e/certs/client.crt -key test/e2e/certs/client.key 2>&1 | tee /tmp/tls-with-client-cert.log`
      4. Verify the WITH-cert connection succeeds (TLS handshake completes):
         `grep -i 'verify return:1' /tmp/tls-with-client-cert.log`
      5. Compare: step 2 should show rejection, step 4 should show success — proving mTLS enforcement.
         Save both outputs as evidence.
    Expected Result: Without client cert → TLS handshake rejected (alert/error). With client cert → TLS handshake succeeds (verify return:1). This proves the coordinator enforces mTLS.
    Failure Indicators: Both connections succeed (mTLS not enforced) OR both fail (TLS misconfigured)
    Evidence: .sisyphus/evidence/task-10-tls-reject.log
  ```

  **Commit**: YES (groups with Tasks 9, 11, 12)

- [x] 11. Verify OTel Tracing with Collector

  **What to do**:
  - Stop any running cluster
  - Start cluster with OTel overlay: `docker compose -f test/e2e/docker-compose.yml -f test/e2e/docker-compose.otel.yml up -d --build`
  - Wait for all services healthy (including Jaeger)
  - Verify Jaeger UI is accessible: `curl http://localhost:16686/`
  - Clean test project and compile: `HG_COORDINATOR=localhost:9000 ./bin/hgbuild -v make -C test/e2e/testdata/ -j4`
    (CRITICAL: `make` subcommand has DisableFlagParsing — use HG_COORDINATOR env var, NOT --coordinator after `make`)
  - Wait a few seconds for traces to flush
  - Query Jaeger API for traces: `curl 'http://localhost:16686/api/traces?service=hybridgrid-coordinator&limit=5'`
    (Default service names: `hybridgrid-coordinator` from `cmd/hg-coord/main.go:249`, `hybridgrid-worker` from `cmd/hg-worker/main.go:343`)
  - Verify traces exist with spans for the compilation flow
  - Stop the OTel cluster, restart plain cluster

  **Must NOT do**:
  - Do NOT configure custom trace sampling — use defaults
  - Do NOT set up persistent trace storage — in-memory is fine for testing

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: OTel + Jaeger setup, trace API querying, span verification
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO (requires cluster restart with OTel overlay)
  - **Parallel Group**: Wave 4
  - **Blocks**: None
  - **Blocked By**: Task 10

  **References**:

  **Pattern References**:
  - `test/e2e/docker-compose.otel.yml` — Created in Task 1; OTel overlay with Jaeger
  - `internal/observability/tracing/tracer.go:44-129` — Tracer Init() function, OTLP exporter setup

  **API/Type References**:
  - `cmd/hg-coord/main.go:103-133` — Tracing flags: --tracing-enable, --tracing-endpoint, --tracing-service-name
  - `internal/observability/tracing/grpc.go` — gRPC interceptors that create spans

  **WHY Each Reference Matters**:
  - `tracer.go` — Shows the OTLP exporter endpoint format (gRPC to collector:4317)
  - `main.go` — Exact flag names for enabling tracing
  - `grpc.go` — Proves spans are created for each gRPC call (what to look for in Jaeger)

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: OTel traces appear in Jaeger after compilation
    Tool: Bash (curl)
    Preconditions: OTel cluster running with Jaeger
    Steps:
      1. Stop any running cluster: `docker compose -f test/e2e/docker-compose.yml down 2>/dev/null`
      2. Start OTel cluster: `docker compose -f test/e2e/docker-compose.yml -f test/e2e/docker-compose.otel.yml up -d --build`
      3. Wait for healthy (90s timeout — Jaeger takes longer)
      4. Verify Jaeger UI: `curl -s -o /dev/null -w '%{http_code}' http://localhost:16686/` — expect 200
      5. Clean and compile: `make -C test/e2e/testdata/ clean && HG_COORDINATOR=localhost:9000 ./bin/hgbuild -v make -C test/e2e/testdata/ -j4 2>&1`
         (CRITICAL: `make` has DisableFlagParsing — must use HG_COORDINATOR env var)
      6. Wait 5s for trace flush: `sleep 5`
      7. Query Jaeger services: `curl -s http://localhost:16686/api/services | python3 -m json.tool`
      8. Verify `hybridgrid-coordinator` and/or `hybridgrid-worker` appear in service list
         (Default service names: `hybridgrid-coordinator` from `cmd/hg-coord/main.go:249`, `hybridgrid-worker` from `cmd/hg-worker/main.go:343`)
      9. Query traces: `curl -s 'http://localhost:16686/api/traces?service=hybridgrid-coordinator&limit=5' | python3 -m json.tool`
      10. Verify at least 1 trace with spans
    Expected Result: Jaeger UI accessible, `hybridgrid-coordinator` and/or `hybridgrid-worker` services registered, traces with compilation spans present
    Failure Indicators: Jaeger returns 404, no services listed, no traces after compilation
    Evidence: .sisyphus/evidence/task-11-otel-traces.json

  Scenario: OTel cleanup — restart plain cluster
    Tool: Bash
    Steps:
      1. Stop: `docker compose -f test/e2e/docker-compose.yml -f test/e2e/docker-compose.otel.yml down -v`
      2. Start plain: `docker compose -f test/e2e/docker-compose.yml up -d`
      3. Wait for healthy
    Expected Result: Plain cluster running again
    Evidence: .sisyphus/evidence/task-11-cleanup.log
  ```

  **Commit**: YES (groups with Tasks 9, 10, 12)

- [x] 12. Run CPython Stress Test

  **What to do**:
  - Use the existing stress test infrastructure at `test/stress/`
  - Run the stress test: `cd test/stress && ./start.sh` OR manually:
    1. `docker compose -f test/stress/docker-compose.yml up -d --build`
    2. Wait for services to be healthy
    3. Execute the CPython build test inside the builder container
  - Monitor the build progress via coordinator dashboard: `curl http://localhost:8080/api/v1/stats`
  - Capture:
    - Total build time (distributed)
    - Number of files compiled
    - Cache hit rate after second build (if applicable)
    - Worker distribution (which workers got how many tasks)
  - Clean up stress test containers: `docker compose -f test/stress/docker-compose.yml down -v`
  - Restart the plain E2E cluster if needed for final wave

  **Must NOT do**:
  - Do NOT modify `test/stress/` files
  - Do NOT tune CPython build configuration
  - Do NOT fail the task if stress test takes longer than expected — document timing

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Long-running Docker-based test with monitoring and evidence capture
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO (test/stress/docker-compose.yml binds ports 9000:9000 and 8080:8080, same as E2E cluster — port conflict)
  - **Parallel Group**: Wave 4 (sequential AFTER Task 11)
  - **Blocks**: None
  - **Blocked By**: Task 4, Task 11 (must stop OTel cluster and free ports before stress test)

  **References**:

  **Pattern References**:
  - `test/stress/start.sh` — Automated stress test runner script
  - `test/stress/run-test.sh` — CPython build test: local → distributed → cache
  - `test/stress/docker-compose.yml` — 1 coordinator + 5 workers + builder
  - `test/stress/Dockerfile.base` — Debian image with build tools

  **WHY Each Reference Matters**:
  - `start.sh` — The entry point; may need to be run as-is or adapted if paths are wrong
  - `run-test.sh` — Shows the actual CPython download and build steps
  - `docker-compose.yml` — Different port mappings and resource limits from E2E compose

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: CPython stress test completes with distributed compilation
    Tool: Bash
    Preconditions: Docker Desktop running, sufficient disk space (~2GB for CPython source + build)
    Steps:
      1. Stop ALL running clusters (E2E and OTel): `docker compose -f test/e2e/docker-compose.yml down 2>/dev/null; docker compose -f test/e2e/docker-compose.yml -f test/e2e/docker-compose.otel.yml down 2>/dev/null`
         (CRITICAL: Task 11 may leave OTel cluster or plain cluster running — free ports 9000/8080 before stress test)
      2. Navigate: work from test/stress/ directory
      3. Build and start: `docker compose -f test/stress/docker-compose.yml up -d --build`
      4. Wait for readiness (120s — NOTE: `test/stress/docker-compose.yml` has a broken healthcheck using `/health` which doesn't exist on coordinator; do NOT rely on Docker health status. Instead poll `/metrics` directly):
         ```bash
         for i in $(seq 1 60); do
           HTTP=$(curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/metrics 2>/dev/null || true)
           [ "$HTTP" = "200" ] && break
           sleep 2
         done
         ```
         Verify coordinator responds, then check workers: `curl -s http://localhost:8080/api/v1/workers | python3 -c "import sys,json; d=json.load(sys.stdin); assert d['count']>=1, f'only {d[\"count\"]} workers'"`
      5. Execute stress test or follow start.sh steps
      6. Monitor: periodically `curl http://localhost:8080/api/v1/stats`
      7. Capture final stats after completion
      8. Record total build time
    Expected Result: CPython builds successfully with distributed compilation. Multiple workers receive tasks. Build completes within 30 minutes.
    Failure Indicators: Build fails with compiler errors, workers crash, coordinator OOM
    Evidence: .sisyphus/evidence/task-12-stress-test.log

  Scenario: Stress test cleanup
    Tool: Bash
    Steps:
      1. Stop: `docker compose -f test/stress/docker-compose.yml down -v --remove-orphans`
      2. Verify no orphaned containers: `docker ps -a | grep -c stress` — should be 0
    Expected Result: All stress test containers removed
    Evidence: .sisyphus/evidence/task-12-cleanup.log
  ```

  **Commit**: YES (groups with Tasks 9, 10, 11)

---

## Final Verification Wave (MANDATORY — after ALL implementation tasks)

> 4 review agents run in PARALLEL. ALL must APPROVE. Rejection → fix → re-run.

- [x] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify evidence exists in `.sisyphus/evidence/`. For each "Must NOT Have": verify no production code was modified (`git diff --name-only` shows only `test/e2e/` and `.sisyphus/` files). Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

  **QA Scenarios:**

  ```
  Scenario: All Must Have features have evidence
    Tool: Bash
    Steps:
      1. Read the plan's "Must Have" section and extract each requirement
      2. For each Must Have, search `.sisyphus/evidence/` for corresponding evidence:
         `ls .sisyphus/evidence/ | grep -i '<feature-keyword>'`
      3. Count total Must Haves vs those with evidence
    Expected Result: Every Must Have has at least one non-empty evidence file
    Evidence: .sisyphus/evidence/f1-compliance-audit.log

  Scenario: No production code was modified
    Tool: Bash
    Steps:
      1. Run: `git diff --name-only HEAD`
      2. Filter: `git diff --name-only HEAD | grep -v -E '^(test/e2e/|\.sisyphus/|\.gitignore)' | wc -l`
      3. Assert the count is 0
    Expected Result: Zero files outside test/e2e/, .sisyphus/, .gitignore were modified
    Evidence: .sisyphus/evidence/f1-no-prod-changes.log
  ```

- [x] F2. **Evidence Completeness Check** — `unspecified-high`
  For every task (1-12), verify that `.sisyphus/evidence/task-{N}-*.{ext}` files exist and are non-empty. Check that each evidence file corresponds to a QA scenario in the plan. Flag missing evidence.
  Output: `Evidence Files [N/N] | Coverage [N/N tasks] | VERDICT`

  **QA Scenarios:**

  ```
  Scenario: Every task has at least one evidence file
    Tool: Bash
    Steps:
      1. For each task N (1-12), check: `ls .sisyphus/evidence/task-${N}-* 2>/dev/null | wc -l`
      2. Verify count >= 1 for every task
      3. Check all evidence files are non-empty: `find .sisyphus/evidence/ -name 'task-*' -empty | wc -l` — expect 0
    Expected Result: 12/12 tasks have evidence, 0 empty files
    Evidence: .sisyphus/evidence/f2-evidence-completeness.log

  Scenario: Evidence files match planned QA scenarios
    Tool: Bash
    Steps:
      1. Extract all `Evidence:` paths from the plan: `grep -o 'Evidence: .*' .sisyphus/plans/e2e-verification.md | sed 's/Evidence: //'`
         (CRITICAL: macOS BSD `grep` lacks `-P` flag — use `grep -o` + `sed` instead of `grep -oP`)
      2. For each expected path, verify file exists: `test -f <path> && echo "OK" || echo "MISSING"`
      3. Count OK vs MISSING
    Expected Result: All planned evidence paths exist as files
    Evidence: .sisyphus/evidence/f2-evidence-mapping.log
  ```

- [x] F3. **Findings Report Compilation** — `writing`
  Compile all bugs/issues discovered during verification into `.sisyphus/findings.md`. Include: coordinator /health bug, --no-fallback missing, Docker image compiler gap, and any new issues found during testing. Categorize by severity (critical/medium/low).
  Output: `.sisyphus/findings.md` with structured findings

  **QA Scenarios:**

  ```
  Scenario: Findings report exists and has required sections
    Tool: Bash
    Steps:
      1. Verify file exists: `test -f .sisyphus/findings.md`
      2. Check for severity categories: `grep -c -E '(critical|medium|low)' .sisyphus/findings.md`
      3. Check for known issues: `grep -c -E '(/health|no-fallback|compiler gap)' .sisyphus/findings.md`
    Expected Result: File exists, contains severity categories, documents at least 3 known issues
    Evidence: .sisyphus/evidence/f3-findings-check.log

  Scenario: No critical findings left undocumented
    Tool: Bash
    Steps:
      1. Grep all evidence files for errors/failures: `grep -ril -E '(error|fail|FAIL)' .sisyphus/evidence/task-* 2>/dev/null`
      2. For each error found, verify it appears in findings.md or is an expected/passing test scenario
    Expected Result: All unexpected errors from evidence are documented in findings.md
    Evidence: .sisyphus/evidence/f3-error-coverage.log
  ```

- [x] F4. **Scope Fidelity Check** — `deep`
  Verify no production code was modified: `git diff --name-only` should only show files under `test/e2e/`, `.sisyphus/`, and `.gitignore`. No files under `cmd/`, `internal/`, `gen/`, `proto/`. Verify all test artifacts are in correct locations.
  Output: `Modified Files [N] | In-Scope [N/N] | VERDICT`

  **QA Scenarios:**

  ```
  Scenario: All modified files are within allowed scope
    Tool: Bash
    Steps:
      1. List all modified files: `git diff --name-only HEAD`
      2. Check each file is in allowed paths: `git diff --name-only HEAD | grep -v -E '^(test/e2e/|\.sisyphus/|\.gitignore)$'`
      3. Assert zero out-of-scope files
      4. Specifically verify forbidden directories untouched:
         `git diff --name-only HEAD | grep -E '^(cmd/|internal/|gen/|proto/)' | wc -l` — expect 0
    Expected Result: Zero files modified outside test/e2e/, .sisyphus/, .gitignore
    Evidence: .sisyphus/evidence/f4-scope-check.log

  Scenario: Test artifacts are in correct locations
    Tool: Bash
    Steps:
      1. Verify Docker infra: `test -f test/e2e/Dockerfile.worker && test -f test/e2e/docker-compose.yml`
      2. Verify test data: `ls test/e2e/testdata/*.c | wc -l` — expect >= 3
      3. Verify evidence dir: `ls .sisyphus/evidence/ | wc -l` — expect >= 20
      4. Verify cert gen script: `test -f test/e2e/gen-certs.sh`
    Expected Result: All test infrastructure artifacts exist at planned locations
    Evidence: .sisyphus/evidence/f4-artifact-locations.log
  ```

---

## Commit Strategy

- **1**: `test(e2e): add Docker infrastructure for E2E verification` — test/e2e/Dockerfile.worker, test/e2e/docker-compose*.yml
- **2**: `test(e2e): add C test project and TLS certs generation` — test/e2e/testdata/, test/e2e/gen-certs.sh
- **3**: `test(e2e): verify compilation pipeline + cache + dashboard` — .sisyphus/evidence/task-4-* through task-8-*
- **4**: `test(e2e): verify fallback + TLS + OTel + stress test` — .sisyphus/evidence/task-9-* through task-12-*
- **5**: `docs(e2e): compile findings report` — .sisyphus/findings.md

---

## Success Criteria

### Verification Commands
```bash
# All evidence files exist
ls .sisyphus/evidence/task-*.* | wc -l  # Expected: ≥20 evidence files

# No production code modified
git diff --name-only | grep -v -E '^(test/e2e/|\.sisyphus/|\.gitignore)' | wc -l  # Expected: 0

# Docker cleanup complete
docker compose -f test/e2e/docker-compose.yml ps  # Expected: no running containers

# Findings documented
test -f .sisyphus/findings.md  # Expected: file exists
```

### Final Checklist
- [x] All "Must Have" features verified with evidence
- [x] All "Must NOT Have" guardrails respected
- [x] Zero production code modifications
- [x] Findings report complete with categorized bugs
- [x] Docker cleanup complete (no orphaned containers)
