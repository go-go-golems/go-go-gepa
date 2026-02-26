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
    - Path: 2026-02-18--cozodb-extraction/cozo-relationship-js-runner/plugin_loader.go
      Note: extractor descriptor decode and metadata export
    - Path: geppetto/pkg/js/modules/geppetto/module.go
      Note: geppetto registers geppetto/plugins module
    - Path: geppetto/pkg/js/modules/geppetto/plugins_module.go
      Note: current extractor and optimizer plugin contract implementation in core framework
    - Path: go-go-gepa/cmd/gepa-runner/js_runtime.go
      Note: go-go-gepa runtime wiring and dependency on geppetto registration
    - Path: go-go-gepa/cmd/gepa-runner/plugin_loader.go
      Note: optimizer descriptor decode and metadata shape
    - Path: go-go-gepa/cmd/gepa-runner/run_recorder.go
      Note: plugin metadata persistence schema lacks registry identifier
ExternalSources: []
Summary: Evidence-based migration plan to move plugin contracts out of geppetto and into go-go-gepa, with registry identifier propagation.
LastUpdated: 2026-02-26T11:40:46-05:00
WhatFor: Guide the extraction of extractor/optimizer plugin contracts and runtime wiring from geppetto into go-go-gepa.
WhenToUse: Use when implementing or reviewing plugin ownership, runtime registration, and plugin metadata propagation changes.
---


# Migration plan: extractor and optimizer plugins

## Executive Summary

`geppetto` currently owns `require("geppetto/plugins")` and exports both extractor and optimizer plugin contract helpers, but this is not core inference framework behavior. The plugin contracts are runtime-specific app concerns (GEPA optimizer + extraction runners) and should be owned by `go-go-gepa`.

This plan moves plugin contract registration/validation into `go-go-gepa` and keeps `geppetto` focused on core session/profile/tool runtime APIs. The migration also introduces explicit `registryIdentifier` propagation so plugin registry provenance is carried in runtime metadata, reports, and recorder storage.

## Problem Statement and Scope

### Problem

1. Ownership mismatch:
- `geppetto` exports plugin descriptor contracts (`defineExtractorPlugin`, `defineOptimizerPlugin`) via `geppetto/plugins` even though contract consumers are GEPA/extractor runners, not core framework internals.

2. Coupling and duplication:
- `go-go-gepa` already duplicates optimizer contract validation in `cmd/gepa-runner/scripts/lib/gepa_plugin_contract.js` and again in Go loader decode logic.
- extractor runners duplicate canonicalization and metadata shaping.

3. Missing registry identifier carriage:
- plugin metadata currently carries `id`, `name`, `apiVersion`, `kind`, but no explicit registry identifier across optimize/eval/extract telemetry and outputs.

### Scope

In scope:
1. Move plugin contract module ownership to `go-go-gepa`.
2. Define compatibility strategy for existing `require("geppetto/plugins")` scripts.
3. Introduce and carry `registryIdentifier` end-to-end.
4. Define phased implementation and tests.

Out of scope:
1. Replacing `require("geppetto")` inference/session APIs.
2. Rewriting optimizer algorithm internals.
3. Reworking profile registry semantics.

## Current-State Architecture (Evidence)

### 1. `geppetto` currently registers plugin helper module

Evidence:
1. `geppetto/pkg/js/modules/geppetto/module.go:27` defines `PluginsModuleName = ModuleName + "/plugins"`.
2. `geppetto/pkg/js/modules/geppetto/module.go:58` registers `reg.RegisterNativeModule(PluginsModuleName, mod.pluginsLoader)`.

Implication:
- Any runtime that registers `geppetto` automatically carries plugin helper contracts, even if plugin contracts are unrelated to that host.

### 2. `geppetto/plugins` contains both extractor + optimizer descriptor logic

Evidence:
1. `geppetto/pkg/js/modules/geppetto/plugins_module.go:11-12` defines extractor and optimizer API version constants.
2. `geppetto/pkg/js/modules/geppetto/plugins_module.go:23-68` defines `defineExtractorPlugin`.
3. `geppetto/pkg/js/modules/geppetto/plugins_module.go:70-85` defines `wrapExtractorRun` and canonicalizes run input.
4. `geppetto/pkg/js/modules/geppetto/plugins_module.go:89-134` defines `defineOptimizerPlugin`.

Implication:
- Core framework package carries application-level plugin contracts.

### 3. `go-go-gepa` already has optimizer contract duplication

Evidence:
1. `go-go-gepa/cmd/gepa-runner/scripts/lib/gepa_plugin_contract.js:1-49` reimplements optimizer contract helper.
2. `go-go-gepa/cmd/gepa-runner/plugin_loader.go:145-179` revalidates descriptor metadata in Go.

Implication:
- Contract rules live in multiple places and can drift.

### 4. Extractor runner still depends on `geppetto/plugins`

Evidence:
1. `2026-02-18--cozodb-extraction/cozo-relationship-js-runner/scripts/relation_extractor_template.js:1` requires `geppetto/plugins`.
2. `2026-02-18--cozodb-extraction/cozo-relationship-js-runner/scripts/relation_extractor_reflective.js:1` requires `geppetto/plugins`.
3. `2026-02-18--cozodb-extraction/cozo-relationship-js-runner/main.go:433` registers `gp.Register(reg, gpOptions)` to make geppetto modules available.

Implication:
- Extractor contract ownership also currently anchored in geppetto.

### 5. Current metadata carries plugin ID but not registry identifier

Evidence:
1. Optimize/eval recorder persists `plugin_id` and `plugin_name` (`go-go-gepa/cmd/gepa-runner/run_recorder.go:301-303`, schema at `:446-447`).
2. Optimize/eval attach plugin id/name from descriptor (`go-go-gepa/cmd/gepa-runner/main.go:328-330`, `go-go-gepa/cmd/gepa-runner/eval_command.go:156-158`).
3. Extractor metadata map includes `plugin_id`/`plugin_name` but no registry identifier (`2026-02-18--cozodb-extraction/cozo-relationship-js-runner/plugin_loader.go:24-29`).

Implication:
- We cannot reliably trace plugin resolution provenance across registries.

## Gap Analysis

1. Architectural gap:
- Plugin contracts sit in the core module (`geppetto`), violating separation of concerns.

2. Runtime compatibility gap:
- Existing scripts depend on `require("geppetto/plugins")`; abrupt removal would break runners.

3. Consistency gap:
- Optimizer rules exist in JS contract helper + loader decode paths; extractor run canonicalization is duplicated between geppetto helper and extractor loader.

4. Observability gap:
- No explicit registry identifier carried across context/options/reporting/recorder outputs.

## Proposed Solution

### Target ownership

1. Create plugin contract module in `go-go-gepa`, for example:
- `go-go-gepa/pkg/js/modules/gepaplugins/module.go`
- exports both extractor + optimizer helpers.

2. Expose canonical module names from `go-go-gepa` registration helper:
- primary: `gepa/plugins`
- compatibility alias (phase-limited): `geppetto/plugins`

3. Keep `geppetto` focused on core runtime module:
- `require("geppetto")` remains for inference/sessions/profiles/tools.
- plugin contracts become a non-core dependency.

### Registry identifier carriage

Add a first-class field: `registryIdentifier`.

Contract behavior:
1. Descriptor optional field: `registryIdentifier?: string`.
2. If omitted, runtime injects default (CLI/config value) into metadata.
3. Loader metadata structs carry this field.

Propagation path:
1. Descriptor decode -> `meta.RegistryIdentifier`.
2. Host context (`create(ctx)`) includes `pluginRegistryIdentifier`.
3. Hook options tags include `registry_identifier`.
4. Recorder schema adds `plugin_registry_identifier`.
5. JSON reports/metadata include `plugin.registryIdentifier`.

### API sketch

```go
// go-go-gepa/pkg/js/modules/gepaplugins/register.go
const ModuleName = "gepa/plugins"
const LegacyModuleName = "geppetto/plugins" // temporary alias

func Register(reg *require.Registry, opts Options) {
    reg.RegisterNativeModule(ModuleName, loader)
    if opts.EnableLegacyAlias {
        reg.RegisterNativeModule(LegacyModuleName, loader)
    }
}
```

```js
// descriptor shape
module.exports = plugins.defineOptimizerPlugin({
  apiVersion: plugins.OPTIMIZER_PLUGIN_API_VERSION,
  kind: "optimizer",
  id: "my.optimizer",
  name: "My Optimizer",
  registryIdentifier: "team-alpha", // new
  create(ctx) { ... }
});
```

```go
// loader metadata
 type optimizerPluginMeta struct {
   APIVersion         string
   Kind               string
   ID                 string
   Name               string
   RegistryIdentifier string
 }
```

### Compatibility strategy

Phase-compatible import strategy:
1. New docs/examples point to `require("gepa/plugins")`.
2. Legacy scripts using `require("geppetto/plugins")` continue working through alias for one release window.
3. Deprecation warning emitted when legacy alias is loaded.
4. Alias removed in hard-cut release after migration completion.

## Pseudocode: Runtime Wiring

```text
newJSRuntime(scriptRoot):
  create goja VM + require registry
  register core geppetto module (require("geppetto"))
  register gepa plugins module (require("gepa/plugins"))
  optionally register compatibility alias (require("geppetto/plugins"))
  return runtime
```

```text
loadPlugin(script):
  descriptor = require(script)
  meta = decode(descriptor)
  if meta.registryIdentifier empty:
     meta.registryIdentifier = runtime default registry ID
  instance = descriptor.create(hostContext + registryIdentifier)
  return instance, meta
```

## Phased Implementation Plan

### Phase 1: Introduce `go-go-gepa` plugin module ownership

Files:
1. Add `go-go-gepa/pkg/js/modules/gepaplugins/*` for contract helpers.
2. Add tests analogous to current geppetto plugin helper tests.

Validation:
- unit tests for descriptor validation and extractor run normalization.

### Phase 2: Wire runtimes to new module and alias

Files:
1. `go-go-gepa/cmd/gepa-runner/js_runtime.go` register new module.
2. `2026-02-18--cozodb-extraction/cozo-relationship-js-runner/main.go` register new module.
3. Keep compatibility alias path enabled.

Validation:
- optimizer smoke scripts still pass.
- extractor scripts still load unchanged.

### Phase 3: Introduce registry identifier carriage

Files:
1. `go-go-gepa/cmd/gepa-runner/plugin_loader.go` add metadata field.
2. `go-go-gepa/cmd/gepa-runner/main.go` and `eval_command.go` propagate field.
3. `go-go-gepa/cmd/gepa-runner/run_recorder.go` schema + inserts + reports.
4. `2026-02-18--cozodb-extraction/cozo-relationship-js-runner/plugin_loader.go` and metadata output.

Validation:
- reports/DB rows include `plugin_registry_identifier`.
- backward compatibility for existing rows and descriptors.

### Phase 4: Documentation and deprecation rollout

Files:
1. update docs referencing `geppetto/plugins`:
   - `geppetto/pkg/doc/topics/13-js-api-reference.md`
   - `geppetto/pkg/doc/topics/14-js-api-user-guide.md`
   - runner readmes/scripts.
2. publish migration guidance and timeline.

Validation:
- no stale references to old module in active docs/examples (except explicit compatibility section).

### Phase 5: Remove legacy alias from geppetto

Files:
1. remove plugin loader registration from geppetto module.
2. delete `geppetto/pkg/js/modules/geppetto/plugins_module.go` after migration window.

Validation:
- grep-based check for `require("geppetto/plugins")` in maintained paths.

## Testing and Validation Strategy

1. Unit tests:
- descriptor validation edge cases (required fields, bad kind/apiVersion).
- extractor run canonicalization defaults (`timeoutMs` and transcript checks).
- registry identifier defaulting rules.

2. Integration tests:
- optimize/eval with scripts requiring `gepa/plugins`.
- temporary compatibility test with `geppetto/plugins` alias.
- extractor run path including `--include-metadata` output verification.

3. Recorder/report tests:
- schema migration adds nullable `plugin_registry_identifier`.
- insert/select/report behavior for old and new rows.

4. Static checks:
- `rg -n 'require\("geppetto/plugins"\)'` for remaining call sites.

## Risks, Alternatives, and Open Questions

### Risks

1. Breaking scripts if alias removal is too early.
2. Drift between optimizer/extractor contract semantics if split ownership is partial.
3. SQLite migration bugs if recorder schema change is not backward-safe.

### Alternatives considered

1. Keep contracts in geppetto:
- rejected because plugin contracts are app/runtime specific, not core framework APIs.

2. Move only optimizer contract:
- rejected because extractor contract has same ownership issue and would preserve split-brain behavior.

3. Keep JS-only contract helpers without Go metadata changes:
- rejected because registry identifier requirement needs loader/report/DB carriage.

### Open questions

1. Canonical default value for `registryIdentifier`:
- static default (`"default"`) vs CLI/config-sourced identifier.
2. Deprecation window length for `geppetto/plugins` alias.
3. Whether extractor runner should be absorbed into `go-go-gepa` as first-class command package.

## References

1. `geppetto/pkg/js/modules/geppetto/module.go`
2. `geppetto/pkg/js/modules/geppetto/plugins_module.go`
3. `geppetto/pkg/js/modules/geppetto/module_test.go`
4. `go-go-gepa/cmd/gepa-runner/js_runtime.go`
5. `go-go-gepa/cmd/gepa-runner/plugin_loader.go`
6. `go-go-gepa/cmd/gepa-runner/main.go`
7. `go-go-gepa/cmd/gepa-runner/eval_command.go`
8. `go-go-gepa/cmd/gepa-runner/run_recorder.go`
9. `go-go-gepa/cmd/gepa-runner/scripts/lib/gepa_plugin_contract.js`
10. `2026-02-18--cozodb-extraction/cozo-relationship-js-runner/main.go`
11. `2026-02-18--cozodb-extraction/cozo-relationship-js-runner/plugin_loader.go`
12. `2026-02-18--cozodb-extraction/cozo-relationship-js-runner/scripts/relation_extractor_template.js`
13. `2026-02-18--cozodb-extraction/cozo-relationship-js-runner/scripts/relation_extractor_reflective.js`
14. `geppetto/pkg/doc/topics/13-js-api-reference.md`
15. `geppetto/pkg/doc/topics/14-js-api-user-guide.md`
