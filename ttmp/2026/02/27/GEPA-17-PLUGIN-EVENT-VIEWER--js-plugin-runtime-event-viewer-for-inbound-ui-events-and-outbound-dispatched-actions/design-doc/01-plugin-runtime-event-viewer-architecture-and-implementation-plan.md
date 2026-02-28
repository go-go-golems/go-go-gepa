---
Title: Plugin Runtime Event Viewer Architecture and Implementation Plan
Ticket: GEPA-17-PLUGIN-EVENT-VIEWER
Status: active
Topics:
    - js-vm
    - event-streaming
    - go-go-os
    - hypercard
    - inventory-app
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/27/GEPA-14-VM-JS-PROGRAMS--js-vm-programs-integration-for-hypercard-in-inventory-go-go-os/design-doc/01-js-vm-programs-hypercard-runtime-and-backend-command-wiring.md
      Note: Prior ticket context consulted for terminology and runtime framing
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/apps/arc-agi-player/src/launcher/module.tsx
      Note: Real-world adapter pattern example for app-window integration
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/cards/runtime.ts
      Note: RuntimeDebugEvent contract and emit helper evidence
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/chat/debug/EventViewerWindow.tsx
      Note: Existing event stream viewer interaction-pattern evidence
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/chat/debug/eventBus.ts
      Note: Bounded event stream bus pattern evidence
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/PluginCardRenderer.tsx
      Note: Inbound UI handler emission boundary evidence
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx
      Note: Runtime eventCard invocation and intent dispatch orchestration evidence
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/desktopContributions.ts
      Note: Desktop command/adapter integration extension points
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/pluginIntentRouting.ts
      Note: Intent ingest and downstream forwarding behavior evidence
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/debug/debugSlice.ts
      Note: Debug event storage/filter capacity evidence
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/debug/useStandardDebugHooks.ts
      Note: Sanitization and ingest hook evidence for reusable debug wiring
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/features/pluginCardRuntime/pluginCardRuntimeSlice.ts
      Note: Timeline and pending queue semantics evidence
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/features/pluginCardRuntime/selectors.ts
      Note: Runtime timeline and pending-intent selector evidence
ExternalSources: []
Summary: In-depth architecture research and implementation blueprint for adding a JS-plugin runtime event viewer that shows inbound UI events and outbound dispatched actions.
LastUpdated: 2026-02-27T23:08:12-05:00
WhatFor: Provide intern-friendly, evidence-backed guidance for implementing and validating plugin runtime event observability.
WhenToUse: Use when implementing GEPA-17 or diagnosing plugin runtime behavior across UI events, VM intents, and Redux dispatch routing.
---


# Plugin Runtime Event Viewer Architecture and Implementation Plan

## Executive summary

`GEPA-17-PLUGIN-EVENT-VIEWER` asks for a dedicated event viewer for running JS plugins (HyperCard-style card sessions) that shows:

1. events going in (UI interaction -> VM handler invocation), and
2. actions going out (runtime intents that are dispatched/routed into Redux/system actions).

The codebase already has most primitives needed for this, but they are split across three systems:

1. plugin runtime host/routing (`PluginCardSessionHost`, `pluginIntentRouting`, `pluginCardRuntimeSlice`),
2. generic debug event infrastructure (`RuntimeDebugEvent`, `debugSlice`, `RuntimeDebugPane`, `useStandardDebugHooks`),
3. chat-specific event streaming UI (`eventBus` + `EventViewerWindow`).

The key finding is that outgoing intents are partially observable today (via plugin runtime timeline), but inbound UI events are not recorded in a first-class stream. Also, the existing debug infrastructure exists but is not wired into the plugin runtime host path.

Recommended direction:

1. instrument plugin runtime flow with structured `RuntimeDebugEvent` emissions,
2. introduce correlation IDs from inbound interaction through outbound intent routing,
3. add a plugin-focused viewer UI using the existing debug slice (not a new storage system),
4. expose it via desktop contribution window command(s) and/or runtime debug window integration.

This provides low-risk observability with minimal architectural churn and reuses existing state, sanitization, and pane interaction patterns.

---

## Problem statement and scope

### Problem

Developers currently cannot answer simple runtime debugging questions quickly for a running plugin session:

1. Which UI handler fired?
2. What args were passed into the VM event handler?
3. Which intents did the VM return?
4. Which intents were blocked vs forwarded?
5. Which concrete Redux/system actions were finally dispatched?

Without a unified stream, troubleshooting plugin behavior requires jumping between runtime code, Redux state snapshots, and ad-hoc logging.

### Scope

In scope:

1. JS plugin runtime event visibility from `PluginCardRenderer` interaction to routing dispatch.
2. Viewer focused on plugin sessions/cards (inventory/go-go-os runtime cards included).
3. Intern-onboarding documentation and implementation plan.

Out of scope:

1. Redesigning the plugin intent contract (`RuntimeIntent`) itself.
2. Replacing chat SEM event viewer architecture.
3. Full production telemetry pipeline beyond local dev/debug observability.

---

## Current-state architecture (evidence-backed)

## 1) Where inbound UI events originate

UI node events in plugin-rendered trees are represented by `UIEventRef` (`handler`, optional `args`) in runtime UI types (`packages/engine/src/plugin-runtime/uiTypes.ts:1-44`).

`PluginCardRenderer` maps node interactions to the host callback:

1. `eventHandler(ref, onEvent, payload)` merges static args and dynamic payload (`PluginCardRenderer.tsx:25-36`),
2. button/input/counter nodes call `eventHandler(...)` (`PluginCardRenderer.tsx:85-114`),
3. renderer emits `onEvent(handler, args)` to host (`PluginCardRenderer.tsx:7, 38, 148`).

This is the earliest reliable point for "event going in".

## 2) How VM handler execution happens

`PluginCardSessionHost` owns session runtime orchestration.

Relevant flow:

1. projects global/session/card context (`projectGlobalState`, `PluginCardSessionHost.tsx:32-66`),
2. receives UI callback in `emitRuntimeEvent` (`PluginCardSessionHost.tsx:309-354`),
3. calls `runtimeService.eventCard(...)` (`PluginCardSessionHost.tsx:318-326`),
4. iterates returned intents and calls `dispatchRuntimeIntent(intent, context)` (`PluginCardSessionHost.tsx:332-340`).

`QuickJSCardRuntimeService.eventCard(...)` evaluates VM event handler and validates result as `RuntimeIntent[]` (`packages/engine/src/plugin-runtime/runtimeService.ts:273-293`).

Contract for intents is in `contracts.ts` (`scope` variants: card/session/domain/system, `contracts.ts:13-39`).

## 3) How intents are ingested and forwarded

`dispatchRuntimeIntent` in `pluginIntentRouting.ts` does two things:

1. always dispatches `ingestRuntimeIntent(...)` first (`pluginIntentRouting.ts:81-88`),
2. then conditionally forwards domain/system intents to downstream actions (`pluginIntentRouting.ts:94-118`).

Routing details:

1. domain -> synthetic Redux action type `${domain}/${actionType}` with `meta.source='plugin-runtime'` (`pluginIntentRouting.ts:69-79`),
2. system -> mapped actions (`nav.back`, `nav.go`, `notify`, `window.close`) via `toSystemAction` (`pluginIntentRouting.ts:30-67`),
3. capability checks happen before forwarding (`pluginIntentRouting.ts:94-106`).

## 4) What plugin runtime state already tracks

`pluginCardRuntimeSlice` tracks per-session state and runtime timeline:

1. `sessions`, `timeline`, `pendingDomainIntents`, `pendingSystemIntents`, `pendingNavIntents` (`pluginCardRuntimeSlice.ts:56-62`),
2. timeline entry structure includes scope/actionType/domain/command/outcome/reason (`pluginCardRuntimeSlice.ts:14-26`),
3. `ingestRuntimeIntent` applies local state actions or queues domain/system intents, then appends timeline (`pluginCardRuntimeSlice.ts:235-320`),
4. timeline is capped (`MAX_TIMELINE_ENTRIES=300`, `pluginCardRuntimeSlice.ts:95, 163-197`).

Selectors expose timeline and pending queues (`selectors.ts:26-35`).

Important: timeline captures intent ingestion outcomes, but not the initial inbound UI interaction event itself.

## 5) Existing debug infrastructure that is reusable

The engine already has generic debug event plumbing:

1. `RuntimeDebugEvent` + hooks + `emitRuntimeDebugEvent(...)` (`cards/runtime.ts:1-71`),
2. `debugSlice` with event capacity/filtering/selection (`debugSlice.ts:4-98`),
3. `useStandardDebugHooks()` that dispatches `ingestEvent` and sanitizes payload (`useStandardDebugHooks.ts:45-58`),
4. `StandardDebugPane` + `RuntimeDebugPane` UI (`StandardDebugPane.tsx:29-70`, `RuntimeDebugPane.tsx:15-182`).

`sanitizeDebugValue` already redacts sensitive key patterns and truncates large payloads (`useStandardDebugHooks.ts:6-43`).

### Gap in current wiring

Although helper APIs exist, the plugin host path (`PluginCardSessionHost` + `pluginIntentRouting`) currently does not call `emitRuntimeDebugEvent(...)` or use `useStandardDebugHooks()`.

## 6) Existing event viewer implementation patterns (chat)

`EventViewerWindow` + `eventBus` form a mature event-stream viewer pattern in chat:

1. event bus retains bounded history per conversation (`eventBus.ts:15-18, 82-87`),
2. viewer supports filtering, pause/resume, auto-scroll, and YAML export (`EventViewerWindow.tsx:114-440`),
3. tests cover thresholds/filtering/export/history caps (`EventViewerWindow.test.ts:10-163`, `eventBus.test.ts:95-104`).

These patterns can be reused for plugin viewer UX, even if data source remains Redux debug slice.

## 7) Windowing and integration extension points

Desktop shell supports composable window content adapters and command handlers through contributions:

1. `DesktopContribution.windowContentAdapters/commands/startupWindows` (`desktopContributions.ts:36-43`),
2. composed adapters are merged with defaults (`useDesktopShellController.tsx:1152-1159`),
3. window body rendering runs adapter pipeline (`useDesktopShellController.tsx:1162-1174`),
4. app modules already use this pattern (`apps/arc-agi-player/src/launcher/module.tsx:35-67, 92-97`).

This is the preferred integration mechanism for opening a dedicated viewer window.

---

## Gap analysis

## Gap 1: no first-class inbound event stream

Inbound UI event trigger information (`handler`, merged args, source node path) is not persisted in timeline or debug slice today.

Impact:

1. cannot correlate user interaction to VM intents deterministically,
2. difficult to debug "why handler X ran" or "wrong payload passed" issues.

## Gap 2: outbound forwarding details are fragmented

Timeline records ingest outcomes, but viewer-level dispatch details are incomplete:

1. domain/system forwarding decision and final dispatched action shape are not emitted as dedicated debug events,
2. system intent mapping behavior (`toSystemAction`) is hidden unless stepping through code.

## Gap 3: no correlation IDs across interaction lifecycle

No explicit `interactionId` links:

1. inbound UI event,
2. VM `eventCard` call,
3. each returned intent,
4. forwarding/block decision and final dispatch.

Without correlation, high-volume sessions are hard to inspect.

## Gap 4: potential routing inconsistency for missing session

In `dispatchRuntimeIntent`, forwarding authorization checks are gated by `runtimeSession?.capabilities` (`pluginIntentRouting.ts:94-106`).

If runtime session lookup fails, domain/system forwarding paths can still execute (`pluginIntentRouting.ts:108-117`) even though `ingestRuntimeIntent` would mark timeline as denied for missing session (`pluginCardRuntimeSlice.ts:238-241`).

This mismatch is both a correctness risk and an observability risk; viewer instrumentation should make this visible and implementation should likely harden this behavior.

## Gap 5: existing pending queues appear unconsumed

`pendingDomainIntents`, `pendingSystemIntents`, and `pendingNavIntents` are populated (`pluginCardRuntimeSlice.ts:275-307`) and selectors exist (`selectors.ts:28-35`), but no runtime consumption path is present in current engine usage search.

This reinforces that immediate observability should target the actual forwarding path (`dispatchRuntimeIntent`) rather than pending queue consumption.

## Gap 6: debug slice exists by default but viewer isn’t mounted

`createAppStore` includes `debug` and `pluginCardRuntime` reducers by default (`createAppStore.ts:55-60`), but a plugin-event-focused viewer surface is not mounted in app runtime by default.

---

## Design options

## Option A: extend `pluginCardRuntime.timeline` only

Approach:

1. add inbound event entries to plugin runtime timeline,
2. enrich timeline entries with forwarding details,
3. build viewer directly on timeline selector.

Pros:

1. single plugin-runtime state source,
2. deterministic lifecycle inside one slice.

Cons:

1. pollutes runtime domain slice with debug concerns,
2. requires broader schema changes and migration of existing timeline semantics,
3. lower reuse of existing debug UI/filtering/sanitization.

Verdict: possible, but not preferred for a debug-focused feature.

## Option B (recommended): use `RuntimeDebugEvent` + `debugSlice` as event ledger

Approach:

1. instrument plugin runtime host/routing to emit structured debug events,
2. correlate via interaction/intent identifiers in `meta`,
3. build plugin event viewer on debug slice selectors (or derived selectors),
4. optionally reuse interaction UX from chat event viewer.

Pros:

1. leverages existing generic debug infrastructure,
2. minimal impact on runtime domain state semantics,
3. existing sanitization + filtering + capacity behavior,
4. easiest phased rollout.

Cons:

1. debug slice default capacity (300) may be low for long sessions,
2. requires explicit filtering by `kind` prefixes to isolate plugin events.

Verdict: best balance of speed, risk, and architecture fit.

## Option C: separate plugin event bus + dedicated viewer storage

Approach:

1. replicate chat-style bus/history for plugin sessions,
2. build dedicated viewer component and state.

Pros:

1. fully custom behavior and retention,
2. can be isolated from other debug traffic.

Cons:

1. duplicates existing debug infrastructure,
2. more maintenance and divergent tooling.

Verdict: overkill for current requirements.

---

## Recommended architecture

## 1) Event model

Introduce a plugin runtime trace taxonomy using `RuntimeDebugEvent.kind` values:

1. `plugin.ui.event.in`
2. `plugin.vm.event.start`
3. `plugin.vm.event.result`
4. `plugin.vm.event.error`
5. `plugin.intent.route.start`
6. `plugin.intent.route.forwarded`
7. `plugin.intent.route.denied`
8. `plugin.intent.route.ignored`

Use `meta` to carry correlation and routing context:

```ts
interface PluginRuntimeTraceMeta {
  sessionId: string;
  windowId: string;
  stackId: string;
  cardId: string;
  interactionId: string;     // one per inbound UI event
  intentId?: string;         // one per returned intent
  intentIndex?: number;      // index in returned intent array
  handler?: string;
  forwardedActionType?: string;
  deniedReason?: string;
}
```

## 2) Correlation strategy

1. Generate `interactionId` in `emitRuntimeEvent` before `eventCard`.
2. For each intent from VM result, generate `intentId = `${interactionId}:${index}``.
3. Pass trace metadata into `dispatchRuntimeIntent` so router instrumentation can emit route outcome events.
4. Optionally allow `ingestRuntimeIntent` caller-provided IDs for exact timeline/debug join.

## 3) Instrumentation points

### In `PluginCardSessionHost`

At minimum:

1. emit `plugin.ui.event.in` when `emitRuntimeEvent` receives `handler/args`.
2. emit `plugin.vm.event.start` before `runtimeService.eventCard`.
3. emit `plugin.vm.event.result` with duration and intent count after success.
4. emit `plugin.vm.event.error` on exception path.
5. emit `plugin.intent.route.start` before each `dispatchRuntimeIntent` call.

### In `pluginIntentRouting`

Emit route outcome events after authorization/mapping decisions:

1. `...forwarded` for domain/system actions actually dispatched,
2. `...denied` when capability/session denies,
3. `...ignored` when system payload cannot map to concrete action.

Also harden missing-session behavior so forwarding does not occur when session is absent.

## 4) Optional API hardening

Change `dispatchRuntimeIntent` signature to return structured result:

```ts
interface DispatchRuntimeIntentResult {
  intentScope: RuntimeIntent['scope'];
  forwarded: boolean;
  forwardedActionType?: string;
  outcome: 'forwarded' | 'denied' | 'ignored';
  reason?: string;
}
```

This makes instrumentation and tests deterministic and avoids inference from side effects.

## 5) Viewer composition

Build `PluginRuntimeEventViewerWindow` as plugin-focused panel:

1. consumes debug events filtered by `kind.startsWith('plugin.')`,
2. filter by `sessionId`, `cardId`, direction (`inbound` vs `outbound`), outcome,
3. grouped view by `interactionId` (expand/collapse),
4. details panel with full payload and forwarded action snapshot,
5. export visible events to YAML (reuse chat viewer pattern).

## 6) Integration surface

Two practical launch surfaces:

1. embed in `RuntimeCardDebugWindow` as a new section/tab (lowest friction),
2. open dedicated `app` window via contribution command + content adapter (`desktopContributions.ts` + app launcher modules pattern).

For intern onboarding and debugging live sessions, dedicated window is preferable; for rapid v1 delivery, embedding in existing runtime debug window is simplest.

---

## Pseudocode and key flows

### A) Host instrumentation flow

```ts
function emitRuntimeEvent(handler: string, args?: unknown) {
  const interactionId = nanoid();
  const t0 = performance.now();

  emitDebug('plugin.ui.event.in', {
    meta: { interactionId, sessionId, cardId, windowId, handler },
    payload: { args },
  });

  try {
    emitDebug('plugin.vm.event.start', {
      meta: { interactionId, sessionId, cardId, windowId, handler },
    });

    const intents = runtimeService.eventCard(...);

    emitDebug('plugin.vm.event.result', {
      durationMs: performance.now() - t0,
      meta: { interactionId, sessionId, cardId, windowId, handler },
      payload: { intentCount: intents.length, intents },
    });

    intents.forEach((intent, index) => {
      const intentId = `${interactionId}:${index}`;
      emitDebug('plugin.intent.route.start', {
        meta: { interactionId, intentId, intentIndex: index, sessionId, cardId, windowId },
        payload: { intent },
      });

      dispatchRuntimeIntent(intent, {
        dispatch,
        getState,
        sessionId,
        cardId,
        windowId,
        trace: { interactionId, intentId, intentIndex: index, emitDebug },
      });
    });
  } catch (err) {
    emitDebug('plugin.vm.event.error', {
      durationMs: performance.now() - t0,
      meta: { interactionId, sessionId, cardId, windowId, handler },
      payload: { message: String(err) },
    });
  }
}
```

### B) Routing outcome instrumentation flow

```ts
function dispatchRuntimeIntent(intent, context): DispatchRuntimeIntentResult {
  context.dispatch(ingestRuntimeIntent(...));

  const session = getRuntimeSession(context);
  if (!session) {
    context.trace?.emitDebug('plugin.intent.route.denied', {
      meta: { ...context.trace, deniedReason: `missing_session:${context.sessionId}` },
      payload: { intent },
    });
    return { forwarded: false, outcome: 'denied', reason: `missing_session:${context.sessionId}` };
  }

  if (intent.scope === 'domain') {
    const decision = authorizeDomainIntent(...);
    if (!decision.allowed) return denied(...);

    const action = toDomainAction(intent, context);
    context.dispatch(action);
    emitForwarded(action.type);
    return { forwarded: true, outcome: 'forwarded', forwardedActionType: action.type };
  }

  if (intent.scope === 'system') {
    const decision = authorizeSystemIntent(...);
    if (!decision.allowed) return denied(...);

    const action = toSystemAction(intent, context);
    if (!action) return ignored('invalid_system_payload');

    context.dispatch(action);
    emitForwarded(action.type);
    return { forwarded: true, outcome: 'forwarded', forwardedActionType: action.type };
  }

  return { forwarded: false, outcome: 'ignored', reason: 'local_state_intent' };
}
```

### C) Viewer selector sketch

```ts
const selectPluginTraceEvents = createSelector([selectDebugState], (debug) =>
  debug.events.filter((e) => e.kind.startsWith('plugin.')),
);

const selectPluginTraceBySession = (sessionId: string) =>
  createSelector([selectPluginTraceEvents], (events) =>
    events.filter((e) => e.meta?.sessionId === sessionId),
  );

const selectGroupedInteractions = createSelector([selectPluginTraceEvents], (events) => {
  const groups = new Map<string, RuntimeDebugEvent[]>();
  for (const event of events) {
    const key = String(event.meta?.interactionId ?? 'uncorrelated');
    const list = groups.get(key) ?? [];
    list.push(event);
    groups.set(key, list);
  }
  return Array.from(groups.entries());
});
```

---

## Implementation plan (phased, file-level)

## Phase 1: event schema + trace plumbing

Primary goals:

1. finalize `kind` taxonomy and `meta` shape,
2. define trace context type for routing.

Files:

1. `packages/engine/src/cards/runtime.ts`
2. `packages/engine/src/components/shell/windowing/pluginIntentRouting.ts`
3. optional helper file: `packages/engine/src/components/shell/windowing/pluginRuntimeTrace.ts`

Deliverables:

1. typed helper(s) for plugin trace emission,
2. optional `DispatchRuntimeIntentResult` return contract.

## Phase 2: instrumentation in host and router

Primary goals:

1. emit inbound UI + VM lifecycle events in `PluginCardSessionHost`,
2. emit routing outcome events in `pluginIntentRouting`,
3. add missing-session routing guard.

Files:

1. `packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx`
2. `packages/engine/src/components/shell/windowing/pluginIntentRouting.ts`

Deliverables:

1. correlated event stream (`interactionId` and `intentId`),
2. deterministic route outcome logging.

## Phase 3: viewer UI

Primary goals:

1. build plugin-specific viewer component over debug slice,
2. support filters, grouping, and event details.

Files (proposed):

1. `packages/engine/src/hypercard/debug/PluginRuntimeEventViewerWindow.tsx`
2. `packages/engine/src/hypercard/debug/index.ts` (exports)
3. optional stories/tests in `components/widgets` and `hypercard/debug`.

Deliverables:

1. usable plugin event stream UI,
2. storybook scenarios with synthetic traces.

## Phase 4: window/command integration

Primary goals:

1. enable opening viewer for active session(s),
2. support command/context-menu pathway.

Integration choices:

1. v1: embed in `RuntimeCardDebugWindow`.
2. v2: dedicated `appKey` window via contributions.

Likely files:

1. `packages/engine/src/hypercard/debug/RuntimeCardDebugWindow.tsx`
2. app launcher module contributions (example pattern in `apps/arc-agi-player/src/launcher/module.tsx`).

## Phase 5: docs and runbook

Primary goals:

1. add concise developer usage docs,
2. include intern-oriented "how to trace a plugin event" flow.

Likely files:

1. `go-go-os/docs/js-api-user-guide-reference.md`
2. new frontend debug playbook entry.

---

## Testing and validation strategy

## Unit tests (core)

1. `pluginIntentRouting`:
   - missing session denies forwarding,
   - capability-denied emits denied trace,
   - domain/system forwarded emits forwarded trace with action type.
2. trace helper:
   - sanitization/redaction preserved,
   - event IDs/timestamps assigned.
3. selectors:
   - grouping by `interactionId`,
   - session/card filters.

## Existing test extension targets

1. `packages/engine/src/__tests__/plugin-intent-routing.test.ts` (routing behavior baseline exists).
2. `packages/engine/src/__tests__/plugin-card-runtime.test.ts` (timeline semantics baseline exists).
3. `packages/engine/src/chat/debug/EventViewerWindow.test.ts` and `eventBus.test.ts` for UX pattern parity ideas.

## UI/story validation

1. Story with one interaction -> multiple intents -> mixed denied/forwarded outcomes.
2. Story with high-volume stream to verify filter/pause/autoscroll behavior.

## Manual runbook

1. launch app with plugin-enabled card stack,
2. trigger click/input/counter events,
3. verify viewer shows `plugin.ui.event.in` before `plugin.vm.event.result`,
4. verify each intent has matching routing outcome event,
5. intentionally deny capability and verify denied entries,
6. export filtered stream and inspect YAML.

---

## Risks, tradeoffs, and mitigations

## Risk 1: event volume and performance

Issue:

1. debug slice default capacity is 300 events (`debugSlice.ts:4,17-20`),
2. high-frequency plugin interactions may evict recent context quickly.

Mitigation:

1. add configurable capacity action or plugin-viewer-local retention,
2. add filters by session/card/kind to reduce render load.

## Risk 2: payload sensitivity and size

Issue:

1. plugin args/intents may contain secrets or large blobs.

Mitigation:

1. always route payloads through `sanitizeDebugValue` (`useStandardDebugHooks.ts:15-43`),
2. truncate large arrays/strings, redact sensitive keys.

## Risk 3: routing behavior drift

Issue:

1. duplication between slice authorization and router forwarding checks can drift.

Mitigation:

1. add explicit routing result contract and tests,
2. add missing-session hard guard.

## Risk 4: UI placement churn

Issue:

1. deciding between embedded debug panel and dedicated window may cause rework.

Mitigation:

1. build viewer component standalone first,
2. wire into embedded surface in v1, dedicated window in v2.

---

## Alternatives considered and why not chosen

1. Expand plugin timeline only: rejected due coupling debug concerns to runtime domain state.
2. New plugin-only event bus: rejected due duplication of existing debug infra.
3. Rely only on console logs: rejected due poor discoverability, no filtering/correlation/export.

---

## Open questions

1. Should plugin event viewer be always-on in dev builds, or explicitly opened via command/context action?
2. Should `RuntimeDebugEvent` gain first-class `sessionId/windowId` fields, or keep these in `meta` for backward compatibility?
3. Do we want exact timeline intent ID alignment (requires caller-provided ingest IDs), or is debug-stream-only correlation sufficient for v1?
4. Should system actions forwarded from plugin include standardized meta (`source: plugin-runtime`) similar to domain actions?
5. Is a dedicated export format needed for plugin traces, or is YAML parity with chat event viewer sufficient?

---

## Recommended v1/v2 rollout

## V1 (fastest path)

1. instrument host + router with debug events,
2. add plugin viewer component filtered from debug slice,
3. mount inside existing `RuntimeCardDebugWindow`.

Expected result: immediate visibility for incoming UI events and outgoing routed actions without windowing-level changes.

## V2 (desktop-native tooling)

1. add desktop command/context action to open viewer window for current session,
2. render via contribution adapter (`content.kind='app'`, `appKey` prefix strategy),
3. optionally add per-session persistence and richer export.

Expected result: a first-class runtime inspector window usable in ARC/inventory shells.

---

## References

Primary runtime/event routing evidence:

1. `packages/engine/src/components/shell/windowing/PluginCardRenderer.tsx` (event emit path): lines 7, 25-36, 85-114, 148
2. `packages/engine/src/plugin-runtime/uiTypes.ts` (UI event refs): lines 1-44
3. `packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx` (eventCard + dispatch flow): lines 309-340
4. `packages/engine/src/plugin-runtime/runtimeService.ts` (VM `eventCard`): lines 273-293
5. `packages/engine/src/components/shell/windowing/pluginIntentRouting.ts` (ingest + route): lines 81-118
6. `packages/engine/src/features/pluginCardRuntime/pluginCardRuntimeSlice.ts` (timeline, pending queues, ingest outcomes): lines 14-26, 56-62, 235-320
7. `packages/engine/src/features/pluginCardRuntime/selectors.ts` (timeline/pending selectors): lines 26-35
8. `packages/engine/src/features/pluginCardRuntime/capabilityPolicy.ts` (authorization): lines 41-54

Debug infrastructure evidence:

1. `packages/engine/src/cards/runtime.ts` (`RuntimeDebugEvent`, `emitRuntimeDebugEvent`): lines 1-71
2. `packages/engine/src/debug/useStandardDebugHooks.ts` (sanitize + dispatch): lines 15-58
3. `packages/engine/src/debug/debugSlice.ts` (capacity/filter/select): lines 4-98
4. `packages/engine/src/debug/StandardDebugPane.tsx` (wiring): lines 29-70
5. `packages/engine/src/components/shell/RuntimeDebugPane.tsx` (UI): lines 15-182

Related viewer pattern evidence:

1. `packages/engine/src/chat/debug/eventBus.ts` (bounded history + stream): lines 15-18, 63-95, 97-117
2. `packages/engine/src/chat/debug/EventViewerWindow.tsx` (viewer UX): lines 114-440
3. `packages/engine/src/chat/debug/EventViewerWindow.test.ts` (filter/export behavior): lines 62-163
4. `packages/engine/src/chat/debug/eventBus.test.ts` (history cap): lines 95-104

Window integration evidence:

1. `packages/engine/src/components/shell/windowing/desktopContributions.ts` (contribution extension points): lines 36-43, 118-130
2. `packages/engine/src/components/shell/windowing/useDesktopShellController.tsx` (adapter composition/render): lines 1152-1174
3. `apps/arc-agi-player/src/launcher/module.tsx` (real adapter pattern): lines 35-67, 92-97
4. `packages/engine/src/app/createAppStore.ts` (default reducers include debug/plugin runtime): lines 55-60

Prior ticket context:

1. `ttmp/2026/02/27/GEPA-14-VM-JS-PROGRAMS--js-vm-programs-integration-for-hypercard-in-inventory-go-go-os/design-doc/01-js-vm-programs-hypercard-runtime-and-backend-command-wiring.md`
