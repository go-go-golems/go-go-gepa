---
Title: UI DSL architecture audit and proposal for card VM widgets
Ticket: GEPA-25-UI-DSL-EVOLUTION
Status: active
Topics:
    - architecture
    - frontend
    - go-go-os
    - hypercard
    - js-vm
    - ui
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/confirm-runtime/src/components/ConfirmRequestWindowHost.tsx
      Note: |-
        Real composition host that maps structured widget types to modern widget components
        Composition host demonstrating structured widget-type to component mapping
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/PluginCardRenderer.tsx
      Note: |-
        React rendering semantics for each UI node kind
        Maps UI nodes into rendered React widgets
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx
      Note: Runtime session lifecycle, render/event execution, and global-state projection
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/widgets/index.ts
      Note: |-
        Engine widget inventory (33 exported widgets)
        Current engine widget inventory used for DSL expansion proposals
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/plugin-runtime/stack-bootstrap.vm.js
      Note: |-
        VM-exposed UI DSL helper surface (`ui.*`) and stack host contract
        Defines VM helper DSL surface and stack host contract
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/plugin-runtime/uiSchema.ts
      Note: |-
        Runtime UI validation rules and supported node kinds
        Runtime validation rules and supported node kinds
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/plugin-runtime/uiTypes.ts
      Note: |-
        Runtime UI node type contract consumed by host renderer
        Declares runtime UI node union used by renderer and validation
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/theme/desktop/tokens.css
      Note: Theme token system and confirm/widget look primitives
ExternalSources: []
Summary: Evidence-based intern guide and proposal for evolving card VM UI DSL into a concise but more powerful model that reuses modern engine widgets.
LastUpdated: 2026-02-28T19:45:00Z
WhatFor: Align runtime card authoring ergonomics with newer engine widget capabilities while preserving sandbox safety and a small DSL surface.
WhenToUse: Use when implementing or reviewing card VM UI DSL changes, renderer updates, or migration planning for richer runtime card UIs.
---


# UI DSL architecture audit and proposal for card VM widgets

## Executive Summary

The current VM UI DSL is intentionally small and stable, but it is now significantly behind the capabilities of the engine widget layer.

Observed facts:

1. Card VMs expose 8 UI helpers (`text`, `button`, `input`, `row`, `column`, `panel`, `badge`, `table`) in `stack-bootstrap.vm.js`.
2. Runtime schema and renderer support only a small node set, with partial drift (`counter` exists in schema/renderer but is not exposed as a VM helper).
3. Engine widget surface now exports 33 reusable widgets (`components/widgets/index.ts`), including strong confirm-oriented widgets added recently (selectable lists/tables, schema form renderer, image/grid/rating/upload primitives).
4. Confirm runtime already proves a concise composition pattern that maps structured widget types to those components in one host (`ConfirmRequestWindowHost.tsx`).
5. Runtime card renderer still uses mostly inline styles and raw HTML controls rather than engine `data-part` primitives, so runtime-card look and behavior diverge from the newer widget system.

Proposal in one sentence:

Keep the DSL concise by preserving the current primitive nodes, then add one extensible `widget` node family (with a small curated initial set) and migrate rendering to engine visual primitives/tokens, not ad hoc inline styling.

This gives us power without turning the DSL into a full React clone.

---

## Problem Statement and Scope

### Problem

We want VM-authored cards to be able to express richer interactions elegantly, while keeping the DSL easy to learn, safe, and maintainable.

Today there is a mismatch:

1. The DSL is minimal and mostly suitable for simple forms and status panels.
2. The engine has already invested in richer, accessible, tokenized widgets.
3. Runtime-card rendering does not fully inherit that richer widget look/behavior model.

### Scope of this proposal

In scope:

1. UI DSL shape exposed to card VMs (`ui.panel`, `ui.text`, etc.).
2. Runtime UI node schema and renderer contract.
3. Mapping selected newer engine widgets into DSL-friendly forms.
4. Visual/look alignment via existing parts/tokens primitives.
5. Migration strategy with compatibility and testing.

Out of scope:

1. Replacing QuickJS runtime architecture.
2. Full redesign of confirm-runtime protocol.
3. Supporting every engine widget directly in VM DSL.
4. Arbitrary user CSS or direct React component injection from VM code.

---

## Fundamentals for New Interns

This section explains the end-to-end runtime path from VM code to rendered UI.

## 1) Where the `ui.*` DSL comes from

The VM helper API is defined in `packages/engine/src/plugin-runtime/stack-bootstrap.vm.js`.

Key lines:

1. DSL helper object: lines 1-32.
2. Stack registration through `defineStackBundle(...)`: lines 37-43.
3. Render entrypoint `__stackHost.render(...)`: lines 148-155.
4. Event entrypoint `__stackHost.event(...)`: lines 157-217.

Current helper methods exposed to VM code:

1. `ui.text(content)`
2. `ui.button(label, props)`
3. `ui.input(value, props)`
4. `ui.row(children)`
5. `ui.column(children)`
6. `ui.panel(children)`
7. `ui.badge(text)`
8. `ui.table(rows, { headers })`

These are the author-facing primitives card scripts use.

## 2) Runtime schema and type validation

The render result from QuickJS is validated before React rendering.

Key files:

1. `plugin-runtime/uiTypes.ts`: node type union.
2. `plugin-runtime/uiSchema.ts`: runtime validator.

Important behavior:

1. Unsupported `kind` throws at validation (`uiSchema.ts:93`).
2. Strict checks exist for event refs and key props.
3. Unknown kinds are rejected early before renderer.

## 3) Host rendering and event dispatch

`PluginCardSessionHost` manages runtime lifecycle and invokes the VM.

Key responsibilities in `PluginCardSessionHost.tsx`:

1. Start runtime session and load stack bundle (`105-221`).
2. Project global state for VM render (`33-64`, `263-286`).
3. Call VM render and event entrypoints (`288-313`, `333-378`).
4. Route emitted intents to reducer/domain/system handlers.

`PluginCardRenderer.tsx` converts `UINode` trees into React elements.

Key behavior:

1. Switch-style rendering by `node.kind` (`39+`).
2. `button` maps to `Btn`; `input` maps to raw `<input>`.
3. Layout and table rendering currently use inline styles.

## 4) Intent pipeline and capability policy

Card handlers emit intents via `dispatchCardAction`, `dispatchSessionAction`, `dispatchDomainAction`, `dispatchSystemCommand`.

Routing path:

1. VM returns intents from `__stackHost.event`.
2. Host calls `dispatchRuntimeIntent(...)`.
3. Reducer ingests timeline/pending intents (`pluginCardRuntimeSlice.ts`).
4. Domain/system commands are authorized by capability policy before dispatch.

Key files:

1. `components/shell/windowing/pluginIntentRouting.ts`
2. `features/pluginCardRuntime/pluginCardRuntimeSlice.ts`
3. `features/pluginCardRuntime/capabilityPolicy.ts`

This matters because richer widgets often imply richer events; we must preserve this safety gate.

---

## Current UI DSL Audit

## 1) Current node surface

From `uiTypes.ts`, node kinds are:

1. Layout: `panel`, `row`, `column`
2. Display: `text`, `badge`
3. Controls: `button`, `input`
4. Data/utility: `table`

## 2) Contract drift and behavior gaps

### A) `counter` exists in schema/renderer but is dead surface

Evidence:

1. `uiTypes.ts` includes `kind: 'counter'`.
2. `uiSchema.ts` validates `counter`.
3. `PluginCardRenderer.tsx` renders `counter`.
4. Stack VM helpers do not expose `ui.counter(...)`.

Effect:

1. DSL surface and runtime contract are out of sync.
2. This node has no demonstrated usage and should be removed instead of expanded.

### B) `button.props.variant` is typed but effectively ignored

Evidence:

1. `uiTypes.ts` defines `button.props.variant?: string`.
2. `PluginCardRenderer.tsx` creates `<Btn ...>` but does not pass `variant`.
3. `Btn.tsx` supports `variant: 'default' | 'primary' | 'danger'` and maps it to `data-variant`.

Effect:

1. DSL suggests a styling option that does not work in runtime cards.

### C) Renderer look is partly outside parts/tokens system

Evidence:

1. `PluginCardRenderer.tsx` uses many inline styles for `panel`, `row`, `column`, `badge`, `table`.
2. `input` renders raw `<input>` with no `data-part` hooks.
3. Engine style architecture is strongly token/part based (`parts.ts`, `tokens.css`, `primitives.css`).

Effect:

1. Runtime-card visuals diverge from newer widget system consistency.
2. Harder to theme globally and harder to evolve appearance safely.

### D) Table is minimal and not composition-friendly

Current `ui.table` only supports:

1. `headers: string[]`
2. `rows: unknown[][]`

No built-in:

1. row selection model,
2. column formatting metadata,
3. search,
4. action callbacks.

By comparison, `SelectableDataTable` supports row keys, selection mode, search, custom column formatting/cell state.

---

## Engine Widget Inventory and New Additions

## 1) Engine widget breadth

`components/widgets/index.ts` exports 33 widgets.

These include:

1. baseline UI primitives (`Btn`, `DataTable`, `FormView`, `Toast`, etc.),
2. structural controls (`TabControl`, `DropdownMenu`, `ContextMenu`, `ToolPalette`),
3. confirm-focused interaction widgets (`SelectableList`, `SelectableDataTable`, `SchemaFormRenderer`, `ImageChoiceGrid`, `RatingPicker`, `GridBoard`, `FilePickerDropzone`, `RequestActionBar`).

## 2) Recent additions (evidence from git log)

From `go-go-os` widget commit history:

1. `48c2724` (2026-02-23): reusable confirm-oriented widget primitives.
2. `d2aa8fa` (2026-02-23): added `RatingPicker` and `GridBoard` and host usage.
3. `f747e66` (2026-02-24): menu runtime and context menu foundation.
4. `f336c9c` (2026-02-27): context menu visual polish.

## 3) Composition proof: confirm runtime host

`ConfirmRequestWindowHost.tsx` is a strong pattern reference.

It dispatches structured widget types to engine components:

1. `select` -> `SelectableList`
2. `form` -> `SchemaFormRenderer`
3. `table` -> `SelectableDataTable`
4. `image` -> `ImageChoiceGrid`
5. `rating` -> `RatingPicker`
6. `grid` -> `GridBoard`
7. `upload` -> `FilePickerDropzone`
8. with `RequestActionBar` as shared action shell

This is exactly the kind of concise-but-powerful composition model we can adapt for card VM DSL evolution.

---

## Gap Analysis (Against Target Outcome)

Target outcome requested:

1. concise,
2. elegant,
3. powerful,
4. does not attempt to cover all possibilities.

Current state vs target:

1. Concise: yes.
2. Elegant: partially, but contract drift and style inconsistency reduce elegance.
3. Powerful: no, compared to available widget capabilities.
4. Scoped: yes, but now underpowered for typical interactive card needs.

Main gaps:

1. DSL/renderer drift (`counter` dead branch, `button.variant` pass-through gap).
2. Primitive-only interaction model where compound widgets already exist and are battle-tested.
3. Runtime-card look is less aligned with tokenized parts architecture.
4. No formal extension pattern for richer widgets that preserves schema safety.

---

## Proposed DSL Evolution (Concise + Powerful)

Design principle:

Keep simple things simple. Add a small extension point for richer widgets. Do not expose all engine widgets directly.

## 1) Keep primitive v1 API minimal

Preserve existing helpers for compatibility:

1. `panel`, `row`, `column`, `text`, `badge`, `button`, `input`, `table`.

## 2) Add one extensible node family: `kind: 'widget'`

Instead of exploding top-level kinds, add:

```ts
{ kind: 'widget', widget: WidgetKind, props: WidgetPropsByKind[WidgetKind] }
```

Initial `WidgetKind` (curated set only):

1. `select` -> wraps `SelectableList`
2. `tableSelect` -> wraps `SelectableDataTable`
3. `grid` -> wraps `GridBoard`
4. `fileUpload` -> wraps `FilePickerDropzone`

Explicitly excluded from VM DSL scope in this revision:

1. `schemaForm`
2. `imageChoice`
3. `rating`
4. `actionBar`

This is small enough to stay elegant, but strong enough to cover most real operator-card interactions.

## 3) Author ergonomics: helper sugar on top of `widget`

Expose sugar methods in VM helper object, but compile each to one `widget` node:

1. `ui.select(props)`
2. `ui.tableSelect(props)`
3. `ui.grid(props)`
4. `ui.fileUpload(props)`

This keeps author code readable while keeping runtime schema compact.

## 4) Appearance contract: constrained look props (not arbitrary CSS)

Add optional `appearance` bag on most nodes:

```ts
appearance?: {
  tone?: 'default' | 'muted' | 'success' | 'warning' | 'danger';
  density?: 'compact' | 'default' | 'comfortable';
  emphasis?: 'low' | 'normal' | 'high';
}
```

Rules:

1. No raw style object in VM payload.
2. Renderer maps appearance to `data-variant` / `data-state` / part classes.
3. Final look remains token-controlled.

This gives visual control without opening uncontrolled CSS complexity.

## 5) Renderer alignment with engine look system

Update `PluginCardRenderer` so primitive nodes also use parts/tokens conventions.

Examples:

1. `input` -> use `data-part="field-input"`.
2. `panel`/`row`/`column` -> use runtime-card part wrappers instead of inline styles.
3. `table` -> either map to `DataTable` widget or add part-driven markup.
4. `button` -> pass through `variant` to `Btn`.

---

## API Sketch

Illustrative (not final):

```ts
type WidgetKind =
  | 'select'
  | 'tableSelect'
  | 'grid'
  | 'fileUpload';

type UINodeV2 =
  | PrimitiveNodesV1
  | {
      kind: 'widget';
      widget: WidgetKind;
      props: Record<string, unknown>;
      appearance?: Appearance;
    };
```

VM helper sugar:

```js
ui.select({
  items: [{ id: 'p1', label: 'Plan A' }, { id: 'p2', label: 'Plan B' }],
  mode: 'single',
  searchable: true,
  onChange: { handler: 'setPlan' },
});

ui.tableSelect({
  columns: ['id', 'title', 'priority'],
  rows: [
    ['a1', 'Ship release notes', 'high'],
    ['a2', 'Refactor renderer styles', 'med'],
  ],
  selectionMode: 'single',
  onChange: { handler: 'selectTicket' },
});
```

---

## Runtime Flow (Proposed)

Pseudo-flow:

```ts
// VM side
ui.select(props) -> { kind: 'widget', widget: 'select', props }

// host validation
validateUINode(tree):
  if kind === 'widget': validate widget discriminator and schema

// renderer
switch node.kind:
  case 'widget':
    switch node.widget:
      case 'select':
        return <SelectableList ... />
      case 'tableSelect':
        return <SelectableDataTable ... />
      ...

// event bridge
widget callbacks emit existing UIEventRef handlers
-> PluginCardRenderer onEvent(handler, args)
-> PluginCardSessionHost.eventCard(...)
-> intents pipeline unchanged
```

Key property:

No changes needed to intent architecture, capability policy, or QuickJS session lifecycle.

---

## Implementation Plan (Phased)

## Phase 0: Correctness and alignment baseline

1. Remove dead `counter` node support from `uiTypes`, `uiSchema`, and renderer.
2. Pass `button.props.variant` through `PluginCardRenderer` into `Btn`.
3. Add schema/renderer/helper parity tests for all primitive nodes.

Files:

1. `plugin-runtime/uiTypes.ts`
2. `plugin-runtime/uiSchema.ts`
3. `components/shell/windowing/PluginCardRenderer.tsx`
4. `plugin-runtime/uiSchema.test.ts`
5. `plugin-runtime/runtimeService.integration.test.ts`

## Phase 1: Introduce `widget` node and two high-value widgets

Start small:

1. `select` (`SelectableList`)
2. `tableSelect` (`SelectableDataTable`)

Why these first:

1. They cover many data navigation workflows.
2. They already have mature components and usage in confirm runtime.

## Phase 2: Add data and media selection set

1. `tableSelect`
2. `grid`
3. `fileUpload`

Gate each by dedicated schema tests + runtime stories.

## Phase 3: Look/system alignment and migration helpers

1. Remove most inline styles from primitive rendering.
2. Add runtime-card parts/tokens hooks.
3. Provide migration notes and optional codemod for common patterns.

---

## Testing and Validation Strategy

## 1) Contract tests

1. Validate every helper output shape in bootstrap-level tests.
2. Ensure `uiSchema` rejects malformed widget payloads with precise errors.

## 2) Integration tests

1. Extend `runtimeService.integration.test.ts` with fixture stacks that render each new widget kind.
2. Verify event callbacks return expected intent payloads.

## 3) Renderer behavior tests

1. Add focused tests for `PluginCardRenderer` widget-kind mappings.
2. Validate `variant` propagation and appearance mapping.

## 4) Storybook verification

1. Add plugin-runtime stories that render primitive and widget nodes.
2. Confirm token/theme parity with core widget stories.

## 5) Manual smoke

1. Existing card stacks (`BookTracker`, ARC demo card) continue to render unchanged.
2. New sample stack uses `select` and `tableSelect` widgets end-to-end.

---

## Risks, Alternatives, and Open Questions

## Risks

1. Schema creep if too many widget kinds are added too quickly.
2. Event payload inconsistency across widgets if callback contracts are not standardized.
3. Performance regressions for large tables/grids rendered in VM cards.

Mitigations:

1. Curated initial widget set only.
2. Shared callback payload conventions (`onChange`, `onSubmit`, `onPrimary`, etc.).
3. Add rendering performance checks in integration stories/tests.

## Alternatives considered

### Alternative A: Keep primitive DSL only

Pros:

1. Minimal complexity.

Cons:

1. Continues capability mismatch.
2. Encourages ad hoc, repetitive card UI code.
3. Leaves newer widget investment underused for VM cards.

Rejected.

### Alternative B: Expose all 33 widgets directly in DSL

Pros:

1. Maximum power.

Cons:

1. Large API surface.
2. Hard onboarding and maintenance burden.
3. High compatibility risk.

Rejected.

### Alternative C: One generic `ui.custom(component, props)`

Pros:

1. Very flexible.

Cons:

1. Weak static/schema safety.
2. Hard to version and document.
3. Easy to violate runtime constraints.

Rejected for now.

## Open questions

1. Should `fileUpload` remain in VM DSL or require additional capability gating?
2. Should table formatting DSL allow simple per-column format hints (`number`, `currency`, `date`) initially?

---

## Recommended Initial Scope

To stay concise and elegant, ship this first:

1. Phase 0 parity fixes (remove dead `counter` + button variant wiring + renderer part alignment).
2. `kind: 'widget'` infrastructure.
3. Only two widget kinds in first rollout:
   - `select`
   - `tableSelect`

Then evaluate adoption and only add next widgets when a concrete use case appears.

This keeps the DSL small and coherent while unlocking significantly better UI quality.

---

## References

Core runtime and DSL:

1. `go-go-os/packages/engine/src/plugin-runtime/stack-bootstrap.vm.js`
2. `go-go-os/packages/engine/src/plugin-runtime/uiTypes.ts`
3. `go-go-os/packages/engine/src/plugin-runtime/uiSchema.ts`
4. `go-go-os/packages/engine/src/plugin-runtime/runtimeService.ts`
5. `go-go-os/packages/engine/src/plugin-runtime/runtimeService.integration.test.ts`

Host rendering and intent routing:

1. `go-go-os/packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx`
2. `go-go-os/packages/engine/src/components/shell/windowing/PluginCardRenderer.tsx`
3. `go-go-os/packages/engine/src/components/shell/windowing/pluginIntentRouting.ts`
4. `go-go-os/packages/engine/src/features/pluginCardRuntime/pluginCardRuntimeSlice.ts`
5. `go-go-os/packages/engine/src/features/pluginCardRuntime/selectors.ts`
6. `go-go-os/packages/engine/src/features/pluginCardRuntime/capabilityPolicy.ts`

Widget system and look:

1. `go-go-os/packages/engine/src/components/widgets/index.ts`
2. `go-go-os/packages/engine/src/components/widgets/SelectableList.tsx`
3. `go-go-os/packages/engine/src/components/widgets/SelectableDataTable.tsx`
4. `go-go-os/packages/engine/src/components/widgets/GridBoard.tsx`
5. `go-go-os/packages/engine/src/components/widgets/FilePickerDropzone.tsx`
6. `go-go-os/packages/engine/src/components/widgets/ContextMenu.tsx`
7. `go-go-os/packages/engine/src/theme/desktop/tokens.css`
8. `go-go-os/packages/engine/src/theme/desktop/primitives.css`

Composition reference:

1. `go-go-os/packages/confirm-runtime/src/components/ConfirmRequestWindowHost.tsx`
2. `go-go-os/packages/confirm-runtime/src/types.ts`
3. `go-go-os/packages/confirm-runtime/src/proto/confirmProtoAdapter.ts`
4. `go-go-os/packages/engine/docs/theming-and-widget-playbook.md`
