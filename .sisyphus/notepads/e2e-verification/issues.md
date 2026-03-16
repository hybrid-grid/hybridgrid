# E2E Verification - Issues

## [2026-03-15T21:46:33Z] Known Issues (Do NOT Fix)

### From Research Phase
- Coordinator `/health` endpoint missing — documented in README but not implemented
- `--no-fallback` flag documented in README but not in CLI code
- Docker compose healthcheck for coordinator silently broken (uses nonexistent /health)
