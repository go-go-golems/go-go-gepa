# Tasks

## Completed

- [x] Create ticket workspace and base docs for `GEPA-01-EXTRACT-GEPPETTO-PLUGINS`.
- [x] Produce evidence-backed architecture analysis for plugin ownership boundaries.
- [x] Identify extractor + optimizer consumer blast radius (`geppetto`, `go-go-gepa`, legacy runners, extraction runner).
- [x] Define migration architecture with phased compatibility strategy.
- [x] Define how to carry a `registryIdentifier` in plugin metadata/reporting.
- [x] Maintain chronological investigation diary with command logs and findings.
- [x] Align implementation direction with user constraints: ignore `gepa/`, hard-cut legacy `geppetto/plugins`, no compatibility alias, keep `gepa_plugin_contract.js`.
- [x] Remove untracked files from `geppetto/` workspace as requested.

## Active Implementation Sprint

- [x] Task 1: Remove `geppetto/plugins` module registration/implementation from `geppetto` and delete helper tests tied to that module.
- [x] Task 2: Update `geppetto` docs to remove plugin-helper API claims and describe hard-cut behavior.
- [x] Task 3: Migrate extractor scripts in `2026-02-18--cozodb-extraction/cozo-relationship-js-runner` off `require("geppetto/plugins")` while preserving descriptor compatibility.
- [x] Task 4: Add hard-cut regression checks/tests so `geppetto/plugins` does not silently return.
- [x] Task 5: Verify builds/tests in touched repos and commit per task (small, reviewable commits). (`cozo-relationship-js-runner` Go test is environment-blocked by missing go.sum entries; captured in diary)
- [x] Task 6: Update GEPA-01 design/tasks/changelog/diary to reflect final no-alias implementation state.
