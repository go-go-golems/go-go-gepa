# Tasks

## Completed (Research + Design)

- [x] Analyze JS runtime path in `go-go-gepa` (`js_runtime.go`, `plugin_loader.go`, runner scripts).
- [x] Run ticket-local CLI/SQLite experiments and capture outputs in `scripts/`.
- [x] Create focused candidate-run design doc with sqlite storage model.
- [x] Create focused dataset-generate design doc for geppetto+LLM bootstrap data.
- [x] Update investigation diary with commands, failures, and decisions.
- [x] Apply v2 command constraints from user feedback:
  - external `--script`,
  - split candidate config and input file for candidate-run,
  - no output section in YAML,
  - Glazed-first command authoring.

## Pending (Implementation Work)

- [x] Add `candidate` command group and `candidate run` subcommand.
- [x] Add plugin loader support for `run()` in candidate-run mode.
- [x] Implement `gepa.candidate-run/v2` config loading with strict schema (no script/output/input keys).
- [x] Require `--input-file` as separate file from candidate config.
- [x] Require external `--script` for candidate-run.
- [x] Add `gepa_candidate_runs` sqlite table and insert path.
- [x] Add `dataset` command group and `dataset generate` subcommand.
- [x] Add dataset-generator plugin contract loader (`gepa.dataset-generator/v1`).
- [x] Implement `gepa.dataset-generate/v2` config loading with strict schema (no script/output keys).
- [x] Require external `--script` for dataset-generate.
- [x] Route output/storage only via CLI flags (`--output-dir`, `--output-db`, etc.).
- [x] Add generated dataset sqlite tables and JSONL output pipeline.
- [x] Handle dataset-generator token-limit truncation via stop-reason continuation logic in ticket-local `exp-11` script.
- [ ] Add integration tests for candidate run and dataset generate.
- [ ] Add CLI help/examples for both commands using Glazed conventions.
