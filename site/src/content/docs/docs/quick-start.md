---
title: Quick start
description: Build the binaries and run a coordinator, worker, and dashboard on one machine.
---

## Build from source

Requires Go 1.25 (the module pins it via a `toolchain` directive — `go build`
will download it automatically if your local Go is older).

```bash
git clone https://github.com/h3nr1-d14z/hybridgrid.git
cd hybridgrid
make build
```

This produces four binaries in `bin/`: `hgbuild`, `hg-coord`, `hg-worker`,
`hg-dashboard`. `make build` also builds the dashboard's React frontend first
(`make build-ui` — requires Node 20+) since `hg-dashboard` embeds it via
`go:embed`.

## Start a coordinator

```bash
./bin/hg-coord serve
```

Default ports: **9000** (gRPC — workers and `hgbuild` talk to this) and
**8080** (HTTP — `/health`, `/metrics`, `/log-level`; ops only, no UI here).

## Start a worker

On the same machine or any other reachable one:

```bash
./bin/hg-worker serve --coordinator=localhost:9000
```

Default ports: **50052** (gRPC) and **9090** (HTTP `/health`, `/metrics`). The
worker auto-detects its compilers, architecture, Docker availability, and any
Flutter/Unity installs, then reports them at handshake.

If you omit `--coordinator`, the worker tries **mDNS auto-discovery** on the
LAN instead — convenient for ad-hoc clusters, but pass the flag explicitly
(or set `HG_COORDINATOR`) on anything routed, containerized, or behind NAT.

## Start the dashboard

```bash
./bin/hg-dashboard serve --insecure --port 8081
```

`--insecure` is required until you configure TLS between the dashboard and
coordinator (see [Configuration](/docs/configuration/)). Open
`http://localhost:8081` — see [Dashboard](/docs/dashboard/) for a tour.

## Run your first build

```bash
export HG_COORDINATOR=localhost:9000

# Groups files into one build the dashboard can show:
./bin/hgbuild build main.c utils.c -C localhost:9000

# Or wrap an existing Makefile (distributes per-file, doesn't group into
# one dashboard "build" yet — see Troubleshooting):
HG_COORDINATOR=localhost:9000 ./bin/hgbuild make -j8
```

Watch it land on the dashboard's Overview and Builds pages in real time.

## Or skip all of this with Docker Compose

```bash
docker compose up -d --scale worker=3
```

Brings up 1 coordinator + 3 workers + the dashboard, wired together on a
private network. See [Docker & clustering](/docs/docker-and-clustering/) for
scaling further and the caveats around running multiple workers on one host.
