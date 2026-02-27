---
Title: Frontend split execution plan and package graph
Ticket: GEPA-10-FRONTEND-SPLIT-CLEANUP
Status: active
Topics:
    - architecture
    - frontend
    - go-go-os
    - go-go-app-inventory
    - wesen-os
    - bundling
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../go-go-os/apps/os-launcher/src/app/modules.tsx
      Note: Launcher imports inventory module from app source path
    - Path: ../../../../../../../go-go-os/apps/os-launcher/src/app/store.ts
      Note: Launcher imports inventory reducers from app source path
    - Path: ../../../../../../../go-go-os/package.json
      Note: Current frontend build ownership includes apps/inventory and apps/os-launcher
    - Path: ../../../../../../../wesen-os/cmd/wesen-os-launcher/main.go
      Note: Launcher server mounts root UI handler and backend modules
    - Path: ../../../../../../../wesen-os/pkg/launcherui/handler.go
      Note: Composition repo embeds launcher dist assets
ExternalSources: []
Summary: 'Execution-ready plan for splitting frontend ownership: platform in go-go-os, inventory app in go-go-app-inventory, launcher bundling in wesen-os.'
LastUpdated: 2026-02-27T19:35:00-05:00
WhatFor: Primary implementation plan for GEPA-10 frontend split.
WhenToUse: Use as onboarding guide and task execution reference for intern implementation.
---


# Frontend split execution plan and package graph

## Executive Summary

This ticket operationalizes the frontend split into three ownership zones:

1. `go-go-os` keeps reusable frontend platform packages (engine/common + backend host package).
2. `go-go-app-inventory` owns the inventory app frontend code and backend domain logic.
3. `wesen-os` owns composition and distribution, including launcher frontend bundling and Go binary assembly.

Current state is partially split on the backend, but not yet on the frontend:

1. `go-go-os` still contains all frontend apps including `apps/inventory`.
2. `go-go-app-inventory` is currently backend-only.
3. `wesen-os` serves embedded launcher assets but has no JS workspace yet.

This document is written for a new intern to start implementation immediately. It provides:

1. File-backed current-state evidence.
2. Concrete target layout.
3. Phase-by-phase migration with acceptance criteria.
4. API surface guidance for stable cross-repo composition.
5. Risk controls and validation commands.

## Problem Statement

We need a clean dependency tree and clear ownership boundaries.

Current coupling issues:

1. `go-go-os/package.json` still builds inventory + launcher directly (`go-go-os/package.json:9-18`).
2. Launcher depends on inventory via same-repo workspace and source-path imports (`go-go-os/apps/os-launcher/package.json:14-19`, `.../src/app/modules.tsx:4`, `.../src/app/store.ts:8-9`).
3. `go-go-app-inventory` has no JS workspace today (backend-only).
4. `wesen-os` embeds `pkg/launcherui/dist` (`wesen-os/pkg/launcherui/handler.go:12-22`) but has no frontend build ownership (`package.json`/`pnpm-workspace.yaml` missing).

Without this split:

1. Inventory app lifecycle is tightly coupled to platform repo changes.
2. Composition runtime cannot independently build/ship UI artifacts.
3. Externalization and pluginization roadmap remains blocked by source-level coupling.

## Proposed Solution

### Target repo layout

1. `go-go-os`
   - Keep: `packages/engine`, `packages/desktop-os`, `packages/confirm-runtime`, shared tooling/docs.
   - Keep: backend host package at `go-go-os/go-go-os/pkg/backendhost`.
   - Remove: `apps/inventory`.
2. `go-go-app-inventory`
   - Add: `apps/inventory` frontend package.
   - Keep: backend packages (`pkg/inventorydb`, `pkg/inventorytools`, `pkg/backendcomponent`).
   - Export public launcher/reducer entrypoints for composition.
3. `wesen-os`
   - Add: frontend workspace and `apps/os-launcher`.
   - Own: launcher dist production and sync into `pkg/launcherui/dist`.
   - Keep: composition server `cmd/wesen-os-launcher`.

### Package graph (logical)

```text
go-go-os (platform)
  ├── @hypercard/engine
  ├── @hypercard/desktop-os
  └── @hypercard/confirm-runtime

go-go-app-inventory (app)
  └── @hypercard/inventory -> depends on @hypercard/* platform packages

wesen-os (composition runtime)
  └── @hypercard/os-launcher -> depends on @hypercard/inventory + @hypercard/desktop-os + @hypercard/engine
     (build output copied to pkg/launcherui/dist and embedded by Go)
```

### Current-state evidence

1. `go-go-os` root workspace currently includes all `packages/*` + `apps/*` (`go-go-os/pnpm-workspace.yaml:1-3`).
2. Inventory app is still in `go-go-os/apps/inventory` (`go-go-os/apps/inventory/package.json`).
3. Launcher imports inventory internals directly (`go-go-os/apps/os-launcher/src/app/modules.tsx:4`, `.../store.ts:8-9`).
4. `wesen-os` launcher runtime mounts embedded UI handler at `/` (`wesen-os/cmd/wesen-os-launcher/main.go:241-242`).
5. `wesen-os` has `pkg/launcherui/dist/.embedkeep` but no JS tooling bootstrap yet.

## Design Decisions

1. Keep platform packages centralized in `go-go-os`.
   - Rationale: avoid divergence in foundational APIs used by multiple apps.
2. Move inventory frontend with `mv` into app repo.
   - Rationale: preserves intent/history and enforces domain ownership.
3. Move launcher bundling to `wesen-os`.
   - Rationale: composition runtime should own shipping artifact assembly.
4. Replace launcher imports of app private paths with exported package APIs.
   - Rationale: cross-repo compatibility and long-term modularity.
5. Use phased cutover with acceptance checks at each phase.
   - Rationale: keeps migration debuggable and reversible.

## Alternatives Considered

1. Keep all frontend in `go-go-os` and only split backend.
   - Rejected: does not achieve clear app ownership or composition independence.
2. Duplicate platform packages into `wesen-os`.
   - Rejected: creates forked platform surfaces and maintenance drift.
3. Big-bang move of all apps and launcher in one commit.
   - Rejected: too risky; hard to isolate breakage.
4. Keep source-path imports (`@hypercard/inventory/src/...`) indefinitely.
   - Rejected: incompatible with proper package boundaries across repos.

## Implementation Plan

### Phase 0: Baseline and safety rails

Goals:

1. Record current behavior and known breakpoints.
2. Ensure fast regression detection.

Actions:

1. Run baseline build/test in each repo and capture results in diary.
2. Snapshot package import graph for launcher and inventory.
3. Mark known stale scripts as migration hazards.

Commands:

```bash
cd go-go-os && npm run build && npm run test
cd go-go-app-inventory && GOWORK=off go test ./...
cd wesen-os && GOWORK=off go test ./...
```

Acceptance:

1. Baseline results captured.
2. Any existing failures documented as pre-existing.

### Phase 1: Extract inventory frontend into go-go-app-inventory

Goals:

1. Move inventory app ownership.
2. Keep app build/test operational in new repo.

Actions:

1. Bootstrap JS workspace files in `go-go-app-inventory`.
2. Move `apps/inventory` from `go-go-os` using `mv`.
3. Recreate/adjust app-local scripts and tsconfig references.

Command shape:

```bash
cd go-go-app-inventory
mkdir -p apps
mv ../go-go-os/apps/inventory ./apps/inventory
```

Acceptance:

1. `@hypercard/inventory` builds in `go-go-app-inventory`.
2. App scripts (`dev`, `build`) work from new repo.

### Phase 2: Introduce stable public app exports

Goals:

1. Stop launcher from importing app internal file paths.
2. Provide package-level contract for composition.

Current coupling to remove:

1. `go-go-os/apps/os-launcher/src/app/modules.tsx:4`
2. `go-go-os/apps/os-launcher/src/app/store.ts:8-9`

Actions:

1. Add explicit public export surface in `@hypercard/inventory`.
2. Refactor launcher imports to use these exports.

API sketch:

```ts
// apps/inventory/src/launcher/public.ts
export { inventoryLauncherModule } from './module';
export { inventoryReducer } from '../features/inventory/inventorySlice';
export { salesReducer } from '../features/sales/salesSlice';
```

```json
{
  "exports": {
    ".": "./src/index.ts",
    "./launcher": "./src/launcher/public.ts",
    "./reducers": "./src/launcher/public.ts"
  }
}
```

Acceptance:

1. No launcher import references `@hypercard/inventory/src/*`.
2. Launcher tests pass with package exports only.

### Phase 3: Bootstrap frontend workspace in wesen-os

Goals:

1. Make `wesen-os` the frontend composition/bundling owner.
2. Host `apps/os-launcher` in composition repo.

Actions:

1. Add `package.json`, `pnpm-workspace.yaml`, and root `tsconfig.json` in `wesen-os`.
2. Move `apps/os-launcher` from `go-go-os` into `wesen-os/apps/os-launcher`.
3. Wire dependencies for launcher to consume platform + inventory packages.
4. Add `dev/build/test` scripts in `wesen-os`.

Important spike before full move:

1. Validate cross-repo dependency strategy (workspace globs vs `file:` links) works in your checkout model.

Acceptance:

1. `wesen-os/apps/os-launcher` can run `dev`, `build`, and `test`.
2. Launcher can render inventory module without source-path hacks.

### Phase 4: Dist assembly and binary packaging in wesen-os

Goals:

1. Dist production and embed sync are fully composition-owned.
2. One command produces runnable launcher binary with current frontend.

Actions:

1. Add script pipeline in `wesen-os`:
   - `launcher:frontend:build`
   - `launcher:ui:sync`
   - `launcher:binary:build`
2. Sync built frontend into `wesen-os/pkg/launcherui/dist`.
3. Keep `//go:embed all:dist` contract unchanged (`wesen-os/pkg/launcherui/handler.go:12`).

Pseudocode:

```bash
npm run launcher:frontend:build
rm -rf pkg/launcherui/dist/*
cp -R apps/os-launcher/dist/* pkg/launcherui/dist/
go build ./cmd/wesen-os-launcher
```

Acceptance:

1. `GET /` serves launcher frontend (root mount from `main.go:241-242`).
2. `GET /api/os/apps` returns healthy app manifests.
3. `GET /api/apps/inventory/...` routes still work.

### Phase 5: Cleanup and docs hardening

Goals:

1. Remove stale ownership/build references.
2. Leave newcomer-friendly docs in all repos.

Actions:

1. Remove/fix stale `go-go-os` launcher scripts (example: smoke script calling missing `launcher:binary:build` at `go-go-os/scripts/smoke-go-go-os-launcher.sh:58`).
2. Update repo READMEs with final ownership boundaries.
3. Add runbook in `wesen-os` as canonical “how to run full stack”.

Acceptance:

1. No dead launcher build commands in `go-go-os`.
2. New engineer can follow docs and run system from `wesen-os` only.

## Open Questions

1. Should `apps/todo`, `apps/crm`, and `apps/book-tracker-debug` remain platform examples in `go-go-os` or move into dedicated app repos later?
2. Which cross-repo package strategy is preferred for day-to-day local development in `wesen-os`?
3. Should Storybook stay platform-centric in `go-go-os`, or should composition-level stories also live in `wesen-os`?

## References

1. `go-go-os/package.json:9-18`
2. `go-go-os/pnpm-workspace.yaml:1-3`
3. `go-go-os/apps/os-launcher/package.json:14-19`
4. `go-go-os/apps/os-launcher/src/app/modules.tsx:4`
5. `go-go-os/apps/os-launcher/src/app/store.ts:8-9`
6. `go-go-os/packages/engine/package.json:6-15`
7. `go-go-os/packages/desktop-os/package.json:18-20`
8. `go-go-os/scripts/smoke-go-go-os-launcher.sh:58`
9. `wesen-os/pkg/launcherui/handler.go:12-22`
10. `wesen-os/cmd/wesen-os-launcher/main.go:241-242`
