---
title: Troubleshooting
description: Common problems and the known, tracked bugs behind some of them.
---

## "no workers match requirements"

1. Worker hasn't registered the right capability — check `gcc`/`g++`/`clang`
   are actually on the worker's `PATH`.
2. Worker's heartbeat expired (coordinator TTL defaults to 60s).
3. Architecture mismatch and Docker isn't available on the worker for
   cross-compilation.

Check the coordinator log for `cpp_compilers=[...]` per worker to confirm
what it actually detected.

## Worker not reachable from the coordinator

The coordinator dials workers back to dispatch tasks — it isn't purely
request/response. If it can't reach a worker:

```bash
hg-worker serve --coordinator=coord:9000 --advertise-address=192.168.1.50:50052
```

Always required behind NAT, Docker port remapping, or Kubernetes.

## Compilation keeps falling back to local

Check `hgbuild -v ...` output for the actual reason. Usually: coordinator not
running, worker timeout, or network partition between client and
coordinator.

## `hgbuild cc`/`make` fail with "error reading server preface: EOF"

Earlier versions silently ignored `--coordinator` for `cc`/`c++`/`make`/`ninja`/
`wrap` (they run with flag parsing disabled), so a build could quietly fall
back to mDNS discovery or the wrong address. All of them now honour
`--coordinator`/`-C` (see the [CLI reference](/docs/cli-reference/)), and the
wrappers export it as `HG_COORDINATOR` to the compile steps they spawn.

Whether the `server preface: EOF` failure itself is gone is **not confirmed**
against a live coordinator. It points at a transport mismatch (plaintext vs.
TLS, or an mDNS-advertised address that isn't the coordinator's gRPC port). One
known mismatch: `--tls-cert`/`--tls-key`/`--tls-ca` and `--insecure` given to
`hgbuild make`/`ninja`/`wrap` are not forwarded to the compiler calls the build
spawns, so those dial the coordinator in plaintext. If you still hit the error,
check that the coordinator is not TLS-only, and set `HG_COORDINATOR` explicitly
(`host:9000`) instead of relying on mDNS.

## `cc`/`c++` tasks show up ungrouped on the dashboard

`hgbuild build`, `hgbuild make`, `ninja`, and `wrap` each create a build
session (the wrappers export `HG_BUILD_ID`), so their tasks group into one
build on the Builds page. A bare `hgbuild cc`/`c++` call only sends a
`build_id` when `HG_BUILD_ID` is already exported — otherwise it shows up as
an individual task. Export `HG_BUILD_ID` yourself (any non-empty value) when
driving `cc`/`c++` from your own build script.

## Running two workers on the same machine collapses into one

`hg-worker`'s default ID is `worker-<hostname>` with no port/PID suffix — a
second worker process on the same host overwrites the first one's coordinator
registry entry unless you pass a distinct `--worker-id` to each. Alternatively
use separate containers/VMs (each gets its own hostname), e.g.
`docker compose up -d --scale worker=N`.

## Cache not working

```bash
ls -la ~/.hybridgrid/cache      # check permissions
hgbuild cache clear             # nuke and retry
```

## Dashboard shows a blank/500 page after adding workers with unusual capabilities

Fixed as of the current dashboard build — earlier versions assumed the
`compilers`/`architectures`/`build_types` arrays in the worker API response
were never `null` (Go encodes an empty slice as JSON `null`, not `[]`), which
crashed the Workers page for any worker reporting no compilers (e.g. a bare
Flutter/Unity-only worker). If you're on an older build, rebuild the
dashboard frontend.
