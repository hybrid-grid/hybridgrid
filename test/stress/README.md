# HybridGrid Stress Test

Stress test for HybridGrid distributed compilation using CPython as the test project.

## Setup

- **Coordinator**: 0.25 CPU, 256MB RAM
- **Workers**: 5 × (0.5 CPU, 512MB RAM) - simulates slower machines
- **Builder**: 1 CPU, 1GB RAM

Total resource usage: ~3.75 CPU cores, ~4GB RAM

## Test Project

**CPython v3.12.0**
- ~500 C source files
- Typical compile time: 8-12 minutes (single machine)
- Good stress test for distributed compilation

## Running the Test

```bash
# From this directory
./start.sh

# Or manually
docker compose up -d
docker compose exec builder bash /workspace/run-test.sh
```

## What Gets Tested

1. **Local Build (baseline)**: `make -j4` with regular gcc
2. **Distributed Build**: `hgbuild make -j8` using workers
3. **Cache Hit Test**: Rebuild to verify caching works

## Expected Results

On M1 Mac Air with 5 workers (0.5 CPU each):

| Test | Expected Time | Notes |
|------|---------------|-------|
| Local (j4) | ~8-10 min | Single machine baseline |
| Distributed | ~5-7 min | Distributed across workers |
| Cache rebuild | ~30-60s | Most files cached |

Expected speedup: **1.3-1.8x**

Note: Speedup is limited because:
- Workers are CPU-limited (0.5 CPU each)
- Network overhead for small files
- Preprocessing still done locally

## Monitoring

- **Dashboard**: http://localhost:8080
- **Logs**: `docker compose logs -f`
- **Worker status**: `docker compose exec builder hgbuild workers`

## Cleanup

```bash
docker compose down -v
```

## Troubleshooting

**Workers not connecting**
```bash
# Check coordinator logs
docker compose logs coordinator

# Check worker logs
docker compose logs worker-1
```

**Build fails**
```bash
# Enter builder container
docker compose exec builder bash

# Run manually
cd /workspace/cpython
hgbuild -v make
```

**Out of memory**
- Reduce worker count in docker-compose.yml
- Increase memory limits
