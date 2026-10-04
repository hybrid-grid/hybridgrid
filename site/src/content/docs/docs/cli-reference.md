---
title: CLI reference
description: Every hgbuild subcommand, what it talks to, and when to reach for it.
---

All subcommands share these persistent flags: `--coordinator`/`-C` (address,
auto-discovers via mDNS if empty), `--insecure` (default `true`), `--no-fallback`,
`--timeout`, `--tls-cert`/`--tls-key`/`--tls-ca`, and `--tracing-*` for
OpenTelemetry.

:::note
`hgbuild cc`/`c++`/`make`/`ninja` currently ignore `--coordinator` in some
paths — set `HG_COORDINATOR` as an environment variable instead if remote
compilation isn't connecting. Tracked in the repo's `TODO.md`.
:::

## Compiling

**`hgbuild cc [flags] [files...]`** / **`hgbuild c++ ...`**
Drop-in `gcc`/`g++` replacement. Preprocesses locally, checks the local cache,
sends the translation unit to the coordinator, falls back to local compile on
any coordinator/network failure (unless `--no-fallback`).

**`hgbuild make [make-args...]`** / **`hgbuild ninja [ninja-args...]`**
Wraps the build tool: sets `CC`/`CXX` to route through `hgbuild cc`/`c++` so
every translation unit in the build gets distributed. Use `-v` for
`[cache]`/`[remote]`/`[local]` per-file status.

**`hgbuild wrap <command> [args...]`**
Generic version of `make`/`ninja` — wraps any build command that respects
`CC`/`CXX` env vars.

**`hgbuild build [files...]`**
Submits files directly as one build session (sets `build_id`, so it shows up
grouped on the dashboard's Builds page — unlike `cc`/`make`, see the note
above). Flags: `--compiler`, `--args` (compiler args), `--arch`, `-o`/`--output`,
`-t`/`--type` (`cpp`, `flutter`, `unity`, `rust`, `go` — only `cpp`/`flutter`/`unity`
have a working executor today).

## Multi-platform builds

**`hgbuild flutter build apk|appbundle --project <path>`**
Distributes a Flutter Android build to a worker with the Flutter SDK +
Android SDK + Gradle (see [Flutter & Unity](/docs/flutter-and-unity/)).
Flags: `--build-mode` (`debug`/`profile`/`release`), `--flavor`.

**`hgbuild unity build <platform> --project <path> --build-method <Class.Method>`**
Runs a real `Unity -batchmode -quit -nographics` build on a worker with
Unity Hub installed. `<platform>`: `android`, `ios`, `windows`, `linux`,
`macos`, `webgl`. Flags: `--unity-version`, `--scripting-backend`
(`mono`/`il2cpp`).

## Fleet inspection

**`hgbuild status`** — coordinator health, active/queued task counts.

**`hgbuild workers`** — list registered workers and their capabilities.

**`hgbuild graph`** — visualize the build dependency graph.

## Cache & config

**`hgbuild cache stats`** / **`hgbuild cache clear`** — local
content-addressable cache (xxhash-keyed) stats and eviction.

**`hgbuild config show`** / **`hgbuild config init`** — print or scaffold
`~/.hybridgrid/config.yaml`.

**`hgbuild version`** — print the client version.
