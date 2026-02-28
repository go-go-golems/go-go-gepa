---
Title: Split HyperCard + runtime plugin architecture into dedicated package separate from desktop engine
Ticket: GEPA-26-HYPERCARD-RUNTIME-SPLIT
Status: active
Topics:
    - architecture
    - frontend
    - go-go-os
    - hypercard
    - js-vm
    - plugins
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-26-HYPERCARD-RUNTIME-SPLIT--split-hypercard-runtime-plugin-architecture-into-dedicated-package-separate-from-desktop-engine/design-doc/01-hypercard-runtime-package-split-architecture-and-migration-guide.md
      Note: Primary intern onboarding and migration design document
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-26-HYPERCARD-RUNTIME-SPLIT--split-hypercard-runtime-plugin-architecture-into-dedicated-package-separate-from-desktop-engine/reference/01-investigation-diary-hypercard-runtime-package-split.md
      Note: Detailed investigation diary with commands, findings, and decision trail
ExternalSources: []
Summary: Deep architecture and migration design for splitting HyperCard/runtime plugin systems from engine shell/windowing into dedicated package boundaries.
LastUpdated: 2026-02-28T15:35:00Z
WhatFor: Give new contributors a complete mental model and implementation plan for package decoupling.
WhenToUse: Use when implementing or reviewing runtime/plugin package separation and dependency-boundary hardening.
---

# Split HyperCard + runtime plugin architecture into dedicated package separate from desktop engine

## Overview

This ticket documents a full architecture pass over current runtime/plugin integration in `go-go-os`, then proposes a migration path that splits HyperCard and VM runtime concerns into dedicated packages while keeping desktop shell/windowing stable.

Primary outputs:

1. Intern-oriented architecture walkthrough from zero context.
2. Explicit dependency-map and coupling analysis across engine, desktop-os, demo apps, and arc-agi-player.
3. Phased package-split design with API boundaries, migration sequencing, and risk controls.

## Key Links

1. Design doc: `design-doc/01-hypercard-runtime-package-split-architecture-and-migration-guide.md`
2. Diary: `reference/01-investigation-diary-hypercard-runtime-package-split.md`
3. Tasks: `tasks.md`

## Status

Current status: **active**

## Topics

- architecture
- frontend
- go-go-os
- hypercard
- js-vm
- plugins

## Tasks

See [tasks.md](./tasks.md) for execution checklist.

## Changelog

See [changelog.md](./changelog.md) for incremental updates.
