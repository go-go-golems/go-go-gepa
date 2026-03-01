---
Title: Implementation Plan for Code Review Bugfixes
Ticket: GEPA-06-CODE-REVIEW-BUGFIXES
Status: active
Topics:
    - bug
    - gepa
    - optimizer
    - plugins
    - runner
    - events
    - go
DocType: planning
Intent: long-term
Owners: []
RelatedFiles:
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/plugin_loader.go
      Note: Primary optimizer context propagation plan
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/pkg/backendmodule/module.go
      Note: Primary SSE behavior change planned
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/pkg/dataset/generator/plugin_loader.go
      Note: Primary dataset context propagation plan
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/pkg/jsbridge/call_and_resolve.go
      Note: Intentional context fallback reviewed as non-bug
ExternalSources: []
Summary: ""
LastUpdated: 2026-03-01T11:06:18.164643432-05:00
WhatFor: ""
WhenToUse: ""
---


# Implementation Plan for Code Review Bugfixes

## Objective

Fix the three reported review issues while minimizing API churn and preserving existing behavior for successful runs.

## Inputs and Confirmed Findings

- `pkg/backendmodule/module.go:351+` writes SSE headers/body before checking run existence.
- `cmd/gepa-runner/plugin_loader.go:220` calls `jsbridge.CallAndResolve(context.Background(), ...)`.
- `pkg/dataset/generator/plugin_loader.go:162` calls `jsbridge.CallAndResolve(context.Background(), ...)`.
- Additional production scan result:
  - `pkg/jsbridge/call_and_resolve.go:155` uses `context.Background()` only as nil-context fallback in `withDefaultTimeout`, which is intentional and should remain.

## Scope

- In scope:
  - Backend SSE endpoint correctness for unknown `run_id`.
  - Context propagation through optimizer plugin call path.
  - Context propagation through dataset generator call path.
  - Regression tests for each changed behavior.
  - Ticket docs (tasks, changelog, diary) and closure.
- Out of scope:
  - Broad refactor of all plugin APIs beyond context threading.
  - Changes to timeout defaults in `jsbridge`.

## Design and Change Strategy

## 1. SSE Existence Validation Before Stream Start

- File: `pkg/backendmodule/module.go`
- Change:
  - In `handleRunEvents`, validate run existence (`GetRun`) before setting `text/event-stream` headers or writing any bytes.
  - If run is missing, return `404` immediately via `http.NotFound`.
- Test:
  - Add/extend module tests to assert unknown run events endpoint returns `404` and does not include SSE preamble output.

## 2. Optimizer Plugin Context Propagation

- Files:
  - `cmd/gepa-runner/plugin_loader.go`
  - `cmd/gepa-runner/main.go`
  - `cmd/gepa-runner/eval_command.go`
  - `cmd/gepa-runner/candidate_run_command.go`
  - relevant tests in `cmd/gepa-runner/*_test.go`
- Change:
  - Add `ctx context.Context` parameter to `callPluginFunction`.
  - Thread `ctx` into plugin methods that invoke JS bridge (`Dataset`, `Evaluate`, `Run`, `Merge`, `InitialCandidate`, `SelectComponents`, `ComponentSideInfo`).
  - Update call sites to pass command/optimizer callback contexts.
- Compatibility:
  - Keep semantics for non-canceled contexts unchanged.

## 3. Dataset Generator Context Propagation

- Files:
  - `pkg/dataset/generator/plugin_loader.go`
  - `pkg/dataset/generator/generation.go`
  - `pkg/dataset/generator/run.go`
  - `cmd/gepa-runner/dataset_generate_command.go`
  - relevant tests in `cmd/gepa-runner/dataset_generator_loader_test.go`
- Change:
  - Add `ctx context.Context` parameter to `Plugin.GenerateOne`.
  - Thread `ctx` through `GenerateRows` and `RunWithRuntime`.
  - Update command and test call sites.

## 4. Remaining `context.Background()` Review

- Keep `pkg/jsbridge/call_and_resolve.go:withDefaultTimeout` fallback as-is.
- Ensure no other new production `context.Background()` calls are introduced in changed paths.

## Verification Plan

- Unit/integration targets:
  - `go test ./pkg/backendmodule -count=1`
  - `go test ./cmd/gepa-runner -count=1`
  - `go test ./pkg/dataset/generator -count=1` (if tests exist; otherwise no-op package compile)
- Additional check:
  - `rg -n "context\\.Background\\(\\)" --glob '!**/*_test.go' --glob '!ttmp/**'`

## Task Execution Order

1. Fix SSE endpoint pre-stream validation + test.
2. Propagate context through optimizer plugin flow + tests.
3. Propagate context through dataset generator flow + tests.
4. Run test suite and production `context.Background()` scan.
5. Update docs/changelog/diary and close ticket.

## Risks and Mitigations

- Risk: Signature changes break many call sites.
  - Mitigation: Compile/test after each task, update call sites in same commit.
- Risk: Context changes alter behavior in existing tests.
  - Mitigation: Keep default behavior same when context is active; only improve cancellation responsiveness.
