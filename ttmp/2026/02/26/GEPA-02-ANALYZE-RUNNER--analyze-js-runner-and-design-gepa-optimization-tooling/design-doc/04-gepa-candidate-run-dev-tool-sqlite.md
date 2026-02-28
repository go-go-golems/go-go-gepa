---
Title: 'gepa candidate run v2: glazed command, external script, split config/input files'
Ticket: GEPA-02-ANALYZE-RUNNER
Status: active
Topics:
    - gepa
    - runner
    - goja
    - js-bindings
    - sqlite
    - tooling
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/gepa-runner/main.go
      Note: root CLI currently lacks candidate command group
    - Path: cmd/gepa-runner/eval_command.go
      Note: closest existing command wiring to mirror with Glazed
    - Path: cmd/gepa-runner/plugin_loader.go
      Note: current evaluate-only plugin callable extraction
    - Path: cmd/gepa-runner/run_recorder.go
      Note: sqlite recorder baseline for new candidate-run table
    - Path: cmd/gepa-runner/dataset.go
      Note: reusable JSON/YAML parsing helpers for candidate/input files
ExternalSources: []
Summary: V2 candidate-run design constrained to separate candidate config and input files, external --script flag, no output config in YAML, and Glazed-first command wiring.
LastUpdated: 2026-02-26T13:20:00-05:00
WhatFor: Implement candidate-run as a strict dev-tool building block.
WhenToUse: Use for implementing `gepa candidate run` in go-go-gepa.
---

# `gepa candidate run` v2

## 1. Influence from GEPA-01

GEPA-01 hard-cut means GEPA-02 must assume:

1. No dependency on `require("geppetto/plugins")`.
2. `--script` is always an external CLI input.
3. Command/runtime ownership stays in `go-go-gepa`.

## 2. Fixed Constraints (v2)

1. Separate files:
   - candidate config file (`--config`)
   - input file (`--input-file`)
2. `--script` is required on CLI and not stored in YAML.
3. No output section in YAML; all output/storage controls are CLI flags.
4. Command implemented with Glazed conventions.

## 3. CLI Contract

```bash
gepa candidate run \
  --script ./scripts/my_candidate_runner.js \
  --config ./configs/candidate-run.yaml \
  --input-file ./inputs/example-001.json \
  --output-format json \
  --record --record-db ./.gepa-runner/runs.sqlite
```

Required flags:

1. `--script <path>`
2. `--config <path>` candidate config only
3. `--input-file <path>` input payload only

Optional flags:

1. `--record`
2. `--record-db <path>`
3. `--candidate-id <string>`
4. `--reflection-used <string>`
5. `--tags <k=v,k=v>`
6. `--output-format <table|json|yaml|text>`
7. standard Glazed output flags (`--output`, `--fields`, etc)

## 4. Candidate Config Schema (`gepa.candidate-run/v2`)

`--config` contains candidate behavior inputs only.

```yaml
apiVersion: gepa.candidate-run/v2

candidate:
  prompt: |
    Solve carefully and return only final answer.
  planner_prompt: |
    Identify operation type before solving.

metadata:
  candidate_id: cand-007
  reflection_used: merge-from-cand-003-004
  tags:
    suite: smoke
    branch: gepa-02

runtime:
  profile: default
  engine_overrides:
    temperature: 0.2
```

Not allowed in YAML:

1. `script`
2. output/storage paths or format (`output`, `output_dir`, etc)
3. input example content

## 5. Input File Schema

`--input-file` is a standalone JSON/YAML object, e.g.:

```json
{
  "question": "14+9",
  "answer": "23"
}
```

This file is intentionally separate from candidate config to keep run input swap-friendly.

## 6. Glazed Command Authoring Shape

Command pattern (per `glazed-command-authoring`):

1. `type CandidateRunCommand struct { *cmds.CommandDescription }`
2. `type CandidateRunSettings struct { ... glazed tags ... }`
3. constructor uses:
   - `cmds.NewCommandDescription(...)`
   - `cmds.WithFlags(fields.New(...))`
   - `cmds.WithSections(settings.NewGlazedSchema(), cli.NewCommandSettingsSection(), geppetto sections... )`
4. execution method decodes via `vals.DecodeSectionInto(schema.DefaultSlug, settings)`
5. wire in cobra with `cli.BuildCobraCommandFromCommand(...)`

Suggested settings struct:

```go
type CandidateRunSettings struct {
    ScriptPath     string `glazed:"script"`
    ConfigPath     string `glazed:"config"`
    InputFile      string `glazed:"input-file"`
    Record         bool   `glazed:"record"`
    RecordDB       string `glazed:"record-db"`
    CandidateID    string `glazed:"candidate-id"`
    ReflectionUsed string `glazed:"reflection-used"`
    Tags           string `glazed:"tags"`
    OutputFormat   string `glazed:"output-format"`
}
```

## 7. Runtime and Plugin Behavior

1. Load `--script` plugin.
2. Require `run(input, options)` callable for candidate-run mode.
3. Do not use `evaluate()` fallback by default.
4. Build options from profile/engine/tag context.
5. Execute once and emit raw output.

## 8. SQLite Storage

Keep dedicated table:

```sql
CREATE TABLE IF NOT EXISTS gepa_candidate_runs (
  run_id TEXT PRIMARY KEY,
  timestamp_ms INTEGER NOT NULL,
  plugin_id TEXT,
  plugin_name TEXT,
  candidate_id TEXT,
  reflection_used TEXT,
  tags_json TEXT,
  candidate_json TEXT NOT NULL,
  input_json TEXT NOT NULL,
  output_json TEXT NOT NULL,
  metadata_json TEXT,
  config_json TEXT,
  status TEXT NOT NULL,
  error TEXT
);
```

## 9. Execution Flow

```text
parse CLI (Glazed)
  -> read candidate config file
  -> read input file
  -> load plugin from --script
  -> run once
  -> emit output through Glazed processor
  -> persist sqlite row when --record=true
```

## 10. Implementation Tasks (candidate-run only)

1. Add `candidate` command group and `run` subcommand (Glazed wiring).
2. Add `run()` callable support in plugin loader.
3. Add config-file parser for `gepa.candidate-run/v2`.
4. Enforce no script/output/input fields in config schema.
5. Add `gepa_candidate_runs` insert path.
6. Add tests for split files + required `--script` behavior.
