---
Title: Promise-Aware Plugin Bridge and Streaming Events Implementation Plan
Ticket: GEPA-04-ASYNC-PLUGIN-PROMISES
Status: active
Topics:
    - gepa
    - plugins
    - goja
    - runner
    - events
    - js-bindings
    - go
    - tooling
DocType: planning
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../geppetto/pkg/js/modules/geppetto/api_sessions.go
      Note: Existing JS Promise and streaming API surface
    - Path: cmd/gepa-runner/js_runtime.go
      Note: Runtime/eventloop context used by loader calls
    - Path: cmd/gepa-runner/plugin_loader.go
      Note: Sync optimizer plugin call paths to migrate to Promise-aware settlement
    - Path: pkg/dataset/generator/plugin_loader.go
      Note: Sync dataset plugin call path to migrate to Promise-aware settlement
ExternalSources: []
Summary: Intern-oriented implementation plan and scoping for Promise-returning JS plugins and streaming event propagation.
LastUpdated: 2026-02-26T17:23:46.55087674-05:00
WhatFor: ""
WhenToUse: ""
---


# Promise-Aware Plugin Bridge and Streaming Events Implementation Plan

## 1) Problem Statement

`geppetto` JS supports async inference and event streaming already, but `go-go-gepa` plugin execution paths are synchronous:

- Optimizer/candidate loader calls JS methods and immediately decodes returned values.
- Dataset generator loader does the same for `generateOne`.

Result: plugin authors cannot safely return Promises or expose real-time stream events from JS to CLI/storage.

## 2) Current System (What Exists Today)

### Geppetto JS side (already async-capable)

- `session.runAsync(...)` => `Promise<Turn>`
- `session.start(...)` => `RunHandle` with:
  - `promise`
  - `on(eventType, callback)` for streaming events

Relevant file:
- `/home/manuel/workspaces/2026-02-22/add-gepa-optimizer/geppetto/pkg/js/modules/geppetto/api_sessions.go`

### go-go-gepa side (currently sync plugin bridge)

- Optimizer plugin loader:
  - `/home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/plugin_loader.go`
  - Methods: `Evaluate`, `Run`, `Dataset`, `Merge`, `InitialCandidate`, `SelectComponents`, `ComponentSideInfo`
  - All use synchronous callable invocation and `decodeJSReturnValue(...)`.
- Dataset generator plugin loader:
  - `/home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/pkg/dataset/generator/plugin_loader.go`
  - Method: `GenerateOne`
  - Same sync assumption.
- Shared runtime:
  - `/home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/js_runtime.go`
  - Uses goja + node eventloop + runtimeowner.

## 3) Goals and Non-Goals

### Goals

1. Allow plugin methods to return either:
   - plain value (existing behavior), or
   - `Promise` resolving to a value.
2. Preserve backward compatibility for all existing sync plugins.
3. Add a plugin-facing event emission channel so JS can surface stream events while async work is running.
4. Provide deterministic timeout/cancel behavior for Promise settlement.

### Non-Goals (for GEPA-04 baseline)

1. Full optimizer-loop redesign.
2. Rewriting geppetto APIs.
3. Mandatory DB schema expansion for every event type on day 1 (start with minimal event model first).

## 4) Proposed Design

## 4.1 Add a shared JS return-settlement helper

Create a shared helper in go-go-gepa (suggested new package: `pkg/jsbridge`):

- Input:
  - runtime references (`*goja.Runtime`, eventloop/runner context),
  - raw JS return value,
  - timeout/cancel controls.
- Output:
  - settled Go value (`any`) or error.

Behavior:

1. If return value is not Promise => same as today.
2. If Promise:
   - attach `then`/`catch` handlers on runtime owner thread,
   - wait for settle via channel/condition,
   - honor context deadline/cancel,
   - convert rejection reason to rich error text.
3. After settle, route to existing decode logic (`decodeJSReturnValue` equivalent).

## 4.2 Extend plugin option contracts with event sink

For plugin method options, add optional emitter:

- `options.events.emit(eventObj)` (or `options.emitEvent(eventObj)`).

Host side behavior:

1. Accept emitted event payload as generic object.
2. Add envelope fields on host side:
   - `ticket/run id` (if available),
   - timestamp,
   - plugin id, event sequence.
3. Route to:
   - CLI stream output when `--stream` flag is enabled,
   - optional persistence sink when configured.

Back-compat: absence of emitter should not fail plugins.

## 4.3 Apply to both loaders

1. `cmd/gepa-runner/plugin_loader.go`:
   - update all JS method invocations (`Run`, `Evaluate`, etc.) to use settlement helper.
2. `pkg/dataset/generator/plugin_loader.go`:
   - update `GenerateOne` path to use same helper.

Important: keep return decoding behavior identical after settlement.

## 4.4 CLI and persistence phase split

Phase A (required for ticket completion):
- Promise support + optional live console output from event sink.

Phase B (optional follow-up if time permits):
- Persist stream events to sqlite tables:
  - candidate run events table,
  - dataset generation events table.

## 5) Detailed Implementation Steps (Intern Runbook)

### Step 1: Baseline tests before refactor

Add failing tests that codify expected new behavior:

1. Promise-returning `run()` resolves correctly.
2. Promise-returning `evaluate()` resolves correctly.
3. Promise-returning `generateOne()` resolves correctly.
4. Rejected Promise produces actionable Go error.
5. Hung Promise obeys timeout/cancel.
6. Sync-return plugins still pass unchanged.

### Step 2: Build settlement helper

1. Implement Promise detection (`*goja.Promise`) in helper.
2. Implement settle waiter that is eventloop-safe.
3. Ensure no cross-thread VM access.
4. Unit-test helper directly with:
   - immediate resolve,
   - delayed resolve,
   - rejection,
   - timeout.

### Step 3: Integrate helper in optimizer plugin loader

Touch:
- `cmd/gepa-runner/plugin_loader.go`

Replace direct post-call decode with:
1. settle,
2. decode.

Run tests for candidate/eval/optimizer paths.

### Step 4: Integrate helper in dataset plugin loader

Touch:
- `pkg/dataset/generator/plugin_loader.go`

Same migration pattern as step 3.

### Step 5: Add event emitter in options

1. Add emitter function to options maps passed into plugin calls.
2. Define minimal event schema (required keys):
   - `type` (string),
   - `data` (object or string),
   - `level` (optional),
   - `ts` added by host.
3. Add host-side validation and panic-safe handling.

### Step 6: Wire CLI streaming toggle

1. Add `--stream` flags to relevant commands (`candidate run`, `dataset generate`; optimizer path optional).
2. On event emission, print structured line (JSONL) to stderr/stdout.
3. Keep final command output unchanged to avoid breaking scripts.

### Step 7: Documentation and examples

1. Add JS plugin example scripts demonstrating Promise + `session.start().on(...)`.
2. Update runner README/help text with:
   - sync vs async return contract,
   - recommended timeout defaults,
   - streaming event usage.

## 6) Suggested Milestones and Estimates

1. Milestone M1 (Promise settlement core): 0.5-1 day
2. Milestone M2 (Both loaders integrated + tests): 0.5-1 day
3. Milestone M3 (Event sink + CLI streaming): 0.5-1 day
4. Milestone M4 (Docs/examples hardening): 0.25-0.5 day

Total practical estimate: 2-3 days for solid delivery.

## 7) Risks and Mitigations

1. Risk: deadlock when waiting Promise settlement on wrong goroutine.
Mitigation: all Promise inspection/callback attachment through runtime owner.

2. Risk: rejected Promise loses useful context.
Mitigation: include rejection payload and plugin method name in wrapped error.

3. Risk: event streaming breaks deterministic output pipelines.
Mitigation: stream output guarded by explicit `--stream`; retain stable final output.

4. Risk: feature drift into full event-store design.
Mitigation: keep sqlite persistence for emitted events as separate follow-up milestone.

## 8) Acceptance Criteria

1. Promise-returning `run`, `evaluate`, and `generateOne` are supported.
2. Existing sync plugins still pass current tests unchanged.
3. Timeout/cancel behavior is deterministic and tested.
4. At least one example plugin shows real-time event emission during async run.
5. Command docs explain async contract and streaming flag behavior.

## 9) Out of Scope Clarification for Intern

Do not modify:

1. `gepa/` repository code (reference only).
2. `2026-02-18--cozodb-extraction` (reference only).
3. Core geppetto APIs unless absolutely required by blocker.

Primary implementation target is `go-go-gepa`.

## 10) Execution Order Recommendation

1. Promise settlement helper + tests.
2. Optimizer loader integration.
3. Dataset loader integration.
4. Event emitter + CLI streaming.
5. Docs and examples.
