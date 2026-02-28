# Tasks

## Done

- [x] Create GEPA-07 ticket workspace and baseline docs.
- [x] Re-baseline `go-go-os` checkout and confirm full source availability.
- [x] Map current backend module host contracts (`AppBackendModule`, registry, lifecycle, namespaced route guard).
- [x] Map launcher frontend module contracts and composition path.
- [x] Map timeline/event ingestion path and debug windows reusable for GEPA.
- [x] Map go-go-gepa script runtime, plugin loader, async call bridge, and stream event contracts.
- [x] Produce detailed pre-research starting-point design doc for team onboarding.
- [x] Record chronological investigation diary for reproducibility.
- [x] Run `docmgr doctor` validation for GEPA-07 ticket docs.
- [x] Upload pre-research bundle to reMarkable and verify cloud listing.

## Next

- [ ] Phase 0 spike: define and validate GEPA event -> SEM/timeline mapping with a minimal mock run.
- [ ] Define GEPA backend module API schema (`/scripts`, `/runs`, `/runs/{id}/events`, `/runs/{id}/timeline`).
- [ ] Implement initial in-repo `gepa` backend module and wire into backendhost registry.
- [ ] Implement initial `apps/gepa` launcher module (static include) using host API resolver pattern.
- [ ] Add smoke tests for GEPA namespaced routes and run lifecycle.
- [ ] Design process-based backend plugin manager protocol for out-of-repo modules.
- [ ] Evaluate and decide dynamic frontend plugin loading strategy (post-Phase 1).
