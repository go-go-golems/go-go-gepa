---
Title: Dedicated HyperCard Tools app and typed runtime-card editor routing
Ticket: GEPA-20-HYPERCARD-TOOLS-APP
Status: active
Topics:
    - hypercard
    - go-go-os
    - inventory-app
    - arc-agi
    - frontend
    - modules
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-28T00:16:57.281895573-05:00
WhatFor: "Track GEPA-20 hard-cutover work that moves runtime-card editor routing to a dedicated hypercard-tools launcher app with typed runtime-card references."
WhenToUse: "Use as the ticket entrypoint for implementation status, tasks, diary, and validation."
---

# Dedicated HyperCard Tools app and typed runtime-card editor routing

## Overview

GEPA-20 implements the ideal architecture for runtime-card editor routing:

1. Dedicated `hypercard-tools` launcher module.
2. Typed runtime-card editor identity (`RuntimeCardRef`).
3. Hard-cutover away from inventory-owned editor rendering and legacy `code-editor:*` keys.

Current status:

1. Code changes completed across `go-go-os`, `go-go-app-inventory`, and `wesen-os`.
2. Automated validation completed (tests + typechecks).
3. Manual smoke test remains open (task 7.3).

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- hypercard
- go-go-os
- inventory-app
- arc-agi
- frontend
- modules

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
