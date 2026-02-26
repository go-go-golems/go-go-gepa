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
Summary: Ticket index for hard-cut removal of geppetto/plugins and follow-up registry identifier propagation.
LastUpdated: 2026-02-26T12:34:00-05:00
WhatFor: Coordinate investigation and implementation planning for plugin contract extraction.
WhenToUse: Start here to navigate design, diary, tasks, and changelog for this ticket.
---


# Extract geppetto extractor and optimizer plugins into go-go-gepa

## Overview

This ticket documents the hard-cut removal of `geppetto/plugins` from core geppetto runtime, migration of affected scripts, and follow-up work to carry `registryIdentifier`.

## Primary Deliverables

1. Design doc:
- `design-doc/01-migration-plan-extractor-and-optimizer-plugins.md`

2. Investigation diary:
- `reference/01-investigation-diary.md`

## Current Status

1. Hard-cut removal of `geppetto/plugins` is implemented and committed.
2. Extractor scripts that depended on `geppetto/plugins` were migrated to plain descriptors.
3. Remaining work is registry identifier propagation through plugin metadata/reporting.

## Key Decisions

1. Plugin contract helpers are not core framework APIs and should move out of `geppetto`.
2. No compatibility alias (`geppetto/plugins`) is provided; removal is immediate.
3. Keep local helper `gepa_plugin_contract.js` in `go-go-gepa` for optimizer scripts.
4. Introduce `registryIdentifier` in plugin metadata and propagate through reports/recorders.

## Tasks

See [tasks.md](./tasks.md).

## Changelog

See [changelog.md](./changelog.md).
