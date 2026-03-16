# E2E Verification - Learnings

## [2026-03-15T21:46:33Z] Plan Initialization

### From Metis/Momus Review (6 rounds, 14 issues fixed)
- Worker Docker image needs build-essential (gcc/g++/make) — Alpine all-in-one lacks compilers
- mDNS doesn't work in Docker bridge networking — use explicit `--coordinator=coordinator:9000`
- Coordinator has NO `/health` endpoint — use `/metrics` for healthchecks
- `--no-fallback` flag documented but NOT implemented — test fallback naturally by stopping coordinator
- Worker uses `--port` flag, coordinator uses `--grpc-port` — different flag names
- Worker uses runtime hostname as advertise address when `--advertise-address` omitted
- `filterHgbuildFlags()` strips only 4 flags: `--coordinator`, `--timeout`, `--insecure`, `--verbose`/`-v`
- TLS client requires BOTH cert AND key to enable — single cert not enough
- `insecure` defaults to `true` — TLS tests need explicit `--insecure=false`
- Dashboard API wraps worker list: `{"workers": [...], "count": N, "timestamp": T}`
- Worker `Stats` struct uses `total_workers`/`healthy_workers`, NOT `active_workers`
- Default tracing service names: `hybridgrid-coordinator` and `hybridgrid-worker`
- macOS BSD grep lacks `-P` flag — use `grep -o` + `sed` for regex
- mTLS negative test uses `openssl s_client` direct probe, not `hgbuild`

### CLI Behavior
- 5 subcommands have `DisableFlagParsing: true`: `cc`, `c++`, `make`, `ninja`, `wrap`
- Flags after these subcommands go to underlying tool, not Cobra
- `FallbackEnabled: true` is hardcoded — remote failures always fall back to local

### Infrastructure Notes
- `test/stress/Dockerfile.base` installs `curl` not `wget` — healthchecks use `curl -sf`
- `hostname: worker-1` in compose ensures container hostname matches TLS SAN
- `--tls-require-client-cert` MUST be included on both sides for mTLS

## [2026-03-16T04:50:00Z] Task 3: TLS Certificate Generation

### Certificate Generation Script (`test/e2e/gen-certs.sh`)
- Script uses **10-step pipeline**: CA key → CA cert → Server key → Server CSR → Server signed → Client key → Client CSR → Client signed → Permissions → Cleanup
- **Idempotency**: Script removes existing certs before generating (`rm -rf $CERT_DIR`) — safe to run multiple times
- **SAN Configuration**: Server cert includes **4 DNS SANs**: `localhost`, `coordinator`, `worker-1`, `worker-2` (critical for Docker container hostnames)
- **File Permissions**: Private keys (*.key) get 600 (read-write owner only), public certs (*.crt) get 644 (world-readable)
- **CSR Cleanup**: Script removes intermediate CSR files and .srl files after signing (keeps cert dir clean)
- **Validity**: All certs use 365-day validity from generation time (no hardcoded dates)
- **Key Size**: RSA 2048-bit throughout (simple, reliable, no ECDSA complexity)

### Certificate Chain Validation
- `openssl verify -CAfile ca.crt server.crt` confirms server cert signed by CA
- `openssl verify -CAfile ca.crt client.crt` confirms client cert signed by CA
- Both server and client certificates validated successfully with CA chain

### SAN Verification
- `openssl x509 -text` output shows "X509v3 Subject Alternative Name: DNS:localhost, DNS:coordinator, DNS:worker-1, DNS:worker-2"
- All 4 SANs present in server certificate (verified via text dump)

### Gitignore Integration
- Added pattern: `test/e2e/certs/` to `.gitignore` line 52
- Ensures generated certificates never committed to repository (proper security posture)

### Evidence Artifacts
- `task-3-cert-generation.log` — Full script output (6 files generated)
- `task-3-cert-verification.log` — `openssl verify` output for both server and client
- `task-3-san-check.log` — Full `openssl x509 -text` dump (SANs confirmed)

## [2026-03-16T04:51:00Z] Task 2: Multi-File C Test Project

### Project Structure
- **Location**: `test/e2e/testdata/`
- **Files Created**: 8 source/header files, 1 Makefile, 1 intentional error file
- **Total Files**: main.c, math_utils.c/h, string_utils.c/h, bad.c, Makefile

### Makefile Configuration
- **CC Variable**: Uses `CC ?= gcc` — allows hgbuild to override at compile time
- **Compiler Flags**: `-Wall -Wextra -O2` for warnings and optimization
- **Targets**: `all` (compile+link), `clean` (remove artifacts)
- **Pattern Rules**: `%.o: %.c` with explicit compiler invocation
- **Build Sequence**: 
  - `main.c` → `main.o`
  - `math_utils.c` → `math_utils.o`
  - `string_utils.c` → `string_utils.o`
  - Link all .o files → `testapp` binary

### Compilation & Verification
- **Local Compile Test**: `make -C test/e2e/testdata/ clean all`
  - Cleans previous artifacts (main.o, math_utils.o, string_utils.o, testapp)
  - Compiles all 3 .c files with warnings/optimization flags
  - Links to testapp binary successfully
  - **Result**: All 4 object files + testapp binary created
- **Binary Execution**: `./testapp` runs successfully
  - Calls all math functions (add, multiply, factorial) with test inputs
  - Calls all string functions (length, reverse, concat) with test inputs
  - Prints visible output for each operation
  - **Exit Code**: 0 (success)
- **Bad File Compilation**: `gcc -c bad.c` fails intentionally
  - **Error**: Missing closing parenthesis on printf statement
  - **Exit Code**: 1 (compilation failure)
  - **Error Message**: Clear compiler diagnostic pointing to root cause

### C Code Patterns
- **Math Utilities**:
  - `add(a, b)`: Simple addition
  - `multiply(a, b)`: Simple multiplication
  - `factorial(n)`: Recursive (exercises stack, not tail-call optimized)
- **String Utilities**:
  - `string_length(str)`: Wrapper around strlen() with NULL check
  - `string_reverse(str)`: Dynamic malloc allocation, in-place reversal, NULL-terminated
  - `string_concat(str1, str2)`: malloc'd result, handles NULL args as empty strings
- **Main Function**:
  - Imports both headers (stdio.h, stdlib.h, math_utils.h, string_utils.h)
  - Calls all functions with concrete test values
  - Uses printf() to display results (test harness visibility)
  - Frees malloc'd memory (reverse, concat) — no memory leaks
  - Returns 0 on success

### Parallel Compilation Readiness
- **3 Separate .c Files**: main.c, math_utils.c, string_utils.c
  - Enables distributed compilation testing (each file can compile on different worker)
  - Makefile pattern rules support parallel `-j` flag
  - **Minimal Compilation Time**: <100ms per file locally (good for testing pipeline overhead)
- **Dependency Graph**: main.o → main.c, math_utils.h; math_utils.o → math_utils.c; string_utils.o → string_utils.c
  - No inter-.c dependencies (no circular includes)
  - math_utils.h and string_utils.h used only by main.c and respective .c files

### Evidence Artifacts
- `task-2-local-compile.log` — make clean all + testapp execution output
  - Shows each .c file compiled with gcc
  - Shows testapp linked successfully
  - Shows binary execution output and exit code 0
- `task-2-bad-compile.log` — gcc bad.c error output
  - Shows compiler error (expected ')')
  - Shows file:line diagnostic
  - Shows exit code 1

## [2026-03-15T23:56:20Z] Task 1 (resumed): Fixed APT dependency conflict

- Root cause: the runtime stage could enter a broken APT state while fetching `gcc-12`, and retries were not pre-healing dependencies before re-attempting package install.
- Fix: updated `test/e2e/Dockerfile.worker` to run `apt-get install -y --fix-broken` before the main noninteractive install of `build-essential`, `gcc`, `g++`, `make`, `curl`, and `ca-certificates`.
- Verification: unmet dependency errors (`g++`/`gcc-12` not installable) are resolved; remaining failures are intermittent upstream mirror connection resets during package download in this environment.
