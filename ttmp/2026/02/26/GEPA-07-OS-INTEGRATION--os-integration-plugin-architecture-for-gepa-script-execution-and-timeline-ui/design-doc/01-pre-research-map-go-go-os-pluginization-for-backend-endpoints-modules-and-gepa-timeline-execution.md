---
Title: 'Pre-research map: go-go-os pluginization for backend endpoints/modules and GEPA timeline execution'
Ticket: GEPA-07-OS-INTEGRATION
Status: active
Topics:
    - gepa
    - plugins
    - architecture
    - events
    - tooling
    - runner
    - geppetto
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go-go-gepa/cmd/gepa-runner/plugin_loader.go
      Note: GEPA plugin hook/event sink execution surface.
    - Path: go-go-gepa/pkg/jsbridge/emitter.go
      Note: GEPA plugin event envelope schema.
    - Path: go-go-os/apps/inventory/src/launcher/renderInventoryApp.tsx
      Note: Reusable UI pattern for folder/chat/timeline command wiring.
    - Path: go-go-os/apps/os-launcher/src/app/modules.tsx
      Note: Current static launcher module composition path.
    - Path: go-go-os/go-inventory-chat/cmd/go-go-os-launcher/inventory_backend_module.go
      Note: Inventory module reference implementation for future plugin modules.
    - Path: go-go-os/go-inventory-chat/cmd/go-go-os-launcher/main.go
      Note: Current static backend module composition and lifecycle wiring.
    - Path: go-go-os/go-inventory-chat/internal/backendhost/manifest_endpoint.go
      Note: Module manifest endpoint contract (/api/os/apps).
    - Path: go-go-os/go-inventory-chat/internal/backendhost/module.go
      Note: Backend module contract baseline.
    - Path: go-go-os/go-inventory-chat/internal/backendhost/routes.go
      Note: Namespaced route policy and legacy alias guard.
    - Path: go-go-os/packages/engine/src/chat/ws/wsManager.ts
      Note: Timeline hydration and event replay behavior.
ExternalSources: []
Summary: Pre-research baseline for adding OS-level plugin integration so standalone backend modules can expose GEPA script discovery/execution and timeline windows in go-go-os.
LastUpdated: 2026-02-27T09:30:00-05:00
WhatFor: Starting-point research document for implementing GEPA as a backend/UI plugin in go-go-os.
WhenToUse: Use before implementation to understand existing contracts, gaps, target APIs, and phased rollout plan.
---


# Pre-research map: go-go-os pluginization for backend endpoints/modules and GEPA timeline execution

## Executive summary

This document is a pre-research starting point for implementing `GEPA-07-OS-INTEGRATION`.

The current `go-go-os` architecture already has a strong backend module host pattern and a strong launcher module pattern, but both are currently wired statically:

- backend modules are registered in-process at startup (`go-inventory-chat/cmd/go-go-os-launcher/main.go:188-217`),
- launcher app modules are imported statically in `apps/os-launcher/src/app/modules.tsx:1-12`.

That makes the inventory backend a good template, but not yet a true pluggable system for third-party modules.

The key integration insight is:

1. Keep the existing `/api/apps/<app-id>` namespaced backend contract and lifecycle manager as the core.
2. Add a plugin manager layer that can register modules from separate plugin packages/processes.
3. Add a GEPA backend module that exposes local JS script inventory + run execution + run events.
4. Reuse existing timeline/event windows in the OS UI by emitting SEM-compatible events and timeline entities per GEPA run.

This pre-research recommends a phased approach:

- Phase 1: in-repo GEPA module using existing backendhost interfaces (lowest risk, fastest proof).
- Phase 2: plugin process protocol for out-of-repo modules.
- Phase 3: dynamic launcher module discovery (beyond static TS imports).

## Problem statement and requested outcome

### User-requested outcome

You want:

- plugins separate from the OS codebase,
- ability to register backend endpoints/modules,
- ability to expose `go-go-gepa` functionality in OS UI,
- an app/folder view that shows locally available JS scripts,
- ability to run scripts and view timeline/event windows.

### Why this is non-trivial

The current system supports modular composition, but composition is compile-time/static in both backend and frontend. There is no runtime plugin process registry and no existing GEPA backend module in `go-go-os`.

## Scope of this pre-research

This is an architecture and API pre-research document, not implementation. It includes:

- current-state architecture mapping with file-backed evidence,
- identified extension points,
- concrete proposed APIs for GEPA integration,
- phased research and delivery plan,
- explicit risks and open questions.

It does not include code changes in this step.

## Current-state architecture map

## 1) Backend module host (Go)

`go-go-os` already has a backend module contract that looks like a plugin host interface.

### Existing contract

`AppBackendModule` has:

- manifest,
- route mount,
- init/start/stop/health lifecycle methods,

defined in `go-inventory-chat/internal/backendhost/module.go:8-25`.

`ModuleRegistry` enforces unique app IDs and ordered module access in `internal/backendhost/registry.go:8-66`.

`LifecycleManager` handles startup ordering, required-module validation, health checks for required modules, and reverse-order stop in `internal/backendhost/lifecycle.go:13-114`.

`MountNamespacedRoutes` enforces `/api/apps/<app-id>` path prefixing and blocks legacy aliases (`/chat`, `/ws`, `/api/timeline`) in `internal/backendhost/routes.go:10-67`.

`/api/os/apps` manifest + health endpoint exists in `internal/backendhost/manifest_endpoint.go:23-54`.

### What this means

The backend host is already plugin-ready at the interface level, but registration source is still static code in `main.go`.

## 2) Inventory backend module as reference implementation

Inventory backend module is implemented in `cmd/go-go-os-launcher/inventory_backend_module.go:21-129` and is the best concrete template for future plugins.

It demonstrates:

- module manifest capabilities (`chat`, `ws`, `timeline`, `profiles`, `confirm`) at `:45-58`,
- namespaced route mounting of chat/ws/timeline/profile APIs at `:61-103`,
- module health semantics for `/api/os/apps`.

The main launcher binary composes that module via:

- `backendhost.NewModuleRegistry(...)` at `main.go:188-196`,
- lifecycle startup at `:203-208`,
- namespaced mount loop at `:213-217`.

This is exactly the pattern a GEPA module should follow initially.

## 3) Launcher frontend module system (TypeScript)

### Existing app module contract

Launcher app modules use `LaunchableAppModule` from `packages/desktop-os/src/contracts/launchableAppModule.ts:22-29` with:

- app manifest,
- optional state reducer slice,
- window launch builder,
- optional desktop contributions,
- window renderer,
- optional registration hook.

Host context already provides endpoint resolvers:

- `resolveApiBase(appId)`
- `resolveWsBase(appId)`

in `launcherHostContext.ts:3-10` and used by `apps/os-launcher/src/App.tsx:22-35`.

### Static composition today

Current composition is static imports:

- `apps/os-launcher/src/app/modules.tsx:1-12`
- `apps/os-launcher/src/app/registry.ts:1-4`

So frontend pluginization is not yet dynamic. This is a major gap for out-of-repo UI plugins.

## 4) Chat/timeline data plane already exists and is reusable

### Event and timeline ingestion flow

Conversation runtime:

- `ConversationManager` ensures module registration and WS/session handling (`packages/engine/src/chat/runtime/conversationManager.ts:36-72`).
- `WsManager` hydrates timeline snapshot via `GET /api/timeline` and replays buffered WS events (`packages/engine/src/chat/ws/wsManager.ts:95-97`, `:360-448`).

SEM handling:

- `semRegistry` maps incoming SEM event types to Redux timeline/chat actions (`packages/engine/src/chat/sem/semRegistry.ts:55-74`, `:315-413`).
- timeline entities are upserted into `timelineSlice` (`packages/engine/src/chat/state/timelineSlice.ts:132-285`).

### Debug windows already implemented

The OS already has:

- live event viewer (`EventViewerWindow`) in `packages/engine/src/chat/debug/EventViewerWindow.tsx`,
- timeline structured debugger + YAML export (`TimelineDebugWindow`) in `packages/engine/src/chat/debug/TimelineDebugWindow.tsx`,
- inventory command wiring to open those windows (`apps/inventory/src/launcher/renderInventoryApp.tsx:497-503`, `:551-553`, `:672-681`, `:1016-1023`).

This is important: a GEPA timeline window can be delivered quickly by routing GEPA run events into the same SEM/timeline pipeline.

## 5) GEPA runtime surfaces already available in go-go-gepa

`go-go-gepa` already has the pieces needed to run local JS scripts and stream progress.

### Script/plugin contracts

- Optimizer plugin descriptor contract in `cmd/gepa-runner/plugin_loader.go:18-193`.
- Dataset generator plugin contract in `pkg/dataset/generator/plugin_loader.go:17-217`.
- JS helper module `require("gepa/plugins")` defines descriptor validators in `cmd/gepa-runner/gepa_plugins_module.go:12-132`.

### Async + event stream support

- Promise-aware execution bridge in `pkg/jsbridge/call_and_resolve.go:36-81`.
- Structured plugin event envelope in `pkg/jsbridge/emitter.go:10-88`.
- CLI stream output format (`stream-event {...}`) in `cmd/gepa-runner/plugin_stream.go:13-33`.

### Candidate run and dataset generation entrypoints

- Candidate run command (`gepa-runner candidate run`) in `cmd/gepa-runner/candidate_run_command.go:43-323`.
- Dataset generation command (`gepa-runner dataset generate`) in `cmd/gepa-runner/dataset_generate_command.go:40-167`.

### Run persistence

- Candidate run table + write path in `cmd/gepa-runner/candidate_run_store.go:11-125`.
- Optimization/eval run recorder tables in `cmd/gepa-runner/run_recorder.go:18-349`.

This means GEPA can be integrated as a backend module without inventing execution engines from scratch.

## Gap analysis

### Gap A: backend module registration source is static

Today, module registration happens in `main.go` via direct constructor calls. There is no runtime loader, no plugin discovery folder, and no plugin process handshake.

### Gap B: launcher app module registration is static imports

`apps/os-launcher` hard-codes module list. There is no dynamic frontend module loading from `/api/os/apps` or plugin manifest metadata.

### Gap C: no GEPA app/module in go-go-os yet

No existing `/api/apps/gepa/*` backend namespace and no launcher module for GEPA.

### Gap D: no standardized bridge from GEPA plugin-stream events to OS SEM timeline

GEPA emits plugin events; OS expects SEM envelopes/timeline upserts for robust timeline windows.

### Gap E: local script discovery security and policy are undefined

No current allowlist/path policy for exposing and executing local JS scripts via OS UI.

## Proposed architecture (starting point)

## Architecture direction

Recommended architecture is a two-layer plugin strategy:

1. Module layer (Go backendhost compatible):
- plugin contributes `AppBackendModule` semantics (manifest/routes/lifecycle/health).

2. UI layer (launcher module):
- plugin contributes frontend module metadata + either static bundle hook (Phase 1) or dynamic module host mechanism (Phase 2/3).

For GEPA initial delivery, prioritize backend module first and reuse existing UI primitives in inventory-style launcher module.

## 1) Backend plugin manager (new)

Add a plugin manager between launcher main and `backendhost.NewModuleRegistry`.

### Target responsibilities

- discover plugin specs from configured directories,
- load in-process module adapters and/or connect out-of-process plugin processes,
- validate app IDs and capabilities,
- instantiate `AppBackendModule` adapters,
- register modules into existing backendhost lifecycle.

### Suggested shape (Go pseudocode)

```go
type BackendPluginSpec struct {
    PluginID   string
    AppID      string
    Kind       string // "inproc" | "process"
    Command    []string
    Env        map[string]string
    Required   bool
    Capabilities []string
}

type BackendPluginManager interface {
    Discover(ctx context.Context) ([]BackendPluginSpec, error)
    LoadModule(ctx context.Context, spec BackendPluginSpec) (backendhost.AppBackendModule, error)
}

func BuildModuleRegistry(ctx context.Context, manager BackendPluginManager, builtins []backendhost.AppBackendModule) (*backendhost.ModuleRegistry, error) {
    specs, err := manager.Discover(ctx)
    if err != nil { return nil, err }

    modules := append([]backendhost.AppBackendModule{}, builtins...)
    for _, spec := range specs {
        mod, err := manager.LoadModule(ctx, spec)
        if err != nil { return nil, fmt.Errorf("load plugin %s: %w", spec.PluginID, err) }
        modules = append(modules, mod)
    }
    return backendhost.NewModuleRegistry(modules...)
}
```

## 2) GEPA backend module contract (new namespace)

Define `app_id = "gepa"` and mount under `/api/apps/gepa/*`.

### Initial endpoint set (draft)

- `GET /api/apps/gepa/scripts`
  - list discovered local script descriptors.
- `POST /api/apps/gepa/runs`
  - start a run (candidate run or dataset generation).
- `GET /api/apps/gepa/runs/{run_id}`
  - run metadata/status.
- `GET /api/apps/gepa/runs/{run_id}/events`
  - stream events (SSE or WS).
- `GET /api/apps/gepa/runs/{run_id}/timeline`
  - timeline snapshot for hydration.
- `POST /api/apps/gepa/runs/{run_id}/cancel`
  - cooperative cancel.

### Why this shape

It mirrors existing chat runtime structure:

- command endpoint (`POST`),
- stream endpoint (`ws` or SSE),
- snapshot endpoint for hydration,
- status endpoint for polling/fallback.

This aligns with how `WsManager` and timeline hydration already work (`wsManager.ts:95-97`, `:360-448`).

## 3) GEPA script inventory model

### Discovery sources

- explicit configured script directories,
- workspace-specific directories,
- optional per-user directory.

### Recommended script manifest fields

```json
{
  "script_id": "local.examples.coaching-entity-sentiment",
  "path": "/abs/path/to/script.js",
  "kind": "dataset-generator",
  "api_version": "gepa.dataset-generator/v1",
  "name": "Coaching Entity/Sentiment Longitudinal Generator",
  "registry_identifier": "local.examples",
  "mtime_ms": 1700000000000,
  "size_bytes": 12345
}
```

### Validation source

Use existing descriptor validation in go-go-gepa loader code (`plugin_loader.go`, `generator/plugin_loader.go`) rather than duplicating schema logic.

## 4) GEPA event to SEM/timeline bridge

### Existing GEPA event format

`jsbridge.Event` includes:

- `sequence`,
- `timestamp_ms`,
- plugin metadata,
- `type`, `level`, `message`, `data`, `payload`,

from `pkg/jsbridge/emitter.go:10-23`.

### Proposed mapping

Convert each GEPA plugin event to a SEM envelope event family:

- `gepa.run.start`
- `gepa.run.progress`
- `gepa.run.warning`
- `gepa.run.error`
- `gepa.run.complete`
- `gepa.dataset.row.generated`

Also project meaningful timeline entities (`timeline.upsert`) with kinds like:

- `gepa_status`
- `gepa_metric`
- `gepa_row`

Then register timeline renderer normalizers and renderers (same extension model as hypercard):

- renderer registration pattern in `rendererRegistry.ts:28-95`,
- optional props normalization in `timelinePropsRegistry.ts:21-43`.

### Bridge pseudocode

```go
func mapGepaEventToSem(runID string, ev jsbridge.Event) SemEnvelope {
    semType := "gepa.run.progress"
    if strings.Contains(strings.ToLower(ev.Level), "error") {
        semType = "gepa.run.error"
    }

    return SemEnvelope{
        Sem: true,
        Event: SemEvent{
            Type: semType,
            ID: fmt.Sprintf("%s:%d", runID, ev.Sequence),
            Data: map[string]any{
                "runId": runID,
                "sequence": ev.Sequence,
                "message": ev.Message,
                "level": ev.Level,
                "pluginId": ev.PluginID,
                "pluginMethod": ev.PluginMethod,
                "payload": ev.Payload,
            },
        },
    }
}
```

## 5) Launcher UI integration strategy

## Phase 1 (pragmatic)

Add `gepaLauncherModule` as a normal compile-time launcher module, but keep backend logic pluginized first.

Why:

- fastest path to user value,
- keeps frontend risk low,
- validates endpoint and timeline contracts early.

The module should render:

- a folder-like GEPA home window (similar to inventory folder pattern `renderInventoryApp.tsx:830-875`),
- script list panel,
- run control panel,
- event viewer / timeline debug window links.

It should use `resolveApiBase('gepa')` and `resolveWsBase('gepa')` exactly like inventory does (`apps/inventory/src/launcher/module.tsx:51-54`).

## Phase 2/3 (true external UI plugins)

Introduce runtime UI plugin manifests (fetched from backend plugin manager) and a dynamic module host capable of loading sandboxed frontend bundles.

This is a larger security/runtime problem and should not block GEPA backend integration.

## API reference draft for GEPA backend module

## 1) List scripts

`GET /api/apps/gepa/scripts`

Response:

```json
{
  "scripts": [
    {
      "script_id": "local.examples.coaching-entity-sentiment",
      "path": "/home/.../exp-11-coaching-dataset-generator.js",
      "kind": "dataset-generator",
      "api_version": "gepa.dataset-generator/v1",
      "name": "Coaching Entity/Sentiment Longitudinal Generator",
      "registry_identifier": "local.examples",
      "mtime_ms": 1772181000000,
      "size_bytes": 12943
    }
  ]
}
```

## 2) Start run

`POST /api/apps/gepa/runs`

Request:

```json
{
  "script_id": "local.examples.coaching-entity-sentiment",
  "mode": "dataset_generate",
  "profile": "gpt5nano",
  "stream": true,
  "input": {
    "config_path": "/abs/path/to/exp-11-coaching-dataset-config.yaml",
    "overrides": {
      "count": 10,
      "dry_run": true
    }
  }
}
```

Response:

```json
{
  "run_id": "gepa-run-20260227-123456",
  "status": "running",
  "ws_url": "/api/apps/gepa/runs/gepa-run-20260227-123456/ws",
  "timeline_url": "/api/apps/gepa/runs/gepa-run-20260227-123456/timeline"
}
```

## 3) Stream events

`GET /api/apps/gepa/runs/{run_id}/ws` (or SSE alternative)

Frame payload (SEM envelope):

```json
{
  "sem": true,
  "event": {
    "type": "gepa.run.progress",
    "id": "gepa-run-20260227-123456:17",
    "data": {
      "runId": "gepa-run-20260227-123456",
      "sequence": 17,
      "message": "building entity timeline",
      "level": "info",
      "pluginMethod": "generateOne",
      "payload": {
        "type": "row.generated",
        "data": {
          "row_index": 3
        }
      }
    }
  }
}
```

## 4) Timeline snapshot

`GET /api/apps/gepa/runs/{run_id}/timeline`

Return timeline snapshot compatible with existing hydrate flow expectations (`wsManager.ts:367-379`).

## Sequence diagrams

### A) Server startup with plugin manager

```text
go-go-os-launcher main
  -> BackendPluginManager.Discover()
  -> BackendPluginManager.LoadModule(spec...)
  -> backendhost.NewModuleRegistry(builtins + plugin modules)
  -> backendhost.NewLifecycleManager(registry).Startup(...)
  -> RegisterAppsManifestEndpoint(/api/os/apps)
  -> MountNamespacedRoutes(/api/apps/<app-id>/*)
  -> serve launcher UI /
```

### B) GEPA run + timeline window

```text
User opens GEPA folder window
  -> UI calls GET /api/apps/gepa/scripts
  -> User selects script and clicks Run
  -> UI POST /api/apps/gepa/runs
  -> Backend starts go-go-gepa run worker
  -> GEPA plugin emits jsbridge events
  -> GEPA module maps events -> SEM envelopes + timeline.upsert
  -> UI EventViewerWindow receives stream
  -> UI TimelineDebugWindow loads /timeline snapshot + live updates
```

## Concrete file map for implementation planning

## go-go-os backend (likely changes)

- `go-inventory-chat/internal/backendhost/module.go`
  - optional extension of manifest metadata for plugin provenance.
- `go-inventory-chat/internal/backendhost/manifest_endpoint.go`
  - include plugin source metadata if useful.
- `go-inventory-chat/cmd/go-go-os-launcher/main.go`
  - replace direct module constructor list with plugin manager composition.
- New: `go-inventory-chat/internal/pluginhost/*`
  - discovery, process supervision, module adapters.
- New: `go-inventory-chat/internal/gepamodule/*`
  - script listing, run orchestration, event bridge, timeline snapshot.

## go-go-os frontend (likely changes)

- `apps/os-launcher/src/app/modules.tsx`
  - Phase 1 static include of `gepaLauncherModule`.
- New: `apps/gepa/src/launcher/module.tsx`
  - GEPA launcher module.
- New: `apps/gepa/src/launcher/renderGepaApp.tsx`
  - script list, run controls, open timeline windows.
- Optional later: dynamic frontend plugin registry host in `packages/desktop-os`.

## go-go-gepa integration touchpoints

- Reuse runtime and plugin loaders:
  - `cmd/gepa-runner/plugin_loader.go`
  - `pkg/dataset/generator/plugin_loader.go`
- Reuse stream/event bridge format:
  - `pkg/jsbridge/emitter.go`
  - `cmd/gepa-runner/plugin_stream.go`
- Optional reuse of persistence tables:
  - `cmd/gepa-runner/candidate_run_store.go`
  - `cmd/gepa-runner/run_recorder.go`

## Testing and validation plan (pre-research recommendation)

## Backend tests

1. Plugin manager discovery/load tests:
- malformed spec rejection,
- duplicate app ID rejection,
- missing required plugin failure behavior.

2. GEPA module API tests:
- list scripts response schema,
- run lifecycle transitions,
- cancel behavior.

3. Event/timeline bridge tests:
- GEPA event -> SEM type mapping,
- timeline snapshot hydration compatibility.

## Frontend tests

1. GEPA module command routing and window dedupe.
2. Script list render + run trigger.
3. Event viewer and timeline debug windows opening from GEPA windows.

## End-to-end smoke

Extend launcher smoke pattern (`scripts/smoke-go-go-os-launcher.sh`) with:

- `GET /api/apps/gepa/scripts` returns at least one script (fixture dir),
- start run endpoint returns run id,
- stream endpoint yields at least one event,
- timeline endpoint returns non-empty snapshot for run.

## Risks and tradeoffs

## Key risks

1. Plugin process isolation vs latency:
- out-of-process gives isolation but adds protocol complexity.

2. Dynamic frontend plugin loading security:
- runtime JS bundle loading needs sandboxing/signature policy.

3. Event taxonomy sprawl:
- unconstrained GEPA event types can make timeline rendering noisy.

4. Local script execution safety:
- must avoid arbitrary path traversal and uncontrolled process execution.

## Tradeoff recommendation

- Take backend pluginization first; keep UI static for first ship.
- Standardize GEPA event taxonomy before broad plugin support.
- Use strict allowlist roots for script discovery.

## Open questions

1. Should external backend plugins use HTTP, gRPC, or NDJSON stdio protocol?
2. Is frontend dynamic plugin loading required in the first GEPA milestone, or is static launcher module acceptable initially?
3. For GEPA runs, should timeline state be persisted durably (SQLite) or in-memory only for first pass?
4. Should `/api/os/apps` include frontend module metadata eventually, or should there be a separate `/api/os/plugins` endpoint?
5. What is the trust model for local script directories (workspace-only vs user-home)?

## Recommended phased plan

## Phase 0: Contract spike (1-2 days)

- Define GEPA module endpoint schemas.
- Build mapping prototype from `jsbridge.Event` to SEM/timeline entities.
- Validate with EventViewerWindow + TimelineDebugWindow only.

## Phase 1: In-repo GEPA backend module + static UI module (short-term delivery)

- Add `gepa` backend module using existing backendhost lifecycle.
- Add `apps/gepa` launcher module statically in `apps/os-launcher`.
- Deliver script list + run + timeline windows.

## Phase 2: Backend plugin manager (separate plugin packages/processes)

- Add discovery + load manager.
- Support plugin specs from folder config.
- Include health/status in `/api/os/apps`.

## Phase 3: Dynamic frontend plugin loading (optional hardening milestone)

- Design and implement UI plugin manifests + secure runtime loading.

## References

Core backend host and inventory module:

- `go-go-os/go-inventory-chat/internal/backendhost/module.go`
- `go-go-os/go-inventory-chat/internal/backendhost/registry.go`
- `go-go-os/go-inventory-chat/internal/backendhost/lifecycle.go`
- `go-go-os/go-inventory-chat/internal/backendhost/routes.go`
- `go-go-os/go-inventory-chat/internal/backendhost/manifest_endpoint.go`
- `go-go-os/go-inventory-chat/cmd/go-go-os-launcher/main.go`
- `go-go-os/go-inventory-chat/cmd/go-go-os-launcher/inventory_backend_module.go`

Launcher module system and inventory UI:

- `go-go-os/packages/desktop-os/src/contracts/launchableAppModule.ts`
- `go-go-os/packages/desktop-os/src/contracts/launcherHostContext.ts`
- `go-go-os/packages/desktop-os/src/registry/createAppRegistry.ts`
- `go-go-os/apps/os-launcher/src/app/modules.tsx`
- `go-go-os/apps/os-launcher/src/App.tsx`
- `go-go-os/apps/inventory/src/launcher/module.tsx`
- `go-go-os/apps/inventory/src/launcher/renderInventoryApp.tsx`

Chat/timeline/event plumbing:

- `go-go-os/packages/engine/src/chat/runtime/conversationManager.ts`
- `go-go-os/packages/engine/src/chat/ws/wsManager.ts`
- `go-go-os/packages/engine/src/chat/sem/semRegistry.ts`
- `go-go-os/packages/engine/src/chat/state/timelineSlice.ts`
- `go-go-os/packages/engine/src/chat/debug/EventViewerWindow.tsx`
- `go-go-os/packages/engine/src/chat/debug/TimelineDebugWindow.tsx`

GEPA runner/plugin/runtime surfaces:

- `go-go-gepa/cmd/gepa-runner/main.go`
- `go-go-gepa/cmd/gepa-runner/candidate_run_command.go`
- `go-go-gepa/cmd/gepa-runner/dataset_generate_command.go`
- `go-go-gepa/cmd/gepa-runner/plugin_loader.go`
- `go-go-gepa/cmd/gepa-runner/plugin_stream.go`
- `go-go-gepa/cmd/gepa-runner/gepa_plugins_module.go`
- `go-go-gepa/pkg/jsbridge/emitter.go`
- `go-go-gepa/pkg/jsbridge/call_and_resolve.go`
- `go-go-gepa/pkg/dataset/generator/run.go`
- `go-go-gepa/pkg/dataset/generator/plugin_loader.go`

