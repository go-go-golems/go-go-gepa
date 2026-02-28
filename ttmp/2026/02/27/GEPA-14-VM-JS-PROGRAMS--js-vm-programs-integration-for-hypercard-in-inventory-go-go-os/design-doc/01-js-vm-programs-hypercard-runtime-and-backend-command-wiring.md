---
Title: JS VM Programs, HyperCard Runtime, and Backend Command Wiring
Ticket: GEPA-14-VM-JS-PROGRAMS
Status: active
Topics:
    - js-vm
    - hypercard
    - inventory-app
    - go-go-os
    - arc-agi
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../go-go-app-arc-agi-3/pkg/backendmodule/client.go
      Note: Upstream ARC API translation and reset/action contracts
    - Path: ../../../../../../../go-go-app-arc-agi-3/pkg/backendmodule/routes.go
      Note: |-
        ARC backend commands and session/game action routes
        ARC backend command/session route contract mapping
    - Path: ../../../../../../../go-go-app-inventory/apps/inventory/src/domain/pluginBundle.vm.js
      Note: |-
        Real-world VM card bundle patterns for local/domain/system action dispatch
        Reference VM bundle patterns for card/session/domain/system dispatch usage
    - Path: ../../../../../../../go-go-app-inventory/apps/inventory/src/domain/stack.ts
      Note: Plugin capability policy for inventory stack
    - Path: ../../../../../../../go-go-app-inventory/pkg/pinoweb/hypercard_events.go
      Note: Hypercard event types and timeline mapping for card lifecycle
    - Path: ../../../../../../../go-go-app-inventory/pkg/pinoweb/hypercard_extractors.go
      Note: Parser/validation for hypercard card v2 payloads
    - Path: ../../../../../../../go-go-app-inventory/pkg/pinoweb/prompts/runtime-card-policy.md
      Note: |-
        LLM authoring contract for runtime cards and handler APIs
        Runtime card generation contract provided to LLM output
    - Path: ../../../../../../../go-go-os/apps/arc-agi-player/src/api/arcApi.ts
      Note: |-
        Existing frontend API client pattern for ARC backend endpoints
        Existing frontend request pattern for ARC module endpoints
    - Path: ../../../../../../../go-go-os/packages/engine/src/app/createAppStore.ts
      Note: Root reducer shape and hypercardArtifacts slice key
    - Path: ../../../../../../../go-go-os/packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx
      Note: |-
        Runtime session loading, global-state projection, render/event dispatch wiring
        Runtime load/render/event orchestration and global-state projection
    - Path: ../../../../../../../go-go-os/packages/engine/src/components/shell/windowing/pluginIntentRouting.ts
      Note: |-
        Domain/system intent routing and command mapping
        Intent ingest and domain/system downstream dispatch behavior
    - Path: ../../../../../../../go-go-os/packages/engine/src/features/pluginCardRuntime/pluginCardRuntimeSlice.ts
      Note: |-
        Runtime state, capability gating, timeline, pending intent queues
        Capability gating
    - Path: ../../../../../../../go-go-os/packages/engine/src/hypercard/artifacts/artifactProjectionMiddleware.ts
      Note: |-
        Timeline artifact projection and runtime card registry registration
        Artifact upsert projection and runtime-card registration
    - Path: ../../../../../../../go-go-os/packages/engine/src/hypercard/artifacts/artifactRuntime.ts
      Note: |-
        Runtime card id/code extraction from hypercard.card.v2 payloads
        Runtime card id/code extraction from hypercard payloads
    - Path: ../../../../../../../go-go-os/packages/engine/src/plugin-runtime/runtimeService.ts
      Note: |-
        QuickJS VM lifecycle, timeouts, memory limits, render/event evaluation
        QuickJS VM session lifecycle
    - Path: ../../../../../../../go-go-os/packages/engine/src/plugin-runtime/stack-bootstrap.vm.js
      Note: |-
        VM bootstrap host API and dispatch* intent functions exposed to cards
        VM authoring API and dispatch* handler context
    - Path: ttmp/2026/02/27/GEPA-12-ARC-AGI-OS-BACKEND-MODULE--arc-agi-backend-module-integration-for-go-go-os-and-wesen-os/design-doc/01-arc-agi-backend-module-architecture-and-implementation-guide.md
      Note: Related backend architecture ticket context
    - Path: ttmp/2026/02/27/GEPA-13-ARC-AGI-WIDGET--arc-agi-game-controller-widget-for-go-go-os-desktop/design/01-arc-agi-player-widget-design.md
      Note: Related frontend widget ticket context
ExternalSources: []
Summary: End-to-end intern-oriented architecture guide for how HyperCard JS VM cards run in go-go-os today, how actions/state are exposed, and how to wire ARC backend commands into VM-driven cards with a concrete phased implementation plan.
LastUpdated: 2026-02-27T22:32:44-05:00
WhatFor: Provide a deep implementation blueprint for integrating ARC game control into HyperCard VM cards in inventory/go-go-os while preserving runtime safety and capability boundaries.
WhenToUse: Use when implementing or reviewing VM card runtime integration, backend command bridging, runtime-card generation policies, and ARC command wiring from hypercards.
---


# JS VM Programs, HyperCard Runtime, and Backend Command Wiring

## Executive Summary

This ticket documents, in implementation detail, how JavaScript VM-powered HyperCard cards currently run in `go-go-os` and `inventory`, and how to extend that runtime so cards can reliably drive `go-go-app-arc-agi` backend commands.

The most important findings are:

1. The runtime already exposes the right primitive APIs to VM cards: `dispatchCardAction`, `dispatchSessionAction`, `dispatchDomainAction`, and `dispatchSystemCommand` are provided in the VM handler context (`stack-bootstrap.vm.js:170-212`).
2. The runtime already projects global app state into VM cards (`PluginCardSessionHost.tsx:32-66`, `261-300`), but there is naming drift between what cards/policies assume and actual reducer keys (`artifacts` vs `hypercardArtifacts`).
3. Domain/system capability gating exists and is enforced (`pluginCardRuntimeSlice.ts:268-310`, `capabilityPolicy.ts:41-54`, `pluginIntentRouting.ts:94-106`).
4. Runtime card injection from chat-generated HyperCard events is already implemented and works (`artifactProjectionMiddleware.ts:24-26`, `artifactRuntime.ts:108-119`, `runtimeCardRegistry.ts:31-34`).
5. The ARC backend module already exposes stable routes for game/session/reset/action workflows (`routes.go:143-260`) and already has an existing frontend integration pattern (`arcApi.ts:16-69`, `ArcPlayerWindow.tsx:49-120`).

So this is not a greenfield problem. We do not need to invent a new VM runtime. We need a clean bridge layer that maps VM intents to ARC backend commands and reflects ARC state back into VM-visible state.

### Recommended direction

Use a dedicated `arc` domain bridge with explicit command/event contracts:

1. VM cards emit `dispatchDomainAction('arc', 'command.request', payload)` (or `dispatchSystemCommand('arc.command', payload)` if we prefer the system namespace).
2. A frontend bridge (listener middleware or host component effect) performs HTTP calls to `/api/apps/arc-agi/*`.
3. Bridge writes normalized ARC state into Redux (session/game/frame/action history/errors).
4. VM cards read ARC state from `globalState.domains.arc` and render playable cards.

This preserves runtime capability checks, keeps VM cards declarative, avoids direct fetch from VM code, and reuses the already-proven ARC route surface.

---

## Audience and Reading Order (Intern First)

If you are new to this codebase, read in this order:

1. `go-go-os/packages/engine/src/plugin-runtime/stack-bootstrap.vm.js` to understand what APIs VM cards get.
2. `runtimeService.ts` to understand how JS runs, times out, and returns UI/intents.
3. `PluginCardSessionHost.tsx` to understand the browser-side session/render/event loop.
4. `pluginCardRuntimeSlice.ts` and `pluginIntentRouting.ts` to understand where intents become app actions.
5. `artifactProjectionMiddleware.ts` + `artifactRuntime.ts` to understand how chat artifacts become runtime cards.
6. `go-go-app-arc-agi-3/pkg/backendmodule/routes.go` to understand what ARC commands are callable.
7. `go-go-os/apps/arc-agi-player/src/api/arcApi.ts` for working request payload examples.

---

## Problem Statement and Scope

### User objective

Enable HyperCard VM cards (including generated cards) to play or interact with ARC game sessions by wiring VM action intents to `go-go-app-arc-agi` backend commands, while exposing enough ARC state into the VM to render interactive cards.

### In scope

1. Deep architecture map of current VM runtime.
2. Exact action/state exposure points and constraints.
3. Gap analysis for ARC backend command wiring.
4. Concrete proposed contracts and phased implementation plan.

### Out of scope (this ticket)

1. Full production implementation of the new bridge.
2. UX polish for final ARC game cards.
3. Replacing the VM engine (QuickJS is already in place).

---

## Current-State Architecture (Evidence-Backed)

## 1) VM runtime lifecycle and execution model

### 1.1 Runtime creation and guardrails

`QuickJSCardRuntimeService` creates one VM per runtime session and applies limits:

- Memory limit: `runtime.setMemoryLimit(...)` (`runtimeService.ts:180`)
- Stack limit: `runtime.setMaxStackSize(...)` (`runtimeService.ts:181`)
- Interrupt deadline check: `runtime.setInterruptHandler(...)` (`runtimeService.ts:182`)

Default limits are currently:

- `memoryLimitBytes: 32 MiB`
- `stackLimitBytes: 1 MiB`
- `loadTimeoutMs: 1000`
- `renderTimeoutMs: 100`
- `eventTimeoutMs: 100`

(from `runtimeService.ts:32-38`).

This means VM card code is already sandboxed for accidental infinite loops and memory blowups.

### 1.2 Bundle loading and session model

`loadStackBundle(stackId, sessionId, code)` executes VM bundle code once for the session and stores the session VM in `this.vms` (`runtimeService.ts:201-218`).

Important behavior:

- Duplicate session IDs are rejected (`runtimeService.ts:202-204`).
- Bundle metadata is read through `globalThis.__stackHost.getMeta()` (`runtimeService.ts:196-199`).
- Initial session/card state and list of card IDs are exposed via bundle metadata (`contracts.ts:40-49`).

### 1.3 Runtime mutation support

Runtime supports live card mutation:

- `defineCard` (`runtimeService.ts:220-229`)
- `defineCardRender` (`runtimeService.ts:231-240`)
- `defineCardHandler` (`runtimeService.ts:242-251`)

Integration tests confirm dynamic card insertion and patching behavior (`runtimeService.integration.test.ts:118-162`).

This is the key mechanism used by runtime card injection from chat artifacts.

### 1.4 Render/event execution and type validation

- Render path: `renderCard(...)` evals `__stackHost.render(...)` and validates returned UI tree with `validateUINode` (`runtimeService.ts:253-271`, `uiSchema.ts:17-99`).
- Event path: `eventCard(...)` evals `__stackHost.event(...)` and validates returned intents with `validateRuntimeIntents` (`runtimeService.ts:273-293`, `intentSchema.ts:7-62`).

This gives us strong runtime boundary checks at host side.

---

## 2) VM authoring API (what card code can do)

The bootstrap source (`stack-bootstrap.vm.js`) defines VM globals:

- `defineStackBundle`
- `defineCard`
- `defineCardRender`
- `defineCardHandler`
- `ui.*` helper (`text`, `button`, `input`, `row`, `column`, `panel`, `badge`, `table`)

(see `stack-bootstrap.vm.js:37-127`).

During handler execution, VM cards get context with dispatch functions:

```js
{
  cardState,
  sessionState,
  globalState,
  dispatchCardAction,
  dispatchSessionAction,
  dispatchDomainAction,
  dispatchSystemCommand
}
```

(source: `stack-bootstrap.vm.js:203-212`).

### Intent shapes emitted by VM

Intent schema in host contracts (`contracts.ts:13-38`):

- `card`: `{ scope: 'card', actionType, payload? }`
- `session`: `{ scope: 'session', actionType, payload? }`
- `domain`: `{ scope: 'domain', domain, actionType, payload? }`
- `system`: `{ scope: 'system', command, payload? }`

---

## 3) Runtime host integration in React/Redux

`PluginCardSessionHost` is the runtime orchestration component.

### 3.1 Load flow

1. Register runtime session in Redux (`registerRuntimeSession`) with stack capabilities (`PluginCardSessionHost.tsx:112-119`).
2. Load plugin bundle via `runtimeService.loadStackBundle` (`141`).
3. Mark session ready (`146`).
4. Inject pending runtime cards from registry (`149-165`).
5. Apply bundle initial session/card state by dispatching runtime intents (`167-203`).

### 3.2 Render flow

- Compute projected global state from store (`261-283`).
- Call `runtimeService.renderCard(...)` with card/session/global state (`294-300`).
- Render returned UI tree through `PluginCardRenderer` (`372`).

### 3.3 Event flow

- UI events call `emitRuntimeEvent` (`309-354`).
- Host calls `runtimeService.eventCard(...)` (`318-326`).
- For each returned intent, host calls `dispatchRuntimeIntent(intent, ...)` (`332-340`).

---

## 4) How actions are currently exposed and processed

## 4.1 Local card/session state actions

`pluginCardRuntimeSlice` handles `card` and `session` intents synchronously:

- `patch`: shallow merge
- `set`: deep-path set via `{ path, value }`
- `reset`: clear object

(see `applyStateAction`, `pluginCardRuntimeSlice.ts:132-161`).

These are applied directly to runtime session state (`244-266`).

## 4.2 Domain/system intents and capability gates

- Domain capability check: `authorizeDomainIntent` (`268-273`).
- System capability check: `authorizeSystemIntent` (`289-293`).

Allowed domain/system intents are appended to pending queues:

- `pendingDomainIntents` (`275-283`)
- `pendingSystemIntents` (`304`)
- `pendingNavIntents` when `command` starts with `nav.` (`305-307`)

(all in `pluginCardRuntimeSlice.ts`).

## 4.3 Immediate routing behavior (important)

`dispatchRuntimeIntent` does two things:

1. Always dispatches `ingestRuntimeIntent(...)` first (`pluginIntentRouting.ts:82-88`).
2. Then, for allowed domain/system intents, immediately dispatches downstream actions:
   - Domain: action type `${domain}/${actionType}` (`69-79`, `108-110`)
   - System: mapped commands for `nav.go`, `nav.back`, `notify`, `window.close` (`30-67`, `113-117`)

This is validated in `plugin-intent-routing.test.ts:20-210`.

### Practical implication

There are two potential routing surfaces today:

1. Pending queues in runtime slice.
2. Immediate dispatch in `pluginIntentRouting.ts`.

Current codebase appears to rely mainly on immediate routing. A repo-wide search shows dequeue actions/selectors are defined but not consumed by non-test code (`rg` findings over `pluginCardRuntimeSlice` usage).

---

## 5) How state is exposed to VM cards

`projectGlobalState` in `PluginCardSessionHost.tsx:32-66` builds:

```ts
{
  self: { stackId, sessionId, cardId, windowId },
  domains: { ...rootState slices except known shell slices },
  nav: { current, param, depth, canBack },
  system: { focusedWindowId, runtimeHealth }
}
```

Known shell slices excluded are:

- `pluginCardRuntime`
- `windowing`
- `notifications`
- `debug`

(`PluginCardSessionHost.tsx:42-44`).

### Notable naming mismatch

`createAppStore` mounts artifacts under `hypercardArtifacts` (`createAppStore.ts:59`).

But inventory VM bundle reads `globalState.domains.artifacts` (`pluginBundle.vm.js:33-35`).

So generated/viewer cards depending on `domains.artifacts` can miss data unless an alias is introduced or bundle is updated.

### Policy/docs mismatch to watch

`runtime-card-policy.md` references `globalState.nav.cardId` (`runtime-card-policy.md:118`), while runtime projection uses `nav.current` (`PluginCardSessionHost.tsx:54`).

This is another intern trap and should be normalized.

---

## 6) HyperCard runtime-card ingestion pipeline

The card generation pipeline already exists and is robust.

### 6.1 Backend generation constraints

Inventory middleware injects runtime-card policy instructions:

- Policy middleware composes `widget-policy` + `runtime-card-policy` (`hypercard_middleware.go:33-35`).
- Generator middleware can enforce presence of `<hypercard:card:v2>` block (`109-117`).

### 6.2 Backend extraction and event emission

Runtime card extractor (`hypercard_extractors.go`) parses `<hypercard:card:v2>` payload and enforces required fields:

- `card.id` required (`309-316`)
- `card.code` required (`317-323`)

On success emits `hypercard.card.v2` event (`335-342`).

### 6.3 Frontend artifact projection and runtime injection

Frontend artifact projection listens to timeline events and upserts artifacts (`artifactProjectionMiddleware.ts:35-49`).

If artifact contains runtime card fields, it registers the runtime card (`24-26`).

Runtime card fields are extracted from hypercard payloads in `artifactRuntime.ts:108-119` and `195-217`.

`PluginCardSessionHost` injects registry cards into loaded sessions (`149-165`, `230-247`).

This is the bridge from assistant-generated card code to executable VM cards.

---

## 7) Inventory plugin bundle as reference implementation

Inventory stack declares capabilities for plugin cards (`stack.ts:42-48`):

- Domain: `inventory`, `sales`
- System: `nav.go`, `nav.back`, `notify`, `window.close`

The bundle demonstrates all runtime dispatch types:

- Domain dispatch examples:
  - `inventory/updateQty` (`pluginBundle.vm.js:367-373`)
  - `inventory/saveItem` (`378-381`)
  - `inventory/createItem` (`457-465`)
  - `inventory/receiveStock` (`526`)
- System dispatch examples:
  - `nav.go` (`190-192`, `217-224`)
  - `nav.back` (`347`, `636-637`, `681-682`)
  - `notify` (`383`, `471`, `531`, `595-596`)
- Card-local state updates:
  - multiple `dispatchCardAction('set'/'patch', ...)` calls (`362-365`, `440-443`, etc.)

This is exactly the pattern ARC bridge should follow, except domain/system target contract changes from inventory CRUD to ARC commands/state.

---

## 8) ARC backend module command/state surface

ARC module currently exposes namespaced routes:

- `GET /games` (`routes.go:25-36`)
- `GET /games/{id}` (`38-54`)
- `POST /sessions` (`56-78`)
- `GET /sessions/{id}` (`94-108`)
- `DELETE /sessions/{id}` (`108-121`)
- `POST /sessions/{id}/games/{game}/reset` (`143-151`, `195-208`)
- `POST /sessions/{id}/games/{game}/actions` (`152-159`, `216-260`)
- `GET /sessions/{id}/events` (`127-135`)
- `GET /sessions/{id}/timeline` (`138-140`)
- `GET /schemas/{id}` (`177-193`)

Backend enforces reset-before-action via per-session/game GUID map (`routes.go:232-239`, `sessions.go:29-52`).

Event timeline/state is kept in `SessionEventStore` (`events.go:36-95`).

The HTTP client maps to upstream ARC routes (`client.go:51-127`).

### Existing frontend consumer for ARC

`go-go-os/apps/arc-agi-player` already consumes this backend:

- API routes in RTK Query (`arcApi.ts:16-69`)
- Session lifecycle and action/reset flow in UI (`ArcPlayerWindow.tsx:49-120`)

This is the best available reference for payload shape and interaction sequence.

---

## Gap Analysis Against Requested Outcome

Requested outcome: “wire backend commands so hypercard cards can play the game or interact with it.”

### Gap A: No ARC domain in VM-capable inventory store

Inventory store currently has domain reducers for `inventory`, `sales`, chat/timeline/profile, etc. (`apps/inventory/src/app/store.ts:12-21`).

No reducer/middleware exists for ARC session/game/frame state in inventory runtime context.

### Gap B: No intent bridge for async backend operations

`dispatchRuntimeIntent` immediate domain routing emits plain Redux actions (`pluginIntentRouting.ts:69-79`, `108-110`).

There is no existing async effect layer bound to plugin runtime domain intents for backend HTTP calls.

### Gap C: Pending intent queues are not operationalized

Runtime slice stores pending domain/system intents (`pluginCardRuntimeSlice.ts:59-62`, `275-307`) but there is no non-test consumer path that drains/executes them.

This means asynchronous command orchestration is currently underdeveloped.

### Gap D: State projection naming drift

VM cards/policy examples assume `domains.artifacts` / `nav.cardId`, while host projection provides `domains.hypercardArtifacts` / `nav.current`.

This inconsistency will confuse generated cards and interns.

### Gap E: Capability declarations do not include ARC commands yet

Inventory plugin capability list currently permits only inventory/sales + nav/notify/window.close (`stack.ts:44-47`).

ARC command channels need explicit allowlisting.

---

## Proposed Solution

## 1) Introduce an ARC Bridge Domain (recommended)

Add a dedicated `arc` domain in inventory frontend state, with two layers:

1. `arcBridge` reducer state (session list, selected game, current frame, action history, request status/errors).
2. `arcBridge` effect processor (listener middleware or host component effect) that executes HTTP commands.

### Why domain bridge (instead of VM direct fetch)

1. Maintains capability and audit trails in runtime timeline.
2. Keeps VM cards deterministic and testable.
3. Reuses existing routing/security gates.
4. Avoids exposing arbitrary network APIs to untrusted/generated VM code.

---

## 2) Define explicit VM intent contract for ARC

Two valid contract styles exist.

### Option A (preferred): Domain contract

VM emits:

- `dispatchDomainAction('arc', 'command.request', {...})`
- `dispatchDomainAction('arc', 'state.patch', {...})` only if needed for local optimistic state.

Bridge processes `arc/command.request` payloads and writes backend results as normal reducer actions.

Suggested payload:

```ts
interface ArcCommandRequest {
  op:
    | 'games.list'
    | 'session.open'
    | 'session.close'
    | 'game.reset'
    | 'game.action'
    | 'session.events.poll'
    | 'session.timeline.get';
  requestId?: string;
  sessionId?: string;
  gameId?: string;
  action?: string;
  data?: Record<string, unknown>;
  reasoning?: unknown;
}
```

### Option B: System command contract

VM emits `dispatchSystemCommand('arc.command', payload)`.

This keeps ARC in system namespace but requires extending `pluginIntentRouting.ts` mapping.

Recommendation: prefer Option A to keep ARC as domain behavior and avoid overloading system commands.

---

## 3) Runtime intent processing architecture

### Short-term compatible approach

Keep existing `dispatchRuntimeIntent` immediate behavior for existing cards, and add ARC bridge on top of emitted domain actions.

In this model:

1. VM emits domain intent.
2. `ingestRuntimeIntent` records timeline and queue (current behavior).
3. Immediate dispatch emits `arc/command.request`.
4. ARC bridge consumes `arc/command.request` and performs HTTP call.

No breaking changes to existing inventory cards.

### Medium-term cleanup approach

Unify on queue-based consumption for all domain/system intents to avoid dual-path complexity.

This requires adding a runtime intent effect host that drains `pendingDomainIntents` and `pendingSystemIntents` and then dequeues by ID.

If we do this, we should deprecate immediate dispatch in `pluginIntentRouting.ts` to prevent duplicate execution.

---

## 4) ARC state exposure contract into VM

Expose normalized ARC state under `globalState.domains.arc`, e.g.:

```ts
interface ArcDomainState {
  games: Array<{ game_id: string; name?: string }>;
  sessions: Record<
    string,
    {
      status: 'active' | 'closed' | 'loading' | 'error';
      selectedGameId?: string;
      currentFrame?: Record<string, unknown>;
      events?: Array<Record<string, unknown>>;
      timeline?: Record<string, unknown>;
      error?: string;
    }
  >;
  ui: {
    activeSessionId?: string;
    loadingOps: Record<string, boolean>;
  };
}
```

VM cards then render ARC game controls by reading this state.

---

## 5) Fix projection aliases for generated cards

Update `projectGlobalState` output to include compatibility aliases:

```ts
domains: {
  ...domains,
  artifacts: domains.hypercardArtifacts ?? domains.artifacts,
}
nav: {
  ...,
  cardId: opts.cardId, // alias to current
  current: opts.cardId,
}
```

This aligns runtime with policy docs and reduces LLM-generated card breakage.

---

## 6) Capability policy updates

Extend inventory stack capabilities for ARC bridge.

If using domain contract:

- Add `arc` to plugin `capabilities.domain` in `stack.ts`.

If using system contract:

- Add `arc.command` (and any subcommands) to `capabilities.system` and update system router mapping.

Domain contract is cleaner and more composable.

---

## Concrete API/Flow Sketches

## A) VM card command emission example

```js
handlers: {
  startSession({ dispatchDomainAction }, args) {
    dispatchDomainAction('arc', 'command.request', {
      op: 'session.open',
      source_url: 'hypercard-vm',
      tags: ['hypercard', 'arc']
    });
  },
  resetGame({ dispatchDomainAction }, args) {
    dispatchDomainAction('arc', 'command.request', {
      op: 'game.reset',
      sessionId: String(args?.sessionId || ''),
      gameId: String(args?.gameId || ''),
    });
  },
  action({ dispatchDomainAction }, args) {
    dispatchDomainAction('arc', 'command.request', {
      op: 'game.action',
      sessionId: String(args?.sessionId || ''),
      gameId: String(args?.gameId || ''),
      action: String(args?.action || ''),
      data: args?.data || {},
    });
  }
}
```

## B) ARC bridge effect pseudocode

```ts
if (action.type === 'arc/command.request') {
  const cmd = action.payload as ArcCommandRequest;
  dispatch(arcCommandStarted(cmd));

  try {
    switch (cmd.op) {
      case 'games.list':
        dispatch(arcGamesReceived(await GET('/api/apps/arc-agi/games')));
        break;
      case 'session.open':
        dispatch(arcSessionOpened(await POST('/api/apps/arc-agi/sessions', cmd.data ?? {})));
        break;
      case 'game.reset':
        dispatch(arcFrameReceived(await POST(`/api/apps/arc-agi/sessions/${cmd.sessionId}/games/${cmd.gameId}/reset`, {})));
        break;
      case 'game.action':
        dispatch(arcFrameReceived(await POST(`/api/apps/arc-agi/sessions/${cmd.sessionId}/games/${cmd.gameId}/actions`, {
          action: cmd.action,
          data: cmd.data,
          reasoning: cmd.reasoning,
        })));
        break;
      // ...others
    }

    dispatch(arcCommandSucceeded(cmd));
  } catch (err) {
    dispatch(arcCommandFailed({ cmd, error: String(err) }));
  }
}
```

## C) VM card render using exposed ARC state

```js
render({ globalState }) {
  const arc = globalState?.domains?.arc || {};
  const activeSession = arc?.ui?.activeSessionId;
  const frame = activeSession ? arc?.sessions?.[activeSession]?.currentFrame : null;

  return ui.panel([
    ui.text('ARC Controller'),
    activeSession ? ui.badge('session: ' + activeSession) : ui.text('No session'),
    frame ? ui.text('state: ' + String(frame.state || 'unknown')) : ui.text('No frame yet'),
    ui.button('Open Session', { onClick: { handler: 'startSession' } }),
    ui.button('Reset', { onClick: { handler: 'resetGame', args: { sessionId: activeSession, gameId: 'bt11-fd9df0622a1a' } } }),
  ]);
}
```

---

## Implementation Plan (Phased, File-Level)

## Phase 0: Contract alignment and safety (small, high ROI)

1. Fix state projection aliases in `PluginCardSessionHost.tsx`.
2. Update runtime-card policy docs to reflect canonical nav/artifacts shape (or vice versa).
3. Add tests verifying alias presence in projected global state.

Files:

- `go-go-os/packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx`
- `go-go-app-inventory/pkg/pinoweb/prompts/runtime-card-policy.md`

## Phase 1: ARC domain state in inventory app

1. Add `arc` reducer slice with session/game/frame/event/timeline state.
2. Register reducer in inventory app store.
3. Add selectors for VM-facing state.

Files:

- `go-go-app-inventory/apps/inventory/src/features/arc/arcSlice.ts` (new)
- `go-go-app-inventory/apps/inventory/src/app/store.ts`

## Phase 2: ARC command bridge (frontend async layer)

1. Implement command effect executor for `arc/command.request` actions.
2. Map operations to `/api/apps/arc-agi/*` endpoints.
3. Normalize API failures to reducer error state + optional toast.

Files (one of these patterns):

- `.../features/arc/arcBridgeMiddleware.ts` (listener middleware)
- or `.../components/arc/ArcBridgeHost.tsx` (component side-effect runner)

If middleware pattern is chosen and `createAppStore` needs extensibility:

- `go-go-os/packages/engine/src/app/createAppStore.ts` (optional extension point for extra middleware)

## Phase 3: VM card contracts and capabilities

1. Add `arc` to stack plugin domain capabilities.
2. Add/update ARC HyperCard cards (static or runtime-generated policy examples).
3. Ensure generated cards use bridge command contract.

Files:

- `go-go-app-inventory/apps/inventory/src/domain/stack.ts`
- `go-go-app-inventory/apps/inventory/src/domain/pluginBundle.vm.js` (if adding built-in ARC cards)
- `go-go-app-inventory/pkg/pinoweb/prompts/runtime-card-policy.md`

## Phase 4: Queue model cleanup (optional, but recommended)

1. Introduce pending-intent consumer and explicit dequeue semantics.
2. De-duplicate immediate routing path in `pluginIntentRouting.ts`.
3. Keep compatibility toggle if needed.

Files:

- `go-go-os/packages/engine/src/components/shell/windowing/pluginIntentRouting.ts`
- `go-go-os/packages/engine/src/features/pluginCardRuntime/pluginCardRuntimeSlice.ts`
- new effect host/middleware for queue draining

## Phase 5: Observability and docs

1. Add debug panel for ARC bridge commands/results.
2. Add onboarding docs for interns and card authors.
3. Add sample runtime card templates for ARC actions.

---

## Testing and Validation Strategy

## Unit tests

1. `plugin runtime projection` tests:
   - projected global state contains `domains.artifacts` alias and `nav.cardId` alias.
2. `arc slice` tests:
   - reducers for command started/succeeded/failed paths.
3. `bridge` tests:
   - command payload maps to correct endpoint, method, and body.
   - reset-before-action error path surfaced correctly.
4. `capability policy` tests:
   - `arc` domain allowed/denied behavior.

## Integration tests

1. VM event intent path:
   - card emits `domain arc/command.request` -> backend mock called -> state updated -> rerender.
2. Artifact runtime card injection:
   - `hypercard.card.v2` event registers card and card can drive ARC bridge.
3. End-to-end mock test using existing arc mock handlers from `arc-agi-player` where possible.

## Manual smoke flow

1. Start launcher with inventory + arc backend modules.
2. Generate or open ARC control card.
3. List games.
4. Open session.
5. Reset game.
6. Send ACTION1..ACTION6 and verify frame changes.
7. Confirm timeline/events retrieval and session close.

---

## Detailed End-to-End Sequences

This section is intentionally explicit so an intern can debug the system with no prior context.

## Sequence 1: Generated card appears in runtime and becomes clickable

1. Assistant emits `<hypercard:card:v2>` payload with `card.id` and `card.code` per policy (`runtime-card-policy.md` and `hypercard_extractors.go` validation).
2. Inventory backend extractor parses YAML and emits `hypercard.card.v2` sem event (`hypercard_extractors.go:335-342`).
3. Frontend timeline projection receives event and upserts timeline entity (`hypercardCard.tsx:63-71`).
4. Artifact projection middleware inspects timeline entity props and extracts artifact/runtime card fields (`artifactProjectionMiddleware.ts:12-26`, `artifactRuntime.ts:195-217`).
5. `registerRuntimeCard(cardId, code)` stores card code in runtime registry (`runtimeCardRegistry.ts:31-34`).
6. `PluginCardSessionHost` injects pending cards on session load and on registry change (`PluginCardSessionHost.tsx:149-165`, `230-247`).
7. User navigates to that card id; runtime can now render and event it like built-in cards.

Failure points and what they mean:

1. Missing `runtime card.id/code` error at backend means model output violated tag policy contract.
2. Runtime injection failure with syntax error means extracted `card.code` is not a valid JS expression expected by `defineCard`.
3. Card not found at render means registry injection did not happen before navigation; check registry listeners and session readiness.

## Sequence 2: Button click in VM card updates local state

1. User clicks a `ui.button(... onClick.handler=...)` in rendered tree.
2. `PluginCardRenderer` calls `onEvent(handler, args)` (`PluginCardRenderer.tsx:25-36`, `85-90`).
3. Session host invokes `runtimeService.eventCard(...)` (`PluginCardSessionHost.tsx:318-326`).
4. VM handler calls `dispatchCardAction('patch'|'set'|'reset', payload)` in bootstrap host (`stack-bootstrap.vm.js:170-176`).
5. Host validates emitted intents and dispatches `ingestRuntimeIntent`.
6. Runtime slice mutates `cardState[cardId]` via `applyStateAction` (`pluginCardRuntimeSlice.ts:244-255`).
7. React rerender triggers `renderCard` with updated `cardState`.

This is the lowest-risk path and should be used for local UI controls (filters, drafts, toggles).

## Sequence 3: VM command triggers ARC backend operation (proposed bridge path)

1. VM handler emits `dispatchDomainAction('arc', 'command.request', cmdPayload)`.
2. Host routes domain intent through `dispatchRuntimeIntent`:
   - timeline/queue ingest occurs first,
   - downstream domain action `type='arc/command.request'` is dispatched if allowed.
3. ARC bridge effect sees `arc/command.request` and performs HTTP call against `/api/apps/arc-agi/*`.
4. Bridge dispatches state updates:
   - started/loading,
   - success with normalized payload,
   - failure with normalized error.
5. VM card rerenders reading `globalState.domains.arc`.

Key guarantee to preserve:

The VM never directly performs fetch. Network access is centralized in trusted frontend code.

## Sequence 4: ARC action fails due missing GUID

1. VM emits `game.action` command before reset.
2. ARC backend route checks GUID mapping and rejects with 400 (`routes.go:232-239`).
3. ARC bridge catches error and writes `commandFailed` state for the VM to render.
4. Card UI displays \"call reset first\" and disables action buttons until `game.reset` succeeds.

This is expected behavior and should be treated as UX logic, not runtime failure.

---

## Contract Catalog for Implementation

Use this as the single source of truth when implementing Phase 1/2.

## VM -> Bridge command contract

```ts
type ArcCommandOp =
  | 'games.list'
  | 'game.get'
  | 'session.open'
  | 'session.get'
  | 'session.close'
  | 'game.reset'
  | 'game.action'
  | 'session.events.poll'
  | 'session.timeline.get';

interface ArcCommandRequest {
  op: ArcCommandOp;
  requestId?: string;
  sessionId?: string;
  gameId?: string;
  action?: string; // ACTION1..ACTION7 for game.action
  data?: Record<string, unknown>;
  reasoning?: unknown;
  params?: Record<string, string | number | boolean>;
}
```

Validation rules:

1. `game.reset` requires `sessionId` and `gameId`.
2. `game.action` requires `sessionId`, `gameId`, `action`.
3. `action` must normalize to `ACTION1..ACTION7` (match backend expectations).
4. `session.events.poll` optional `after_seq` comes via `params`.

## Bridge -> Backend endpoint mapping

```ts
const endpointByOp = {
  'games.list':        ['GET',    '/api/apps/arc-agi/games'],
  'game.get':          ['GET',    '/api/apps/arc-agi/games/{gameId}'],
  'session.open':      ['POST',   '/api/apps/arc-agi/sessions'],
  'session.get':       ['GET',    '/api/apps/arc-agi/sessions/{sessionId}'],
  'session.close':     ['DELETE', '/api/apps/arc-agi/sessions/{sessionId}'],
  'game.reset':        ['POST',   '/api/apps/arc-agi/sessions/{sessionId}/games/{gameId}/reset'],
  'game.action':       ['POST',   '/api/apps/arc-agi/sessions/{sessionId}/games/{gameId}/actions'],
  'session.events.poll':['GET',   '/api/apps/arc-agi/sessions/{sessionId}/events?after_seq={n}'],
  'session.timeline.get':['GET',  '/api/apps/arc-agi/sessions/{sessionId}/timeline'],
} as const;
```

## Bridge state contract exposed to VM

```ts
interface ArcBridgeState {
  games: Array<{ game_id: string; name?: string }>;
  sessions: Record<
    string,
    {
      session_id: string;
      status: 'active' | 'closed' | 'loading' | 'error';
      selected_game_id?: string;
      current_frame?: Record<string, unknown>;
      last_action?: string;
      last_error?: string;
      events_after_seq?: number;
      events?: Array<Record<string, unknown>>;
      timeline?: Record<string, unknown>;
    }
  >;
  requests: Record<
    string,
    {
      op: string;
      status: 'pending' | 'success' | 'error';
      started_at: string;
      ended_at?: string;
      error?: string;
    }
  >;
  ui: {
    active_session_id?: string;
    active_game_id?: string;
  };
}
```

Card-author guidance:

1. Render with defensive null checks (`globalState?.domains?.arc ?? {}`).
2. Never assume `current_frame` exists before reset.
3. Disable action buttons unless `available_actions` includes the action.

---

## File-by-File Implementation Blueprint

This is the most literal implementation plan to hand to an intern.

## `go-go-app-inventory/apps/inventory/src/features/arc/arcSlice.ts` (new)

Implement:

1. `commandRequested`
2. `commandStarted`
3. `commandSucceeded`
4. `commandFailed`
5. `setActiveSession`
6. `setActiveGame`

Reducer behavior:

1. Keep normalized per-session frame/event/timeline.
2. Keep request status map keyed by `requestId`.
3. Bound request history length if needed (avoid unbounded growth).

## `go-go-app-inventory/apps/inventory/src/features/arc/arcBridge.ts` (new)

Implement command executor:

1. Accept plain Redux action `arc/command.request`.
2. Perform HTTP using `fetch` or shared helper.
3. Dispatch start/success/failure actions.
4. Normalize backend error envelopes (`{ error: { message } }`) into string.

If this is listener middleware:

1. Register middleware in store setup.
2. Ensure middleware runs after default middleware.

If this is component effect host:

1. Mount once near app root.
2. Subscribe to queued commands from state.
3. Dequeue after completion.

## `go-go-app-inventory/apps/inventory/src/app/store.ts`

1. Register new `arc` reducer in `createAppStore` domain reducers.
2. If middleware approach chosen and `createAppStore` cannot inject middleware, either:
   - extend engine API to accept extra middleware factories, or
   - build custom store in inventory app.

## `go-go-app-inventory/apps/inventory/src/domain/stack.ts`

1. Add `arc` to plugin `capabilities.domain`.
2. Optionally keep `inventory` and `sales` unchanged for backward compatibility.

## `go-go-app-inventory/apps/inventory/src/domain/pluginBundle.vm.js`

1. Add one reference card (`arcControl`) with handlers:
   - list games,
   - open session,
   - reset,
   - action.
2. Use `dispatchDomainAction('arc', 'command.request', ...)`.
3. Read `globalState.domains.arc` for status/frame/action availability.

## `go-go-os/packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx`

1. Add projection aliases:
   - `domains.artifacts` alias to `domains.hypercardArtifacts`.
   - `nav.cardId` alias to `nav.current`.
2. Keep existing projection fields to avoid regression.

## `go-go-app-inventory/pkg/pinoweb/prompts/runtime-card-policy.md`

1. Add ARC command examples using chosen command contract.
2. Clarify correct runtime projection names (`nav.current` or aliases if introduced).
3. Add explicit warning that runtime cards should not invent unknown domain/system names.

---

## Intern Debugging Runbook

When a VM ARC card \"does nothing\", debug in this fixed order.

1. Did VM emit intent?
   - Check `pluginCardRuntime.timeline` entry for this session/card.
   - If missing: handler not wired or UI event ref incorrect.
2. Was intent denied?
   - Check timeline `outcome='denied'` and `reason`.
   - If denied: capability policy mismatch in stack config.
3. Did domain action dispatch?
   - Confirm `arc/command.request` appears in Redux devtools/logging.
4. Did bridge call backend?
   - Check network request to `/api/apps/arc-agi/...`.
5. Did backend reject?
   - Parse `error.message` from response.
   - Common case: missing guid due reset not called first.
6. Did reducer update state?
   - Verify `state.arc.sessions[sessionId]` updates.
7. Did VM read the correct path?
   - Confirm card reads `globalState.domains.arc`.

Top symptoms and likely causes:

1. \"Card not found\" after generation:
   - runtime card not injected; inspect artifact projection and registry.
2. \"Runtime interrupted\":
   - heavy loop in render/handler exceeded timeout.
3. Button click no-op with no error:
   - capability denied and silently blocked at routing layer.
4. \"Artifact not found\" in viewer:
   - state path drift (`artifacts` alias missing).

---

## Risks, Constraints, and Mitigations

## Risk 1: Duplicate execution from dual routing paths

Cause: both immediate routing and pending queues exist.

Mitigation:

1. Keep one clear authoritative path for ARC commands (immediate domain action + bridge consumer).
2. If queue path is adopted, deprecate immediate path in same milestone.

## Risk 2: Policy/runtime drift for state shape

Cause: docs and runtime projection disagree (`artifacts`, `nav.cardId`).

Mitigation:

1. Add compatibility aliases.
2. Add projection contract tests.
3. Keep one documented canonical state shape for card authors.

## Risk 3: Generated card safety

Cause: generated runtime cards may emit unsupported actions or malformed payloads.

Mitigation:

1. Keep capability allowlists strict.
2. Validate bridge command payloads and return structured errors.
3. Improve runtime-card policy prompt examples for ARC commands.

## Risk 4: ARC session guid contract failures

Cause: action before reset.

Mitigation:

1. Bridge tracks session/game readiness state.
2. VM cards disable action buttons until reset success.
3. Surface backend `missing guid` errors in card UI.

---

## Alternatives Considered

## Alternative A: Fetch directly inside VM cards

Rejected because:

1. Breaks host capability boundaries.
2. Harder to test and audit.
3. Exposes too much network surface to generated code.

## Alternative B: Encode ARC only as system commands

Possible, but less expressive than domain contract and conflates app-domain behavior with shell-level commands.

## Alternative C: Reuse full `arc-agi-player` store/module inside inventory VM host

Possible long-term, but high coupling and heavier integration than required for command wiring. Better to share API client contracts and keep inventory bridge lightweight.

---

## Open Questions

1. Should ARC bridge command namespace be `domain: arc` (recommended) or `system: arc.*`?
2. Do we want to formalize queue-based intent execution now, or defer to a runtime refactor ticket?
3. Should projection aliases be temporary compatibility shims or long-term canonical contract?
4. Should generated ARC cards be constrained to a curated command subset initially (`session.open`, `game.reset`, `game.action`) for safety?

---

## Recommended Next Steps (Actionable)

1. Implement Phase 0 immediately (alias and policy alignment) to remove current authoring drift.
2. Implement Phase 1 + 2 together (arc domain state + command bridge).
3. Add one built-in static ARC control card in `pluginBundle.vm.js` as reference card.
4. Extend runtime-card policy with ARC-specific examples and payload schema.
5. Decide in follow-up ticket whether to migrate to queue-first runtime intent execution.

---

## References (Primary Evidence)

### VM runtime core

- `go-go-os/packages/engine/src/plugin-runtime/runtimeService.ts:32-38`
- `go-go-os/packages/engine/src/plugin-runtime/runtimeService.ts:167-186`
- `go-go-os/packages/engine/src/plugin-runtime/runtimeService.ts:201-218`
- `go-go-os/packages/engine/src/plugin-runtime/runtimeService.ts:220-251`
- `go-go-os/packages/engine/src/plugin-runtime/runtimeService.ts:253-305`
- `go-go-os/packages/engine/src/plugin-runtime/stack-bootstrap.vm.js:122-233`
- `go-go-os/packages/engine/src/plugin-runtime/contracts.ts:13-172`
- `go-go-os/packages/engine/src/plugin-runtime/intentSchema.ts:7-62`
- `go-go-os/packages/engine/src/plugin-runtime/uiSchema.ts:17-99`
- `go-go-os/packages/engine/src/plugin-runtime/runtimeService.integration.test.ts:22-162`

### Host/render/routing integration

- `go-go-os/packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx:32-373`
- `go-go-os/packages/engine/src/components/shell/windowing/PluginCardRenderer.tsx:38-149`
- `go-go-os/packages/engine/src/components/shell/windowing/pluginIntentRouting.ts:30-119`
- `go-go-os/packages/engine/src/features/pluginCardRuntime/pluginCardRuntimeSlice.ts:47-369`
- `go-go-os/packages/engine/src/features/pluginCardRuntime/capabilityPolicy.ts:1-55`
- `go-go-os/packages/engine/src/__tests__/plugin-intent-routing.test.ts:20-210`
- `go-go-os/packages/engine/src/__tests__/plugin-card-runtime.test.ts:18-246`

### Artifact/runtime-card pipeline

- `go-go-os/packages/engine/src/hypercard/artifacts/artifactProjectionMiddleware.ts:7-52`
- `go-go-os/packages/engine/src/hypercard/artifacts/artifactRuntime.ts:95-220`
- `go-go-os/packages/engine/src/hypercard/artifacts/artifactsSlice.ts:41-134`
- `go-go-os/packages/engine/src/plugin-runtime/runtimeCardRegistry.ts:27-94`
- `go-go-os/packages/engine/src/hypercard/timeline/hypercardCard.tsx:27-133`
- `go-go-os/packages/engine/src/hypercard/artifacts/artifactProjectionMiddleware.test.ts:59-93`

### Inventory runtime-card policy and bundle

- `go-go-app-inventory/apps/inventory/src/domain/stack.ts:42-48`
- `go-go-app-inventory/apps/inventory/src/domain/pluginBundle.vm.js:21-35`
- `go-go-app-inventory/apps/inventory/src/domain/pluginBundle.vm.js:367-390`
- `go-go-app-inventory/apps/inventory/src/domain/pluginBundle.vm.js:445-532`
- `go-go-app-inventory/apps/inventory/src/domain/pluginBundle.vm.js:600-685`
- `go-go-app-inventory/apps/inventory/src/features/inventory/inventorySlice.ts:13-39`
- `go-go-app-inventory/pkg/pinoweb/prompts/runtime-card-policy.md:104-154`
- `go-go-app-inventory/pkg/pinoweb/hypercard_middleware.go:33-55`
- `go-go-app-inventory/pkg/pinoweb/hypercard_extractors.go:83-91`
- `go-go-app-inventory/pkg/pinoweb/hypercard_extractors.go:309-342`
- `go-go-app-inventory/pkg/pinoweb/hypercard_events.go:27-34`
- `go-go-app-inventory/pkg/pinoweb/hypercard_events.go:253-377`

### ARC backend command surface

- `go-go-app-arc-agi-3/pkg/backendmodule/module.go:92-120`
- `go-go-app-arc-agi-3/pkg/backendmodule/routes.go:25-260`
- `go-go-app-arc-agi-3/pkg/backendmodule/client.go:51-127`
- `go-go-app-arc-agi-3/pkg/backendmodule/sessions.go:18-70`
- `go-go-app-arc-agi-3/pkg/backendmodule/events.go:36-95`
- `go-go-app-arc-agi-3/pkg/backendmodule/reflection.go:27-55`
- `go-go-app-arc-agi-3/docs/arc-agi-app-module-user-guide.md:95-233`

### Existing ARC frontend integration pattern

- `go-go-os/apps/arc-agi-player/src/api/arcApi.ts:16-69`
- `go-go-os/apps/arc-agi-player/src/components/ArcPlayerWindow.tsx:49-120`
- `go-go-os/apps/arc-agi-player/src/features/arcPlayer/arcPlayerSlice.ts:24-65`
- `go-go-os/apps/arc-agi-player/src/launcher/module.tsx:35-103`

### Related tickets consulted

- `go-go-gepa/ttmp/2026/02/27/GEPA-12-ARC-AGI-OS-BACKEND-MODULE--arc-agi-backend-module-integration-for-go-go-os-and-wesen-os/design-doc/01-arc-agi-backend-module-architecture-and-implementation-guide.md`
- `go-go-gepa/ttmp/2026/02/27/GEPA-13-ARC-AGI-WIDGET--arc-agi-game-controller-widget-for-go-go-os-desktop/design/01-arc-agi-player-widget-design.md`
