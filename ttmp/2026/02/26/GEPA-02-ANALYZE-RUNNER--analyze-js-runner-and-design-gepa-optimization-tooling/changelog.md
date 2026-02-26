# Changelog

## 2026-02-26

- Initialized ticket workspace.

## 2026-02-26

- Completed deeper JS-runner analysis and captured focused experiment evidence in ticket-local `scripts/`.
- Key findings from experiments:
  - `gepa-runner` has no `candidate` command.
  - `gepa-runner` has no `dataset` command.
  - `eval` does not accept `--candidate`.
  - Plugin loader rejects run-only plugins because `evaluate()` is currently mandatory.
  - Existing eval recorder captures aggregate eval data but not explicit candidate-run metadata fields.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/scripts/exp-00-eval-profile-registry-error.txt
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/scripts/exp-01-eval-smoke-success.txt
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/scripts/exp-01-runs.sqlite
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/scripts/exp-02-run-only-plugin-fails.txt
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/scripts/exp-03-candidate-command-missing.txt
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/scripts/exp-04-dataset-command-missing.txt
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/scripts/exp-05-eval-no-candidate-flag.txt

## 2026-02-26

- Added new narrow-scope design docs aligned with building-block requirements:
  - `04-gepa-candidate-run-dev-tool-sqlite.md` (single-run dev tool, optional sqlite recording, no eval aggregation)
  - `05-gepa-dataset-generate-llm-bootstrap.md` (geppetto+LLM starter dataset generation, file + sqlite output)
- Marked these as current authoritative docs in ticket `index.md`.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/design-doc/04-gepa-candidate-run-dev-tool-sqlite.md
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/design-doc/05-gepa-dataset-generate-llm-bootstrap.md
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/index.md
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/tasks.md

## 2026-02-26

Added focused GEPA-02 building-block designs for candidate run and dataset generate, backed by JS-runner experiments and sqlite schema inspection.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/design-doc/04-gepa-candidate-run-dev-tool-sqlite.md — candidate run design
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/design-doc/05-gepa-dataset-generate-llm-bootstrap.md — dataset generation design
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/reference/01-investigation-diary.md — phase-6 command log and evidence

## 2026-02-26

Updated GEPA-02 design docs to v2 constraints:

1. external `--script` required for both `candidate run` and `dataset generate`,
2. candidate-run uses separate files (`--config` for candidate config, `--input-file` for run input),
3. no output section in YAML configs; output/storage routing is CLI-only flags,
4. command authoring guidance switched to explicit Glazed patterns.

Also documented GEPA-01 influence on GEPA-02 (hard-cut context) and refreshed index/tasks accordingly.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/design-doc/04-gepa-candidate-run-dev-tool-sqlite.md — v2 candidate-run constraints and Glazed wiring
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/design-doc/05-gepa-dataset-generate-llm-bootstrap.md — v2 dataset-generate constraints and Glazed wiring
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/index.md — v2 constraint summary
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/tasks.md — v2 implementation checklist updates

## 2026-02-26

Validated GEPA-02 docs after v2 rewrite and uploaded a new reMarkable bundle `GEPA-02 Candidate Run + Dataset Generate Design v2.pdf`.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/reference/01-investigation-diary.md — phase-8 prompt-context and v2 rewrite details
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/changelog.md — delivery record


## 2026-02-26

Validated GEPA-02 docs with docmgr doctor and uploaded the refreshed bundle to reMarkable at /ai/2026/02/26/GEPA-02-ANALYZE-RUNNER.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/changelog.md — delivery entry
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/reference/01-investigation-diary.md — phase-7 upload evidence

## 2026-02-26

Implemented the first GEPA-02 building block in code: `dataset generate`.

### What shipped

- Added `gepa-runner dataset generate` (Glazed command) under a new `dataset` command group.
- Added dataset-generator plugin contract loading for `gepa.dataset-generator/v1`.
- Extended `require("gepa/plugins")` with:
  - `DATASET_GENERATOR_API_VERSION`
  - `defineDatasetGenerator(...)`
- Implemented `gepa.dataset-generate/v2` config parsing with strict rejection of YAML `script` and output-routing keys.
- Implemented CLI-owned output routing:
  - `--output-dir` writes JSONL + metadata JSON
  - `--output-db` writes sqlite rows
- Added generated dataset sqlite tables:
  - `gepa_generated_datasets`
  - `gepa_generated_dataset_rows`
- Added an example script:
  - `cmd/gepa-runner/scripts/arithmetic_dataset_generator.js`

### Validation

- `go test ./cmd/gepa-runner -count=1` passed.
- `go run ./cmd/gepa-runner dataset generate --help` shows the new command.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/dataset_generate_command.go
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/dataset_generate_config.go
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/dataset_generate_store.go
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/dataset_generator_loader.go
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/gepa_plugins_module.go
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/main.go
