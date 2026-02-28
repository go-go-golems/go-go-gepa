---
Title: 'Implementation plan: dedicated HyperCard Tools app and typed runtime-card editor routing'
Ticket: GEPA-20-HYPERCARD-TOOLS-APP
Status: active
Topics:
    - hypercard
    - go-go-os
    - inventory-app
    - arc-agi
    - frontend
    - modules
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-app-inventory/apps/inventory/src/launcher/renderInventoryApp.tsx
      Note: Inventory cleanup removing direct code editor window ownership
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/apps/hypercard-tools/src/launcher/module.tsx
      Note: Dedicated tools app module rendering runtime-card editors
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/hypercard/editor/editorLaunch.ts
      Note: Hard-cutover payload construction to hypercard-tools app key routing
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/hypercard/editor/runtimeCardRef.ts
      Note: Defines typed runtime-card reference and editor routing codec contract
    - Path: workspaces/2026-02-22/add-gepa-optimizer/wesen-os/apps/os-launcher/src/__tests__/launcherHost.test.tsx
      Note: Host-level regression tests for encoded editor instances and fallback rendering
    - Path: workspaces/2026-02-22/add-gepa-optimizer/wesen-os/apps/os-launcher/src/app/modules.tsx
      Note: Launcher registry integration for hypercard-tools module
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-28T00:17:00.925491113-05:00
WhatFor: Define and execute a hard-cutover plan that moves runtime-card editor routing to a dedicated HyperCard Tools app and typed runtime-card references.
WhenToUse: Use during GEPA-20 implementation, review, and future extension of HyperCard tooling windows.
---


# Implementation plan: dedicated HyperCard Tools app and typed runtime-card editor routing

## Executive Summary

This ticket hard-cuts runtime-card editor window routing from inventory-owned logic to a dedicated launcher app module: `hypercard-tools`.  
The core objective is to eliminate ad-hoc app-key conventions (`code-editor:<cardId>`) that are not launcher-module-safe, and replace them with a typed contract based on `RuntimeCardRef` and shared encode/decode helpers.

The implementation introduces:

1. A dedicated `hypercard-tools` launchable app module in `go-go-os/apps`.
2. Typed runtime-card identity (`ownerAppId + cardId`) for editor routing and payload construction.
3. Shared helper functions that own app-key and instance-id formatting/parsing.
4. A hard cutover of all runtime-card "Edit" pathways to this module.
5. Removal of inventory-specific editor rendering branches.

## Problem Statement

Current runtime-card editor windows open with `appKey = code-editor:<cardId>`, but launcher routing only resolves registered module IDs.  
Because `code-editor` is not a registered launchable module, the shell renders `Unknown app module`.

This exposed deeper architectural debt:

1. Editor routing is encoded as loose string conventions in shared engine code.
2. The runtime tooling surface (editor/debug windows) is effectively coupled to inventory.
3. Card identity is not expressed as a typed, app-scoped contract at editor-routing boundaries.

Without a hard cutover, future app reuse (arc-agi and others) will continue to risk collisions and routing drift.

## Proposed Solution

### A. Dedicated tools module

Create `go-go-os/apps/hypercard-tools` with launcher export `@hypercard/hypercard-tools/launcher`.  
Manifest ID: `hypercard-tools`.

This module is responsible for rendering tooling windows, starting with runtime-card editor.

### B. Typed runtime-card identity and routing helpers

Introduce a small contract in engine HyperCard editor layer:

```ts
export interface RuntimeCardRef {
  ownerAppId: string;
  cardId: string;
}
```

Add pure helpers:

1. `encodeRuntimeCardEditorInstanceId(ref: RuntimeCardRef): string`
2. `decodeRuntimeCardEditorInstanceId(instanceId: string): RuntimeCardRef | null`
3. `buildRuntimeCardEditorAppKey(ref: RuntimeCardRef): string` (using `formatAppKey('hypercard-tools', instanceId)`)

This centralizes all editor routing format decisions.

### C. Hard cutover in editor launch

Replace legacy payload construction in `editorLaunch.ts` so all edit actions open:

`appKey = hypercard-tools:<encoded-instance-id>`

Window IDs and dedupe keys also move to the new namespace.

### D. Host integration

Register `hypercardToolsLauncherModule` in OS launcher module composition and add TS path aliases in `wesen-os`.

### E. Remove inventory-owned editor rendering branch

Inventory launcher no longer renders `CodeEditorWindow` directly and no longer maintains `code-editor-` instance parsing.

## Design Decisions

1. Hard cutover, no compatibility shims.
Reason: user explicitly requested no backwards compatibility and this is a good migration point.

2. Keep `RuntimeCardRef` contract close to editor-launch path first.
Reason: fixes immediate routing failures while establishing reusable contract for broader runtime-card standardization.

3. Use a dedicated app module rather than window-content adapter hacks.
Reason: aligns with launcher architecture and keeps app ownership explicit.

4. Use shared encoder/decoder helpers.
Reason: prevents future format drift and cross-repo mismatch.

## Alternatives Considered

1. Patch inventory launcher to parse legacy `code-editor:<cardId>` app keys.
Rejected: keeps editor ownership in inventory and does not scale to multi-app runtime-card tooling.

2. Keep editor as non-module adapter in shell.
Rejected: bypasses module registry conventions and reintroduces implicit coupling.

3. Extend `parseAppKey` to allow nested colons in instance IDs.
Rejected: broad contract change in desktop core for a single feature; higher blast radius.

## Implementation Plan

### Phase 1: Contract and helpers

1. Add `RuntimeCardRef` and helper codecs in engine hypercard editor package.
2. Add unit tests for encode/decode and app-key helper behavior.

### Phase 2: HyperCard Tools app module

1. Scaffold `go-go-os/apps/hypercard-tools`.
2. Implement launcher module with `renderWindow` that decodes editor instance IDs and renders `CodeEditorWindow`.
3. Export module entrypoints and include package in TS project references.

### Phase 3: Cutover call sites

1. Update `buildCodeEditorWindowPayload` and `openCodeEditor`.
2. Update timeline/runtime-debug edit buttons to pass typed refs.
3. Remove inventory code-editor route branch.

### Phase 4: Launcher integration

1. Register module in `wesen-os` launcher module list.
2. Add path aliases for `@hypercard/hypercard-tools` and `@hypercard/hypercard-tools/launcher`.
3. Verify unknown-app regression is gone.

### Phase 5: Validation and cleanup

1. Run targeted tests and typechecks.
2. Perform manual smoke path:
runtime debug -> Edit -> save/inject -> reflected in registry/session.
3. Update ticket tasks/changelog/diary and commit.

## API Sketches

```ts
// engine/hypercard/editor/runtimeCardRef.ts
export interface RuntimeCardRef {
  ownerAppId: string;
  cardId: string;
}

export function encodeRuntimeCardEditorInstanceId(ref: RuntimeCardRef): string;
export function decodeRuntimeCardEditorInstanceId(instanceId: string): RuntimeCardRef | null;
export function buildRuntimeCardEditorAppKey(ref: RuntimeCardRef): string;
```

```ts
// hypercard-tools launcher module render path
renderWindow({ instanceId }) {
  const ref = decodeRuntimeCardEditorInstanceId(instanceId);
  if (!ref) return <UnknownToolsWindow instanceId={instanceId} />;
  return <CodeEditorWindow cardId={ref.cardId} initialCode={getEditorInitialCode(ref.cardId)} />;
}
```

## Validation Strategy

1. Unit tests for helper codecs.
2. Unit tests for launcher module parsing/render fallback.
3. Workspace typecheck for changed projects.
4. Manual regression test on Edit from runtime debug and timeline cards/widgets.

## Open Questions

1. Should runtime card registry internals also be migrated from plain `cardId` string to composite `RuntimeCardRef` key in this same ticket or a follow-up?
2. Should `hypercard-tools` expose additional tools windows immediately (event viewer, runtime registry inspector), or stay editor-only in phase 1?

## References

1. GEPA-19 hard cutover ticket (prior runtime-card flow changes)
2. `go-go-os/packages/engine/src/hypercard/editor/editorLaunch.ts`
3. `go-go-app-inventory/apps/inventory/src/launcher/renderInventoryApp.tsx`
4. `wesen-os/apps/os-launcher/src/app/modules.tsx`
