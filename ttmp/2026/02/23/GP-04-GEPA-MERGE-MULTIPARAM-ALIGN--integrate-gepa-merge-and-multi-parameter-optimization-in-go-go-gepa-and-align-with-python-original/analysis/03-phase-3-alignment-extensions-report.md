---
Title: Phase 3 Alignment Extensions Report
Ticket: GP-04-GEPA-MERGE-MULTIPARAM-ALIGN
Status: active
Topics:
    - architecture
    - migration
    - testing
    - geppetto
    - inference
DocType: analysis
Intent: implementation-summary
Owners: []
RelatedFiles:
    - Path: ../../../../../../../go-go-gepa/cmd/gepa-runner/main.go
      Note: |-
        Runner-level wiring for scheduler and extension hooks
        Phase 3 runner flags and hook wiring
    - Path: ../../../../../../../go-go-gepa/cmd/gepa-runner/plugin_loader.go
      Note: |-
        Plugin hook loading and decoding for phase 3
        Phase 3 plugin extension hook loading and decoding
    - Path: ../../../../../../../go-go-gepa/cmd/gepa-runner/scripts/toy_math_optimizer.js
      Note: Hook usage examples
    - Path: ../../../../../../../go-go-gepa/pkg/optimizer/gepa/config.go
      Note: |-
        Merge scheduler configuration knobs
        Phase 3 merge scheduler knobs
    - Path: ../../../../../../../go-go-gepa/pkg/optimizer/gepa/optimizer.go
      Note: |-
        Merge due scheduling, event hooks, component and side-info hooks
        Phase 3 merge scheduling and hook implementation
ExternalSources: []
Summary: Summary of phase 3 optional alignment features implemented after phase 1-2 port/hardening.
LastUpdated: 2026-02-23T12:45:00-05:00
WhatFor: Capture exactly what was implemented for optional phase 3 and how it can be used.
WhenToUse: Use when enabling advanced merge scheduling, plugin-driven component routing, or event observability.
---


# Phase 3 Alignment Extensions Report

## Delivered scope

Phase 3 is complete and implemented in commit `d9a6e75` in `go-go-gepa`.

This phase delivered four optional alignment enhancements that were left open after Phase 1-2:

- Python-like merge scheduling support (`stagnation_due` with `merges_due` counter)
- explicit optimizer event hooks for merge/mutate lifecycle observability
- plugin extension points for component selection and component-side side-info shaping
- explicit seedless initialization mode based on plugin-provided initial candidate

## 1) Merge scheduling aligned with Python intent

Implemented in:

- `go-go-gepa/pkg/optimizer/gepa/config.go`
- `go-go-gepa/pkg/optimizer/gepa/optimizer.go`

New config knobs:

- `MergeScheduler`: `probabilistic` (default) or `stagnation_due`
- `MaxMergesDue`: cap for internal due counter

Behavior summary:

- `probabilistic` keeps existing behavior (`MergeProbability` each iteration)
- `stagnation_due` increments due counter on non-accepted iterations and spends due merges when possible

This preserves the lightweight architecture while introducing a Python-style “merge after stagnation” behavior.

## 2) Optimizer event observability hooks

Implemented in:

- `go-go-gepa/pkg/optimizer/gepa/optimizer.go`
- `go-go-gepa/cmd/gepa-runner/main.go`

New event model:

- `mutate_attempted`, `mutate_accepted`, `mutate_rejected`
- `merge_attempted`, `merge_accepted`, `merge_rejected`

Event payload includes:

- iteration
- parent/child IDs
- operation
- updated keys
- parent/baseline/child scores
- accepted flag
- calls used

CLI exposure:

- `--show-events` emits events during optimization.

## 3) JS plugin extension points (optimize-anything style)

Implemented in:

- `go-go-gepa/cmd/gepa-runner/plugin_loader.go`
- `go-go-gepa/cmd/gepa-runner/main.go`

Optional plugin hooks now supported:

- `selectComponents(...)` (or alias `chooseComponents`)
- `componentSideInfo(...)` (aliases `sideInfoForComponent`, `buildSideInfo`)
- `initialCandidate(...)` (alias `getInitialCandidate`)

Optimizer integration:

- `SetComponentSelectorFunc(...)`
- `SetSideInfoFunc(...)`

This allows JS plugins to influence which component is optimized each iteration and how side-info context is shaped per component.

## 4) Seedless initialization mode decision and implementation

Decision: implemented, but only with explicit plugin support.

New runner flag:

- `--seedless`

Behavior:

- if no `--seed`, `--seed-file`, or `--seed-candidate` is provided and `--seedless` is set, runner calls plugin `initialCandidate()`
- if plugin does not provide a candidate or returns empty candidate, runner fails with explicit error

This avoids implicit weak defaults and keeps initialization policy in task-specific plugin code.

## Validation

Executed and passed:

```bash
GOWORK=off GOTOOLCHAIN=go1.25.7 go test ./... -count=1
GOWORK=off GOTOOLCHAIN=go1.25.7 make lint
```

## Example usage

```bash
./gepa-runner optimize \
  --script ./cmd/gepa-runner/scripts/toy_math_optimizer.js \
  --seedless \
  --merge-prob 0.25 \
  --merge-scheduler stagnation_due \
  --max-merges-due 3 \
  --show-events \
  --max-evals 120 \
  --batch-size 8 \
  --profile 4o-mini
```

## Residual follow-ups (optional)

- persist event stream into run recorder tables for post-run analysis
- document plugin hook contracts in `cmd/gepa-runner/README.md` with concrete JSON examples
