# Experimental distributed Unity IL2CPP build

This pilot targets a Linux x86_64 standalone Player. Use the updated coordinator, CLI and workers together. A Linux x86_64 Editor worker needs a Unity Hub Editor with Linux IL2CPP support and a working license, Python 3, bubblewrap with user namespaces enabled, and hgbuild on PATH. Copy [the sample build method](../../examples/unity/HybridGridLinuxPlayer.cs) into Assets/Editor/ in a Unity 2022.3 project with an enabled scene. Alternatively use a project method that selects IL2CPP and writes the complete Player under Build/ with an .x86_64 executable.

Install the launcher on the Editor worker and set its toolchain root to the LLVM installation actually used by Unity Bee for this target. Use the native compiler invocation in the Unity build log to identify it. The launcher expects bin/clang++ inside this root. Use a dedicated pilot cluster of Linux x86_64 compile workers with the matching Clang version and target ABI. The current generic C++ scheduler does not enforce identical toolchain versions.

```sh
install -m 755 scripts/unity/hg-unity-il2cpp /usr/local/bin/hg-unity-il2cpp
export HG_UNITY_LLVM_ROOT=/path/to/unity/linux/llvm
hg-unity-il2cpp --probe
hg-worker serve --coordinator HOST:9000 --max-parallel 1
```

Start the Editor worker with this environment. The worker advertises distributed IL2CPP after detecting Unity and passing the launcher probe. Configure the nested hgbuild client with access to the same coordinator (including HG_TOKEN for token-authenticated deployments). The pilot targets a plaintext LAN coordinator. Run at least one additional C++ worker with a free slot. Set the Editor worker's maximum parallel tasks to one to reserve it for the parent build.

```sh
hgbuild --coordinator=HOST:9000 unity build linux --project ./Game \
  --build-method HybridGridLinuxPlayer.Build --distributed-il2cpp
```

Unity drives the normal build graph. A per-build mount namespace exposes the wrapper at Unity's Clang driver path while preserving the original LLVM installation for preprocessing and linking. The launcher expands Clang response files and routes compile commands for generated .cpp sources under il2cppOutput through hgbuild. Linking and packaging run on the Editor worker. Incompatible toolchains or a Unity version that bypasses the selected compiler need separate validation. A successful Editor exit with zero observed HybridGrid objects fails the pilot build.

The coordinator skips whole-Player cache hits in this mode. The compiler wrapper bypasses the client-side object cache for generated sources. For a distribution benchmark, use fresh object caches and inspect C++ task completions under the parent BuildID on distinct worker IDs. Compare the resulting Player with a clean local build and run it before timing. A completed build returns Build/hybridgrid-player.tar.gz with the executable and data directory together.
