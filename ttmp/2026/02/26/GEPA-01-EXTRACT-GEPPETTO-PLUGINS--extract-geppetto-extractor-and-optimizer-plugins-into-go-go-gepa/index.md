---
Title: Extract geppetto extractor and optimizer plugins into go-go-gepa
Ticket: GEPA-01-EXTRACT-GEPPETTO-PLUGINS
Status: complete
Topics:
    - architecture
    - plugins
    - extractor
    - optimizer
    - gepa
    - geppetto
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go-go-gepa/ttmp/2026/02/26/GEPA-01-EXTRACT-GEPPETTO-PLUGINS--extract-geppetto-extractor-and-optimizer-plugins-into-go-go-gepa/design-doc/01-migration-plan-extractor-and-optimizer-plugins.md
      Note: primary migration design
    - Path: go-go-gepa/ttmp/2026/02/26/GEPA-01-EXTRACT-GEPPETTO-PLUGINS--extract-geppetto-extractor-and-optimizer-plugins-into-go-go-gepa/reference/01-investigation-diary.md
      Note: chronological investigation evidence
ExternalSources: []
Summary: Ticket index for completed go-go-gepa implementation of plugin module ownership and registryIdentifier propagation.
LastUpdated: 2026-02-26T13:40:00-05:00
WhatFor: Record GEPA-01 implementation outcomes and traceability for follow-up work.
WhenToUse: Start here to navigate current design, diary, tasks, and changelog.
---

# Extract geppetto extractor and optimizer plugins into go-go-gepa

## Overview

This ticket tracks the completed GEPA-01 implementation in `go-go-gepa`:

1. own plugin module behavior in go-go-gepa,
2. propagate `registryIdentifier` through loader/runtime/reporting/storage,
3. keep no compatibility alias.

## Current Status

1. `go-go-gepa` now owns optimizer plugin contract helpers via native module `require("gepa/plugins")`.
2. `registryIdentifier` is carried through loader metadata, host context, hook tags, reports, and sqlite (`plugin_registry_identifier`).
3. Runner example scripts were migrated to `require("gepa/plugins")` and package tests cover decode defaults + recorder migration.
4. Scope remained constrained to `go-go-gepa/`; `gepa/` and `2026-02-18--cozodb-extraction/` stayed reference-only.

## Primary Deliverables

1. Design doc: `design-doc/01-migration-plan-extractor-and-optimizer-plugins.md`
2. Investigation diary: `reference/01-investigation-diary.md`
3. Execution checklist: `tasks.md`

## Tasks

See [tasks.md](./tasks.md).

## Changelog

See [changelog.md](./changelog.md).
