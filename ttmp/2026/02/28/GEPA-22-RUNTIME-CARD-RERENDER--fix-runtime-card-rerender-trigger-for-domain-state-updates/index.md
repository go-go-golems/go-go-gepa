---
Title: Fix runtime card rerender trigger for domain state updates
Ticket: GEPA-22-RUNTIME-CARD-RERENDER
Status: active
Topics:
    - go-go-os
    - hypercard
    - frontend
    - inventory-app
    - architecture
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-22-RUNTIME-CARD-RERENDER--fix-runtime-card-rerender-trigger-for-domain-state-updates/design-doc/01-implementation-plan-runtime-card-rerender-trigger-fix-domain-projection-subscription.md
      Note: Detailed plan for suggested rerender fix
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-22-RUNTIME-CARD-RERENDER--fix-runtime-card-rerender-trigger-for-domain-state-updates/reference/01-intern-handoff-rerender-bug-and-fix-strategy.md
      Note: Intern-oriented quick handoff and bug explanation
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-22-RUNTIME-CARD-RERENDER--fix-runtime-card-rerender-trigger-for-domain-state-updates/tasks.md
      Note: Granular not-started execution checklist
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-28T00:34:01.185339049-05:00
WhatFor: Track the planned fix for runtime-card rerender invalidation when domain-only Redux updates do not trigger host recomputation.
WhenToUse: Use as GEPA-22 entrypoint before and during implementation kickoff.
---


# Fix runtime card rerender trigger for domain state updates

## Overview

GEPA-22 captures the fix plan for the rerender gap documented in GEPA-14 intern Q&A.

Suggested fix to implement later:

1. explicit domain projection selector subscription,
2. stable projection fingerprint/reference,
3. memo dependency wiring in `PluginCardSessionHost` tree computation.

Status for now:

1. planning and tasks are complete,
2. implementation has not started yet.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

Execution state: **not started**.

## Topics

- go-go-os
- hypercard
- frontend
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
