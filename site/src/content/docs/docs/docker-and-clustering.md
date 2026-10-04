---
title: Docker & clustering
description: Running the fleet in Docker Compose, scaling workers, and what to know before moving to Kubernetes.
---

## Docker Compose (LAN/dev)

```bash
docker compose up -d --scale worker=3
```

Brings up `coordinator` (ports 9000 gRPC, 8080 HTTP), `dashboard` (8081), and
3 `worker` replicas on a private bridge network, coordinator address wired as
`coordinator:9000` via Compose's internal DNS. Scale further at any time:

```bash
docker compose up -d --scale worker=8
```

Each container gets its own hostname, which sidesteps a real worker-ID
collision bug (see below) — this is the easy way to see a multi-worker
cluster without touching a second physical machine.

For Flutter, use the separate `docker-compose.flutter.yml` (a Flutter-capable
worker image is much larger and not something you want in the default dev
stack).

## Moving to a real multi-machine cluster / Kubernetes

mDNS auto-discovery only works within one LAN broadcast domain — it will
**not** work across Kubernetes pods or routed networks. For anything beyond
Docker Compose on one host:

- Point workers at the coordinator explicitly:
  `--coordinator=hg-coord.namespace.svc.cluster.local:9000` (or `HG_COORDINATOR`).
- Set `--advertise-address` on each worker to something the coordinator can
  dial back into — the coordinator calls workers back to dispatch tasks, it's
  not request/response-only. In Kubernetes, that's typically the pod IP via
  the downward API.
- Turn on `--token` (shared secret) and TLS/mTLS once you're off a trusted
  LAN — see [Configuration](/docs/configuration/).
- No Kubernetes manifests ship in the repo yet — only `docker-compose.yml`.
  You'll need to write a `Deployment`/`Service` for `hg-coord` and a
  `Deployment`/`StatefulSet` for `hg-worker` yourself.

## Known limitation: worker-ID collisions on one host

`hg-worker`'s ID is generated as `worker-<hostname>` — hostname only, no
port or PID suffix, and there's no `--worker-id` override flag. Two
`hg-worker` processes on the **same machine** (different `--port`, not in
separate containers) will silently overwrite each other's registry entry on
the coordinator instead of showing up as two workers. Containers/pods each
get their own hostname and aren't affected. Tracked in the repo's `TODO.md`.
