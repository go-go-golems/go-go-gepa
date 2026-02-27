# Tasks

## Completed

- [x] Create ticket workspace `GEPA-09-REPO-SPLIT-ARCHITECTURE`
- [x] Map current backend host contracts and lifecycle behavior
- [x] Map current frontend launcher/module contracts and runtime assumptions
- [x] Identify split blockers and coupling points (build, embed, imports, routing)
- [x] Produce long-form architecture design document (10+ pages target)
- [x] Produce chronological research diary with command evidence
- [x] Produce v2 renamed topology design for `wesen-os` + `go-go-app-inventory`
- [x] Produce detailed execution task board for first-plan implementation

## Next (V2 execution sequence)

- [ ] Create repository `wesen-os` and bootstrap backend host core package
- [ ] Create repository `go-go-app-inventory` and extract inventory backend from `go-go-os/go-inventory-chat`
- [ ] Port `internal/inventorydb` and `internal/pinoweb` into `go-go-app-inventory`
- [ ] Define host-agnostic inventory backend `Component` API in `go-go-app-inventory`
- [ ] Add lifecycle, route, and reflection tests for `go-go-app-inventory`
- [ ] Port generic backend host package into `wesen-os` (`module/registry/lifecycle/routes/manifest`)
- [ ] Implement `wesen-os` inventory adapter over `go-go-app-inventory` component
- [ ] Implement `wesen-os` GEPA adapter over `go-go-gepa`
- [ ] Finalize frontend artifact ingestion path from `go-go-os` into `wesen-os`
- [ ] Add `wesen-os` launcher smoke tests for `/api/os/apps` and namespaced routes
- [ ] Pin dependency versions in `wesen-os` release metadata (`go-go-os`, `go-go-gepa`, `go-go-app-inventory`)
- [ ] Cut over docs and naming to `wesen-os` and `go-go-app-inventory` with no compatibility aliases

## Plan Reference

- `design-doc/02-v2-wesen-os-composition-plan-go-go-os-go-go-gepa-go-go-app-inventory.md`
