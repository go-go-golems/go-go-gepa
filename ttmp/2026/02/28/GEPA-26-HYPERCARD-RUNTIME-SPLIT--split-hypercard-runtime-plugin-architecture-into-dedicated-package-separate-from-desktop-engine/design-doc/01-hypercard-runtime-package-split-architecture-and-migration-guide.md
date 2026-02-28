---
Title: HyperCard runtime package split architecture and migration guide
Ticket: GEPA-26-HYPERCARD-RUNTIME-SPLIT
Status: active
Topics:
    - architecture
    - frontend
    - go-go-os
    - hypercard
    - js-vm
    - plugins
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../go-go-app-arc-agi-3/apps/arc-agi-player/src/launcher/module.tsx
      Note: External consumer dependency pattern on engine runtime adapter surface
    - Path: ../../../../../../../go-go-os/package.json
      Note: Workspace-level scripts and package topology baseline
    - Path: ../../../../../../../go-go-os/packages/desktop-os/src/store/createLauncherStore.ts
      Note: Host-level store assembly and reserved reducer key model
    - Path: ../../../../../../../go-go-os/packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx
      Note: Primary coupling seam between desktop shell and runtime service
    - Path: ../../../../../../../go-go-os/packages/engine/src/components/shell/windowing/pluginIntentRouting.ts
      Note: Bridge between runtime intents and desktop/domain dispatch
    - Path: ../../../../../../../go-go-os/packages/engine/src/features/pluginCardRuntime/pluginCardRuntimeSlice.ts
      Note: Runtime session state
    - Path: ../../../../../../../go-go-os/packages/engine/src/hypercard/artifacts/artifactProjectionMiddleware.ts
      Note: HyperCard artifact projection and runtime card registration flow
    - Path: ../../../../../../../go-go-os/packages/engine/src/index.ts
      Note: Current mixed export surface proving package responsibility overlap
    - Path: ../../../../../../../go-go-os/packages/engine/src/plugin-runtime/runtimeService.ts
      Note: Core QuickJS runtime service and session lifecycle
    - Path: ../../../../../../../go-go-os/packages/engine/src/plugin-runtime/stack-bootstrap.vm.js
      Note: VM bootstrap API and ui helper contract
ExternalSources: []
Summary: Intern-onboarding architecture document and phased design for splitting HyperCard and plugin runtime concerns out of engine shell/windowing.
LastUpdated: 2026-02-28T15:44:00Z
WhatFor: Provide implementation-grade clarity for package boundary extraction and runtime integration stability.
WhenToUse: Use when planning or implementing package split, dependency cleanup, and runtime/desktop ownership boundaries.
---


# HyperCard runtime package split architecture and migration guide

## Executive Summary

This document explains the current runtime plugin architecture in `go-go-os` from first principles, then proposes how to split HyperCard and VM runtime concerns into dedicated packages separate from desktop shell/windowing.

Current state is functional but tightly coupled:

1. `@hypercard/engine` exports everything: desktop shell, windowing reducers, plugin runtime, HyperCard artifact runtime, and many widgets.
2. `PluginCardSessionHost` lives under shell/windowing while depending on plugin runtime reducers, desktop selectors, and HyperCard artifact state.
3. Apps consume runtime entrypoints via `@hypercard/engine/desktop-hypercard-adapter`, which creates implicit coupling between app launchers and engine internals.

Recommended target state:

1. Desktop-first package for shell/windowing (`@hypercard/desktop-engine`).
2. HyperCard/runtime-first package for QuickJS plugin runtime, runtime state, artifact projection, card host components (`@hypercard/hypercard-runtime`).
3. Small bridge package for desktop+runtime adapter wiring (`@hypercard/desktop-hypercard`).

This keeps runtime innovation fast without destabilizing shell/windowing and clarifies ownership for contributors.

## Audience and Reading Guide

This document is written for a new intern who does not know:

1. this monorepo layout,
2. the HyperCard model,
3. the runtime plugin pipeline,
4. why package boundaries matter here.

Recommended read order:

1. Repository fundamentals (`What exists today`).
2. Runtime architecture walkthrough (`How a plugin card actually runs`).
3. Problem statement (`Why split now`).
4. Target architecture and migration plan (`How to split safely`).

## Repository Fundamentals

## 1) Workspace overview

Root workspace (`go-go-os`) currently hosts:

1. shared packages
   - `packages/engine`
   - `packages/desktop-os`
   - `packages/confirm-runtime`
2. demo apps
   - `apps/todo`
   - `apps/crm`
   - `apps/book-tracker-debug`
3. nested Go backend host module under `go-go-os/go-go-os/pkg/backendhost`

Primary workspace config:

1. root `package.json` defines workspaces `packages/*`, `apps/*`
2. `pnpm-workspace.yaml` mirrors that workspace scope

## 2) What each package does today

### `@hypercard/engine`

Current mixed-responsibility package:

1. desktop shell UI and windowing
2. desktop core state (`windowing` reducers/selectors/actions)
3. plugin runtime (`QuickJSCardRuntimeService`, stack bootstrap, contracts, UI schema)
4. HyperCard runtime/artifact logic
5. generic widgets, theme, diagnostics, debug utilities

This concentration is the root packaging issue.

### `@hypercard/desktop-os`

Host/composition layer:

1. app manifest contracts
2. registry and render helpers for app windows
3. store assembly helpers that call `createAppStore` from engine

`desktop-os` depends on `@hypercard/engine`, and currently treats several engine reducers as reserved keys:

1. `pluginCardRuntime`
2. `windowing`
3. `notifications`
4. `debug`
5. `hypercardArtifacts`

### `@hypercard/confirm-runtime`

Runtime package for operator confirmations using engine widgets.

It currently depends on `@hypercard/engine` and demonstrates a mature compound-widget orchestration pattern.

## 3) Export surface reality

Engine currently exports all major domains from one barrel (`packages/engine/src/index.ts`):

1. app utilities
2. cards
3. chat
4. plugin runtime
5. widgets
6. debug/diagnostics
7. hypercard
8. notifications and plugin runtime state
9. theme and shared types

Additionally, engine exposes subpath entrypoints:

1. `@hypercard/engine/desktop-core`
2. `@hypercard/engine/desktop-react`
3. `@hypercard/engine/desktop-hypercard-adapter`

These are practical seams we can formalize into separate packages.

## Runtime Plugin Architecture (Current State)

This section explains the runtime in end-to-end sequence form.

## 1) Core runtime data model

### `CardStackDefinition`

From `packages/engine/src/cards/types.ts`:

1. `id`, `name`, `icon`, `homeCard`
2. `plugin.bundleCode` (raw VM JavaScript bundle)
3. `plugin.capabilities` (`domain`, `system` allowlists)
4. `cards` metadata map

The `cards` map is mostly metadata in plugin mode; real render/handler logic comes from VM bundle.

### Plugin bundle loading pattern

Typical app stack pattern:

1. `pluginBundle.vm.js` authored as runtime script
2. app imports it as raw string: `import bundleCode from './pluginBundle.vm.js?raw'`
3. stack sets `plugin.bundleCode = bundleCode`

This pattern exists in demo apps and `go-go-app-arc-agi-3`.

## 2) VM bootstrap contract

`packages/engine/src/plugin-runtime/stack-bootstrap.vm.js` defines global runtime primitives:

1. `globalThis.defineStackBundle(factory)`
2. `globalThis.defineCard(...)`
3. `globalThis.defineCardRender(...)`
4. `globalThis.defineCardHandler(...)`
5. `globalThis.ui` helper object with primitive DSL builders

Host-facing control plane is exposed as `globalThis.__stackHost`:

1. `getMeta()`
2. `render(cardId, cardState, sessionState, globalState)`
3. `event(cardId, handlerName, args, cardState, sessionState, globalState)`
4. define/patch methods for runtime card mutation

## 3) QuickJS runtime service

`packages/engine/src/plugin-runtime/runtimeService.ts`:

1. uses `quickjs-emscripten` and singlefile runtime
2. maintains `SessionVm` per `sessionId`
3. load path
   - create VM
   - eval bootstrap script
   - eval bundle code
   - read and validate metadata via `__stackHost.getMeta()`
4. render path
   - eval `__stackHost.render(...)`
   - validate output with `validateUINode`
5. event path
   - eval `__stackHost.event(...)`
   - validate returned intents via `validateRuntimeIntents`
6. mutation path
   - `defineCard`, `defineCardRender`, `defineCardHandler`

Service boundaries are clean and mostly pure runtime concern, making it a strong candidate for extraction.

## 4) UI tree and intent contracts

Plugin runtime contracts (`contracts.ts` and schema validators) define:

1. UINode tree transport shape
2. Runtime intents with four scopes:
   - `card`
   - `session`
   - `domain`
   - `system`

Schema guards exist for both rendered UI and intent payloads, reducing host risk from malformed VM output.

## 5) Runtime state slice

`packages/engine/src/features/pluginCardRuntime/pluginCardRuntimeSlice.ts`:

1. tracks runtime sessions (`loading|ready|error`)
2. maintains per-session `sessionState` and per-card `cardState`
3. records runtime timeline entries
4. queues pending domain/system/nav intents

Capability policy module controls domain/system authorization before host dispatch.

Selectors (`selectProjectedRuntimeDomains` etc.) provide:

1. runtime-local state access
2. projection of global app domain slices into VM `globalState.domains`

This entire feature module is not shell-specific and should move with runtime package.

## 6) Card host component in shell layer

`PluginCardSessionHost` currently lives in shell/windowing:

1. reads desktop navigation session state
2. registers runtime session in plugin runtime slice
3. loads VM bundle via `QuickJSCardRuntimeService`
4. applies initial session/card patch intents
5. injects runtime cards from registry
6. renders tree through `PluginCardRenderer`
7. dispatches event intents via `dispatchRuntimeIntent`

This component is the largest coupling junction.

## 7) Intent routing bridge

`pluginIntentRouting.ts` bridges runtime intents to host actions:

1. always ingests intents into runtime timeline
2. for `domain` intents
   - emits Redux action type `${domain}/${actionType}`
3. for `system` intents
   - routes `nav.back`, `nav.go`, `notify`, `window.close`

This file mixes runtime concerns and desktop command semantics.

## 8) HyperCard artifact projection and runtime card injection

Artifact middleware (`hypercard/artifacts/artifactProjectionMiddleware.ts`):

1. listens to timeline entity events
2. extracts HyperCard widget/card upserts
3. writes `hypercardArtifacts` state
4. if runtime card code present, registers card in runtime card registry

Registry (`plugin-runtime/runtimeCardRegistry.ts`) then supports injection into loaded sessions.

This is runtime-domain functionality and should move with runtime package, not remain in shell package.

## 9) Desktop shell integration point

Default adapters (`defaultWindowContentAdapters.tsx`) create a card adapter that renders `PluginCardSessionHost`.

Desktop shell itself is generic windowing, but it currently exports and co-locates runtime-specific host components in the same barrel.

## Problem Statement

Current architecture creates high coupling and unclear ownership:

1. package-level coupling
   - engine owns both desktop shell and runtime core
2. component-level coupling
   - runtime host component resides inside shell/windowing
3. app-level coupling
   - launchers import runtime adapter from engine subpath
4. store-level coupling
   - default app store wires runtime and hypercard reducers as engine core

Symptoms:

1. difficult to reason about what is "desktop framework" vs "hypercard runtime"
2. larger blast radius for runtime changes
3. unclear package API for third-party apps that only need one side
4. harder testing boundaries and release cadence partitioning

## Goals and Non-Goals

## Goals

1. establish explicit package boundaries:
   - desktop shell/windowing package
   - hypercard runtime/plugin package
2. preserve runtime behavior and app compatibility during migration
3. reduce import coupling from apps and downstream repos
4. keep development ergonomics for launchers and store composition

## Non-Goals

1. rewriting runtime VM or intent model
2. redesigning UI DSL in this ticket
3. replacing Redux with another state architecture
4. changing external app behavior as first objective

## Target Package Architecture

Recommended package layout:

1. `@hypercard/desktop-engine`
2. `@hypercard/hypercard-runtime`
3. `@hypercard/desktop-hypercard` (small adapter bridge)
4. keep `@hypercard/desktop-os` as host/composition contract package
5. keep `@hypercard/confirm-runtime` depending on desktop/widget packages as needed

## 1) `@hypercard/desktop-engine`

Owns:

1. desktop core state (`windowing`)
2. shell/windowing components and adapters that are runtime-agnostic
3. desktop command infrastructure/context menus
4. shared widgets/theme/parts that are general-purpose

Does not own:

1. QuickJS runtime
2. plugin runtime contracts/schemas
3. HyperCard artifact projection
4. plugin runtime reducers/selectors

## 2) `@hypercard/hypercard-runtime`

Owns:

1. plugin runtime core (`runtimeService`, `stack-bootstrap`, schema/contracts)
2. runtime state feature (`pluginCardRuntime` reducers/selectors/capability policy)
3. HyperCard artifact runtime and projection middleware
4. `PluginCardRenderer`, `PluginCardSessionHost`, `pluginIntentRouting`
5. runtime card registry

Depends on:

1. React + Redux peer deps
2. minimal desktop abstractions through adapter interfaces (not full shell internals)

## 3) `@hypercard/desktop-hypercard`

Tiny package for integration glue:

1. provides `createHypercardCardContentAdapter`
2. maps desktop windowing context into runtime host props
3. optionally exposes convenience composition helpers for launchers

Reason:

1. avoids circular imports between desktop and runtime packages
2. keeps each core package clean

## 4) Dependency direction

Desired direction:

1. `desktop-engine` -> no dependency on runtime package
2. `hypercard-runtime` -> no dependency on shell internals
3. `desktop-hypercard` -> depends on both `desktop-engine` and `hypercard-runtime`
4. apps -> depend on `desktop-os` + whichever runtime/bridge packages they need

## 5) Proposed public APIs

### `@hypercard/hypercard-runtime`

High-level exports:

1. types and contracts
   - `CardStackDefinition` extensions as needed
   - runtime intent/UI types
2. runtime service
   - `QuickJSCardRuntimeService`
3. runtime state
   - `pluginCardRuntimeReducer`
   - actions/selectors/capability utilities
4. HyperCard state
   - `hypercardArtifactsReducer`
   - artifact selectors/actions
   - projection middleware factory
5. UI runtime host
   - `PluginCardRenderer`
   - `PluginCardSessionHost`
   - `dispatchRuntimeIntent`
6. registry utilities
   - `registerRuntimeCard`, `injectPendingCardsWithReport`, etc.

### `@hypercard/desktop-engine`

High-level exports:

1. `desktop-core` reducers/actions/selectors
2. `desktop-react` shell components and contribution/adaptation primitives
3. general widget/theme systems
4. optional `createBaseDesktopStore` that excludes runtime-specific reducers

### `@hypercard/desktop-hypercard`

High-level exports:

1. `createHypercardCardContentAdapter`
2. optional `withHypercardRuntime` helper for shell composition

## Current-to-Target Mapping Table

| Current location | Target package |
| --- | --- |
| `engine/src/plugin-runtime/*` | `hypercard-runtime` |
| `engine/src/features/pluginCardRuntime/*` | `hypercard-runtime` |
| `engine/src/hypercard/*` | `hypercard-runtime` |
| `engine/src/components/shell/windowing/PluginCardSessionHost.tsx` | `hypercard-runtime` |
| `engine/src/components/shell/windowing/PluginCardRenderer.tsx` | `hypercard-runtime` |
| `engine/src/components/shell/windowing/pluginIntentRouting.ts` | `hypercard-runtime` |
| `engine/src/components/shell/windowing/defaultWindowContentAdapters.tsx` (runtime adapter part) | `desktop-hypercard` |
| `engine/src/desktop/core/*` | `desktop-engine` |
| `engine/src/components/shell/windowing/*` (generic shell) | `desktop-engine` |

## Migration Plan (Phased)

## Phase 0: Baseline and freeze

1. add explicit architecture ticket docs (this ticket)
2. lock down behavior via targeted tests around:
   - runtime service integration
   - plugin session host rerender
   - intent routing
   - artifact projection
3. add dependency graph snapshot in CI (import boundaries)

Exit criteria:

1. stable baseline tests pass
2. agreed ownership matrix accepted

## Phase 1: Create new package skeletons

1. scaffold `packages/hypercard-runtime`
2. scaffold `packages/desktop-engine`
3. scaffold `packages/desktop-hypercard`
4. add minimal exports and build scripts

No functional movement yet; only package shells and no-op adapters.

Exit criteria:

1. packages build in CI
2. zero runtime behavior change

## Phase 2: Move pure runtime modules

Move without React/shell concerns first:

1. `plugin-runtime/*`
2. `features/pluginCardRuntime/*`
3. runtime contracts/types

Add compatibility re-exports from `@hypercard/engine` for one transition window.

Exit criteria:

1. runtime unit/integration tests pass from new package
2. old imports still compile through compatibility exports

## Phase 3: Move HyperCard artifact runtime

Move:

1. `hypercard/artifacts/*`
2. runtime card registry glue

Adjust store builders to import reducers/middleware from new package.

Exit criteria:

1. artifact projection tests pass
2. runtime card injection still works in existing demo flows

## Phase 4: Move runtime React host components

Move:

1. `PluginCardRenderer`
2. `PluginCardSessionHost`
3. `pluginIntentRouting`

Introduce host adapters/interfaces so runtime host no longer imports desktop selectors directly.

Suggested interface seam:

1. runtime host accepts `navContext` and `windowFocus` from props or callbacks
2. desktop bridge package resolves those from windowing state

Exit criteria:

1. card windows function unchanged
2. no direct import from runtime package into desktop core internals

## Phase 5: Extract bridge package

Move runtime-specific adapter creation out of desktop-engine:

1. `createHypercardCardContentAdapter` into `@hypercard/desktop-hypercard`

Update apps:

1. replace `@hypercard/engine/desktop-hypercard-adapter` imports with bridge package import

Exit criteria:

1. launcher modules compile with new adapter import path
2. desktop-engine no longer exports runtime host components

## Phase 6: Harden boundaries and remove compatibility shims

1. remove transitional re-exports from old engine package
2. enforce lint/import rules:
   - desktop package cannot import runtime package internals
   - runtime package cannot import shell internals
3. update docs and starter templates

Exit criteria:

1. clean dependency DAG
2. all apps migrated
3. no deprecated path usage

## Testing Strategy

## 1) Unit tests

Runtime package:

1. `runtimeService.integration.test.ts`
2. `intentSchema` + `uiSchema` tests
3. `pluginCardRuntimeSlice` reducer tests
4. artifact parsing/projection tests

Desktop package:

1. windowing reducer tests
2. command/context action tests
3. shell behavior tests

Bridge package:

1. adapter rendering tests with mocked runtime host
2. integration tests for card-window creation path

## 2) Integration tests

1. open card window -> load plugin bundle -> render tree -> click event -> intent dispatch -> domain/system handling
2. timeline projection producing runtime card injection
3. multi-session runtime behavior isolation

## 3) Application smoke tests

Validate `apps/todo`, `apps/crm`, `apps/book-tracker-debug`, and `go-go-app-arc-agi-3` launchers:

1. import path migration
2. runtime card rendering
3. navigation commands (`nav.go`, `nav.back`)

## 4) CI strategy

1. package-level typecheck/build matrix
2. dependency-boundary static checks
3. targeted runtime integration suite as required gate

## Rollback Plan

If migration phase fails:

1. retain compatibility re-exports for previous phase
2. keep bridge adapter dual-wired for one release window
3. revert only failing moved modules; avoid wide rollback
4. gate rollout by package-level feature flags where needed

## Risks and Mitigations

## Risk 1: Circular dependencies after split

Mitigation:

1. maintain strict dependency direction with bridge package
2. add import-lint rules early

## Risk 2: Runtime host loses nav/window context behavior

Mitigation:

1. introduce explicit host adapter contract before moving component
2. lock behavior via integration tests on nav/back/focus interactions

## Risk 3: App breakage in external repos

Mitigation:

1. one transition release with compatibility exports
2. migration guide with exact import replacement table
3. smoke test with `go-go-app-arc-agi-3` before shim removal

## Risk 4: Store composition drift

Mitigation:

1. provide canonical `createLauncherStore` example using new packages
2. document reserved reducer keys by package ownership

## Alternatives Considered

## Alternative A: Keep all runtime + shell in `@hypercard/engine`

Pros:

1. no migration work

Cons:

1. coupling remains high
2. ownership remains unclear
3. external consumer ergonomics remain poor

Decision: rejected.

## Alternative B: Move only QuickJS runtime service, keep host components in shell

Pros:

1. low effort

Cons:

1. major coupling remains (`PluginCardSessionHost` and runtime reducers)
2. limited architectural payoff

Decision: rejected.

## Alternative C: Merge runtime into `desktop-os` instead

Pros:

1. fewer packages

Cons:

1. violates separation of host contracts vs runtime execution
2. harder for non-desktop consumers of runtime

Decision: rejected.

## Implementation Notes for New Interns

## 1) Local setup quickstart

From `go-go-os` root:

```bash
npm install
npm run build
npm run test
npm run storybook
```

For runtime-focused iteration:

```bash
cd packages/engine
npx vitest run src/plugin-runtime/runtimeService.integration.test.ts
```

## 2) Where to start reading code

Read in this order:

1. `packages/engine/src/cards/types.ts`
2. `packages/engine/src/plugin-runtime/stack-bootstrap.vm.js`
3. `packages/engine/src/plugin-runtime/runtimeService.ts`
4. `packages/engine/src/features/pluginCardRuntime/pluginCardRuntimeSlice.ts`
5. `packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx`
6. `packages/engine/src/components/shell/windowing/pluginIntentRouting.ts`
7. `packages/engine/src/hypercard/artifacts/artifactProjectionMiddleware.ts`
8. app launcher example: `apps/todo/src/launcher/module.tsx`

## 3) Mental model checklist

Before coding split work, make sure you can explain:

1. how VM code is loaded and validated
2. where runtime session state lives
3. how runtime intents become Redux actions
4. how HyperCard timeline entities create runtime card injections
5. why `PluginCardSessionHost` is the coupling center

## 4) Concrete migration work items for first PR

Good starter PR for intern:

1. create `packages/hypercard-runtime` with copied `plugin-runtime/*`
2. add package exports + tests
3. wire compatibility re-exports in engine
4. no UI moves yet

This keeps scope constrained and reviewable.

## Open Questions

1. Should `hypercardArtifacts` remain runtime package state or become a dedicated artifact package later?
2. Do we keep `PluginCardSessionHost` in runtime package permanently, or split host and pure runtime presenter components?
3. Should `desktop-os` eventually depend on `desktop-engine` only, with runtime wiring opt-in via bridge?
4. How long should compatibility re-exports be supported before hard cut?

## Recommendation

Proceed with three-package split (`desktop-engine`, `hypercard-runtime`, `desktop-hypercard`) using phased migration and compatibility shims.

Rationale:

1. maximizes architectural clarity,
2. minimizes immediate breakage,
3. keeps runtime and shell evolution independently scalable.

## References

Current package and export surfaces:

1. `go-go-os/package.json`
2. `go-go-os/packages/engine/package.json`
3. `go-go-os/packages/desktop-os/package.json`
4. `go-go-os/packages/confirm-runtime/package.json`
5. `go-go-os/packages/engine/src/index.ts`
6. `go-go-os/packages/engine/src/desktop-core.ts`
7. `go-go-os/packages/engine/src/desktop-react.ts`
8. `go-go-os/packages/engine/src/desktop-hypercard-adapter.ts`

Runtime/plugin internals:

1. `go-go-os/packages/engine/src/cards/types.ts`
2. `go-go-os/packages/engine/src/plugin-runtime/stack-bootstrap.vm.js`
3. `go-go-os/packages/engine/src/plugin-runtime/runtimeService.ts`
4. `go-go-os/packages/engine/src/plugin-runtime/contracts.ts`
5. `go-go-os/packages/engine/src/plugin-runtime/runtimeCardRegistry.ts`
6. `go-go-os/packages/engine/src/features/pluginCardRuntime/pluginCardRuntimeSlice.ts`
7. `go-go-os/packages/engine/src/features/pluginCardRuntime/selectors.ts`
8. `go-go-os/packages/engine/src/features/pluginCardRuntime/capabilityPolicy.ts`
9. `go-go-os/packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx`
10. `go-go-os/packages/engine/src/components/shell/windowing/pluginIntentRouting.ts`
11. `go-go-os/packages/engine/src/components/shell/windowing/defaultWindowContentAdapters.tsx`

HyperCard artifacts and store composition:

1. `go-go-os/packages/engine/src/hypercard/artifacts/artifactProjectionMiddleware.ts`
2. `go-go-os/packages/engine/src/hypercard/artifacts/artifactsSlice.ts`
3. `go-go-os/packages/engine/src/hypercard/artifacts/artifactRuntime.ts`
4. `go-go-os/packages/engine/src/app/createAppStore.ts`
5. `go-go-os/packages/engine/src/desktop/core/state/windowingSlice.ts`

App consumption references:

1. `go-go-os/apps/todo/src/launcher/module.tsx`
2. `go-go-os/apps/crm/src/launcher/module.tsx`
3. `go-go-os/apps/book-tracker-debug/src/launcher/module.tsx`
4. `go-go-app-arc-agi-3/apps/arc-agi-player/src/launcher/module.tsx`
5. `go-go-app-arc-agi-3/apps/arc-agi-player/src/app/store.ts`
