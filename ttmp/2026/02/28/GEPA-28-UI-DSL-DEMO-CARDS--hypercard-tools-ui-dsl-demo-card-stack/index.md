---
Title: HyperCard Tools UI DSL demo card stack
Ticket: GEPA-28-UI-DSL-DEMO-CARDS
Status: active
Topics:
    - frontend
    - go-go-os
    - hypercard
    - ui-dsl
    - demo
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-28-UI-DSL-DEMO-CARDS--hypercard-tools-ui-dsl-demo-card-stack/design-doc/01-ui-dsl-demo-cards-architecture-implementation-plan-and-intern-onboarding-guide.md
      Note: Primary architecture and implementation handoff for the intern.
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-28-UI-DSL-DEMO-CARDS--hypercard-tools-ui-dsl-demo-card-stack/reference/01-investigation-and-implementation-diary.md
      Note: Chronological diary with command evidence and implementation notes.
ExternalSources: []
Summary: Build a full HyperCard Tools UI DSL demo stack, document architecture and implementation details for a new intern, and deliver docs via reMarkable.
LastUpdated: 2026-02-28T18:35:00-05:00
WhatFor: Track the design, implementation, validation, and documentation for HyperCard Tools UI DSL demo cards.
WhenToUse: Use this ticket when onboarding engineers to HyperCard runtime cards or when extending demo coverage for the current UI DSL surface.
---

# HyperCard Tools UI DSL demo card stack

## Overview

This ticket delivers a dedicated, navigable demo stack under HyperCard Tools to showcase the current UI DSL widget surface, layout patterns, state mutation patterns, and system-intent behavior.

The implementation includes:

- a full card stack in `apps/hypercard-tools`,
- launcher wiring so opening HyperCard Tools lands on the demo folder/stack experience,
- documentation aimed at a new intern with no prior context,
- a diary and explicit phased task tracking.

## Key Links

- Design + intern guide: `design-doc/01-ui-dsl-demo-cards-architecture-implementation-plan-and-intern-onboarding-guide.md`
- Diary: `reference/01-investigation-and-implementation-diary.md`
- Task checklist: `tasks.md`
- Ticket history: `changelog.md`

## Status

Current status: **active**.

Research and implementation are in progress.

## Topics

- frontend
- go-go-os
- hypercard
- ui-dsl
- demo

## Structure

- `design-doc/`: primary architecture + implementation handoff
- `reference/`: diary with step-by-step implementation notes
- `playbooks/`: deterministic validation sequences (to be added)
- `scripts/`: ad-hoc helpers if needed
- `various/`: intermediate working notes and exports
- `archive/`: deprecated artifacts
