---
Title: JS VM programs integration for HyperCard in inventory/go-go-os
Ticket: GEPA-14-VM-JS-PROGRAMS
Status: complete
Topics:
    - js-vm
    - hypercard
    - inventory-app
    - go-go-os
    - arc-agi
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-app-arc-agi-3/pkg/backendmodule/routes.go
      Note: ARC backend command API analyzed
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/plugin-runtime/runtimeService.ts
      Note: Core runtime implementation analyzed
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/27/GEPA-14-VM-JS-PROGRAMS--js-vm-programs-integration-for-hypercard-in-inventory-go-go-os/design-doc/01-js-vm-programs-hypercard-runtime-and-backend-command-wiring.md
      Note: Primary architecture deliverable
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/27/GEPA-14-VM-JS-PROGRAMS--js-vm-programs-integration-for-hypercard-in-inventory-go-go-os/design-doc/02-intern-q-a-store-boundaries-sessions-runtime-intents-and-re-rendering.md
      Note: Follow-up intern Q&A addendum covering stores, sessions, intents, rerendering, and pending queues
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/27/GEPA-14-VM-JS-PROGRAMS--js-vm-programs-integration-for-hypercard-in-inventory-go-go-os/reference/01-investigation-diary.md
      Note: Chronological research diary
ExternalSources: []
Summary: Deep architecture research and implementation blueprint for wiring ARC backend commands/state into HyperCard JS VM cards, with intern-focused follow-up Q&A addendum.
LastUpdated: 2026-02-28T14:27:56.519507551-05:00
WhatFor: Track GEPA-14 deliverables and cross-link core docs for intern onboarding and implementation handoff.
WhenToUse: Use when starting implementation work for VM-to-ARC command/state integration.
---



# JS VM programs integration for HyperCard in inventory/go-go-os

## Overview

`GEPA-14-VM-JS-PROGRAMS` documents how VM cards run today, how actions/state are exposed, and how to wire ARC backend commands so HyperCard cards can play/interact with ARC games.

## Key Documents

- Design doc: `design-doc/01-js-vm-programs-hypercard-runtime-and-backend-command-wiring.md`
- Follow-up Q&A addendum: `design-doc/02-intern-q-a-store-boundaries-sessions-runtime-intents-and-re-rendering.md`
- Diary: `reference/01-investigation-diary.md`

## Deliverable Status

- [x] In-depth architecture research completed
- [x] 10+ page design doc completed
- [x] Follow-up intern Q&A addendum completed
- [x] Chronological diary completed
- [x] `docmgr` relate/changelog/doctor complete
- [x] reMarkable upload complete

## Related Tickets Consulted

- `GEPA-12-ARC-AGI-OS-BACKEND-MODULE`
- `GEPA-13-ARC-AGI-WIDGET`
- `GEPA-17-PLUGIN-EVENT-VIEWER`

## Tasks

See [tasks.md](./tasks.md).

## Changelog

See [changelog.md](./changelog.md).
