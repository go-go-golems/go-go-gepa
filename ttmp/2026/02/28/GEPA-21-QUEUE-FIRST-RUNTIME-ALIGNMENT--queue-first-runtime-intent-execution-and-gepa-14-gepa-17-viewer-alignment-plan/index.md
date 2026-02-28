---
Title: Queue-first runtime intent execution and GEPA-14/GEPA-17 viewer alignment plan
Ticket: GEPA-21-QUEUE-FIRST-RUNTIME-ALIGNMENT
Status: active
Topics:
    - go-go-os
    - hypercard
    - event-streaming
    - js-vm
    - inventory-app
    - architecture
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-21-QUEUE-FIRST-RUNTIME-ALIGNMENT--queue-first-runtime-intent-execution-and-gepa-14-gepa-17-viewer-alignment-plan/design-doc/01-implementation-plan-queue-first-runtime-execution-and-event-viewer-alignment-deferred.md
      Note: Primary deferred implementation plan for intern onboarding
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-21-QUEUE-FIRST-RUNTIME-ALIGNMENT--queue-first-runtime-intent-execution-and-gepa-14-gepa-17-viewer-alignment-plan/reference/01-intern-handoff-notes-and-glossary.md
      Note: Intern glossary and practical safe-start notes
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-21-QUEUE-FIRST-RUNTIME-ALIGNMENT--queue-first-runtime-intent-execution-and-gepa-14-gepa-17-viewer-alignment-plan/tasks.md
      Note: Granular deferred backlog for future execution
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-28T00:28:00.557431646-05:00
WhatFor: Plan a future queue-first runtime intent execution model and align GEPA-17 event viewer semantics with the GEPA-14 runtime intent lifecycle, while deferring implementation for now.
WhenToUse: Use when starting the deferred queue-first implementation or onboarding an intern to this subsystem.
---


# Queue-first runtime intent execution and GEPA-14/GEPA-17 viewer alignment plan

## Overview

This ticket is a deferred implementation plan only.

1. We are not executing this migration today.
2. Current behavior remains unchanged:
   - immediate routing in `dispatchRuntimeIntent(...)` remains active,
   - pending runtime queues remain in state and are intentionally ignored operationally.
3. The deliverable in this ticket is planning quality for an intern:
   - full architecture explanation,
   - phased implementation design,
   - granular task backlog.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

Execution mode for now: **deferred / planning only**.

## Topics

- go-go-os
- hypercard
- event-streaming
- js-vm
- inventory-app
- architecture

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
