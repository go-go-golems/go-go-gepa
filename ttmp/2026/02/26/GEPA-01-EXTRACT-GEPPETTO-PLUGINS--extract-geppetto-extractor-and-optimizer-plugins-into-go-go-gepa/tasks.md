# Tasks

## Completed

- [x] Create ticket workspace and base docs for `GEPA-01-EXTRACT-GEPPETTO-PLUGINS`.
- [x] Produce evidence-backed architecture analysis for plugin ownership boundaries.
- [x] Remove `geppetto/plugins` from geppetto core (hard cut).
- [x] Keep local `go-go-gepa` optimizer helper contract file (`gepa_plugin_contract.js`).

## Reinstated Implementation Plan (go-go-gepa focused)

- [x] Add go-go-gepa native plugin module ownership (`gepa/plugins` in go-go-gepa runtime).
- [x] Extend plugin metadata decode with `registryIdentifier` in `plugin_loader.go`.
- [x] Propagate `registryIdentifier` into host context + hook tags in optimize/eval flows.
- [x] Add `plugin_registry_identifier` persistence to sqlite recorder schema and inserts.
- [x] Include `registryIdentifier` in run/eval report outputs.
- [x] Add tests for decode/default behavior + recorder migration + metadata visibility in report rows.
- [x] Update docs/examples in go-go-gepa ticket docs to reflect the above implementation.

## Scope Guardrails

- [x] `gepa/` marked reference-only.
- [x] `2026-02-18--cozodb-extraction/` marked reference-only.
- [x] Keep active implementation changes constrained to `go-go-gepa/` unless explicitly requested.
