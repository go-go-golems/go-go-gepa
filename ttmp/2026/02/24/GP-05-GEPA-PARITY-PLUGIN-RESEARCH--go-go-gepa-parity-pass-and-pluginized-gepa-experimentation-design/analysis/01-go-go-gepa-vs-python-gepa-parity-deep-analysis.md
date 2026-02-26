---
Title: go-go-gepa vs python gepa parity deep analysis
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
    - Path: ../../../../../../../go-go-gepa/cmd/gepa-runner/main.go
      Note: Runner wiring of plugin hooks into optimizer hooks
    - Path: ../../../../../../../go-go-gepa/pkg/optimizer/gepa/config.go
      Note: Go config knobs for merge scheduler, optimizable keys, and component selector
    - Path: ../../../../../../../go-go-gepa/pkg/optimizer/gepa/optimizer.go
      Note: |-
        Go optimizer loop, parent selection, component selection, and minibatch sampling behavior
        Core optimizer loop and parity behavior
    - Path: ../../../../../../../go-go-gepa/pkg/optimizer/gepa/pareto.go
      Note: |-
        Candidate-level non-dominance computation in Go
        Go Pareto front computation
    - Path: src/gepa/core/engine.py
      Note: |-
        Python optimization control flow, acceptance, and merge scheduling behavior
        Python optimization control flow
    - Path: src/gepa/core/state.py
      Note: |-
        Python GEPA state initialization and frontier tracking by frontier type
        Python frontier state semantics
    - Path: src/gepa/gepa_utils.py
      Note: Python Pareto dominator pruning and frequency-based sampling
    - Path: src/gepa/strategies/batch_sampler.py
      Note: |-
        Python epoch-shuffled minibatch behavior
        Python minibatch strategy
    - Path: src/gepa/strategies/candidate_selector.py
      Note: Python candidate selection strategies and Pareto selector
    - Path: src/gepa/strategies/component_selector.py
      Note: Python module selection strategies
ExternalSources: []
Summary: Detailed parity analysis between go-go-gepa and Python GEPA with implementation blueprint for initial frontier seeding, frontier semantics, component selection, and minibatch policies.
LastUpdated: 2026-02-24T10:44:00-05:00
WhatFor: Guide the next implementation pass that raises behavioral parity with Python GEPA while preserving Go ergonomics.
WhenToUse: Use when planning or reviewing parity work on parent selection, frontier tracking, component selection, and batch sampling.
---


# go-go-gepa vs Python GEPA parity deep analysis

## 1. Executive summary

This analysis compares `go-go-gepa` and Python `gepa` for four core optimization mechanics:

1. initial Pareto set computation,
2. Pareto frontier computation,
3. component/module selection,
4. minibatch computation.

Observed result: `go-go-gepa` is intentionally compact and candidate-centric, while Python GEPA is stateful and frontier-centric across validation keys. The Go implementation is already extensible (merge hook, component selector hook, side-info hook, event hook), but its current semantics differ from Python in ways that materially affect exploration behavior, reproducibility, and how frontier diversity is exploited.

The primary parity work should therefore focus on state model and sampling policy, not only on adding more hooks.

## 2. Problem statement and scope

The current task is a parity pass over `go-go-gepa` against Python GEPA, with emphasis on how each system:

1. seeds and updates Pareto state,
2. chooses parents from that state,
3. chooses which prompt component to edit,
4. samples minibatches for mutate/merge acceptance.

Scope boundaries:

1. In scope: optimizer semantics, strategy wiring, data flow from runner/plugin into optimizer, and a phased parity plan.
2. Out of scope: broad API redesign of all runner commands, non-parity UX work, and replacing JS plugin architecture.

## 3. Current-state architecture (evidence-based)

### 3.1 go-go-gepa core loop

Go optimizer control flow is concentrated in `Optimize(...)`.

1. Derive optimizable keys from config or seed (`deriveOptimizableKeys`) and initialize single seed node.
2. Evaluate seed on one initial sampled minibatch.
3. Iterate until evaluation budget is consumed.
4. Select parent from a frontier-like subset.
5. Optionally select second parent for merge.
6. Sample minibatch constrained by remaining budget and operation multiplier.
7. Evaluate parent(s), propose child (mutate or merge), evaluate child on same minibatch.
8. Accept child if dominates baseline objective vector or improves scalar score beyond epsilon.

Evidence:

1. `go-go-gepa/pkg/optimizer/gepa/optimizer.go:242`
2. `go-go-gepa/pkg/optimizer/gepa/optimizer.go:285`
3. `go-go-gepa/pkg/optimizer/gepa/optimizer.go:304`
4. `go-go-gepa/pkg/optimizer/gepa/optimizer.go:339`
5. `go-go-gepa/pkg/optimizer/gepa/optimizer.go:451`
6. `go-go-gepa/pkg/optimizer/gepa/optimizer.go:516`

### 3.2 Python GEPA core loop

Python GEPA splits behavior into state, engine, proposer, and strategy objects.

1. Engine initializes state via full valset evaluation of seed candidate.
2. Reflective proposer selects candidate and minibatch via pluggable strategies.
3. Merge proposer is separately scheduled and consumes merge budget/state.
4. Accepted proposals trigger full valset evaluation and state frontier updates.

Evidence:

1. `gepa/src/gepa/core/engine.py:287`
2. `gepa/src/gepa/core/engine.py:406`
3. `gepa/src/gepa/core/engine.py:484`
4. `gepa/src/gepa/core/state.py:483`
5. `gepa/src/gepa/proposer/reflective_mutation/reflective_mutation.py:138`
6. `gepa/src/gepa/proposer/merge.py:290`

### 3.3 Runner/plugin integration in Go

`gepa-runner` wires JS plugin hooks into Go optimizer hooks.

1. `evaluate` is required and called per `(candidate, example)`.
2. Optional hooks: `merge`, `initialCandidate`, `selectComponents`, `componentSideInfo`.
3. CLI flags map directly to config knobs for batch size, merge scheduler, component selector, and keys.

Evidence:

1. `go-go-gepa/cmd/gepa-runner/plugin_loader.go:37`
2. `go-go-gepa/cmd/gepa-runner/plugin_loader.go:108`
3. `go-go-gepa/cmd/gepa-runner/plugin_loader.go:291`
4. `go-go-gepa/cmd/gepa-runner/plugin_loader.go:319`
5. `go-go-gepa/cmd/gepa-runner/main.go:246`
6. `go-go-gepa/cmd/gepa-runner/main.go:275`
7. `go-go-gepa/cmd/gepa-runner/main.go:284`

## 4. Detailed parity analysis for the four target behaviors

### 4.1 Initial Pareto set computation

#### go-go-gepa behavior

Go initializes with only seed candidate in pool and evaluates seed on an initial sampled minibatch, not full validation set.

1. Seed node inserted into pool with `ID=0`.
2. `sampleBatchIndices(...)` picks initial subset.
3. Seed evaluation cache/statistics are therefore partial at startup.

Evidence:

1. `go-go-gepa/pkg/optimizer/gepa/optimizer.go:267`
2. `go-go-gepa/pkg/optimizer/gepa/optimizer.go:285`

#### Python behavior

Python initializes full frontier structures from full valset seed evaluation.

1. `initialize_gepa_state(...)` executes seed against full valset evaluator.
2. State stores per-val-id best score/front (`pareto_front_valset`, `program_at_pareto_front_valset`).
3. Also stores objective front(s) when objective scores are present.

Evidence:

1. `gepa/src/gepa/core/state.py:204`
2. `gepa/src/gepa/core/state.py:205`
3. `gepa/src/gepa/core/state.py:206`
4. `gepa/src/gepa/core/state.py:650`

#### Parity implication

Go parent selection starts from sparse, minibatch-biased evidence, while Python starts from full valset-aligned state. This changes early-iteration dynamics and candidate selection pressure.

### 4.2 Pareto frontier computation

#### go-go-gepa behavior

Go computes a candidate-level non-dominated set on demand across pool-level mean objective vectors.

1. Build objective vector per candidate from cached global stats.
2. If multi-objective keys > 1, run `ParetoFront(...)` over candidate vectors.
3. If single objective, fallback to top-k by scalar mean score.
4. Select parent by weighted random among resulting candidate indices.

Evidence:

1. `go-go-gepa/pkg/optimizer/gepa/optimizer.go:647`
2. `go-go-gepa/pkg/optimizer/gepa/optimizer.go:660`
3. `go-go-gepa/pkg/optimizer/gepa/pareto.go:45`

#### Python behavior

Python tracks frontier incrementally per configured frontier type:

1. `instance`: per validation example,
2. `objective`: per objective metric,
3. `hybrid`: combined mapping,
4. `cartesian`: per `(val_id, objective)`.

Candidate selection then prunes dominated programs across frontier memberships and samples proportionally to frontier frequency.

Evidence:

1. `gepa/src/gepa/core/state.py:540`
2. `gepa/src/gepa/core/state.py:442`
3. `gepa/src/gepa/core/state.py:430`
4. `gepa/src/gepa/gepa_utils.py:37`
5. `gepa/src/gepa/gepa_utils.py:90`
6. `gepa/src/gepa/strategies/candidate_selector.py:18`

#### Parity implication

Go’s frontier semantics are global candidate-vector non-dominance; Python’s are keyed-frontier dominance and frequency-of-coverage. They are not equivalent and can produce different parent distributions even under identical raw scores.

### 4.3 Which module/component to optimize

#### go-go-gepa behavior

Go component selection pipeline:

1. `paramKeys` are derived from configured keys or seed keys.
2. Optional hook (`componentSelectorFn`) can override selection each iteration.
3. Fallback behavior is `round_robin` or `all` via `ComponentSelector` config.
4. Round-robin pointer is maintained in candidate node (`NextParamIndex`).

Evidence:

1. `go-go-gepa/pkg/optimizer/gepa/optimizer.go:1023`
2. `go-go-gepa/pkg/optimizer/gepa/optimizer.go:733`
3. `go-go-gepa/pkg/optimizer/gepa/optimizer.go:755`
4. `go-go-gepa/pkg/optimizer/gepa/optimizer.go:1084`

#### Python behavior

Python module selection is strategy-based and receives richer context:

1. Selector called with state, trajectories, minibatch scores, candidate index, and candidate.
2. Default strategies are `round_robin` and `all`.
3. Round-robin index for each candidate is kept in state-level array.

Evidence:

1. `gepa/src/gepa/proposer/reflective_mutation/reflective_mutation.py:261`
2. `gepa/src/gepa/strategies/component_selector.py:10`
3. `gepa/src/gepa/core/state.py:170`

#### Parity implication

Go has comparable strategy outcomes for simple cases, but Python selectors can be trajectory-informed by default interface shape. Go’s hook can emulate this only if traces/score context are provided to it.

### 4.4 Minibatch computation

#### go-go-gepa behavior

Go samples without replacement via random permutation each iteration, bounded by budget and operation multiplier.

1. Batch size reduced when remaining eval budget cannot afford full mutation/merge triplet plan.
2. `sampleBatchIndices` is stateless except RNG.

Evidence:

1. `go-go-gepa/pkg/optimizer/gepa/optimizer.go:320`
2. `go-go-gepa/pkg/optimizer/gepa/optimizer.go:326`
3. `go-go-gepa/pkg/optimizer/gepa/optimizer.go:855`

#### Python behavior

Python reflective path uses `EpochShuffledBatchSampler`:

1. Shuffle once per epoch,
2. pad to minibatch multiple using least-frequent IDs,
3. deterministic slicing by iteration index.

Merge path uses dedicated val-overlap stratified sampling.

Evidence:

1. `gepa/src/gepa/strategies/batch_sampler.py:17`
2. `gepa/src/gepa/strategies/batch_sampler.py:50`
3. `gepa/src/gepa/strategies/batch_sampler.py:58`
4. `gepa/src/gepa/proposer/merge.py:258`

#### Parity implication

Go’s sampling is simpler but less reproducible in epoch structure and does not distinguish reflective vs merge sampling policy beyond shared random minibatch.

## 5. Key behavior gaps and why they matter

### Gap A: Seed initialization bias

1. Go’s initial state is minibatch-only.
2. Python’s initial state is full-valset grounded.

Risk: parent-selection pressure may overfit early sampled subset in Go.

### Gap B: Frontier semantics mismatch

1. Go uses global candidate-level vector non-dominance.
2. Python uses keyed frontier mappings and coverage-weighted selector.

Risk: different exploration diversity and parent reuse patterns.

### Gap C: Strategy-interface richness mismatch

1. Go component selector hook has narrow context.
2. Python module selector can consume trajectories and scores.

Risk: reduced ability to run trajectory-aware research variants in Go.

### Gap D: Sampling policy mismatch

1. Go: per-iteration random permutation.
2. Python: epoch-shuffled/padded reflective batches + dedicated merge subsampling.

Risk: weaker reproducibility and different acceptance statistics under same budget.

## 6. Proposed parity architecture for go-go-gepa

### 6.1 State model additions

Add explicit optimization state for:

1. valset evaluations by candidate,
2. frontier mappings by type (`instance`, `objective`, `hybrid`, `cartesian`),
3. per-candidate selector cursor (module pointer),
4. sampled-batch bookkeeping for deterministic replay.

### 6.2 Evaluation flow split

Separate:

1. reflective minibatch eval path,
2. full valset eval path for accepted candidates,
3. merge-specific subsample path.

### 6.3 Frontier and parent selection pipeline

Replace on-the-fly global front with:

1. incremental frontier updates on accepted/full-eval candidates,
2. dominator pruning across frontier memberships,
3. frequency-weighted parent sampling.

### 6.4 Component selection context upgrade

Extend component selector input to include:

1. minibatch example IDs,
2. current candidate minibatch scores,
3. optional trace summaries.

This allows parity with Python selector interface without forcing Python internals into Go.

## 7. Pseudocode sketches

### 7.1 Seed and frontier initialization

```text
seed_eval = full_val_evaluate(seed_candidate)
state.add_candidate(seed_candidate, parents=[])
state.update_frontiers(candidate=0, val_eval=seed_eval, frontier_type)
```

### 7.2 Parent selection (Python-like)

```text
front_mapping = state.get_frontier_mapping(frontier_type)
survivors = remove_dominated_programs(front_mapping, per_program_scores)
weights = frequency_on_front(survivors, front_mapping)
parent = weighted_sample(weights, rng)
```

### 7.3 Iteration evaluation split

```text
parent = candidate_selector(state)
batch_ids = reflective_batch_sampler.next_minibatch_ids(train_loader, state)
proposal = propose_mutation_or_merge(parent, batch_ids, state)
if proposal.accepted_on_subsample:
    val_eval = full_val_evaluate(proposal.candidate)
    state.add_candidate(proposal.candidate, parents)
    state.update_frontiers(new_id, val_eval)
```

## 8. Phased implementation plan

### Phase 1: Frontier-state scaffolding

1. Introduce frontier-type enum/config in Go optimizer state.
2. Add state structures for keyed frontier mappings.
3. Add full valset evaluation function path for seed and accepted children.

### Phase 2: Parent selection parity

1. Implement dominator-pruning utility and frequency-weighted sampling.
2. Gate behind config strategy for safe rollout (`candidate_selector=pareto_coverage` etc.).

### Phase 3: Batch sampler abstraction

1. Add `BatchSampler` interface in Go.
2. Implement `EpochShuffledBatchSampler` equivalent.
3. Add merge-subsample selector with overlap floor.

### Phase 4: Component-selector context parity

1. Extend component selector input payload and JS bridge schema.
2. Keep old fields for compatibility.

### Phase 5: Validation hardening

1. Golden tests for frontier updates.
2. Deterministic batch-sequence tests.
3. Parent sampling distribution smoke tests.

## 9. Testing and validation strategy

### Unit tests

1. Seed initialization populates full frontier maps.
2. `instance/objective/hybrid/cartesian` mappings update correctly on ties/improvements.
3. Dominator pruning and weighted sampling produce expected candidate set.
4. Epoch-shuffled sampler produces deterministic cycle and correct padding.

### Integration tests

1. Runner optimize with fixed seed reproduces same sequence of selected parents and minibatches.
2. Merge path uses overlap floor and refuses insufficient overlap.
3. JS selector hook receives new extended context fields.

### Regression checks

1. Existing script examples still load and pass smoke tests.
2. Current single-objective workflows remain stable under compatibility mode.

## 10. Risks, alternatives, and open questions

### Risks

1. Full-valset evaluations may increase runtime if not controlled by policy.
2. Overfitting parity to one Python configuration might reduce Go simplicity.
3. Frontier-type expansion increases state and serialization complexity.

### Alternatives

1. Keep current Go core and add optional Python-like mode only.
2. Emulate only parent selector behavior without full frontier state parity.
3. Implement parity entirely in plugin/userland. This is weaker because several decisions are currently internal to optimizer.

### Open questions

1. Should Go default remain compact mode, with parity mode opt-in?
2. Which frontier type should be default for multi-objective plugins (`instance` or `hybrid`)?
3. Should full valset evaluation be mandatory on every accepted child or policy-driven?

## 11. Recommended near-term decisions

1. Adopt a dual-mode plan: `compact` (current semantics) and `parity` (Python-like).
2. Implement frontier-state + batch-sampler abstractions before adding more plugin hooks.
3. Preserve the existing JS plugin contract while adding additive fields/hooks.

## 12. References

Primary evidence bundle for this ticket:

1. `ttmp/2026/02/24/GP-05-GEPA-PARITY-PLUGIN-RESEARCH--go-go-gepa-parity-pass-and-pluginized-gepa-experimentation-design/sources/01-go-optimizer-hooks-and-types.txt`
2. `ttmp/2026/02/24/GP-05-GEPA-PARITY-PLUGIN-RESEARCH--go-go-gepa-parity-pass-and-pluginized-gepa-experimentation-design/sources/02-go-optimizer-main-loop.txt`
3. `ttmp/2026/02/24/GP-05-GEPA-PARITY-PLUGIN-RESEARCH--go-go-gepa-parity-pass-and-pluginized-gepa-experimentation-design/sources/03-go-optimizer-selection-batching.txt`
4. `ttmp/2026/02/24/GP-05-GEPA-PARITY-PLUGIN-RESEARCH--go-go-gepa-parity-pass-and-pluginized-gepa-experimentation-design/sources/11-py-state.txt`
5. `ttmp/2026/02/24/GP-05-GEPA-PARITY-PLUGIN-RESEARCH--go-go-gepa-parity-pass-and-pluginized-gepa-experimentation-design/sources/12-py-engine.txt`
6. `ttmp/2026/02/24/GP-05-GEPA-PARITY-PLUGIN-RESEARCH--go-go-gepa-parity-pass-and-pluginized-gepa-experimentation-design/sources/14-py-candidate-selector.txt`
7. `ttmp/2026/02/24/GP-05-GEPA-PARITY-PLUGIN-RESEARCH--go-go-gepa-parity-pass-and-pluginized-gepa-experimentation-design/sources/15-py-batch-sampler.txt`
