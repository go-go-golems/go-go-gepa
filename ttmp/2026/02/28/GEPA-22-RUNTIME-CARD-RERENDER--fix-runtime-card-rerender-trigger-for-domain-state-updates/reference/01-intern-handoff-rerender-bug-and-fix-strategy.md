---
Title: 'Intern handoff: rerender bug and fix strategy'
Ticket: GEPA-22-RUNTIME-CARD-RERENDER
Status: active
Topics:
    - go-go-os
    - hypercard
    - frontend
    - inventory-app
    - architecture
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-28T01:40:00-05:00
WhatFor: "Give the intern a concise operational map for the rerender bug and the recommended fix path."
WhenToUse: "Use when understanding or extending GEPA-22 rerender invalidation behavior."
---

# Intern handoff: rerender bug and fix strategy

## Goal

Understand why domain updates can fail to rerender runtime cards and how to implement the planned fix safely.

## Context

### Current bug in one sentence

`PluginCardSessionHost` may reuse a cached tree when only domain slices change, because its memo dependencies do not include a subscribed domain projection dependency.

### Current behavior path

1. Host renders card tree with `renderCard(...)` in `useMemo`.
2. Global state projection reads `store.getState()` on demand.
3. If selected memo deps do not change, render is skipped.

## Quick Reference

### Files you must read first

1. `go-go-os/packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx`
2. `go-go-os/packages/engine/src/features/pluginCardRuntime/selectors.ts`
3. `go-go-gepa/ttmp/2026/02/27/GEPA-14-VM-JS-PROGRAMS--js-vm-programs-integration-for-hypercard-in-inventory-go-go-os/design-doc/02-intern-q-a-store-boundaries-sessions-runtime-intents-and-re-rendering.md`

### Recommended fix summary

- Add explicit domain projection selector subscription.
- Add stable projection fingerprint/reference dependency.
- Include dependency in tree memo invalidation path.

### Implementation status

Core GEPA-22 host fix is implemented in `go-go-os` with regression coverage.

## Usage Examples

### Detect stale render scenario

```text
update inventory reducer only
-> no cardState/sessionState/nav change
-> tree memo dependencies unchanged
-> renderCard not called
-> UI shows stale stock value
```

### Desired behavior after fix

```text
update projected domain slice
-> projection fingerprint changes
-> tree memo invalidates
-> renderCard called
-> card UI reflects latest domain data
```

## Related

1. `../design-doc/01-implementation-plan-runtime-card-rerender-trigger-fix-domain-projection-subscription.md`
2. `../tasks.md`
