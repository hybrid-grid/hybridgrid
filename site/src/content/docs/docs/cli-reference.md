---
title: CLI reference
description: Every hgbuild subcommand, what it talks to, and when to reach for it.
---

All subcommands share these persistent flags: `--coordinator`/`-C` (address,
auto-discovers via mDNS if empty), `--insecure` (default `true`), `--no-fallback`,
`--timeout`, `--tls-cert`/`--tls-key`/`--tls-ca`, and `--tracing-*` for
OpenTelemetry.

:::note
The coordinator address is resolved in this order: `--coordinator`/`-C`, the
`HG_COORDINATOR` environment variable, then mDNS auto-discovery. `hgbuild cc`
and `hgbuild c++` run with flag parsing disabled and parse `--coordinator`/`-C`
(plus `--timeout`, `--arch`, `--insecure`, `--no-fallback`, `-v`) themselves
before handing the rest to the compiler. `hgbuild make`, `ninja`, and `wrap`
pass their arguments straight through to the wrapped tool (`make -C dir` is
make's own flag), so for those set `HG_COORDINATOR` in the environment.
:::

:::caution
If `cc`/`c++`/`make` still fail with `error reading server preface: EOF`, see
[Troubleshooting](/docs/troubleshooting/) — that symptom has not been confirmed
fixed.
:::

## Compiling

**`hgbuild cc [flags] [files...]`** / **`hgbuild c++ ...`**
Drop-in `gcc`/`g++` replacement. Preprocesses locally, checks the local cache,
sends the translation unit to the coordinator, falls back to local compile on
any coordinator/network failure (unless `--no-fallback`). The request carries a
`build_id` only when `HG_BUILD_ID` is set in the environment (`hgbuild make`
does this for you); a lone `cc` call is not grouped into a build.

**`hgbuild make [make-args...]`** / **`hgbuild ninja [ninja-args...]`**
Wraps the build tool: sets `CC`/`CXX` to route through `hgbuild cc`/`c++` so
every translation unit in the build gets distributed, and exports one
`HG_BUILD_ID` so all of them are grouped as a single build on the dashboard.
Use `-v` for `[cache]`/`[remote]`/`[local]` per-file status.

**`hgbuild wrap <command> [args...]`**
Generic version of `make`/`ninja` — wraps any build command that respects
`CC`/`CXX` env vars (also exports `HG_BUILD_ID`).

**`hgbuild build [files...]`**
Submits files directly as one build session (sets `build_id`, so it shows up
grouped on the dashboard's Builds page). Reads the coordinator address the same
way as the other subcommands (`--coordinator`, `HG_COORDINATOR`, mDNS). Flags: `--compiler`, `--args` (compiler args), `--arch`, `-o`/`--output`,
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
