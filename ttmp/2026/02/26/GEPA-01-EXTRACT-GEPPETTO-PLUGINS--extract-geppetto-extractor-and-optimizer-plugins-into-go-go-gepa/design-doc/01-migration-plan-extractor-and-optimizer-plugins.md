---
Title: 'Migration plan: extractor and optimizer plugins'
Ticket: GEPA-01-EXTRACT-GEPPETTO-PLUGINS
Status: active
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
      Note: geppetto module registration no longer includes geppetto/plugins
    - Path: geppetto/pkg/js/modules/geppetto/module_test.go
      Note: hard-cut regression test asserts require("geppetto/plugins") fails
    - Path: geppetto/pkg/doc/topics/13-js-api-reference.md
      Note: JS API reference updated to state plugin helpers are no longer exported by geppetto
    - Path: geppetto/pkg/doc/topics/14-js-api-user-guide.md
      Note: JS guide updated to remove geppetto/plugins authoring guidance
    - Path: go-go-gepa/cmd/gepa-runner/scripts/lib/gepa_plugin_contract.js
      Note: retained local optimizer contract helper (explicitly kept)
    - Path: 2026-02-18--cozodb-extraction/cozo-relationship-js-runner/scripts/relation_extractor_template.js
      Note: extractor script migrated off geppetto/plugins helper import
    - Path: 2026-02-18--cozodb-extraction/cozo-relationship-js-runner/scripts/relation_extractor_reflective.js
      Note: extractor script migrated off geppetto/plugins helper import
ExternalSources: []
Summary: Hard-cut migration plan and implementation notes for removing geppetto/plugins with no compatibility alias, while keeping local GEPA optimizer helper usage.
LastUpdated: 2026-02-26T12:34:00-05:00
WhatFor: Document the no-alias migration and remaining follow-up work (registry identifier propagation).
WhenToUse: Use when implementing remaining plugin metadata work or validating that legacy plugin module imports are gone.
---

# Migration plan: extractor and optimizer plugins

## Executive Summary

This ticket now follows a hard-cut policy:

1. `geppetto` no longer registers `require("geppetto/plugins")`.
2. No compatibility alias is provided.
3. `go-go-gepa` keeps local optimizer helper usage via `./lib/gepa_plugin_contract`.
4. Extractor scripts that used `geppetto/plugins` were migrated to plain descriptor exports.

The remaining follow-up scope in this ticket is registry identifier carriage (`registryIdentifier`) through loader/reporting/recording surfaces.

## Problem Statement and Scope

### Problem addressed

`geppetto/plugins` mixed application plugin-contract behavior into a core framework module. That coupling created:

1. unclear ownership boundaries,
2. duplicated validation logic across repos,
3. brittle import behavior for runners that depended on helper registration details.

### Scope in this implementation pass

Completed in this pass:

1. remove plugin helper module registration from `geppetto`,
2. delete `plugins_module.go` from geppetto,
3. update geppetto tests/docs to reflect removal,
4. migrate extractor scripts away from `geppetto/plugins`.

Out of scope in this pass:

1. replacing `require("geppetto")` core runtime APIs,
2. changing optimizer algorithm internals,
3. implementing full registry identifier persistence changes.

## Current-State Architecture (Post-Change)

### 1) Geppetto core module surface

Observed state:

1. `pkg/js/modules/geppetto/module.go` now registers only `ModuleName = "geppetto"`.
2. `pkg/js/modules/geppetto/plugins_module.go` is removed.
3. `pkg/js/modules/geppetto/module_test.go` now includes a regression assertion that `require("geppetto/plugins")` fails.

Consequence:

- plugin helper imports from geppetto are a hard error by design.

### 2) Optimizer contract ownership in GEPA runtime

Observed state:

1. `go-go-gepa` keeps local helper `cmd/gepa-runner/scripts/lib/gepa_plugin_contract.js`.
2. bundled optimizer scripts in `go-go-gepa` already use that local helper path.

Consequence:

- optimizer plugin authoring in `go-go-gepa` is decoupled from geppetto plugin helper exports.

### 3) Extractor scripts migrated off geppetto/plugins

Observed state:

1. `cozo-relationship-js-runner` scripts now export plain descriptor objects with explicit `apiVersion` + `kind`.
2. helper import `require("geppetto/plugins")` was removed from those scripts.

Consequence:

- extractor script loading no longer depends on geppetto plugin helper module.

## Design Decisions

1. Hard cut, no alias:
- rejected any temporary `geppetto/plugins` compatibility alias.

2. Keep `gepa_plugin_contract.js` (optimizer helper) in go-go-gepa:
- preserves ergonomic local JS descriptor validation without reintroducing geppetto coupling.

3. Prefer plain descriptor exports for extractor scripts:
- simplest migration path where host loaders already validate descriptor schema.

## Testing and Validation

Completed validation:

1. `go test ./pkg/js/modules/geppetto -count=1` (geppetto) passed.
2. geppetto pre-commit hook executed full test/lint successfully during commit.
3. static grep check across targeted repos confirms no runtime `require("geppetto/plugins")` usage remains in maintained code paths (excluding explicit negative-test/docs mention).

Validation blocker captured:

1. `cozo-relationship-js-runner` local `go test` is currently blocked in this environment by missing `go.sum` entries when run with `GOWORK=off`.
2. This is a repo environment/dependency state issue, not a compile error from the script change itself.

## Risks and Follow-ups

### Residual risks

1. External downstream scripts outside this workspace may still import `geppetto/plugins` and now fail.
2. Registry provenance is still incomplete until `registryIdentifier` is wired through loaders/recorders.

### Remaining follow-ups

1. Add `registryIdentifier` to optimizer + extractor metadata structs.
2. Propagate to host context/tags/report JSON.
3. Extend recorder schema with `plugin_registry_identifier`.
4. Add integration tests asserting presence of new metadata field.

## Implementation Record (Commits)

1. `geppetto` commit `d102477`:
- removed module registration + deleted plugins module + added hard-cut regression test.

2. `geppetto` commit `a9c2e61`:
- updated JS docs to remove plugin helper API claims.

3. `cozo-relationship-js-runner` commit `b694000`:
- removed `geppetto/plugins` imports from extractor scripts and switched to explicit descriptor export.

## References

1. `geppetto/pkg/js/modules/geppetto/module.go`
2. `geppetto/pkg/js/modules/geppetto/module_test.go`
3. `geppetto/pkg/doc/topics/13-js-api-reference.md`
4. `geppetto/pkg/doc/topics/14-js-api-user-guide.md`
5. `go-go-gepa/cmd/gepa-runner/scripts/lib/gepa_plugin_contract.js`
6. `2026-02-18--cozodb-extraction/cozo-relationship-js-runner/scripts/relation_extractor_template.js`
7. `2026-02-18--cozodb-extraction/cozo-relationship-js-runner/scripts/relation_extractor_reflective.js`
