---
Title: 'Design and Implementation Plan: split chat runtime out of engine and introduce generic visibility context'
Ticket: GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT
Status: active
Topics:
    - architecture
    - frontend
    - go-go-os
    - hypercard
    - plugins
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/apps/book-tracker-debug/src/app/stories/BookTrackerDebugApp.stories.tsx
      Note: Existing app-level story destination pattern
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/desktop-os/src/store/createLauncherStore.ts
      Note: Desktop composition entrypoint for runtime reducers
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/chat/index.ts
      Note: Chat surface currently bundled into engine
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/contextActionVisibility.ts
      Note: Generic visibility filter primitives to preserve
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/types.ts
      Note: Visibility contracts and context target types
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/useDesktopShellController.tsx
      Note: Current in-engine visibility context derives chat profile state
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/widgets/BookTracker.stories.tsx
      Note: Runtime-coupled story candidate for relocation
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/widgets/ChatConversationWindow.stories.tsx
      Note: Chat runtime story candidate for relocation
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/index.ts
      Note: Engine barrel currently exports chat
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/hypercard-runtime/src/app/createAppStore.ts
      Note: Current app-store factory ownership and reducer composition
ExternalSources: []
Summary: Design for splitting chat + runtime concerns out of engine, moving misplaced stories, and introducing a generic visibility-context contract for context-menu policy gating.
LastUpdated: 2026-02-28T16:27:48.331611036-05:00
WhatFor: Onboard engineers to current coupling issues and provide a phased implementation plan to produce a clean package DAG.
WhenToUse: Use when implementing or reviewing engine/chat/runtime package boundaries and context-menu visibility policy architecture.
---


# Design and Implementation Plan: split chat runtime out of engine and introduce generic visibility context

## Executive Summary

This document proposes a package-boundary cleanup across `go-go-os` so that the dependency graph is explicit, acyclic, and easier to reason about for both runtime behavior and Storybook development.

The core proposal has three coordinated parts:

1. Move runtime-coupled stories out of `packages/engine` into the package that owns the runtime concern.
2. Replace shell-internal chat profile introspection with a generic, externally injectable visibility-context resolver.
3. Move all chat runtime code from `packages/engine/src/chat` into a dedicated package (`packages/chat-runtime`) and treat it as a consumer of engine primitives, not part of engine core.

Today, engine is not purely windowing/core UI. It exports chat (`packages/engine/src/index.ts:15-16`) and includes chat/runtime dependencies in its package (`packages/engine/package.json:24-35`). At the same time, hypercard runtime depends on engine (`packages/hypercard-runtime/package.json:18`) and builds app stores that include engine reducers (`packages/hypercard-runtime/src/app/createAppStore.ts:56-63`). This works functionally, but blurs ownership and increases coupling between desktop shell, runtime host concerns, and chat behavior.

The desired end state is:

- `@hypercard/engine`: desktop/windowing core, shared widgets, base diagnostics/notifications, generic policy primitives.
- `@hypercard/hypercard-runtime`: plugin cards, runtime host, artifact projection, card session machinery.
- `@hypercard/chat-runtime`: conversation runtime, profile handling, timeline decode, chat windows and debug panes.
- `@hypercard/desktop-os`: composition layer that chooses which runtime packages to include and wires store + shell policy.

This split keeps features reusable while preserving a clear DAG:

```text
apps/*, desktop-os
   -> @hypercard/chat-runtime (optional)
   -> @hypercard/hypercard-runtime (optional)
   -> @hypercard/engine (required)

@hypercard/chat-runtime -> @hypercard/engine
@hypercard/hypercard-runtime -> @hypercard/engine
@hypercard/engine -> (no internal package deps)
```

The visibility-context redesign is the key decoupling mechanism: shell should not assume `state.chatProfiles`. Instead, packages that care about policy (chat, auth, org permissions, feature flags) should provide resolver logic into engine.

## Audience, Scope, and Reading Guide

### Audience

This document is written for an intern or new team member who has not worked in this repository before. It assumes familiarity with TypeScript and React, but not with the `go-go-os` package structure.

### In Scope

- Package DAG cleanup for engine/chat/runtime concerns.
- Story ownership cleanup (runtime stories must live with runtime owners).
- Visibility policy architecture for context-menu actions.
- Phased implementation and migration plan.

### Out of Scope

- Backend API redesign for chat/hypercard.
- Visual redesign of desktop shell widgets.
- Runtime behavior changes unrelated to dependency boundaries.

### Reading Order

1. Current-state architecture and evidence.
2. Why visibility exists and what it means.
3. Target architecture and APIs.
4. Phased implementation steps.
5. Testing and rollout risk controls.

## Problem Statement

The current package structure mixes three kinds of concerns:

1. **Engine core concerns**: desktop windowing, menu/context infrastructure, generic widgets.
2. **Runtime concerns**: hypercard plugin card session host and runtime reducers.
3. **Chat concerns**: timeline transport, profile state, WebSocket hydration, chat window composition.

This causes two concrete issues:

### 1) Dependency ownership is ambiguous

Evidence:

- Engine exports chat directly (`packages/engine/src/index.ts:15-18`).
- Engine chat barrel exports runtime hooks and WS machinery (`packages/engine/src/chat/index.ts:17-34`).
- Hypercard runtime depends on engine (`packages/hypercard-runtime/package.json:18`) and imports engine reducers/diagnostics in store factory (`packages/hypercard-runtime/src/app/createAppStore.ts:2-9`).

Result: the line between "desktop core" and "feature runtime" is unclear.

### 2) Story placement violates package boundaries

Evidence:

- Engine stories import runtime store factory from `@hypercard/hypercard-runtime`:
  - `DesktopShell.stories.tsx:5`
  - `ChatConversationWindow.stories.tsx:4`
  - `BookTracker.stories.tsx:5`
- Another engine story still imports a deleted engine-local createAppStore path:
  - `RuntimeCardDebugWindow.stories.tsx:4` (`../../app/createAppStore`)

Result: Storybook can fail for reasons unrelated to engine core and encourages cross-package imports that should not exist.

### 3) Desktop shell has implicit chat coupling

The shell visibility pipeline is generic in concept, but one resolver is hardcoded to chat profile state shape.

Evidence:

- `resolveActionVisibilityContext` reads `state.chatProfiles` (`useDesktopShellController.tsx:232-289`).
- Result is fed to generic visibility filtering (`useDesktopShellController.tsx:917-919`).

This is a soft coupling: engine does not import chat runtime symbols, but it still embeds knowledge of chat-state schema.

## Current-State Architecture (Evidence-Based)

### Workspace and package composition

`go-go-os` top-level scripts already include multiple packages (`package.json:10`):

- `packages/engine`
- `packages/hypercard-runtime`
- `packages/confirm-runtime`
- `packages/desktop-os`
- several apps

TS references confirm this multi-package shape (`tsconfig.json:4-12`).

### Runtime store ownership

`createAppStore` currently lives in hypercard runtime, not engine.

Evidence:

- Store factory file: `packages/hypercard-runtime/src/app/createAppStore.ts`.
- It includes reducers: `pluginCardRuntime`, `windowing`, `notifications`, `debug`, `hypercardArtifacts` (`createAppStore.ts:56-63`).

This means runtime packages compose an app store from engine + runtime reducers.

### Desktop shell policy flow

Desktop shell has a generic policy evaluator:

- Action visibility model: `types.ts:40-57`.
- Filter implementation: `contextActionVisibility.ts:101-123`.

At runtime, shell flow is:

```text
context target -> build raw actions -> derive visibility context -> apply filters -> open menu
```

Concrete invocation in code:

- `resolveActionVisibilityContext(store.getState(), target)` (`useDesktopShellController.tsx:917`)
- `applyActionVisibility(...)` (`useDesktopShellController.tsx:918`)

But context derivation is chat-aware by default (`state.chatProfiles`, lines `232-289`).

### Story ownership state

There is already an app-level BookTracker debug story at:

- `apps/book-tracker-debug/src/app/stories/BookTrackerDebugApp.stories.tsx`

And engine tests explicitly validate app story presence (`packages/engine/src/__tests__/storybook-app-smoke.test.ts:6-10`).

This is strong evidence that runtime demo stories can and should live outside engine.

## Gap Analysis Against Target Outcome

Target outcome from request:

1. Clean package DAG.
2. Stories in owning package/app.
3. Engine detached from chat/hypercard runtime assumptions.
4. Visibility mechanism understandable and controlled by external packages.

Gaps:

1. Engine still includes chat runtime source and exports chat publicly.
2. Engine stories still instantiate runtime stores and runtime cards.
3. Shell visibility context is not provider-driven.
4. Consumers have no first-class API to inject visibility identity/claims.

## What "Visibility" Means (and Why It Exists)

### Definition

In this codebase, visibility means **policy gating of context-menu actions before user execution**.

An action can be:

- shown and executable,
- shown but disabled,
- hidden.

This is controlled by action metadata (`allowedProfiles`, `allowedRoles`, optional `when`) plus an evaluated runtime context.

### Why this exists

Without visibility gating, every command handler must manually re-check permissions. That leads to inconsistent UX and duplicated policy logic.

Visibility system gives two protections:

1. **UX-level correctness**: users only see actions relevant to them.
2. **Command-level safety**: command dispatch path also checks allowed commands (`isContextCommandAllowed`).

### Is it chat-specific?

No. Chat is one use case (profiles/roles), but visibility is generally useful for:

- feature flags (`beta`, `internal`),
- workspace mode (`readonly`, `offline`),
- tenant policies,
- module capabilities,
- selection state constraints.

The system should therefore remain in engine, but context derivation should be externalized.

### Should we remove it entirely?

Possible but not recommended.

If removed:

- each module re-implements ad hoc policy checks,
- menu consistency regresses,
- disabled-vs-hidden semantics become fragmented.

Recommendation: keep visibility system, but make context provider generic and package-owned by consumers.

## Proposed Architecture

## 1) Package DAG and responsibilities

### `@hypercard/engine` (core)

Owns:

- desktop windowing and menu runtime,
- context/action models,
- policy evaluator primitives (`applyActionVisibility`),
- generic widgets and theme,
- notifications + base diagnostics.

Does not own:

- chat transport/profile logic,
- hypercard card runtime hosts,
- runtime demo apps.

### `@hypercard/chat-runtime` (new package)

Owns:

- everything under current `engine/src/chat/*`,
- chat windows and debug windows,
- chat reducers/selectors,
- ws/timeline/runtime hooks,
- chat-specific visibility resolver helpers.

Depends on:

- `@hypercard/engine` desktop/menu/widget primitives.

### `@hypercard/hypercard-runtime`

Owns:

- plugin card runtime reducers,
- artifact projection,
- plugin session host/renderers,
- runtime card tooling/debug windows,
- app store factory for runtime-composed apps.

Depends on:

- `@hypercard/engine`.

### `@hypercard/desktop-os`

Owns:

- launcher composition and module wiring,
- reducer composition contract,
- selection of optional packages (chat/hypercard).

## 2) Visibility-context redesign (generic and injectable)

Current flow mixes two concerns in engine controller:

- context derivation (chat-specific state knowledge)
- context application (generic action filtering)

Proposed split:

- Engine: apply filters only.
- External module: derive context from app state.

### API sketch

```ts
// engine: types
export interface DesktopVisibilityContextResolverArgs {
  state: unknown;
  target: DesktopContextTargetRef;
  invocation?: DesktopCommandInvocation;
}

export interface DesktopVisibilityContextResolver {
  resolve(args: DesktopVisibilityContextResolverArgs): DesktopActionVisibilityContext;
}
```

```ts
// engine: shell props
export interface DesktopShellProps {
  // existing props...
  visibilityContextResolver?: DesktopVisibilityContextResolver;
}
```

```ts
// engine: default fallback
const DEFAULT_RESOLVER: DesktopVisibilityContextResolver = {
  resolve: ({ target }) => ({ target }),
};
```

Controller use:

```ts
const resolver = props.visibilityContextResolver ?? DEFAULT_RESOLVER;
const visibilityContext = resolver.resolve({
  state: store.getState(),
  target,
  invocation,
});
const visibleItems = applyActionVisibility(rawItems, visibilityContext);
```

### Optional resolver composition utility

Allow multiple packages to contribute to one final context:

```ts
export function composeVisibilityResolvers(
  ...resolvers: DesktopVisibilityContextResolver[]
): DesktopVisibilityContextResolver {
  return {
    resolve(args) {
      return resolvers.reduce((acc, r) => ({ ...acc, ...r.resolve(args) }), {
        target: args.target,
      });
    },
  };
}
```

### Chat-side resolver (in new `chat-runtime` package)

```ts
export function createChatVisibilityResolver(opts?: {
  selectProfilesState?: (root: unknown) => ChatProfilesStateLike | null;
}): DesktopVisibilityContextResolver {
  const read = opts?.selectProfilesState ?? defaultChatProfilesSelector;

  return {
    resolve({ state, target }) {
      const profiles = read(state);
      if (!profiles) return { target };

      const scopeKey = target.conversationId ? `conv:${target.conversationId}` : undefined;
      const scoped = scopeKey ? profiles.selectedByScope?.[scopeKey] : undefined;
      const profile = scoped?.profile ?? profiles.selectedProfile ?? undefined;
      const registry = scoped?.registry ?? profiles.selectedRegistry ?? undefined;
      const roles = resolveProfileRoles(profiles.availableProfiles, profile);

      return { target, profile, registry, roles };
    },
  };
}
```

This preserves current behavior without embedding chat schema into engine.

## 3) Story relocation strategy

Stories should live where ownership and dependencies naturally belong.

### Move out of engine

1. `components/widgets/BookTracker.stories.tsx`
2. `components/widgets/RuntimeCardDebugWindow.stories.tsx`
3. `components/widgets/CodeEditorWindow.stories.tsx`
4. runtime-centric variants inside `DesktopShell.stories.tsx`
5. `components/widgets/ChatConversationWindow.stories.tsx` (to chat package)

### Destination

- `apps/book-tracker-debug` for app-specific BookTracker flows.
- `packages/hypercard-runtime` stories for runtime card/editor/session host demos.
- `packages/chat-runtime` stories for chat conversation/debug behaviors.
- keep only pure shell/windowing stories in engine.

## 4) Chat package extraction

Create `packages/chat-runtime` and migrate `packages/engine/src/chat/*`.

### Package exports

```ts
// packages/chat-runtime/src/index.ts
export * from './components/ChatConversationWindow';
export * from './debug/EventViewerWindow';
export * from './debug/TimelineDebugWindow';
export * from './runtime/useConversation';
export * from './runtime/useProfiles';
export * from './runtime/useSetProfile';
export * from './state/profileSlice';
export * from './state/chatSessionSlice';
export * from './state/chatWindowSlice';
export * from './state/timelineSlice';
export * from './visibility/createChatVisibilityResolver';
```

### Import adjustments

- `engine/src/index.ts`: remove `export * from './chat'`.
- desktop integration modules import chat APIs from `@hypercard/chat-runtime`.
- app stores include chat reducers from `@hypercard/chat-runtime` where needed.

## Detailed Rationale and Tradeoffs

## Why not keep chat in engine?

Pros of keeping in engine:

- fewer packages to manage,
- fewer import path migrations.

Cons:

- engine remains umbrella package with mixed responsibilities,
- harder to enforce clean DAG,
- Storybook breakages from unrelated runtime concerns,
- difficult for consumers who want shell without chat transport stack.

Decision: move chat out.

## Why resolver injection instead of hardcoded state shape?

Pros:

- engine stays agnostic to app schema,
- modules choose policy semantics,
- easier testing of policy behavior,
- future support for non-chat policies without shell modifications.

Cons:

- one extra integration step in app wiring,
- slight increase in API surface.

Decision: resolver injection is justified by decoupling value.

## Why not move all stories to one central "integration stories" package?

Pros:

- single place to browse mixed demos.

Cons:

- ownership ambiguity returns,
- package-specific story tests become weaker,
- contributors edit stories outside owning package.

Decision: place stories with owning package/app; optionally maintain a curated integration Storybook index later.

## Implementation Plan (Phased)

## Phase 0: Ticket and design alignment

Deliverables:

- this design doc,
- migration task breakdown,
- risk and validation checklist.

No runtime changes.

## Phase 1: Introduce resolver API in engine

Files (planned):

- `packages/engine/src/components/shell/windowing/types.ts`
- `packages/engine/src/components/shell/windowing/desktopShellTypes.ts`
- `packages/engine/src/components/shell/windowing/useDesktopShellController.tsx`
- `packages/engine/src/components/shell/windowing/index.ts`

Steps:

1. Add resolver types and props.
2. Add default resolver.
3. Replace chat-specific resolver call with injected resolver.
4. Remove internal `readChatProfilesState` logic from engine controller.

Validation:

- unit tests for default behavior (no resolver),
- unit tests for injected resolver and `unauthorized` behavior.

## Phase 2: Create `packages/chat-runtime`

Files (planned):

- new package scaffold (`package.json`, `tsconfig.json`, `src/index.ts`)
- move `engine/src/chat/**/*` into `chat-runtime/src/**/*`

Steps:

1. Copy/move chat files preserving structure.
2. Update imports from relative engine paths to package imports as needed.
3. Add dependency on `@hypercard/engine` for shared shell/widget primitives.
4. Export chat APIs from new package.
5. Remove `./chat` barrel export from engine root.

Validation:

- `npm run build -w packages/chat-runtime`
- workspace typecheck
- chat tests pass in new package location.

## Phase 3: Wire visibility resolver from chat-runtime in composition layers

Files (planned):

- `packages/desktop-os` integration points
- app modules that instantiate `DesktopShell` with chat actions

Steps:

1. Implement `createChatVisibilityResolver` in chat-runtime.
2. Compose resolver in app shell wiring.
3. Ensure context-menu actions gated by profiles/roles still behave exactly as before.

Validation:

- profile switch scenario test,
- message/conversation action visibility tests,
- command dispatch blocked when action hidden/disabled.

## Phase 4: Story relocation and cleanup

### 4A: BookTracker

- Keep canonical app story at `apps/book-tracker-debug/src/app/stories/BookTrackerDebugApp.stories.tsx`.
- Remove runtime-booktracker story from engine.

### 4B: Hypercard runtime editor/debug stories

Move from engine to `packages/hypercard-runtime`:

- `RuntimeCardDebugWindow.stories.tsx`
- `CodeEditorWindow.stories.tsx`
- any plugin runtime stories tied to artifact/card registry.

### 4C: Chat stories

Move to `packages/chat-runtime`:

- `ChatConversationWindow.stories.tsx`
- chat debug stories currently living in engine widgets taxonomy.

### 4D: Desktop shell stories

Keep in engine only if no runtime-specific imports are needed. Runtime-driven shell demos belong to `desktop-os` or integration stories.

Validation:

- Storybook taxonomy check script (`scripts/storybook/check-taxonomy.mjs`),
- `npm run storybook:check`,
- targeted story render smoke.

## Phase 5: Package DAG verification and final hardening

Steps:

1. Verify no `@hypercard/hypercard-runtime` imports in engine source/stories.
2. Verify no chat schema assumptions in engine controller.
3. Verify engine package dependencies do not include chat-only transport stack after extraction.
4. Update docs and migration notes.

Validation commands (expected):

```bash
rg -n "@hypercard/hypercard-runtime" packages/engine/src
rg -n "chatProfiles" packages/engine/src/components/shell/windowing
npm run build
npm run test
npm run storybook:check
```

## Pseudocode: End-to-End Flow After Refactor

```ts
// desktop-os composition layer
import { createChatVisibilityResolver } from '@hypercard/chat-runtime';
import { composeVisibilityResolvers } from '@hypercard/engine/desktop-react';

const visibilityResolver = composeVisibilityResolvers(
  createChatVisibilityResolver(),
  createFeatureFlagResolver(),
  createTenantPolicyResolver(),
);

<DesktopShell
  stack={stack}
  contributions={contributions}
  visibilityContextResolver={visibilityResolver}
/>
```

Engine runtime behavior:

```ts
function resolveContextMenuItemsForTarget(target) {
  const rawItems = buildRawContextMenuItemsForTarget(target);
  const ctx = (props.visibilityContextResolver ?? DEFAULT_RESOLVER)
    .resolve({ state: store.getState(), target });
  return applyActionVisibility(rawItems, ctx);
}
```

Action declaration by any package:

```ts
const actions = [
  {
    id: 'conversation.export',
    label: 'Export Transcript',
    commandId: 'conversation.export',
    visibility: {
      allowedRoles: ['admin'],
      unauthorized: 'hide',
    },
  },
  {
    id: 'message.create-task',
    label: 'Create Task',
    commandId: 'message.create-task',
    visibility: {
      allowedProfiles: ['agent'],
      unauthorized: 'disable',
    },
  },
];
```

## Diagrams

## Current (simplified)

```text
@hypercard/engine
  |- windowing core
  |- widgets
  |- chat runtime + ws + profile state
  |- some runtime-coupled stories

@hypercard/hypercard-runtime
  -> depends on @hypercard/engine

@hypercard/desktop-os
  -> depends on @hypercard/engine + @hypercard/hypercard-runtime
```

## Target (simplified)

```text
                +-------------------------+
                |     @hypercard/engine   |
                | windowing + core widgets|
                +-----------+-------------+
                            ^
                            |
        +-------------------+-------------------+
        |                                       |
+-------+------------------+       +------------+----------------+
| @hypercard/chat-runtime  |       | @hypercard/hypercard-runtime|
| chat UI + ws + profile   |       | plugin cards + artifacts     |
+-----------+--------------+       +---------------+--------------+
            ^                                      ^
            |                                      |
            +---------------+----------------------+ 
                            |
                 +----------+-----------+
                 | @hypercard/desktop-os|
                 | app/module composer   |
                 +----------+-----------+
                            ^
                            |
                         apps/*
```

## Risks and Mitigations

### Risk 1: Breaking public imports from `@hypercard/engine`

Mitigation:

- explicit migration PR with mechanical import updates,
- changelog note for package consumers,
- temporary branch-wide codemod script to rewrite imports.

### Risk 2: Story taxonomy churn

Mitigation:

- move stories in small batches,
- run taxonomy check after each batch,
- keep app-level stories stable.

### Risk 3: Visibility behavior regressions

Mitigation:

- snapshot/unit tests for action filtering,
- profile-role matrix tests,
- parity test: before/after resolver output for existing scenarios.

### Risk 4: Package dependency creep reappears

Mitigation:

- add CI guard queries checking forbidden imports,
- keep clear package ownership docs.

## Alternatives Considered

## Alternative A: Keep chat in engine, only move stories

Rejected because it fixes immediate story breakage but not ownership clarity or long-term DAG cleanliness.

## Alternative B: Keep hardcoded shell chat resolver and create adapter wrappers

Rejected because it preserves implicit schema coupling and prevents other packages from first-class policy injection.

## Alternative C: Remove visibility model and gate only in command handlers

Rejected because UX consistency and command safety degrade; policy logic duplicates across modules.

## Implementation Checklist (Execution-Oriented)

1. Add visibility resolver contract to engine.
2. Extract chat source tree into new `chat-runtime` package.
3. Rewire app imports from `@hypercard/engine` chat exports to `@hypercard/chat-runtime`.
4. Move runtime-coupled stories to owning packages/apps.
5. Remove legacy/stale story imports (`../../app/createAppStore` etc.).
6. Tighten CI checks for forbidden cross-package imports.
7. Update docs, ticket tasks, and changelog.

## Test Strategy

### Unit tests

- `contextActionVisibility` unchanged semantics.
- resolver fallback behavior in shell with no resolver.
- composed resolver merge behavior.

### Integration tests

- context action visibility for profile/role permutations.
- conversation/message context menu gating.
- command blocking via `isContextCommandAllowed` when action hidden/disabled.

### Storybook and build verification

```bash
npm run build -w packages/engine
npm run build -w packages/chat-runtime
npm run build -w packages/hypercard-runtime
npm run build -w packages/desktop-os
npm run storybook:check
npm run test
```

## Rollout Plan

1. Land resolver API first while chat remains in engine (no behavior change).
2. Move chat package and update imports.
3. Move stories package-by-package.
4. Remove engine chat exports.
5. Enforce DAG with CI checks.

This sequencing reduces risk by decoupling API and file-move complexity.

## Open Questions

1. Should `DesktopActionVisibilityContext` remain profile/role-centric, or move to a generic claims model in this same migration?
2. Should `desktop-os` own resolver composition by default, or should each app module provide it explicitly?
3. Do we want a dedicated integration-story package for cross-package demos after ownership cleanup?
4. Should engine package dependencies be split further (for example, moving codemirror/quickjs dependencies out of engine if they are only used by runtime/editor features)?

## References

Observed files used for this design:

- `go-go-os/package.json`
- `go-go-os/tsconfig.json`
- `go-go-os/packages/engine/package.json`
- `go-go-os/packages/engine/src/index.ts`
- `go-go-os/packages/engine/src/chat/index.ts`
- `go-go-os/packages/engine/src/components/shell/windowing/types.ts`
- `go-go-os/packages/engine/src/components/shell/windowing/contextActionVisibility.ts`
- `go-go-os/packages/engine/src/components/shell/windowing/desktopShellTypes.ts`
- `go-go-os/packages/engine/src/components/shell/windowing/useDesktopShellController.tsx`
- `go-go-os/packages/engine/src/components/shell/windowing/DesktopShell.stories.tsx`
- `go-go-os/packages/engine/src/components/widgets/BookTracker.stories.tsx`
- `go-go-os/packages/engine/src/components/widgets/ChatConversationWindow.stories.tsx`
- `go-go-os/packages/engine/src/components/widgets/RuntimeCardDebugWindow.stories.tsx`
- `go-go-os/packages/engine/src/__tests__/storybook-app-smoke.test.ts`
- `go-go-os/packages/hypercard-runtime/package.json`
- `go-go-os/packages/hypercard-runtime/src/index.ts`
- `go-go-os/packages/hypercard-runtime/src/app/createAppStore.ts`
- `go-go-os/packages/desktop-os/package.json`
- `go-go-os/packages/desktop-os/src/store/createLauncherStore.ts`
- `go-go-os/apps/book-tracker-debug/src/app/store.ts`
- `go-go-os/apps/book-tracker-debug/src/app/stories/BookTrackerDebugApp.stories.tsx`
- `go-go-app-arc-agi-3/apps/arc-agi-player/src/launcher/module.tsx`
- `go-go-app-arc-agi-3/apps/arc-agi-player/src/app/store.ts`
