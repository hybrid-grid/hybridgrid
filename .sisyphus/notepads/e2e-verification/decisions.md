# E2E Verification - Decisions

## [2026-03-15T21:46:33Z] Architecture Decisions

### Testing Approach
- Docker Compose on macOS (single machine, bridge network)
- Test workload: Small C project first, then CPython stress test
- Feature coverage: Everything except Build()/StreamBuild() stubs

### Infrastructure Choices
- Base image: Debian bookworm-slim with build-essential (NOT Alpine)
- Pattern source: `test/stress/Dockerfile.base` — proven working
- Healthcheck: `/metrics` for coordinator (since /health missing), `/health` for workers
- TLS: Self-signed certs (1 CA, 1 server cert with SANs, 1 client cert)

### Test Scope (Explicit Exclusions)
- ❌ mDNS (doesn't work in Docker bridge)
- ❌ Build()/StreamBuild() RPC stubs (v0.3.0 scope)
- ❌ `--no-fallback` flag (not implemented)
- ❌ Cross-compilation (same-architecture only)
- ❌ Bug fixes (document only in findings.md)
