---
Title: Engine/Chat/Runtime package DAG cleanup and visibility-context redesign
Ticket: GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT
Status: complete
Topics:
    - architecture
    - frontend
    - go-go-os
    - hypercard
    - plugins
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT--engine-chat-runtime-package-dag-cleanup-and-visibility-context-redesign/design-doc/01-design-and-implementation-plan-split-chat-runtime-out-of-engine-and-introduce-generic-visibility-context.md
      Note: Primary design deliverable
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT--engine-chat-runtime-package-dag-cleanup-and-visibility-context-redesign/reference/01-investigation-diary-engine-chat-runtime-split-and-visibility-context.md
      Note: Chronological diary deliverable
ExternalSources: []
Summary: Research and design ticket for clean package DAG across engine, hypercard runtime, and chat runtime with generic visibility-context injection.
LastUpdated: 2026-02-28T18:02:30.609952552-05:00
WhatFor: Track design and upcoming implementation phases for GEPA-27 package-boundary refactor.
WhenToUse: Use when planning, implementing, or reviewing the engine/chat/runtime split and story relocation work.
---



# Engine/Chat/Runtime package DAG cleanup and visibility-context redesign

## Overview

This ticket defines and documents how to complete the next boundary cleanup after GEPA-26:

1. split chat runtime out of engine,
2. relocate runtime-coupled stories to owning packages/apps,
3. remove shell-internal chat schema assumptions using an injected visibility-context resolver.

## Key Links

- Design doc: `design-doc/01-design-and-implementation-plan-split-chat-runtime-out-of-engine-and-introduce-generic-visibility-context.md`
- Diary: `reference/01-investigation-diary-engine-chat-runtime-split-and-visibility-context.md`
- Tasks: `tasks.md`
- Changelog: `changelog.md`

## Status

Current status: **active**.

Research and design are complete. Implementation phases are listed in `tasks.md` and remain open.

## Topics

- architecture
- frontend
- go-go-os
- hypercard
- plugins

## Structure

- `design-doc/`: primary design and implementation plan
- `reference/`: chronological diary and quick references
- `playbooks/`: to be added during implementation
- `scripts/`: temporary migration helpers/codemods during implementation
