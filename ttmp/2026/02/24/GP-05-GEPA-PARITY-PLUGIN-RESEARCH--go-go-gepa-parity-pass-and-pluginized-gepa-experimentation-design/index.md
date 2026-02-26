---
Title: go-go-gepa parity pass and pluginized GEPA experimentation design
Ticket: GP-05-GEPA-PARITY-PLUGIN-RESEARCH
Status: active
Topics:
    - architecture
    - tools
    - inference
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Research ticket documenting parity gaps between go-go-gepa and Python GEPA and proposing a pluginized experimentation architecture.
LastUpdated: 2026-02-24T10:47:00-05:00
WhatFor: Track and deliver two detailed research documents for parity and plugin extension design.
WhenToUse: Use when planning the next implementation pass over go-go-gepa optimizer semantics and plugin APIs.
---

# go-go-gepa parity pass and pluginized GEPA experimentation design

## Overview

This ticket contains two research deliverables:

1. A detailed parity pass of `go-go-gepa` versus Python `gepa` focused on initial frontier seeding, frontier semantics, component selection, and minibatch policy.
2. A plugin-extension architecture analysis describing where JS hooks can be added to accelerate GEPA variant experimentation.

## Key Links

- **Primary analysis**: `analysis/01-go-go-gepa-vs-python-gepa-parity-deep-analysis.md`
- **Plugin extension analysis**: `analysis/02-plugin-extension-points-for-gepa-workflow-experimentation.md`
- **Investigation diary**: `reference/01-investigation-diary.md`
- **Evidence script**: `scripts/01-collect-parity-and-plugin-evidence.sh`
- **Evidence artifacts**: `sources/`

## Status

Current status: **active**

## Topics

- architecture
- tools
- inference

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
