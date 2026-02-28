# Tasks

## Ticket setup

- [x] Create GEPA-26 ticket workspace and initial documents
- [x] Confirm scope: package split for HyperCard + runtime plugin systems

## Architecture investigation

- [x] Inventory current monorepo package structure (`engine`, `desktop-os`, `confirm-runtime`, app packages)
- [x] Map plugin runtime internals (`runtimeService`, `stack-bootstrap`, contracts, schema)
- [x] Trace runtime state reducers/selectors/capability policy (`pluginCardRuntime`)
- [x] Trace desktop-shell integration points (`PluginCardSessionHost`, `pluginIntentRouting`, default adapters)
- [x] Trace HyperCard artifact projection and runtime card injection path
- [x] Capture app-level consumption patterns in first-party apps and external arc-agi-player package

## Design deliverables

- [x] Write verbose intern onboarding architecture document from fundamentals to runtime details
- [x] Propose target package boundaries and dependency direction
- [x] Define phased migration plan, test strategy, rollback plan, and risk mitigations
- [x] Document explicit non-goals and alternatives considered

## Documentation operations

- [x] Maintain detailed chronological diary while investigating and writing
- [x] Relate high-signal code files to design doc and diary via `docmgr doc relate`
- [x] Update changelog with completion summary
- [x] Run `docmgr doctor --ticket GEPA-26-HYPERCARD-RUNTIME-SPLIT --stale-after 30`
- [x] Upload final bundle to reMarkable and verify remote listing

## Implementation hard-cut execution (no compatibility wrappers)

- [x] Create new `packages/hypercard-runtime` package and move runtime/hypercard source into it
- [x] Remove runtime/hypercard exports from `@hypercard/engine` and keep engine focused on desktop/shell/widgets
- [x] Move runtime host/renderer/intent routing out of engine windowing and expose from `@hypercard/hypercard-runtime`
- [x] Update `go-go-os` apps and package wiring to import runtime APIs from new package
- [x] Update `go-go-app-arc-agi-3` imports and package wiring to new runtime package
- [x] Update/relocate affected tests and stories to match new package boundaries
- [x] Run typecheck/tests for touched packages and document failures precisely if any
- [x] Update GEPA-26 diary and changelog per execution step with commit hashes
- [x] Upload updated GEPA-26 diary bundle to reMarkable after implementation completes
