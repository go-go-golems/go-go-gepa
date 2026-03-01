---
Title: Merge + Multi-Param Integration and Python GEPA Alignment Study
Ticket: GP-04-GEPA-MERGE-MULTIPARAM-ALIGN
Status: active
Topics:
    - architecture
    - migration
    - tools
    - geppetto
    - inference
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: gepa/src/gepa/adapters/optimize_anything_adapter/optimize_anything_adapter.py
      Note: Python adapter handling of per-component reflective datasets
    - Path: gepa/src/gepa/core/engine.py
      Note: |-
        Python execution order and merge scheduling/acceptance behavior
        Python merge scheduling and acceptance flow
    - Path: gepa/src/gepa/optimize_anything.py
      Note: |-
        Python reference API and config topology (optimize_anything)
        Python optimize_anything config and API reference
    - Path: gepa/src/gepa/proposer/merge.py
      Note: |-
        Python merge proposer logic used for alignment analysis
        Python merge proposer behavior reference
    - Path: gepa/src/gepa/strategies/component_selector.py
      Note: Python component selector strategies (round_robin, all)
    - Path: gepa/ttmp/2026/02/23/GP-04-GEPA-MERGE-MULTIPARAM-ALIGN--integrate-gepa-merge-and-multi-parameter-optimization-in-go-go-gepa-and-align-with-python-original/sources/01-imported-merge-multiparam.patch
      Note: |-
        Canonical patch artifact for upstream merge+multi-param commit delta
        Primary upstream patch artifact
    - Path: geppetto/pkg/js/modules/geppetto/plugins_module.go
      Note: |-
        Shared JS optimizer plugin descriptor contract helper
        Shared optimizer plugin descriptor helper contract
    - Path: go-go-gepa/cmd/gepa-runner/main.go
      Note: |-
        Current CLI surface; currently single-param seed prompt only
        Runner CLI integration target
    - Path: go-go-gepa/cmd/gepa-runner/plugin_loader.go
      Note: |-
        Current JS plugin contract loader; currently no merge callback support
        JS plugin merge callback integration target
    - Path: go-go-gepa/pkg/optimizer/gepa/config.go
      Note: Current config surface used by go-go-gepa optimizer
    - Path: go-go-gepa/pkg/optimizer/gepa/optimizer.go
      Note: |-
        Current optimizer core in go-go-gepa; missing merge and multi-parameter machinery
        Current optimizer baseline for gap analysis
    - Path: imported/geppetto-main/cmd/gepa-runner/main.go
      Note: Upstream CLI additions for seed-candidate, merge-prob, optimizable-keys, component-selector
    - Path: imported/geppetto-main/cmd/gepa-runner/plugin_loader.go
      Note: Upstream merge callback hook shape for JS plugins
    - Path: imported/geppetto-main/pkg/optimizer/gepa/optimizer.go
      Note: |-
        Upstream commit with merge + multi-param implementation to integrate
        Upstream merge and multi-parameter implementation reference
ExternalSources: []
Summary: Deep architectural study and implementation blueprint for integrating merge and multi-parameter optimization into go-go-gepa while aligning with Python GEPA and optimize_anything capabilities.
LastUpdated: 2026-02-23T11:05:00-05:00
WhatFor: Guide the integration work with source-grounded comparisons, explicit design choices, and phased delivery tasks.
WhenToUse: Use before and during implementation of merge/multi-param support and when reviewing parity with Python GEPA behavior.
---


# Merge + Multi-Param Integration and Python GEPA Alignment Study

## 1. Executive Summary

This study answers three concrete engineering questions:

1. What exactly changed in the upstream imported Go GEPA branch (`imported/geppetto-main`, commit `7b488b9`) for merge and multi-parameter optimization?
2. How does our current `go-go-gepa` implementation compare to the Python GEPA reference implementation (`gepa/src/gepa/*`), especially `optimize_anything`?
3. How should we integrate merge + multi-param in `go-go-gepa` in a way that preserves our JS plugin ergonomics while borrowing the strongest patterns from Python (component-aware reflection, merge scheduling, and objective-aware selection)?

The short answer:

- `go-go-gepa` is currently a solid single-parameter reflective mutation optimizer with Pareto-aware parent choice, but it is missing multi-parameter evolution primitives and merge/crossover primitives that exist both in upstream imported Go and in Python GEPA.
- The upstream Go patch (`7b488b9`) is a practical, incremental bridge toward Python behavior: it introduces `MergeProbability`, `OptimizableKeys`, component selectors (`round_robin`/`all`), seed-candidate map loading, and optional JS merge callbacks.
- Python GEPA (`optimize_anything`) is still architecturally richer than both Go variants: explicit proposer composition, merge scheduling state machine, richer callback/event model, and stronger adapter-centric component handling.

Recommended strategy:

- Integrate upstream Go merge + multi-param changes into `go-go-gepa` in a first pass (low-risk parity lift).
- In a second pass, evolve Go toward an `optimize_anything`-style envelope by enriching plugin interfaces and side-info/component semantics rather than trying to copy Python internals one-to-one.

## 2. Scope, Inputs, and Method

### 2.1 In-scope repositories

- `go-go-gepa` (target implementation)
- `imported/geppetto-main` (source of upstream merge/multi-param patch)
- `gepa` Python repo (reference behavior)
- `geppetto` (for JS plugin contract helper surface)

### 2.2 Ground-truth artifacts used

Primary evidence is captured in ticket `sources/`:

- `sources/01-imported-merge-multiparam.patch` (full upstream patch)
- `sources/02-*.diff` (pairwise file diffs: upstream vs `go-go-gepa`)
- `sources/05..10` (Python API/config/engine/merge excerpts)
- `sources/11..17` (Go current/upstream excerpts)
- `sources/18..20` (Python component selector and reflective dataset behavior)
- `sources/21..23` (JS plugin contract and loader comparisons)

### 2.3 Comparison method

The analysis uses a layered comparison:

1. Interface layer: config fields, CLI flags, plugin method contracts.
2. Control-flow layer: mutation loop, merge scheduling, acceptance criteria.
3. Data model layer: candidate representation, per-component update bookkeeping, side-info format.
4. Extensibility layer: how easy it is to express `optimize_anything` patterns in JS plugins.

## 3. Current `go-go-gepa` Architecture (Baseline)

### 3.1 Core optimizer model

Current optimizer structure in `go-go-gepa/pkg/optimizer/gepa/optimizer.go:19` is single-parent and single-component-per-iteration by behavior:

- `Optimizer` stores evaluator, reflector, cache, pool, RNG.
- `candidateNode` tracks one `ParentID`, candidate map, evaluations, reflection raw text.
- Mutation targets only one key via `primaryParamKey` (`"prompt"` preferred, fallback first key), see `go-go-gepa/pkg/optimizer/gepa/optimizer.go:158` and `go-go-gepa/pkg/optimizer/gepa/optimizer.go:159`.

Conceptually:

```text
select parent -> evaluate parent minibatch -> reflect on single param -> evaluate child -> accept/reject
```

There is no second parent, no component set selection per iteration, and no merge proposer hook.

### 3.2 Current config surface

`go-go-gepa/pkg/optimizer/gepa/config.go:6` includes:

- `MaxEvalCalls`, `BatchSize`, `FrontierSize`, `RandomSeed`
- reflection prompt/system fields
- `Objective`, `MaxSideInfoChars`, `Epsilon`

It does **not** include:

- merge probability/template/system
- optimizable keys list
- component selector mode

### 3.3 Current runner/plugin contract

`go-go-gepa/cmd/gepa-runner/main.go:43` currently accepts classic flags for script, dataset, seed text, objective, eval budget, record/report outputs.

Notably absent in runner surface:

- `--seed-candidate`
- `--merge-prob`
- `--optimizable-keys`
- `--component-selector`

Plugin loader in `go-go-gepa/cmd/gepa-runner/plugin_loader.go:29` currently binds only `evaluate` and optional `dataset`; merge hooks are absent.

### 3.4 Strengths of current baseline

- Clear and maintainable flow.
- Good cache behavior and stagnation guard (`callsUsed` unchanged + rejected child break condition).
- Deterministic and testable single-lane optimization loop.

### 3.5 Baseline limitations

- Component coupling can’t be modeled (e.g., system prompt + rubric prompt co-evolution).
- No crossover across Pareto-specialized candidates.
- JS plugins cannot inject custom merge semantics.

## 4. What Upstream Imported Go Added (`7b488b9`)

Upstream commit summary is explicit: `:art: Add merge and multi-param`.

Changed files (`imported/geppetto-main`):

- `pkg/optimizer/gepa/config.go`
- `pkg/optimizer/gepa/format.go`
- `pkg/optimizer/gepa/optimizer.go`
- `pkg/optimizer/gepa/reflector.go`
- `cmd/gepa-runner/main.go`
- `cmd/gepa-runner/dataset.go`
- `cmd/gepa-runner/plugin_loader.go`
- `cmd/gepa-runner/scripts/toy_math_optimizer.js`
- `cmd/gepa-runner/README.md`

### 4.1 Config-level additions

`imported/geppetto-main/pkg/optimizer/gepa/config.go:28` onward adds:

- `MergeProbability float64`
- `MergeSystemPrompt string`
- `MergePromptTemplate string`
- `OptimizableKeys []string`
- `ComponentSelector string` with `round_robin` and `all`

This immediately upgrades the control plane from single-param implicit to explicit multi-component strategy configuration.

### 4.2 Optimizer-level additions

`imported/geppetto-main/pkg/optimizer/gepa/optimizer.go:18` onward adds core merge/multi-param mechanics:

- `MergeInput` and optional `MergeFunc` hook (`SetMergeFunc`).
- Candidate node ancestry extension (`Parent2ID`), `Operation` tag (`seed|mutate|merge`), per-key update lineage (`LastUpdated`, `UpdatedKeys`), component pointer (`NextParamIndex`).
- Key-derivation and validation (`deriveOptimizableKeys`).
- Component selection per iteration (`selectComponents`).
- Merge path:
  - optional second parent selection
  - system-aware component merge (`systemAwareMerge`)
  - per-component merge proposal using merge function or reflector merge template
  - acceptance against best parent in current batch

This is the biggest structural step toward Python GEPA behavior while preserving the existing compact optimizer style.

### 4.3 Runner-level additions

`imported/geppetto-main/cmd/gepa-runner/main.go:44` onward adds:

- seed candidate map file support (`--seed-candidate`, JSON/YAML)
- merge probability (`--merge-prob`)
- key restriction (`--optimizable-keys`)
- component selector mode (`--component-selector`)
- runtime merge hook registration from plugin loader (`plugin.HasMerge()`, `plugin.Merge(...)`)

### 4.4 Plugin loader additions

`imported/geppetto-main/cmd/gepa-runner/plugin_loader.go` adds optional merge entrypoint detection:

- accepted plugin function names: `merge`, `mergeCandidate`, `mergePrompt`
- normalization/decoding of merge output as string or keyed object

This keeps JS authorship simple: plugin authors can add merge behavior incrementally.

### 4.5 Upstream patch quality observations

Positive:

- Practical compatibility with existing plugin model.
- Minimal invasive change set.
- Config and runtime flow are coherent.

Cautions:

- Semantics are intentionally “GEPA-inspired”, not full parity with Python merge scheduler/callback semantics.
- Some behavior is heuristic (e.g., merge acceptance and component selection state behavior) and should be covered with stronger tests in `go-go-gepa`.

## 5. Python GEPA (`optimize_anything`) Architecture Highlights

Python’s design is more decomposed and strategy-driven.

### 5.1 Config topology is explicit and compositional

In `gepa/src/gepa/optimize_anything.py`:

- `ReflectionConfig` (`:694`) defines component selectors, minibatch behavior, reflection LM, prompt templates.
- `MergeConfig` (`:720`) controls merge budget (`max_merge_invocations`) and validation overlap floor.
- `GEPAConfig` (`:790`) composes `engine`, `reflection`, `tracking`, optional `merge`, optional `refiner`.

This encourages evolution of each subsystem independently.

### 5.2 `optimize_anything` API is artifact-centric

`optimize_anything(...)` (`:998`) supports:

- string candidate mode
- dict candidate mode (multi-component)
- seedless mode (`seed_candidate=None`) where initial candidate is synthesized from objective/background

It’s an “artifact optimizer API”, not just a prompt loop API.

### 5.3 Reflective mutation is component-aware by design

In `gepa/src/gepa/proposer/reflective_mutation/reflective_mutation.py`:

- component selector yields `components_to_update`
- adapter builds component-specific reflective datasets
- proposer emits per-component updated texts

This naturally supports coupled multi-parameter programs.

### 5.4 Merge is a first-class proposer with its own scheduler

In `gepa/src/gepa/proposer/merge.py:210` and `gepa/src/gepa/core/engine.py:406`:

- merge attempts are explicitly scheduled (`merges_due`, `last_iter_found_new_program`)
- merge candidates are drawn from Pareto-dominator structure
- merge evaluates subsampled overlap-aware ids
- engine decides accept/reject with callback hooks and state updates

The merge flow is not just “random crossover”; it is a planned proposer lane.

### 5.5 Callback/event model gives observability

Python engine triggers event callbacks for:

- candidate selection
- evaluation start/end
- merge attempted/accepted/rejected
- iteration lifecycle

This is useful for debugging and experiment tracking in advanced workflows.

## 6. Alignment Matrix: `go-go-gepa` vs Upstream Go vs Python

### 6.1 Feature parity snapshot

| Capability | go-go-gepa (current) | imported Go (`7b488b9`) | Python GEPA |
|---|---|---|---|
| Single-param reflective mutation | Yes | Yes | Yes |
| Multi-component candidate updates | No (implicit one key) | Yes (`OptimizableKeys`, selector) | Yes (adapter+selector) |
| Merge/crossover primitive | No | Yes (probabilistic + system-aware) | Yes (proposer-driven) |
| Plugin-provided merge callback | No | Yes | N/A (Python adapters/proposers) |
| Seed candidate map file | No | Yes | Yes (dict seed and seedless mode) |
| Merge scheduling state machine | No | Partial | Yes |
| Callback-rich observability | Minimal | Minimal | Strong |
| Refiner loop (post-eval refinement) | No | No | Optional component |

### 6.2 Interpretation

- Upstream Go closes the highest-impact functionality gaps for day-to-day prompt optimization.
- Python remains the architectural north star for long-term extensibility.
- Best path is two-stage: parity-first (imported Go features), then architecture-lift (Python-inspired proposer/adapter layering).

## 7. Detailed Integration Design for `go-go-gepa`

### 7.1 Design principles

1. Preserve existing runner UX where possible.
2. Add merge/multi-param as additive extensions.
3. Keep plugin contract backward-compatible.
4. Converge semantics toward Python where it improves correctness/traceability.

### 7.2 Proposed package-level changes (`pkg/optimizer/gepa`)

Add from upstream with targeted hardening:

- Config fields from imported `config.go`.
- Optimizer structs/functions from imported `optimizer.go`:
  - `MergeInput`, `MergeFunc`, `SetMergeFunc`
  - `selectParentDistinct`, `deriveOptimizableKeys`, `selectComponents`, `systemAwareMerge`, `proposeMerge`
- Reflector merge template support from imported `reflector.go`/`format.go`.

Hardening we should add during integration:

- Additional tests for `deriveOptimizableKeys` error paths and round_robin pointer semantics.
- Test for merge acceptance baseline (`max(parentA,parentB)`) and no-budget behavior.
- Regression tests for `Operation` and ancestry metadata in `Result.Candidates`.

### 7.3 Proposed runner changes (`cmd/gepa-runner`)

Bring over:

- flags: `--seed-candidate`, `--merge-prob`, `--optimizable-keys`, `--component-selector`
- loading/parsing map seeds from JSON/YAML in `dataset.go`
- merge hook registration via plugin loader

Runner behavior specifics:

- keep existing record/report flow from `go-go-gepa` (already present)
- use full `BestCandidate` when `prompt` key absent (already solved in imported runner and should be preserved)

### 7.4 Proposed plugin contract extensions

Current plugin instance contract in go:

- required: `evaluate(input, options)`
- optional: `dataset()`

Proposed additive extension:

- optional merge callback: one of
  - `merge(input, options)`
  - `mergeCandidate(input, options)`
  - `mergePrompt(input, options)`

Merge input shape (recommended):

```json
{
  "candidateA": {"prompt": "...", "critic_prompt": "..."},
  "candidateB": {"prompt": "...", "critic_prompt": "..."},
  "paramKey": "critic_prompt",
  "paramA": "...",
  "paramB": "...",
  "sideInfoA": "...",
  "sideInfoB": "..."
}
```

Accepted outputs:

- string
- object with `<paramKey>` or fallback keys (`prompt`, `mergedPrompt`, etc.)
- object with nested `candidate.<paramKey>`

This mirrors imported logic and preserves plugin ergonomics.

## 8. Bringing `optimize_anything` Strengths into JS Plugin Workflows

The most valuable Python ideas are not Python-specific; they are control-structure patterns.

### 8.1 Pattern A: Component-centric reflective dataset construction

Python adapter strategy:

- build per-component reflective datasets
- each component sees focused feedback slice

JS equivalent design:

- allow plugin-side helper to emit per-key side-info blocks
- optimizer passes key-specific side-info into mutate/merge callbacks

Pseudo-interface:

```javascript
module.exports = defineOptimizerPlugin({
  create(ctx) {
    return {
      evaluate(input, options) {
        // return score + objectives + side_info_by_key
        return {
          score: 0.62,
          objectives: { accuracy: 0.62, cost: -0.04 },
          sideInfoByKey: {
            prompt: "...",
            critic_prompt: "..."
          }
        };
      }
    };
  }
});
```

### 8.2 Pattern B: Merge scheduler semantics

Python separates:

- “merge due?” scheduling
- “merge candidate generation”
- “acceptance decision”

Go can mimic this without full proposer framework by adding explicit merge scheduler state to optimizer loop, e.g.:

```text
if merge_due && last_iter_accepted:
  attempt merge
  if accepted:
    consume merge_due
    continue
  else:
    keep merge_due (or decrement by policy)
# then run reflective mutation
```

### 8.3 Pattern C: Better objective-aware parent selection for merges

Python uses valset overlap and ancestor constraints.

Near-term Go adaptation:

- keep current simple parent selection for parity
- add optional policy hook for merge parent pair selection later

### 8.4 Pattern D: Seedless mode and artifact synthesis

`optimize_anything` supports seedless start (`seed_candidate=None`).

Potential Go/JS adaptation:

- `--seedless` flag requiring objective/background
- initialize candidate via reflection LLM before first evaluation

This is a later phase; not required for merge/multi-param parity.

## 9. Control-Flow Diagrams

### 9.1 Current `go-go-gepa` loop

```text
+----------------+
| Select parent  |
+--------+-------+
         |
         v
+-----------------------------+
| Evaluate parent on minibatch|
+--------+--------------------+
         |
         v
+-----------------------------+
| Reflect mutate single key   |
| (prompt or first key)       |
+--------+--------------------+
         |
         v
+-----------------------------+
| Evaluate child              |
+--------+--------------------+
         |
         v
+-----------------------------+
| Accept if improved          |
+-----------------------------+
```

### 9.2 Target merged + multi-param loop (phase 1)

```text
+----------------+
| Select parentA |
+--------+-------+
         |
         +-------------------------------+
         | merge branch? (probability)   |
         +------------------+------------+
                            |yes
                            v
                  +--------------------+
                  | Select parentB     |
                  +---------+----------+
                            |
                            v
                  +------------------------------+
                  | systemAwareMerge per key      |
                  | then LLM/plugin merge for     |
                  | selected component(s)         |
                  +---------+--------------------+
                            |
                            v
                  +------------------------------+
                  | Evaluate child; accept vs     |
                  | max(parentA,parentB)          |
                  +------------------------------+
                            |
                            no
                            v
                  +------------------------------+
                  | Reflect mutate selected key(s)|
                  | (round_robin or all)         |
                  +------------------------------+
```

### 9.3 Python-inspired long-term architecture

```text
        +---------------------+
        | GEPAEngine          |
        +----------+----------+
                   |
      +------------+-------------+
      |                          |
      v                          v
+-------------+         +----------------+
| Reflective  |         | MergeProposer  |
| Proposer    |         | (scheduled)    |
+------+------+         +--------+-------+
       |                         |
       +-----------+-------------+
                   v
          +----------------+
          | Evaluator/Cache|
          +----------------+
                   |
                   v
          +----------------+
          | Frontier/State |
          +----------------+
```

## 10. Pseudocode for Recommended Go Phase-1 Integration

### 10.1 Optimizer loop pseudocode

```pseudo
function Optimize(seedCandidate, examples):
  keys = deriveOptimizableKeys(cfg, seedCandidate)
  pool = [seedNode(keys)]
  evaluate(seedNode, initialBatch)

  while callsUsed < maxEvalCalls:
    parentA = selectParent(pool)
    mode = choose("merge" with mergeProb if pool>=2 else "mutate")

    if mode == "merge":
      parentB = selectParentDistinct(parentA)
      keysToUpdate = selectComponents(parentA)
      child = systemAwareMerge(parentA, parentB)
      for key in keysToUpdate:
        child[key] = proposeMerge(key, parentA, parentB, sideInfoA[key], sideInfoB[key])
      baseline = max(stats(parentA), stats(parentB))
    else:
      keysToUpdate = selectComponents(parentA)
      child = clone(parentA)
      for key in keysToUpdate:
        child[key] = proposeMutation(key, parentA, sideInfo[key])
      baseline = stats(parentA)

    evaluate(child, batch)
    if accept(child, baseline):
      add child to pool
```

### 10.2 JS plugin merge hook fallback pseudocode

```pseudo
if plugin has merge callback:
  mergedText = plugin.merge(input)
else:
  mergedText = reflector.merge(paramA, paramB, sideInfoA, sideInfoB)
```

### 10.3 Component selection pseudocode

```pseudo
if componentSelector == "all":
  return optimizableKeys
else: # round_robin
  key = optimizableKeys[parent.nextIndex % len(optimizableKeys)]
  parent.nextIndex += 1
  return [key]
```

## 11. Risk Analysis and Mitigations

### 11.1 Technical risks

- Merge-quality regressions due weak merge prompt.
- Overfitting when mutating all components every iteration.
- Candidate lineage bugs (Parent2ID, LastUpdated drift).
- Plugin compatibility breaks if merge output parsing is too strict.

### 11.2 Mitigations

- Add deterministic regression tests around merge acceptance and component selector.
- Keep `round_robin` default.
- Encode robust merge output decoding with clear error messages.
- Keep merge callback optional and preserve existing evaluate-only plugins.

## 12. Validation Plan

### 12.1 Unit tests (must-have)

- `pkg/optimizer/gepa/config_test.go`
  - defaults for merge/multi-param fields
- `pkg/optimizer/gepa/optimizer_test.go`
  - multi-key round_robin progression
  - `all` selector updates all keys
  - merge acceptance vs best-parent baseline
  - ancestry metadata correctness
- `cmd/gepa-runner/plugin_loader_test` (or equivalent)
  - merge callback detection and decoding
  - merge output coercion edge cases
- `cmd/gepa-runner/dataset_test`
  - seed candidate JSON/YAML parse

### 12.2 Smoke tests

- optimize with seed prompt (legacy path)
- optimize with seed candidate map + `--optimizable-keys`
- optimize with `--merge-prob > 0` using toy plugin merge callback
- eval-report compatibility unchanged

### 12.3 Acceptance criteria

- All existing tests pass.
- New merge/multi-param tests pass.
- Legacy plugins run unchanged.
- New plugin merge callbacks work with clear decode errors when malformed.

## 13. Implementation Roadmap (Phased)

### Phase A: Parity lift from imported Go

- Port config + optimizer + reflector + format changes.
- Port runner flag and seed-candidate support.
- Port plugin merge hooks.
- Port toy example merge function updates.

### Phase B: Parity hardening in go-go-gepa

- Add integration tests and regressions not present upstream.
- Improve CLI help docs for new multi-param semantics.
- Add artifactized smoke scripts in ticket sources.

### Phase C: Python-inspired architecture lift (optional but recommended)

- Introduce explicit proposer lane abstraction (reflective and merge proposer objects).
- Add callback/event hooks for merge attempted/accepted/rejected.
- Add optional seedless mode and richer side-info schemas.

## 14. How to Leverage `optimize_anything` Ideas in JS Plugins

### 14.1 Short-term plugin authoring guidance

For plugin authors today, adopt these practices:

- represent candidate as multi-key map from day one, even if starting with one key
- emit objective scores map (not just scalar score)
- include key-specific feedback blocks for future multi-param merges

### 14.2 Medium-term contract enhancement proposal

Add optional plugin methods:

- `selectComponents(input, options)` for custom per-iteration component policy
- `buildSideInfo(input, options)` for key-specific reflective datasets

These should remain optional; core optimizer should provide defaults.

### 14.3 Example optimize_anything-like plugin skeleton (proposed)

```javascript
module.exports = plugins.defineOptimizerPlugin({
  apiVersion: plugins.OPTIMIZER_PLUGIN_API_VERSION,
  kind: "optimizer",
  id: "my.multicomponent.task",
  name: "My Multi-Component Task",
  create(ctx) {
    return {
      dataset() { return loadExamples(); },

      evaluate({ candidate, example }, options) {
        const out = runSystem(candidate, example, options);
        return {
          score: out.scalar,
          objectives: {
            accuracy: out.accuracy,
            cost: -out.tokens,
            latency: -out.ms,
          },
          feedback: out.feedback,
          trace: out.trace,
          sideInfoByKey: out.sideInfoByKey,
        };
      },

      merge({ candidateA, candidateB, paramKey, sideInfoA, sideInfoB }, options) {
        // custom domain-aware merge per key
        return mergeComponent(candidateA[paramKey], candidateB[paramKey], sideInfoA, sideInfoB);
      },
    };
  },
});
```

This is effectively a JS-friendly analogue of Python adapter + proposer interplay.

## 15. Recommendation and Next Decision

### 15.1 Recommended immediate action

Proceed with **Phase A + Phase B** in `go-go-gepa` now:

- import upstream merge/multi-param changes
- keep existing recorder/reporter features
- test and stabilize

### 15.2 Recommended sequencing

1. Land core port + tests.
2. Validate on toy scripts with and without merge callback.
3. Then decide whether to invest in Phase C proposer/callback architecture.

### 15.3 Why this sequence is pragmatic

- It captures the high-value functionality gap quickly.
- It minimizes redesign risk during active migration.
- It creates a reliable baseline for Python-aligned enhancements instead of mixing parity and redesign in one jump.

## 16. Appendix: Key Source Pointers

### 16.1 Current Go baseline

- `go-go-gepa/pkg/optimizer/gepa/config.go:6`
- `go-go-gepa/pkg/optimizer/gepa/optimizer.go:19`
- `go-go-gepa/cmd/gepa-runner/main.go:43`
- `go-go-gepa/cmd/gepa-runner/plugin_loader.go:29`

### 16.2 Upstream Go merge+multi-param

- `imported/geppetto-main/pkg/optimizer/gepa/config.go:28`
- `imported/geppetto-main/pkg/optimizer/gepa/optimizer.go:18`
- `imported/geppetto-main/cmd/gepa-runner/main.go:44`
- `imported/geppetto-main/cmd/gepa-runner/plugin_loader.go:37`

### 16.3 Python reference

- `gepa/src/gepa/optimize_anything.py:694`
- `gepa/src/gepa/optimize_anything.py:998`
- `gepa/src/gepa/optimize_anything.py:1429`
- `gepa/src/gepa/proposer/merge.py:210`
- `gepa/src/gepa/core/engine.py:406`
- `gepa/src/gepa/strategies/component_selector.py:10`
- `gepa/src/gepa/proposer/reflective_mutation/reflective_mutation.py:89`
- `gepa/src/gepa/adapters/optimize_anything_adapter/optimize_anything_adapter.py:508`

### 16.4 JS plugin descriptor helper

- `geppetto/pkg/js/modules/geppetto/plugins_module.go:88`

