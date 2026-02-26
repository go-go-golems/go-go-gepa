---
Title: Extract geppetto extractor and optimizer plugins into go-go-gepa
Ticket: GEPA-01-EXTRACT-GEPPETTO-PLUGINS
Status: active
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
Summary: Ticket index for migrating plugin contract ownership from geppetto into go-go-gepa with registry identifier propagation.
LastUpdated: 2026-02-26T11:40:46-05:00
WhatFor: Coordinate investigation and implementation planning for plugin contract extraction.
WhenToUse: Start here to navigate design, diary, tasks, and changelog for this ticket.
---


# Extract geppetto extractor and optimizer plugins into go-go-gepa

## Overview

This ticket documents how to remove extractor/optimizer plugin contract ownership from `geppetto` and make it a `go-go-gepa` concern, while preserving runtime compatibility and adding a carried `registryIdentifier`.

## Primary Deliverables

1. Design doc:
- `design-doc/01-migration-plan-extractor-and-optimizer-plugins.md`

2. Investigation diary:
- `reference/01-investigation-diary.md`

## Current Status

1. Analysis complete with file-backed architecture mapping.
2. Phased migration plan defined (ownership, compatibility alias, registry identifier carriage).
3. Implementation not started in this ticket yet.

## Key Decisions

1. Plugin contract helpers are not core framework APIs and should move out of `geppetto`.
2. Use `go-go-gepa` as contract owner with canonical module `gepa/plugins`.
3. Keep temporary `geppetto/plugins` alias during migration.
4. Introduce `registryIdentifier` in plugin metadata and propagate through reports/recorders.

## Tasks

See [tasks.md](./tasks.md).

## Changelog

See [changelog.md](./changelog.md).
