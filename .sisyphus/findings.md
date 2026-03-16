# E2E Verification Findings

## Docker Build Failures

### Issue: APT Dependency Conflict in test/e2e/Dockerfile.worker

**File:** `test/e2e/Dockerfile.worker` (lines 41-59)

**Symptom:**
```
The following packages have unmet dependencies:
 g++ : Depends: gcc-12 (>= 12.2.0-1~) but it is not going to be installed
 g++-12 : Depends: gcc-12 (= 12.2.0-14+deb12u1) but it is not going to be installed
 gcc : Depends: gcc-12 (>= 12.2.0-1~) but it is not going to be installed
E: Unmet dependencies. Try 'apt --fix-broken install' with no packages (or specify a solution).
```

**Root Cause:**
Line 55 runs `apt-get -f install -y || true` AFTER the first install failure, but this doesn't work because:
1. The broken state already exists when line 44's install fails
2. The `|| true` silently swallows the error but doesn't fix the dependency tree
3. The retry logic at line 56 doesn't address the broken packages

**Proposed Fix:**
Replace lines 41-59 with:
```dockerfile
RUN set -eux; \
    apt-get update; \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        build-essential \
        gcc \
        g++ \
        make \
        curl \
        ca-certificates; \
    rm -rf /var/lib/apt/lists/*
```

OR (if retries needed):
```dockerfile
RUN set -eux; \
    apt-get update; \
    apt-get install -y --fix-broken; \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        build-essential \
        gcc \
        g++ \
        make \
        curl \
        ca-certificates; \
    rm -rf /var/lib/apt/lists/*
```

**Impact:** HIGH — Blocks all Docker-based E2E tests (Tasks 4-12)

**Workaround for Verification:**
1. Build binaries directly with `go build` instead of Docker multi-stage build
2. Skip Docker image validation for Wave 1
3. Document as blocker for Wave 2+ (cluster startup tests)

---

## Verified Components (Wave 1 Partial)

### ✅ Docker Compose Syntax
- Base compose config: VALID
- TLS overlay: VALID
- OTel overlay: VALID

### ✅ C Test Project
- Compiles cleanly with `make all`
- Binary runs, outputs correct math/string results
- Error path (`bad.c`) produces expected compilation error

### ✅ TLS Certificates
- Generation script works (idempotent)
- CA → server chain: VALID
- CA → client chain: VALID
- Server cert SAN field: Contains all 4 DNS names (localhost, coordinator, worker-1, worker-2)

### ❌ Docker Image Build
- BLOCKED by external infrastructure issue:
  - Debian APT mirror `deb.debian.org` (151.101.2.132) consistently fails on `gcc-12_12.2.0-14+deb12u1_arm64.deb`
  - Error: "Error reading from server. Remote end closed connection"
  - Tested with 3 retry attempts @ 5s backoff — all failed
  - This is NOT a Dockerfile bug — network/mirror issue
  - **Workaround for E2E**: Use locally-built binaries with `docker compose run` instead of image build
