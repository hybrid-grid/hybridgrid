# Hybridgrid Distributed Test Plan

## Test Environment

| Machine | OS | Arch | Role | Specs |
|---------|-----|------|------|-------|
| Mac | macOS | arm64 | Coordinator + Builder | M1/M2 |
| Windows | Windows 11 | amd64 | Worker | x86_64 |
| Raspi 5 | Linux | arm64 | Worker | ARM Cortex-A76 |

**Network:** Same LAN (WiFi/Ethernet)

---

## Phase 1: Setup & Connectivity

### 1.1 Build binaries for each platform
```bash
# On Mac (coordinator)
GOOS=darwin GOARCH=arm64 go build -o bin/hg-coord-darwin ./cmd/hg-coord
GOOS=darwin GOARCH=arm64 go build -o bin/hg-worker-darwin ./cmd/hg-worker
GOOS=darwin GOARCH=arm64 go build -o bin/hgbuild-darwin ./cmd/hgbuild

# For Windows worker
GOOS=windows GOARCH=amd64 go build -o bin/hg-worker.exe ./cmd/hg-worker

# For Raspi worker
GOOS=linux GOARCH=arm64 go build -o bin/hg-worker-linux ./cmd/hg-worker
```

### 1.2 Deploy binaries
- Copy `hg-worker.exe` to Windows machine
- Copy `hg-worker-linux` to Raspberry Pi
- Keep coordinator on Mac

### 1.3 Test connectivity
```bash
# On Mac - start coordinator
./hg-coord serve --grpc-port=9000 --http-port=8080

# On Windows - test connection
./hg-worker.exe serve --coordinator=<MAC_IP>:9000

# On Raspi - test connection
./hg-worker-linux serve --coordinator=<MAC_IP>:9000

# Verify on Mac
curl http://localhost:8080/workers
```

**Success Criteria:**
- [ ] Both workers appear in coordinator dashboard
- [ ] Heartbeats received every 30s
- [ ] No connection errors in logs

---

## Phase 2: mDNS Auto-Discovery

### 2.1 Test zero-config discovery
```bash
# On Mac - start coordinator with mDNS
./hg-coord serve --mdns

# On Windows - start worker without --coordinator flag
./hg-worker.exe serve

# On Raspi - start worker without --coordinator flag
./hg-worker-linux serve
```

**Success Criteria:**
- [ ] Workers auto-discover coordinator via mDNS
- [ ] Registration completes within 5s
- [ ] Works across all 3 platforms

---

## Phase 3: Basic Compilation Tests

### 3.1 Simple C file
```bash
# On Mac
echo 'int main() { return 0; }' > /tmp/test.c
./hgbuild cc /tmp/test.c -o /tmp/test
```

**Expected:** Compiles on one of the workers

### 3.2 C file with includes
```bash
# On Mac
cat > /tmp/hello.c << 'EOF'
#include <stdio.h>
int main() {
    printf("Hello from hybridgrid!\n");
    return 0;
}
EOF
./hgbuild cc /tmp/hello.c -o /tmp/hello
./tmp/hello
```

**Expected:** Preprocessed locally, compiled remotely, runs correctly

### 3.3 Cross-architecture test
```bash
# Force compilation on specific worker
./hgbuild cc /tmp/test.c -o /tmp/test --target=amd64  # Should go to Windows
./hgbuild cc /tmp/test.c -o /tmp/test --target=arm64  # Should go to Raspi or Mac
```

---

## Phase 4: Make Integration

### 4.1 Small project (curl)
```bash
git clone --depth=1 https://github.com/curl/curl /tmp/curl
cd /tmp/curl
autoreconf -fi
./configure --without-ssl
./hgbuild make -j4
```

**Metrics to capture:**
- Total build time with hybridgrid
- Total build time without hybridgrid (`make -j4`)
- Number of tasks distributed to each worker

### 4.2 Medium project (redis)
```bash
git clone --depth=1 https://github.com/redis/redis /tmp/redis
cd /tmp/redis
./hgbuild make -j4
```

### 4.3 Large project (CPython) - optional
```bash
git clone --depth=1 https://github.com/python/cpython /tmp/cpython
cd /tmp/cpython
./configure --disable-test-modules
./hgbuild make -j6
```

---

## Phase 5: Fault Tolerance

### 5.1 Worker disconnect during build
```bash
# Start build
./hgbuild make -j4 &

# Kill Windows worker mid-build
# (On Windows: Ctrl+C or close terminal)

# Observe:
# - Does build continue with remaining worker?
# - Are failed tasks retried on other workers?
# - Does build complete successfully?
```

### 5.2 Coordinator restart
```bash
# While workers are connected, restart coordinator
# Observe:
# - Do workers reconnect automatically?
# - Is there data loss?
```

### 5.3 Network interruption
```bash
# Temporarily disconnect Raspi from network
# Observe:
# - Does coordinator detect worker as unhealthy?
# - Does it stop sending tasks to disconnected worker?
# - Does worker reconnect when network restored?
```

---

## Phase 6: Performance Benchmarks

### 6.1 Latency test
```bash
# Measure single file compilation time
for i in {1..10}; do
    time ./hgbuild cc /tmp/test.c -c -o /tmp/test.o
done
# Compare with local: time gcc /tmp/test.c -c -o /tmp/test.o
```

### 6.2 Throughput test
```bash
# Generate 100 independent C files
for i in {1..100}; do
    echo "int func_$i() { return $i; }" > /tmp/src/file_$i.c
done

# Compile all with hybridgrid
time for f in /tmp/src/*.c; do ./hgbuild cc "$f" -c -o "${f%.c}.o" & done; wait

# Compare with local
time for f in /tmp/src/*.c; do gcc "$f" -c -o "${f%.c}.o" & done; wait
```

### 6.3 Real project benchmark (3 runs each)
```bash
cd /tmp/redis

# Baseline (no hybridgrid)
for i in 1 2 3; do
    make clean
    time make -j4
done

# With hybridgrid
for i in 1 2 3; do
    make clean
    time ./hgbuild make -j4
done
```

---

## Phase 7: Cache Validation

### 7.1 Cache hit test
```bash
# First build (cache miss)
./hgbuild cc /tmp/test.c -c -o /tmp/test.o
# Note: Should show [remote] in output

# Second build (cache hit)
./hgbuild cc /tmp/test.c -c -o /tmp/test.o
# Note: Should show [cache] in output and be instant
```

### 7.2 Cache invalidation
```bash
# Modify source file
echo '// comment' >> /tmp/test.c

# Rebuild (should miss cache)
./hgbuild cc /tmp/test.c -c -o /tmp/test.o
```

---

## Phase 8: Edge Cases

### 8.1 Large file (>1MB source)
```bash
# Generate large C file
python3 -c "print('int arr[] = {' + ','.join(str(i) for i in range(100000)) + '};')" > /tmp/large.c
./hgbuild cc /tmp/large.c -c -o /tmp/large.o
```

### 8.2 Many small files (1000 files)
```bash
for i in {1..1000}; do
    echo "int f$i(){return $i;}" > /tmp/many/f$i.c
done
time ./hgbuild make -j8  # with Makefile that compiles all
```

### 8.3 Compilation errors
```bash
echo 'int main() { syntax error }' > /tmp/bad.c
./hgbuild cc /tmp/bad.c -o /tmp/bad
# Expected: Error message from worker, non-zero exit code
```

---

## Test Results Template

| Test | Mac→Win | Mac→Raspi | Expected | Status |
|------|---------|-----------|----------|--------|
| 1.3 Connectivity | | | Both workers registered | |
| 2.1 mDNS | | | Auto-discovery works | |
| 3.1 Simple C | | | Compiles successfully | |
| 3.2 With includes | | | Preprocessed + compiled | |
| 4.1 curl build | | | Completes, shows speedup | |
| 5.1 Worker disconnect | | | Build continues | |
| 6.3 Redis benchmark | | | ≥1.2x speedup | |
| 7.1 Cache hit | | | Instant second build | |

---

## Success Criteria

1. **Functionality**: All Phase 1-4 tests pass
2. **Reliability**: Phase 5 fault tolerance tests pass
3. **Performance**: ≥1.2x speedup on real projects vs local build
4. **Cross-platform**: Works on all 3 platforms without modification

---

## Known Limitations

- Windows worker needs GCC/MinGW installed
- Raspberry Pi has limited RAM (may fail on large files)
- Network latency adds overhead for small files
- Cross-compilation between arm64/amd64 not yet supported

---

## Quick Start Commands

```bash
# === COORDINATOR (Mac) ===
./hg-coord serve --mdns --grpc-port=9000 --http-port=8080

# === WORKER (Windows) ===
./hg-worker.exe serve --coordinator=<MAC_IP>:9000

# === WORKER (Raspberry Pi) ===
./hg-worker-linux serve --coordinator=<MAC_IP>:9000

# === BUILD (Mac) ===
cd /path/to/project
./hgbuild make -j4
```
