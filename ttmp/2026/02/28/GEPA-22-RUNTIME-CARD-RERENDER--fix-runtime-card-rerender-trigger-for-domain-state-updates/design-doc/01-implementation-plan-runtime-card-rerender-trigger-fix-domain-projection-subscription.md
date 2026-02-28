---
Title: 'Implementation plan: runtime card rerender trigger fix (domain projection subscription)'
Ticket: GEPA-22-RUNTIME-CARD-RERENDER
Status: active
Topics:
    - go-go-os
    - hypercard
    - frontend
    - inventory-app
    - architecture
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/27/GEPA-14-VM-JS-PROGRAMS--js-vm-programs-integration-for-hypercard-in-inventory-go-go-os/design-doc/01-js-vm-programs-hypercard-runtime-and-backend-command-wiring.md
      Note: Broader host/runtime flow context referenced by this implementation plan
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/27/GEPA-14-VM-JS-PROGRAMS--js-vm-programs-integration-for-hypercard-in-inventory-go-go-os/design-doc/02-intern-q-a-store-boundaries-sessions-runtime-intents-and-re-rendering.md
      Note: Source analysis that identifies rerender trigger gap and recommended Option A
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx
      Note: Primary host render memo/dependency location where rerender invalidation fix will be implemented
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/features/pluginCardRuntime/selectors.ts
      Note: Existing selector surface to extend with projected domain selector
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-28T00:34:01.291433297-05:00
WhatFor: Provide a full implementation plan for fixing the current runtime-card rerender gap when domain state changes do not trigger PluginCardSessionHost recomputation.
WhenToUse: Use when starting GEPA-22 implementation after explicit kickoff approval.
---


# Implementation plan: runtime card rerender trigger fix (domain projection subscription)

## Executive Summary

GEPA-22 addresses a known runtime rendering bug: domain-only Redux updates can fail to rerender VM cards because `PluginCardSessionHost` memo dependencies do not subscribe to those domain changes.

Suggested fix to implement (from GEPA-14 Q&A):

1. add explicit domain projection selector subscription,
2. compute stable projection fingerprint,
3. include this fingerprint in `renderCard` memo dependencies.

This ticket is planning-only right now. No code execution should begin until kickoff approval.

## Problem Statement

In current host flow:

1. `tree` is calculated with `useMemo` in `PluginCardSessionHost`.
2. `projectGlobalState(store.getState(), ...)` is pull-based.
3. rerender depends on selected slices (`cardState`, `sessionState`, nav, focus, runtime status), not arbitrary domain slice revisions.

If a change occurs only in a domain slice (for example, inventory stock values), the card tree can stay stale because memo dependencies do not change.

## Proposed Solution

### Recommended fix (Option A)

Introduce a domain projection subscription path:

1. Add selector that derives only domain data relevant to current card/session.
2. Produce stable fingerprint/hash/reference for that projection.
3. Include fingerprint in `tree` `useMemo` dependencies.
4. Keep `projectGlobalState` output semantics unchanged for card runtime contract.

### Why Option A

1. Better precision than forced rerender on every store update.
2. Lower performance risk than global brute-force subscriptions.
3. Preserves app-domain boundaries and future generic card behavior.

### Sketch

```ts
const domainProjection = useSelector((state) =>
  selectProjectedDomainsForRuntimeCard(state, {
    stackId: stack.id,
    sessionId,
    cardId: currentCardId,
  })
);

const domainProjectionFingerprint = useMemo(
  () => stableHash(domainProjection),
  [domainProjection],
);

const tree = useMemo(() => {
  const projectedGlobalState = projectGlobal();
  return runtimeServiceRef.current?.renderCard(
    sessionId,
    currentCardId,
    cardState,
    sessionState,
    projectedGlobalState,
  ) ?? null;
}, [
  cardState,
  sessionState,
  currentCardId,
  runtimeSession?.status,
  projectGlobal,
  domainProjectionFingerprint,
]);
```

## Design Decisions

1. Use explicit selector subscription instead of full-store forced rerender.
Reason: predictable performance and targeted invalidation.

2. Keep projection app-neutral (`globalState.domains` shape still generic).
Reason: VM runtime must stay reusable across apps.

3. Add regression tests before rollout.
Reason: rerender bugs are easy to reintroduce during host refactors.

4. Do not start implementation yet.
Reason: user asked for ticket/plan/tasks only at this time.

## Alternatives Considered

1. Force rerender on all store updates.
Rejected as default approach due avoidable perf churn.

2. Session-global numeric revision counter.
Possible fallback, but less expressive than projection-based invalidation.

3. Per-card custom selectors in each app.
Rejected for core path because host should remain generic, with app-specific projection hooks layered cleanly.

## Implementation Plan

### Phase 0: Kickoff and guardrails

1. Confirm GEPA-22 kickoff approval.
2. Freeze files in scope and expected outcomes.
3. Add feature branch and baseline perf snapshots.

### Phase 1: Baseline instrumentation

1. Add temporary debug counters:
   - render invocations,
   - domain projection changes,
   - state updates without rerender.
2. Reproduce stock-update stale-render scenario.
3. Capture baseline traces for comparison.

### Phase 2: Domain projection selector contract

1. Define selector API for projected domain slices.
2. Implement stable identity strategy:
   - memoized selector output or stable hash.
3. Add unit tests for selector stability and change detection.

### Phase 3: Host integration

1. Wire selector subscription into `PluginCardSessionHost`.
2. Add projection fingerprint into tree memo dependencies.
3. Ensure no behavior regressions for card/session/nav-driven rerenders.

### Phase 4: Performance and correctness hardening

1. Measure rerender frequency before/after.
2. Confirm no excessive rerender loops.
3. Optimize projection size or hashing if needed.

### Phase 5: Tests

1. Add regression test: domain-only state update triggers rerender.
2. Add negative test: unrelated domain update does not trigger rerender if projection excludes it.
3. Add integration test with runtime card reading domain state.

### Phase 6: Documentation and handoff

1. Update GEPA-14 references with implemented outcome.
2. Add “how rerender invalidation works” section to intern docs.
3. Close GEPA-22 with test evidence.

## Open Questions

1. Should projection fingerprinting be hash-based or reference-based only?
2. Where should app-specific domain projection rules live long-term?
3. Do we need a debug toggle in runtime debug window for projection/rerender diagnostics?

## References

1. `go-go-os/packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx`
2. `go-go-os/packages/engine/src/features/pluginCardRuntime/selectors.ts`
3. `go-go-gepa/ttmp/2026/02/27/GEPA-14-VM-JS-PROGRAMS--js-vm-programs-integration-for-hypercard-in-inventory-go-go-os/design-doc/02-intern-q-a-store-boundaries-sessions-runtime-intents-and-re-rendering.md`
4. `go-go-gepa/ttmp/2026/02/27/GEPA-14-VM-JS-PROGRAMS--js-vm-programs-integration-for-hypercard-in-inventory-go-go-os/design-doc/01-js-vm-programs-hypercard-runtime-and-backend-command-wiring.md`
