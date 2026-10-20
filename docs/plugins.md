# Plugin Architecture

Hybridgrid is extensible at four seams. Each tier matches the shape of
its problem — there is deliberately no single plugin mechanism.

| Tier | Boundary | Mechanism | Contract | Status |
|---|---|---|---|---|
| **Builder plugins** | worker-side | external `hg-exec-<id>` executables, exec + JSON | `builderv1alpha1` | stable |
| **Scheduler plugins** | coordinator-side | external gRPC process | `schedulerv1alpha1` | stable |
| **Event webhooks** | coordinator-side | outbound HTTP POST | `eventhooksv1` | stable |
| **UI / consumers** | coordinator-side | gRPC `TelemetryService` | `hybridgrid.v1` | pre-existing |

The design rule everywhere: **the plugin can fail, lag, or lie — the
core must keep working** (timeouts, fallbacks, bounded queues), and
**the contract stays tiny** (the documented failure mode of plugin
systems is contract creep, not crashes).

---

## 1. Builder plugins (`builderv1alpha1`)

Add a build type to the cluster without touching any Go code: write a
program, name it `hg-exec-<id>`, drop it in the worker's
`--builder-dir`, and submit builds with
`hgbuild build -t <id> <project-dir> -- <args>`.

### Discovery and handshake

- The worker scans `--builder-dir` for **executable** files named
  `hg-exec-<id>` (a `.exe` suffix is trimmed on Windows; it is not
  part of the id).
- `<id>` must match `[a-z0-9-]+`.
- The worker runs `hg-exec-<id> capability`; the plugin answers with
  one JSON object on stdout:

```json
{"id": "script", "api": "builderv1alpha1", "description": "..."}
```

- The `api` string must match the worker's contract version
  **exactly**; a mismatch (or any handshake failure) skips that
  plugin without affecting the others.
- Discovered ids are advertised in the worker's
  `WorkerCapabilities.custom_builders`, so scheduler-side capability
  matching and dispatch work unchanged.

### Execution

`hg-exec-<id> execute` is invoked with one JSON object on **stdin**:

```json
{
  "task_id": "task-...",
  "build_id": "make-42",
  "builder_id": "script",
  "work_dir": "/tmp/hg-builder-123",
  "args": ["--release"],
  "env": {"MY_FLAG": "1"},
  "timeout_ms": 120000
}
```

`work_dir` contains the extracted source archive (plain tar or
zstd-compressed tar — whatever the client sent). The plugin:

- writes **console text to stderr** (surfaced to the client),
- writes **the result JSON to stdout** on exit:

```json
{"success": true, "outputs": ["dist/app.bin", "dist/lib.a"]}
```

`outputs` are `work_dir`-relative **file** paths (directories are not
entries; list their files). Outputs are archived back to the client
with SHA-256 checksums. A non-zero exit, invalid stdout JSON, or
`success:false` is a task failure (the error text is the `error` field
or the stderr tail). The worker enforces the timeout by killing the
plugin process.

**Output safety**: output paths that would escape `work_dir` are
rejected by the worker — a plugin can never exfiltrate files outside
its work directory.

### Coordinator behavior

- Routing is by capability: a build with `custom_builder: "<id>"` is
  dispatched to a worker advertising `<id>`, under the same
  fast-budget + bounded-queue backpressure as every other build type
  (`--dispatch-queue-timeout`).
- There is **no coordinator-side result cache** for custom builders:
  cache keying is builder-specific semantics, so v1 leaves caching to
  the builder (it owns its outputs and can fingerprint them).
- Task events use the builder id as the build type, so dashboards and
  webhooks see `"rust"`, `"script"`, etc.

### Go SDK

`sdk/builder` implements the protocol for you:

```go
package main

import (
    "context"
    "github.com/h3nr1-d14z/hybridgrid/sdk/builder"
)

type myBuilder struct{}

func (myBuilder) Capability() builder.Capability {
    return builder.Capability{ID: "mytool", Description: "builds with mytool"}
}

func (myBuilder) Build(ctx context.Context, t builder.Task) (builder.Result, error) {
    // run your toolchain against files under t.WorkDir
    return builder.Result{Success: true, Outputs: []string{"dist/app.bin"}}, nil
}

func main() { builder.Main(myBuilder{}) }
```

Build it, name the binary `hg-exec-mytool`, drop it in
`--builder-dir`. Any language that can read stdin and write stdout
works equally well — a shell script is a valid builder.

### Reference builders

- `examples/builders/hg-exec-script` — the meta-builder: runs the
  project's own `hg-build.sh`/`build.sh` (`.bat`/`.ps1` on Windows).
  Outputs come from a `.hg-outputs` manifest (one path per line) or
  default to everything under `dist/`. With this builder, **any
  toolchain becomes a distributed hybridgrid build with zero plugin
  code**.

---

## 2. Scheduler plugins (`schedulerv1alpha1`)

Replace the coordinator's worker-selection policy with an external
process (any language, gRPC):

```bash
hg-coord serve --scheduler=external \
    --scheduler-plugin-addr=localhost:50099 \
    --scheduler-plugin-timeout=50ms
```

For non-local plugins, authenticate the channel with mTLS (the
coordinator verifies the plugin with a CA and presents its own client
certificate):

```bash
hg-coord serve --scheduler=external \
    --scheduler-plugin-addr=plugin.cluster:50099 \
    --scheduler-plugin-ca=ca.pem \
    --scheduler-plugin-cert=coord.pem \
    --scheduler-plugin-key=coord.key
```

The three flags are all-or-nothing; omit them all for a plaintext
local socket. `examples/schedulers/leastloaded-plugin` serves the
matching side with `-tls-cert`/`-tls-key`/`-tls-client-ca`
(requiring a client certificate).

Contract: `proto/hybridgrid/v1/scheduler_plugin.proto`.

- `Select` receives the full fleet snapshot (load, health, circuit
  state, capabilities, custom builders) plus the dispatch context
  features (source sizes, filename, compiler) and answers with a
  worker id. It must answer within `--scheduler-plugin-timeout`
  (default 50ms).
- **Fallback semantics**: plugin error, timeout, empty answer, or an
  unknown worker id falls back to the in-process scheduler
  (least-loaded). A broken or lagging plugin can never stall
  dispatch.
- `Feedback` reports completions (worker, success, duration,
  context). It is fire-and-forget: the coordinator queues feedback in
  a bounded buffer and drops on overflow — a slow plugin degrades its
  own learning signal, never cluster throughput.
- The plugin may decline to decide (empty answer) — e.g. during its
  own warm-up — and the fallback decides.

Reference: `examples/schedulers/leastloaded-plugin` (least-loaded
policy over the snapshot, logs feedback).

---

## 3. Event webhooks (`eventhooksv1`)

```bash
hg-coord serve --event-webhook https://ci.example.com/hooks/hybridgrid \
               --event-webhook https://slack.example.com/builds
```

Every task event is POSTed (JSON) to each endpoint:

```json
{"type": "task_started",   "ID": "task-...", "BuildType": "cpp", ...}
{"type": "task_completed", "ID": "task-...", "Status": "completed",
 "DurationMs": 8493, "QueueTimeMs": 7522, "WorkerID": "worker-a", ...}
```

- Non-blocking by construction: events queue in a bounded buffer;
  overflow drops with a log line. The dispatch and completion paths
  never wait on HTTP.
- Per-endpoint delivery retries twice with short backoff, then drops.
- Webhook delivery rides alongside the internal telemetry notifier
  (a fan-out composes both).

---

Deliveries can be authenticated: pass `--event-webhook-secret` and
every POST carries `X-Hybridgrid-Signature: sha256=<hex>` — the
HMAC-SHA256 of the raw body under that secret. A receiver verifies:

```python
digest = hmac.new(secret, body, hashlib.sha256).hexdigest()
hmac.compare_digest("sha256=" + digest,
                    request.headers["X-Hybridgrid-Signature"])
```

The signature covers content integrity, not replay — pair with TLS
and endpoint-side dedup if replays matter.

## 4. UI / consumer seam

Pre-existing (the dashboard split): any process can consume cluster
telemetry over the gRPC `TelemetryService` (`GetStats`,
`ListWorkers`, `ListTasks`, task consoles, `StreamEvents`). The React
dashboard added in PR #39 is itself just another consumer of this
seam — the reference for "plugin UIs".

---

## Versioning and compatibility

- Contract versions are explicit strings (`builderv1alpha1`,
  `schedulerv1alpha1`) checked at handshake / construction.
- `alpha` means: **breaking changes are allowed** until the contract
  is marked stable; the repo's release notes will call every
  breaking change. Pin your plugin deployments to tested releases.
- Additive proto changes (new optional fields) are not breaking.
- The safest plugin posture is: tolerate unknown fields, answer
  within the documented timeout, treat unknown verbs as errors.

## Security model

- **Builder plugins run arbitrary code on the worker by design.**
  Whoever can place a binary in `--builder-dir` (or configure it) is
  an admin of that worker — the same trust level as configuring the
  docker executor's images. Workers that accept untrusted builders
  should run them in constrained environments the operator controls.
- Builder outputs are path-confined to the per-task work directory,
  which is removed after the task.
- Scheduler plugins and webhooks are coordinator-side: plugin
  endpoints should not be exposed publicly; webhook URLs may carry
  bearer tokens in the query/headers your receiver requires — with
  `--event-webhook-secret` set, each POST carries
  `X-Hybridgrid-Signature: sha256=<hex>` (HMAC-SHA256 of the body).
- No plugin tier can bypass task auth (workers/clients still
  authenticate to the coordinator as always).

## Prior art and lessons (design notes)

| System | What we took |
|---|---|
| containerd runtime/v2 shims | binary naming convention, minimal execution-scoped contract |
| HashiCorp go-plugin | subprocess + explicit version handshake; caution against coarse versioning |
| CNI/CSI | exec + JSON on stdio — language-agnostic plugins |
| OpenTelemetry Collector | the counter-example: they dropped runtime plugins for compile-time composition because contract stability is a permanent tax — we keep contracts tiny instead |
| Terraform SDKv1 → plugin-framework | get the contract right early or pay an ecosystem rewrite |

Full research notes with sources: `NOTES-plugins.md`.
