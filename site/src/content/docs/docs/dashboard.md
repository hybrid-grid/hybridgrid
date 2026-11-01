---
title: Dashboard
description: A tour of hg-dashboard — the live control-room view of your fleet.
---

`hg-dashboard` is a standalone binary — a separate process, separate Docker
image, separate port from the coordinator — that renders a React SPA and
serves a small REST + WebSocket API translating the coordinator's
`TelemetryService` (gRPC) into something a browser can consume.

```
Browser ──HTTP/REST + WS──▶ hg-dashboard (:8081) ──gRPC──▶ hg-coord (:9000)
```

```bash
./bin/hg-dashboard serve --insecure --coordinator=localhost:9000 --port 8081
```

## Pages

**Overview** — 8 fleet metrics (active/queued/succeeded/failed tasks, cache
hit rate, healthy workers, avg build duration, coordinator uptime), a live
per-worker heartbeat trace (color follows circuit-breaker state: green
closed, amber half-open, red open — and it visibly flatlines when a breaker
trips), cache statistics broken down by Flutter/Unity, a cluster-wide
activity Gantt view, and a real-time event log fed by the WebSocket.

**Builds** — history table + a build-duration chart. Click through to a
build's detail page: status/outcome breakdown (a donut of
success/failed/running/queued), a per-task timeline (queue wait vs. compile
time), the task table, and a console panel with `stdout`/`stderr` tabs.

**Workers** — one card per worker: live heartbeat, executor slot grid
(busy/idle, not just a number), circuit-breaker badge, cores/memory,
success rate, latency, detected compilers. An **Add worker** button opens
instructions (the exact `hg-worker serve` command and a Compose scale
command) — there's no remote-provisioning API, a worker joins by actually
running the binary somewhere.

Any task's console output is also reachable directly at
`/tasks/<task-id>/console`, independent of which build it belonged to.

## Theme and language

Top-right of every page: a light/dark toggle and an EN/VI toggle, both
persisted in the browser (`localStorage`), applied before first paint so
there's no flash of the wrong theme on reload.

## Building it yourself

The frontend lives at `internal/observability/ui/web/` in the main repo (Vite
+ React + TypeScript, TanStack Router/Query, Zustand, Tailwind v4). Run
`npm run build` there (or `make build-ui` from the repo root) before building
`hg-dashboard` — its `go:embed` directive reads the built `web/dist`, which is
gitignored generated output, not committed source.
