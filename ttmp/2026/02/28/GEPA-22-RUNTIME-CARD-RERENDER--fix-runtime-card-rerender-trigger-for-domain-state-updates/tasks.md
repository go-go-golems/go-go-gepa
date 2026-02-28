# Tasks

## Planning Complete, Implementation Not Started

- [ ] 0. Confirm start authorization
- [ ] 0.1 Reconfirm with owner that implementation should start now
- [ ] 0.2 Add kickoff note in changelog with date and owner

- [ ] 1. Reproduce and baseline the bug
- [ ] 1.1 Reproduce domain-only update stale-render case in launcher runtime
- [ ] 1.2 Add temporary logging around `renderCard` invocation count
- [ ] 1.3 Capture baseline perf and rerender traces

- [ ] 2. Selector/projection design
- [ ] 2.1 Define `selectProjectedDomainsForRuntimeCard` selector contract
- [ ] 2.2 Implement stable identity strategy for projection output
- [ ] 2.3 Add unit tests for projection stability and change semantics

- [ ] 3. PluginCardSessionHost integration
- [ ] 3.1 Wire projection selector into `PluginCardSessionHost`
- [ ] 3.2 Compute projection fingerprint or stable reference dependency
- [ ] 3.3 Add dependency to `tree` useMemo so domain updates can invalidate render cache
- [ ] 3.4 Keep existing card/session/nav/runtime triggers intact

- [ ] 4. Correctness tests
- [ ] 4.1 Add regression test: domain-only relevant update causes rerender
- [ ] 4.2 Add negative test: unrelated domain update does not rerender when excluded from projection
- [ ] 4.3 Add integration test with runtime card consuming projected domain state

- [ ] 5. Performance hardening
- [ ] 5.1 Compare rerender counts before/after fix
- [ ] 5.2 Optimize projection/hashing path if rerender cost is high
- [ ] 5.3 Remove temporary instrumentation once metrics are captured

- [ ] 6. Docs and handoff
- [ ] 6.1 Update GEPA-14 docs with implemented rerender fix details
- [ ] 6.2 Update intern handoff reference with final code path
- [ ] 6.3 Record final validation commands and results in changelog

- [ ] 7. Release readiness
- [ ] 7.1 Run typechecks (`go-go-os`, `go-go-app-inventory`, `wesen-os`)
- [ ] 7.2 Run targeted runtime and launcher tests
- [ ] 7.3 Perform manual smoke with stock/domain updates and verify card rerender
- [ ] 7.4 Close GEPA-22 with commit/test links
