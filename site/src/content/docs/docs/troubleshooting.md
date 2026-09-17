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

Known bug: `--coordinator` (and mDNS auto-discovery) don't reliably reach the
compile path for `cc`/`c++`/`make`/`ninja`, even when the address is
TCP-reachable and works fine for `hgbuild build`/`hg-worker`. **Workaround:**
set the `HG_COORDINATOR` environment variable instead of the flag.

## Builds via `make`/`cc` never show up grouped on the dashboard

`hgbuild cc`/`make` never set `build_id` on the compile request, despite
appearances in the code — only `hgbuild build file1 file2 ...` does. Tasks
from `make`/`cc` still get scheduled and cached normally, they just show up
as individual ungrouped tasks rather than one "build" on the dashboard's
Builds page.

## Running two workers on the same machine collapses into one

`hg-worker`'s ID is `worker-<hostname>` with no port/PID suffix and no
override flag — a second worker process on the same host overwrites the
first one's coordinator registry entry. Use separate containers/VMs (each
gets its own hostname) instead, e.g. `docker compose up -d --scale worker=N`.

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
