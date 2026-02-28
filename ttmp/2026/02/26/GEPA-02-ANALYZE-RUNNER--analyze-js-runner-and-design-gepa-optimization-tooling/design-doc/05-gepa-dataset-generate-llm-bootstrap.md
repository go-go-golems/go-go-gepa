---
Title: 'gepa dataset generate v2: glazed command, external script, cli-owned output flags'
Ticket: GEPA-02-ANALYZE-RUNNER
Status: active
Topics:
    - gepa
    - runner
    - goja
    - geppetto
    - dataset
    - sqlite
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/gepa-runner/main.go
      Note: root CLI currently lacks dataset command group
    - Path: cmd/gepa-runner/js_runtime.go
      Note: runtime already supports geppetto-backed JS execution
    - Path: cmd/gepa-runner/plugin_loader.go
      Note: current optimizer-only descriptor loader
    - Path: cmd/gepa-runner/dataset.go
      Note: existing load helpers to reuse/extend
    - Path: cmd/gepa-runner/scripts/lib/gepa_optimizer_common.js
      Note: reusable LLM prompting helper via geppetto
ExternalSources: []
Summary: V2 dataset-generate design requiring external --script, YAML with generation semantics only, and output/storage as CLI flags under Glazed.
LastUpdated: 2026-02-26T13:20:00-05:00
WhatFor: Implement dataset generation as a foundational CLI building block.
WhenToUse: Use when implementing `gepa dataset generate` in go-go-gepa.
---

# `gepa dataset generate` v2

## 1. Influence from GEPA-01

GEPA-01 hard-cut influences GEPA-02 dataset generation:

1. No reliance on `geppetto/plugins` helper module.
2. Script path is explicit external CLI input (`--script`).
3. Implementation stays in go-go-gepa command/runtime code.

## 2. Fixed Constraints (v2)

1. `--script` is provided externally (required CLI flag).
2. YAML config does not carry output directives.
3. Output/storage controls are CLI flags only.
4. Use Glazed for command definition, parsing, and output controls.

## 3. CLI Contract

```bash
gepa dataset generate \
  --script ./scripts/generators/arithmetic_generator.js \
  --config ./configs/dataset-generate.yaml \
  --count 100 \
  --output-dir ./data/generated \
  --output-db ./data/generated/datasets.sqlite \
  --output-format json
```

Required flags:

1. `--script <path>`
2. `--config <path>`

Optional flags:

1. `--count <n>` (overrides config default)
2. `--seed <int>`
3. `--output-dir <path>`
4. `--output-db <path>`
5. `--output-file-stem <name>`
6. `--dry-run`
7. standard Glazed output flags

## 4. Config Schema (`gepa.dataset-generate/v2`)

```yaml
apiVersion: gepa.dataset-generate/v2
name: arithmetic-starter
count: 50
seed: 42

prompting:
  system: |
    You generate arithmetic training examples as JSON only.
  user_template: |
    Create one example with fields:
    - question
    - answer
    Difficulty: {{difficulty}}
  variables:
    difficulty: [easy, medium, hard]

validation:
  required_fields: [question, answer]
  max_retries: 2
  drop_invalid: true
```

Not allowed in YAML:

1. `script`
2. `output` sections/paths/formats
3. sqlite output path

## 5. Glazed Command Authoring Shape

Use standard Glazed command authoring:

1. `DatasetGenerateCommand` with embedded `*cmds.CommandDescription`.
2. `DatasetGenerateSettings` with `glazed` tags.
3. constructor uses `fields.New(...)` + `cmds.WithSections(...)`.
4. decode settings from default section.
5. register via `cli.BuildCobraCommandFromCommand`.

Suggested settings struct:

```go
type DatasetGenerateSettings struct {
    ScriptPath    string `glazed:"script"`
    ConfigPath    string `glazed:"config"`
    Count         int    `glazed:"count"`
    Seed          int    `glazed:"seed"`
    OutputDir     string `glazed:"output-dir"`
    OutputDB      string `glazed:"output-db"`
    OutputFileStem string `glazed:"output-file-stem"`
    DryRun        bool   `glazed:"dry-run"`
}
```

## 6. Dataset Generator Contract

Define dedicated descriptor contract:

```javascript
module.exports = {
  apiVersion: "gepa.dataset-generator/v1",
  kind: "dataset-generator",
  id: "example.arithmetic_gen",
  name: "Example Arithmetic Generator",
  create(ctx) {
    return {
      generateOne(input, options) {
        return {
          row: { question: "2+2", answer: "4" },
          metadata: { difficulty: "easy" }
        };
      }
    };
  }
};
```

Input includes `{index, seed, variables, promptSpec}`.

## 7. Output and Persistence Model (CLI-owned)

CLI flags decide whether/where data is written:

1. `--output-dir`: JSONL + metadata file output.
2. `--output-db`: sqlite persistence enabled.
3. no output routing in YAML.

SQLite tables remain:

1. `gepa_generated_datasets`
2. `gepa_generated_dataset_rows`

## 8. Execution Flow

```text
parse CLI (Glazed)
  -> read config file
  -> apply CLI overrides (count/seed/output*)
  -> load plugin via --script
  -> generate N rows
  -> validate/retry rows
  -> write JSONL/meta when --output-dir set
  -> write sqlite rows when --output-db set
```

## 9. Implementation Tasks (dataset-generate only)

1. Add `dataset` command group and `generate` subcommand (Glazed wiring).
2. Add dataset-generator descriptor loader.
3. Add config parser for `gepa.dataset-generate/v2` with no output/script keys.
4. Add CLI override layer for count/seed/output flags.
5. Add writers for JSONL + sqlite from CLI flags.
6. Add tests validating forbidden YAML output/script keys.
