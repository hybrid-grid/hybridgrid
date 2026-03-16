# Config Validation Timing Fix - COMPLETED

## Issue
Config validation was happening BEFORE CLI flags were applied, causing invalid CLI values to bypass validation.

**Example bug scenario:**
- `./bin/hg-coord serve --grpc-port=8080 --http-port=8080` would NOT get caught by validation
- Instead, it would fail later with "bind: address already in use" (network error)
- Should have failed early with "invalid configuration: ports must be different"

## Root Cause
In `cmd/hg-coord/main.go`:
- Lines 29-32: Validated `config.DefaultConfig()` in `main()` function
- Lines 66-70: CLI flags extracted in `serveCmd.RunE()`
- Lines 123+: Server config created with CLI values (AFTER validation already ran)

**Timeline:**
1. ❌ Line 29: Validate defaults (port=9000, port=8080) → PASS (defaults are valid)
2. ✓ Line 70: Get CLI flags (--grpc-port=8080, --http-port=8080)
3. ✓ Line 123: Create server config with invalid CLI values (no re-validation!)

## Solution
**File:** `cmd/hg-coord/main.go`

### Change 1: Remove early validation (Lines 29-32)
```go
// BEFORE
cfg := config.DefaultConfig()
if err := cfg.Validate(); err != nil {
    log.Fatal().Err(err).Msg("config validation failed")
}

// AFTER
cfg := config.DefaultConfig()
```

### Change 2: Add validation after CLI flags extracted (After Line 70)
```go
// Validate port ranges
if grpcPort < 1 || grpcPort > 65535 {
    return fmt.Errorf("invalid configuration: coordinator.grpc_port must be 1-65535, got %d", grpcPort)
}
if httpPort < 1 || httpPort > 65535 {
    return fmt.Errorf("invalid configuration: coordinator.http_port must be 1-65535, got %d", httpPort)
}
if grpcPort == httpPort {
    return fmt.Errorf("invalid configuration: coordinator.grpc_port and coordinator.http_port must be different, got %d for both", grpcPort)
}
```

## Verification Results
✅ **All checks passed:**

1. **Build:** `go build -o bin/hg-coord ./cmd/hg-coord` → SUCCESS
2. **Validation (conflicting ports):** `./bin/hg-coord serve --grpc-port=8080 --http-port=8080` → Correctly rejects with:
   ```
   Error: invalid configuration: coordinator.grpc_port and coordinator.http_port must be different, got 8080 for both
   ```
3. **Validation (invalid port range):** `./bin/hg-coord serve --grpc-port=99999 --http-port=8080` → Correctly rejects with:
   ```
   Error: invalid configuration: coordinator.grpc_port must be 1-65535, got 99999
   ```
4. **Tests:** 645 tests passed, 0 failures, race-clean ✓

## Impact
- **Files changed:** 1 (`cmd/hg-coord/main.go`)
- **Lines added:** 11 (validation)
- **Lines removed:** 4 (early validation)
- **Regressions:** 0
- **Breaking changes:** None (early validation of defaults was meaningless)

## Future Work
Consider applying same fix to:
- `cmd/hg-worker/main.go` (if it has similar validation timing issue)
- `cmd/hgbuild/main.go` (if it validates configs)

But this task was **COORDINATOR ONLY** per spec.
