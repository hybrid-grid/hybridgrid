## [2026-03-16T10:20:00Z] Task 5: ROOT CAUSE IDENTIFIED

### The Core Issue
OS mismatch between client (macOS) and workers (Linux containers):
- Client preprocesses source on macOS → macOS-specific header paths
- Coordinator requires OS-matching worker for preprocessed source
- All workers are Linux → No match found → "no workers match requirements"

### Why The Design Works This Way
**Preprocessed Mode**: Headers already expanded with OS-specific paths → Worker MUST be same OS
**Raw Source Mode**: Headers NOT expanded → Worker can be any OS (uses Docker for cross-compile)

### Evidence Chain
1. ✅ Workers detect C++ capabilities correctly (gcc, g++)
2. ✅ Workers send capabilities in HandshakeRequest
3. ✅ Coordinator receives and stores capabilities
4. ✅ Coordinator logs confirm: `cpp_compilers=["gcc","g++"]`, `arch=ARCH_ARM64`
5. ❌ Coordinator filters workers by OS: `clientOSFilter="darwin"` but workers are `Os="linux"`
6. ❌ Scheduler returns ErrNoMatchingWorkers

### Root Cause Code Locations
- **Client preprocessing**: `internal/cli/build/build.go:186-195`
- **Coordinator OS filtering**: `internal/coordinator/server/grpc.go:20-29` (Compile handler)
- **Scheduler OS filter**: `internal/coordinator/scheduler/scheduler.go:53-59`
- **Worker OS detection**: `internal/capability/detect.go:22` (`runtime.GOOS`)

### Solution Options
See `.sisyphus/findings.md` for full analysis.

**RECOMMENDED**: Run `hgbuild` inside Linux container (same OS as workers)
- Test FROM worker container: `docker compose exec worker-1 hgbuild cc -c /testdata/main.c -o /tmp/main.o`
- OR: Add hgbuild binary to worker image, mount testdata, compile from inside

**ALTERNATIVE**: Force preprocessing failure → raw source fallback (tests cross-OS path)

### User Decision Required
This is an E2E test design choice, not a product bug. The system is working as designed.
Need user input on which path to test.

- 2026-03-16T09:12:37Z F1 audit follow-up: auxiliary evidence files `task-6-metrics-before.log`, `task-6-metrics-after.log`, `task-7-metrics-summary.txt`, and `task-12-cleanup.log` are empty; `.sisyphus/plans/e2e-verification.md` is also modified in the working tree despite the read-only rule.
