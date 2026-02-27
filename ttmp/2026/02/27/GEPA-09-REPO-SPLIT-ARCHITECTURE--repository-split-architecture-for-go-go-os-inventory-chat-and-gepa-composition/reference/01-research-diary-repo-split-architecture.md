---
Title: 'Research diary: repo split architecture'
Ticket: GEPA-09-REPO-SPLIT-ARCHITECTURE
Status: active
Topics:
    - architecture
    - go-go-os
    - frontend
    - inventory-chat
    - gepa
    - plugins
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go-go-gepa/go.mod
      Note: GEPA dependency baseline
    - Path: go-go-os/README.md
      Note: Monorepo architecture baseline used in diary evidence
    - Path: go-go-os/go-inventory-chat/go.mod
      Note: Backend dependency baseline
    - Path: go-go-os/package.json
      Note: Build workflow and coupling evidence
ExternalSources: []
Summary: Chronological research log for the repository split design, including v2 rename to wesen-os and go-go-app-inventory plus command evidence and task planning.
LastUpdated: 2026-02-27T17:26:00-05:00
WhatFor: Provide continuation context and audit trail for how the repository split design was produced.
WhenToUse: Use when continuing implementation planning, reviewing assumptions, or retracing source evidence.
---


# Research diary: repo split architecture

## Goal

Produce a deep, implementation-ready research document (10+ pages) that defines:

1. how to split into three repos,
2. how the final dependency graph should work,
3. what APIs/contracts each repo should expose,
4. how composition repo startup and runtime initialization should behave,
5. with an explicit no-backwards-compatibility migration approach.

## Context snapshot

The working tree already contained prior GEPA and OS integration work under tickets `GEPA-07` and `GEPA-08`. The new ask was specifically repository architecture and operational split, not immediate code refactor.

Primary source roots used:

- `go-go-os/`
- `go-go-os/go-inventory-chat/`
- `go-go-gepa/`
- `go-go-gepa/ttmp/` for ticket docs and storage

No external internet sources were used because this was an internal architecture analysis tied to local repository contracts.

## Chronological log

## Phase 1: ticket and workspace setup

Commands run:

```bash
docmgr status --summary-only
docmgr ticket list
docmgr ticket create-ticket --ticket GEPA-09-REPO-SPLIT-ARCHITECTURE --title "Repository split architecture for go-go-os, inventory chat, and GEPA composition" --topics architecture,go-go-os,frontend,inventory-chat,gepa,plugins
docmgr doc add --ticket GEPA-09-REPO-SPLIT-ARCHITECTURE --doc-type design-doc --title "Repository split blueprint and implementation roadmap"
docmgr doc add --ticket GEPA-09-REPO-SPLIT-ARCHITECTURE --doc-type reference --title "Research diary: repo split architecture"
```

Findings:

- Existing tickets ran through `GEPA-08`; `GEPA-09` was the next logical id.
- Ticket workspace was created successfully with default files (`index.md`, `tasks.md`, `changelog.md`).

Decision:

- Use `GEPA-09` as a dedicated split-architecture ticket to avoid mixing with existing implementation-heavy tickets.

## Phase 2: repository inventory and boundary scan

Commands run:

```bash
rg --files go-go-os | head -n 200
rg --files go-go-gepa | head -n 200
find go-go-os -maxdepth 3 -type d | sort
rg --files go-go-os/go-inventory-chat | head -n 260
```

Findings:

- `go-go-os` contains both frontend workspace and `go-inventory-chat` backend in one tree.
- `go-go-gepa` is separate and mature enough to be consumed as dependency, but current launcher wiring still uses an internal GEPA module in `go-inventory-chat/internal/gepa`.

Interpretation:

- split strategy should treat current `go-inventory-chat/internal/gepa` as transitional implementation and define a clearer adapter boundary in composition repo.

## Phase 3: backend contract extraction

Commands run:

```bash
sed -n '1,260p' go-go-os/go-inventory-chat/cmd/go-go-os-launcher/main.go
sed -n '1,260p' go-go-os/go-inventory-chat/cmd/go-go-os-launcher/inventory_backend_module.go
sed -n '1,260p' go-go-os/go-inventory-chat/internal/backendhost/module.go
sed -n '1,260p' go-go-os/go-inventory-chat/internal/backendhost/manifest_endpoint.go
sed -n '1,260p' go-go-os/go-inventory-chat/internal/backendhost/routes.go
sed -n '1,320p' go-go-os/go-inventory-chat/internal/backendhost/lifecycle.go
sed -n '1,260p' go-go-os/go-inventory-chat/internal/backendhost/registry.go
```

Findings:

- Generic backend module interface is already coherent and reusable.
- Lifecycle startup/required health check semantics are explicit and robust.
- Route namespacing and forbidden legacy aliases are already encoded.
- Reflection endpoint support exists and provides strong foundation for discoverability.

Interpretation:

- Composition repo should carry this generic host package largely unchanged.
- Domain repos should avoid importing host internals directly to prevent cycles.

## Phase 4: frontend contract extraction

Commands run:

```bash
sed -n '1,260p' go-go-os/apps/os-launcher/src/App.tsx
sed -n '1,260p' go-go-os/apps/os-launcher/src/app/modules.tsx
sed -n '1,260p' go-go-os/packages/desktop-os/src/contracts/launchableAppModule.ts
sed -n '1,260p' go-go-os/packages/desktop-os/src/contracts/launcherHostContext.ts
sed -n '1,260p' go-go-os/packages/desktop-os/src/contracts/appManifest.ts
sed -n '1,320p' go-go-os/packages/desktop-os/src/registry/createAppRegistry.ts
sed -n '1,320p' go-go-os/packages/desktop-os/src/store/createLauncherStore.ts
sed -n '1,320p' go-go-os/packages/desktop-os/src/runtime/buildLauncherContributions.ts
sed -n '1,320p' go-go-os/packages/desktop-os/src/runtime/renderAppWindow.ts
```

Findings:

- Frontend module contracts (`LaunchableAppModule`, `AppManifest`, `LauncherHostContext`) are already strong separation points.
- Host resolves backend endpoints by app id (`/api/apps/${appId}`), matching backend route namespacing.
- Current composition of modules in `apps/os-launcher` is hardcoded import list and should move to package-level dependency imports across repos.

Interpretation:

- repo A should own these contracts and runtime primitives;
- repo B should publish domain module packages implementing those contracts;
- repo C should orchestrate imports and runtime registration.

## Phase 5: route and runtime behavior evidence

Commands run:

```bash
sed -n '1,360p' go-go-os/apps/inventory/src/launcher/renderInventoryApp.tsx
sed -n '1,320p' go-go-os/packages/engine/src/chat/runtime/http.ts
sed -n '1,280p' go-go-os/packages/engine/src/chat/ws/wsManager.ts
sed -n '1,280p' go-go-os/packages/engine/src/chat/runtime/conversationManager.ts
```

Findings:

- Inventory frontend module already uses host context API base and ws base fallback logic.
- Chat runtime endpoint assumptions are stable and namespaced when basePrefix is set.
- WS and timeline hydration flow is clear and reusable after split.

Interpretation:

- no route redesign needed for split; route governance should focus on preserving namespaced-only model and removing legacy aliases.

## Phase 6: build pipeline and coupling evidence

Commands run:

```bash
cat go-go-os/package.json
cat go-go-os/apps/os-launcher/package.json
sed -n '1,260p' go-go-os/scripts/sync-launcher-ui.sh
sed -n '1,260p' go-go-os/scripts/build-go-go-os-launcher.sh
sed -n '1,260p' go-go-os/scripts/smoke-go-go-os-launcher.sh
sed -n '1,260p' go-go-os/go-inventory-chat/internal/launcherui/handler.go
```

Findings:

- launcher binary build currently depends on direct filesystem sync from frontend dist into Go embed directory.
- this is a core coupling point to be redesigned for multi-repo flow.

Interpretation:

- composition repo should keep embed behavior (for single binary), but artifact ingestion must come from repo A/B outputs rather than in-tree app build outputs.

## Phase 7: GEPA and dependency evidence

Commands run:

```bash
sed -n '1,320p' go-go-os/go-inventory-chat/internal/gepa/module.go
cat go-go-os/go-inventory-chat/go.mod
cat go-go-gepa/go.mod
```

Findings:

- Internal GEPA module in `go-inventory-chat` already mirrors desired routes and reflection model.
- dependency versions show drift between inventory-chat and go-go-gepa in core shared libs.

Interpretation:

- composition repo should pin versions and own compatibility testing matrix.
- GEPA adapter should be explicit and isolated to simplify future plugin extraction.

## Phase 8: drafting decisions

Key architecture decisions made while drafting:

1. Three-repo split exactly as requested.
2. Keep generic backend host API in composition repo.
3. Keep repo B host-agnostic via component contract to avoid cyclic dependency.
4. Retain namespaced routes; no legacy compatibility shim.
5. Use reflection and schema endpoints as mandatory discoverability surface.
6. Define hard-cut migration phases with no dual-mode support.

Rejected direction:

- keeping a compatibility bridge for legacy routes; rejected because explicit instruction was no backwards compatibility.

## Phase 9: documentation edits

Files authored:

- `design-doc/01-repository-split-blueprint-and-implementation-roadmap.md`
- `reference/01-research-diary-repo-split-architecture.md` (this file)

Supporting ticket files to update next:

- `index.md`
- `tasks.md`
- `changelog.md`

## Quick reference: resulting architecture recommendation

1. Repo A (`hypercard-frontend`): frontend platform packages only.
2. Repo B (`hypercard-inventory-chat`): inventory domain backend + frontend module packages.
3. Repo C (`go-go-os-composition`): backend host, launcher command, adapters, product binary.

Critical API rule:

- all backend app APIs under `/api/apps/<app-id>/*`.

Critical bootstrap rule:

- backend modules start before HTTP server is exposed;
- frontend launcher derives base paths from app id through host context.

## Usage example for future investigations

If a future contributor needs to continue this work:

1. Open this diary and the design doc.
2. Confirm contracts still match source line evidence in references.
3. Create implementation tasks per migration phase.
4. Keep no-compat constraints explicit in every PR.

## Outstanding follow-ups

1. Decide exact GEPA adapter mode for phase 1 (library vs subprocess).
2. Define artifact transport for frontend bundles between repo A and repo C.
3. Add contract-test suites in repo C and enforce them in repo B CI.

## Phase 10: v2 rename and task-board revision

New user direction required a naming and topology update:

1. `go-go-os-composition` renamed to `wesen-os`.
2. `hypercard-inventory-chat` renamed to `go-go-app-inventory`.
3. First execution plan reframed as composition of:
   - `go-go-os`
   - `go-go-gepa`
   - `go-go-app-inventory`
   into `wesen-os`.

Commands run:

```bash
docmgr doc add --ticket GEPA-09-REPO-SPLIT-ARCHITECTURE --doc-type design-doc --title "V2 wesen-os composition plan (go-go-os + go-go-gepa + go-go-app-inventory)"
find go-go-os/go-inventory-chat -maxdepth 3 -type f | sort
```

Outcome:

- Added v2 design doc with detailed phased task board and explicit extraction boundaries.
- Updated ticket index/tasks/changelog to make v2 doc the active implementation reference.

## Phase 11: backend-only split execution kickoff

Objective for this execution run:

1. create detailed backend-only execution tasks in ticket,
2. execute tasks one by one across the new repos,
3. commit each completed task slice,
4. keep an explicit commit-by-commit diary.

Execution task slices:

1. Task S1: ticket tasks + v2 plan finalized and committed.
2. Task S2: initialize `go-go-app-inventory` as extracted backend repo and move inventory backend sources with `mv`.
3. Task S3: initialize `wesen-os` backend host core and move host runtime sources with `mv`.
4. Task S4: wire inventory adapter in `wesen-os` to consume `go-go-app-inventory`.
5. Task S5: compile/test baseline and document remaining gaps.

This diary section will be updated after each task commit with:

- exact files moved,
- exact commit hash and message,
- validation commands and results.

## Phase 12: backend-only split execution (task-by-task with commits)

This phase executed the backend-only split directly in code across the target repos, with commits per slice and `mv`-first migration where possible.

### S1. Ticket/task baseline commit

Commit:

- `go-go-gepa@25b9212` - `docs(gepa-09): add v2 wesen-os backend split plan and task board`

Outcome:

- Ticket workspace captured v2 naming (`wesen-os`, `go-go-app-inventory`).
- Execution board prepared before code migration.

### S2. Move inventory backend from `go-go-os` into `go-go-app-inventory`

Primary move operations (representative):

```bash
mv go-go-os/go-inventory-chat/internal/inventorydb go-go-app-inventory/pkg/inventorydb
mv go-go-os/go-inventory-chat/internal/pinoweb go-go-app-inventory/pkg/pinoweb
mv go-go-os/go-inventory-chat/cmd/go-go-os-launcher/tools_inventory*.go go-go-app-inventory/pkg/inventorytools/
mv go-go-os/go-inventory-chat/cmd/hypercard-inventory-seed go-go-app-inventory/cmd/inventory-seed
```

Implementation notes:

- Initialized module in extracted repo:
  - `go mod init github.com/go-go-golems/go-go-app-inventory`
- Converted extracted tool registry into reusable package API:
  - `package main` -> `package inventorytools`
  - exported `InventoryToolNames` and `InventoryToolFactories`
- Updated imports to extracted package paths.

Commits:

- `go-go-app-inventory@45127d1` - `feat: extract inventory backend packages from go-go-os`
- `go-go-os@4f6c181` - `refactor: move inventory backend sources to go-go-app-inventory`

Validation:

```bash
cd go-go-app-inventory && GOWORK=off go test ./...
```

Result:

- Pass.

### S3. Move backend host + launcher runtime into `wesen-os`

Primary move operations (representative):

```bash
mv go-go-os/go-inventory-chat/internal/backendhost wesen-os/pkg/backendhost
mv go-go-os/go-inventory-chat/internal/launcherui wesen-os/pkg/launcherui
mv go-go-os/go-inventory-chat/internal/gepa wesen-os/pkg/gepa
mv go-go-os/go-inventory-chat/cmd/go-go-os-launcher wesen-os/cmd/wesen-os-launcher
```

Implementation notes:

- Initialized module:
  - `go mod init github.com/go-go-golems/wesen-os`
- Added local replace for development wiring:
  - `go mod edit -replace github.com/go-go-golems/go-go-app-inventory=../go-go-app-inventory`
- Rewrote imports from old in-tree paths to:
  - `github.com/go-go-golems/wesen-os/pkg/...`
  - `github.com/go-go-golems/go-go-app-inventory/pkg/...`
- Renamed command identifiers to `wesen-os-launcher`.

Commits:

- `wesen-os@59bd4c6` - `feat: move os backend host and launcher into wesen-os`
- `go-go-os@dc4dd17` - `refactor: remove moved backend host and launcher sources`

### S4. Test regression and fix

Issue discovered:

- `TestProfileAPI_CRUDRoutesAreMounted` failed with:
  - `unexpected profile API contract key: registry`

Root cause:

- Integration contract helper `assertProfileListItemContract` allowed list-item keys that omitted `registry`, but runtime payload now includes it.

Fix:

- Added `"registry"` to allowed keys in:
  - `wesen-os/cmd/wesen-os-launcher/main_integration_test.go`

Validation:

```bash
cd wesen-os && GOWORK=off go test ./...
```

Result:

- Pass.

### S5. Post-move health check and remaining gaps

Sanity check:

```bash
cd go-go-os/go-inventory-chat && GOWORK=off go test ./...
```

Result:

- `./...` matched no packages (expected at current extraction state).

Known remaining backend-only gaps:

1. Formal host-agnostic `Component` interface in `go-go-app-inventory` still needs explicit package boundary.
2. `wesen-os/pkg/gepa` is still copied/migrated code and not yet replaced by adapter over `go-go-gepa` APIs.
3. Backend CI/smoke automation for multi-repo composition still pending.
