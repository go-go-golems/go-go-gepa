---
Title: Apps Browser UX and technical reference
Ticket: GEPA-11-APPS-BROWSER-UI-WIDGET
Status: active
Topics:
    - frontend
    - ui
    - backend
    - architecture
    - go-go-os
    - wesen-os
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../go-go-app-inventory/pkg/backendcomponent/component.go
      Note: |-
        Inventory module capabilities and route map
        Inventory manifest capabilities and mounted route tree
    - Path: ../../../../../../../go-go-os/go-go-os/pkg/backendhost/manifest_endpoint.go
      Note: |-
        /api/os/apps and /api/os/apps/{id}/reflection endpoint behavior
        Implements /api/os/apps and /api/os/apps/{app_id}/reflection endpoints
    - Path: ../../../../../../../go-go-os/go-go-os/pkg/backendhost/module.go
      Note: |-
        Core BackendModule and reflection data model contract
        Defines BackendModule manifest and reflection payload contracts used by Apps Browser
    - Path: ../../../../../../../go-go-os/go-go-os/pkg/backendhost/routes.go
      Note: |-
        Namespaced mount model and forbidden legacy aliases
        Defines namespaced mount policy and forbidden legacy aliases
    - Path: ../../../../../../../pinocchio/pkg/webchat/http/profile_api.go
      Note: |-
        Profile-related endpoints mounted under inventory namespaced base
        Profile endpoints surfaced under inventory namespaced /api subtree
    - Path: ../../../../../../../wesen-os/cmd/wesen-os-launcher/main.go
      Note: |-
        Runtime composition and endpoint mounting sequence
        Registers manifest endpoint and mounts module namespaces
    - Path: pkg/backendmodule/module.go
      Note: |-
        GEPA reflection API + schema-discovery surface
        GEPA reflection APIs/schemas and run endpoints
ExternalSources: []
Summary: End-to-end UX and technical guide for an Apps Browser widget that lists mounted modules, shows health and capabilities, and drills into reflective API metadata.
LastUpdated: 2026-02-27T22:58:00-05:00
WhatFor: Give UX/design and implementation teams one shared understanding of how mounted app/module discovery works and how to design an Apps Browser window around it.
WhenToUse: Use when designing or implementing the Apps Browser UI widget and related backend integration.
---


# Apps Browser UX and technical reference

## Executive Summary

This ticket defines a new **Apps Browser** widget/window whose job is to help users and operators see what backend app modules are mounted in the running OS instance.

In plain language, the widget should answer these questions immediately:

1. What apps/modules are currently mounted?
2. Are they healthy?
3. Which ones are required for startup?
4. What can each module do (capabilities)?
5. Which modules provide deeper machine-readable docs (reflection)?
6. For reflective modules, what APIs and schemas can we inspect?

The technical foundation already exists in backend contracts:

1. `GET /api/os/apps` lists mounted modules with health/capabilities.
2. `GET /api/os/apps/{app_id}/reflection` provides richer API/schema docs for modules that implement reflection.

Today, `wesen-os` mounts this endpoint and currently returns at least:

1. `inventory` (required, healthy, non-reflective in current code)
2. `gepa` (optional, healthy, reflective)

This document is intentionally written for a mixed audience. It is detailed enough for implementation engineers, but also translated into design-friendly language for UX work.

## Problem Statement

The launcher currently knows app modules at build-time (static frontend registry), but there is no dedicated runtime inspection UI for mounted backend modules.

That creates several user and operator gaps:

1. No explicit module health dashboard in the windowed UX.
2. No simple way to discover dynamic backend capabilities.
3. No guided path to reflective API metadata (if available).
4. No clear distinction between:
   - an app that exists in frontend registry,
   - an app that is mounted in backend runtime,
   - an app that is currently healthy.

For a UX designer, the problem is not only “add another list”. The core design challenge is representing **system truth** and **confidence**:

1. registration truth: module exists in manifest feed,
2. runtime truth: module health is passing/failing,
3. discoverability truth: module supports reflection or not,
4. actionability truth: there is enough metadata to inspect APIs and schemas.

The Apps Browser window should make these states visually obvious.

## Current Architecture (How this works today)

### 1) BackendModule contract in go-go-os

The base contract lives in `go-go-os/go-go-os/pkg/backendhost/module.go`.

Each module provides:

1. `Manifest()` with app id, name, description, required, capabilities.
2. `MountRoutes()` to mount module routes under a sub-mux.
3. lifecycle hooks `Init/Start/Stop/Health`.
4. optionally `Reflection()` if the module supports discoverable metadata.

Reflection model fields already include:

1. `capabilities`
2. `docs`
3. `apis`
4. `schemas`

This is exactly what an Apps Browser “inspect” panel needs.

### 2) Module manifest endpoint and reflection endpoint

`go-go-os/go-go-os/pkg/backendhost/manifest_endpoint.go` registers:

1. `GET /api/os/apps`
2. `GET /api/os/apps/{app_id}/reflection`

`/api/os/apps` response includes per app:

1. `app_id`
2. `name`
3. `description`
4. `required`
5. `capabilities[]`
6. `healthy`
7. optional `health_error`
8. optional `reflection { available, url, version }`

Reflection route behavior:

1. `404` if app id unknown.
2. `501` if app exists but does not implement reflection.
3. `500` if reflection returns internal error.
4. `200` with `ModuleReflectionDocument` when available.

### 3) Namespacing model (important for UX labels)

`go-go-os/go-go-os/pkg/backendhost/routes.go` enforces app namespacing.

Mounted app routes are under:

1. `/api/apps/{app_id}/...`

Legacy aliases like `/chat`, `/ws`, `/api/timeline` are intentionally forbidden for long-term architecture hygiene.

UX implication:

1. the Apps Browser should visually emphasize canonical namespaced base URLs.
2. if you show endpoint examples, always show namespaced forms first.

### 4) Runtime composition in wesen-os

`wesen-os/cmd/wesen-os-launcher/main.go` wires this together:

1. create module registry (`inventory`, `gepa`, future modules)
2. startup lifecycle with required app checks
3. register `/api/os/apps`
4. mount each module at `/api/apps/{app_id}`
5. mount launcher UI at `/`

This means Apps Browser can live purely in frontend and call runtime endpoints directly.

### 5) Current mounted modules and capabilities

Inventory module (`go-go-app-inventory/pkg/backendcomponent/component.go`) manifest:

1. `app_id = inventory`
2. `required = true`
3. capabilities: `chat`, `ws`, `timeline`, `profiles`, `confirm`
4. currently non-reflective (in this runtime wrapper)

GEPA module (`go-go-gepa/pkg/backendmodule/module.go`) manifest:

1. `app_id = gepa`
2. `required = false`
3. capabilities: `script-runner`, `events`, `timeline`, `schemas`, `reflection`
4. reflective; publishes APIs and schema references.

## Live API Evidence (Current running system)

From the currently running `wesen-os` instance:

### `GET /api/os/apps`

```json
{
  "apps": [
    {
      "app_id": "inventory",
      "name": "Inventory",
      "description": "Inventory chat runtime, profiles, timeline, and confirm APIs",
      "required": true,
      "capabilities": ["chat", "ws", "timeline", "profiles", "confirm"],
      "healthy": true
    },
    {
      "app_id": "gepa",
      "name": "GEPA",
      "description": "GEPA script runner backend module",
      "required": false,
      "capabilities": ["script-runner", "events", "timeline", "schemas", "reflection"],
      "reflection": {
        "available": true,
        "url": "/api/os/apps/gepa/reflection",
        "version": "v1"
      },
      "healthy": true
    }
  ]
}
```

### `GET /api/os/apps/gepa/reflection`

Returns structured metadata for docs, APIs, schemas and capability stability labels.

### `GET /api/os/apps/inventory/reflection`

Returns:

1. HTTP `501 Not Implemented`
2. body: `reflection not implemented`

This is an important design state, not an error in the app itself.

## Endpoint Catalog for Apps Browser

The Apps Browser needs a small core endpoint set to function.

| Endpoint | Method | Purpose | Expected Status |
|---|---|---|---|
| `/api/os/apps` | GET | Primary module list + health + capabilities + reflection hint | 200 |
| `/api/os/apps/{app_id}/reflection` | GET | Detailed docs/API/schema metadata | 200 / 501 / 404 |

Useful optional drill-down endpoints (module-specific):

| Module | Endpoint | Method | Notes |
|---|---|---|---|
| `gepa` | `/api/apps/gepa/scripts` | GET | list local runnable scripts |
| `gepa` | `/api/apps/gepa/schemas/{schema_id}` | GET | schema docs for run APIs |
| `inventory` | `/api/apps/inventory/api/chat/profiles` | GET | profile list from mounted profile API |

For design scope, we recommend starting with only `/api/os/apps` and reflection route. Module-specific links can be presented as “advanced inspect actions”.

## Data Model Reference (UX-friendly)

### A) App list model (`/api/os/apps`)

TypeScript interface suggestion:

```ts
export interface AppsManifestResponse {
  apps: AppManifestDocument[];
}

export interface AppManifestDocument {
  app_id: string;
  name: string;
  description?: string;
  required: boolean;
  capabilities?: string[];
  reflection?: {
    available: boolean;
    url?: string;
    version?: string;
  };
  healthy: boolean;
  health_error?: string;
}
```

UX interpretation by field:

1. `required=true`: badge like “Core” or “Required for startup”.
2. `healthy=false`: show warning state and health_error details.
3. `capabilities`: concise chips/tags.
4. `reflection.available=true`: show “Inspectable” action.

### B) Reflection model (`/api/os/apps/{id}/reflection`)

TypeScript interface suggestion:

```ts
export interface ModuleReflectionDocument {
  app_id: string;
  name: string;
  version?: string;
  summary?: string;
  capabilities?: ReflectionCapability[];
  docs?: ReflectionDocLink[];
  apis?: ReflectionAPI[];
  schemas?: ReflectionSchemaRef[];
}

export interface ReflectionCapability {
  id: string;
  stability?: string;
  description?: string;
}

export interface ReflectionAPI {
  id: string;
  method: string;
  path: string;
  summary?: string;
  request_schema?: string;
  response_schema?: string;
  error_schema?: string;
  tags?: string[];
}

export interface ReflectionSchemaRef {
  id: string;
  format: string;
  uri?: string;
  embedded?: unknown;
}
```

UX interpretation by field:

1. `apis`: table view candidate (method + path + summary).
2. `schemas`: clickable list to view schema JSON.
3. `capabilities[].stability`: ideal for beta/stable badges.
4. `docs`: links to design/playbook material.

### C) State model for the Apps Browser window

```ts
type LoadState = 'idle' | 'loading' | 'loaded' | 'error';

type ReflectionState =
  | { kind: 'idle' }
  | { kind: 'loading'; appId: string }
  | { kind: 'loaded'; appId: string; doc: ModuleReflectionDocument }
  | { kind: 'unsupported'; appId: string; reason: string } // 501 case
  | { kind: 'error'; appId: string; message: string };

interface AppsBrowserState {
  loadState: LoadState;
  apps: AppManifestDocument[];
  fetchError?: string;
  selectedAppId?: string;
  reflection: ReflectionState;
  lastRefreshedAt?: string;
}
```

## UX Design Guidance (for non-technical designers)

### Primary window layout

Recommended two-pane layout:

1. Left: app list (cards or compact rows).
2. Right: details panel for selected app.

```text
+---------------------------------------------------------------+
| Apps Browser                                       Refresh ⟳  |
+---------------------------+-----------------------------------+
| [Search app/capability]   | App: Inventory                   |
|                           | Required: Yes                    |
| ● Inventory      healthy  | Health: Healthy                  |
| ● GEPA           healthy  | Capabilities: chat ws timeline   |
|                           | Reflection: Not available         |
|                           |                                   |
|                           | [Open namespaced API base]        |
+---------------------------+-----------------------------------+
```

### Card/row anatomy for each app

Each row should contain:

1. app name
2. app id (smaller text)
3. health status dot + label
4. required/optional badge
5. first 2-3 capability chips
6. reflection availability icon

### Details panel sections

For selected app, show sections in order:

1. Identity
2. Runtime health
3. Capabilities
4. Endpoints
5. Reflection (if available)

If reflection is not available:

1. show an informational block, not a red error block,
2. message example: “This module does not publish reflective API docs yet.”

### Empty/error/loading states

1. Loading: skeleton list + “Discovering mounted apps...”.
2. Error (apps list fetch failed): retry CTA + short diagnostic text.
3. Reflection 501: neutral “not supported yet”.
4. Reflection 404: likely stale selection; prompt list refresh.

### Sorting/filtering suggestions

Default sort:

1. required first,
2. then unhealthy first (to surface issues),
3. then name ascending.

Useful filters:

1. `Required`
2. `Healthy`
3. `Has reflection`
4. capability chip filters (e.g., `timeline`, `profiles`, `schemas`)

### Microcopy recommendations

Preferred labels:

1. “Mounted Apps” (not “Services”)
2. “Reflection available” / “Reflection not available yet”
3. “Health check”
4. “Required at startup”

Avoid labels that imply failure where there is none:

1. do not mark 501 reflection as “broken”.

## Interaction Flows

### Flow 1: Open widget and inspect app health

```text
User opens Apps Browser
  -> UI requests GET /api/os/apps
  -> receives apps[]
  -> list renders with health badges
  -> user selects one app
  -> details panel renders manifest data
```

### Flow 2: Open reflection details (supported app)

```text
User selects GEPA
  -> clicks "Inspect API metadata"
  -> UI requests GET /api/os/apps/gepa/reflection
  -> response 200 with docs/apis/schemas
  -> UI renders API table + schema links
```

### Flow 3: Reflection not supported

```text
User selects Inventory
  -> clicks "Inspect API metadata"
  -> UI requests GET /api/os/apps/inventory/reflection
  -> response 501
  -> UI shows informative "not available yet" panel
```

## Pseudocode: minimal client implementation

```ts
async function loadApps(): Promise<AppManifestDocument[]> {
  const res = await fetch('/api/os/apps');
  if (!res.ok) throw new Error(`apps fetch failed: ${res.status}`);
  const payload: AppsManifestResponse = await res.json();
  return payload.apps ?? [];
}

async function loadReflection(app: AppManifestDocument): Promise<ModuleReflectionDocument | null> {
  if (!app.reflection?.available || !app.reflection.url) return null;
  const res = await fetch(app.reflection.url);
  if (res.status === 501) {
    throw new ReflectionUnsupportedError(app.app_id, 'Reflection not implemented');
  }
  if (res.status === 404) {
    throw new Error(`App not found: ${app.app_id}`);
  }
  if (!res.ok) {
    throw new Error(`Reflection fetch failed: ${res.status}`);
  }
  return (await res.json()) as ModuleReflectionDocument;
}
```

## API Signature Reference

Backend registration entrypoint:

```go
func RegisterAppsManifestEndpoint(mux *http.ServeMux, registry *ModuleRegistry)
```

Module contract summary:

```go
type AppBackendModule interface {
  Manifest() AppBackendManifest
  MountRoutes(mux *http.ServeMux) error
  Init(ctx context.Context) error
  Start(ctx context.Context) error
  Stop(ctx context.Context) error
  Health(ctx context.Context) error
}
```

Optional reflective contract:

```go
type ReflectiveAppBackendModule interface {
  Reflection(ctx context.Context) (*ModuleReflectionDocument, error)
}
```

## Design Decisions

### Decision 1: Build around `/api/os/apps` as single source of truth

Reason:

1. It already includes health + required + capabilities + reflection hint.
2. It is runtime truth, not compile-time assumption.

### Decision 2: Reflection drill-down is optional per app

Reason:

1. contract supports non-reflective modules.
2. UX must treat reflection absence as normal variation.

### Decision 3: Keep Apps Browser read-only in phase 1

Reason:

1. first value is operational clarity and discoverability.
2. avoids coupling to app-specific mutations.

### Decision 4: Keep module-specific endpoint browsing as advanced mode

Reason:

1. primary audience needs clarity first, depth second.
2. reflection already gives enough structure for future “API explorer” expansion.

## Alternatives Considered

### Alternative A: derive app list from static frontend registry only

Rejected because:

1. static registry does not represent runtime health.
2. cannot represent backend-only modules or startup failures accurately.

### Alternative B: add separate “health endpoint” per app for browser list

Rejected because:

1. duplicates existing behavior already surfaced in `/api/os/apps`.
2. increases coordination burden across modules.

### Alternative C: require every app to implement reflection before shipping browser

Rejected because:

1. too strict for incremental adoption.
2. blocks useful first version.

## Implementation Plan

### Phase 1: Data plumbing and list view

1. Add frontend data client for `/api/os/apps`.
2. Add Apps Browser window with list + detail panel.
3. Add loading/error/retry states.

### Phase 2: Reflection inspect panel

1. Add reflection fetch on demand.
2. Render API table and schema links.
3. Support 501/404/500 states with clear UX copy.

### Phase 3: polish and UX handoff hardening

1. capability filters and search.
2. badges and visual hierarchy refinement.
3. link-outs to module base paths and known APIs.

## Risks and Constraints

1. Current launcher module list is static in frontend (`modules.tsx`), while mounted backend list is runtime-driven. The browser must clearly indicate it is showing backend mount truth.
2. Inventory currently does not implement reflection in this runtime wrapper; inspect experience will be asymmetric between modules.
3. Some module routes are mounted through composed handlers (`/api/` subtree in inventory), so endpoint lists should rely on reflection when available and curated docs otherwise.

## Open Questions

1. Should Apps Browser eventually compare frontend-registered apps vs backend-mounted apps and show mismatch warnings?
2. Should we add inventory reflection support as a follow-up ticket so both first-party modules expose inspectable APIs uniformly?
3. Should “required” and “healthy” be represented as separate icons or a single status chip stack in launcher visual language?
4. Should the widget support auto-refresh polling, or manual refresh only in phase 1?

## References

1. `go-go-os/go-go-os/pkg/backendhost/module.go`
2. `go-go-os/go-go-os/pkg/backendhost/manifest_endpoint.go`
3. `go-go-os/go-go-os/pkg/backendhost/routes.go`
4. `wesen-os/cmd/wesen-os-launcher/main.go`
5. `go-go-app-inventory/pkg/backendcomponent/component.go`
6. `go-go-gepa/pkg/backendmodule/module.go`
7. `pinocchio/pkg/webchat/http/profile_api.go`
8. `wesen-os/apps/os-launcher/src/App.tsx`
9. `wesen-os/apps/os-launcher/src/app/modules.tsx`
10. `wesen-os/scripts/smoke-wesen-os-launcher.sh`
