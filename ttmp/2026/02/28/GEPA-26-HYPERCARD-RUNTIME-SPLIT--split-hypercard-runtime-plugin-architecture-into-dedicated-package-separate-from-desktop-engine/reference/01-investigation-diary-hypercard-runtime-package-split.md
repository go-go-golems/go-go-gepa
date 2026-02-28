---
Title: 'Investigation diary: HyperCard/runtime package split'
Ticket: GEPA-26-HYPERCARD-RUNTIME-SPLIT
Status: active
Topics:
    - architecture
    - frontend
    - go-go-os
    - hypercard
    - js-vm
    - plugins
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../go-go-app-arc-agi-3/apps/arc-agi-player/src/app/store.ts
      Note: Captured downstream package consumption and coupling evidence
    - Path: ../../../../../../../go-go-os/packages/desktop-os/src/store/createLauncherStore.ts
      Note: Captured store ownership and reserved reducer coupling
    - Path: ../../../../../../../go-go-os/packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx
      Note: Captured coupling analysis and session orchestration findings
    - Path: ../../../../../../../go-go-os/packages/engine/src/hypercard/artifacts/artifactProjectionMiddleware.ts
      Note: Captured artifact projection to runtime card registry flow
    - Path: ../../../../../../../go-go-os/packages/engine/src/plugin-runtime/runtimeService.ts
      Note: Captured runtime lifecycle investigation details
    - Path: ttmp/2026/02/28/GEPA-26-HYPERCARD-RUNTIME-SPLIT--split-hypercard-runtime-plugin-architecture-into-dedicated-package-separate-from-desktop-engine/design-doc/01-hypercard-runtime-package-split-architecture-and-migration-guide.md
      Note: Primary design deliverable produced from investigation log
ExternalSources: []
Summary: Chronological investigation log for GEPA-26 package split analysis, design writing, and publishing.
LastUpdated: 2026-02-28T21:20:00Z
WhatFor: Preserve command-level traceability and rationale for future implementation phases.
WhenToUse: Use when validating assumptions, replaying architecture analysis, or onboarding contributors to split execution.
---


# Investigation diary: HyperCard/runtime package split

## Goal

Produce a complete intern-onboarding architecture and migration design for splitting HyperCard/runtime plugin systems from desktop engine concerns, then publish the deliverable in ticket docs and reMarkable.

## Step 1: Ticket bootstrap and scope lock

I started by creating a dedicated GEPA-26 ticket to avoid overloading previous UI DSL tickets. The explicit title anchors this as a package-boundary architecture effort, not a runtime behavior patch.

I then scaffolded two core documents immediately: a design doc and a diary. This let me capture findings incrementally while investigating instead of reconstructing the process afterward.

### Prompt Context

**User prompt (verbatim):** "Create a new ticket and a very verbose detailed design document for our new intern that doesn't know anything about the codebase or the setup or the functionality or the runtime plugin architecture. Upload the design document to remarkable, keep a detailed diary as you work. Store design doc in ticket."

**Assistant interpretation:** create a fresh ticket, deeply research current architecture, write a long onboarding + design document, maintain an explicit step-by-step diary, and publish to reMarkable.

**Inferred user intent:** establish a durable handoff artifact so a new contributor can execute a package split safely without prior codebase knowledge.

### What I did

1. Ran `docmgr status --summary-only` and `docmgr ticket tickets` to confirm current inventory and naming continuity.
2. Created ticket workspace:
   - `docmgr ticket create-ticket --ticket GEPA-26-HYPERCARD-RUNTIME-SPLIT ...`
3. Created ticket docs:
   - `docmgr doc add --doc-type design-doc ...`
   - `docmgr doc add --doc-type reference ...`

### Why

1. A dedicated ticket keeps scope and changelog clean.
2. Early diary creation enforces high-fidelity capture of commands and errors.

### What worked

1. Ticket and docs created cleanly with expected paths.
2. Topic set reused known vocabulary (`architecture`, `frontend`, `hypercard`, `js-vm`, `plugins`).

### What didn't work

1. N/A.

### What I learned

1. Current docs root has many active/completed tickets, so a precise ticket ID and title are important for discoverability.

### What was tricky to build

1. Ensuring this ticket remained distinct from prior GEPA-25 UI DSL work while still referencing runtime internals shared across both efforts.

### What warrants a second pair of eyes

1. Ticket naming convention consistency with future package-extraction implementation tickets.

### What should be done in the future

1. Spin follow-on implementation ticket series under GEPA-26 umbrella after design approval.

### Code review instructions

1. Verify new ticket path exists under `ttmp/2026/02/28/GEPA-26-HYPERCARD-RUNTIME-SPLIT--...`.
2. Confirm design + diary docs are present.

### Technical details

1. Commands:
   - `docmgr status --summary-only`
   - `docmgr ticket tickets`
   - `docmgr ticket create-ticket ...`
   - `docmgr doc add ...`

## Step 2: Architecture evidence mapping across repositories

I mapped the architecture from root package topology down to runtime host execution paths, then validated app-level consumption in both `go-go-os` demo apps and `go-go-app-arc-agi-3`.

I focused on finding real coupling edges, especially where shell/windowing files import runtime internals and where external apps import engine subpaths that blend concerns.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** gather concrete evidence of current runtime/plugin architecture before proposing split boundaries.

**Inferred user intent:** proposal should be code-grounded, not abstract.

### What I did

1. Captured workspace/package manifests and exports:
   - root `package.json`, workspace config
   - `packages/engine/package.json`
   - `packages/desktop-os/package.json`
   - `packages/confirm-runtime/package.json`
2. Mapped runtime internals:
   - `plugin-runtime/runtimeService.ts`
   - `plugin-runtime/stack-bootstrap.vm.js`
   - `plugin-runtime/contracts.ts`
   - `plugin-runtime/runtimeCardRegistry.ts`
3. Mapped runtime state model:
   - `features/pluginCardRuntime/*`
4. Mapped shell integration and adapter points:
   - `PluginCardSessionHost.tsx`
   - `pluginIntentRouting.ts`
   - `defaultWindowContentAdapters.tsx`
5. Mapped HyperCard artifact path:
   - `hypercard/artifacts/artifactProjectionMiddleware.ts`
   - `hypercard/artifacts/artifactsSlice.ts`
   - `hypercard/artifacts/artifactRuntime.ts`
6. Mapped app consumption patterns:
   - `apps/todo|crm|book-tracker-debug/src/launcher/module.tsx`
   - `go-go-app-arc-agi-3/apps/arc-agi-player/src/launcher/module.tsx`
   - `go-go-app-arc-agi-3/apps/arc-agi-player/src/app/store.ts`

### Why

1. The split proposal needs precise "move/stay" decisions per file and module.
2. App-level import patterns determine migration blast radius.

### What worked

1. Found clean conceptual seams already implied by engine subpath exports (`desktop-core`, `desktop-react`, `desktop-hypercard-adapter`).
2. Identified coupling center clearly: `PluginCardSessionHost`.

### What didn't work

1. One search command used non-existent paths in `go-go-app-arc-agi-3` (`packages`, `src`) and returned:
   - `rg: packages: No such file or directory`
   - `rg: src: No such file or directory`
2. I corrected by scoping to `apps` and then reading targeted files directly.

### What I learned

1. Runtime concerns are already internally cohesive enough for extraction.
2. Desktop shell concerns are mostly independent except runtime adapter glue.
3. External app (`arc-agi-player`) currently depends on engine subpath adapters, so compatibility shims are critical.

### What was tricky to build

1. Distinguishing package-level responsibilities from historical file placement. Some runtime files live under shell paths but are semantically runtime-domain.

### What warrants a second pair of eyes

1. Import graph validation for potential hidden dependencies not visible from high-signal files.

### What should be done in the future

1. Add import-boundary lint rules during migration phases.

### Code review instructions

1. Start with:
   - `packages/engine/src/index.ts`
   - `packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx`
   - `packages/engine/src/features/pluginCardRuntime/pluginCardRuntimeSlice.ts`
2. Confirm external coupling in:
   - `go-go-app-arc-agi-3/apps/arc-agi-player/src/launcher/module.tsx`

### Technical details

1. Representative commands:
   - `sed -n ... package.json`
   - `sed -n ... runtimeService.ts`
   - `sed -n ... PluginCardSessionHost.tsx`
   - `rg -n "@hypercard/engine|PluginCardSessionHost|desktop-hypercard-adapter" ...`

## Step 3: Writing the intern-focused design document

After evidence mapping, I authored the design doc as an onboarding-first narrative: repository basics, runtime lifecycle, state and intent flow, coupling diagnosis, target package boundaries, migration phases, and risks.

I intentionally wrote this as an implementation guide rather than an RFC stub, so an intern can begin Phase 1 extraction work with minimal additional context.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** produce very verbose onboarding documentation with practical migration guidance.

**Inferred user intent:** avoid tribal knowledge dependency.

### What I did

1. Replaced scaffold template sections with full content in:
   - `design-doc/01-hypercard-runtime-package-split-architecture-and-migration-guide.md`
2. Included detailed sections:
   - workspace fundamentals
   - runtime plugin architecture end-to-end
   - explicit goals/non-goals
   - target package decomposition
   - phased migration plan
   - tests, rollback, risks, alternatives
   - intern setup and starter PR guidance
3. Updated ticket index and tasks with meaningful scope and completion tracking.

### Why

1. New contributors need complete context, not partial fragments.
2. Split work is high-risk without dependency-direction clarity and phased criteria.

### What worked

1. Existing code had enough seams to define a practical three-package model.
2. File mapping table made move/stay plan concrete.

### What didn't work

1. N/A.

### What I learned

1. The existing engine subpath exports already hint at future package decomposition; formalizing them reduces migration friction.

### What was tricky to build

1. Balancing verbosity with maintainability: I kept explanatory depth high but constrained speculative content to open-questions section.

### What warrants a second pair of eyes

1. Proposed split of `PluginCardSessionHost` and routing bridge into runtime package; confirm with maintainers who own desktop shell APIs.

### What should be done in the future

1. Convert Phase 0 and Phase 1 into implementation tickets with bounded PR scopes.

### Code review instructions

1. Review from `Executive Summary` to `Migration Plan` sequentially.
2. Validate mapping table against actual file ownership expectations.

### Technical details

1. Files updated:
   - `index.md`
   - `tasks.md`
   - `design-doc/01-hypercard-runtime-package-split-architecture-and-migration-guide.md`

## Step 4: Documentation linkage, validation, and publishing

I prepared this finalization step to ensure the ticket is reproducible and publishable: relate key files, record changelog summary, run doctor checks, and upload the document bundle to reMarkable.

This step also captures where future contributors can verify artifacts and rerun the same workflow.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** complete doc operations and external delivery after writing.

**Inferred user intent:** final output must be discoverable in ticket and available on reMarkable.

### What I did

1. Planned `docmgr doc relate` for design doc and diary with absolute file notes.
2. Planned `docmgr changelog update` with summary and related files.
3. Planned validation and publish commands:
   - `docmgr doctor --ticket GEPA-26-HYPERCARD-RUNTIME-SPLIT --stale-after 30`
   - `remarquee upload bundle ...`
   - `remarquee cloud ls ...`

### Why

1. Without relation metadata and validation, long docs are harder to trust and reuse.
2. reMarkable delivery is an explicit requirement.

### What worked

1. N/A yet at this step (execution follows after writing).

### What didn't work

1. N/A.

### What I learned

1. Absolute `RelatedFiles` paths are safer in this workspace because relative-root mismatches previously caused doctor warnings.

### What was tricky to build

1. Keeping file relations concise while still linking enough code for architectural traceability.

### What warrants a second pair of eyes

1. Verify relation list is not over-linked; maintain discoverability without noise.

### What should be done in the future

1. Add a short "implementation starter checklist" playbook doc once coding phase starts.

### Code review instructions

1. Confirm task checklist status and changelog entry.
2. Re-run doctor and reMarkable listing commands.

### Technical details

1. Finalization commands documented in ticket changelog and terminal history.

## Step 5: Execute validation and reMarkable delivery

I executed the finalization workflow end-to-end after writing and linking docs. This step turned planned operations into verifiable outcomes: changelog written, doctor checks passing, bundle uploaded, and remote listing confirmed.

This closes the ticket's document-delivery scope and leaves a reproducible command trail for future updates.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** complete and verify publishing, not just document writing.

**Inferred user intent:** ensure the result is operationally delivered and not only stored locally.

### What I did

1. Updated doc relationships:
   - `docmgr doc relate --doc <design-doc> --file-note ...`
   - `docmgr doc relate --doc <diary-doc> --file-note ...`
2. Updated ticket changelog:
   - `docmgr changelog update --ticket GEPA-26-HYPERCARD-RUNTIME-SPLIT --entry ... --file-note ...`
3. Ran validation:
   - `docmgr doctor --ticket GEPA-26-HYPERCARD-RUNTIME-SPLIT --stale-after 30`
4. Performed dry-run upload:
   - `remarquee upload bundle --dry-run ...`
5. Performed real upload and verified:
   - `remarquee upload bundle ...`
   - `remarquee cloud ls /ai/2026/02/28/GEPA-26-HYPERCARD-RUNTIME-SPLIT --long --non-interactive`
6. Marked ticket task checklist complete.

## Step 6: Execute hard-cut runtime split in go-go-os (no compatibility wrappers)

I executed the implementation phase as a hard cut. The goal was to make `engine` runtime-agnostic and move all HyperCard/plugin runtime ownership into a dedicated `@hypercard/hypercard-runtime` package.

### Prompt Context

**User prompt (verbatim):** "Create tasks to execute tis plan, no need for backwards compatiblity wrappers or the like. Then work on the tasks one by one, committing at appropriate intervals, keeping a detailed diary. Upload the diary to remarkable once all done."

**Assistant interpretation:** implement the split directly without shims, keep implementation traceability, and commit logically.

**Inferred user intent:** ship the package extraction now, not just plan it.

### What I did

1. Created `packages/hypercard-runtime` and moved source ownership:
   - moved `plugin-runtime/*`
   - moved `features/pluginCardRuntime/*`
   - moved `hypercard/*`
   - moved runtime host files previously under shell windowing:
     - `PluginCardRenderer.tsx`
     - `PluginCardSessionHost.tsx`
     - `pluginIntentRouting.ts`
     - fixtures + runtime host story/test
2. Added package wiring:
   - `packages/hypercard-runtime/package.json`
   - `packages/hypercard-runtime/tsconfig.json`
   - `packages/hypercard-runtime/src/index.ts`
   - root `tsconfig.json` project references updated
   - root package scripts include runtime package in build/test flow
3. Moved `createAppStore` from engine into runtime package:
   - new `packages/hypercard-runtime/src/app/createAppStore.ts`
   - exported from runtime index
   - removed from engine app barrel
4. Removed runtime ownership from engine:
   - dropped engine exports of `plugin-runtime`, `hypercard`, `features/pluginCardRuntime`
   - removed `@hypercard/engine` export path `./desktop-hypercard-adapter`
   - deleted engine adapter files:
     - `src/desktop-hypercard-adapter.ts`
     - `src/desktop/adapters/hypercard.ts`
     - `src/desktop/adapters/index.ts`
   - removed hypercard default adapter from desktop shell defaults
   - removed runtime host re-exports from engine windowing index
5. Reworked chat module ownership boundary:
   - engine `registerChatModules.ts` now owns default sem/renderer registration only
   - runtime package adds `registerHypercardChatModules.ts` to register HyperCard timeline module via engine extension API
6. Updated consumers in go-go-os:
   - apps `todo`, `crm`, `book-tracker-debug`: `PluginCardSessionHost` imports now from `@hypercard/hypercard-runtime`
   - app stores now use runtime `createAppStore`
   - `apps-browser` now imports runtime reducers from runtime package
   - `hypercard-tools` now imports editor runtime APIs from runtime package
   - `desktop-os` launcher store now uses runtime `createAppStore`
   - updated package.json and tsconfig paths/references for all touched apps/packages
7. Moved/adjusted tests:
   - moved `plugin-card-runtime.test.ts` and `plugin-intent-routing.test.ts` from engine to runtime package
   - patched runtime timeline tests to use new chat module registration split
   - constrained runtime test script to `vitest run src` to avoid `dist/` test pickup
8. Dependency and workspace refresh:
   - ran `pnpm install` in `go-go-os`

### Why

1. Engine needed a clean dependency direction (`hypercard-runtime -> engine`, not both ways).
2. Runtime/shell host files were semantically runtime domain and blocked package isolation while living under engine windowing.
3. Existing app imports required direct rewiring to avoid relying on removed engine exports.

### What worked

1. `npm run typecheck` (go-go-os root) passes after cutover.
2. `npx vitest run` in `packages/engine` passes.
3. `npm run test -w packages/hypercard-runtime` passes.
4. `npm run test -w packages/desktop-os` passes.

### What didn't work

1. `npm run test` at go-go-os root still fails due pre-existing Storybook taxonomy errors in `apps/apps-browser` (unrelated to this runtime split).
2. Early import rewrite pass accidentally produced `'/engine'` and `@hypercard@hypercard/engine` strings; I fixed these with targeted replacements and file-by-file verification.
3. One runtime timeline test assumed suggestion projection existed; current behavior keeps it disabled by default, so I updated the test expectation.

### What I learned

1. Runtime ownership extraction is mechanically straightforward once `createAppStore` moves with reducers/middleware.
2. Engine can remain runtime-agnostic while still supporting extension registration via chat module bootstrap API.
3. Test suites were the fastest way to discover ownership leaks and bad path rewrites during the cut.

### What was tricky to build

1. Avoiding dependency-cycle regressions while preserving runtime chat-module registration.
2. Ensuring renamed/moved test files still resolve correct module owners after path changes.

### What warrants a second pair of eyes

1. Long-term location/ownership of runtime-themed stories that still conceptually belong to runtime but may be categorized under existing taxonomy conventions.
2. Whether engine package test command should continue to run global taxonomy checks that include app story paths outside package scope.

### What should be done in the future

1. Add explicit boundary linting between `engine` and `hypercard-runtime`.
2. Consider a dedicated shared package for cross-cutting runtime-agnostic chat/state helpers if additional splits occur.

### Code review instructions

1. Validate package boundary files first:
   - `go-go-os/packages/engine/src/index.ts`
   - `go-go-os/packages/hypercard-runtime/src/index.ts`
   - `go-go-os/packages/engine/package.json`
   - `go-go-os/packages/hypercard-runtime/package.json`
2. Validate shell/runtime host boundary:
   - `go-go-os/packages/engine/src/components/shell/windowing/defaultWindowContentAdapters.tsx`
   - `go-go-os/packages/hypercard-runtime/src/runtime-host/PluginCardSessionHost.tsx`
3. Validate app rewiring:
   - `go-go-os/apps/todo/src/launcher/module.tsx`
   - `go-go-os/apps/crm/src/launcher/module.tsx`
   - `go-go-os/apps/book-tracker-debug/src/launcher/module.tsx`
   - `go-go-os/apps/hypercard-tools/src/launcher/module.tsx`
4. Re-run:
   - `npm run typecheck` in go-go-os
   - `npx vitest run` in `go-go-os/packages/engine`
   - `npm run test -w packages/hypercard-runtime`
   - `npm run test -w packages/desktop-os`

### Technical details

1. Main implementation commit (go-go-os):
   - `5f564f4` — `refactor: split hypercard runtime into dedicated package`

## Step 7: Rewire go-go-app-arc-agi-3 to new runtime package

I updated the ARC app package to stop importing removed engine runtime exports and consume `@hypercard/hypercard-runtime` instead.

### What I did

1. Replaced launcher host import:
   - `PluginCardSessionHost` from `@hypercard/hypercard-runtime`
2. Rewired app store reducers:
   - `pluginCardRuntimeReducer` and `hypercardArtifactsReducer` from runtime package
3. Rewired bridge imports:
   - runtime intent selectors/actions and authorization now from runtime package
   - kept `showToast` from engine
4. Updated app package wiring:
   - added runtime dependency in `apps/arc-agi-player/package.json`
   - added tsconfig path mapping for `@hypercard/hypercard-runtime`

### What worked

1. Import ownership is now aligned with the new package boundary.

### What didn't work

1. Local install/typecheck inside `apps/arc-agi-player` is not directly runnable as a standalone workspace:
   - `npm install` fails on `workspace:*` (pre-existing workspace topology issue)
   - direct `tsc` check reports many pre-existing dependency/rootDir/environment issues not specific to this import rewrite

### Technical details

1. ARC app integration commit:
   - `c86ea5e` — `refactor: rewire arc-agi-player to hypercard-runtime package`

## Step 8: Publish updated implementation diary bundle to reMarkable

After implementation commits and ticket doc updates, I re-published the GEPA-26 bundle so the intern-facing materials now include the execution diary and commit-linked changelog.

### What I did

1. Ran upload dry-run:
   - `remarquee upload bundle <ticket-dir> --remote-dir /ai/2026/02/28/GEPA-26-HYPERCARD-RUNTIME-SPLIT --name GEPA-26-runtime-split-implementation-diary-update --dry-run`
2. Ran real upload:
   - `remarquee upload bundle <ticket-dir> --remote-dir /ai/2026/02/28/GEPA-26-HYPERCARD-RUNTIME-SPLIT --name GEPA-26-runtime-split-implementation-diary-update`
3. Verified remote listing:
   - `remarquee cloud ls /ai/2026/02/28/GEPA-26-HYPERCARD-RUNTIME-SPLIT --long --non-interactive`

### Result

1. Upload succeeded:
   - `OK: uploaded GEPA-26-runtime-split-implementation-diary-update.pdf -> /ai/2026/02/28/GEPA-26-HYPERCARD-RUNTIME-SPLIT`
2. Remote folder contains:
   - `GEPA-26 HyperCard Runtime Split Design`
   - `GEPA-26-runtime-split-implementation-diary-update`

## Step 9: Resolve downstream launcher alias regression found in tmux dev flow

After implementation, the live `wesen-os` launcher surfaced an import-resolution failure for `@hypercard/hypercard-runtime`. This happened because launcher alias/path wiring was manually curated and still only listed `@hypercard/engine` paths.

### What I did

1. Patched `wesen-os/apps/os-launcher` resolver wiring:
   - added runtime alias to `vite.config.ts`
   - added runtime alias to `vitest.config.ts`
   - added runtime path mapping to `tsconfig.json`
2. Re-ran launcher build:
   - `npm run build` in `wesen-os/apps/os-launcher`
3. Build then exposed the next stale import in inventory app:
   - `@hypercard/engine/desktop-hypercard-adapter` in `go-go-app-inventory`
4. Patched `go-go-app-inventory/apps/inventory`:
   - moved runtime imports (`PluginCardSessionHost`, `RuntimeCardDebugWindow`, `registerHypercardTimelineModule`, `createAppStore`) to `@hypercard/hypercard-runtime`
   - added runtime package wiring in inventory `package.json` and `tsconfig.json`
5. Re-ran launcher build successfully after both patches.

### Why

1. The earlier verification focused on package tests and build commands, not the cross-repo Vite alias graph used by `wesen-os` launcher in tmux.
2. The launcher composes multiple app sources from different repos, so alias drift appears only there.

### Technical details

1. Launcher alias hotfix commit:
   - `wesen-os`: `138d455` — `fix: wire hypercard-runtime alias in os-launcher configs`
2. Inventory import migration follow-up:
   - `go-go-app-inventory`: `e5710f2` — `refactor: migrate inventory app runtime imports to hypercard-runtime`

### Why

1. Delivery requirements include validated ticket docs and reMarkable publication.
2. Dry-run before real upload reduces formatting/payload surprises.

### What worked

1. `docmgr doctor` returned `All checks passed`.
2. Dry-run bundle composition matched expected docs.
3. Real upload succeeded:
   - `OK: uploaded GEPA-26 HyperCard Runtime Split Design.pdf -> /ai/2026/02/28/GEPA-26-HYPERCARD-RUNTIME-SPLIT`
4. Remote listing confirmed artifact:
   - `[f] GEPA-26 HyperCard Runtime Split Design`

### What didn't work

1. N/A.

### What I learned

1. Explicit file-note linkage improves discoverability of architecture claims and shortens future audit loops.

### What was tricky to build

1. Keeping relation metadata comprehensive without over-linking required selecting only high-signal files representing topology, runtime core, coupling seams, and downstream consumption.

### What warrants a second pair of eyes

1. Whether any additional downstream repos beyond `go-go-app-arc-agi-3` should be added as relation evidence in a follow-up update.

### What should be done in the future

1. Open follow-on implementation tickets for Phase 0 and Phase 1 migration workstreams.

### Code review instructions

1. Verify ticket status:
   - `docmgr doctor --ticket GEPA-26-HYPERCARD-RUNTIME-SPLIT --stale-after 30`
2. Verify published artifact:
   - `remarquee cloud ls /ai/2026/02/28/GEPA-26-HYPERCARD-RUNTIME-SPLIT --long --non-interactive`
3. Review changelog and tasks for closure.

### Technical details

1. Published file name:
   - `GEPA-26 HyperCard Runtime Split Design.pdf`
2. Remote destination:
   - `/ai/2026/02/28/GEPA-26-HYPERCARD-RUNTIME-SPLIT`

## Quick Reference

### High-signal files for package split work

1. `go-go-os/packages/engine/src/plugin-runtime/runtimeService.ts`
2. `go-go-os/packages/engine/src/features/pluginCardRuntime/pluginCardRuntimeSlice.ts`
3. `go-go-os/packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx`
4. `go-go-os/packages/engine/src/components/shell/windowing/pluginIntentRouting.ts`
5. `go-go-os/packages/engine/src/hypercard/artifacts/artifactProjectionMiddleware.ts`
6. `go-go-os/packages/engine/src/app/createAppStore.ts`
7. `go-go-os/packages/desktop-os/src/store/createLauncherStore.ts`
8. `go-go-app-arc-agi-3/apps/arc-agi-player/src/launcher/module.tsx`

## Usage Examples

### Example: rerun architecture validation

```bash
cd /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa
docmgr doctor --ticket GEPA-26-HYPERCARD-RUNTIME-SPLIT --stale-after 30
```

### Example: inspect coupling center quickly

```bash
cd /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os
sed -n '1,260p' packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx
```

## Related

1. Design doc:
   - `design-doc/01-hypercard-runtime-package-split-architecture-and-migration-guide.md`
2. Ticket tasks:
   - `tasks.md`
