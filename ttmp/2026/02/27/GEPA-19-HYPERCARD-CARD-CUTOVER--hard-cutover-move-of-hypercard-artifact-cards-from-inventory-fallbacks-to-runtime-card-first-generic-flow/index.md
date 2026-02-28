---
Title: Hard cutover move of HyperCard artifact/cards from inventory fallbacks to runtime-card-first generic flow
Ticket: GEPA-19-HYPERCARD-CARD-CUTOVER
Status: complete
Topics:
    - js-vm
    - hypercard
    - go-go-os
    - inventory-app
    - arc-agi
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-app-inventory/apps/inventory/src/domain/pluginBundle.vm.js
      Note: |-
        Inventory fallback card implementations targeted for removal
        Removed fallback viewer card implementations
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-app-inventory/apps/inventory/src/domain/stack.ts
      Note: |-
        Inventory fallback card metadata targeted for removal
        Removed fallback viewer card metadata
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/hypercard/artifacts/artifactRuntime.test.ts
      Note: Updated tests for runtime-card-required artifact opening
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/hypercard/artifacts/artifactRuntime.ts
      Note: |-
        Current fallback routing logic targeted for removal
        Runtime-card-first artifact open hard cutover implementation
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/hypercard/timeline/hypercardCard.tsx
      Note: Gated artifact controls by runtime card id
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/hypercard/timeline/hypercardWidget.tsx
      Note: |-
        Template-based widget edit fallback targeted for removal
        Removed template fallback editor routing and gated controls by runtime card id
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/27/GEPA-19-HYPERCARD-CARD-CUTOVER--hard-cutover-move-of-hypercard-artifact-cards-from-inventory-fallbacks-to-runtime-card-first-generic-flow/design-doc/01-hard-cutover-implementation-plan-move-hypercard-artifact-card-flow-off-inventory-fallback-cards.md
      Note: Primary implementation plan
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/27/GEPA-19-HYPERCARD-CARD-CUTOVER--hard-cutover-move-of-hypercard-artifact-cards-from-inventory-fallbacks-to-runtime-card-first-generic-flow/reference/01-implementation-diary.md
      Note: Chronological execution diary
ExternalSources: []
Summary: Hard-cutover ticket to remove inventory fallback artifact cards and enforce runtime-card-first generic HyperCard opening behavior.
LastUpdated: 2026-02-28T14:27:57.011935979-05:00
WhatFor: Plan, execute, and audit the migration away from inventory-coupled fallback card behavior.
WhenToUse: Use during GEPA-19 implementation and review.
---



# Hard cutover move of HyperCard artifact/cards from inventory fallbacks to runtime-card-first generic flow

## Overview

GEPA-19 removes template/inventory fallback behavior from HyperCard artifact/card opening and enforces runtime-card-first semantics.

## Key Documents

- Implementation plan: `design-doc/01-hard-cutover-implementation-plan-move-hypercard-artifact-card-flow-off-inventory-fallback-cards.md`
- Implementation diary: `reference/01-implementation-diary.md`

## Related Tickets

- `GEPA-14-VM-JS-PROGRAMS`
- `GEPA-17-PLUGIN-EVENT-VIEWER`
- `GEPA-18-ARC-AGI-FRONTEND-EXTRACTION`

## Status

- [x] Ticket workspace created
- [x] Detailed implementation plan authored
- [x] Granular task list authored
- [x] Engine hard cutover implemented
- [x] Inventory fallback removal implemented
- [x] Validation + final handoff complete

## Tasks

See [tasks.md](./tasks.md).

## Changelog

See [changelog.md](./changelog.md).
