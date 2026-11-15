---
title: What is Hybrid-Grid
description: A distributed build system for C/C++, Flutter, and Unity that spreads compilation across a LAN of worker machines.
---

Hybrid-Grid is a Go-based distributed build system. It spreads compilation work
across multiple machines on your network instead of running everything on one
box — the way `distcc`/`icecc` do for C/C++, but with a coordinator, a
scheduler, a content-addressable cache, and a live web dashboard on top.

## The three pieces

```
hgbuild (CLI) ──gRPC──▶ hg-coord (Coordinator) ──gRPC──▶ hg-worker (Node 1..N)
```

- **`hgbuild`** — the CLI you actually run. It wraps `make`/`ninja`, acts as a
  drop-in `cc`/`c++` replacement, or drives `flutter`/`unity` builds.
- **`hg-coord`** — the coordinator. Workers register with it, it schedules
  incoming tasks to the best available worker (P2C — Power of Two Choices),
  and it trips a circuit breaker per worker when one starts failing.
- **`hg-worker`** — runs on every machine that contributes compute. It
  reports its capabilities (compilers, architectures, Docker, Flutter SDK,
  Unity installs) and executes the tasks the coordinator hands it.

A fourth binary, **`hg-dashboard`**, is a standalone web UI that watches the
coordinator over gRPC and gives you a live view of the fleet — see
[Dashboard](/docs/dashboard/).

## What it can build today

| Type | Status |
|---|---|
| C/C++ | Production — cross-compilation, MSVC flag translation, Docker cross-compile via dockcross |
| Flutter (Android) | Working — distributed `flutter build apk`/`appbundle` via Docker workers |
| Unity | Working — real `Unity -batchmode` execution on workers with Unity Hub installed (not containerized — Editor licensing has to happen on the worker itself) |
| Rust / Go / Node.js | Protocol scaffolding exists (`RustConfig`, `GoConfig`, `NodeConfig`); no executor wired up yet |

## Why a coordinator instead of just SSH-and-run

Three things you don't get from a bag of SSH scripts:

- **Content-addressable cache** — identical inputs (hashed) skip
  recompilation entirely, locally or across the fleet.
- **Fault tolerance** — a circuit breaker isolates a flaky worker instead of
  hanging every build that hits it; if the coordinator itself is
  unreachable, `hgbuild` falls back to compiling locally.
- **Observability** — the [dashboard](/docs/dashboard/) shows worker health,
  build history, and per-task console output in real time, backed by
  Prometheus metrics on both binaries.

Ready to try it? Go to [Quick start](/docs/quick-start/).
