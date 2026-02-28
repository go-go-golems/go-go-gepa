---
Title: 'Intern Q&A: Store Boundaries, Sessions, Runtime Intents, and Re-rendering'
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
    - Path: ../../../../../../../go-go-app-arc-agi-3/apps/arc-agi-player/src/api/arcApi.ts
      Note: Existing frontend API mapping to ARC backend routes
    - Path: ../../../../../../../go-go-app-arc-agi-3/apps/arc-agi-player/src/app/store.ts
      Note: |-
        ARC player store composition and ownership boundary
        ARC app store boundary evidence
    - Path: ../../../../../../../go-go-app-arc-agi-3/apps/arc-agi-player/src/launcher/module.tsx
      Note: Per-window store lifetime in ARC app host
    - Path: ../../../../../../../go-go-app-arc-agi-3/pkg/backendmodule/routes.go
      Note: |-
        ARC backend command/session/events route surface
        ARC backend route contract for runtime bridge
    - Path: ../../../../../../../go-go-os/packages/engine/src/app/createAppStore.ts
      Note: Shared engine base reducer model
    - Path: ../../../../../../../go-go-os/packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx
      Note: |-
        Runtime load/render/event orchestration and global state projection
        Rerender and emitRuntimeEvent orchestration evidence
    - Path: ../../../../../../../go-go-os/packages/engine/src/components/shell/windowing/pluginIntentRouting.ts
      Note: |-
        Runtime intent ingest and immediate downstream dispatch behavior
        dispatchRuntimeIntent ingest+forward behavior
    - Path: ../../../../../../../go-go-os/packages/engine/src/components/shell/windowing/useDesktopShellController.tsx
      Note: session id generation and card window creation behavior
    - Path: ../../../../../../../go-go-os/packages/engine/src/desktop/core/state/windowingSlice.ts
      Note: cardSessionId nav stack creation and cleanup
    - Path: ../../../../../../../go-go-os/packages/engine/src/features/pluginCardRuntime/pluginCardRuntimeSlice.ts
      Note: |-
        Runtime session/card state, timeline, and pending queue semantics
        Session intents and pending queue behavior
    - Path: ../../../../../../../go-go-os/packages/engine/src/features/pluginCardRuntime/selectors.ts
      Note: Timeline and pending-queue selectors
    - Path: ../../../../../../../go-go-os/packages/engine/src/plugin-runtime/contracts.ts
      Note: Runtime intent type contracts
    - Path: ../../../../../../../go-go-os/packages/engine/src/plugin-runtime/runtimeService.ts
      Note: QuickJS session lifecycle and load/render/event/dispose contract
    - Path: ../../../../../../../go-go-os/packages/engine/src/plugin-runtime/stack-bootstrap.vm.js
      Note: |-
        VM-facing dispatch APIs available to card handlers
        VM dispatch* APIs and intent creation
    - Path: ttmp/2026/02/27/GEPA-17-PLUGIN-EVENT-VIEWER--js-plugin-runtime-event-viewer-for-inbound-ui-events-and-outbound-dispatched-actions/design-doc/01-plugin-runtime-event-viewer-architecture-and-implementation-plan.md
      Note: Event viewer deep-dive referenced by this Q&A
ExternalSources: []
Summary: Detailed intern-facing answers for GEPA-14 follow-up questions on store separation, generic HyperCard architecture, session lifetimes, intent flow, and pending queue strategy.
LastUpdated: 2026-02-28T00:10:00-05:00
WhatFor: Answer GEPA-14 follow-up architecture questions with implementation-grade guidance and diagrams for intern onboarding.
WhenToUse: Use when implementing ARC gameplay cards, runtime intent routing, session lifecycle behavior, and rerender/event debug behavior in JS plugin stacks.
---


# Intern Q&A: Store Boundaries, Sessions, Runtime Intents, and Re-rendering

## Executive Summary

Short answers first:

1. `arc-agi` should have its own app store. It already does in `arc-agi-player` (`store.ts:12-24`) and per-window host creation (`module.tsx:71-77`). Keep that boundary.
2. HyperCard cards/artifacts should move to a reusable "HyperCard Platform" package boundary in `go-go-os/packages/engine`, with inventory and arc-agi as consumers. Do not keep inventory assumptions in runtime contracts.
3. Your rerender bug is real: card render currently depends on selected runtime/session/windowing state, not all domain slice changes. If stock updates do not touch those dependencies, render will not rerun.
4. A `sessionId` is the host-side identity of one plugin runtime instance + one card nav stack. It is created when opening a card window and usually dies when that window closes.
5. `emitRuntimeEvent` is the host bridge from UI interaction to VM handler execution. `dispatchRuntimeIntent` is the host router from VM intents into Redux actions. `ingestRuntimeIntent` is the reducer entry point that records/applies/queues runtime intents in `pluginCardRuntime` state.
6. "Session intents" (GEPA-14 section 4.1) are `scope: 'session'` intents that mutate `sessionState` (shared across all cards in a runtime session).
7. Pending queues are not useless, but they are underused right now. Decide one model: keep and drain them intentionally (queue-first), or remove them and keep immediate routing only.

The rest of this document explains each answer in detail with architecture diagrams, pseudocode, and migration plans.

## Audience and Approach

This is written for a new intern who is starting from zero context. The goal is to make runtime behavior explainable from source code, not intuition.

Reading order if you are new:

1. Session and store mental model.
2. Intent flow end-to-end.
3. Rerender mechanics and the bug.
4. Reorg plan for generic HyperCard platform.
5. Pending queue strategy decision.

## 1) Should ARC-AGI have its own store?

Yes. Keep ARC-AGI as its own application state boundary.

Evidence:

- ARC app has dedicated store factory with app/domain reducers in `arc-agi-player/src/app/store.ts:12-24`.
- ARC app creates a per-window store instance in `ArcPlayerHost` (`module.tsx:72-75`), then wraps window content in Redux `Provider` (`module.tsx:76`).
- Engine has shared base reducers pattern in `createAppStore.ts:54-61` where engine slices are standardized and domain reducers are appended.

Why this is good:

- ARC domain state (`arcPlayer`, `arcApi`) is high-churn and game-specific. Inventory does not need to subscribe to it.
- It keeps development velocity: ARC feature changes do not force cross-app reducer contracts.
- It supports multiple app windows with isolated local state if needed.

What to standardize across stores:

- Base engine slices: `pluginCardRuntime`, `windowing`, `notifications`, `debug`, `hypercardArtifacts`.
- Cross-app runtime conventions: global projection shape and domain namespace contracts.

Recommended policy:

- One store per app shell/window host.
- Shared engine contract package for runtime projection and intents.
- Explicit bridge reducers for domain commands (for example `arcBridge`).

## 2) How to make HyperCard cards/artifacts generic beyond inventory? (multirepo reorg)

### Current issue

Runtime mechanisms are generic, but usage patterns still leak inventory assumptions:

- Inventory VM bundle reads domain slices like `globalState.domains.inventory` and `globalState.domains.sales` (`pluginBundle.vm.js:25-31`).
- Runtime-card policy examples are inventory-specific (`runtime-card-policy.md:114-117`).
- Artifact helper defaults stack id to `inventory` (`artifactRuntime.ts:252`).

This makes it easy to think HyperCard == inventory plugin, which is not the intent.

### Target architecture

```text
+-----------------------+        +---------------------+
|   go-go-app-inventory |        | go-go-app-arc-agi-3 |
|  (domain app)         |        | (domain app)        |
|  inventory reducers   |        | arc reducers        |
|  inventory card packs |        | arc card packs      |
+-----------+-----------+        +----------+----------+
            |                               |
            v                               v
+-------------------------------------------------------+
| go-go-os/packages/engine (HyperCard Platform core)    |
| - runtimeService (QuickJS)                            |
| - PluginCardSessionHost                               |
| - intent contracts + routing                          |
| - artifacts projection + runtime card registry        |
| - debug/event viewer hooks                            |
+-------------------------------------------------------+
```

### Proposed multirepo organization

1. `go-go-os/packages/engine`
   - Owns runtime engine, intent contracts, session host, artifact runtime primitives.
   - No inventory-specific domain names in core APIs.

2. `go-go-app-inventory`
   - Owns inventory stack bundle(s), inventory domain reducers, inventory prompts/policies.

3. `go-go-app-arc-agi-3`
   - Owns ARC stack bundle(s), ARC domain reducers, ARC command adapters and typed API client.

4. `go-go-gepa` docs tickets
   - Architecture decisions and migration runbooks.

### Practical refactor checklist

1. Introduce app-neutral global projection contract docs (`globalState.domains.<domainName>` only).
2. Remove `inventory` default from artifact open payload builder; require stack id from artifact metadata.
3. Split prompt/policy templates into:
   - generic runtime-card policy
   - app-specific supplements (inventory, arc).
4. Publish reusable card helpers as packages (for example `@hypercard/cards-core`, `@hypercard/cards-arc`).
5. Add sample ARC stack bundle in arc-agi repo, not engine repo.

### API contract example for generic cards

```ts
// app-neutral runtime projection
interface RuntimeGlobalProjection {
  self: { stackId: string; sessionId: string; cardId: string; windowId: string };
  nav: { current: string; param?: string; depth: number; canBack: boolean };
  system: { focusedWindowId: string | null; runtimeHealth: { status: string } };
  domains: Record<string, unknown>; // domain-owned namespaces only
}
```

## 3) Why state changes do not rerender cards now

### What currently triggers rerender

In `PluginCardSessionHost`:

- `tree` is computed in `useMemo` (`PluginCardSessionHost.tsx:286-307`).
- Dependencies are:
  - `cardState`
  - `sessionState`
  - `currentCardId`
  - runtime status and projection callback dependencies.

`projectGlobal` callback is recreated from selected deps (`261-283`), and it calls `store.getState()` inside (`263`), but it does **not** subscribe to arbitrary domain slice changes.

### Why stock updates can be invisible

If a stock change only updates `domains.inventory` and does not change:

- current nav,
- focused window,
- runtime status,
- card/session local state,

then `useMemo` dependencies may not change, so `renderCard` is not re-invoked.

### Diagram of current behavior

```text
Redux domain update (inventory.qty)
            |
            v
PluginCardSessionHost selectors changed?
(cardState/sessionState/nav/runtime/focus)
            |
       +----+----+
       |         |
      yes        no
       |         |
  useMemo reruns  cached tree reused
  renderCard()    (no rerender)
```

### Recommended fix options

Option A (recommended): Add explicit domain projection selector subscription.

- Use a selector that computes a stable projection fingerprint for `globalState.domains` relevant to current card.
- Include fingerprint in `tree` dependencies.

Option B: Add per-session `globalRevision` counter.

- Increment when any selected domain slice changes.
- Depend on `globalRevision` in `useMemo`.

Option C: Force render on all store updates.

- Simple but expensive; avoid unless debug mode.

### Example fix sketch (Option A)

```ts
const projectedDomainSnapshot = useSelector((state) => {
  const projection = projectDomainStateForCard(state, currentCardId);
  return stableHash(projection); // or memoized reference
});

const tree = useMemo(() => {
  const projectedGlobalState = projectGlobal();
  return runtimeService.renderCard(..., projectedGlobalState);
}, [cardState, sessionState, currentCardId, projectedDomainSnapshot, projectGlobal, runtimeSession?.status]);
```

## 4) Session ID, session lifetime, and what a session is for

### What a session is

A session in this runtime has three coupled meanings:

1. Runtime session: a QuickJS VM instance in `runtimeService.vms` keyed by `sessionId` (`runtimeService.ts:158`, `201-212`).
2. UI navigation session: card nav stack in `windowing.sessions[sessionId]` (`windowingSlice.ts:14`, `61-68`).
3. State scope: per-session `sessionState` and per-card state map in `pluginCardRuntime.sessions[sessionId]` (`pluginCardRuntimeSlice.ts:47-53`, `203-213`).

So "session" is not just an ID; it is the coherence key joining VM, nav history, and runtime state.

### Where sessionId comes from

- Generated by `nextSessionId()` in desktop controller (`useDesktopShellController.tsx:117-121`).
- Used when opening card windows (`540`, `549`; `776`, `785`).

### Lifetime

- Starts when card window opens and host registers runtime session.
- VM loaded once via `loadStackBundle(stack.id, sessionId, bundleCode)` (`PluginCardSessionHost.tsx:141`).
- Ends when host unmounts / window closes:
  - `disposeSession(sessionId)` in runtime service (`PluginCardSessionHost.tsx:255-257`, `runtimeService.ts:295-305`)
  - `removeRuntimeSession({sessionId})` in plugin runtime slice (`257`)
  - nav session cleanup in windowing on close (`windowingSlice.ts:92-95`).

### Why this matters

- Opening same stack in two windows gives two sessions (isolated session/card state).
- Duplicate session IDs are rejected by runtime service (`runtimeService.ts:202-204`).

## 5) `emitRuntimeEvent`, Redux, `dispatchRuntimeIntent`, and `ingestRuntimeIntent`

This is the most important mental model.

### End-to-end flow

```text
UI click/input in rendered card
   -> PluginCardRenderer onClick/onChange
   -> emitRuntimeEvent(handler,args) [host]
   -> runtimeService.eventCard(...) [QuickJS VM]
   -> VM handler executes, calling dispatch* helpers
   -> VM returns RuntimeIntent[]
   -> dispatchRuntimeIntent(intent, ctx) [host router]
      -> dispatch(ingestRuntimeIntent(...)) [Redux reducer]
      -> maybe dispatch downstream domain/system action immediately
```

### Who dispatches and who receives

1. VM card handler dispatches logical intents by pushing to `__runtimeIntents` through:
   - `dispatchCardAction`
   - `dispatchSessionAction`
   - `dispatchDomainAction`
   - `dispatchSystemCommand`
   defined in `stack-bootstrap.vm.js:170-201`.

2. Host (`PluginCardSessionHost`) receives returned intents (`332-340`) and calls `dispatchRuntimeIntent`.

3. `dispatchRuntimeIntent` first dispatches Redux `ingestRuntimeIntent` (`pluginIntentRouting.ts:82-88`).

4. `ingestRuntimeIntent` reducer:
   - applies card/session state mutations directly,
   - appends timeline entry,
   - queues domain/system intents when capability allows (`pluginCardRuntimeSlice.ts:244-310`).

5. `dispatchRuntimeIntent` then also performs immediate side routing for domain/system:
   - domain intent -> Redux action `${domain}/${actionType}` (`108-110`)
   - known system commands -> mapped actions (`nav.go`, `nav.back`, `notify`, `window.close`) (`30-67`, `113-117`).

### Are intents Redux actions?

- Runtime intents are **not** native Redux actions by themselves. They are runtime contract objects (`contracts.ts:13-38`).
- They become Redux actions in two ways:
  1. wrapped into `ingestRuntimeIntent(...)` action,
  2. translated into downstream domain/system Redux actions by `dispatchRuntimeIntent`.

### What is `ingestRuntimeIntent` and who owns it?

- It is an action creator/reducer in `pluginCardRuntimeSlice` (`235-320`).
- Owner: engine runtime state slice (`pluginCardRuntime`).
- Responsibility: canonical bookkeeping and state mutation for runtime intents (timeline + local state + queues + capability checks).

## 6) GEPA-14 section 4.1: what are "session intents"?

Session intents are intents with:

```ts
{ scope: 'session', actionType: 'patch' | 'set' | 'reset', payload?: unknown }
```

Defined in `contracts.ts:19-23`.

Semantics in reducer (`pluginCardRuntimeSlice.ts:258-266`):

- `patch`: shallow merge object into `sessionState`.
- `set`: deep set by `{ path, value }`.
- `reset`: clear `sessionState` object.

Use case:

- Cross-card state in one session (for example selected game, mode flags, pagination, thresholds).
- If card A sets session state, card B in same session can read it via `render({ sessionState })`.

Not global:

- Session intents do not change app domain slices directly.
- They are session-local; closing session drops this state.

## 7) GEPA-14 section 4.3: what are pending queues, and should we remove them?

### What they are for

`pluginCardRuntimeSlice` maintains:

- `pendingDomainIntents` (`59`, `275-283`)
- `pendingSystemIntents` (`60`, `304`)
- `pendingNavIntents` (`61`, `305-307`)

They are envelopes of authorized intents waiting for some consumer to dequeue (`322-350`).

So the design intent appears to be queue-based effect execution and replay/debug observability.

### Current reality

In the current host path, `dispatchRuntimeIntent` immediately routes domain/system actions after ingest (`pluginIntentRouting.ts:108-117`).

That means two surfaces coexist:

1. Immediate side-dispatch path (actively used).
2. Pending queues (mostly bookkeeping/debug right now).

### Should we remove them?

Not yet. They are valuable for:

- deterministic event viewer/replay,
- async effect workers,
- failure retry / dead-letter handling,
- decoupling VM intent generation from side-effect timing.

But you must choose one canonical execution model.

### Recommendation

Adopt queue-first execution for GEPA-14 follow-up and GEPA-17 viewer alignment:

1. `dispatchRuntimeIntent` only ingests.
2. Dedicated `RuntimeIntentEffectHost` drains queues and executes domain/system effects.
3. On success/failure, dequeue and append outcome telemetry.

If you do not want queue-first now, then simplify aggressively:

- remove pending queues and related selectors/actions,
- keep timeline only,
- keep immediate routing only.

Mixed mode causes confusion for interns and future contributors.

## 8) ARC backend wiring: concrete runtime bridge

ARC routes are already suitable:

- create session: `POST /api/apps/arc-agi/sessions` (`routes.go:56-76`)
- reset game: `POST /api/apps/arc-agi/sessions/{session}/games/{game}/reset` (`143-151`, `195-208`)
- action: `POST /api/apps/arc-agi/sessions/{session}/games/{game}/actions` (`152-160`, `216-260`)
- events/timeline: `GET .../events`, `GET .../timeline` (`127-140`)

Frontend API mapping already exists in `arcApi.ts:25-68`.

### VM-side command style (recommended)

```js
handlers: {
  start(ctx, args) {
    ctx.dispatchDomainAction('arc', 'command.request', {
      type: 'session.open',
      sourceUrl: args?.sourceUrl,
    });
  },
  move(ctx, args) {
    ctx.dispatchDomainAction('arc', 'command.request', {
      type: 'game.action',
      sessionId: ctx.sessionState.sessionId,
      gameId: ctx.sessionState.gameId,
      action: args?.action,
      data: args?.data ?? {},
    });
  },
}
```

### Host bridge reducer/effect example

```ts
// action emitted by runtime router for domain 'arc'
{ type: 'arc/command.request', payload: {...}, meta: { source: 'plugin-runtime', sessionId, cardId } }

// listener middleware
startListening({
  actionCreator: arcCommandRequested,
  effect: async (action, api) => {
    // call arcApi endpoints, dispatch success/failure + state updates
  },
});
```

### State exposure back to cards

Keep ARC projection in a stable domain namespace:

```ts
globalState.domains.arc = {
  session: { id, status },
  game: { id, state },
  frame: {...},
  actionHistory: [...],
  lastError: null,
}
```

Then card render methods can react to ARC state naturally.

## 9) Event viewer relation (GEPA-17)

Your questions about intents and pending queues overlap with GEPA-17.

Use GEPA-17 design as debugging companion:

- inbound event trace should capture `emitRuntimeEvent` calls,
- outbound trace should capture each intent and routing outcome,
- pending queue status should be visible per session.

If queue-first is adopted, GEPA-17 viewer becomes the operational source of truth for plugin effect lifecycle.

## 10) Implementation plan for intern

Phase 1: correctness and observability

1. Fix rerender trigger for domain updates.
2. Add debug logging hook around `emitRuntimeEvent` and `dispatchRuntimeIntent`.
3. Expose timeline + queue counters in a basic panel.

Phase 2: ARC bridge

1. Add `arc` domain command action contracts (`command.request`, `command.succeeded`, `command.failed`).
2. Implement listener middleware that calls ARC routes via existing API layer.
3. Write projection selector that maps ARC reducer state into `globalState.domains.arc`.

Phase 3: generic platform cleanup

1. Remove inventory defaults from artifact open helper.
2. Split runtime-card prompt policy into generic + app overlays.
3. Move ARC card bundles to arc repo and inventory bundles to inventory repo; keep engine repo runtime-only.

Phase 4: queue model decision

1. Choose queue-first or immediate-only.
2. Delete the alternative path.
3. Update docs/tests/event viewer accordingly.

## Design Decisions

1. Keep app-specific stores and avoid monolithic global store for all apps.
2. Keep engine runtime generic; domain logic stays in app repos.
3. Treat `sessionId` as the cross-layer identity for VM+nav+state.
4. Resolve pending queue ambiguity with one canonical execution model.
5. Prioritize observability (GEPA-17 alignment) before adding more VM complexity.

## Alternatives Considered

### Alternative A: Single global store for all apps

Rejected because:

- high coupling,
- higher accidental rerenders,
- difficult ownership boundaries between app teams.

### Alternative B: Let VM call backend directly via fetch

Rejected because:

- weak capability control,
- testability and security issues,
- bypasses runtime intent and observability pipeline.

### Alternative C: Keep mixed immediate+queue mode indefinitely

Rejected because:

- conceptually hard to teach,
- ambiguous effect source of truth,
- higher bug risk in retry/replay/event viewer paths.

## Open Questions

1. For rerender fix, do we want broad domain snapshot hashing or explicit per-card selectors?
2. Should ARC runtime intents be `scope: 'domain'` only, or should high-risk operations be mapped through `scope: 'system'` commands?
3. Do we want session persistence/restore across app reloads, or are sessions intentionally ephemeral per window lifetime?
4. Which queue model do we ratify for 2026Q1: queue-first or immediate-only?

## References

1. GEPA-14 primary design doc: `design-doc/01-js-vm-programs-hypercard-runtime-and-backend-command-wiring.md`
2. GEPA-14 diary: `reference/01-investigation-diary.md`
3. GEPA-17 event viewer design: `ttmp/2026/02/27/GEPA-17-PLUGIN-EVENT-VIEWER--js-plugin-runtime-event-viewer-for-inbound-ui-events-and-outbound-dispatched-actions/design-doc/01-plugin-runtime-event-viewer-architecture-and-implementation-plan.md`
