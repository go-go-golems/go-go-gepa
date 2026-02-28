---
Title: Phase 2 ARC bridge for HyperCard runtime command execution
Ticket: GEPA-23-ARC-BRIDGE-PHASE-2
Status: active
Topics:
    - arc-agi
    - go-go-os
    - hypercard
    - js-vm
    - inventory-app
    - architecture
    - frontend
    - backend
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-app-arc-agi-3/apps/arc-agi-player/src/api/arcApi.ts
      Note: Existing ARC API wrappers to reuse in bridge op mapping
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-app-arc-agi-3/apps/arc-agi-player/src/bridge/middleware.ts
      Note: ARC bridge middleware executes arc/command.request lifecycle and API mappings
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-app-arc-agi-3/apps/arc-agi-player/src/domain/pluginBundle.ts
      Note: Demo VM card handlers dispatch ARC command requests
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-app-arc-agi-3/apps/arc-agi-player/src/domain/stack.ts
      Note: ARC demo HyperCard stack definition and capability policy
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-app-arc-agi-3/apps/arc-agi-player/src/launcher/module.tsx
      Note: Folder launch behavior providing React game and HyperCard stack entrypoints
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-app-arc-agi-3/pkg/backendmodule/routes.go
      Note: Backend route contract source for command execution mapping
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx
      Note: Runtime card host where event handlers and rerender behavior must consume ARC bridge updates
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/pluginIntentRouting.ts
      Note: |-
        Intent routing layer where arc/command.request dispatch path is defined
        Generic routing emits runtime metadata consumed by ARC bridge middleware
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/features/pluginCardRuntime/pluginCardRuntimeSlice.ts
      Note: Runtime ingest/timeline primitives and potential correlation metadata integration point
ExternalSources: []
Summary: Delivery ticket for implementing the ARC bridge that maps HyperCard VM runtime intents to ARC backend commands and projects command/session/game results back into Redux for card rerendering.
LastUpdated: 2026-02-28T05:49:00-05:00
WhatFor: Track implementation, testing, and rollout of Phase 2 ARC command bridge capabilities proposed in GEPA-14.
WhenToUse: Use when executing ARC bridge work from contract freeze through validation and closure.
---




# Phase 2 ARC bridge for HyperCard runtime command execution

## Overview

This ticket operationalizes the Phase 2 ARC bridge architecture from GEPA-14 into a concrete implementation stream. The design doc defines the command contracts, effect execution model, state shape, projection requirements, and acceptance criteria. `tasks.md` breaks execution into granular phases from kickoff and type contracts through integration tests and rollout cleanup.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- arc-agi
- go-go-os
- hypercard
- js-vm
- inventory-app
- architecture
- frontend
- backend

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
