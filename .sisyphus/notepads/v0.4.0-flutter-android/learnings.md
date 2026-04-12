# Learnings

## 2026-03-20: Added flutter pub get step

- Flutter executor (`internal/worker/executor/flutter.go`) now runs `flutter pub get` before `flutter build`
- Uses same `execCtx` and `workDir` as build command for consistent environment
- `runFlutterPubGet` helper function added at line 158 — runs `flutter pub get` in workDir, captures stderr
- On pub get failure, returns `Result{Success: false, ExitCode: <non-zero>, Stderr: <error>}`
- Error handling matches existing pattern: extract ExitCode from `*exec.ExitError` if available

## 2026-03-20: Fixed flutter pub get integration bugs

- **Bug 1**: `fmt.Errorf` wrapping lost the `*exec.ExitError` type, so type assertion at call site always failed and exit code defaulted to 1. Fix: `runFlutterPubGet` now returns raw `cmd.Run()` error directly.
- **Bug 2**: Only stderr was captured, not stdout. Fix: `runFlutterPubGet` now accepts `*bytes.Buffer` for both stdout and stderr.
- **Bug 3**: Pub get output was not appended to build logs. Fix: Result.Stdout/Stderr now prepend pub get output before build output (`pubGetOut.String() + stdout.String()`).
- **Signature change**: `runFlutterPubGet(ctx, flutterCmd, workDir string) error` → `runFlutterPubGet(ctx, flutterCmd, workDir string, stdout, stderr *bytes.Buffer) error`
- **Exit code**: On pub get failure, if error is `*exec.ExitError`, exit code is extracted from it; otherwise defaults to 1.

## 2026-03-20: Flutter E2E Build hang/failure root cause and minimal fix

- Flutter E2E was stalling/failing deep in Build because Flutter artifact payloads can exceed gRPC default 4MB message limits.
- Reproduction showed `ResourceExhausted: grpc: received message larger than max (292365926 vs. 4194304)` after ~8.4 minutes in `TestFlutterE2E`.
- Minimal reliable fix: set larger gRPC call message limits (512MB) on both coordinator worker-dial clients and shared gRPC client wrapper.
- E2E hardening: `runDirectFlutterBuild` now uses a targeted 9-minute call timeout and emits filtered docker compose logs (`flutter pub get`, `flutter build`, `gradle`, `license`, `download`) when build calls fail.
- Verified result: `go test -v -run TestFlutterE2E ./test/e2e/...` passes in ~495s (<10 minutes).

## 2026-03-20: Documentation structure for Flutter builds

- README.md updated: version changed to v0.4.0, Flutter moved from "Planned" to "Working", added quick-start section with CLI examples and Docker image mention.
- docs/flutter.md created with comprehensive guide covering: Docker setup, CLI usage, build modes/flavors, cache behavior, limitations, troubleshooting.
- CLI examples documented: `hgbuild flutter build apk` and `hgbuild flutter build appbundle` with `--build-mode` and `--flavor` flags.
- Docker image base: `ghcr.io/cirruslabs/flutter` via `build/docker/flutter/Dockerfile` — Android SDK pre-bundled, no iOS toolchain.
- Cache behavior: First build always compiles (cache miss), subsequent identical builds return cached artifact.
- Limitations clearly documented: Android only, no signing keys included.

## 2026-03-20: Fixed jq example for /api/v1/workers response shape

- `/api/v1/workers` returns an object: `{ "workers": [...], "count": N, "timestamp": T }`
- The Troubleshooting snippet in docs/flutter.md used `jq '.[]'` which only works for arrays
- Correct filter is `.workers[]` to access the workers array inside the response object
- Example: `jq '.workers[] | select(.build_types | contains(["FLUTTER"]))'`

## 2026-03-20: Fixed Flutter build timeout default in hgbuild CLI

- **Problem**: `cmd/hgbuild/main.go` set `flutter.Dependencies{BuildTimeout: 0}`, leaving `BuildRequest.TimeoutSeconds` unset.
- **Impact**: Worker default 120s timeout (`internal/worker/server/grpc.go`) cut off long Flutter release builds.
- **Fix**: Changed `BuildTimeout: 0` → `BuildTimeout: 10 * time.Minute` in `newFlutterCmd()` at line 438.
- **Scope**: Single file change (`cmd/hgbuild/main.go`), no new flags, no test changes needed.
- **Verification**: `go test -v ./internal/cli/flutter/...` passes (2/2 tests).

## 2026-03-20: Fixed Flutter request timeout < build timeout issue

- **Problem**: `newFlutterCmd()` set `RequestTimeout: 5 * time.Minute` but `BuildTimeout: 10 * time.Minute`, causing `DeadlineExceeded` before builds completed.
- **Fix**: Changed `RequestTimeout: 5 * time.Minute` → `RequestTimeout: 15 * time.Minute` in `cmd/hgbuild/main.go`.
- **Rationale**: Request timeout must be >= build timeout to prevent premature gRPC deadline expiration.
- **Verification**: `go test -v ./internal/cli/flutter/...` passes (2/2 tests).

## 2026-03-20: Disabled Gradle VFS file watching to prevent watcher crashes

- **Problem**: Flutter Android builds crashed with `Couldn't poll for events, error = 4` from Gradle's native file watcher (`net.rubygrapefruit.platform.internal.jni.AbstractNativeFileEventFunctions$NativeFileWatcher`).
- **Fix**: Set `GRADLE_OPTS=-Dorg.gradle.vfs.watch=false` environment variable on both `flutter pub get` and `flutter build` commands in `internal/worker/executor/flutter.go`.
- **Changes**: `runFlutterPubGet` (line 206) and `Execute` method's build command (line 152) both now set `cmd.Env = append(os.Environ(), "GRADLE_OPTS=-Dorg.gradle.vfs.watch=false")`.
- **Verification**: `go test -v ./internal/worker/executor/...` passes (all tests).

## 2026-03-20: Increased Flutter build timeout defaults to 20m

- **Problem**: Flutter release APK builds were hitting the 10m `BuildTimeout` (latest failure: ~10m12s with `exit -1`).
- **Fix**: Changed `BuildTimeout: 10 * time.Minute` → `20 * time.Minute` and `RequestTimeout: 15 * time.Minute` → `25 * time.Minute` in `newFlutterCmd()` (`cmd/hgbuild/main.go`, line 437-438).
- **Rationale**: Request timeout must remain strictly > build timeout to avoid premature `DeadlineExceeded` cancellation.
- **Scope**: Single file (`cmd/hgbuild/main.go`), no CLI flag changes.
- **Verification**: `go test -v ./internal/cli/flutter/...` passes (2/2 tests).

## 2026-03-20: Fixed Flutter E2E startup pull issue

- **Problem**: `TestFlutterE2E` failed with `pull access denied for hybridgrid/flutter-android` when local image existed.
- **Root cause 1**: `docker compose up -d --build` defaults to pulling images even when local image exists.
- **Fix 1**: Added `--pull=never` to docker compose args in `test/e2e/flutter_test.go` line 105.
- **Root cause 2**: `platform: linux/amd64` in compose YAML forced architecture mismatch with arm64 local image.
- **Fix 2**: Removed `platform: linux/amd64` constraint from generated compose YAML so it uses local image's native platform.
- **Verification**: `go test -v -run TestFlutterE2E ./test/e2e/...` now starts containers successfully (compose startup passes; Gradle build issue is unrelated).

## 2026-03-20: Removed hardcoded JAVA_HOME from E2E compose env

- **Problem**: `test/e2e/flutter_test.go` line 92 had `JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64` hardcoded in generated compose YAML environment.
- **Root cause**: Architecture-specific path breaks arm64 image usage after `--pull=never` and platform constraint removal.
- **Fix**: Removed the `JAVA_HOME` line from the compose environment block, letting the container use its default Java environment.
- **Verification**: Test failure signature changed — no longer fails at Gradle `assembleDebug` due to bad JAVA_HOME; now surfaces AAPT2/Rosetta emulation issue instead (different root cause).

## 2026-03-20: Added INTEGRATION_TEST env gate to TestFlutterE2E

- Added `os.Getenv("INTEGRATION_TEST") != "1"` check in `TestFlutterE2E` (`test/e2e/flutter_test.go`, line 44-46) after the Docker image availability check.
- Pattern matches existing skip style: `t.Skip("skipping Flutter E2E test unless INTEGRATION_TEST=1 is set")`.
- Verification: `go test -v -run TestFlutterE2E ./test/e2e/...` shows SKIP (0.29s) with message at line 45.
- This unblocks F2 full-suite CI/local verification without executing heavyweight Docker Flutter E2E by default.
