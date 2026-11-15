---
title: Flutter & Unity builds
description: Distributing Flutter Android builds and running real Unity batch-mode builds on remote workers.
---

Both of these submit the whole project as a `BuildRequest` (archived,
content-hashed) to the coordinator, which schedules it to a worker
advertising the matching capability — unlike C/C++, they aren't split into
many small remote-compile calls.

## Flutter (Android only)

Requires a worker running the pre-built `hybridgrid/flutter-android` Docker
image (Flutter SDK + Android SDK + Gradle) — or start one from the bundled
Compose file:

```bash
docker compose -f docker-compose.flutter.yml up -d
```

```bash
hgbuild flutter build apk --project /path/to/flutter/project
hgbuild flutter build apk --project /path/to/flutter/project --build-mode release --flavor staging
hgbuild flutter build appbundle --project /path/to/flutter/project --build-mode profile
```

The first build of a given source hash always runs `flutter pub get` +
`flutter build` (cache miss). An identical second run returns the cached
artifact. iOS isn't supported — no signing keys ship in the image, and iOS
builds need a macOS host anyway.

## Unity

This one runs a **real Unity Editor in batch mode** on the worker — it is
not a stub. There's no Docker image for it: Unity Editor licensing has to
happen on the actual worker machine, so the worker needs Unity Hub + an
Editor version + an activated license installed directly.

**1. Install Unity on the worker** at a path the worker auto-detects:

| OS | Path |
|---|---|
| macOS | `/Applications/Unity/Hub/Editor/<version>/Unity.app` |
| Linux | `~/Unity/Hub/Editor/<version>/Editor/Unity` |
| Windows | `C:\Program Files\Unity\Hub\Editor\<version>\Editor\Unity.exe` |

Activate the license once (Personal or Pro) — `hg-worker` does not do this
for you. Start `hg-worker serve` and it reports the detected version(s) and
build targets (by inspecting `PlaybackEngines`) at handshake automatically.

**2. Add a build method to the Unity project** — the standard batch-mode
convention, not something `hgbuild` generates:

```csharp
// Assets/Editor/BuildScript.cs
public static class BuildScript {
    public static void BuildAndroid() {
        BuildPipeline.BuildPlayer(scenes, "build/app.apk", BuildTarget.Android, BuildOptions.None);
    }
}
```

**3. Run the build:**

```bash
hgbuild unity build android \
  --project /path/to/UnityProject \
  --build-method BuildScript.BuildAndroid \
  --unity-version 2022.3.10f1 \
  --scripting-backend il2cpp
```

Platforms: `android`, `ios`, `windows`, `linux`, `macos`, `webgl`.

The project archive excludes `Library/`, `Temp/`, `Build/`, `Logs/`, and
`.git/` before upload, so a multi-gigabyte `Library/` cache never crosses the
network. The worker runs:

```
Unity -batchmode -quit -nographics -projectPath <dir> \
  -logFile <path> -buildTarget Android -executeMethod BuildScript.BuildAndroid
```

and streams back the build log plus whatever artifacts your build method
produced.
