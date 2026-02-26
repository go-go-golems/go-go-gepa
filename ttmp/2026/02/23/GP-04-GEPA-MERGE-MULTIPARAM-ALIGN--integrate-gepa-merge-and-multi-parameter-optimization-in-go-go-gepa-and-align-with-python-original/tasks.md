# Tasks

## Phase 0: Research and Planning (Completed)

- [x] Create ticket `GP-04-GEPA-MERGE-MULTIPARAM-ALIGN`.
- [x] Add analysis and diary documents.
- [x] Capture upstream merge/multi-param patch artifact from `imported/geppetto-main` commit `7b488b9`.
- [x] Capture file-by-file diffs between upstream and `go-go-gepa` for optimizer/runner/plugin files.
- [x] Capture Python GEPA reference excerpts (`optimize_anything`, `api.optimize`, `merge proposer`, `engine merge flow`, `component selector`).
- [x] Write in-depth 10+ page alignment and integration study.
- [x] Add reproducible artifact collection script under ticket `scripts/`.
- [x] Upload study document to reMarkable and verify remote location.

## Phase 1: Core Merge + Multi-Param Port into go-go-gepa

- [x] Port config additions to `go-go-gepa/pkg/optimizer/gepa/config.go`:
- [x] `MergeProbability`, `MergeSystemPrompt`, `MergePromptTemplate`, `OptimizableKeys`, `ComponentSelector`.
- [x] Port optimizer structural additions to `go-go-gepa/pkg/optimizer/gepa/optimizer.go`:
- [x] `MergeInput`, `MergeFunc`, `SetMergeFunc`, dual-parent lineage metadata, component selector machinery.
- [x] Port merge prompt/template + reflector merge support to `go-go-gepa/pkg/optimizer/gepa/format.go` and `go-go-gepa/pkg/optimizer/gepa/reflector.go`.
- [x] Port runner flag surface and seed-candidate map support to `go-go-gepa/cmd/gepa-runner/main.go` and `go-go-gepa/cmd/gepa-runner/dataset.go`.
- [x] Port optional plugin merge callback support in `go-go-gepa/cmd/gepa-runner/plugin_loader.go`.
- [x] Update toy optimizer script to include merge callback example.

## Phase 2: Hardening and Test Coverage

- [x] Add/extend unit tests for config defaults and selector behavior.
- [x] Add optimizer tests for merge acceptance baseline (child vs max(parentA,parentB)).
- [x] Add optimizer tests for round-robin and all-components update modes.
- [x] Add plugin loader tests for merge callback detection/output decoding.
- [x] Add dataset loader tests for seed-candidate JSON/YAML parsing and error paths.
- [x] Add smoke test coverage that loads packaged example scripts and validates optional hook exposure.
- [x] Decouple example script contract helper from `require("geppetto/plugins")` for compatibility with released Geppetto versions.
- [x] Run `GOWORK=off GOTOOLCHAIN=go1.25.7 go test ./... -count=1` in `go-go-gepa`.
- [x] Run `GOWORK=off GOTOOLCHAIN=go1.25.7 make lint` in `go-go-gepa`.

## Phase 3: Python Alignment Enhancements (Optional but Recommended)

- [x] Design a merge scheduling policy closer to Python (`merges_due`, `last_iter_found_new_program`) without over-complicating current Go architecture.
- [x] Evaluate adding callback/event hooks for merge attempted/accepted/rejected observability.
- [x] Define JS plugin extension points inspired by `optimize_anything` adapter patterns (component-side side_info, optional component selection hooks).
- [x] Decide whether to add seedless initialization mode in Go runner.

## Phase 4: Delivery

- [x] Update this ticket changelog after each implementation chunk with commit hashes.
- [x] Keep diary updated with exact commands, failures, and validation outcomes.
- [x] Upload the final implementation report (post-port) to reMarkable.
