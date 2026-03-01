---
Title: 'Migration plan: extractor and optimizer plugins'
Ticket: GEPA-01-EXTRACT-GEPPETTO-PLUGINS
Status: complete
Topics:
    - architecture
    - plugins
    - extractor
    - optimizer
    - gepa
    - geppetto
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: geppetto/pkg/js/modules/geppetto/module.go
      Note: hard-cut removed geppetto/plugins from core module registration
    - Path: go-go-gepa/cmd/gepa-runner/js_runtime.go
      Note: runtime wiring point where new go-go-gepa plugin module should be registered
    - Path: go-go-gepa/cmd/gepa-runner/plugin_loader.go
      Note: optimizer metadata decode and registryIdentifier propagation point
    - Path: go-go-gepa/cmd/gepa-runner/main.go
      Note: optimize flow metadata/report propagation
    - Path: go-go-gepa/cmd/gepa-runner/eval_command.go
      Note: eval flow metadata/report propagation
    - Path: go-go-gepa/cmd/gepa-runner/run_recorder.go
      Note: sqlite schema and persistence for plugin metadata
    - Path: go-go-gepa/cmd/gepa-runner/scripts/lib/gepa_plugin_contract.js
      Note: current local optimizer contract helper retained in go-go-gepa
ExternalSources: []
Summary: Implemented go-go-gepa ownership of optimizer plugin contract module plus end-to-end registryIdentifier carriage.
LastUpdated: 2026-02-26T13:40:00-05:00
WhatFor: Capture GEPA-01 implementation decisions and final technical state.
WhenToUse: Use when maintaining plugin metadata flow and GEPA runner reporting/storage.
---

# Migration Plan: Extractor and Optimizer Plugins

## Executive Summary

GEPA-01 implementation is complete in `go-go-gepa`:
1. Plugin contract ownership was moved into a new native JS module `require("gepa/plugins")`.
2. `registryIdentifier` is now carried end-to-end from descriptor decode through host context, hook tags, CLI/report JSON, and sqlite persistence.
3. No compatibility alias was introduced.

## Final State

1. `cmd/gepa-runner/gepa_plugins_module.go` registers `gepa/plugins` and exports:
   - `OPTIMIZER_PLUGIN_API_VERSION`
   - `defineOptimizerPlugin(...)`
2. `optimizerPluginMeta` includes `RegistryIdentifier` and defaults to `local`.
3. `pluginRegistryIdentifier` is injected into plugin `create(hostContext)` and propagated in hook tags (`initialCandidate`, `evaluate`, `merge`, `selectComponents`, `componentSideInfo`).
4. Recorder schema stores `plugin_registry_identifier` in `gepa_runs` with additive migration support for existing DBs.
5. Eval/report outputs include `registryIdentifier` in printed/json plugin metadata.
6. Runner sample scripts now import `require("gepa/plugins")`.

## Implementation Notes

### Module ownership

1. Added native module file `cmd/gepa-runner/gepa_plugins_module.go`.
2. Registered module in `cmd/gepa-runner/js_runtime.go`.
3. Migrated runner example scripts to import `gepa/plugins`.

### Metadata decode and propagation

1. Extended `optimizerPluginMeta` and decode path in `cmd/gepa-runner/plugin_loader.go`.
2. Added robust defaulting for missing JS values (avoid accidental `"undefined"` string propagation).
3. Added plugin registry tags in optimize/eval command flows.

### Storage and reporting

1. Added `plugin_registry_identifier` to `runRecorderConfig`, `runRecord`, insert SQL, schema create SQL, and migration logic.
2. Updated eval report row/summary queries and table rendering to include registry identifier.
3. Updated `--out-report` payload generation in optimize/eval commands.

### Tests

1. Added plugin loader tests for:
   - default registry identifier,
   - explicit registry identifier,
   - host context injection.
2. Added recorder tests for:
   - persisted `plugin_registry_identifier`,
   - legacy schema migration adding the new column.
3. Extended eval report tests to assert registry identifier visibility.

## Decisions and Constraints

1. No compatibility alias for `geppetto/plugins`.
2. `gepa/` and `2026-02-18--cozodb-extraction/` are reference-only for this workstream.
3. Focus implementation in `go-go-gepa` plus GEPA-01 ticket docs.

## Residual Risks

1. External plugin scripts that do not use `defineOptimizerPlugin` may still set malformed descriptor fields.
2. Existing sqlite rows created before migration can still contain null registry values (report path defaults these to `local`).

## Validation Checklist

1. `go test ./cmd/gepa-runner -count=1` passed.
2. Unit tests cover descriptor default/explicit registry behavior.
3. Unit tests cover sqlite migration and report query visibility for `plugin_registry_identifier`.

## References

1. `go-go-gepa/cmd/gepa-runner/js_runtime.go`
2. `go-go-gepa/cmd/gepa-runner/plugin_loader.go`
3. `go-go-gepa/cmd/gepa-runner/main.go`
4. `go-go-gepa/cmd/gepa-runner/eval_command.go`
5. `go-go-gepa/cmd/gepa-runner/run_recorder.go`
6. `go-go-gepa/cmd/gepa-runner/scripts/lib/gepa_plugin_contract.js`
