# gepa-runner

`gepa-runner` runs a GEPA-style reflective optimization loop on top of:

- Geppetto inference/runtime
- JavaScript plugins loaded via go-go-gepa module helpers (`require("gepa/plugins")`)

Current implementation includes:

- reflection-based mutation
- optional merge/crossover
- multi-parameter candidate optimization
- Pareto-aware parent selection (when evaluator returns multiple objectives)
- optional SQLite run recording

## Quick start

```bash
gepa-runner optimize \
  --script ./cmd/gepa-runner/scripts/toy_math_optimizer.js \
  --seed "Answer the question. Respond with only the final answer." \
  --max-evals 200 \
  --batch-size 8 \
  --out-prompt best_prompt.txt \
  --out-report run_report.json \
  --profile 4o-mini
```

Dataset can come from plugin `dataset()` or a file:

```bash
gepa-runner optimize \
  --script ./my_optimizer.js \
  --dataset ./data/train.jsonl \
  --seed-file ./seed_prompt.txt
```

## Included example scripts

Under `cmd/gepa-runner/scripts/`:

- `toy_math_optimizer.js`
  - baseline arithmetic optimizer, now using shared helper library
- `multi_param_math_optimizer.js`
  - multi-parameter candidate (`prompt`, `planner_prompt`, `critic_prompt`)
  - plugin-side component selection and per-component side-info shaping
- `seedless_heuristic_merge_optimizer.js`
  - seedless initialization via `initialCandidate()`
  - non-LLM heuristic merge callback
- `optimize_anything_style_optimizer.js`
  - component-metadata adapter style inspired by optimize-anything patterns
  - multi-objective scoring + component-aware hooks
- `smoke_noop_optimizer.js`
  - minimal smoke plugin

Shared JS utilities are in:

- `cmd/gepa-runner/scripts/lib/gepa_optimizer_common.js`
- `cmd/gepa-runner/scripts/arithmetic_dataset_generator.js` (dataset-generate example)

## Optimize flags (important)

Core:

- `--script` JS plugin path (required)
- `--dataset` optional JSON/JSONL dataset file
- `--max-evals` evaluator call budget
- `--batch-size` minibatch size
- `--objective` optional natural-language objective prefix for reflection/merge prompts

Seeding:

- `--seed` prompt text
- `--seed-file` prompt file
- `--seed-candidate` JSON/YAML object map for multi-param seed
- `--seedless` use plugin `initialCandidate()` when no seed/seed-file/seed-candidate is provided

Merge / scheduler:

- `--merge-prob` merge attempt probability
- `--merge-scheduler` `probabilistic` (default) or `stagnation_due`
- `--max-merges-due` cap for internal due counter (`stagnation_due` mode)

Multi-param:

- `--optimizable-keys` comma-separated candidate keys to optimize
- `--component-selector` `round_robin` (default) or `all`

Observability / outputs:

- `--show-events` print mutate/merge attempted/accepted/rejected events
- `--out-prompt` write best candidate `prompt` key
- `--out-report` write full JSON result
- `--record` persist run metrics to SQLite
- `--record-db` SQLite path (default: `.gepa-runner/runs.sqlite`)

## Multi-parameter example

`seed-candidate.yaml`:

```yaml
prompt: |
  Solve the task and return final answer only.
planner_prompt: |
  Produce a short plan before solving.
critic_prompt: |
  Identify likely mistakes and verify output.
```

Run:

```bash
gepa-runner optimize \
  --script ./my_optimizer.js \
  --dataset ./data/train.jsonl \
  --seed-candidate ./seed-candidate.yaml \
  --optimizable-keys prompt,planner_prompt,critic_prompt \
  --component-selector round_robin
```

## Seedless mode example

Plugin provides `initialCandidate()` and run uses `--seedless`:

```bash
gepa-runner optimize \
  --script ./my_optimizer.js \
  --dataset ./data/train.jsonl \
  --seedless
```

If `initialCandidate()` is missing or empty, command fails explicitly.

## JS plugin contract

Plugin descriptor:

```js
const plugins = require("gepa/plugins");

module.exports = plugins.defineOptimizerPlugin({
  apiVersion: plugins.OPTIMIZER_PLUGIN_API_VERSION,
  kind: "optimizer",
  id: "my.task",
  name: "My Task",
  create(ctx) {
    return {
      evaluate(input, options) {
        return { score: 0.0 };
      }
    };
  }
});
```

## Dataset generate (building block)

Generate synthetic rows with a dataset-generator script and v2 config:

```bash
gepa-runner dataset generate \
  --script ./cmd/gepa-runner/scripts/arithmetic_dataset_generator.js \
  --config ./dataset-generate.yaml \
  --count 10 \
  --output-dir ./tmp/generated \
  --output-db ./tmp/generated/datasets.sqlite
```

Config YAML (`gepa.dataset-generate/v2`) only defines generation semantics.
Output routing (`--output-dir`, `--output-db`, etc.) is CLI-only.

### Required hook

- `evaluate(input, options) -> object | number`

`input` fields:

- `candidate` map of strings
- `example` dataset item
- `exampleIndex` index of example

Return:

- required: `score` (number; higher is better)
- optional:
  - `objectiveScores` or `objectives` map of numbers
  - `output`, `feedback`, `trace`, `notes` / `evaluatorNotes`

### Optional hooks

- `dataset() -> array`
  - used when `--dataset` is not provided

- `merge(input, options) -> string | object`
  - aliases recognized: `mergeCandidate`, `mergePrompt`
  - `input` includes `candidateA`, `candidateB`, `paramKey`, `paramA`, `paramB`, `sideInfoA`, `sideInfoB`

- `initialCandidate(options) -> string | object`
  - alias recognized: `getInitialCandidate`
  - used by `--seedless`

- `selectComponents(input, options) -> string | string[]`
  - alias recognized: `chooseComponents`
  - `input` includes `operation`, `parentId`, `parent2Id`, `candidate`, `availableKeys`, `nextParamIndex`

- `componentSideInfo(input, options) -> string | object`
  - aliases recognized: `sideInfoForComponent`, `buildSideInfo`
  - `input` includes `operation`, `paramKey`, `examples`, `evals`, `maxChars`, `default`

`options` fields passed to hooks:

- `profile`
- `engineOptions`
- `tags`

## Async plugin methods and stream events

`gepa-runner` now supports Promise-returning plugin hooks in addition to synchronous returns.

Supported Promise-aware hooks include:

- optimizer path: `run`, `evaluate`, `merge`, `initialCandidate`, `selectComponents`, `componentSideInfo`, `dataset`
- dataset path: `generateOne`

### Event emission hooks

For hooks that receive `options`, plugins can emit stream events while asynchronous work is running:

- `options.emitEvent(payload)`
- `options.events.emit(payload)`

When command flag `--stream` is enabled (`candidate run` / `dataset generate`), events are printed as they arrive:

```text
stream-event {"kind":"plugin_stream","command":"candidate_run","event":{...}}
```

### Candidate run async example

```bash
gepa-runner candidate run \
  --script ./candidate-plugin.js \
  --config ./candidate-config.yaml \
  --input-file ./candidate-input.json \
  --stream \
  --output-format json
```

### Dataset generate async example

```bash
gepa-runner dataset generate \
  --script ./dataset-plugin.js \
  --config ./dataset-config.yaml \
  --stream \
  --dry-run
```

See built-in Glazed help pages for full examples and troubleshooting:

- `gepa-runner help gepa-runner-async-plugin-contract`
- `gepa-runner help gepa-runner-candidate-run-streaming-example`
- `gepa-runner help gepa-runner-dataset-generate-streaming-example`

## Merge return decoding rules

`merge(...)` can return:

- a string (merged text)
- an object with one of:
  - `<paramKey>`
  - `prompt`
  - `merged`
  - `mergedPrompt`
  - `text`
- or `{ candidate: { <paramKey>: "..." } }`

## Event stream

With `--show-events`, `optimize` prints one line per event, e.g.:

```text
[event] iter=7 type=merge_attempted op=merge parent=4 parent2=2 child=9 accepted=false baseline=0.812500 child=0.790000 keys=prompt
```

Event types:

- `mutate_attempted`, `mutate_accepted`, `mutate_rejected`
- `merge_attempted`, `merge_accepted`, `merge_rejected`

## Recorded runs

Persist optimize/eval metrics:

```bash
gepa-runner optimize \
  --script ./cmd/gepa-runner/scripts/toy_math_optimizer.js \
  --seed "ok seed" \
  --max-evals 8 \
  --batch-size 2 \
  --record \
  --record-db ./tmp/gepa-runs.sqlite
```

Inspect:

```bash
gepa-runner eval-report --db ./tmp/gepa-runs.sqlite --limit-runs 20 --format table
gepa-runner eval-report --db ./tmp/gepa-runs.sqlite --limit-runs 20 --format json
```
