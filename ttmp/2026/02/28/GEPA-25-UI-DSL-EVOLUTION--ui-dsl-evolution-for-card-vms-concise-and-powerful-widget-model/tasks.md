# Tasks

## Ticket setup

- [x] Create GEPA-25 ticket workspace with design-doc and diary docs
- [x] Confirm scope: improve card-VM UI DSL, review existing widgets, propose elegant/powerful direction

## Architecture and evidence mapping

- [x] Trace VM-exposed DSL source (`stack-bootstrap.vm.js`) and node schema contracts
- [x] Trace runtime rendering pipeline (`PluginCardSessionHost` + `PluginCardRenderer`)
- [x] Inventory current DSL node kinds, events, and validation behavior
- [x] Inventory engine widget surface (`components/widgets`) and identify recently added widgets
- [x] Trace confirm-runtime composition patterns that already leverage newer widgets
- [x] Capture concrete file/line evidence for core claims and constraints

## Analysis and proposal

- [x] Document current-state fundamentals for intern onboarding
- [x] Identify gaps/risks in current DSL (contract drift, styling mismatch, capability boundaries)
- [x] Propose concise/elegant next-version DSL shape (core + selected compound widgets)
- [x] Provide API sketches and pseudocode for host/runtime/renderer changes
- [x] Provide phased implementation and migration plan
- [x] Provide testing/validation strategy and rollout guardrails

## Delivery and publishing

- [x] Maintain chronological investigation diary with commands/findings/decisions
- [x] Relate key files to ticket docs via `docmgr doc relate`
- [x] Update ticket changelog with completion summary
- [x] Run `docmgr doctor --ticket GEPA-25-UI-DSL-EVOLUTION --stale-after 30`
- [x] Upload final bundled docs to reMarkable (dry-run + real upload + remote listing verification)

## Scope correction follow-up (2026-02-28)

- [x] Remove `ui.counter` from proposal and runtime contract
- [x] Remove `schemaForm`, `rating`, `imageChoice`, and `actionBar` from proposed VM DSL widget surface
- [x] Update intern-facing design doc to reflect constrained elegant widget set
- [x] Implement runtime code changes for `counter` removal (`uiTypes`, `uiSchema`, `PluginCardRenderer`, tests)
- [x] Update diary and changelog with per-step commits
- [x] Re-run validation (`engine` tests + `docmgr doctor`) and refresh reMarkable upload
