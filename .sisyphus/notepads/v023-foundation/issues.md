# Issues — v0.2.3 Foundation

## Known Issues (from Metis review)

1. **Config struct mismatch**: `config.TracingConfig` has 4 fields but `tracing.Config` has 8 — Task 1 expands before OTel wiring
2. **LogConfig.File is dead code**: Made functional in Task 8 before Task 9 adds lumberjack
3. **Hardcoded `Insecure: true`**: At `cmd/hg-worker/main.go:110` — removed in Task 4
4. **No HTTP mux for coordinator log-level**: May need dashboard mux extension or separate HTTP server

## Platform-Specific Constraints
- MSVC/Windows tests excluded (macOS dev environment)
- Docker tests must work without daemon (mock Docker client)
