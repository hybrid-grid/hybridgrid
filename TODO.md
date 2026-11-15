# TODO

## Backend bugs (found while demoing the dashboard, 2026-09-18)

- [ ] **`hgbuild cc`/`hgbuild c++`/`hgbuild make` ignore `--coordinator`/`-C`.**
      Remote compilation fails with `rpc error: code = Unavailable desc =
      connection error: desc = "error reading server preface: EOF"` when the
      coordinator address comes from mDNS auto-discovery or from the
      `--coordinator` flag, even though the address is TCP-reachable
      (`nc` succeeds) and the same coordinator works fine when the worker
      or `hgbuild build` connect to it. Workaround: set the
      `HG_COORDINATOR` env var instead — that path works. Needs
      investigation into `getCoordinatorAddress()` / flag parsing in
      `cmd/hgbuild/main.go` for the `cc`/`c++`/`make` subcommands
      specifically (the `build` subcommand is unaffected apart from not
      reading env at all — see below).

- [ ] **`hgbuild build` doesn't read `HG_COORDINATOR`; `hgbuild cc`/`make`
      don't set `build_id` on the `CompileRequest`.**
      `cmd/hgbuild/main.go`'s `build` subcommand requires `--coordinator`
      explicitly (mDNS/env fallback isn't wired for it), while the
      `cc`/`c++`/`make` compile path never sets `CompileRequest.BuildId`
      from `taskid.BuildSessionID()`/`HG_BUILD_ID` — despite a code
      comment claiming it does (`cmd/hgbuild/main.go:326` area). Net
      effect: per-file compiles via `make`/`cc` never group into a
      "build" the coordinator/dashboard can show under
      `/api/v1/builds` — they only ever show as individual ungrouped
      tasks. Only `hgbuild build file1 file2 ...` (with `-C` passed
      explicitly) produces a real grouped build today.

- [ ] **Worker ID collides when running 2+ workers on the same host.**
      `cmd/hg-worker/main.go:204` sets
      `caps.WorkerId = fmt.Sprintf("worker-%s", hostname)` — hostname-only,
      no port/PID/random suffix, and there's no `--worker-id` override
      flag. Starting a second `hg-worker` on the same machine (different
      `--port`/`--advertise-address`) silently overwrites the first
      worker's coordinator registry entry instead of registering as a
      second node — confirmed live: `/api/v1/workers` stayed at
      `count: 1` and just swapped `address` to the newer process's port.
      Only affects multi-worker-per-host setups (common for local dev/demo
      and bare-metal boxes running several worker processes); a container
      per worker (each with its own hostname) sidesteps it, as would
      Kubernetes pods. Fix: include port (or a random suffix / explicit
      `--worker-id` flag) in the generated ID when not overridden.

## Dashboard follow-ups (nice-to-have, not blocking)

- [ ] `hg-dashboard` only accepts `--coordinator`/`--insecure` via CLI
      flags, not env vars (`HG_COORDINATOR`, `HG_INSECURE`), unlike
      `hg-worker`/`hgbuild` which accept both. Would make the
      `docker-compose.yml` `dashboard` service config cleaner.
- [ ] No Kubernetes manifests yet (Deployment/Service for
      `hg-coord`/`hg-worker`/`hg-dashboard`) — only `docker-compose.yml`
      for LAN/dev. Needed if running the fleet across a real cluster
      network instead of mDNS-discoverable LAN.
