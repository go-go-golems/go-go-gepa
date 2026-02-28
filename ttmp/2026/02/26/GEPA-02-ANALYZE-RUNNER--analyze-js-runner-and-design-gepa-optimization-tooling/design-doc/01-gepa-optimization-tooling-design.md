---
Title: GEPA Optimization Tooling Design
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
    - Path: cmd/gepa-runner/js_runtime.go
    - Path: cmd/gepa-runner/main.go
      Note: CLI entrypoint - optimize/eval commands
    - Path: cmd/gepa-runner/plugin_loader.go
      Note: JS plugin loading and validation
    - Path: cmd/gepa-runner/run_recorder.go
      Note: Storage schema - SQLite recording layer with identified gaps
    - Path: pkg/optimizer/gepa/config.go
      Note: Config struct with all optimizer parameters
    - Path: pkg/optimizer/gepa/optimizer.go
      Note: Core optimizer loop - event system
    - Path: pkg/optimizer/gepa/types.go
      Note: Core types - Candidate
ExternalSources: []
Summary: 'Design for GEPA optimization tooling: synthetic dataset generation, YAML-configured middleware runs, enriched storage with candidate/reflection metadata, and analysis/benchmarking support.'
LastUpdated: 2026-02-26T12:00:00-05:00
WhatFor: Guide implementation of GEPA optimization and benchmarking tools
WhenToUse: When building or extending the gepa-runner optimization harness
---


# GEPA Optimization Tooling Design

## 1. Executive Summary

This document proposes a set of tools for running structured GEPA optimization experiments. The current `gepa-runner` CLI supports running single optimization passes with JS plugins and recording results to SQLite. However, it lacks:

1. **Synthetic dataset generation** -- no tooling exists to create test datasets from templates or generators.
2. **YAML-driven experiment configuration** -- middleware stacks and optimizer parameters are set entirely through CLI flags; there is no declarative experiment file.
3. **Enriched metadata storage** -- the current SQLite schema records candidate scores and evaluations but does not store per-iteration state (which candidate was selected as parent, what operation was attempted, what reflection was used), experiment configuration snapshots, or structured tags like experiment names and variant labels.
4. **Benchmarking and analysis support** -- no tooling for running repeated trials, comparing configurations, or producing aggregate statistics.

The proposed design introduces three new commands (`gepa-runner generate-dataset`, `gepa-runner experiment`, `gepa-runner analyze`) and extends the storage schema to capture the full experiment lifecycle with enough detail for rigorous post-hoc analysis.

## 2. Problem Statement and Scope

### 2.1 What exists today

The `gepa-runner` CLI (`cmd/gepa-runner/main.go`) provides two commands:

- **`optimize`** -- Runs a GEPA optimization loop. Accepts a JS plugin (`--script`), a dataset, a seed prompt, and optimizer parameters as flags. Optionally records to SQLite via `--record`.
- **`eval`** -- Evaluates a single candidate against a dataset and optionally records.

The optimizer core (`pkg/optimizer/gepa/optimizer.go`, 1185 lines) implements the evolutionary loop with:
- Pool-based candidate tracking (`candidateNode` with ID, parentage, operation type)
- Evaluation caching by candidate hash + example index
- Event emission (`OptimizerEvent` with type, iteration, scores, acceptance)
- Hooks for merge, component selection, and side-info formatting

The recording layer (`cmd/gepa-runner/run_recorder.go`, 602 lines) persists to three SQLite tables:
- `gepa_runs` -- Run-level metadata (plugin, config, timing, best score)
- `gepa_candidate_metrics` -- Per-candidate aggregate scores and lineage
- `gepa_eval_examples` -- Per-(candidate, example) evaluation details

### 2.2 What is missing

**For running experiments:**

1. No way to define an experiment configuration as a single YAML file that bundles: optimizer parameters, plugin path, dataset path, seed candidate, middleware/reflection settings, and experiment metadata (name, tags, variant label).
2. No way to run the same experiment N times with different random seeds for statistical significance.
3. No synthetic dataset generation pipeline.

**For storage and analysis:**

4. The `gepa_runs` table does not record the full experiment configuration (only scattered fields like `max_evals`, `batch_size`). There is no snapshot of the YAML config or the optimizer `Config` struct.
5. The `gepa_candidate_metrics` table does not store `parent2_id` (merge second parent), `operation` type, `updated_keys`, or `created_at` timestamp. These exist in the in-memory `CandidateEntry` but are dropped during recording (see `run_recorder.go:161-185`).
6. There is no per-iteration event log. The `OptimizerEvent` struct is emitted to stdout via `--show-events` but never persisted. This means we cannot reconstruct the optimization trajectory post-hoc.
7. There is no experiment-level grouping. When running the same config 10 times, there is no way to link those runs or compute aggregate statistics across them.
8. Result data (the final `Result` struct with all candidates) is only written as a JSON file via `--out-report`. It is not queryable.

### 2.3 Desired outcomes

A researcher should be able to:

1. Write a YAML experiment file that fully specifies an optimization run.
2. Run `gepa-runner experiment run experiment.yaml --trials 5` and get 5 recorded runs.
3. Query the database to compare configurations: "which merge scheduler gives better scores on average?"
4. Reconstruct the full optimization trajectory of any run: which parent was selected, what operation was applied, whether the child was accepted, and what the reflection LLM said.
5. Generate synthetic datasets from templates for controlled benchmarking.

## 3. Current-State Architecture

### 3.1 JS runner architecture

The JS runtime is set up in `cmd/gepa-runner/js_runtime.go` (76 lines):

```go
func newJSRuntime(scriptRoot string) (*jsRuntime, error) {
    // Creates goja VM + event loop
    // Registers geppetto native module for LLM inference
    // Enables require() with node_modules resolution rooted at scriptRoot
}
```

Plugins are loaded in `plugin_loader.go` (682 lines) via the `gepa.optimizer/v1` contract:
- Plugin descriptor: `{apiVersion, kind, id, name, create(hostContext)}`
- Plugin instance: `{evaluate(), dataset?(), initialCandidate?(), merge?(), selectComponents?(), componentSideInfo?()}`

The `hostContext` passed to `create()` includes: `app`, `scriptPath`, `scriptRoot`, `profile`, `engineOptions`. This is the primary extension point for passing experiment-level configuration to plugins.

### 3.2 Optimizer core data flow

```
main.go:RunIntoWriter()
  |
  |-- Resolve seed candidate (text/file/candidate file/seedless)
  |-- Load dataset (file or plugin.dataset())
  |-- Create Geppetto inference engine (reflection LLM)
  |-- Load JS plugin and extract hooks
  |-- Build Config struct from CLI flags
  |-- Create Optimizer
  |-- Wire hooks (merge, componentSelector, sideInfo, eventHook)
  |-- Run Optimize(ctx, seedCandidate, examples)
  |      |
  |      |-- For each iteration:
  |      |     selectParent() -> selectKeys() -> evaluate parent batch
  |      |     -> mutate or merge -> evaluate child batch -> accept/reject
  |      |     -> emit OptimizerEvent
  |      |
  |      |-- Return Result{BestCandidate, BestStats, CallsUsed, Candidates}
  |
  |-- Record to SQLite if --record
  |-- Write outputs (prompt, report)
```

### 3.3 Storage schema gaps (evidence)

The `candidateMetricRow` struct (`run_recorder.go:63-75`) does NOT include:
- `Parent2ID` (merge second parent) -- exists in `CandidateEntry.Parent2ID`
- `Operation` (seed/mutate/merge) -- exists in `CandidateEntry.Operation`
- `UpdatedKeys` -- exists in `CandidateEntry.UpdatedKeys`
- `CreatedAt` -- exists in `CandidateEntry.CreatedAt`

The `runRecord` struct (`run_recorder.go:37-61`) does NOT include:
- Full config snapshot (merge scheduler, component selector, random seed, etc.)
- Experiment name, variant label, or tags
- Trial index within an experiment group

## 4. Proposed Solution

### 4.1 YAML experiment configuration

Introduce an experiment file format that bundles everything needed for a run:

```yaml
# experiment.yaml
apiVersion: gepa.experiment/v1
metadata:
  name: "math-optimizer-merge-comparison"
  tags:
    domain: "arithmetic"
    variant: "stagnation-due"
  description: "Compare merge schedulers on toy math task"

plugin:
  script: ./scripts/toy_math_optimizer.js

dataset:
  # Option A: file path
  path: ./data/math_examples.json
  # Option B: let plugin provide it (omit path)
  # Option C: generator reference
  # generator: ./scripts/generate_math_dataset.js

seed:
  # One of: text, file, candidate, seedless
  candidate:
    prompt: |
      Solve the math problem step by step. Return only the final numeric answer.

optimizer:
  max-evals: 200
  batch-size: 8
  merge-prob: 0.3
  merge-scheduler: stagnation_due
  max-merges-due: 3
  component-selector: round_robin
  optimizable-keys: [prompt]
  objective: "Maximize exact-match accuracy on arithmetic questions"
  max-side-info-chars: 8000
  random-seed: 0  # 0 = auto (different per trial)
  epsilon: 0.0

reflection:
  # Optional overrides for reflection/merge prompts
  system-prompt: "You are an expert prompt engineer."
  # prompt-template: ...
  # merge-system-prompt: ...
  # merge-prompt-template: ...

# Geppetto engine settings (passed through to inference engine factory)
engine:
  provider: anthropic
  model: claude-sonnet-4-20250514
  temperature: 0.7
  max-tokens: 1024

record:
  enabled: true
  db: .gepa-runner/runs.sqlite
```

**Design rationale:**
- The YAML structure mirrors the existing CLI flags closely to avoid surprise.
- `metadata` section provides experiment identity for grouping and querying.
- `engine` section replaces the Geppetto profile/flag system for reproducibility -- the exact model and parameters are captured in the experiment file.
- `seed.candidate` supports the multi-parameter map directly.

### 4.2 Enriched storage schema

Extend the SQLite schema with these changes:

#### 4.2.1 New table: `gepa_experiments`

Groups runs into experiments:

```sql
CREATE TABLE IF NOT EXISTS gepa_experiments (
  experiment_id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT,
  tags_json TEXT NOT NULL DEFAULT '{}',
  config_yaml TEXT NOT NULL,      -- full YAML snapshot
  config_hash TEXT NOT NULL,       -- SHA256 of normalized config
  created_at_ms INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_gepa_experiments_name
  ON gepa_experiments (name);
```

#### 4.2.2 Extended `gepa_runs` table

Add columns:

```sql
ALTER TABLE gepa_runs ADD COLUMN experiment_id TEXT
  REFERENCES gepa_experiments(experiment_id);
ALTER TABLE gepa_runs ADD COLUMN trial_index INTEGER;
ALTER TABLE gepa_runs ADD COLUMN random_seed INTEGER;
ALTER TABLE gepa_runs ADD COLUMN config_json TEXT;  -- full optimizer Config as JSON
ALTER TABLE gepa_runs ADD COLUMN merge_scheduler TEXT;
ALTER TABLE gepa_runs ADD COLUMN merge_prob REAL;
ALTER TABLE gepa_runs ADD COLUMN component_selector TEXT;
ALTER TABLE gepa_runs ADD COLUMN optimizable_keys_json TEXT;
ALTER TABLE gepa_runs ADD COLUMN engine_model TEXT;
ALTER TABLE gepa_runs ADD COLUMN engine_provider TEXT;
```

#### 4.2.3 Extended `gepa_candidate_metrics` table

Add the missing lineage and operation fields:

```sql
ALTER TABLE gepa_candidate_metrics ADD COLUMN parent2_id INTEGER DEFAULT -1;
ALTER TABLE gepa_candidate_metrics ADD COLUMN operation TEXT DEFAULT 'unknown';
ALTER TABLE gepa_candidate_metrics ADD COLUMN updated_keys_json TEXT DEFAULT '[]';
ALTER TABLE gepa_candidate_metrics ADD COLUMN created_at_ms INTEGER;
```

#### 4.2.4 New table: `gepa_iteration_events`

Persist the `OptimizerEvent` stream:

```sql
CREATE TABLE IF NOT EXISTS gepa_iteration_events (
  run_id TEXT NOT NULL,
  iteration INTEGER NOT NULL,
  event_type TEXT NOT NULL,
  operation TEXT NOT NULL,
  parent_id INTEGER NOT NULL,
  parent2_id INTEGER NOT NULL DEFAULT -1,
  child_id INTEGER NOT NULL DEFAULT -1,
  updated_keys_json TEXT DEFAULT '[]',
  parent_score REAL,
  parent2_score REAL,
  baseline_score REAL,
  child_score REAL,
  accepted INTEGER NOT NULL DEFAULT 0,
  calls_used INTEGER NOT NULL DEFAULT 0,
  reflection_raw TEXT,     -- the LLM reflection text for this iteration
  timestamp_ms INTEGER NOT NULL,
  PRIMARY KEY (run_id, iteration, event_type)
);
CREATE INDEX IF NOT EXISTS idx_gepa_events_run
  ON gepa_iteration_events (run_id, iteration);
```

This table captures the full trajectory of the optimization. Each row is one event (mutate_attempted, merge_accepted, etc.) anchored to a run and iteration.

#### 4.2.5 Schema migration strategy

Use a `schema_version` table and apply migrations incrementally:

```sql
CREATE TABLE IF NOT EXISTS gepa_schema_version (
  version INTEGER PRIMARY KEY
);
```

On startup, check the version and apply any missing `ALTER TABLE` statements. New columns use defaults so existing data is preserved.

### 4.3 New commands

#### 4.3.1 `gepa-runner experiment run`

```
gepa-runner experiment run <experiment.yaml> [flags]
  --trials N          Number of independent optimization runs (default 1)
  --parallel N        Max concurrent runs (default 1)
  --dry-run           Validate config and print resolved settings without running
  --override KEY=VAL  Override a config field (e.g., --override optimizer.batch-size=16)
```

**Behavior:**
1. Parse and validate experiment YAML.
2. Create or find `gepa_experiments` row (keyed by config_hash).
3. For each trial `i` in `[0, N)`:
   a. Resolve random seed (auto-generate if `random-seed: 0`).
   b. Create `gepa_runs` row with `experiment_id`, `trial_index`, full config snapshot.
   c. Run the optimization loop (reuse existing `Optimize()` code path).
   d. Record enriched `gepa_candidate_metrics` with lineage fields.
   e. Record `gepa_iteration_events` from event hook.
   f. Finalize run record.
4. Print summary (per-trial best scores, mean/stddev across trials).

#### 4.3.2 `gepa-runner generate-dataset`

```
gepa-runner generate-dataset [flags]
  --generator <path>  JS generator script
  --output <path>     Output JSON/JSONL file
  --count N           Number of examples to generate
  --seed N            Random seed for reproducibility
  --template <path>   YAML template for non-JS generation
```

**Generator plugin contract** (`gepa.dataset-generator/v1`):

```javascript
const descriptor = defineDatasetGenerator({
  apiVersion: "gepa.dataset-generator/v1",
  kind: "dataset-generator",
  id: "math.arithmetic",
  name: "Arithmetic Dataset Generator",
  create(options) {
    return {
      // Generate `count` examples
      generate(count, seed) {
        const examples = [];
        // ... generate examples ...
        return examples;
      },
      // Optional: describe the schema
      schema() {
        return {
          fields: ["question", "answer"],
          description: "Arithmetic QA pairs"
        };
      }
    };
  }
});
module.exports = descriptor;
```

**YAML template alternative** (for simpler cases):

```yaml
# template.yaml
apiVersion: gepa.dataset-template/v1
schema:
  question: string
  answer: string
templates:
  - pattern: "What is {a} + {b}?"
    answer: "{a+b}"
    variables:
      a: { type: int, min: 1, max: 100 }
      b: { type: int, min: 1, max: 100 }
  - pattern: "What is {a} * {b}?"
    answer: "{a*b}"
    variables:
      a: { type: int, min: 1, max: 12 }
      b: { type: int, min: 1, max: 12 }
```

#### 4.3.3 `gepa-runner analyze`

Query and aggregate experiment results:

```
gepa-runner analyze [subcommand] [flags]

Subcommands:
  runs          List runs with filtering
  compare       Compare experiments or configurations
  trajectory    Show iteration-by-iteration trajectory for a run
  candidates    List candidates for a run with lineage
  summary       Aggregate statistics across trials
```

**Examples:**

```bash
# List all runs for an experiment
gepa-runner analyze runs --experiment "math-optimizer-merge-comparison"

# Compare two experiment variants
gepa-runner analyze compare \
  --experiment "math-opt-probabilistic" \
  --experiment "math-opt-stagnation-due" \
  --metric best_mean_score

# Show trajectory of a specific run
gepa-runner analyze trajectory --run-id "gepa-optimize-1740567890123456789"

# Summary statistics across trials
gepa-runner analyze summary --experiment "math-optimizer-merge-comparison"
```

Output uses Glazed table formatting for terminal display and supports `--output json` for programmatic use.

### 4.4 Event recording implementation

Wire the existing `EventHook` to the recorder. In the `experiment run` command:

```go
// Pseudocode for the event recording hook
recorder := newRunRecorder(cfg)
opt.SetEventHook(func(event gepaopt.OptimizerEvent) {
    recorder.RecordEvent(event)  // append to in-memory buffer
    if showEvents {
        fmt.Fprintf(w, "[event] iter=%d ...\n", event.Iteration)
    }
})
```

The recorder buffers events and flushes them in the same transaction as the run finalization.

### 4.5 hostContext extension for experiment metadata

Pass experiment configuration to the JS plugin's `create(hostContext)`:

```javascript
// hostContext will include:
{
  app: "gepa-runner",
  scriptPath: "/abs/path/to/plugin.js",
  scriptRoot: "/abs/path/to/",
  profile: "default",
  engineOptions: { ... },
  // NEW: experiment metadata
  experiment: {
    name: "math-optimizer-merge-comparison",
    tags: { domain: "arithmetic", variant: "stagnation-due" },
    trialIndex: 0,
    randomSeed: 42
  }
}
```

This lets plugins adapt behavior based on experiment parameters (e.g., use a specific random seed for dataset sampling).

## 5. Pseudocode and Key Flows

### 5.1 Experiment run flow

```
ExperimentRunCommand.Run(ctx, yamlPath, trials):
  expConfig = parseExperimentYAML(yamlPath)
  validate(expConfig)

  configHash = sha256(canonicalize(expConfig))
  expID = ensureExperiment(db, expConfig, configHash)

  results = []
  for trial in range(trials):
    seed = expConfig.optimizer.randomSeed
    if seed == 0:
      seed = crypto/rand.Int63()

    cfg = buildOptimizerConfig(expConfig, seed)
    plugin = loadPlugin(expConfig.plugin.script, hostContext(expConfig, trial, seed))
    dataset = loadDataset(expConfig, plugin)
    seedCandidate = resolveSeed(expConfig)
    engine = createEngine(expConfig.engine)
    reflector = createReflector(cfg, engine)

    recorder = newRunRecorder(db, expID, trial, seed, expConfig)
    opt = NewOptimizer(cfg, evalFn(plugin), reflector)
    wireHooks(opt, plugin, recorder)

    result = opt.Optimize(ctx, seedCandidate, dataset)
    recorder.RecordOptimizeResult(result)
    recorder.Close(true, nil)

    results = append(results, result)

  printSummary(results)
```

### 5.2 Enhanced RecordOptimizeResult

```
recorder.RecordOptimizeResult(res):
  // Existing behavior (preserved)
  for candidate in res.Candidates:
    row.RunID = ...
    row.CandidateID = candidate.ID
    row.ParentID = candidate.ParentID
    row.CandidateHash = candidate.Hash
    row.MeanScore = candidate.GlobalStats.MeanScore
    // ... existing fields ...

    // NEW: lineage and operation fields
    row.Parent2ID = candidate.Parent2ID
    row.Operation = candidate.Operation
    row.UpdatedKeysJSON = json.Marshal(candidate.UpdatedKeys)
    row.CreatedAtMs = candidate.CreatedAt.UnixMilli()
```

### 5.3 Event recording

```
recorder.RecordEvent(event):
  row = eventRow{
    RunID: r.run.RunID,
    Iteration: event.Iteration,
    EventType: string(event.Type),
    Operation: event.Operation,
    ParentID: event.ParentID,
    Parent2ID: event.Parent2ID,
    ChildID: event.ChildID,
    UpdatedKeysJSON: json.Marshal(event.UpdatedKeys),
    ParentScore: event.ParentScore,
    Parent2Score: event.Parent2Score,
    BaselineScore: event.BaselineScore,
    ChildScore: event.ChildScore,
    Accepted: event.Accepted,
    CallsUsed: event.CallsUsed,
    TimestampMs: time.Now().UnixMilli(),
  }
  r.events = append(r.events, row)
```

## 6. Implementation Phases

### Phase 1: Storage enrichment (low risk, high value)

**Files to modify:**
- `cmd/gepa-runner/run_recorder.go` -- Add missing fields to `candidateMetricRow`, add event recording, add schema migration
- `pkg/optimizer/gepa/optimizer.go` -- Expose `ReflectionRaw` on events (it's already on candidateNode)

**Deliverables:**
1. Schema migration to add missing columns
2. `gepa_iteration_events` table
3. `gepa_experiments` table
4. `gepa_schema_version` tracking
5. Event buffer and flush in recorder
6. Tests for migration idempotency

**Estimated scope:** ~300 lines of Go (schema + migration + event recording)

### Phase 2: Experiment YAML config

**Files to create/modify:**
- `cmd/gepa-runner/experiment_config.go` -- YAML parsing and validation
- `cmd/gepa-runner/experiment_command.go` -- `experiment run` command
- `cmd/gepa-runner/main.go` -- Register new command

**Deliverables:**
1. `ExperimentConfig` struct with YAML tags
2. Validation (required fields, enum checks, path resolution)
3. Config-to-optimizer-params mapping
4. Multi-trial loop with seed management
5. `--dry-run` mode
6. Tests for YAML parsing edge cases

**Estimated scope:** ~500 lines of Go

### Phase 3: Synthetic dataset generation

**Files to create/modify:**
- `cmd/gepa-runner/generate_dataset_command.go` -- `generate-dataset` command
- `cmd/gepa-runner/scripts/lib/gepa_dataset_contract.js` -- Generator plugin contract
- `cmd/gepa-runner/scripts/generate_math_dataset.js` -- Example generator

**Deliverables:**
1. Dataset generator plugin contract
2. YAML template engine (simpler alternative)
3. `generate-dataset` command with JSON/JSONL output
4. Example generators for math and text tasks
5. Tests

**Estimated scope:** ~400 lines of Go + ~100 lines of JS

### Phase 4: Analysis commands

**Files to create/modify:**
- `cmd/gepa-runner/analyze_command.go` -- `analyze` subcommands
- Uses Glazed output layer for table/JSON formatting

**Deliverables:**
1. `analyze runs` -- List/filter runs
2. `analyze compare` -- Cross-experiment comparison
3. `analyze trajectory` -- Per-iteration event trace
4. `analyze candidates` -- Candidate lineage tree
5. `analyze summary` -- Aggregate statistics (mean, stddev, min, max)

**Estimated scope:** ~600 lines of Go

### Phase 5: Documentation and examples

1. End-to-end example: generate dataset, write experiment YAML, run trials, analyze results
2. Plugin authoring guide with dataset generator contract
3. SQL query cookbook for advanced analysis
4. Migration notes for existing databases

## 7. Testing Strategy

### 7.1 Unit tests

- Schema migration: test that migrations are idempotent and preserve existing data
- YAML parsing: test valid configs, missing required fields, type mismatches
- Event recording: test that all event fields round-trip through SQLite
- Config resolution: test that YAML fields correctly map to `gepaopt.Config`

### 7.2 Integration tests

- End-to-end experiment run with the `smoke_noop_optimizer.js` plugin
- Multi-trial run with seed verification (same seed = same result)
- Dataset generation + experiment run pipeline
- Analysis queries on recorded data

### 7.3 Regression tests

- Ensure existing `optimize` and `eval` commands work unchanged
- Ensure old databases without new columns still work (migration backcompat)

## 8. Risks, Alternatives, and Open Questions

### 8.1 Risks

1. **Schema migration complexity.** SQLite ALTER TABLE is limited (no DROP COLUMN, no column renames). The migration strategy must be additive only. Mitigated by using default values for all new columns.

2. **Experiment YAML versioning.** If the config format changes, old YAML files may become invalid. Mitigated by the `apiVersion` field and a versioned parser.

3. **Parallel trials.** SQLite does not handle concurrent writes well. If `--parallel > 1`, runs should use WAL mode and serialize writes. Alternative: each trial writes to a temp DB, merged at the end.

### 8.2 Alternatives considered

**Alternative A: JSON config instead of YAML.**
Rejected because YAML is more human-friendly for experiment configs with multi-line strings (prompts, objectives). The Go YAML library (`gopkg.in/yaml.v3`) is already a dependency.

**Alternative B: Separate benchmarking tool.**
Instead of extending `gepa-runner`, build a separate `gepa-bench` binary. Rejected because the integration with the existing plugin system, recorder, and optimizer is tight enough that a separate binary would duplicate significant code.

**Alternative C: Use the existing `--out-report` JSON for analysis.**
Rejected because JSON files are not queryable across runs. SQLite enables cross-run queries, aggregation, and joining that are essential for benchmarking.

**Alternative D: Store events as JSONL files instead of SQLite.**
Considered for simplicity, but rejected because:
- Cross-run queries require loading all files into memory
- SQLite indexes enable fast filtering
- The recorder already uses SQLite

### 8.3 Open questions

1. **Should the experiment command support Geppetto profiles or require explicit engine config?** Currently, the `optimize` command uses Pinocchio profiles (`--profile`). The experiment YAML could either embed engine settings directly or reference a profile name. Direct embedding is more reproducible; profiles are more convenient. Recommendation: support both, prefer direct embedding.

2. **Should analysis commands use Glazed output formatting or raw SQL?** Glazed integration provides table/JSON/CSV output modes for free. Recommendation: use Glazed.

3. **Should we add a `gepa-runner experiment sweep` command for grid/random search?** This would generate experiment YAML variants from a sweep definition. Useful but can be deferred to a later phase.

4. **How should the reflection LLM's raw response be stored in events?** The `ReflectionRaw` string can be large (multiple KB). Storing it in every event row may bloat the database. Options: (a) store only on `_accepted` events, (b) store in a separate table with FK, (c) always store but compress. Recommendation: store on all events since debugging failed mutations is valuable; add a `--compact-recording` flag to skip if desired.

5. **Should the `hostContext` include the full YAML config or just metadata?** Passing the full config lets plugins access any experiment parameter. Passing only metadata keeps the interface clean. Recommendation: pass metadata + explicit `experimentConfig` key for plugins that need it.

## 9. References

### Key source files

| File | Lines | Role |
|------|-------|------|
| `cmd/gepa-runner/main.go` | 428 | CLI entrypoint, optimize command |
| `cmd/gepa-runner/run_recorder.go` | 602 | SQLite recording |
| `cmd/gepa-runner/plugin_loader.go` | 682 | JS plugin loading |
| `cmd/gepa-runner/js_runtime.go` | 76 | Goja runtime setup |
| `cmd/gepa-runner/eval_command.go` | ~150 | Eval-only command |
| `pkg/optimizer/gepa/optimizer.go` | 1185 | Core optimization loop |
| `pkg/optimizer/gepa/config.go` | 130 | Config struct |
| `pkg/optimizer/gepa/types.go` | 67 | Core data types |
| `pkg/optimizer/gepa/reflector.go` | 139 | LLM reflection |
| `pkg/optimizer/gepa/format.go` | 156 | Side-info formatting |
| `pkg/optimizer/gepa/pareto.go` | ~100 | Pareto front |

### Related tickets

- **GP-05-GEPA-PARITY-PLUGIN-RESEARCH** -- Parity analysis and v2 plugin extension architecture
- **GEPA-01-EXTRACT-GEPPETTO-PLUGINS** -- Plugin contract extraction and registry identifiers
- **GP-04-GEPA-MERGE-MULTIPARAM-ALIGN** -- Merge and multi-parameter alignment

### JS plugin contracts

- `cmd/gepa-runner/scripts/lib/gepa_plugin_contract.js` -- Plugin descriptor validation
- `cmd/gepa-runner/scripts/lib/gepa_optimizer_common.js` -- Shared utilities for plugins
