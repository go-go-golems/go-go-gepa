---
Title: UI DSL demo cards architecture, implementation plan, and intern onboarding guide
Ticket: GEPA-28-UI-DSL-DEMO-CARDS
Status: active
Topics:
    - frontend
    - go-go-os
    - hypercard
    - ui-dsl
    - demo
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/apps/hypercard-tools/src/launcher/module.tsx
      Note: HyperCard Tools launcher entrypoint and current launch behavior.
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/hypercard-runtime/src/plugin-runtime/stack-bootstrap.vm.js
      Note: UI DSL factory and VM bootstrap host contract.
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/hypercard-runtime/src/plugin-runtime/uiTypes.ts
      Note: Source-of-truth UI DSL node type contract.
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/hypercard-runtime/src/plugin-runtime/uiSchema.ts
      Note: Runtime validator for UI node trees emitted by plugin cards.
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/hypercard-runtime/src/runtime-host/PluginCardRenderer.tsx
      Note: Renderer mapping from UI DSL nodes to React widgets.
    - Path: /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/hypercard-runtime/src/runtime-host/PluginCardSessionHost.tsx
      Note: Runtime session lifecycle, rendering loop, and event dispatch.
Summary: Detailed intern-focused architecture and implementation plan for delivering HyperCard Tools UI DSL demo cards and wiring them as the default HyperCard Tools launch experience.
LastUpdated: 2026-02-28T18:48:00-05:00
WhatFor: Onboard a new engineer and provide an executable plan to build and maintain a comprehensive UI DSL demo stack.
WhenToUse: Use when implementing or extending UI DSL demo cards, or when debugging HyperCard runtime card behavior end-to-end.
---

# UI DSL demo cards architecture, implementation plan, and intern onboarding guide

## 1. Executive summary

This document explains how HyperCard UI DSL cards currently work, where they live, how they are launched, and how events/state move through the runtime. It then defines an implementation plan to create a full demo stack under HyperCard Tools that showcases the entire active UI DSL surface.

The key scope is:

1. Build a folder-like Home card that acts as a demo catalog.
2. Add a complete set of demo cards covering every supported UI DSL widget kind.
3. Keep the HyperCard runtime-card editor behavior intact.
4. Make HyperCard Tools open into the demo stack by default.

The active UI DSL surface in this codebase (after previous scope reductions) is:

- `ui.panel`
- `ui.row`
- `ui.column`
- `ui.text`
- `ui.badge`
- `ui.button`
- `ui.input`
- `ui.table`

`ui.counter` is out of scope and intentionally unsupported. `schemaForm`, `rating`, `imagechoice`, and `actionbar` are also intentionally excluded from this demo scope.

## 2. Problem statement and goals

### 2.1 Current gap

HyperCard Tools currently launches to a static explanatory app window. It does not open an interactive card stack that helps developers explore UI DSL features quickly.

Evidence:

- In `apps/hypercard-tools/src/launcher/module.tsx`, launch payload uses `content.kind: 'app'` with `appKey: formatAppKey(APP_ID, HOME_INSTANCE)` and the home renderer is static text (`module.tsx:11-29`, `module.tsx:61-71`).

This means:

- no “folder/catalog” user experience for demos,
- no canonical demo cards for widget behavior,
- onboarding friction for new engineers trying to author plugin cards.

### 2.2 Target outcome

When a user clicks HyperCard Tools from the desktop icon, they should land in a full card-stack demo workspace (similar to app-like navigation), with:

- a home catalog card,
- category cards for layout/data/input/actions,
- clear examples of state mutation and intent handling,
- deterministic back/home navigation.

### 2.3 Non-goals

- Reintroducing removed widgets (`counter`, `schemaForm`, `rating`, `imagechoice`, `actionbar`).
- Expanding the UI DSL schema contract in this ticket.
- Backwards-compatibility wrappers for old APIs.

## 3. Fundamentals for an intern: how this system is wired

This section is intentionally explicit for someone new to the codebase.

### 3.1 Where the UI DSL contract is defined

The UI DSL node schema is defined in TypeScript union types:

- `packages/hypercard-runtime/src/plugin-runtime/uiTypes.ts`

The union currently supports exactly five shape families:

1. layout containers: `panel`, `row`, `column`
2. text primitives: `text`, `badge`
3. actions: `button`
4. input: `input`
5. data table: `table`

Runtime validation lives in:

- `packages/hypercard-runtime/src/plugin-runtime/uiSchema.ts`

Any unsupported node kind throws a runtime validation error (for example `counter`). This is already tested in `uiSchema.test.ts`.

### 3.2 How UI DSL helpers are injected in VM

Plugin card JS code does not import React. Instead, it gets a `ui` factory from VM bootstrap code:

- `packages/hypercard-runtime/src/plugin-runtime/stack-bootstrap.vm.js`

At runtime, the bootstrap sets `globalThis.ui = __ui` with methods like `panel`, `text`, `button`, etc. These return plain object nodes (`{ kind: 'text', ... }`) consumed by runtime renderer.

The same bootstrap also provides:

- `defineStackBundle(factory)`
- `defineCard(...)`
- `defineCardRender(...)`
- `defineCardHandler(...)`
- `globalThis.__stackHost` entrypoints for meta/render/event calls.

### 3.3 How render and event loops execute

`PluginCardSessionHost` is the orchestration component:

- `packages/hypercard-runtime/src/runtime-host/PluginCardSessionHost.tsx`

Core flow:

1. Register runtime session in Redux (`registerRuntimeSession`).
2. Load bundle code into QuickJS runtime (`loadStackBundle`).
3. Render current card by calling VM `render` and validating node tree.
4. Render tree via `PluginCardRenderer`.
5. On user event, call VM handler, receive intents, dispatch intents into runtime + Redux.

Important detail: the host projects a `globalState` object into card VM with `domains`, `nav`, `self`, and `system` sections (`PluginCardSessionHost.tsx:33-63`, `263-285`).

### 3.4 How system navigation and toasts work

VM handlers emit intents through helper dispatchers. For navigation and toast behaviors, handlers use `dispatchSystemCommand` with command names like:

- `nav.go`
- `nav.back`
- `notify`
- `window.close`

Intent translation to Redux actions is in:

- `packages/hypercard-runtime/src/runtime-host/pluginIntentRouting.ts`

`nav.go` maps to `sessionNavGo`, `nav.back` maps to `sessionNavBack`, `notify` maps to `showToast`.

### 3.5 How nodes map to React widgets

`PluginCardRenderer` converts node kinds to React components:

- `panel/column/row` -> flex containers
- `text/badge` -> spans
- `button` -> `Btn` from engine
- `input` -> native input with `onChange` payload `{ value }`
- `table` -> native table

Source:

- `packages/hypercard-runtime/src/runtime-host/PluginCardRenderer.tsx`

## 4. Current HyperCard Tools behavior and why we change it

Current behavior in HyperCard Tools launcher:

1. App icon click launches `buildHomeWindowPayload` with `content.kind: 'app'`.
2. `renderWindow` returns static home text for `instanceId === 'home'`.
3. Non-home instances try to decode runtime card editor refs and render `CodeEditorWindow`.

This is useful for editor routing but does not provide a demo stack.

Desired updated behavior:

1. Default app launch opens a card window (`content.kind: 'card'`) pointing to HyperCard Tools demo stack home.
2. A `WindowContentAdapter` renders `PluginCardSessionHost` for this stack.
3. Existing encoded editor app-key windows still render via `renderWindow` for editor flows.

In short, HyperCard Tools should become both:

- a default UI DSL demo stack app,
- and a host for runtime-card editor windows.

## 5. Proposed architecture for demo cards

### 5.1 New files in `apps/hypercard-tools`

Proposed additions:

```text
apps/hypercard-tools/src/domain/
  pluginBundle.authoring.d.ts
  pluginBundle.vm.js
  pluginBundle.ts
  stack.ts
```

### 5.2 Stack identity and card map

Define one stack constant, e.g.:

- `id: 'hypercardToolsUiDslDemo'`
- `homeCard: 'home'`
- `cards`: placeholder metadata for each demo card
- `plugin.bundleCode`: JS VM bundle string imported from `pluginBundle.ts`

### 5.3 Demo card taxonomy

Recommended card set (concise but complete):

1. `home` (folder/catalog index)
2. `layouts` (`panel`, `row`, `column`)
3. `textAndBadges` (`text`, `badge`)
4. `buttons` (action flows + toast)
5. `inputs` (`input` change handling into card/session state)
6. `tables` (`table` with headers/rows)
7. `stateAndNav` (combined state patch/set/reset + nav/back)
8. `allWidgetsPlayground` (one composed sample card)

This covers every active widget kind and shows realistic composition patterns.

### 5.4 Folder-like home card behavior

Home card should behave like a “tool folder”:

- show an index table with `Card`, `Focus`, and `Action` columns,
- show “Open <Demo>” buttons,
- include a quick “Reset demo state” action,
- include explanatory text on what each card demonstrates.

### 5.5 State model

Use both card-local and session state to teach authoring patterns.

- `cardState`: per-card ephemeral fields (`draft`, `counterLikeText`, `filter`, etc.)
- `sessionState`: cross-card values (`activeDemo`, `note`, `visitCount`)

Use action types supported by runtime slice:

- `patch`
- `set`
- `reset`

Pseudo-intent examples:

```js
// card-local patch
ctx.dispatchCardAction('patch', { draft: 'hello' });

// nested set
ctx.dispatchCardAction('set', { path: 'form.title', value: 'Widget Tour' });

// session-wide patch
ctx.dispatchSessionAction('patch', { lastVisited: 'tables' });
```

### 5.6 Navigation patterns

Use explicit helper wrappers in VM bundle:

```js
function go(ctx, cardId, param) {
  ctx.dispatchSystemCommand('nav.go', param ? { cardId, param } : { cardId });
}

function back(ctx) {
  ctx.dispatchSystemCommand('nav.back');
}

function toast(ctx, message) {
  ctx.dispatchSystemCommand('notify', { message });
}
```

This keeps handlers readable and gives interns a pattern to copy.

## 6. Detailed implementation plan

### Phase A: scaffold domain stack and plugin bundle

1. Add `stack.ts` with card metadata and `CardStackDefinition` export.
2. Add `pluginBundle.vm.js` with `defineStackBundle(({ ui }) => ({ ... }))`.
3. Add `pluginBundle.ts` raw-string export pattern matching other apps.
4. Add `pluginBundle.authoring.d.ts` to improve IDE authoring safety.

### Phase B: implement demo cards and handlers

For each card:

1. implement `render` that emits valid `UINode` tree only,
2. add handlers for all buttons/inputs,
3. route navigation through `nav.go`/`nav.back`.

Important constraint: never emit removed/unsupported kinds.

### Phase C: update launcher wiring

In `apps/hypercard-tools/src/launcher/module.tsx`:

1. Create stack session window payload builder returning `content.kind: 'card'`.
2. Add a `WindowContentAdapter` that renders `PluginCardSessionHost` for this stack.
3. Keep `renderWindow` path for encoded editor instances unchanged.
4. Keep unknown-instance fallback for malformed editor keys.

Pseudo-shape:

```tsx
function buildWorkspaceWindowPayload(reason): OpenWindowPayload {
  return {
    id: `window:hypercard-tools:${instanceId}`,
    content: {
      kind: 'card',
      card: {
        stackId: STACK.id,
        cardId: STACK.homeCard,
        cardSessionId: `hypercard-tools-session:${instanceId}`,
      },
    },
  };
}

function createHypercardToolsCardAdapter(): WindowContentAdapter {
  return {
    id: 'hypercard-tools.card-window',
    canRender: (window) => window.content.kind === 'card' && window.content.card?.stackId === STACK.id,
    render: (window) => <PluginCardSessionHost ... />,
  };
}
```

### Phase D: tests and verification

Because launch payload kind changes from `app` to `card` for HyperCard Tools icon launch, tests must be updated.

Primary impact area:

- `wesen-os/apps/os-launcher/src/__tests__/launcherHost.test.tsx`

Expected changes:

- assertions that grouped `hypercard-tools` with app-kind modules should now expect `card` kind for icon launch payload,
- keep tests for encoded editor instance rendering intact.

### Phase E: docs + diary + upload

1. Update ticket tasks as each phase completes.
2. Record command outputs and any failures in diary.
3. Upload design doc and diary bundle to reMarkable.

## 7. Validation strategy

### 7.1 Functional manual checks

1. Launch OS frontend.
2. Click HyperCard Tools icon.
3. Confirm card-based demo home opens (not static text).
4. Navigate every demo card from home.
5. Trigger input/button handlers and confirm updates.
6. Confirm back navigation works.
7. Trigger notify action and verify toast appears.
8. Confirm runtime-card editor still opens for encoded instance IDs.

### 7.2 Automated checks

Minimum:

1. `pnpm --filter @wesen/os-launcher test -- launcherHost.test.tsx`
2. package tests/typecheck in `go-go-os` for touched modules.

If CI-style full run is too slow locally, run targeted suites and note limits in diary.

## 8. Risks and pitfalls

### 8.1 Risk: accidental unsupported DSL node kinds

Cause: demo authoring accidentally reuses old widget names.

Mitigation:

- rely on runtime schema validation (`uiSchema.ts`),
- keep card bundle examples strictly within active kinds.

### 8.2 Risk: launcher regressions for editor windows

Cause: changing launch payload kind can accidentally break `renderWindow` path assumptions.

Mitigation:

- keep `renderWindow` logic for encoded `editor~stack~card` instances,
- add targeted tests for both launch path and editor path.

### 8.3 Risk: state mutation confusion for interns

Cause: difference between card state and session state is subtle.

Mitigation:

- include explicit demo cards that show both side by side,
- keep helper functions and comments in plugin bundle.

## 9. Alternatives considered

### Alternative 1: Keep static HyperCard Tools home and open demos via commands

Rejected because:

- still hides most behavior behind non-discoverable commands,
- weak onboarding value,
- poor parity with “folder-like” experience requested.

### Alternative 2: Add demos as Storybook-only examples

Rejected because:

- does not validate actual VM runtime path,
- misses end-to-end event/intent/session behavior,
- less useful for card authors who need real runtime semantics.

### Alternative 3: Expand DSL surface in the same ticket

Rejected because:

- mixes feature expansion with demo/onboarding scope,
- increases regression risk,
- conflicts with explicit scope reduction decisions made earlier.

## 10. Intern onboarding walkthrough

### Step 1: understand the three contracts

1. Node contract (`uiTypes.ts`)
2. VM authoring contract (`stack-bootstrap.vm.js`)
3. Renderer contract (`PluginCardRenderer.tsx`)

If all three agree, your card will render correctly.

### Step 2: clone a known-good pattern

Use a smaller existing plugin bundle as base pattern, for example:

- `apps/book-tracker-debug/src/domain/pluginBundle.vm.js`

It already uses helper functions, navigation wrappers, and card maps.

### Step 3: implement demos card-by-card

Do not attempt everything in one giant card. Keep one concern per card.

### Step 4: wire launcher only after bundle is stable

This makes debugging easier. Verify stack host first (runtime/renderer), then make it the default launch window.

### Step 5: lock behavior with tests

Update launcher host tests when launch payload semantics change.

## 11. Expected deliverables and acceptance criteria

### Deliverables

1. New HyperCard Tools domain stack + plugin bundle demo cards.
2. Updated launcher wiring and contributions.
3. Updated test expectations in OS launcher suite.
4. Ticket docs (design + diary + tasks + changelog).
5. reMarkable upload confirmation.

### Acceptance criteria

1. HyperCard Tools icon launch opens demo card stack home.
2. All active UI DSL widgets are demonstrated at least once.
3. Editor instances still render correctly.
4. Tests for launcher behavior pass.
5. Docs are complete enough for a new intern to continue work without verbal handoff.

## 12. Implementation checklist snapshot

- [x] Investigation and architecture mapping
- [x] Detailed intern-oriented design plan
- [ ] HyperCard Tools demo stack implementation
- [ ] Launcher/test cutover
- [ ] Final validation and reMarkable uploads

## 13. References

- `/home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/apps/hypercard-tools/src/launcher/module.tsx`
- `/home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/hypercard-runtime/src/plugin-runtime/uiTypes.ts`
- `/home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/hypercard-runtime/src/plugin-runtime/uiSchema.ts`
- `/home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/hypercard-runtime/src/plugin-runtime/stack-bootstrap.vm.js`
- `/home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/hypercard-runtime/src/runtime-host/PluginCardRenderer.tsx`
- `/home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/hypercard-runtime/src/runtime-host/PluginCardSessionHost.tsx`
- `/home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/hypercard-runtime/src/runtime-host/pluginIntentRouting.ts`
- `/home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/apps/book-tracker-debug/src/domain/pluginBundle.vm.js`
- `/home/manuel/workspaces/2026-02-22/add-gepa-optimizer/wesen-os/apps/os-launcher/src/__tests__/launcherHost.test.tsx`
