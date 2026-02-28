# v0.3.0 Implementation Plan

**Date:** 2026-02-28
**Status:** Draft
**Author:** v03-researcher agent

## Overview

v0.3.0 extends HybridGrid beyond C/C++ compilation to support Flutter and Unity project builds, and expands discovery from LAN-only mDNS to WAN-capable registry. The existing proto definitions (`build.proto`) already include `FlutterConfig`, `UnityConfig`, `FlutterCapability`, `UnityCapability`, `BuildType_BUILD_TYPE_FLUTTER`, and `BuildType_BUILD_TYPE_UNITY`; the `BuildRequest`/`BuildResponse` messages support these via `oneof config`. The task is to implement the coordinator/worker logic and CLI commands to make these work end-to-end, plus a new WAN registry service.

---

## 1. Flutter Build Distribution

### 1.1 How Flutter Builds Work

Flutter CLI builds (`flutter build <target>`) produce platform-specific artifacts:

| Target | Command | Output | SDK Needed |
|--------|---------|--------|------------|
| Android APK | `flutter build apk` | `build/app/outputs/flutter-apk/app-release.apk` | Android SDK, JDK |
| Android AAB | `flutter build appbundle` | `build/app/outputs/bundle/release/app-release.aab` | Android SDK, JDK |
| iOS | `flutter build ios` | `build/ios/iphoneos/Runner.app` | Xcode (macOS only) |
| Web | `flutter build web` | `build/web/` directory | None extra |
| Linux | `flutter build linux` | `build/linux/x64/release/bundle/` | GTK dev libs |
| Windows | `flutter build windows` | `build/windows/x64/runner/Release/` | Visual Studio |
| macOS | `flutter build macos` | `build/macos/Build/Products/Release/` | Xcode (macOS only) |

**Key insight:** Unlike C/C++ compilation, Flutter builds are monolithic -- the entire project must be present on the worker. Individual Dart files cannot be compiled independently due to Dart's AOT compilation pipeline and tree-shaking.

### 1.2 Parallelization Opportunities

Flutter builds themselves are not easily parallelizable at the compilation unit level (unlike C/C++ object files). However, distribution is valuable for:

1. **Multi-platform builds** -- Build Android, iOS, Web, Linux, Windows, macOS in parallel on different workers, each having the required platform SDK.
2. **Multi-flavor builds** -- Build `dev`, `staging`, `production` flavors concurrently on different workers.
3. **Offloading** -- Move long build (3-10 min) off developer laptop to powerful build farm.
4. **Android Gradle parallelism** -- Workers can be configured with `org.gradle.parallel=true` and `org.gradle.workers.max=N` in `gradle.properties` for internal parallelism.

### 1.3 Worker Requirements

Workers supporting Flutter builds need:

- Flutter SDK (version-matched to project's `.fvmrc` or `pubspec.yaml` constraint)
- For Android: Android SDK, JDK 17+
- For iOS/macOS: Xcode (macOS only, cannot run on Linux/Windows)
- For Web: Chrome (for testing), no extra SDK for building
- For Linux: `libgtk-3-dev`, `ninja-build`, `libstdc++`
- For Windows: Visual Studio 2022 with "Desktop development with C++"

**Current capability detection** (`internal/capability/detect.go:330-384`) already detects Flutter SDK, Android SDK, Xcode availability, and reports supported platforms. This is sufficient.

### 1.4 Build Flow

```
hgbuild flutter build apk --release
    |
    +-> 1. Parse CLI args, detect project root (pubspec.yaml)
    +-> 2. Archive project (tar.zst, exclude build/, .dart_tool/)
    +-> 3. Compute source_hash from pubspec.lock + lib/**/*.dart
    +-> 4. Check cache (keyed by hash + platform + flavor + mode)
    +-> 5. Send BuildRequest{build_type=FLUTTER, flutter_config={...}}
    +-> 6. Coordinator selects worker with Flutter capability + platform
    +-> 7. Worker extracts archive, runs `flutter build <target>`
    +-> 8. Worker archives output artifacts, returns BuildResponse
    +-> 9. Client extracts artifacts to local build/ directory
    +-> 10. Store artifacts in cache
```

For large projects (>10MB), use `StreamBuild` to stream the source archive.

### 1.5 Proto Changes Needed

The existing proto is **sufficient**. No new messages needed. Specifically:
- `FlutterConfig` has `flutter_version`, `build_mode`, `flavor`, `dart_defines`
- `FlutterCapability` has `sdk_version`, `platforms`, `android_sdk`, `xcode_available`
- `BuildRequest.flutter_config` already wired
- `BuildResponse.artifacts` + `artifact_list` supports returning APK/AAB/IPA files

**Minor addition suggested:** Add `target` field to `FlutterConfig` (apk, appbundle, ios, web, linux, windows, macos) since `TargetPlatform` enum doesn't distinguish APK vs AAB.

```protobuf
message FlutterConfig {
  string flutter_version = 1;
  string build_mode = 2;            // debug, profile, release
  string flavor = 3;
  map<string, string> dart_defines = 4;
  string target = 5;                // NEW: apk, appbundle, ios, web, linux, windows, macos
  repeated string extra_args = 6;   // NEW: pass-through args to flutter build
}
```

### 1.6 New Packages/Modules

| Package | Purpose |
|---------|---------|
| `internal/worker/executor/flutter.go` | Flutter build executor (extract archive, run `flutter build`, package artifacts) |
| `internal/cli/flutter/flutter.go` | CLI handler for `hgbuild flutter build` commands |
| `internal/cli/flutter/archive.go` | Project archiving with smart exclusion (.dart_tool, build/, ios/Pods/) |

### 1.7 Complexity Estimate

| Component | Effort |
|-----------|--------|
| Flutter executor (worker side) | Medium -- extract tar.zst, run flutter CLI, collect artifacts |
| CLI command parsing | Low -- similar to existing `hgbuild make/ninja` |
| Project archiving | Medium -- smart exclusion, streaming for large projects |
| Cache key generation for Flutter | Low -- hash pubspec.lock + lib/**/*.dart |
| Coordinator routing (already works) | None -- `matchesCapability` already handles `BUILD_TYPE_FLUTTER` |
| Testing | Medium -- mock Flutter SDK, verify archive/extract |

---

## 2. Unity Build Distribution

### 2.1 How Unity Builds Work

Unity command-line builds use:

```bash
Unity -batchmode -nographics -quit \
  -projectPath /path/to/project \
  -executeMethod BuildScript.PerformBuild \
  -buildTarget StandaloneWindows64 \
  -logFile -
```

Key characteristics:
- **Requires Unity Editor license** per machine (Personal, Plus, Pro, or Build Server license)
- **Build Server licensing** allows floating headless build entitlements for CI/CD
- **`-executeMethod`** invokes a static C# method defined in `Assets/Editor/` scripts
- **Project must be fully present** on the build machine (Assets, Packages, ProjectSettings)
- **Build targets:** StandaloneWindows64, StandaloneOSX, StandaloneLinux64, Android, iOS, WebGL

### 2.2 Parallelization Opportunities

1. **Multi-platform builds** -- Build Windows, macOS, Linux, Android, iOS, WebGL on different workers simultaneously.
2. **Asset pipeline offloading** -- Unity's asset import (textures, models, shaders) is CPU-intensive and benefits from powerful workers.
3. **IL2CPP builds** -- IL2CPP scripting backend (C++ code generation from C#) is very slow; offloading to multi-core workers cuts time significantly.
4. **Addressable asset bundles** -- Different asset groups can potentially be built on different workers.

### 2.3 Worker Requirements

- Unity Editor installed (specific version matching project)
- **License activated** (floating Build Server license ideal for farms)
- For Android: Android SDK, JDK, NDK
- For iOS: Xcode (macOS only)
- For Windows: Visual Studio (for IL2CPP)
- Sufficient disk space (Unity projects + Library cache can be 10-50GB)

### 2.4 Build Flow

```
hgbuild unity build --target StandaloneWindows64 --method BuildScript.Build
    |
    +-> 1. Parse CLI args, detect project root (Assets/ + ProjectSettings/)
    +-> 2. Archive project (tar.zst, exclude Library/, Temp/, Logs/)
    +-> 3. Compute source_hash from Assets/**/* + Packages/manifest.json
    +-> 4. Check cache (keyed by hash + target + scripting_backend)
    +-> 5. Send BuildRequest{build_type=UNITY, unity_config={...}}
    +-> 6. Coordinator selects worker with matching Unity version + target
    +-> 7. Worker extracts archive, runs Unity -batchmode ...
    +-> 8. Worker archives build output, returns BuildResponse
    +-> 9. Client extracts artifacts
    +-> 10. Store in cache
```

**Important:** Unity projects are large (often 1-50GB). StreamBuild is essential. Consider also a worker-side Library cache (Unity's import cache) keyed by project ID to avoid re-importing assets on every build.

### 2.5 Proto Changes Needed

Existing proto is **mostly sufficient**. Suggested additions:

```protobuf
message UnityConfig {
  string unity_version = 1;
  string build_method = 2;          // e.g., "BuildScript.Build"
  string scripting_backend = 3;     // mono, il2cpp
  map<string, string> extra_args = 4;
  string build_target = 5;          // NEW: StandaloneWindows64, Android, etc.
  repeated string scenes = 6;       // NEW: specific scenes to include
}

message UnityCapability {
  repeated string versions = 1;
  string license_type = 2;          // pro, personal, build_server
  repeated TargetPlatform build_targets = 3;
  bool il2cpp_available = 4;        // NEW: IL2CPP support
}
```

### 2.6 New Packages/Modules

| Package | Purpose |
|---------|---------|
| `internal/worker/executor/unity.go` | Unity build executor (extract, invoke Unity CLI, collect output) |
| `internal/cli/unity/unity.go` | CLI handler for `hgbuild unity build` commands |
| `internal/cli/unity/archive.go` | Project archiving (exclude Library/, Temp/) |

### 2.7 Complexity Estimate

| Component | Effort |
|-----------|--------|
| Unity executor (worker side) | Medium-High -- Unity CLI is finicky, error parsing from logs needed |
| CLI command parsing | Low |
| Project archiving | High -- Unity projects are huge, need streaming + smart exclusion |
| License management | Medium -- Detect license type, handle "license in use" errors |
| Coordinator routing | None -- already works for `BUILD_TYPE_UNITY` checking caps.Unity |
| Library cache on worker | High -- persistent cache per project to avoid re-import |
| Testing | High -- need Unity installed, or extensive mocking |

---

## 3. WAN Registry

### 3.1 Current State: LAN-Only mDNS

Current discovery (`internal/discovery/mdns/`) uses Zeroconf/mDNS:
- **Announcer** (`announcer.go`): Workers broadcast `_hybridgrid._tcp` on LAN
- **Browser** (`browser.go`): Coordinator discovers workers via mDNS browse
- **CoordAnnouncer**: Coordinator broadcasts `_hybridgrid-coord._tcp` for worker auto-connect
- **Limitation:** mDNS only works on local network segment (same subnet)

### 3.2 Design Goals

1. Workers on different networks (home, office, cloud VMs) can join a build cluster
2. Secure authentication (tokens, TLS) -- don't accept arbitrary workers
3. NAT traversal for workers behind routers
4. Graceful degradation -- LAN mDNS continues to work alongside WAN
5. Low operational overhead -- should not require running extra infrastructure

### 3.3 Architecture Options

| Option | Pros | Cons |
|--------|------|------|
| **A. Central Registry Service** | Simple, reliable, NAT-friendly | Single point of failure, requires hosting |
| **B. DHT (Kademlia)** | Fully decentralized, no SPOF | Complex, slow lookup, NAT issues |
| **C. Cloud Relay (TURN-like)** | Works through any NAT | High latency, bandwidth cost |
| **D. Hybrid: Registry + Direct Connect** | Best of A + direct when possible | Medium complexity |

**Recommended: Option D -- Hybrid Registry + Direct Connect**

### 3.4 Detailed Architecture

```
                    ┌──────────────────┐
                    │  WAN Registry    │
                    │  (hg-registry)   │
                    │                  │
                    │ - Worker list    │
                    │ - Health tracking│
                    │ - Token auth     │
                    │ - NAT info       │
                    └────────┬─────────┘
                         gRPC/TLS
              ┌──────────────┼──────────────┐
              │              │              │
    ┌─────────▼──┐  ┌───────▼────┐  ┌──────▼─────┐
    │ Coordinator│  │ WAN Worker │  │ WAN Worker │
    │  (Office)  │  │  (Cloud)   │  │  (Home)    │
    │            │  │            │  │            │
    │ + LAN      │  └────────────┘  └────────────┘
    │   Workers  │
    └────────────┘
```

**Flow:**
1. WAN Workers register with the Registry (gRPC over TLS, token auth)
2. Coordinator queries Registry for available WAN workers
3. For each build task, coordinator tries LAN workers first (lower latency)
4. If no LAN worker available/capable, falls back to WAN workers via Registry
5. Coordinator connects directly to WAN worker's public address for build requests
6. If direct connection fails (NAT), falls through to relay mode

### 3.5 NAT Traversal Strategy

**Tier 1: Direct connection** -- Worker has public IP or port forwarding configured. Worker reports public address to registry. Coordinator connects directly.

**Tier 2: STUN-assisted** -- Worker behind NAT uses STUN to discover its public IP:port. Reports to registry. Works for Full Cone and Address-Restricted NAT.

**Tier 3: Relay** -- Worker behind Symmetric NAT. Traffic relays through the Registry service. Higher latency but always works. Only used for control plane; large source archives should be avoided for relay-only workers (or use chunked transfer with backpressure).

### 3.6 Registry Service Design

New standalone service: `hg-registry`

```protobuf
// New proto file: proto/hybridgrid/v1/registry.proto

service RegistryService {
  // Worker registration
  rpc RegisterWorker(RegisterWorkerRequest) returns (RegisterWorkerResponse);

  // Worker heartbeat (keep-alive)
  rpc Heartbeat(HeartbeatRequest) returns (HeartbeatResponse);

  // Worker deregistration
  rpc DeregisterWorker(DeregisterWorkerRequest) returns (DeregisterWorkerResponse);

  // Coordinator queries for available workers
  rpc ListWorkers(ListWorkersRequest) returns (ListWorkersResponse);

  // Query workers by capability
  rpc FindWorkers(FindWorkersRequest) returns (FindWorkersResponse);

  // NAT traversal assistance
  rpc GetConnectionInfo(GetConnectionInfoRequest) returns (GetConnectionInfoResponse);
}

message RegisterWorkerRequest {
  string auth_token = 1;
  WorkerCapabilities capabilities = 2;
  string public_address = 3;       // Worker's public address if known
  string private_address = 4;      // Worker's private/LAN address
  NATType nat_type = 5;            // Detected NAT type
}

enum NATType {
  NAT_UNKNOWN = 0;
  NAT_NONE = 1;                    // Public IP, no NAT
  NAT_FULL_CONE = 2;
  NAT_ADDRESS_RESTRICTED = 3;
  NAT_PORT_RESTRICTED = 4;
  NAT_SYMMETRIC = 5;
}
```

### 3.7 Security

- **Token-based authentication:** Cluster invite tokens (short-lived, rotatable)
- **mTLS:** Registry <-> Worker, Registry <-> Coordinator, Coordinator <-> Worker
- **Worker verification:** Registry verifies worker identity via challenge-response
- **Rate limiting:** Protect registry from abuse

Existing `internal/security/` package has TLS config helpers that can be extended.

### 3.8 Integration with Existing Code

The `WorkerInfo.DiscoverySource` field already supports `"mdns"` and `"wan"`. The scheduler (`scheduler.go:314`) already gives LAN workers a +20 score bonus (`ScoreLANSource`). WAN workers automatically get lower priority, which is correct behavior (prefer low-latency LAN workers).

Changes needed:
- `internal/discovery/wan/` -- New package for WAN registry client
- `internal/discovery/wan/client.go` -- gRPC client to registry
- `internal/discovery/wan/nat.go` -- NAT type detection (STUN)
- Coordinator merges LAN (mDNS) + WAN (registry) workers into single registry
- Worker optionally registers with WAN registry in addition to LAN mDNS

### 3.9 New Packages/Modules

| Package | Purpose |
|---------|---------|
| `cmd/hg-registry/` | WAN registry service binary |
| `internal/registry/service/` | Registry gRPC service implementation |
| `internal/registry/store/` | Worker store (in-memory + optional persistence) |
| `internal/registry/auth/` | Token management, mTLS validation |
| `internal/discovery/wan/client.go` | WAN registry client (used by workers + coordinator) |
| `internal/discovery/wan/nat.go` | NAT type detection |
| `proto/hybridgrid/v1/registry.proto` | Registry service proto definition |

### 3.10 Complexity Estimate

| Component | Effort |
|-----------|--------|
| Registry proto + codegen | Low |
| Registry service (core CRUD + heartbeat) | Medium |
| Token auth + mTLS integration | Medium -- extends existing security/ package |
| NAT type detection | Medium-High -- STUN integration, multiple NAT types |
| Relay mode | High -- bidirectional gRPC streaming for relay |
| WAN client for workers | Medium |
| WAN client for coordinator | Medium -- merge LAN + WAN sources |
| Coordinator scoring (already done) | None -- ScoreLANSource already differentiates |
| `hg-registry` binary + CLI | Low |
| Testing | High -- need to test NAT scenarios |

---

## 4. Implementation Order

### Phase 1: Flutter Builds (2-3 weeks)

Priority: High -- Most requested feature, simpler than Unity.

1. Add `target` and `extra_args` to `FlutterConfig` proto, regenerate
2. Implement `internal/worker/executor/flutter.go`
3. Implement `internal/cli/flutter/` (CLI parsing, project archiving)
4. Wire up coordinator `Build()` method to forward Flutter requests
5. Tests with mock Flutter SDK + integration test with real Flutter project
6. Update `hgbuild` CLI to accept `hgbuild flutter build <target>` commands

### Phase 2: Unity Builds (2-3 weeks)

Priority: Medium -- Valuable but complex (licensing, large projects).

1. Add `build_target`, `scenes`, `il2cpp_available` to proto, regenerate
2. Implement `internal/worker/executor/unity.go`
3. Implement `internal/cli/unity/` (CLI parsing, project archiving with streaming)
4. Worker-side Library cache management
5. License detection and error handling
6. Tests (extensive mocking required)

### Phase 3: WAN Registry (3-4 weeks)

Priority: Medium -- Extends reach of the system, but LAN works for thesis demo.

1. Define `registry.proto`, generate code
2. Implement `internal/registry/service/` (core service)
3. Implement `internal/registry/auth/` (tokens, mTLS)
4. Implement `internal/discovery/wan/client.go` (worker + coordinator side)
5. Implement `cmd/hg-registry/` binary
6. NAT detection (`internal/discovery/wan/nat.go`)
7. Relay mode (can be deferred to v0.3.1 if time-constrained)
8. Integration tests

### Phase 4: Integration + Polish (1 week)

1. Dashboard updates (show Flutter/Unity builds, WAN workers)
2. Documentation updates
3. End-to-end integration tests
4. Release prep

---

## 5. Dependencies Between Features

```
Flutter Builds ──> (independent, can start immediately)
Unity Builds   ──> (independent, can start immediately)
WAN Registry   ──> (independent, can start immediately)

Flutter Builds ──> shares archiving logic with ──> Unity Builds
                   (extract common archive package first)

WAN Registry   ──> uses existing ──> security/tls package
               ──> uses existing ──> coordinator/registry interface
```

All three features are independent and can be developed in parallel. The shared archiving logic between Flutter and Unity suggests implementing Flutter first, extracting the archive utilities, then reusing them for Unity.

---

## 6. Risks and Open Questions

1. **Flutter version pinning** -- Projects may use FVM or `.flutter-version`. Workers need version management or Docker containers with pre-installed Flutter versions. Consider Docker-based Flutter workers (e.g., `cirrusci/flutter` images).

2. **Unity licensing cost** -- Build Server licenses cost ~$30/seat/month. For thesis, Unity Personal (free) works but has splash screen. Need to document licensing implications clearly.

3. **Unity project size** -- Large Unity projects (10-50GB) may be impractical to transfer over WAN. Worker-side project caching (rsync-style delta sync) should be considered for v0.3.1.

4. **NAT traversal reliability** -- Symmetric NAT (common on mobile hotspots, enterprise networks) requires relay mode. Initial v0.3.0 can require direct connectivity or port forwarding, with relay as stretch goal.

5. **Registry high availability** -- Single registry is a SPOF. For v0.3.0, single instance is acceptable; v0.4.0 could add Raft-based replication.

6. **StreamBuild implementation** -- Current `StreamBuild` on coordinator (`server/grpc.go:365-397`) returns "not implemented." Must be completed for Flutter/Unity (projects >10MB are common).

7. **Worker-to-coordinator Build forwarding** -- Current `Build()` method (`server/grpc.go:351-362`) returns "not implemented." Must be implemented to route Flutter/Unity builds to workers, similar to existing `Compile()` forwarding.
