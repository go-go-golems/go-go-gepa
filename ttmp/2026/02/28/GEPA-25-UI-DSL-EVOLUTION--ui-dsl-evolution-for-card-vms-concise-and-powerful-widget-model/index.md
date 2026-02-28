---
Title: 'UI DSL evolution for card VMs: concise and powerful widget model'
Ticket: GEPA-25-UI-DSL-EVOLUTION
Status: active
Topics:
    - architecture
    - frontend
    - go-go-os
    - hypercard
    - js-vm
    - ui
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/02/28/GEPA-25-UI-DSL-EVOLUTION--ui-dsl-evolution-for-card-vms-concise-and-powerful-widget-model/design-doc/01-ui-dsl-architecture-audit-and-proposal-for-card-vm-widgets.md
      Note: Primary architecture and proposal deliverable
    - Path: ttmp/2026/02/28/GEPA-25-UI-DSL-EVOLUTION--ui-dsl-evolution-for-card-vms-concise-and-powerful-widget-model/reference/01-investigation-diary-ui-dsl-and-widget-evolution.md
      Note: Chronological command-and-findings diary
ExternalSources: []
Summary: Research and proposal ticket for evolving card VM UI DSL to stay concise while reusing richer engine widget capabilities.
LastUpdated: 2026-02-28T19:49:00Z
WhatFor: Provide implementation-ready guidance for UI DSL contract evolution, renderer alignment, and phased adoption.
WhenToUse: Use when planning or implementing card VM UI DSL improvements in go-go-os.
---



# UI DSL evolution for card VMs: concise and powerful widget model

## Overview

This ticket analyzes the card VM UI DSL (`ui.panel`, `ui.text`, etc.), compares it with the expanded engine widget system, and proposes a concise, elegant next-step DSL model.

Primary outcomes:

1. Evidence-backed architecture audit of DSL exposure, schema validation, renderer behavior, and intent routing.
2. Inventory of existing and recently added engine widgets relevant to VM card authoring.
3. Intern-focused, implementation-ready proposal for a compact DSL v2 (`core primitives + curated widget extension pattern`).

## Key Links

1. Design doc: `design-doc/01-ui-dsl-architecture-audit-and-proposal-for-card-vm-widgets.md`
2. Diary: `reference/01-investigation-diary-ui-dsl-and-widget-evolution.md`

## Status

Current status: **active**

## Topics

- architecture
- frontend
- go-go-os
- hypercard
- js-vm
- ui

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
