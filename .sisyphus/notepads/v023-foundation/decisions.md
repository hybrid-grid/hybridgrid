# Decisions — v0.2.3 Foundation

## [2026-03-15T05:47] Pre-Plan Decisions

1. **Config file loading**: CLI flags only for v0.2.3 — no `--config` flag refactoring
2. **Auth interceptor**: Explicitly deferred to v0.3.0 with TODO comment
3. **Log-level endpoint**: Both coordinator and worker, extracted to `internal/logging` package
4. **Log rotation**: Lumberjack dependency (cross-platform, in-process rotation)
5. **Test coverage targets**: Per-package (executor→60%, capability→60%, cli/build→60%, cli/output→70%)
6. **CHANGELOG**: One-time generation + Makefile target for future
