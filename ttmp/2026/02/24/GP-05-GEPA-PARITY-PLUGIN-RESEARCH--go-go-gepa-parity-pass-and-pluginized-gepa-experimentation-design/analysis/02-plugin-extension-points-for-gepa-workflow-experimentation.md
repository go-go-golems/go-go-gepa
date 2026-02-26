---
Title: plugin extension points for GEPA workflow experimentation
Ticket: GP-05-GEPA-PARITY-PLUGIN-RESEARCH
Status: active
Topics:
    - architecture
    - tools
    - inference
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../go-go-gepa/cmd/gepa-runner/README.md
      Note: Public plugin contract documentation
    - Path: ../../../../../../../go-go-gepa/cmd/gepa-runner/js_runtime.go
      Note: JS runtime module registration and require surface
    - Path: ../../../../../../../go-go-gepa/cmd/gepa-runner/main.go
      Note: |-
        Runner-to-optimizer hook wiring points
        Hook wiring from JS to optimizer
    - Path: ../../../../../../../go-go-gepa/cmd/gepa-runner/plugin_loader.go
      Note: |-
        Current JS plugin contract and decoded optional hooks
        Current JS plugin contract and decoder behavior
    - Path: ../../../../../../../go-go-gepa/cmd/gepa-runner/scripts/lib/gepa_plugin_contract.js
      Note: Local plugin descriptor contract helper
    - Path: ../../../../../../../go-go-gepa/pkg/optimizer/gepa/optimizer.go
      Note: |-
        Internal optimization phases and hook insertion points
        Potential hook insertion points across phases
    - Path: src/gepa/core/adapter.py
      Note: |-
        Python adapter contract for evaluation/reflection/proposal
        Python adapter extensibility model
    - Path: src/gepa/proposer/reflective_mutation/reflective_mutation.py
      Note: Python proposer interfaces and context-rich strategy hooks
    - Path: src/gepa/strategies/batch_sampler.py
      Note: Python batch sampler extension model
    - Path: src/gepa/strategies/candidate_selector.py
      Note: |-
        Python candidate selection strategy model
        Python strategy object model
ExternalSources: []
Summary: Extension-architecture proposal for JS plugins across GEPA workflow phases, enabling rapid experimentation with sampling, selection, frontier, acceptance, and merge strategies.
LastUpdated: 2026-02-24T10:45:00-05:00
WhatFor: Define a pluginized experimentation surface so GEPA variants can be tested in JS without frequent Go-core edits.
WhenToUse: Use when designing or implementing new optimizer hooks and JS plugin APIs for research iteration speed.
---


# Plugin extension points for GEPA workflow experimentation

## 1. Executive summary

`go-go-gepa` already supports a useful JS plugin surface, but most research-critical policy decisions still live in Go core (parent selection semantics, batch policy, frontier handling, acceptance policy, merge scheduling details). This document maps current hookability and proposes a staged plugin extension architecture so experimentation can happen primarily in JS while preserving a stable, typed Go execution core.

The design recommendation is a layered contract:

1. keep the current evaluator-oriented contract as `gepa.optimizer/v1`,
2. introduce additive phase hooks in `gepa.optimizer/v2`,
3. keep safe defaults in Go when hooks are omitted.

## 2. Current plugin surface (what is already pluggable)

### 2.1 Loader contract today

The plugin loader currently recognizes:

1. required: `evaluate(input, options)`,
2. optional: `dataset`, `merge`, `initialCandidate`, `selectComponents`, `componentSideInfo`.

Evidence:

1. `go-go-gepa/cmd/gepa-runner/plugin_loader.go:37`
2. `go-go-gepa/cmd/gepa-runner/plugin_loader.go:108`
3. `go-go-gepa/cmd/gepa-runner/plugin_loader.go:291`
4. `go-go-gepa/cmd/gepa-runner/plugin_loader.go:319`
5. `go-go-gepa/cmd/gepa-runner/plugin_loader.go:357`

### 2.2 Runner wiring today

Runner maps plugin hooks to optimizer hooks:

1. `SetMergeFunc(...)`,
2. `SetComponentSelectorFunc(...)`,
3. `SetSideInfoFunc(...)`,
4. optional event streaming to stdout.

Evidence:

1. `go-go-gepa/cmd/gepa-runner/main.go:275`
2. `go-go-gepa/cmd/gepa-runner/main.go:276`
3. `go-go-gepa/cmd/gepa-runner/main.go:284`
4. `go-go-gepa/cmd/gepa-runner/main.go:292`
5. `go-go-gepa/cmd/gepa-runner/main.go:300`

### 2.3 What remains hard-coded in Go

Despite current hooks, these are still internal policies:

1. parent selection from frontier (`selectParent`),
2. frontier computation mode,
3. minibatch sampler (`sampleBatchIndices`),
4. acceptance policy (`acceptChild`),
5. merge scheduling state machine (`shouldAttemptMerge`, `updateMergeDue`),
6. mutation proposal structure beyond component-level text replacement.

Evidence:

1. `go-go-gepa/pkg/optimizer/gepa/optimizer.go:642`
2. `go-go-gepa/pkg/optimizer/gepa/optimizer.go:855`
3. `go-go-gepa/pkg/optimizer/gepa/optimizer.go:516`
4. `go-go-gepa/pkg/optimizer/gepa/optimizer.go:528`

## 3. Workflow-phase map and hook opportunities

Define optimizer workflow as explicit phases:

1. seed resolution,
2. candidate parent selection,
3. operation scheduling (mutate vs merge),
4. batch selection,
5. proposal generation,
6. evaluation and acceptance,
7. frontier update,
8. observability/reporting.

Current JS coverage:

1. phase 1 partial (`initialCandidate`),
2. phase 5 partial (`merge`, `selectComponents`, `componentSideInfo`),
3. phase 4/6 only indirectly via `evaluate`.

Gaps:

1. no JS control for parent/frontier strategy,
2. no JS control for batch sampler,
3. no JS control for acceptance comparator,
4. no JS control for merge scheduling policy,
5. no JS control for frontier update policy.

## 4. Proposed v2 plugin architecture

## 4.1 Design principles

1. Additive: keep `v1` behavior unchanged when new hooks are absent.
2. Safe fallback: Go defaults remain canonical fallback.
3. Typed boundaries: each hook gets explicit input/output schema.
4. Deterministic options: pass RNG seed/context into hooks where needed.
5. Budget safety: Go core remains final authority for `MaxEvalCalls` and guardrails.

## 4.2 Proposed hook families

### A. Dataset and batching

1. `dataset(options) -> array` (already exists)
2. `sampleBatch(input, options) -> { indices: number[] }`

`sampleBatch` input proposal:

```json
{
  "phase": "mutate|merge|seed_init",
  "datasetSize": 1234,
  "batchSizeRequested": 8,
  "budgetRemaining": 40,
  "iteration": 12,
  "operationMultiplier": 2,
  "history": {
    "recentBatchIndices": [1,2,3]
  }
}
```

### B. Parent/frontier selection

1. `selectParents(input, options) -> { parentId: number, parent2Id?: number }`
2. `computeFrontier(input, options) -> { candidateIds: number[] }` (optional override)

Input should include:

1. candidate stats and objectives,
2. current frontier snapshots,
3. merge eligibility context.

### C. Operation scheduling

1. `scheduleOperation(input, options) -> { operation: "mutate"|"merge" }`
2. `onIterationOutcome(input, options)` callback for adaptive schedulers.

This makes `merges_due` experiments easy to run without recompiling Go.

### D. Proposal generation

Current:

1. `merge`, `selectComponents`, `componentSideInfo`.

Add:

1. `mutate(input, options) -> { candidatePatch, raw }`

This allows plugin-side custom mutation logic for multi-field coupled edits, similar to Python custom proposer patterns.

### E. Acceptance policy

1. `acceptCandidate(input, options) -> { accept: boolean, reason?: string }`

Input includes:

1. parent and child minibatch stats,
2. objective vectors,
3. baseline policy decision as `defaultAccept`.

Go should still enforce hard constraints (e.g., budget integrity, result validity).

### F. Frontier update policy

1. `updateFrontier(input, options) -> { frontierDelta }` (advanced; optional)

Default remains Go internal. Hook is for research-only frontier semantics.

### G. Observability and telemetry

Current Go event stream is callback-only on Go side. Add plugin callback hooks:

1. `onEvent(event, options)` for online adaptation and richer trace capture.

## 5. API sketch for `gepa.optimizer/v2`

```js
const plugins = require("./lib/gepa_plugin_contract_v2");

module.exports = plugins.defineOptimizerPlugin({
  apiVersion: "gepa.optimizer/v2",
  kind: "optimizer",
  id: "research.my_variant",
  name: "My GEPA variant",
  create(ctx) {
    return {
      dataset() {},
      evaluate(input, options) {},

      initialCandidate(options) {},
      sampleBatch(input, options) {},
      selectParents(input, options) {},
      scheduleOperation(input, options) {},
      selectComponents(input, options) {},
      componentSideInfo(input, options) {},
      mutate(input, options) {},
      merge(input, options) {},
      acceptCandidate(input, options) {},
      onEvent(event, options) {},
    };
  }
});
```

## 6. Go bridge changes required

### 6.1 Optimizer hook interfaces

Add new hook types in `pkg/optimizer/gepa/optimizer.go`:

1. `BatchSamplerFunc`,
2. `ParentSelectorFunc`,
3. `OperationSchedulerFunc`,
4. `MutateFunc`,
5. `AcceptFunc`.

Use existing hook style (`SetXFunc(...)`) like:

1. `SetMergeFunc` (`go-go-gepa/pkg/optimizer/gepa/optimizer.go:201`)
2. `SetComponentSelectorFunc` (`go-go-gepa/pkg/optimizer/gepa/optimizer.go:218`)

### 6.2 Plugin loader decoding

Extend loader with optional callables and decoders, mirroring current method patterns:

1. detect method via `findOptionalCallable(...)`,
2. marshal structured input maps,
3. decode typed return payloads with fallback error messages.

Evidence pattern exists in:

1. `go-go-gepa/cmd/gepa-runner/plugin_loader.go:129`
2. `go-go-gepa/cmd/gepa-runner/plugin_loader.go:319`
3. `go-go-gepa/cmd/gepa-runner/plugin_loader.go:499`

### 6.3 Runner wiring

Map plugin capabilities to optimizer hooks in `main.go` using existing pattern used for merge/select/side-info.

## 7. Comparison to Python extension model

Python GEPA already externalizes many decisions through strategy objects and adapter methods:

1. `CandidateSelector`,
2. `BatchSampler`,
3. `ReflectionComponentSelector`,
4. `GEPAAdapter.make_reflective_dataset`,
5. optional custom proposer.

Evidence:

1. `gepa/src/gepa/strategies/candidate_selector.py:11`
2. `gepa/src/gepa/strategies/batch_sampler.py:13`
3. `gepa/src/gepa/strategies/component_selector.py:10`
4. `gepa/src/gepa/core/adapter.py:58`
5. `gepa/src/gepa/proposer/reflective_mutation/reflective_mutation.py:89`

The proposed Go `v2` plugin architecture is the JS analog of this strategy decomposition.

## 8. Pseudocode for hook dispatch with fallback

```text
if plugin.hasSampleBatch():
    batchIdx = plugin.sampleBatch(ctx)
else:
    batchIdx = goDefaultSampleBatch()

if plugin.hasSelectParents():
    parent, parent2 = plugin.selectParents(state)
else:
    parent = goDefaultSelectParent()

if plugin.hasAcceptCandidate():
    accepted = plugin.acceptCandidate(compareInput)
else:
    accepted = goDefaultAccept(baseline, child)
```

## 9. Phased rollout plan

### Phase 1: Low-risk additions

1. Add `sampleBatch`, `scheduleOperation`, `acceptCandidate` hooks.
2. Preserve exact existing behavior when absent.
3. Add schema validation and explicit error reporting.

### Phase 2: Parent/frontier experiments

1. Add `selectParents` hook.
2. Expose read-only frontier snapshot payload.

### Phase 3: Advanced proposal hooks

1. Add `mutate` hook for full candidate patch proposals.
2. Add optional `onEvent` callback.

### Phase 4: Frontier override (optional)

1. Add frontier-update override hook only after baseline parity mode is stable.

## 10. Test strategy

1. Contract tests: each optional hook accepted/rejected with clear errors.
2. Fallback tests: no-hook plugin must keep current behavior.
3. Hook precedence tests: plugin hook must override Go default for that phase only.
4. Budget safety tests: plugin cannot violate eval budget constraints.
5. Replay tests: deterministic plugin + seed must replay same schedule.

## 11. Risks and mitigations

### Risk 1: Contract explosion

Mitigation:

1. split into `v1` stable and `v2` experimental,
2. document each hook as optional and independent.

### Risk 2: Non-deterministic research plugins

Mitigation:

1. pass `seed` and deterministic context,
2. require hook outputs to be pure-data, no side effects in core loop decisions.

### Risk 3: Debug complexity

Mitigation:

1. standardize event envelopes,
2. persist hook inputs/outputs in recorder under optional debug mode.

## 12. Recommended immediate next step

Implement Phase 1 hooks (`sampleBatch`, `scheduleOperation`, `acceptCandidate`) before parity-state refactor lands. This gives researchers immediate leverage while major optimizer-state changes are being built.

## 13. References

Primary evidence bundle for this ticket:

1. `ttmp/2026/02/24/GP-05-GEPA-PARITY-PLUGIN-RESEARCH--go-go-gepa-parity-pass-and-pluginized-gepa-experimentation-design/sources/07-go-plugin-loader-contract.txt`
2. `ttmp/2026/02/24/GP-05-GEPA-PARITY-PLUGIN-RESEARCH--go-go-gepa-parity-pass-and-pluginized-gepa-experimentation-design/sources/06-go-runner-wiring.txt`
3. `ttmp/2026/02/24/GP-05-GEPA-PARITY-PLUGIN-RESEARCH--go-go-gepa-parity-pass-and-pluginized-gepa-experimentation-design/sources/01-go-optimizer-hooks-and-types.txt`
4. `ttmp/2026/02/24/GP-05-GEPA-PARITY-PLUGIN-RESEARCH--go-go-gepa-parity-pass-and-pluginized-gepa-experimentation-design/sources/09-go-plugin-docs.txt`
5. `ttmp/2026/02/24/GP-05-GEPA-PARITY-PLUGIN-RESEARCH--go-go-gepa-parity-pass-and-pluginized-gepa-experimentation-design/sources/17-py-reflective-proposer.txt`
6. `ttmp/2026/02/24/GP-05-GEPA-PARITY-PLUGIN-RESEARCH--go-go-gepa-parity-pass-and-pluginized-gepa-experimentation-design/sources/14-py-candidate-selector.txt`
7. `ttmp/2026/02/24/GP-05-GEPA-PARITY-PLUGIN-RESEARCH--go-go-gepa-parity-pass-and-pluginized-gepa-experimentation-design/sources/15-py-batch-sampler.txt`
8. `ttmp/2026/02/24/GP-05-GEPA-PARITY-PLUGIN-RESEARCH--go-go-gepa-parity-pass-and-pluginized-gepa-experimentation-design/sources/20-py-adapter-contract.txt`
