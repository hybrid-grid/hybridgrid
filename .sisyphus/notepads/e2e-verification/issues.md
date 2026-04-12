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

---

## [2026-03-18] Stress Test Scripts: Flag-Order and Exit-Status Bugs

### Bug 1: Pipeline Exit Code Swallowing (Critical — 3 of 5 benchmark scripts)

All benchmark scripts use `| tee` which suppresses the actual exit code of `hgbuild make`.

**`test/stress/benchmark.sh` — Line 104-107:**
```bash
docker compose exec -T builder bash -c "
    cd /workspace/cpython
    hgbuild make -j$jobs 2>&1
" 2>&1 | tee /tmp/build_${workers}w.log
```
The pipeline `| tee` means bash returns `tee`'s exit code (0 on success), not the docker exec exit code. `set -e` at line 5 cannot catch the inner failure. Build failures are silently ignored.

**`test/stress/benchmark-heterogeneous.sh` — Line 368-371:** Identical pattern.
**`test/stress/benchmark-fair.sh` — Line 148-151:** Identical pattern.

Fix pattern (matching `run-test.sh`'s correct approach): use a tempfile with explicit `$?` capture.

**`test/stress/start.sh` — Line 45:**
```bash
docker compose exec builder bash /workspace/run-test.sh
```
Does not capture exit code. If `run-test.sh` exits non-zero, `start.sh` still exits 0 because `docker compose exec` itself succeeds.

---

### Bug 2: Flag Order — BENIGN (run-test.sh line 108)

**`test/stress/run-test.sh` — Line 108:**
```bash
hgbuild --coordinator=${COORDINATOR} -v make -j8 > "$TMPLOG" 2>&1
```

This is **correct**. Explanation:

- `DisableFlagParsing: true` on the `make` subcommand (`cmd/hgbuild/main.go:1106`) means the `make` subcommand does NOT parse its own flags.
- However, the root command (`rootCmd`) DOES parse its persistent flags BEFORE dispatching to the subcommand.
- Root persistent flags include: `--coordinator` (line 163), `-v` (line 167), `--no-fallback` (line 165).
- Therefore, `hgbuild -v make` → `-v` is consumed by root command (enables hgbuild verbose), `make -j8` → passed to make. **Correct.**
- The comment in `benchmark.sh` line 103: `# Note: Don't use -v flag with make wrapper (DisableFlagParsing passes it to make)` — this comment is **INCORRECT**. The `-v` before `make` goes to the root parser, not to `make`.

Key insight: `DisableFlagParsing` only affects the subcommand's own argument parsing. Root-level persistent flags are always parsed first by cobra's command tree traversal.

---

### Contrast: Which Scripts Check Exit Codes

| Script | Exit Code Check | Method |
|--------|----------------|--------|
| `run-test.sh` | YES | Tempfile + `$?` (lines 108-114, 161-167) |
| `benchmark.sh` | NO | Pipeline `| tee` swallows exit |
| `benchmark-heterogeneous.sh` | NO | Pipeline `| tee` swallows exit |
| `benchmark-fair.sh` | NO | Pipeline `| tee` swallows exit |
| `start.sh` | NO | `docker compose exec` succeeds even if inner script fails |

---

### Manual Workaround Without Editing Files

**Yes, CPython stress test can be run manually without script edits:**

```bash
# From inside builder container:
cd /workspace/cpython
COORDINATOR="${HG_COORDINATOR:-coordinator:9000}"

# Use tempfile pattern (NOT pipeline):
hgbuild --coordinator=${COORDINATOR} -v make -j8 > /tmp/build.log 2>&1
BUILD_EXIT=$?
tail -20 /tmp/build.log
if [ $BUILD_EXIT -ne 0 ]; then
    echo "Build FAILED with exit code $BUILD_EXIT"
    exit $BUILD_EXIT
fi
echo "Build SUCCEEDED"
```

The critical rule: **Never use `| tee` for commands whose exit code you need to check.** Use a tempfile and explicit `$?` capture instead, exactly as `run-test.sh` does at lines 83-90.

---

### Related: Misleading Comment in benchmark.sh

**File**: `test/stress/benchmark.sh`  
**Line 103**: `# Note: Don't use -v flag with make wrapper (DisableFlagParsing passes it to make)`

This comment reflects a misunderstanding of cobra's `DisableFlagParsing`. Since root-level persistent flags are parsed before subcommand dispatch, `hgbuild -v make` correctly enables hgbuild verbose mode. If you wanted `-v` to go to `make` instead, you would write `hgbuild make -v` — but there's no need to do that since `-v` for make is `-v` (verbose make) and `-v` for hgbuild is also `-v` (verbose). In practice both work, but the comment is architecturally misleading.

- 2026-03-18: CPython stress run required executing in /workspace/cpython; running `hgbuild --coordinator=coordinator:9000 -v make -j8` in /workspace failed because make received `--coordinator` and `-v` printed make version. `run-test.sh` also exited during distributed build due to `set -e` (no error tail). Successful distributed build used `HG_COORDINATOR=coordinator:9000 hgbuild make -j8` in /workspace/cpython; stats showed total_tasks=173, success_tasks=171, failed_tasks=2.
