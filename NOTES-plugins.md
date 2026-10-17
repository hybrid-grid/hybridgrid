# Everything-as-a-Plugin — Research & Design Notes

Working notes for the plugin architecture campaign. These feed the final
`docs/plugins.md` and the thesis report. Vietnamese summary at the bottom.

## 1. Prior art researched (with sources)

| System | Plugin model | What we took / rejected | Source |
|---|---|---|---|
| containerd Runtime v2 | external shim binary per runtime (`containerd-shim-runc-v2` in PATH), ttrpc over unix socket, API "minimal and scoped to the execution lifecycle" | **Pattern for worker-side builder plugins**: `hg-exec-<id>` binary convention, minimal contract, per-task process | https://github.com/containerd/containerd/blob/main/docs/runtime-v2.md |
| HashiCorp go-plugin | subprocess, handshake line on stdout (protocol version + socket + cert), gRPC/net-rpc, AutoMTLS, hclog forwarding; "very basic protocol version … incremented to invalidate any previous plugins" | **Pattern for coordinator-side scheduler plugin**; also the caution: coarse versioning is weak — we use explicit `v1alpha1` string in handshake | https://pkg.go.dev/github.com/hashicorp/go-plugin , https://github.com/hashicorp/go-plugin/blob/main/docs/internals.md |
| CNI/CSI | exec binary, JSON stdin→stdout, exit codes, spec version in API version field | **Cheapest builder protocol option**: exec + JSON, language-agnostic, no gRPC stubs needed for plugin authors | (general knowledge) |
| OpenTelemetry Collector | NO runtime plugins — compile-time composition via ocb builder | **Caution**: contract stability for runtime plugins is a permanent maintenance tax; version discipline required | https://opentelemetry.io/docs/collector/extend/ocb |
| Plugin-system failure modes | "weak boundaries", contract creep, silent breakage after host upgrades | **Design rule**: keep contracts tiny; refuse feature requests into the contract; explicit version | https://thechris.in/articles/why-most-plugin-systems-fail-quietly , https://levelup.gitconnected.com/why-most-plugin-systems-fail-quietly-3cfa6a881499 |
| Go stdlib `plugin` | dlopen, Linux-only, same-toolchain lock-in | **Rejected**: repo requires Windows workers | (general knowledge) |
| Terraform providers | largest plugin ecosystem; had to rewrite SDKv1 → plugin-framework because contract v1 was weak | **Caution**: get the contract right early or pay a full-ecosystem rewrite | (general knowledge) |

## 2. Design decisions (the ones that matter)

### D1 — Three plugin tiers, one per problem shape
- **Builder plugin** (worker-side, exec-based, CNI/containerd-shim style): the
  flagship. Users add toolchains ("user tự tạo builder") without touching Go
  code of coord/worker/CLI. Any language that can read JSON and exit.
- **Scheduler plugin** (coordinator-side, gRPC subprocess, go-plugin style):
  external decision engines (learned policies) plug into dispatch.
- **Hooks** (HTTP webhooks): event fan-out, cheapest tier.
- UI plugins: already shipped as the TelemetryService gRPC boundary —
  PR #39 (React rewrite) consumed it without any coordinator change.

### D2 — Why exec/JSON for builders (not gRPC)
- Compile tasks run seconds; a process spawn per task (~1–5 ms) is noise vs
  the native executor already spawning compiler processes per TU.
- Language-agnostic "code, plug and play": bash/python/rust/go all equal.
- No stub generation, no socket lifecycle, no health pings to maintain.
- containerd uses a daemon+RPC because it manages long-lived containers;
  our builds are one-shot — exec matches the workload shape.

### D3 — Wire contract (proto, additive only)
- `WorkerCapabilities.custom_builders: repeated string = 27` — worker
  advertises plugin builder ids.
- `BuildRequest.custom_builder: string` — routes to a plugin builder when
  non-empty (enum stays for core types; BUILD_TYPE_RUST=5 etc. remain hints).
- `BuildRequest.config` oneof gains `CustomConfig{args, env}`.
- Reuse of the existing archive pipeline: `source_archive` (tar.zst) in,
  `artifacts` (tar.zst) out — same as flutter today.

### D4 — Coordinator backpressure for ALL build types (P0, done first)
- flutter/unity selection moved onto `dispatchLoop` (atomic
  select-then-book) + the same fast-budget (~700 ms, 8 attempts) → bounded
  queue (`--dispatch-queue-timeout`, 256 waiters) → completion/registration
  signals as Compile().
- Fixes pre-existing debt: flutter/unity used to hard-fail under overload
  while cpp queued.
- Capacity is part of the filter match: a full worker never wins a dispatch.
- Task events now report coordinator-side queue_ms (worker's field is
  always zero on this path).

### D5 — Builder protocol v1alpha1 (exec contract)
- Discovery: worker scans `--builder-dir` for `hg-exec-*` executables.
- Handshake: `hg-exec-<id> capability` → one JSON line on stdout:
  `{"id":"rust","api":"builderv1alpha1","description":"..."}`
- Execute: `hg-exec-<id> execute` with task JSON on stdin
  (`task_id, build_id, builder_id, work_dir, args, env, timeout_ms`);
  plugin writes **result JSON to stdout** (`{"success":true,"outputs":[...]}`)
  and **console text to stderr**; outputs are work_dir-relative paths that
  get archived back.
- Non-zero exit / invalid JSON = task failure (stderr becomes error text).
- Version policy: `api` string must match exactly; breaking changes bump
  to `builderv1` — old plugins keep working until they adopt it.

### D6 — Security/trust boundary (documented, not code)
- Whoever can deploy a builder binary to a worker IS admin of that worker
  (same trust level as the docker executor images today).
- Scheduler plugin channel: mTLS/token using existing security package.
- Webhooks: coordinator makes outbound calls only; URLs are admin config.

### D7 — Scheduler plugin semantics
- Proxy implements `scheduler.Scheduler` (+LearningScheduler feedback,
  fire-and-forget with drop-on-slow so a lagging plugin can't stall dispatch).
- Timeout per Select (default tens of ms); on timeout/error → fallback to
  configured inner scheduler (leastloaded default) — the uncertainty-
  escalation pattern discussed w.r.t. calibrated confidence.

## 3. Implementation log

### P0 (this PR)
- `dispatchRequest.filter/label` — filter-mode requests on dispatchLoop.
- `selectByFilter` — least-loaded healthy worker with spare capacity.
- `selectBuildWorker` — fast budget + queue + await, returns queue wait.
- flutter/unity handlers: selection via `selectBuildWorker`, manual
  `IncrementTasks` removed (dispatchLoop books atomically), events report
  real queue_ms. `selectFlutterWorker`/`selectUnityWorker` deleted.
- Tests (race): `TestBuild_Flutter_QueuesUntilSlotFrees`,
  `TestBuild_Flutter_QueueTimeout`, `TestBuild_Unity_QueuesUntilSlotFrees`.

## Tóm tắt tiếng Việt (cho báo cáo)
Ba tầng plugin theo hình dạng bài toán: builder (exec+JSON, tham khảo
containerd-shim/CNI), scheduler (subprocess gRPC, tham khảo go-plugin),
hooks (HTTP webhook). Loại Go stdlib plugin (không Windows). Bài học từ
OTel Collector + Terraform: chi phí thật là giữ contract ổn định — nên
contract để nhỏ nhất, version tường minh (`v1alpha1`). P0: đưa selection
của flutter/unity vào cùng cơ chế backpressure của cpp (đỡ nợ kỹ thuật,
events báo queue_ms thật).
