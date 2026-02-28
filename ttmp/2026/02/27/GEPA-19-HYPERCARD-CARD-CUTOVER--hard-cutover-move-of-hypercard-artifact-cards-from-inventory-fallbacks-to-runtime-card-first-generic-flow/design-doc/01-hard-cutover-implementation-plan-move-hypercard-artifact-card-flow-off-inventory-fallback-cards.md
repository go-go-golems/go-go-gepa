---
Title: 'Hard Cutover Implementation Plan: Move HyperCard Artifact/Card Flow off Inventory Fallback Cards'
Ticket: GEPA-19-HYPERCARD-CARD-CUTOVER
Status: active
Topics:
    - js-vm
    - hypercard
    - go-go-os
    - inventory-app
    - arc-agi
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../go-go-app-inventory/apps/inventory/src/domain/pluginBundle.vm.js
      Note: |-
        Remove reportViewer/itemViewer implementations during hard cutover
        Remove fallback card implementations
    - Path: ../../../../../../../go-go-app-inventory/apps/inventory/src/domain/stack.ts
      Note: |-
        Remove static inventory fallback viewer cards from stack metadata
        Remove fallback card metadata entries
    - Path: ../../../../../../../go-go-os/packages/engine/src/hypercard/artifacts/artifactRuntime.test.ts
      Note: |-
        Update tests for hard cutover behavior
        Cutover test coverage target
    - Path: ../../../../../../../go-go-os/packages/engine/src/hypercard/artifacts/artifactRuntime.ts
      Note: |-
        Remove inventory/template fallback routing; require runtime-card-first opening
        Primary fallback-to-runtime cutover target
    - Path: ../../../../../../../go-go-os/packages/engine/src/hypercard/timeline/hypercardCard.tsx
      Note: |-
        Keep card timeline rendering aligned with runtime-card-first open behavior
        Runtime-card-first artifact opening behavior target
    - Path: ../../../../../../../go-go-os/packages/engine/src/hypercard/timeline/hypercardWidget.test.ts
      Note: |-
        Update timeline test coverage for new runtime-card-only behavior
        Timeline behavior test target
    - Path: ../../../../../../../go-go-os/packages/engine/src/hypercard/timeline/hypercardWidget.tsx
      Note: |-
        Remove template editor fallback and gate open/edit on runtimeCardId
        Template fallback edit path removal target
ExternalSources: []
Summary: Detailed hard-cutover plan to move hypercard artifact/card opening away from inventory template fallback cards and enforce runtime-card-first generic behavior.
LastUpdated: 2026-02-28T00:30:00-05:00
WhatFor: Execute a deliberate breaking migration that removes inventory-specific fallback cards and template-based routing from the HyperCard runtime flow.
WhenToUse: Use when implementing the GEPA-19 migration and validating runtime-card-first behavior across engine and inventory app surfaces.
---


# Hard Cutover Implementation Plan

## Executive Summary

This ticket performs a hard cutover of HyperCard artifact/card opening behavior. We are removing inventory-specific fallback routing (`templateToCardId`, default `stackId: inventory`, static `reportViewer`/`itemViewer` cards) and enforcing a runtime-card-first model.

After this cutover:

1. Artifact windows open only when `runtimeCardId` is present.
2. Timeline widget/card renderers no longer rely on template-driven fallback card IDs.
3. Inventory app no longer ships generic fallback viewer card implementations.
4. Missing runtime card IDs are treated as unsupported instead of silently routed into inventory placeholders.

No backward compatibility is required.

## Problem Statement

The current flow still carries inventory-era fallback assumptions:

1. Engine artifact open helper maps templates into inventory card IDs (`reportViewer`/`itemViewer`).
2. Engine defaults missing stack IDs to `inventory`.
3. Inventory stack and VM bundle contain static fallback viewer cards that act as generic artifact catch-alls.
4. This makes HyperCard behavior non-generic and binds runtime card delivery to inventory artifacts.

These assumptions block clean multi-app usage (ARC and future apps) and create ambiguous ownership between engine and one specific app.

## Proposed Solution

Perform a hard cutover in two codebases (`go-go-os`, `go-go-app-inventory`) with explicit breaking changes:

1. In engine artifact runtime:
   - remove `templateToCardId` and template-icon fallback logic,
   - require `runtimeCardId` to build artifact open payloads,
   - stop defaulting `stackId` to `inventory`; use provided stack or explicit runtime default contract.
2. In engine timeline renderers:
   - for widget/card entities, enable open/edit only when runtime card identity exists,
   - remove template-based code editor routing.
3. In inventory app:
   - remove `reportViewer` and `itemViewer` card metadata from stack,
   - remove `reportViewer` and `itemViewer` handler/render definitions from VM bundle.
4. Update tests to validate hard-cutover behavior and remove template fallback expectations.

## Detailed Design Decisions

1. Break fast on missing `runtimeCardId`.
Reason: fallback behavior hides producer defects and keeps inventory coupling alive.

2. Remove inventory-specific card IDs from engine routing code.
Reason: engine must stay app-agnostic.

3. Keep runtime-card registration/injection path unchanged.
Reason: this is already the correct generic mechanism (`runtimeCardId` + `runtimeCardCode`).

4. Remove inventory fallback cards entirely rather than deprecating.
Reason: user requested hard cutover and no backward compatibility.

## Alternatives Considered

1. Soft deprecation with warnings and dual path.
Rejected due explicit hard-cutover requirement and ongoing conceptual complexity.

2. Keep template fallback in engine but move fallback cards to another app.
Rejected because it still preserves non-generic routing semantics.

3. Add bridge layer that synthesizes runtime cards when missing.
Rejected as hidden magic; producers should emit proper runtime cards.

## Granular Implementation Plan

### Phase A: Ticket/Planning artifacts

1. Create GEPA-19 workspace.
2. Author this implementation plan.
3. Build granular task list with checkboxes.
4. Start chronological diary.

### Phase B: Engine hard cutover (go-go-os)

1. Update `artifactRuntime.ts`:
   - remove `templateToCardId` export and usage,
   - remove template-based icon fallback,
   - require `runtimeCardId` in `buildArtifactOpenWindowPayload`.
2. Update `hypercardWidget.tsx`:
   - remove template-based edit flow,
   - open/edit buttons only when runtime card id is present.
3. Update `hypercardCard.tsx` open behavior if needed for strict runtime-card-first semantics.
4. Update tests:
   - `artifactRuntime.test.ts`
   - `hypercardWidget.test.ts`
   - any dependent snapshot/tests.
5. Run targeted test suite and fix regressions.
6. Commit phase B changes with a dedicated commit.

### Phase C: Inventory fallback removal (go-go-app-inventory)

1. Remove `reportViewer`/`itemViewer` from `INVENTORY_CARD_META` in `stack.ts`.
2. Remove `reportViewer`/`itemViewer` implementations from `pluginBundle.vm.js`.
3. Run inventory test/lint/build checks used by repo.
4. Commit phase C changes with a dedicated commit.

### Phase D: Ticket completion and validation

1. Update tasks checklist with completed status per task.
2. Append diary entries with exact commands, failures, and commit hashes.
3. Update changelog with completed phases and related files.
4. Relate files in ticket docs via `docmgr doc relate`.
5. Run `docmgr doctor --ticket GEPA-19-HYPERCARD-CARD-CUTOVER --stale-after 30`.
6. Mark ticket ready for next follow-up (ARC-side producer validation).

## Pseudocode Diff Sketch

```ts
// before
const cardId = input.runtimeCardId ?? templateToCardId(input.template);
const stackId = (input.stackId ?? 'inventory').trim() || 'inventory';

// after (hard cutover)
const runtimeCardId = cleanString(input.runtimeCardId);
if (!runtimeCardId) return undefined;
const stackId = cleanString(input.stackId) ?? 'runtime';
const cardId = runtimeCardId;
```

```tsx
// before widget renderer
const cardId = templateToCardId(template);
openCodeEditor(cardId)

// after widget renderer
if (!runtimeCardId) return;
openCodeEditor(runtimeCardId)
```

## Validation Matrix

1. Artifact with runtime card id:
   - opens runtime card window,
   - uses runtime card id for editor.
2. Artifact without runtime card id:
   - open button hidden or no-op by design,
   - no fallback viewer card open.
3. Inventory app stack:
   - no `reportViewer`/`itemViewer` cards left.
4. Engine tests:
   - no expectations for template fallback routing.

## Risks and Mitigations

1. Risk: existing widget flows without runtime cards become non-openable.
Mitigation: expected hard-cutover behavior; explicitly documented.

2. Risk: unhandled references to removed helper (`templateToCardId`).
Mitigation: full ripgrep scan + compile/test pass.

3. Risk: hidden reliance on removed inventory cards.
Mitigation: repo-wide search for `reportViewer|itemViewer` and fail if still referenced in runtime paths.

## Post-Cutover Follow-up (not required in this ticket)

1. Ensure ARC-side card producers always emit `hypercard.card.v2` with card id/code.
2. Add explicit stack-id field in producer payloads if multi-stack artifact opening is needed.
3. Expand event-viewer telemetry to flag dropped open attempts due missing runtime card id.
