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
