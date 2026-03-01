---
Title: Move chat UI out of engine into chat-runtime
Ticket: GEPA-29
Status: complete
Topics:
    - architecture
    - frontend
    - go-go-os
    - chat-runtime
    - hypercard
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-28T20:00:54.984821949-05:00
WhatFor: ""
WhenToUse: ""
---


# Move chat UI out of engine into chat-runtime

## Overview

GEPA-29 extracted chat-specific UI composition from `@hypercard/engine` into `@hypercard/chat-runtime`, including widget components, sidebar, message typing, and stylesheet ownership.

The ticket also introduced explicit chat theme loading via `@hypercard/chat-runtime/theme` and removed implicit chat styling from `@hypercard/engine/theme`.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **complete**

## Topics

- architecture
- frontend
- go-go-os
- chat-runtime
- hypercard

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
