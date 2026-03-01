---
Title: Repository rename migration plan (go-go-os -> go-go-os-frontend)
Ticket: GEPA-32-FRONTEND-REPO-RENAME
Status: active
Topics:
    - frontend
    - migration
    - go-go-os
    - wesen-os
    - architecture
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-app-arc-agi-3/apps/arc-agi-player/tsconfig.json
      Note: ARC app alias rewiring
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-app-inventory/apps/inventory/tsconfig.json
      Note: Inventory dependency path rewiring
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os-backend/README.md
      Note: Backend extraction source wording alignment
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os-frontend/README.md
      Note: Repo identity rename details
    - Path: workspaces/2026-02-22/add-gepa-optimizer/wesen-os/apps/os-launcher/tsconfig.json
      Note: Primary downstream alias rewiring
ExternalSources: []
Summary: Prescriptive implementation plan for local and cross-repo migration from go-go-os to go-go-os-frontend.
LastUpdated: 2026-03-01T07:12:33.564491518-05:00
WhatFor: Guide execution of the post-rename migration so builds/tests continue to work across dependent repos.
WhenToUse: Use when applying or reviewing repository/path alias updates tied to the go-go-os-frontend rename.
---


# Repository rename migration plan (go-go-os -> go-go-os-frontend)

## Executive Summary

`go-go-os` has already been renamed on GitHub to `go-go-os-frontend`. The remaining work is local and cross-repo: align local directory naming, Git remotes, and all workspace path references that currently hardcode `go-go-os`.

This migration keeps behavior unchanged while preventing broken local builds/tests caused by stale relative paths (primarily in `wesen-os`, `go-go-app-inventory`, and `go-go-app-arc-agi-3`).

## Problem Statement

Multiple repos in the workspace import frontend source via filesystem aliases pointing at `../go-go-os/...`. After a local directory rename to match GitHub, these paths fail to resolve and frontend compilation/tests break.

The frontend repository itself also still uses the old remote URL and repo identity text in key docs.

## Proposed Solution

1. Rename the local repository directory from `go-go-os` to `go-go-os-frontend`.
2. Update the renamed repo's `origin` remote to `git@github.com:go-go-golems/go-go-os-frontend.git`.
3. Replace cross-repo filesystem path references:
   - `wesen-os`: TS path maps, Vite/Vitest aliases, test fixture paths, README workspace tree references.
   - `go-go-app-inventory`: TS path maps/references and shared Vite helper aliases.
   - `go-go-app-arc-agi-3`: TS path maps.
4. Update frontend/backed docs where the repo identity should now reference `go-go-os-frontend`.
5. Run targeted validation (`vitest`, `tsc -b`, and/or package tests as appropriate) in touched repos.
6. Record every step in the diary, including command output and any failures/fixes.

## Design Decisions

1. Preserve package/runtime naming where it is product-facing.
   - Example: visible UI title strings like "go-go-os Launcher" are not renamed unless they are explicitly repository-path concerns.
2. Limit edits to operational path/identity references.
   - Avoid broad wording churn in historical docs unless required for active onboarding or active tooling.
3. Keep commits repo-scoped.
   - One focused commit per impacted repo makes review and rollback straightforward.

## Alternatives Considered

1. Keep local directory as `go-go-os` and rely on GitHub redirect.
   - Rejected because local relative-path references still encode `go-go-os`; this leaves a mismatch with canonical repo identity and future automation.
2. Add symlink `go-go-os -> go-go-os-frontend`.
   - Rejected because symlinks hide unresolved references and add OS/tooling portability risk.
3. Bulk update all historical documentation across all tickets.
   - Rejected as unnecessary scope expansion; historical references remain valid context for past snapshots.

## Implementation Plan

1. Ticket setup and scoped inventory of affected files.
2. Local repo rename + remote normalization.
3. `wesen-os` path rewiring + tests.
4. `go-go-app-inventory` path rewiring + tests.
5. `go-go-app-arc-agi-3` path rewiring + tests.
6. Frontend/backed README identity touch-ups.
7. Final validation sweep and ticket bookkeeping (`tasks`, `changelog`, diary, relations, doctor).

## Open Questions

1. Whether the `wesen` fork repo will also be renamed to `wesen/go-go-os-frontend` (currently still `wesen/go-go-os`).
2. Whether any CI workflows outside this workspace still clone `go-go-os` by old path and need separate updates.

## References

1. `wesen-os/apps/os-launcher/tsconfig.json`
2. `wesen-os/apps/os-launcher/vite.config.ts`
3. `wesen-os/apps/os-launcher/vitest.config.ts`
4. `go-go-app-inventory/apps/inventory/tsconfig.json`
5. `go-go-app-inventory/tooling/vite/createHypercardViteConfig.ts`
6. `go-go-app-arc-agi-3/apps/arc-agi-player/tsconfig.json`
