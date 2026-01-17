# Troubleshooting Guide

## Common Issues

### Workers Not Discovered

**Symptoms:** Coordinator shows 0 workers, builds fail immediately

**Causes & Solutions:**

1. **mDNS blocked by firewall**
   ```bash
   # Linux: Allow mDNS
   sudo ufw allow 5353/udp

   # macOS: Check if Bonjour is enabled
   sudo launchctl list | grep mDNS
   ```

2. **Different subnets**
   - mDNS only works on same network segment
   - Use static worker configuration instead:
   ```yaml
   coordinator:
     static_workers:
       - "192.168.2.10:50052"
       - "192.168.2.11:50052"
   ```

3. **Worker not running**
   ```bash
   # Check worker status
   curl http://worker-ip:9090/health
   ```

### Connection Refused

**Symptoms:** `connection refused` errors

**Solutions:**

1. **Check service is running**
   ```bash
   # Coordinator
   curl http://localhost:8080/health

   # Worker
   curl http://localhost:9090/health
   ```

2. **Check ports are open**
   ```bash
   netstat -tlnp | grep -E '9000|8080|50052|9090'
   ```

3. **Docker networking**
   ```bash
   # Use host.docker.internal on Docker Desktop
   hg-worker serve --coordinator=host.docker.internal:9000
   ```

### Circuit Breaker Open

**Symptoms:** Tasks failing with "circuit breaker open"

**Diagnosis:**
```bash
# Check dashboard for circuit states
open http://localhost:8080

# Check metrics
curl -s http://localhost:8080/metrics | grep circuit
```

**Solutions:**

1. **Identify failing worker**
   - Dashboard shows circuit state per worker
   - Check worker logs for errors

2. **Fix underlying issue**
   - Compiler not installed
   - Disk full
   - Resource exhaustion

3. **Wait for recovery**
   - Circuit auto-resets after `open_timeout` (default 30s)
   - Half-open state tests with single request

### Compilation Timeouts

**Symptoms:** Tasks timeout, stderr shows "deadline exceeded"

**Solutions:**

1. **Increase timeout**
   ```yaml
   worker:
     task_timeout: 10m  # Default is 5m
   ```

2. **Check worker resources**
   ```bash
   # On worker machine
   top -bn1 | head -20
   df -h
   ```

3. **Reduce concurrent tasks**
   ```yaml
   worker:
     max_concurrent_tasks: 2  # Reduce from default 4
   ```

### Cache Not Working

**Symptoms:** Cache hit rate is 0%, rebuilding everything

**Diagnosis:**
```bash
# Check cache directory
ls -la ~/.hybridgrid/cache/

# Check cache metrics
curl -s http://localhost:8080/metrics | grep cache
```

**Solutions:**

1. **Cache disabled**
   ```yaml
   cache:
     enabled: true  # Ensure this is set
   ```

2. **Cache directory permissions**
   ```bash
   chmod 755 ~/.hybridgrid/cache
   ```

3. **Different compiler flags**
   - Cache key includes all compiler arguments
   - Ensure consistent build commands

### Docker Executor Fails

**Symptoms:** "Docker not available" errors

**Solutions:**

1. **Docker not installed**
   ```bash
   docker --version
   ```

2. **Docker socket permissions**
   ```bash
   sudo usermod -aG docker $USER
   # Log out and back in
   ```

3. **Pull dockcross images**
   ```bash
   docker pull dockcross/linux-arm64
   docker pull dockcross/linux-armv7
   ```

### TLS Handshake Errors

**Symptoms:** "TLS handshake failed", "certificate verify failed"

**Solutions:**

1. **Certificate expired**
   ```bash
   openssl x509 -in server.crt -noout -dates
   ```

2. **Wrong CA certificate**
   - Ensure client has correct CA cert
   - Check cert chain is complete

3. **Hostname mismatch**
   ```bash
   openssl x509 -in server.crt -noout -text | grep DNS
   ```

### High Memory Usage

**Symptoms:** OOM kills, slow performance

**Solutions:**

1. **Limit concurrent tasks**
   ```yaml
   worker:
     max_concurrent_tasks: 2
   coordinator:
     max_concurrent_tasks: 500
   ```

2. **Limit cache size**
   ```yaml
   cache:
     max_size_gb: 5
   ```

3. **Check for memory leaks**
   ```bash
   # Enable pprof
   curl http://localhost:8080/debug/pprof/heap > heap.prof
   go tool pprof heap.prof
   ```

## Debug Mode

Enable verbose logging:

```bash
# Environment variable
HG_LOG_LEVEL=debug hg-coord serve

# Or config file
logging:
  level: debug
  format: json
```

## Log Analysis

### Common Log Patterns

**Worker registration:**
```
level=info msg="worker registered" worker_id=worker-1 arch=amd64 source=mdns
```

**Task execution:**
```
level=info msg="task started" task_id=task-123 worker=worker-1 compiler=gcc
level=info msg="task completed" task_id=task-123 duration=234ms success=true
```

**Circuit breaker:**
```
level=warn msg="circuit opened" worker=worker-2 failure_rate=0.65
level=info msg="circuit half-open" worker=worker-2
level=info msg="circuit closed" worker=worker-2
```

**Errors:**
```
level=error msg="compilation failed" error="exit code 1" stderr="..."
level=error msg="worker unreachable" worker=worker-3 error="connection refused"
```

## Health Check Endpoints

| Endpoint | Expected | Meaning |
|----------|----------|---------|
| `GET /health` | `200 OK` | Service running |
| `GET /metrics` | Prometheus data | Metrics available |
| `gRPC HealthCheck` | `healthy=true` | gRPC service healthy |

## Getting Help

1. **Check logs** - Always start with debug logs
2. **Check metrics** - Dashboard and `/metrics` endpoint
3. **GitHub Issues** - Search existing issues or create new one
4. **Include diagnostics:**
   ```bash
   # System info
   uname -a
   go version
   docker version

   # Service status
   curl http://localhost:8080/api/stats
   curl http://localhost:8080/metrics | grep hybridgrid

   # Logs (last 100 lines)
   journalctl -u hg-coord -n 100
   ```
