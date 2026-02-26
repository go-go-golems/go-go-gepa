---
Title: Analyze JS runner and design GEPA optimization tooling
Ticket: GEPA-02-ANALYZE-RUNNER
Status: active
Topics:
    - gepa
    - runner
    - goja
    - js-bindings
    - tooling
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Focused GEPA-02 analysis plus initial implementation of `dataset generate` under v2 constraints."
LastUpdated: 2026-02-26T14:15:00-05:00
WhatFor: "Entry point for GEPA-02 implementation and design work"
WhenToUse: "Use when onboarding to this ticket or locating current authoritative docs"
---

# GEPA-02: Analyze JS Runner and Foundational Building Blocks

## Scope for this ticket

1. Analyze how JS currently runs inside `go-go-gepa`.
2. Design two narrow building blocks (not full optimizer workflows):
   - `gepa candidate run`
   - `gepa dataset generate`
3. Keep chronological investigation diary with command evidence.

## Current Implementation Status

1. `dataset generate` is now implemented in `go-go-gepa` with:
   - Glazed command wiring,
   - `gepa.dataset-generator/v1` plugin loader,
   - `gepa.dataset-generate/v2` config parsing and key restrictions,
   - CLI-owned output routing (`--output-dir`, `--output-db`),
   - JSONL + sqlite persistence path.
2. `candidate run` remains pending.

## Authoritative design docs (current)

1. `design-doc/04-gepa-candidate-run-dev-tool-sqlite.md`
2. `design-doc/05-gepa-dataset-generate-llm-bootstrap.md`

## v2 Constraints Applied

1. `--script` is external CLI input for both commands.
2. Candidate-run uses separate files: config file and input file.
3. No output section in YAML configs; output/storage routing is CLI flags only.
4. Command definitions and parsing are Glazed-first.

## Older docs in this ticket

`design-doc/01..03` are retained as background material but are broader than current scope.

## Supporting docs

1. `reference/01-investigation-diary.md`
2. `tasks.md`
3. `changelog.md`

## Structure

- `design-doc/` primary design deliverables
- `reference/` investigation diary
- `scripts/` experiment logs and temporary artifacts
- `sources/` source captures
