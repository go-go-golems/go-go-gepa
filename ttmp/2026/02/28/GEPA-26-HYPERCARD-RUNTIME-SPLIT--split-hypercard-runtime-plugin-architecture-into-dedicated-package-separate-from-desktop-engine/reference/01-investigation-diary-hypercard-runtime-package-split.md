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
LastUpdated: 2026-02-28T15:50:00Z
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
