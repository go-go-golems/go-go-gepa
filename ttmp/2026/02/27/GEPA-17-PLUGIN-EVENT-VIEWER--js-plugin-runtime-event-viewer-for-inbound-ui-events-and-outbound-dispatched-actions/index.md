---
Title: JS plugin runtime event viewer for inbound UI events and outbound dispatched actions
Ticket: GEPA-17-PLUGIN-EVENT-VIEWER
Status: complete
Topics:
    - js-vm
    - event-streaming
    - go-go-os
    - hypercard
    - inventory-app
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/27/GEPA-17-PLUGIN-EVENT-VIEWER--js-plugin-runtime-event-viewer-for-inbound-ui-events-and-outbound-dispatched-actions/design-doc/01-plugin-runtime-event-viewer-architecture-and-implementation-plan.md
      Note: |-
        Primary architecture and implementation deliverable
        Primary analysis deliverable for GEPA-17
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/27/GEPA-17-PLUGIN-EVENT-VIEWER--js-plugin-runtime-event-viewer-for-inbound-ui-events-and-outbound-dispatched-actions/reference/01-investigation-diary.md
      Note: |-
        Chronological command-level investigation diary
        Chronological diary deliverable for GEPA-17
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx
      Note: |-
        Host runtime event entrypoint and intent dispatch orchestration evidence
        Host runtime event path evidence
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/pluginIntentRouting.ts
      Note: |-
        Intent ingest + forwarding behavior and capability gate evidence
        Intent routing evidence
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/features/pluginCardRuntime/pluginCardRuntimeSlice.ts
      Note: Runtime timeline and pending queue semantics evidence
ExternalSources: []
Summary: Deep architecture investigation and implementation blueprint for adding a running JS-plugin event viewer that traces inbound UI events and outbound dispatched actions.
LastUpdated: 2026-02-28T14:27:56.671437991-05:00
WhatFor: Track GEPA-17 research outputs and implementation guidance for plugin runtime observability.
WhenToUse: Use when implementing or validating plugin event-stream observability in go-go-os/inventory runtime sessions.
---



# JS plugin runtime event viewer for inbound UI events and outbound dispatched actions

## Overview

`GEPA-17-PLUGIN-EVENT-VIEWER` captures how plugin runtime events currently flow and provides an implementation-grade plan for an event viewer that shows both inbound UI events and outbound dispatched actions in running JS plugin sessions.

## Key Documents

- Design doc: `design-doc/01-plugin-runtime-event-viewer-architecture-and-implementation-plan.md`
- Diary: `reference/01-investigation-diary.md`

## Deliverable Status

- [x] In-depth architecture investigation completed
- [x] 5+ page implementation design document completed
- [x] Chronological investigation diary completed
- [x] `docmgr` relate/changelog/doctor validation complete
- [x] reMarkable upload complete

## Related Tickets Consulted

- `GEPA-14-VM-JS-PROGRAMS`

## Tasks

See [tasks.md](./tasks.md).

## Changelog

See [changelog.md](./changelog.md).
