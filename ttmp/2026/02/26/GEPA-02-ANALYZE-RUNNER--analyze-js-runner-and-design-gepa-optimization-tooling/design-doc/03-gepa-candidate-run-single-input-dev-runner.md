---
Title: 'gepa candidate run: single-input dev runner'
Ticket: GEPA-02-ANALYZE-RUNNER
Status: active
Topics:
    - gepa
    - runner
    - goja
    - tooling
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/gepa-runner/dataset.go
      Note: loadSeedCandidateFile reused for candidate resolution
    - Path: cmd/gepa-runner/eval_command.go
      Note: Closest precedent
    - Path: cmd/gepa-runner/plugin_loader.go
      Note: Evaluate() path reused
    - Path: cmd/gepa-runner/run_recorder.go
      Note: ensureRecorderTables pattern reused for new table
ExternalSources: []
Summary: Design for `gepa candidate run` -- a dev tool that takes a candidate + single input, runs it through the plugin once, and stores the raw result.
LastUpdated: 2026-02-26T13:00:00-05:00
WhatFor: Guide implementation of the candidate run command
WhenToUse: When implementing gepa candidate run
---


# `gepa candidate run` -- Single-Input Dev Runner

## 1. What It Is

A dev tool. Takes a candidate (prompt parameters) and a single input, runs the plugin's logic on it once, shows you what happened. No scoring, no datasets, no aggregation. That's what a future `gepa candidate eval` will do.

Think of it like `curl` for GEPA candidates: point it at an input, see the output.

```bash
# Simplest form: one candidate, one input
gepa candidate run \
  --script optimizer.js \
  --candidate candidate.yaml \
  --input '{"question": "2+2", "answer": "4"}'

# Output:
# {
#   "output": {"text": "4"},
#   "candidate": {"prompt": "Solve the math problem..."},
#   "input": {"question": "2+2", "answer": "4"}
# }
```

## 2. Problem

The current `gepa-runner` has no way to run a candidate on a single input and inspect the result. The existing commands both go further than we want:

- **`eval`** (`eval_command.go`) -- always evaluates the *entire* dataset, computes aggregate stats. Cannot run on a single example. Only accepts a single-string prompt, not a multi-parameter candidate.

- **`optimize`** (`main.go`) -- runs the full evolutionary loop. Overkill for "I want to see what this prompt does on this input."

What you end up doing today is writing throwaway JS scripts or shelling out to the LLM directly, losing the plugin's prompt-construction logic.

## 3. The Plugin Problem

The existing plugin contract has `evaluate()` which bundles two things:

```javascript
// toy_math_optimizer.js:30-49
function evaluate(input, options) {
  // Step 1: RUN -- construct prompt, call LLM
  const prompt = `${instruction}\n\nQuestion: ${example.question}\nFinal answer:`;
  const got = common.runUserPrompt(ctx, options, prompt);

  // Step 2: EVAL -- score the output
  const scored = common.exactMatchScore(example.answer, got);
  return { score: scored.score, output: { text: scored.got }, feedback: scored.feedback };
}
```

For `candidate run`, we want step 1 only. Two options:

**Option A: Add a `run()` method to the plugin contract.**
The plugin exposes `run(input, options)` that does the LLM call and returns the raw output without scoring. Plugins that don't implement it fall back to calling `evaluate()` and stripping the score.

**Option B: Call `evaluate()` and just show the output.**
Simpler. The score comes back but we don't display it prominently -- we focus on the output. The plugin doesn't need to change at all.

**Recommendation: Option A with Option B as fallback.** New plugins can provide a clean `run()` method. Old plugins work unchanged via `evaluate()` fallback. The `run()` method is optional.

### New plugin method: `run()`

```javascript
// Plugin instance can optionally implement:
run(input, options) {
  // input: { candidate, example, exampleIndex }
  // Returns: { output, metadata? }
  //   output: the raw LLM response (any shape)
  //   metadata: optional extra info (constructed prompt, model used, latency, etc.)
  const prompt = composePrompt(input.candidate, input.example);
  const got = common.runUserPrompt(ctx, options, prompt);
  return {
    output: { text: got },
    metadata: {
      constructed_prompt: prompt,
    }
  };
}
```

If `run()` is not present on the plugin, fall back to calling `evaluate()` and extracting its `output` field.

## 4. CLI Interface

```
gepa candidate run [flags]

Required:
  --script PATH         JS optimizer plugin

Candidate (one of):
  --candidate PATH      YAML/JSON file with candidate map
  --prompt TEXT         Shorthand for {"prompt": TEXT}
  --prompt-file PATH   Shorthand, read from file

Input (one of):
  --input JSON          Inline JSON example
  --input-file PATH     Example from file

Optional:
  --name TEXT           Label for this run (stored in DB)
  --tags KEY=VAL,...    Arbitrary metadata tags
  --record              Store result in SQLite
  --record-db PATH      SQLite path (default: .gepa-runner/runs.sqlite)
  --output json|yaml|text  Output format (default: json)
  --verbose             Also show constructed prompt and full metadata
```

### Examples

```bash
# Quick test with inline prompt and input
gepa candidate run \
  --script scripts/toy_math_optimizer.js \
  --prompt "Solve step by step. Return only the number." \
  --input '{"question": "6*7", "answer": "42"}'

# Multi-parameter candidate from file
gepa candidate run \
  --script scripts/multi_param_math_optimizer.js \
  --candidate candidates/baseline.yaml \
  --input-file examples/hard_division.json \
  --record --name "baseline-v3"

# Pipe input from stdin
echo '{"question": "100-37"}' | gepa candidate run \
  --script scripts/toy_math_optimizer.js \
  --prompt "Calculate." \
  --input -
```

### Candidate file format

Same as existing `--seed-candidate` in the optimize command:

```yaml
# candidates/baseline.yaml
prompt: |
  Solve the math problem step by step.
  Return only the final numeric answer.
planner_prompt: |
  Before solving, identify the operation type.
critic_prompt: |
  Double-check your arithmetic before answering.
```

Loaded via the existing `loadSeedCandidateFile()` in `dataset.go`.

## 5. Output

### Default (JSON):

```json
{
  "output": {"text": "42"},
  "candidate": {
    "prompt": "Solve the math problem step by step.\nReturn only the final numeric answer."
  },
  "input": {"question": "6*7", "answer": "42"},
  "plugin": {"id": "example.toy_math", "name": "Example: Toy math accuracy"},
  "timestamp": "2026-02-26T13:14:00Z"
}
```

### Verbose (`--verbose`):

```json
{
  "output": {"text": "42"},
  "candidate": {
    "prompt": "Solve the math problem step by step.\nReturn only the final numeric answer."
  },
  "input": {"question": "6*7", "answer": "42"},
  "plugin": {"id": "example.toy_math", "name": "Example: Toy math accuracy"},
  "metadata": {
    "constructed_prompt": "Solve the math problem...\n\nQuestion: 6*7\nFinal answer:",
    "run_method": "run",
    "latency_ms": 1250
  },
  "timestamp": "2026-02-26T13:14:00Z"
}
```

### Text (`--output text`):

```
42
```

Just the raw output text. Useful for piping: `gepa candidate run ... --output text | pbcopy`.

### When falling back to `evaluate()`:

If the plugin doesn't have `run()`, we call `evaluate()` and reshape the result:

```json
{
  "output": {"text": "42"},
  "eval_result": {
    "score": 1.0,
    "feedback": "Correct.",
    "trace": {}
  },
  "candidate": {},
  "input": {},
  "plugin": {},
  "metadata": {
    "run_method": "evaluate_fallback"
  }
}
```

The `eval_result` is included since `evaluate()` computed it anyway, but it's secondary -- the focus is `output`.

## 6. SQLite Storage

### New table: `gepa_candidate_runs`

Dedicated table for individual run results. Separate from the optimizer's `gepa_runs`/`gepa_eval_examples` tables because the semantics are different -- this is one candidate, one input, one execution.

```sql
CREATE TABLE IF NOT EXISTS gepa_candidate_runs (
  run_id TEXT PRIMARY KEY,
  timestamp_ms INTEGER NOT NULL,
  plugin_id TEXT,
  plugin_name TEXT,
  candidate_name TEXT,          -- user-provided --name
  candidate_json TEXT NOT NULL, -- full candidate map
  input_json TEXT NOT NULL,     -- the input example
  output_json TEXT NOT NULL,    -- what came back
  metadata_json TEXT,           -- constructed prompt, latency, etc.
  tags_json TEXT DEFAULT '{}',  -- --tags key=val pairs
  run_method TEXT NOT NULL,     -- "run" or "evaluate_fallback"
  eval_score REAL,              -- only set if evaluate_fallback was used
  eval_result_json TEXT,        -- full eval result if evaluate_fallback
  profile TEXT,                 -- geppetto profile used
  engine_options_json TEXT,     -- engine config snapshot
  duration_ms INTEGER
);

CREATE INDEX IF NOT EXISTS idx_candidate_runs_timestamp
  ON gepa_candidate_runs (timestamp_ms DESC);
CREATE INDEX IF NOT EXISTS idx_candidate_runs_plugin
  ON gepa_candidate_runs (plugin_id);
CREATE INDEX IF NOT EXISTS idx_candidate_runs_name
  ON gepa_candidate_runs (candidate_name);
```

### Run ID format

```
gepa-run-{unix_nano}
```

### What gets stored

Every `--record` invocation writes one row:

| Column | Source |
|--------|--------|
| `run_id` | Generated |
| `timestamp_ms` | `time.Now()` |
| `plugin_id` | Plugin descriptor `.id` |
| `plugin_name` | Plugin descriptor `.name` |
| `candidate_name` | `--name` flag (nullable) |
| `candidate_json` | JSON of the candidate map |
| `input_json` | JSON of the input example |
| `output_json` | JSON of the plugin's output |
| `metadata_json` | Constructed prompt, latency, run method |
| `tags_json` | `--tags` as JSON object |
| `run_method` | `"run"` or `"evaluate_fallback"` |
| `eval_score` | Score from evaluate() if fallback was used |
| `eval_result_json` | Full EvalResult if fallback |
| `profile` | Geppetto profile name |
| `engine_options_json` | Engine config snapshot |
| `duration_ms` | Wall-clock time for the plugin call |

### Querying stored runs

```sql
-- Recent runs
SELECT run_id, candidate_name, plugin_id,
       json_extract(output_json, '$.text') as output_text,
       duration_ms
FROM gepa_candidate_runs
ORDER BY timestamp_ms DESC LIMIT 10;

-- Runs for a specific candidate
SELECT * FROM gepa_candidate_runs
WHERE candidate_name = 'baseline-v3';

-- Filter by tags
SELECT * FROM gepa_candidate_runs
WHERE json_extract(tags_json, '$.experiment') = 'merge-comparison';
```

## 7. Execution Flow

```
1. Parse flags
2. Load JS plugin (existing loadOptimizerPlugin)
3. Resolve candidate:
   --candidate file -> loadSeedCandidateFile()
   --prompt text -> {"prompt": text}
   --prompt-file -> read file -> {"prompt": text}
4. Resolve input:
   --input json -> json.Unmarshal
   --input-file -> read + unmarshal
   --input - -> read stdin + unmarshal
5. Start timer
6. Check if plugin has run() method
7. If run():
   call plugin.Run(candidate, 0, input, opts)
   result = {output, metadata}
8. Else:
   call plugin.Evaluate(candidate, 0, input, opts)
   result = {output: evalResult.Output, eval_result: evalResult}
9. Stop timer
10. Print result in requested format
11. If --record:
    write row to gepa_candidate_runs
```

## 8. Plugin Loader Changes

Add `run()` detection alongside existing optional methods in `plugin_loader.go`:

```go
type optimizerPlugin struct {
    // ... existing fields ...
    runFn goja.Callable  // NEW: optional run() method
}
```

In `loadOptimizerPlugin()`, after existing optional method detection:

```go
runFn := findOptionalCallable(instanceObj, "run", "runCandidate")
```

New method on `optimizerPlugin`:

```go
type RunResult struct {
    Output   any            `json:"output"`
    Metadata map[string]any `json:"metadata,omitempty"`
}

func (p *optimizerPlugin) HasRun() bool {
    return p != nil && p.runFn != nil
}

func (p *optimizerPlugin) Run(
    candidate gepaopt.Candidate,
    exampleIndex int,
    example any,
    opts pluginEvaluateOptions,
) (RunResult, error) {
    input := map[string]any{
        "candidate":    candidate,
        "example":      example,
        "exampleIndex": exampleIndex,
    }
    options := map[string]any{
        "profile":       opts.Profile,
        "engineOptions": opts.EngineOptions,
        "tags":          opts.Tags,
    }
    ret, err := p.runFn(p.instance, p.rt.vm.ToValue(input), p.rt.vm.ToValue(options))
    if err != nil {
        return RunResult{}, errors.Wrap(err, "plugin run: call failed")
    }
    decoded, err := decodeJSReturnValue(ret)
    if err != nil {
        return RunResult{}, errors.Wrap(err, "plugin run: invalid return")
    }
    return decodeRunResult(decoded)
}

func decodeRunResult(v any) (RunResult, error) {
    switch x := v.(type) {
    case string:
        return RunResult{Output: map[string]any{"text": x}}, nil
    case map[string]any:
        r := RunResult{Output: x["output"]}
        if md, ok := x["metadata"].(map[string]any); ok {
            r.Metadata = md
        }
        if r.Output == nil {
            r.Output = x // treat entire return as output
        }
        return r, nil
    default:
        return RunResult{Output: v}, nil
    }
}
```

## 9. Implementation Plan

Three pieces, all in one phase:

**A. Plugin loader extension** (~40 lines in `plugin_loader.go`)
- Detect optional `run()`/`runCandidate` on plugin instance
- `Run()` method + `decodeRunResult()`

**B. Command** (~200 lines, new file `candidate_run_command.go`)
- Glazed command with flags from section 4
- Candidate resolution (reuse `loadSeedCandidateFile`, `resolveSeedText`)
- Input resolution (inline JSON, file, stdin)
- Plugin dispatch: `run()` or `evaluate()` fallback
- Output formatting (json, yaml, text)
- Register as `rootCmd.AddCommand` in `main.go`

**C. SQLite recording** (~150 lines, new file `candidate_run_recorder.go`)
- `gepa_candidate_runs` table creation
- `candidateRunRecorder` struct with `Record()` and `Close()`
- Uses same SQLite file as existing recorder

**Total: ~390 lines of Go.**

### Tests

- Inline prompt + inline input with `smoke_noop_optimizer.js`
- Multi-parameter candidate from file with `multi_param_math_optimizer.js`
- Stdin input piping
- SQLite round-trip: record + query back
- Fallback to `evaluate()` when no `run()` method

## 10. What This Is Not

- **Not an evaluator.** No scoring, no datasets, no aggregation. That's `gepa candidate eval` (future).
- **Not an optimizer.** No loop, no mutation, no reflection. That's `gepa optimize`.
- **Not a batch tool.** One candidate, one input, one run. For N inputs, shell loop or future eval command.

## 11. Open Questions

1. **Should `run()` receive a Turn (geppetto conversation) instead of just an input?** Probably not for v1 -- the input is an opaque example object and the plugin constructs whatever it needs internally.

2. **Should `--output text` extract `.text` from the output automatically?** If `output` is `{"text": "42"}`, should it print `42`? Recommendation: yes, try `output.text`, then `output.response`, then `JSON.stringify(output)`.

3. **Should we add `--dry-run` to show the constructed prompt without calling the LLM?** Would require the plugin to expose prompt construction separately. Defer.

## 12. References

| File | Role |
|------|------|
| `cmd/gepa-runner/plugin_loader.go:206-238` | Existing `Evaluate()` call path |
| `cmd/gepa-runner/plugin_loader.go:129-143` | `findOptionalCallable()` pattern |
| `cmd/gepa-runner/eval_command.go` | Closest existing command (does too much) |
| `cmd/gepa-runner/dataset.go` | `loadSeedCandidateFile()`, `resolveSeedText()` |
| `cmd/gepa-runner/run_recorder.go:434-502` | `ensureRecorderTables()` pattern |
| `scripts/toy_math_optimizer.js:30-49` | Plugin evaluate with run+score bundled |
| `scripts/multi_param_math_optimizer.js:47-95` | Multi-param evaluate example |
| `scripts/lib/gepa_optimizer_common.js:49-60` | `runUserPrompt()` -- the actual LLM call |
