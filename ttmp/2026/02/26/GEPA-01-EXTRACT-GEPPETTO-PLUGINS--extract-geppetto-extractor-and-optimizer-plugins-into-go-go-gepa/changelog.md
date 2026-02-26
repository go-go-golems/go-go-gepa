# Changelog

## 2026-02-26

- Created ticket workspace for `GEPA-01-EXTRACT-GEPPETTO-PLUGINS`.
- Added design doc `design-doc/01-migration-plan-extractor-and-optimizer-plugins.md` with:
  - current-state architecture mapping,
  - evidence-backed gap analysis,
  - proposed target ownership in `go-go-gepa`,
  - phased rollout plan,
  - registry identifier propagation strategy.
- Added reference diary `reference/01-investigation-diary.md` with chronological commands/findings/interpretations.
- Updated index and tasks with current completion status and implementation backlog.

## 2026-02-26

Completed architecture analysis and migration design for moving extractor/optimizer plugin contracts from geppetto to go-go-gepa, including phased compatibility alias and registryIdentifier carriage plan.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/geppetto/pkg/js/modules/geppetto/plugins_module.go — source of contract logic targeted for extraction
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/run_recorder.go — recorder schema impacted by registry identifier carry requirement
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-01-EXTRACT-GEPPETTO-PLUGINS--extract-geppetto-extractor-and-optimizer-plugins-into-go-go-gepa/design-doc/01-migration-plan-extractor-and-optimizer-plugins.md — primary design deliverable


## 2026-02-26

Recorded full investigation diary with command-by-command evidence, call-site inventory, and documentation blast radius for geppetto/plugins consumers.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/2026-02-18--cozodb-extraction/cozo-relationship-js-runner/scripts/relation_extractor_template.js — extractor script currently imports geppetto/plugins
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/geppetto/pkg/doc/topics/14-js-api-user-guide.md — public docs currently point to geppetto/plugins
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-01-EXTRACT-GEPPETTO-PLUGINS--extract-geppetto-extractor-and-optimizer-plugins-into-go-go-gepa/reference/01-investigation-diary.md — chronological evidence log


## 2026-02-26

Resolved doc vocabulary warnings by adding topic slugs (extractor/gepa/optimizer/plugins) and confirmed doctor clean status.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-01-EXTRACT-GEPPETTO-PLUGINS--extract-geppetto-extractor-and-optimizer-plugins-into-go-go-gepa/index.md — ticket topics now validate against vocabulary
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/vocabulary.yaml — added topic vocabulary entries required by doctor


## 2026-02-26

Uploaded a bundled PDF deliverable to reMarkable after successful dry-run and remote listing verification.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-01-EXTRACT-GEPPETTO-PLUGINS--extract-geppetto-extractor-and-optimizer-plugins-into-go-go-gepa/design-doc/01-migration-plan-extractor-and-optimizer-plugins.md — included in uploaded bundle
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-01-EXTRACT-GEPPETTO-PLUGINS--extract-geppetto-extractor-and-optimizer-plugins-into-go-go-gepa/reference/01-investigation-diary.md — included in uploaded bundle


## 2026-02-26 (Hard-Cut Implementation)

- Implemented hard-cut removal of `geppetto/plugins` with no compatibility alias.
- Kept local optimizer helper `go-go-gepa/cmd/gepa-runner/scripts/lib/gepa_plugin_contract.js` as requested.
- Migrated extractor scripts in `cozo-relationship-js-runner` off `geppetto/plugins` imports.
- Added/kept regression behavior in geppetto tests to assert `require("geppetto/plugins")` fails.

### Commits

- `geppetto`: `d102477` — remove `geppetto/plugins` registration/helpers and enforce hard-cut in tests.
- `geppetto`: `a9c2e61` — update JS docs to remove plugin-helper API claims.
- `cozo-relationship-js-runner`: `b694000` — remove `geppetto/plugins` imports from extractor scripts.

### Validation

- `go test ./pkg/js/modules/geppetto -count=1` passed.
- geppetto pre-commit hooks ran full test/lint successfully on commit.
- static grep check shows no remaining runtime `require("geppetto/plugins")` usage in maintained code paths.
- `cozo-relationship-js-runner` `go test` is currently blocked by missing `go.sum` entries in this environment.

## 2026-02-26

Executed hard-cut migration with no geppetto/plugins alias: removed core module registration, migrated extractor scripts, and updated ticket docs to final state.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/2026-02-18--cozodb-extraction/cozo-relationship-js-runner/scripts/relation_extractor_template.js — migrated off geppetto/plugins helper import
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/geppetto/pkg/js/modules/geppetto/module.go — removed plugin module registration
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/geppetto/pkg/js/modules/geppetto/module_test.go — added regression assertion that geppetto/plugins is unavailable
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-01-EXTRACT-GEPPETTO-PLUGINS--extract-geppetto-extractor-and-optimizer-plugins-into-go-go-gepa/design-doc/01-migration-plan-extractor-and-optimizer-plugins.md — final no-alias architecture record

## 2026-02-26

Reverted GEPA-01 planning direction back to go-go-gepa implementation work. Updated design/index/tasks to make remaining work explicit: implement go-go-gepa plugin module ownership and carry `registryIdentifier` through loader/runtime/reporting/sqlite. Also codified scope guardrails that `gepa/` and `2026-02-18--cozodb-extraction/` are reference-only.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-01-EXTRACT-GEPPETTO-PLUGINS--extract-geppetto-extractor-and-optimizer-plugins-into-go-go-gepa/design-doc/01-migration-plan-extractor-and-optimizer-plugins.md — reinstated go-go-gepa implementation plan
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-01-EXTRACT-GEPPETTO-PLUGINS--extract-geppetto-extractor-and-optimizer-plugins-into-go-go-gepa/tasks.md — updated action checklist for immediate implementation
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-01-EXTRACT-GEPPETTO-PLUGINS--extract-geppetto-extractor-and-optimizer-plugins-into-go-go-gepa/index.md — updated ticket scope and status
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/AGENT.md — added scope boundaries (`gepa/` and `2026-02-18--cozodb-extraction/` reference-only)

## 2026-02-26 (Build pass complete in go-go-gepa)

- Added native plugin module ownership in `go-go-gepa` via `require("gepa/plugins")`.
- Added `registryIdentifier` carriage through loader metadata, host context, optimize/eval tags, report JSON payloads, and sqlite recorder rows.
- Added recorder schema migration for `plugin_registry_identifier` and report query support (`id@registry` grouping/printing).
- Added/updated tests for:
  - descriptor default/explicit registry decode,
  - host context injection,
  - recorder persistence + legacy schema migration,
  - eval report registry visibility.
- Fixed decode bug where missing JS descriptor field could become literal `"undefined"` instead of defaulting.
- Migrated runner example scripts to import `require("gepa/plugins")`.

### Validation

- `go test ./cmd/gepa-runner -count=1` passed.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/gepa_plugins_module.go — new go-go-gepa-owned plugin contract module
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/plugin_loader.go — registry metadata decode/default and host context injection
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/main.go — optimize flow tag/report propagation
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/eval_command.go — eval flow tag/report propagation
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/run_recorder.go — sqlite schema/persistence migration
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/eval_report.go — report query/format changes for registry identifier
