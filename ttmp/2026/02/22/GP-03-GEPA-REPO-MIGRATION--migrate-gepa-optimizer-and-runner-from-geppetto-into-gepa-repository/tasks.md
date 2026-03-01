# Tasks

## Phase 1: Repository Split and Ownership Transfer

### A. Ticket setup and planning

- [x] Create ticket `GP-03-GEPA-REPO-MIGRATION` in `gepa/ttmp`.
- [x] Add migration implementation plan document under `analysis/`.
- [x] Add migration diary document under `reference/`.
- [x] Add file relationships for plan/diary docs using `docmgr doc relate`.

### B. Move GEPA docs and workspace root

- [x] Move existing GEPA tickets from `geppetto/ttmp` to `gepa/ttmp`.
- [x] Update docmgr root config to `gepa/ttmp`.
- [x] Verify moved ticket metadata and links remain coherent.

### C. Extract code into `gepa/go-gepa-runner`

- [x] Scaffold `go-gepa-runner` with go-go-golems project setup templates.
- [x] Copy `pkg/optimizer/gepa` into new repository ownership path.
- [x] Copy `cmd/gepa-runner` into new repository ownership path.
- [x] Rewrite imports to `github.com/gepa-ai/gepa/go-gepa-runner/...`.
- [x] Run `go mod tidy` in `go-gepa-runner`.
- [x] Validate `go test ./... -count=1` in `go-gepa-runner`.
- [x] Validate `go build ./cmd/gepa-runner` in `go-gepa-runner`.

### D. Clean geppetto to generic-only behavior

- [x] Remove `geppetto/cmd/gepa-runner` from geppetto tree.
- [x] Remove `geppetto/pkg/optimizer/gepa` from geppetto tree.
- [x] Keep generic optimizer plugin helpers in `geppetto/plugins` module.
- [x] Update geppetto docs to reference external GEPA runner script location.
- [x] Re-run geppetto tests to confirm no regressions after cleanup.
- [x] Commit geppetto migration cleanup.

### E. Harden `go-gepa-runner` standalone ergonomics

- [x] Replace scaffold README TODO with real optimize/eval usage and architecture notes.
- [x] Add migration verification artifacts under new ticket `sources/` (test/build outputs).
- [x] Commit extracted GEPA module and moved tickets in `gepa` repo.

### F. Delivery and publication

- [x] Upload migration plan document to reMarkable.
- [x] Record upload path/receipt in ticket changelog.
- [x] Write diary entries for each completed task batch with commit hashes.
- [x] Mark all remaining tasks complete.
