---
Title: Intern handoff notes and glossary
Ticket: GEPA-21-QUEUE-FIRST-RUNTIME-ALIGNMENT
Status: active
Topics:
    - go-go-os
    - hypercard
    - event-streaming
    - js-vm
    - inventory-app
    - architecture
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-21-QUEUE-FIRST-RUNTIME-ALIGNMENT--queue-first-runtime-intent-execution-and-gepa-14-gepa-17-viewer-alignment-plan/tasks.md
      Note: Step-by-step backlog linked from intern handoff
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/pluginIntentRouting.ts
      Note: Intern glossary references this file for mixed-mode routing behavior
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/features/pluginCardRuntime/pluginCardRuntimeSlice.ts
      Note: Intern onboarding focuses on queue structures and ingest behavior
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-28T00:28:00.784824891-05:00
WhatFor: Provide intern-ready onboarding notes, glossary, and practical guidance for the deferred GEPA-21 queue-first migration.
WhenToUse: Use before and during GEPA-21 implementation kickoff.
---


# Intern handoff notes and glossary

## Goal

Make GEPA-21 understandable to a new engineer by defining:

1. what exists today,
2. what we eventually want,
3. what not to change yet,
4. where to read and how to start safely.

## Context

### Important: not being implemented now

GEPA-21 is intentionally deferred.

For now, the system keeps mixed behavior:

1. runtime intents are ingested into timeline + queues,
2. domain/system intents are also immediately routed,
3. pending queues are present but not the primary execution path.

This ticket stores the future plan and task list only.

## Quick Reference

### Core files

1. `go-go-os/packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx`
2. `go-go-os/packages/engine/src/components/shell/windowing/pluginIntentRouting.ts`
3. `go-go-os/packages/engine/src/features/pluginCardRuntime/pluginCardRuntimeSlice.ts`
4. `go-go-os/packages/engine/src/features/pluginCardRuntime/selectors.ts`
5. `go-go-os/packages/engine/src/chat/debug/EventViewerWindow.tsx`

### What each key function does (today)

- `emitRuntimeEvent(...)`:
  receives UI callback from rendered runtime card, calls VM `eventCard`, gets intents.
- `dispatchRuntimeIntent(...)`:
  dispatches `ingestRuntimeIntent(...)`, then immediately routes domain/system intents.
- `ingestRuntimeIntent(...)` reducer:
  updates timeline, local/session state, and appends queue envelopes.

### Glossary

- Runtime intent: a VM-produced instruction with `scope` (`card`, `session`, `domain`, `system`).
- Session intent: intent tied to one runtime session (`sessionId`) and card context.
- Pending domain queue: state queue for domain action envelopes.
- Pending system queue: state queue for system command envelopes.
- Pending nav queue: subset/duplicate of system queue for `nav.*` commands.
- Queue-first: execution model where ingest writes queue and a dedicated consumer executes effects.
- Mixed mode: current model where ingest + immediate routing both exist.
- Lifecycle telemetry: event stream that marks intent states from produced to terminal outcome.

### Reading order for intern (recommended)

1. GEPA-14 Q&A doc (concepts and terminology).
2. `pluginCardRuntimeSlice.ts` (state model and ingest behavior).
3. `pluginIntentRouting.ts` (where immediate routing still happens).
4. GEPA-17 viewer plan (what observability needs to show).
5. GEPA-21 design plan (how migration should be sequenced).

### What to avoid when implementation starts

1. Do not change queue + routing behavior in one giant PR.
2. Do not remove immediate route before effect host exists and passes tests.
3. Do not break existing event viewer filters while introducing lifecycle states.
4. Do not assume nav queue semantics are finalized; validate with owner first.

## Usage Examples

### Example 1: Trace one UI click in today’s mixed model

```text
click button
-> emitRuntimeEvent
-> VM returns 2 intents
-> dispatchRuntimeIntent(intent A)
   -> ingestRuntimeIntent (timeline + queue)
   -> immediate route
-> dispatchRuntimeIntent(intent B)
   -> ingestRuntimeIntent (timeline + queue)
   -> immediate route
```

### Example 2: Intended queue-first trace (future)

```text
click button
-> emitRuntimeEvent
-> VM returns intents
-> dispatchRuntimeIntent
   -> ingest only (timeline + queue)
-> RuntimeIntentEffectHost drains queue
   -> execute
   -> record outcome
   -> dequeue
```

## Related

1. `../design-doc/01-implementation-plan-queue-first-runtime-execution-and-event-viewer-alignment-deferred.md`
2. `../tasks.md`
3. `../../2026/02/27/GEPA-14-VM-JS-PROGRAMS--js-vm-programs-integration-for-hypercard-in-inventory-go-go-os/design-doc/02-intern-q-a-store-boundaries-sessions-runtime-intents-and-re-rendering.md`
4. `../../2026/02/27/GEPA-17-PLUGIN-EVENT-VIEWER--js-plugin-runtime-event-viewer-for-inbound-ui-events-and-outbound-dispatched-actions/design-doc/01-plugin-runtime-event-viewer-architecture-and-implementation-plan.md`
