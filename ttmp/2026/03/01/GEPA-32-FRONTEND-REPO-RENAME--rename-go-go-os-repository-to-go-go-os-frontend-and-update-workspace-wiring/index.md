---
Title: Rename go-go-os repository to go-go-os-frontend and update workspace wiring
Ticket: GEPA-32-FRONTEND-REPO-RENAME
Status: active
Topics:
    - frontend
    - migration
    - go-go-os
    - wesen-os
    - architecture
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-app-arc-agi-3/apps/arc-agi-player/tsconfig.json
      Note: ARC consumer path rewiring
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-app-inventory/apps/inventory/tsconfig.json
      Note: Inventory consumer path rewiring
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os-frontend/README.md
      Note: Renamed frontend repo identity
    - Path: workspaces/2026-02-22/add-gepa-optimizer/wesen-os/apps/os-launcher/tsconfig.json
      Note: Consumer workspace path rewiring
ExternalSources: []
Summary: Post-GitHub-rename implementation ticket for migrating local workspace and cross-repo aliases from go-go-os to go-go-os-frontend.
LastUpdated: 2026-03-01T07:12:30.905562332-05:00
WhatFor: Use this ticket to track concrete migration steps and validation evidence for the frontend repository rename follow-up.
WhenToUse: When workspace path aliases, remotes, and docs must be aligned to go-go-os-frontend after the GitHub rename.
---


# Rename go-go-os repository to go-go-os-frontend and update workspace wiring

## Overview

This ticket tracks post-GitHub-rename cleanup so the local monorepo workspace and cross-repo tooling consistently use `go-go-os-frontend` instead of `go-go-os` for filesystem paths and remote naming.

The objective is to keep all affected repos runnable after the rename while preserving focused commits and an implementation diary with exact commands, failures, and validation outcomes.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field
- Design doc: `design-doc/01-repository-rename-migration-plan-go-go-os-go-go-os-frontend.md`
- Diary: `reference/01-investigation-diary.md`

## Status

Current status: **active**

## Topics

- frontend
- migration
- go-go-os
- wesen-os
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
