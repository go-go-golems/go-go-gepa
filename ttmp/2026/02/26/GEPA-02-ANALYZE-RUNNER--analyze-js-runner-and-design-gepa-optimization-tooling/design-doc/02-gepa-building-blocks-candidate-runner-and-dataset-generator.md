---
Title: 'GEPA Building Blocks: Candidate Runner and Dataset Generator'
Ticket: GEPA-02-ANALYZE-RUNNER
Status: active
Topics:
    - gepa
    - runner
    - goja
    - optimization
    - benchmarking
    - tooling
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/gepa-runner/dataset.go
      Note: Dataset and seed file loading - reused for candidate resolution
    - Path: cmd/gepa-runner/eval_command.go
      Note: Closest precedent - existing eval command with single-prompt interface
    - Path: cmd/gepa-runner/plugin_loader.go
      Note: Plugin loading and Evaluate() call path - reused by candidate run
    - Path: cmd/gepa-runner/run_recorder.go
      Note: SQLite recording - extended with candidate metadata columns
    - Path: cmd/gepa-runner/scripts/lib/gepa_optimizer_common.js
      Note: JS utilities reused by candidate run
    - Path: cmd/gepa-runner/scripts/lib/gepa_plugin_contract.js
      Note: Plugin contract - extended with defineDatasetGenerator
ExternalSources: []
Summary: 'Design for standalone GEPA building blocks: single-candidate evaluation runner and synthetic dataset generator. No optimizer loop -- just the primitives.'
LastUpdated: 2026-02-26T12:10:00-05:00
WhatFor: Guide implementation of gepa candidate run and gepa dataset generate commands
WhenToUse: When implementing the building-block CLI commands
---


# GEPA Building Blocks: Candidate Runner and Dataset Generator

## 1. Executive Summary

This document designs two standalone building-block commands for the GEPA toolchain:

1. **`gepa candidate run`** -- Run a single candidate (a set of prompt parameters) against one or more inputs from a dataset, using a JS plugin's `evaluate()` function. Store results with rich metadata (candidate identity, run config, per-example scores and outputs) to SQLite and/or output files.

2. **`gepa dataset generate`** -- Generate synthetic datasets using JS generator plugins or YAML templates. Output to JSON files and/or a dataset directory.

These are the *primitives* that the optimizer loop will eventually call, but they are useful on their own for:
- Manual testing of a candidate prompt against specific examples
- Building evaluation baselines before optimization
- Creating reproducible synthetic benchmarks
- Debugging individual evaluations in isolation

The optimizer loop itself is out of scope for this document.

## 2. Problem Statement

### 2.1 What exists today

The current `gepa-runner` CLI has two commands that are coupled to higher-level workflows:

**`eval` command** (`eval_command.go:40-230`): Evaluates a single prompt (string) against an *entire* dataset. It:
- Requires `--prompt` or `--prompt-file` (single string, always mapped to `{"prompt": text}`)
- Iterates over every example in the dataset
- Records all results in a single batch to SQLite
- Cannot evaluate a multi-parameter candidate (the `{"prompt": ..., "planner_prompt": ...}` map form)
- Cannot run against a single example or a subset
- Does not store the candidate config, run tags, or experiment context

**`optimize` command** (`main.go:67-396`): Runs the full evolutionary loop. It:
- Supports multi-parameter candidates via `--seed-candidate`
- Has rich plugin wiring (merge, component selection, side-info hooks)
- Records to SQLite, but only at the end of the full optimization run
- Cannot be used for "run this candidate on these 3 examples and show me the results"

**What's missing:** a command that takes a candidate (multi-parameter YAML/JSON), a specific example or subset of examples, runs `evaluate()` on each, and stores results with metadata. The existing `eval` command is close but has three gaps:
1. Only accepts a single `--prompt` string, not a multi-parameter candidate map
2. Always evaluates the entire dataset (no example selection)
3. Does not store run configuration or candidate metadata beyond a SHA256

For dataset generation, nothing exists. The `dataset()` plugin method returns hardcoded data; there is no generation or parameterization.

### 2.2 Desired outcome

A researcher should be able to:

```bash
# Run a candidate against a single example
gepa candidate run \
  --script optimizer.js \
  --candidate candidate.yaml \
  --input '{"question": "2+2", "answer": "4"}' \
  --record

# Run a candidate against examples 0,2,5 from a dataset
gepa candidate run \
  --script optimizer.js \
  --candidate candidate.yaml \
  --dataset data.json \
  --examples 0,2,5 \
  --record --record-db runs.sqlite

# Run a candidate against the full dataset
gepa candidate run \
  --script optimizer.js \
  --candidate candidate.yaml \
  --dataset data.json \
  --record

# Generate a synthetic dataset
gepa dataset generate \
  --generator generators/math.js \
  --count 100 \
  --seed 42 \
  --output data/math_100.json

# Generate from a YAML template
gepa dataset generate \
  --template templates/arithmetic.yaml \
  --count 50 \
  --output data/arith_50.json
```

## 3. Current-State Analysis

### 3.1 Plugin evaluate interface

The evaluate call path is (`plugin_loader.go:206-238`):

```go
func (p *optimizerPlugin) Evaluate(
    candidate gepaopt.Candidate,   // map[string]string
    exampleIndex int,
    example any,                   // arbitrary JSON value
    opts pluginEvaluateOptions,    // {Profile, EngineOptions, Tags}
) (gepaopt.EvalResult, error)
```

The JS plugin receives:
```javascript
evaluate({candidate, example, exampleIndex}, {profile, engineOptions, tags})
// Returns: {score, objectives?, output?, feedback?, trace?, notes?}
```

This interface already supports multi-parameter candidates and arbitrary examples. The building-block command just needs to call it directly without an optimizer loop.

### 3.2 Candidate representation

From `types.go:12`:
```go
type Candidate map[string]string
```

A candidate YAML file would be:
```yaml
prompt: |
  Solve the math problem step by step.
  Return only the final numeric answer.
planner_prompt: |
  Before solving, identify the operation needed.
```

This is already supported by `loadSeedCandidateFile()` in `dataset.go`, which parses JSON or YAML candidate files. The `--seed-candidate` flag in the optimize command uses exactly this.

### 3.3 Dataset loading

From `dataset.go`, `loadDataset()` handles JSON arrays and JSONL files. The plugin's `dataset()` method is the alternative source. Both return `[]any`.

### 3.4 Recording layer

From `run_recorder.go`, the existing recorder writes to three tables:
- `gepa_runs` -- run-level metadata
- `gepa_candidate_metrics` -- per-candidate aggregate stats
- `gepa_eval_examples` -- per-(candidate, example) evaluation results

The `RecordEvalResult()` method (`run_recorder.go:189-239`) already handles single-candidate evaluation recording. It stores the candidate hash, per-example scores, feedback, output, trace, and raw JSON.

### 3.5 Gaps in existing recorder for building-block use

The `runRecorderConfig` (`run_recorder.go:24-35`) captures:
- `Mode` -- "optimize" or "eval"
- `PluginID`, `PluginName`, `Profile`
- `DatasetSize`, `Objective`, `MaxEvals`, `BatchSize`
- `SeedPrompt` -- only the prompt text

Missing for building-block runs:
1. **Candidate JSON** -- the full candidate map, not just a seed prompt SHA
2. **Run tags/labels** -- arbitrary key-value metadata (e.g., `{"experiment": "baseline", "variant": "v2"}`)
3. **Example selection** -- which examples were evaluated (the indices), since we may not run the full dataset
4. **Config snapshot** -- the YAML config used for this run
5. **Candidate number / identity** -- a user-provided label like "candidate-7" or "baseline"

## 4. Proposed Solution

### 4.1 `gepa candidate run` command

#### 4.1.1 CLI interface

```
gepa candidate run [flags]

Required:
  --script PATH          JS optimizer plugin (provides evaluate())

Candidate (one of):
  --candidate PATH       YAML or JSON file with candidate map
  --prompt TEXT          Single prompt text (shorthand for {"prompt": TEXT})
  --prompt-file PATH    Single prompt from file

Input (one of):
  --input JSON           Single example as inline JSON
  --input-file PATH      Single example from JSON file
  --dataset PATH         Dataset file (JSON array or JSONL)
  --examples INDICES     Comma-separated example indices to evaluate (requires --dataset)
                         If omitted with --dataset, evaluates all examples.

Metadata:
  --name TEXT            Human label for this candidate run (e.g., "baseline-v2")
  --tags KEY=VAL,...     Arbitrary tags stored with the run
  --config PATH          YAML config file (bundles all settings; CLI flags override)

Recording:
  --record               Persist to SQLite (default: false)
  --record-db PATH       SQLite path (default: .gepa-runner/runs.sqlite)

Output:
  --output-json PATH     Write results as JSON to file
  --output-format FMT    Output format: table, json, yaml (default: table)
  --verbose              Show per-example details (default: summary only)
```

#### 4.1.2 YAML config file format

Instead of passing everything as flags, the user can write a config file:

```yaml
# candidate-run.yaml
apiVersion: gepa.candidate-run/v1

script: ./scripts/toy_math_optimizer.js

candidate:
  prompt: |
    Solve the math problem step by step.
    Return only the final numeric answer.

# Optional: inline input examples (alternative to --dataset)
inputs:
  - { question: "2+2", answer: "4" }
  - { question: "10-3", answer: "7" }

# Or reference a dataset file:
# dataset: ./data/math_examples.json
# examples: [0, 2, 5]  # optional subset

metadata:
  name: "baseline-v2"
  tags:
    experiment: "merge-comparison"
    variant: "no-merge"

record:
  enabled: true
  db: .gepa-runner/runs.sqlite

engine:
  provider: anthropic
  model: claude-sonnet-4-20250514
  temperature: 0.7
  max-tokens: 1024
```

Invocation:
```bash
gepa candidate run --config candidate-run.yaml
gepa candidate run --config candidate-run.yaml --tags attempt=3  # override/add tags
```

#### 4.1.3 Execution flow

```
1. Parse config (YAML file + CLI flag overrides)
2. Load JS plugin
3. Resolve candidate:
   - From --candidate file (YAML/JSON map)
   - From --prompt / --prompt-file (wrapped as {"prompt": text})
   - From config YAML candidate section
4. Resolve inputs:
   - From --input (single inline JSON)
   - From --input-file (single JSON file)
   - From --dataset + optional --examples filter
   - From config YAML inputs section
   - From plugin.dataset() as fallback
5. For each (example, index) pair:
   a. Call plugin.Evaluate(candidate, index, example, opts)
   b. Collect EvalResult
   c. Print per-example result if --verbose
6. Compute aggregate stats (mean score, mean objectives)
7. Record to SQLite if --record
8. Write output (table/json/yaml)
```

#### 4.1.4 Output formats

**Table (default):**
```
Candidate: baseline-v2
Plugin: example.toy_math (Example: Toy math accuracy)
Examples: 6

  #  SCORE  FEEDBACK
  0  1.000  Correct.
  1  1.000  Correct.
  2  0.000  Expected "42" but got "41".
  3  1.000  Correct.
  4  1.000  Correct.
  5  1.000  Correct.

Mean score: 0.833333
```

**JSON:**
```json
{
  "candidate": {"prompt": "..."},
  "candidate_name": "baseline-v2",
  "plugin": {"id": "example.toy_math", "name": "Example: Toy math accuracy"},
  "tags": {"experiment": "merge-comparison", "variant": "no-merge"},
  "examples_evaluated": 6,
  "stats": {"mean_score": 0.833333, "n": 6},
  "results": [
    {
      "example_index": 0,
      "score": 1.0,
      "output": {"text": "4"},
      "feedback": "Correct."
    },
    ...
  ]
}
```

### 4.2 Storage schema extensions

#### 4.2.1 Extended `gepa_runs` table

Add columns for building-block run metadata:

```sql
ALTER TABLE gepa_runs ADD COLUMN candidate_name TEXT;
ALTER TABLE gepa_runs ADD COLUMN candidate_json TEXT;
ALTER TABLE gepa_runs ADD COLUMN tags_json TEXT DEFAULT '{}';
ALTER TABLE gepa_runs ADD COLUMN config_yaml TEXT;
ALTER TABLE gepa_runs ADD COLUMN examples_json TEXT;  -- which indices were evaluated
```

The `mode` column gains a new value: `"candidate_run"` (alongside existing `"optimize"` and `"eval"`).

#### 4.2.2 Schema migration

Use an additive migration strategy with version tracking:

```sql
CREATE TABLE IF NOT EXISTS gepa_schema_version (
  version INTEGER PRIMARY KEY
);
```

Migration 1 (from current schema):
```sql
-- Check if column exists before adding (SQLite doesn't support IF NOT EXISTS for ALTER)
-- Handled in Go code with pragma table_info checks
ALTER TABLE gepa_runs ADD COLUMN candidate_name TEXT;
ALTER TABLE gepa_runs ADD COLUMN candidate_json TEXT;
ALTER TABLE gepa_runs ADD COLUMN tags_json TEXT DEFAULT '{}';
ALTER TABLE gepa_runs ADD COLUMN config_yaml TEXT;
ALTER TABLE gepa_runs ADD COLUMN examples_json TEXT;
```

All new columns are nullable with defaults, so existing data is preserved.

#### 4.2.3 Recording flow for `candidate run`

```go
type candidateRunRecorderConfig struct {
    DBPath        string
    PluginID      string
    PluginName    string
    Profile       string
    CandidateName string             // user-provided label
    CandidateJSON string             // full candidate as JSON
    TagsJSON      string             // arbitrary tags as JSON
    ConfigYAML    string             // full config snapshot
    ExamplesJSON  string             // evaluated indices as JSON array
    DatasetSize   int                // total dataset size (may differ from evaluated count)
}
```

The recorder creates a run with `mode = "candidate_run"`, then calls `RecordEvalResult()` with the enriched metadata.

### 4.3 `gepa dataset generate` command

#### 4.3.1 CLI interface

```
gepa dataset generate [flags]

Source (one of):
  --generator PATH       JS generator plugin
  --template PATH        YAML template file

Output:
  --output PATH          Output file (.json or .jsonl)
  --output-dir PATH      Output directory (one file per example)
  --count N              Number of examples to generate (required)
  --seed N               Random seed for reproducibility (default: 0 = random)

Formatting:
  --format FMT           json (default) or jsonl
  --pretty               Pretty-print JSON output
```

#### 4.3.2 JS generator plugin contract

New plugin kind: `gepa.dataset-generator/v1`

```javascript
const plugins = require("./lib/gepa_plugin_contract");

module.exports = plugins.defineDatasetGenerator({
  apiVersion: "gepa.dataset-generator/v1",
  kind: "dataset-generator",
  id: "math.arithmetic",
  name: "Arithmetic Dataset Generator",

  create(ctx) {
    return {
      // Required: generate examples
      generate(count, options) {
        const rng = options.rng;  // seeded random if --seed provided
        const examples = [];
        for (let i = 0; i < count; i++) {
          const a = rng.intN(100);
          const b = rng.intN(100);
          examples.push({
            question: `${a} + ${b}`,
            answer: String(a + b),
          });
        }
        return examples;
      },

      // Optional: describe the schema for validation
      schema() {
        return {
          fields: ["question", "answer"],
          description: "Arithmetic QA pairs",
        };
      },
    };
  },
});
```

The `generate(count, options)` function receives:
- `count` -- number of examples requested
- `options.rng` -- a seeded RNG object with `intN(max)`, `float64()`, `choice(array)` methods
- `options.seed` -- the seed value used

The function returns an array of example objects.

#### 4.3.3 YAML template alternative

For simpler cases where a full JS plugin is overkill:

```yaml
# templates/arithmetic.yaml
apiVersion: gepa.dataset-template/v1
id: math.arithmetic
name: Arithmetic Dataset Template

templates:
  - weight: 3
    pattern:
      question: "What is {a} + {b}?"
      answer: "{result}"
    variables:
      a: { type: int, min: 1, max: 100 }
      b: { type: int, min: 1, max: 100 }
    compute:
      result: "{a} + {b}"

  - weight: 2
    pattern:
      question: "What is {a} * {b}?"
      answer: "{result}"
    variables:
      a: { type: int, min: 2, max: 12 }
      b: { type: int, min: 2, max: 12 }
    compute:
      result: "{a} * {b}"

  - weight: 1
    pattern:
      question: "What is {a} - {b}?"
      answer: "{result}"
    variables:
      a: { type: int, min: 10, max: 100 }
      b: { type: int, min: 1, max: "{a}" }
    compute:
      result: "{a} - {b}"
```

Templates are selected by weight for each example. Variables are sampled uniformly from their ranges. Compute expressions are evaluated with simple integer arithmetic.

This is intentionally limited -- for complex generation logic, use a JS generator.

#### 4.3.4 Execution flow

```
1. Parse flags
2. If --generator:
   a. Load JS generator plugin
   b. Create seeded RNG
   c. Call plugin.generate(count, {rng, seed})
3. If --template:
   a. Parse YAML template
   b. Create seeded RNG
   c. For each example:
      - Select template by weight
      - Sample variables from ranges
      - Evaluate compute expressions
      - Substitute into pattern
4. Write output:
   a. To --output file (JSON array or JSONL)
   b. To --output-dir (one file per example)
5. Print summary: count, schema, output path
```

### 4.4 Plugin contract extension in `gepa_plugin_contract.js`

Add a `defineDatasetGenerator()` helper alongside the existing `defineOptimizerPlugin()`:

```javascript
// In gepa_plugin_contract.js

const DATASET_GENERATOR_API_VERSION = "gepa.dataset-generator/v1";

function defineDatasetGenerator(desc) {
  if (!desc || typeof desc !== "object") {
    throw new Error("defineDatasetGenerator: descriptor must be an object");
  }
  if (desc.apiVersion !== DATASET_GENERATOR_API_VERSION) {
    throw new Error(`defineDatasetGenerator: unsupported apiVersion "${desc.apiVersion}"`);
  }
  if (desc.kind !== "dataset-generator") {
    throw new Error('defineDatasetGenerator: kind must be "dataset-generator"');
  }
  if (!desc.id || typeof desc.id !== "string") {
    throw new Error("defineDatasetGenerator: id is required");
  }
  if (!desc.name || typeof desc.name !== "string") {
    throw new Error("defineDatasetGenerator: name is required");
  }
  if (typeof desc.create !== "function") {
    throw new Error("defineDatasetGenerator: create must be a function");
  }
  return Object.freeze(desc);
}

module.exports = {
  // ... existing exports ...
  DATASET_GENERATOR_API_VERSION,
  defineDatasetGenerator,
};
```

### 4.5 Seeded RNG for JS generators

The Go side creates a seeded `math/rand.Rand` and exposes it to JS as an RNG object:

```go
// Exposed to JS as options.rng
type jsRNG struct {
    rng *rand.Rand
}

func (r *jsRNG) IntN(max int) int     { return r.rng.Intn(max) }
func (r *jsRNG) Float64() float64     { return r.rng.Float64() }
func (r *jsRNG) Choice(arr []any) any { return arr[r.rng.Intn(len(arr))] }
func (r *jsRNG) Shuffle(arr []any)    { r.rng.Shuffle(len(arr), func(i, j int) { arr[i], arr[j] = arr[j], arr[i] }) }
```

This ensures `--seed 42` produces identical datasets across runs.

## 5. Design Decisions

### 5.1 Separate `candidate run` from existing `eval`

The existing `eval` command operates on a single prompt string and always evaluates the full dataset. Rather than modifying it (breaking backward compat), we add `candidate run` as a new command that supports:
- Multi-parameter candidates
- Example subsetting
- Inline input
- Rich metadata (name, tags)
- Config files

The old `eval` command continues to work unchanged.

### 5.2 Config YAML is optional, not required

The building blocks should work with just CLI flags for quick one-off runs:
```bash
gepa candidate run --script optimizer.js --prompt "Solve it" --input '{"q":"2+2","a":"4"}'
```

The YAML config is a convenience for reproducible, documented runs. CLI flags override YAML values.

### 5.3 Recording is opt-in

Like the existing `--record` flag, recording to SQLite is off by default. Quick debugging runs shouldn't require database setup. When enabled, the building blocks use the same SQLite file and tables as the optimizer, so results are queryable together.

### 5.4 Reuse existing plugin loader

`candidate run` uses the same `loadOptimizerPlugin()` and `optimizerPlugin.Evaluate()` code path. No need for a new plugin kind for evaluation -- the `gepa.optimizer/v1` contract already defines `evaluate()`.

Dataset generation *does* introduce a new plugin kind (`gepa.dataset-generator/v1`) because the interface is fundamentally different (generate N examples vs. evaluate a candidate on an example).

### 5.5 YAML template engine is intentionally simple

The template system handles the 80% case (parameterized QA generation with arithmetic). For anything complex (multi-step reasoning, conditional generation, external data), use a JS generator plugin. The template engine is not a general-purpose programming language.

## 6. Implementation Plan

### Phase 1: `gepa candidate run` (core)

**New files:**
- `cmd/gepa-runner/candidate_run_command.go` -- Command definition and execution

**Modified files:**
- `cmd/gepa-runner/main.go` -- Register `candidate run` subcommand
- `cmd/gepa-runner/run_recorder.go` -- Add new columns, schema migration, `RecordCandidateRun()` method

**Steps:**
1. Add `CandidateRunCommand` struct with Glazed flags
2. Implement candidate resolution (file, inline, prompt)
3. Implement input resolution (inline, file, dataset + subset)
4. Wire to existing plugin loader and `Evaluate()`
5. Implement output formatting (table, json, yaml)
6. Add schema migration for new `gepa_runs` columns
7. Implement `RecordCandidateRun()` in recorder
8. Register command in `main.go`
9. Tests: inline input, file input, dataset subset, recording round-trip

**Estimated scope:** ~350 lines of Go

### Phase 2: `gepa candidate run --config` (YAML config)

**New files:**
- `cmd/gepa-runner/candidate_run_config.go` -- YAML config parsing and validation

**Steps:**
1. Define `CandidateRunConfig` struct with YAML tags
2. Implement parsing with `gopkg.in/yaml.v3`
3. Implement merging: YAML defaults + CLI flag overrides
4. Store config snapshot in SQLite when recording
5. Tests: config parsing, override behavior, validation

**Estimated scope:** ~200 lines of Go

### Phase 3: `gepa dataset generate` (JS generators)

**New files:**
- `cmd/gepa-runner/dataset_generate_command.go` -- Command definition
- `cmd/gepa-runner/dataset_generator_loader.go` -- JS generator plugin loading
- `cmd/gepa-runner/scripts/lib/gepa_dataset_contract.js` -- Generator contract (JS side)
- `cmd/gepa-runner/scripts/generators/math_generator.js` -- Example generator

**Steps:**
1. Add `defineDatasetGenerator()` to JS contract library
2. Implement generator plugin loader (similar to optimizer plugin loader)
3. Implement seeded RNG bridge for JS
4. Implement `DatasetGenerateCommand`
5. Write example math generator
6. Tests: generation, seed reproducibility, output formats

**Estimated scope:** ~300 lines of Go + ~80 lines of JS

### Phase 4: `gepa dataset generate --template` (YAML templates)

**New files:**
- `cmd/gepa-runner/dataset_template.go` -- YAML template engine

**Steps:**
1. Define template YAML schema
2. Implement variable sampling (int range, float range, choice)
3. Implement simple compute expressions (arithmetic only)
4. Implement pattern substitution
5. Implement weighted template selection
6. Tests: sampling, compute, edge cases

**Estimated scope:** ~250 lines of Go

## 7. Testing Strategy

### 7.1 Unit tests

- **Candidate resolution:** test YAML file, JSON file, inline prompt, multi-parameter
- **Input resolution:** test inline JSON, file, dataset subset, dataset fallback from plugin
- **Schema migration:** test that new columns are added idempotently
- **Recording:** test that all new fields round-trip through SQLite
- **YAML config parsing:** test valid configs, missing fields, type mismatches, override merging
- **Dataset generation:** test seed reproducibility, count accuracy, schema validation
- **Template engine:** test variable sampling, compute expressions, pattern substitution

### 7.2 Integration tests

- End-to-end `candidate run` with `smoke_noop_optimizer.js`
- End-to-end `candidate run` with `toy_math_optimizer.js` on inline input
- End-to-end `dataset generate` with example generator
- Record to SQLite + query back with `eval-report`

### 7.3 Backward compatibility

- Existing `eval` and `optimize` commands must work unchanged
- Existing SQLite databases must accept the migration without data loss
- Existing JS plugins must work with `candidate run` without modification

## 8. Risks, Alternatives, and Open Questions

### 8.1 Risks

1. **Plugin loading cost.** Every `candidate run` invocation creates a goja VM and loads the plugin. For batch evaluation of many candidates, this overhead adds up. Mitigation: the command evaluates multiple examples in one invocation; for many candidates, a shell loop or future batch mode handles it.

2. **SQLite schema drift.** Adding columns to shared tables means all tools must handle the migration. Mitigation: migration is additive only (new nullable columns with defaults). Version tracking prevents re-running migrations.

### 8.2 Alternatives considered

**Alternative A: Extend `eval` command instead of new `candidate run`.**
Rejected because `eval` has a specific interface (single prompt string, full dataset) that users may depend on. Adding multi-parameter candidates, example subsetting, and config files to it would change its behavior and flag surface significantly.

**Alternative B: Separate binary for dataset generation.**
Rejected because the generator plugins need the same goja runtime and require() infrastructure as optimizer plugins. Sharing the binary avoids duplication.

**Alternative C: Use goja for template evaluation instead of a custom engine.**
Considered running YAML template expressions through goja. Rejected because the template engine needs to be simple and safe (no arbitrary code execution from a YAML file). A purpose-built evaluator for `{a} + {b}` is simpler and more predictable.

### 8.3 Open questions

1. **Should `candidate run` accept piped input?** Reading examples from stdin (`echo '{"q":"2+2"}' | gepa candidate run --script ...`) would enable composition with other tools. Recommendation: yes, add `--input -` to read from stdin.

2. **Should dataset generation support streaming output?** For very large datasets (100K+ examples), holding everything in memory may be problematic. Recommendation: defer; start with in-memory generation, add streaming JSONL mode later if needed.

3. **Should we add `gepa candidate diff` for comparing two candidates?** Running two candidates on the same examples and showing a side-by-side comparison would be useful. Recommendation: defer to a future phase, can be built on top of `candidate run` + `analyze`.

4. **Should the config YAML include engine/profile settings?** The current engine setup goes through Geppetto's profile system. Embedding engine settings in the YAML config improves reproducibility but adds complexity. Recommendation: support both -- `engine:` section in YAML for explicit config, fall back to `--profile` flag for convenience.

5. **Should `candidate run` support running the candidate's `initialCandidate()` as input?** If no candidate is provided, should it ask the plugin? Recommendation: no, that conflates the candidate runner with seedless initialization. Use `gepa candidate run --candidate <(gepa plugin init --script ...)` or similar composition.

## 9. References

### Key source files

| File | Role |
|------|------|
| `cmd/gepa-runner/eval_command.go` | Existing eval command (closest precedent) |
| `cmd/gepa-runner/main.go` | CLI entrypoint, command registration |
| `cmd/gepa-runner/plugin_loader.go` | Plugin loading, `Evaluate()` call path |
| `cmd/gepa-runner/run_recorder.go` | SQLite recording layer |
| `cmd/gepa-runner/dataset.go` | Dataset/seed file loading |
| `cmd/gepa-runner/js_runtime.go` | Goja VM setup |
| `cmd/gepa-runner/scripts/lib/gepa_plugin_contract.js` | Plugin contract validation |
| `cmd/gepa-runner/scripts/lib/gepa_optimizer_common.js` | JS utilities (runUserPrompt, exactMatchScore) |
| `cmd/gepa-runner/scripts/toy_math_optimizer.js` | Example plugin with inline dataset |
| `pkg/optimizer/gepa/types.go` | Candidate, EvalResult, CandidateStats |

### Related tickets

- **GEPA-02-ANALYZE-RUNNER** design doc 01 -- Full optimizer tooling design (deferred)
- **GP-05-GEPA-PARITY-PLUGIN-RESEARCH** -- Plugin extension architecture
- **GEPA-01-EXTRACT-GEPPETTO-PLUGINS** -- Plugin contract ownership
