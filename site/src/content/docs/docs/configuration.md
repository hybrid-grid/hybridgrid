---
title: Configuration
description: config.yaml layout, environment variables, and TLS setup.
---

Config resolves in this order: CLI flags → environment variables → the
config file (`~/.hybridgrid/config.yaml` by default) → built-in defaults.

## Full structure

```yaml
coordinator:
  grpc_port: 9000
  http_port: 8080
  auth_token: ""       # shared secret; leave empty to disable auth
  tls_cert: ""
  tls_key: ""
  mdns_enable: true

worker:
  port: 50052
  coordinator_addr: "localhost:9000"
  auth_token: ""
  max_parallel: 0      # 0 = auto-detect from CPU cores
  work_dir: ""
  timeout: 5m
  heartbeat_sec: 30

client:
  coordinator_addr: ""  # empty = mDNS auto-discovery
  auth_token: ""
  timeout: 2m
  fallback: true         # compile locally if the coordinator is unreachable

cache:
  enable: true
  dir: ~/.hybridgrid/cache
  max_size_mb: 10240
  ttl_hours: 0           # 0 = no expiry

log:
  level: info
  format: console        # or "json"
  file: ""
  rotation:
    max_size_mb: 100
    max_backups: 3
    max_age_days: 28
    compress: true

tls:
  enabled: false
  cert_file: ""
  key_file: ""
  client_ca: ""
  require_client_cert: false   # true = mTLS
  insecure_skip_verify: false

tracing:
  enable: false
  endpoint: "localhost:4317"
  service_name: "hybridgrid"
  sample_rate: 0.01
  insecure: true
  batch_size: 512
```

## Environment variables

The worker and CLI both honor `HG_COORDINATOR` as a fallback when
`--coordinator` isn't passed. Other useful ones seen in the Docker images:
`HG_LOG_LEVEL`, `HG_HEARTBEAT_TTL`, `HG_MAX_CONCURRENT_TASKS`.

## TLS / mTLS

```bash
# Self-signed cert for testing
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes

hg-coord serve --tls-cert=cert.pem --tls-key=key.pem
# add --tls-ca=ca.pem --tls-require-client-cert for mTLS

hg-worker serve --coordinator=coordinator:9000 \
  --tls-cert=client.pem --tls-key=client-key.pem --tls-ca=ca.pem
```

`hg-dashboard` takes the same `--tls-cert`/`--tls-key`/`--tls-ca` flags for
its gRPC connection to the coordinator — pass `--insecure` instead while
TLS isn't configured.

:::caution
`hg-dashboard` only accepts `--coordinator`/`--insecure` via CLI flags today,
not environment variables (unlike `hg-worker`/`hgbuild`). Set them explicitly
in your process manager or Compose file's `command:`.
:::
