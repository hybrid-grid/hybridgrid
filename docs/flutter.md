# Flutter Builds

> Distributed Flutter Android compilation via Hybrid-Grid workers.
>
> **Last Updated:** 2026-03-20 | **Version:** 0.4.0

---

## Table of Contents

1. [Overview](#overview)
2. [Requirements](#requirements)
3. [Docker Setup](#docker-setup)
4. [CLI Usage](#cli-usage)
5. [Build Modes and Flavors](#build-modes-and-flavors)
6. [Cache Behavior](#cache-behavior)
7. [Limitations](#limitations)
8. [Troubleshooting](#troubleshooting)

---

## Overview

Hybrid-Grid v0.4.0 adds distributed Flutter Android compilation. Build APKs and Android App Bundles across your worker pool, dramatically reducing build times for large Flutter projects.

### Supported Targets

| Target | Status |
|--------|--------|
| Android APK | Working |
| Android App Bundle (AAB) | Working |
| iOS | Not supported |
| Web | Not supported |
| Desktop | Not supported |

---

## Requirements

### Coordinator and Workers

- **Coordinator**: `hg-coord serve` running (on any machine)
- **Worker**: Flutter worker with Android SDK must be connected

### Flutter Project

A valid Flutter project with `pubspec.yaml` in the project root:

```
my_flutter_app/
  pubspec.yaml
  lib/
    main.dart
  android/
    app/
      build.gradle
```

### Source Project Structure

The project directory is archived and sent to workers. Ensure:
- `pubspec.yaml` exists and is valid
- `android/` directory contains Gradle build configuration
- No extremely large files (>100MB) in the project root

---

## Docker Setup

### Pre-built Image

The simplest way to run a Flutter worker:

```bash
# Pull the pre-built image
docker pull ghcr.io/h3nr1-d14z/hybridgrid/flutter-android:3.19.6

# Or use docker-compose
docker compose -f docker-compose.flutter.yml up -d
```

### Docker Compose Configuration

```yaml
# docker-compose.flutter.yml
services:
  flutter-worker:
    build:
      context: .
      dockerfile: build/docker/flutter/Dockerfile
      args:
        FLUTTER_VERSION: "3.19.6"
    image: hybridgrid/flutter-android:3.19.6
    container_name: hg-flutter-worker
    environment:
      - HG_LOG_LEVEL=debug
      - FLUTTER_ROOT=/opt/flutter
      - ANDROID_SDK_ROOT=/opt/flutter/bin/cache/android-sdk
      - ANDROID_HOME=/opt/flutter/bin/cache/android-sdk
      - JAVA_HOME=/usr/lib/jvm/java-17-openjdk-arm64
    networks:
      - hybridgrid

networks:
  hybridgrid:
    external: true
```

### Custom Flutter Version

To build with a different Flutter version:

```bash
docker build \
  -f build/docker/flutter/Dockerfile \
  --build-arg FLUTTER_VERSION=3.24.0 \
  -t my-flutter-worker .
```

### Image Contents

The Docker image includes:

| Component | Version | Location |
|-----------|---------|----------|
| Flutter SDK | 3.19.6 (default) | `/opt/flutter` |
| Android SDK | Pre-bundled | `/opt/flutter/bin/cache/android-sdk` |
| Gradle | 8.5 | `/opt/gradle` |
| Java | 17 | `/usr/lib/jvm/java-17-openjdk-arm64` |

### Starting the Worker

```bash
# Start Flutter worker (connects to coordinator on hybridgrid network)
docker compose -f docker-compose.flutter.yml up -d

# Check logs
docker compose -f docker-compose.flutter.yml logs -f

# Stop worker
docker compose -f docker-compose.flutter.yml down
```

---

## CLI Usage

### Command Structure

```
hgbuild flutter build <apk|appbundle> [flags]
```

### Required Flags

| Flag | Description |
|------|-------------|
| `--project` | Path to Flutter project directory |

### Optional Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--build-mode` | `release` | Build mode: `debug`, `profile`, `release` |
| `--flavor` | (none) | Build flavor (e.g., `staging`, `production`) |
| `--output`, `-o` | (none) | Save the built APK/AAB files into this directory |
| `--distribute-abis` | `false` | APK only: build each ABI as its own task, in parallel on different workers (produces split APKs, not a universal APK) |
| `--abis` | `arm64,arm,x64` | ABIs for `--distribute-abis` |

### Examples

#### Debug APK

```bash
hgbuild flutter build apk --project /path/to/flutter/app --build-mode debug
```

#### Release APK with Flavor

```bash
hgbuild flutter build apk --project /path/to/flutter/app --build-mode release --flavor staging
```

#### Profile App Bundle

```bash
hgbuild flutter build appbundle --project /path/to/flutter/app --build-mode profile
```

#### Release App Bundle with Flavor

```bash
hgbuild flutter build appbundle --project /path/to/flutter/app --build-mode release --flavor production
```

### Full Example

```bash
# 1. Start coordinator (if not running)
hg-coord serve

# 2. Start Flutter worker (in another terminal or via docker compose)
docker compose -f docker-compose.flutter.yml up -d

# 3. Wait for worker to register
sleep 5
hgbuild workers

# 4. Build your Flutter app
hgbuild flutter build apk --project ~/projects/my_app --build-mode release

# 5. Check for cached build on second run
hgbuild flutter build apk --project ~/projects/my_app --build-mode release
# Output: Build completed (cache hit)
```

---

## Build Modes and Flavors

### Build Modes

| Mode | Flags | Use Case |
|------|-------|----------|
| `debug` | `--debug` | Development with debugging symbols |
| `profile` | `--profile` | Performance profiling with tracing |
| `release` | `--release` | Production builds (default) |

### Flavors

Flavors map to Gradle product flavors defined in `android/app/build.gradle`:

```groovy
android {
    flavorDimensions "environment"
    productFlavors {
        staging {
            dimension "environment"
            applicationIdSuffix ".staging"
            versionNameSuffix "-staging"
        }
        production {
            dimension "environment"
            // No suffix for production
        }
    }
}
```

### Combined Build Mode and Flavor

The `--build-mode` and `--flavor` flags combine to produce:
- `hgbuild flutter build apk --build-mode release --flavor staging` runs `flutter build apk --release --flavor staging`

---

## Cache Behavior

### How Caching Works

Flutter builds cache based on source archive hash:

1. **First build (cache miss)**
   - Project source is archived (tar+zstd)
   - Hash computed from archive content
   - Worker runs `flutter pub get` then `flutter build <type> --<mode>`
   - Resulting APK/AAB returned and cached locally

2. **Second build (cache hit)**
   - Same source archive hash computed
   - If hash matches cached entry, return cached artifact immediately
   - No compilation performed

3. **Source change (cache miss)**
   - Any change to project files produces new hash
   - Full rebuild required

### Cache Key Components

The cache key is derived from:
- Source archive hash (all project files)
- Build mode (debug/profile/release)
- Output type (apk/appbundle)
- Flavor (if specified)

### Cache Location

Artifacts are cached in `~/.hybridgrid/cache/` alongside C/C++ build artifacts.

### Verifying Cache Hits

Use verbose mode to see cache behavior:

```bash
hgbuild -v flutter build apk --project ~/projects/my_app

# First run shows:
# [remote] Building Flutter APK (worker-1)

# Second run shows:
# [cache]  Flutter APK cache hit
```

### Shared Gradle Build Cache

The cache above only helps when the source is unchanged. After a Dart
edit, Gradle would normally redo its Kotlin/Java compile, dexing, R8 and
resource merging on whatever worker gets the build. The coordinator can
serve a Gradle remote build cache so those task outputs are reused by every
worker:

```bash
# Coordinator: on by default, served on the HTTP port at /cache/gradle/
hg-coord serve --gradle-cache-dir ~/.cache/hybridgrid/gradle-cache \
  --gradle-cache-max-total-mb 5120

# Workers: point Flutter builds at it (or set HG_GRADLE_CACHE_URL)
hg-worker serve --coordinator coord:9000 \
  --gradle-cache-url http://coord:8080/cache/gradle/
```

The worker only edits its temporary copy of the project (`android/gradle.properties`
and `android/settings.gradle[.kts]`), never `~/.gradle`. With `--token`, the
cache requires it as the HTTP password; the token reaches Gradle through an
environment variable and is never written to disk. `GET /cache/gradle/`
returns hit/miss/store counters as JSON.

Measured on the demo app (release APK, Dart file changed between builds):

| Build | Time |
|-------|------|
| Cold, cache empty | 40.0s |
| Same worker, warm cache | 22.9s |
| New worker with empty `~/.gradle` build cache, remote cache warm (25/26 Gradle tasks from cache) | 27.0s |

The Dart compile itself (`compileFlutterBuild*`) is not cacheable by Gradle,
so it still runs on every source change.

---

## Limitations

### Android Only

v0.4.0 supports only Android targets:
- No iOS toolchain (no Xcode)
- No web support
- No desktop support (Windows/macOS/Linux)

### No Signing

The Docker image does not include signing keys. Release builds will:
- Compile successfully
- Produce unsigned APKs/AABs

To sign builds:
1. Extract artifacts from Hybrid-Grid cache
2. Sign manually with `apksigner` or `jarsigner`

### Build Isolation

Each build runs in an isolated temporary directory:
- `flutter pub get` runs fresh each build
- Gradle daemon is stopped after each build
- No shared state between builds

### Large Output Logs

Flutter builds produce significant log output. If logs exceed 256KB, output is truncated with `[output truncated]`.

---

## Troubleshooting

### "no workers match requirements"

Flutter workers may not be available. Check:

```bash
# List available workers
hgbuild workers -v

# Verify worker has Flutter capability
curl http://localhost:8081/api/v1/workers | jq '.workers[] | select(.build_types | contains(["FLUTTER"]))'
```

### "project not found"

The `--project` path must point to a directory containing `pubspec.yaml`:

```bash
# Wrong - points to parent
hgbuild flutter build apk --project ~/projects

# Correct - points to Flutter project root
hgbuild flutter build apk --project ~/projects/my_app
```

### Build Fails with "unsupported Gradle project"

This usually means dependencies are missing. Ensure:
1. `flutter pub get` runs successfully in the project
2. `android/app/build.gradle` exists and is valid
3. Android SDK licenses are accepted

### Build Hangs

Flutter builds can take 5-10 minutes for first builds. If a build appears hung:
1. Check worker logs: `docker compose -f docker-compose.flutter.yml logs`
2. Check coordinator logs: `hg-coord serve` output
3. Use `docker compose -f docker-compose.flutter.yml logs flutter-worker | grep -E "(flutter|gradle)"` for Flutter-specific logs

### "exit code 1" from flutter pub get

Check network connectivity from worker to Flutter pub server:

```bash
docker compose -f docker-compose.flutter.yml exec flutter-worker flutter pub get
```

Common issues:
- No network access to `pub.dev`
- Missing DNS resolution
- Corporate proxy blocking

### gRPC Message Size Error

If you see `grpc: received message larger than max` errors:
- This indicates a very large Flutter build artifact
- The gRPC max message size is 512MB by default
- If artifacts exceed this, builds will fail

---

## Testing

### E2E Test

Run the Flutter E2E test to verify the full pipeline:

```bash
go test -v -run TestFlutterE2E ./test/e2e/...
```

Expected output: Test passes in approximately 8-10 minutes on first run, faster on subsequent runs due to cache.

---

## See Also

- [Quick Start Guide](./quick-start.md) - General Hybrid-Grid quick start
- [Feature Guide](./feature-guide.md) - Full feature documentation
- [System Architecture](./system-architecture.md) - Design and components
