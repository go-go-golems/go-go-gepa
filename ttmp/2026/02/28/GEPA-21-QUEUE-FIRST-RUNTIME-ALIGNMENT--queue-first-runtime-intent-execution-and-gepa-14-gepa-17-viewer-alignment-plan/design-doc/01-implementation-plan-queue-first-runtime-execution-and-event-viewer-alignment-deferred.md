---
Title: 'Implementation plan: queue-first runtime execution and event viewer alignment (deferred)'
Ticket: GEPA-21-QUEUE-FIRST-RUNTIME-ALIGNMENT
Status: active
Topics:
    - go-go-os
    - hypercard
    - event-streaming
    - js-vm
    - inventory-app
    - architecture
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/27/GEPA-14-VM-JS-PROGRAMS--js-vm-programs-integration-for-hypercard-in-inventory-go-go-os/design-doc/02-intern-q-a-store-boundaries-sessions-runtime-intents-and-re-rendering.md
      Note: Prior recommendation to adopt queue-first and align with GEPA-17
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/27/GEPA-17-PLUGIN-EVENT-VIEWER--js-plugin-runtime-event-viewer-for-inbound-ui-events-and-outbound-dispatched-actions/design-doc/01-plugin-runtime-event-viewer-architecture-and-implementation-plan.md
      Note: Existing event viewer lifecycle/telemetry plan to align with queue-first execution
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/chat/debug/EventViewerWindow.tsx
      Note: Viewer surface that needs lifecycle alignment with queue-first model
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx
      Note: Entrypoint for UI event emission and dispatchRuntimeIntent invocation
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/pluginIntentRouting.ts
      Note: Current mixed-mode routing point where ingest and immediate dispatch coexist
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/features/pluginCardRuntime/pluginCardRuntimeSlice.ts
      Note: State contract for timeline and pending runtime queues
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/features/pluginCardRuntime/selectors.ts
      Note: Queue selectors that a future effect host will consume
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-28T00:28:00.597777003-05:00
WhatFor: Give an intern a complete, implementation-grade plan for migrating to queue-first runtime intent execution and aligning GEPA-17 viewer telemetry semantics, while explicitly deferring execution for now.
WhenToUse: Use when the team decides to begin queue-first adoption work after GEPA-20 stabilization.
---


# Implementation plan: queue-first runtime execution and event viewer alignment (deferred)

## Executive Summary

This ticket plans a future migration to a single runtime intent model: queue-first execution.

Today, runtime intents follow a mixed model:

1. `dispatchRuntimeIntent` immediately routes domain/system intents.
2. `ingestRuntimeIntent` also appends domain/system intents into pending queues.

This duality creates ambiguity for correctness, replay, and observability. GEPA-17 event viewer work also depends on having one canonical lifecycle source.

Decision for now:

1. Do not implement this migration today.
2. Keep current mixed behavior.
3. Keep pending queues but ignore them operationally.

This document is a deferred execution plan so an intern can pick up the work with low ambiguity later.

## Problem Statement

GEPA-14 and GEPA-17 uncovered the same architecture tension:

1. runtime slice has pending queues (`pendingDomainIntents`, `pendingSystemIntents`, `pendingNavIntents`),
2. routing layer still executes immediate side effects directly after ingest,
3. queue dequeue/consumer path is not the primary runtime execution path.

Consequences:

1. two conceptual execution models coexist,
2. event viewer has to explain two parallel truths (timeline+queue vs immediate route),
3. harder reasoning about idempotency, retries, replay, and diagnostics.

The problem is not broken functionality right now; it is long-term maintainability and observability drift.

## Proposed Solution

### Target model (future): queue-first runtime effects

Adopt one canonical flow:

1. `dispatchRuntimeIntent` performs validation + ingest only.
2. A dedicated `RuntimeIntentEffectHost` drains queues and executes side effects.
3. Success/failure outcomes are recorded and queues are dequeued with explicit lifecycle events.
4. GEPA-17 viewer reflects this same lifecycle end-to-end.

### Current-state policy (now)

Until the team starts this ticket:

1. keep mixed model as-is,
2. do not remove queues,
3. do not wire queue drains in production path,
4. focus only on planning and onboarding docs in this ticket.

### Lifecycle sketch

```text
UI Event
  -> emitRuntimeEvent(handler,args)
  -> runtimeService.eventCard(...)
  -> intents[]
  -> dispatchRuntimeIntent(intent)
       -> ingestRuntimeIntent (timeline + queue write)
       -> (future) stop here
  -> RuntimeIntentEffectHost (future)
       -> read queue head
       -> execute effect
       -> emit route result telemetry
       -> dequeue
```

### Viewer alignment target

GEPA-17 event viewer should be able to show, per intent:

1. ingested,
2. queued,
3. dequeued/executing,
4. succeeded/failed/denied,
5. final side effect emitted (domain action/system command).

## Design Decisions

1. Keep this ticket deferred right now.
Reason: active stabilization work is in flight and this is a structural refactor.

2. Keep queues in state during defer period.
Reason: preserves diagnostics compatibility and avoids churn before migration.

3. Plan queue-first as the eventual single source of execution truth.
Reason: better replay semantics and cleaner GEPA-17 telemetry model.

4. Separate migration phases into contract, effect host, routing cutover, and viewer alignment.
Reason: intern-friendly sequencing with clear rollback points.

## Alternatives Considered

1. Immediate-only forever (delete queues).
Rejected for target state because it weakens replay and queue-state observability capabilities.

2. Keep mixed model indefinitely.
Rejected for target state because it preserves ambiguity and complexity.

3. Queue-first migration in one large PR.
Rejected because it has high blast radius and difficult debugging surface.

## Implementation Plan

### Phase 0: Preconditions and guardrails

1. Confirm no concurrent refactor touching:
   - `pluginIntentRouting.ts`,
   - `pluginCardRuntimeSlice.ts`,
   - `PluginCardSessionHost.tsx`.
2. Freeze telemetry field names used by GEPA-17 viewer.
3. Add ADR note: queue-first is planned canonical path.

### Phase 1: Contracts and telemetry envelope

1. Introduce explicit runtime intent lifecycle event schema.
2. Add `intentId`/`interactionId` propagation contract from emit -> ingest -> execute.
3. Add structured route result type for effect execution.

### Phase 2: RuntimeIntentEffectHost skeleton

1. Create effect host component/middleware layer (exact placement to decide at implementation start).
2. Read queue selectors (`domain/system/nav`) and process deterministic order.
3. Add in-flight dedupe and at-least-once safeguards.

### Phase 3: Queue drain execution semantics

1. Domain queue:
   - construct domain actions from envelopes,
   - dispatch and record outcome,
   - dequeue on completion.
2. System queue:
   - map to nav/notify/window commands,
   - dispatch and record outcome,
   - dequeue on completion.
3. Nav queue:
   - keep as derived/diagnostic queue or formalize separate consumer policy.

### Phase 4: Routing cutover

1. Update `dispatchRuntimeIntent`:
   - keep ingest,
   - remove immediate domain/system direct dispatch.
2. Ensure capability checks remain consistent (ingest path vs effect host path).
3. Add feature flag for staged rollout in dev/test environments.

### Phase 5: GEPA-17 viewer alignment

1. Emit lifecycle bus entries for queue state transitions.
2. Update viewer filters/tables to show queue status per intent.
3. Add correlation view by `interactionId`:
   - inbound UI event,
   - produced intents,
   - queued/dequeued,
   - final action result.

### Phase 6: Tests

1. Unit tests:
   - queue ordering,
   - idempotency/retry behavior,
   - capability deny behavior.
2. Integration tests:
   - event -> queue -> effect -> dequeue.
3. Viewer tests:
   - correlation renders correctly.

### Phase 7: Cleanup

1. Remove obsolete immediate-route assumptions from docs/tests.
2. Update GEPA-14 and GEPA-17 cross references.
3. Add operator runbook for debugging stalled queues.

### Phase 8: Rollout

1. Enable for internal profile first.
2. Collect telemetry and verify no intent drops.
3. Promote to default once stability criteria pass.

## Detailed Engineering Notes For Intern

### Existing code points to understand first

1. `packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx`
2. `packages/engine/src/components/shell/windowing/pluginIntentRouting.ts`
3. `packages/engine/src/features/pluginCardRuntime/pluginCardRuntimeSlice.ts`
4. `packages/engine/src/features/pluginCardRuntime/selectors.ts`
5. `packages/engine/src/chat/debug/EventViewerWindow.tsx`

### Current mixed behavior summary

```ts
// today
dispatchRuntimeIntent(intent) {
  dispatch(ingestRuntimeIntent(intent)); // timeline + queue write
  if (intent.scope === "domain") dispatch(domainAction); // immediate route
  if (intent.scope === "system") dispatch(systemAction); // immediate route
}
```

### Target behavior summary

```ts
// future
dispatchRuntimeIntent(intent) {
  dispatch(ingestRuntimeIntent(intent)); // no direct side effect route
}

RuntimeIntentEffectHost.tick() {
  const next = selectNextPendingIntent(state);
  const result = execute(next);
  dispatch(recordExecutionResult(next.id, result));
  dispatch(dequeue(next.id));
}
```

## Open Questions

1. Should nav intents remain duplicated in both `pendingSystemIntents` and `pendingNavIntents`, or become a computed view?
2. Do we require exactly-once guarantees or is at-least-once acceptable with idempotent handlers?
3. Will effect host be component-driven, middleware-driven, or both (component for UI stacks, middleware for headless)?
4. Do we need per-session backpressure limits before rollout?
5. How will we represent poison messages / repeated failures?

## References

1. GEPA-14 design doc:
   - `go-go-gepa/ttmp/2026/02/27/GEPA-14-VM-JS-PROGRAMS--js-vm-programs-integration-for-hypercard-in-inventory-go-go-os/design-doc/01-js-vm-programs-hypercard-runtime-and-backend-command-wiring.md`
2. GEPA-14 intern Q&A doc:
   - `go-go-gepa/ttmp/2026/02/27/GEPA-14-VM-JS-PROGRAMS--js-vm-programs-integration-for-hypercard-in-inventory-go-go-os/design-doc/02-intern-q-a-store-boundaries-sessions-runtime-intents-and-re-rendering.md`
3. GEPA-17 event viewer plan:
   - `go-go-gepa/ttmp/2026/02/27/GEPA-17-PLUGIN-EVENT-VIEWER--js-plugin-runtime-event-viewer-for-inbound-ui-events-and-outbound-dispatched-actions/design-doc/01-plugin-runtime-event-viewer-architecture-and-implementation-plan.md`
4. Runtime slice and selectors:
   - `go-go-os/packages/engine/src/features/pluginCardRuntime/pluginCardRuntimeSlice.ts`
   - `go-go-os/packages/engine/src/features/pluginCardRuntime/selectors.ts`
5. Routing + host:
   - `go-go-os/packages/engine/src/components/shell/windowing/pluginIntentRouting.ts`
   - `go-go-os/packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx`
