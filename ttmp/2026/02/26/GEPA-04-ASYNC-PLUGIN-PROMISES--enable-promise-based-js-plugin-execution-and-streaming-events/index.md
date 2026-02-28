---
Title: Enable Promise-based JS plugin execution and streaming events
Ticket: GEPA-04-ASYNC-PLUGIN-PROMISES
Status: complete
Topics:
    - gepa
    - plugins
    - goja
    - runner
    - events
    - js-bindings
    - go
    - tooling
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/plugin_loader.go
      Note: Optimizer plugin loader currently assumes synchronous return values.
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/pkg/dataset/generator/plugin_loader.go
      Note: Dataset generator plugin loader currently assumes synchronous return values.
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/js_runtime.go
      Note: Shared JS runtime/eventloop entrypoint used by commands.
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/geppetto/pkg/js/modules/geppetto/api_sessions.go
      Note: Geppetto JS already exposes Promise and streaming primitives.
ExternalSources: []
Summary: Scope and plan ticket for enabling Promise-returning JS plugins in go-go-gepa and propagating streaming events out of plugin execution paths.
LastUpdated: 2026-02-28T14:27:55.294971719-05:00
WhatFor: ""
WhenToUse: ""
---


# Enable Promise-based JS plugin execution and streaming events

## Overview

`geppetto` JS already supports async execution (`runAsync`, `start`, `RunHandle.on`), but `go-go-gepa` plugin loaders currently expect immediate synchronous return values from plugin methods (`run`, `evaluate`, `generateOne`). This ticket defines and scopes the bridge work needed to support Promise-returning plugin methods and optionally forward streaming events into CLI/storage.

Current status: scoped with implementation plan and task breakdown, ready for execution.

## Key Links

- Planning doc: `planning/01-promise-aware-plugin-bridge-and-streaming-events-implementation-plan.md`
- Diary: `reference/01-implementation-diary.md`
- Tasks: `tasks.md`
- Changelog: `changelog.md`

## Status

Current status: **active**

## Topics

- gepa
- plugins
- goja
- runner
- events
- js-bindings
- go
- tooling

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
