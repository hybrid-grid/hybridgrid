# TODO

Single list of open work. Items were re-verified against the code on 2026-11-28;
the former `CHECKLIST.md` and `plan.md` were folded into this file (the
icecream benchmark plan lives in `docs/benchmark-icecc-plan.md`, the v0.3.0
design in `docs/v030-plan.md`, plugin design notes in
`docs/plugins-design-notes.md`).

## Backend bugs (found while demoing the dashboard, 2026-09-18)

- [ ] **`hgbuild cc`/`c++`/`make`: remote compile via mDNS-discovered or
      `--coordinator` address may still fail with `Unavailable ... error
      reading server preface: EOF`.**
      Original report: the address was TCP-reachable (`nc` succeeds) and the
      same coordinator worked for workers and `hgbuild build`, but remote
      compiles failed unless the `HG_COORDINATOR` env var was used instead.
      Partly fixed: these subcommands run with `DisableFlagParsing`, and the
      wrapper now parses and strips `--coordinator`/`-C` itself
      (`filterHgbuildFlags` in `cmd/hgbuild/main.go`), where before the flag
      was silently ignored. Still unconfirmed: whether the preface-EOF failure
      itself is gone (it points at a transport/TLS or wrong-port mismatch, e.g.
      a plaintext/TLS mismatch or an mDNS-advertised address that is not the
      gRPC port). Needs a live repro with `--coordinator` and with mDNS
      discovery before closing.

- [ ] **Worker ID still defaults to `worker-<hostname>`, so 2+ workers on one
      host collide unless `--worker-id` is set.**
      Partly fixed: `hg-worker serve` now has a `--worker-id` flag
      (`cmd/hg-worker/main.go`). The default is unchanged (hostname only, no
      port/PID/random suffix), so a second `hg-worker` on the same machine
      silently overwrites the first one's coordinator registry entry unless
      the operator passes distinct IDs. Fix idea: derive the default from
      hostname plus port (or a random suffix) when `--worker-id` is not given.
      Containers (own hostname) and Kubernetes pods are unaffected.

## Dashboard follow-ups (nice-to-have, not blocking)

- [ ] `hg-dashboard` only accepts `--coordinator`/`--insecure` via CLI flags,
      not env vars (`HG_COORDINATOR`, `HG_INSECURE`), unlike
      `hg-worker`/`hgbuild`. Verified: `cmd/hg-dashboard/main.go` reads only
      `LOG_LEVEL` from the environment. Would make the `docker-compose.yml`
      `dashboard` service config cleaner.

## Deployment / release

- [ ] No Kubernetes manifests or Helm chart (Deployment/Service for
      `hg-coord`/`hg-worker`/`hg-dashboard`); only `docker-compose.yml` for
      LAN/dev. Needed to run the fleet across a real cluster network instead
      of an mDNS-discoverable LAN. Verified: no Chart.yaml or k8s manifests
      in the repo.
- [ ] Homebrew formula/tap. The `brews:` block in `.goreleaser.yaml` is
      commented out until the `HOMEBREW_TAP_TOKEN` secret and tap repo exist.
      Several docs (`docs/getting-started.md`, `docs/quick-start.md`) still
      show `brew` install steps; verify they work or remove them.

## Features not implemented

- [ ] WAN registry (HTTP register/heartbeat/worker-list endpoints and a WAN
      client for workers) so discovery is not LAN/mDNS-only. Design in
      `docs/v030-plan.md`. Verified: coordinator HTTP only serves `/health`,
      `/metrics`, `/log-level`, `/dispatch-count`, `/events` and the API/UI
      routes; no registry endpoints.
- [ ] Distributed cache sync between nodes (only the local CAS and the
      coordinator's per-builder result cache exist).
- [ ] GPU detection in `internal/capability` (needed for Unity/Flutter
      scheduling). Verified: no GPU references in the package.
- [ ] Built-in build types for Cocos, Rust, Go and Node.js. The proto enum
      has `BUILD_TYPE_COCOS/RUST/GO/NODEJS`, but no executor implements them
      beyond capability hints; today they are covered only through the
      builder plugin system (`docs/plugins.md`). Decide whether to ship
      first-party builders or document plugins as the supported path.

## Quality

- [ ] Reach 80% overall test coverage. Measured 2026-11-28: executor 69.2%,
      capability 60.7%, cli/build 79.0%, cli/output 93.8%; overall figure not
      re-measured.
