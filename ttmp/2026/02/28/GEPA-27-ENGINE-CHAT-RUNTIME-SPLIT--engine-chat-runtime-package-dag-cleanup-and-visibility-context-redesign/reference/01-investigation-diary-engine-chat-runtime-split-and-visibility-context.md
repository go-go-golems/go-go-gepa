---
Title: 'Investigation diary: engine-chat-runtime split and visibility context'
Ticket: GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT
Status: active
Topics:
    - architecture
    - frontend
    - go-go-os
    - hypercard
    - plugins
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT--engine-chat-runtime-package-dag-cleanup-and-visibility-context-redesign/design-doc/01-design-and-implementation-plan-split-chat-runtime-out-of-engine-and-introduce-generic-visibility-context.md
      Note: Diary references the final architecture plan
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/__tests__/storybook-app-smoke.test.ts
      Note: Diary records evidence for app-level story ownership pattern
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/hypercard-runtime/src/hypercard/debug/RuntimeCardDebugWindow.stories.tsx
      Note: Diary records story ownership move out of engine
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/scripts/deps/check-boundaries.mjs
      Note: Diary records package-boundary guardrail implementation
ExternalSources: []
Summary: Chronological diary of creating the GEPA-27 design package split proposal and delivery to reMarkable.
LastUpdated: 2026-02-28T17:02:35-05:00
WhatFor: Help reviewers and future implementers understand what evidence was gathered and why each design decision was made.
WhenToUse: Use when implementing GEPA-27 phases or auditing reasoning behind package-boundary decisions.
---


# Investigation diary: engine-chat-runtime split and visibility context

## Goal

Capture exactly how this ticket was created, researched, documented, and delivered, including the commands and evidence used to derive the design.

## Step 1: Create GEPA-27 ticket scaffold

I started by creating a fresh ticket workspace because the request asked for a new ticket rather than extending an existing active ticket. The ticket title was written to reflect both parts of the effort: dependency-graph cleanup and visibility-context redesign.

I then created two docs immediately: one primary design-doc and one diary/reference doc. This ensured documentation structure was in place before analysis started.

### Prompt Context

**User prompt (verbatim):** "Make a new ticket to move chat runtime out of engine/ so that we have clean packages and clear DAG dependency graph. 

Create a  detailed design and implementation document for our intern who doesn't know the codebase: for: 

- move stories to their respective component (BookTracker prob apps/Booktracker? 
- use vibility context for registering chat stuff, so that we sever all dependencies from engine/ package. Design a visibility context so that it is easy for external modules to control visibility. 
- Move all the chat stuff into its own package as well

For the visibility context thing, explain what it is for, what visibility even means, how and what for it can be used by other packages/ and explain the design rationales and trade offs. I'm not even sure we need that thing, because I don't know what it is for. 

8+ pages document, upload to remarkable, verbose and detailed and written out for our intern. prose paragraphs, bullet points, pseudocode, code snippets , diagrams, etc..."

**Assistant interpretation:** Create a new docmgr ticket, produce a long intern-focused architecture/design proposal with evidence and migration steps, and upload the result to reMarkable.

**Inferred user intent:** Establish a clean long-term architecture boundary and a practical implementation plan that other engineers can execute without ambiguity.

**Commit (code):** N/A (docs-only work in ticket workspace)

### What I did

- Ran `docmgr status --summary-only` to confirm docs root and health.
- Ran `docmgr ticket create-ticket --ticket GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT ...`.
- Ran `docmgr doc add` for design-doc and diary reference docs.

### Why

- New ticket was explicitly requested.
- Separate design and diary docs follow the existing ticket workflow conventions.

### What worked

- Ticket scaffold generated successfully in `go-go-gepa/ttmp` with expected files (`index.md`, `tasks.md`, `changelog.md`).

### What didn't work

- N/A in this step.

### What I learned

- Ticket metadata inherited expected topics and status cleanly from CLI defaults.

### What was tricky to build

- The only subtle point was selecting a ticket id that matched established naming conventions while remaining specific enough for later search and cross-reference.

### What warrants a second pair of eyes

- Ticket naming consistency with internal reporting dashboards if any external automation expects a stricter pattern.

### What should be done in the future

- Add an explicit naming convention note to team docs for new architecture tickets.

### Code review instructions

- Verify ticket path exists and contains expected scaffold docs.
- Run `docmgr ticket list --ticket GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT`.

### Technical details

- Command: `docmgr ticket create-ticket --ticket GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT --title "Engine/Chat/Runtime package DAG cleanup and visibility-context redesign" --topics architecture,frontend,go-go-os,hypercard,plugins`

## Step 2: Gather architecture and coupling evidence

After scaffolding, I gathered line-anchored evidence from package manifests, barrel exports, shell controller logic, runtime store wiring, and story files. I focused on facts needed to justify every major recommendation.

I gathered both package-level and behavior-level evidence: package dependencies, import graphs, and where runtime behavior is coupled to chat-state shape.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Build an evidence-backed model of the current architecture before proposing changes.

**Inferred user intent:** Avoid speculative redesign and ensure refactor steps are credible and reviewable.

**Commit (code):** N/A (analysis-only)

### What I did

- Collected manifests:
  - `go-go-os/package.json`
  - `packages/engine/package.json`
  - `packages/hypercard-runtime/package.json`
  - `packages/desktop-os/package.json`
- Collected critical implementation files:
  - `engine/src/index.ts`
  - `engine/src/chat/index.ts`
  - `engine/src/components/shell/windowing/useDesktopShellController.tsx`
  - `engine/src/components/shell/windowing/contextActionVisibility.ts`
  - `hypercard-runtime/src/app/createAppStore.ts`
  - `desktop-os/src/store/createLauncherStore.ts`
- Collected story ownership evidence:
  - runtime imports in engine stories
  - existing app-level BookTracker story in `apps/book-tracker-debug`
- Collected downstream consumer evidence from `go-go-app-arc-agi-3` imports of `@hypercard/hypercard-runtime`.

### Why

- The design requested a clear DAG and explanation for interns; this requires concrete source-backed description of current edges and boundary violations.

### What worked

- `rg` + `nl -ba` were sufficient to capture all core evidence quickly.

### What didn't work

- One attempted lookup targeted `windowContentAdapter.tsx` while file is `windowContentAdapter.ts`; corrected by querying file list first.

### What I learned

- The problematic coupling in engine is mostly not a compile-time import from chat runtime package; it is a data-shape coupling (`state.chatProfiles`) inside shell controller logic.

### What was tricky to build

- Distinguishing between “hard” dependency (import edges) and “soft” dependency (schema assumptions against global state) is subtle but essential for correct refactor scope.

### What warrants a second pair of eyes

- Confirm whether any unobserved consumers import chat APIs from `@hypercard/engine` in external repositories outside this workspace.

### What should be done in the future

- Add a static import boundary check in CI for each package.

### Code review instructions

- Start with `engine/src/index.ts` and `engine/src/chat/index.ts`.
- Then inspect `useDesktopShellController.tsx` lines `232-289` and `917-919`.
- Validate story import findings with `rg` query in ticket notes.

### Technical details

Key commands executed:

```bash
rg --line-number --no-heading "from '@hypercard/hypercard-runtime'" go-go-os/packages/engine/src
nl -ba go-go-os/packages/engine/src/components/shell/windowing/useDesktopShellController.tsx | sed -n '220,300p'
nl -ba go-go-os/packages/hypercard-runtime/src/app/createAppStore.ts | sed -n '1,140p'
```

## Step 3: Write detailed design and implementation proposal

I wrote an 8+ page design document that explains fundamentals first, then maps current architecture, then presents target architecture with migration phases, pseudocode, diagrams, and tradeoffs.

The document explicitly explains what "visibility" means in this system, why it exists, and why we should keep the policy mechanism but externalize context derivation.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Produce intern-friendly architecture material that can directly drive implementation.

**Inferred user intent:** Ensure a new engineer can execute the refactor with minimal oral transfer.

**Commit (code):** N/A (docs-only)

### What I did

- Replaced design doc skeleton with full content:
  - executive summary,
  - current-state architecture evidence,
  - visibility-context deep explanation,
  - package DAG diagrams,
  - pseudocode for resolver injection,
  - phased implementation plan,
  - test/rollout strategy and risks.
- Included concrete file references from observed code.

### Why

- The request explicitly asked for verbose onboarding content plus design rationale and tradeoffs.

### What worked

- Evidence-first structure made it straightforward to argue for each recommendation.

### What didn't work

- N/A in writing step.

### What I learned

- The repo already has app-level story architecture (`apps/*/stories`) that naturally supports moving BookTracker out of engine without creating a new pattern.

### What was tricky to build

- Balancing “verbose intern onboarding” with “concise/elegant API design” required explicitly separating conceptual explanation from final API surface recommendation.

### What warrants a second pair of eyes

- Validate naming for the proposed new package (`@hypercard/chat-runtime`) against existing naming policy.

### What should be done in the future

- Convert Phase plans into executable subtasks and sequence PRs by phase.

### Code review instructions

- Review design doc sections in order:
  1. Problem Statement
  2. Current-State Architecture
  3. Visibility explanation and API sketch
  4. Phased implementation plan

### Technical details

- Primary output file:
  - `.../design-doc/01-design-and-implementation-plan-split-chat-runtime-out-of-engine-and-introduce-generic-visibility-context.md`

## Step 4: Update ticket bookkeeping and prepare delivery

I converted the ticket from empty scaffold to actionable work package by updating tasks and index summary and documenting this diary.

I then prepared the ticket for `docmgr doctor` and reMarkable bundle upload.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Complete the full delivery loop, not only document authoring.

**Inferred user intent:** Have a usable ticket artifact, not a loose markdown file.

**Commit (code):** N/A

### What I did

- Updated `tasks.md` with concrete migration phases.
- Updated `index.md` overview and key links.
- Updated this diary with chronological technical context.

### Why

- Ticket should be implementation-ready and continuation-friendly.

### What worked

- Existing ticket scaffold was easy to extend into actionable structure.

### What didn't work

- N/A.

### What I learned

- Keeping both high-level design and chronological diary in the same ticket materially improves continuation quality.

### What was tricky to build

- Ensuring task granularity is neither too broad (non-actionable) nor too narrow (noise).

### What warrants a second pair of eyes

- Confirm task sequencing aligns with expected PR boundaries and team review capacity.

### What should be done in the future

- Add owner assignments once implementation starts.

### Code review instructions

- Check `index.md`, `tasks.md`, `changelog.md`, design doc, and diary in one pass.
- Run `docmgr doctor --ticket GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT --stale-after 30`.

### Technical details

- Files updated in ticket workspace:
  - `index.md`
  - `tasks.md`
  - `changelog.md`
  - design doc
  - diary doc

## What worked (overall)

- New ticket creation and document scaffolding.
- Rapid evidence gathering with line-level references.
- Translation from technical evidence to phased design plan.

## What didn't work (overall)

- Initial assumption that all story imports were already fixed; one stale path remains in `RuntimeCardDebugWindow.stories.tsx` and is explicitly called out in design.

## What was tricky to build (overall)

- Separating immediate breakage fixes from the larger architectural split while still producing one coherent plan.
- Explaining visibility-context purpose in a way that is not chat-specific.

## What warrants a second pair of eyes (overall)

- Proposed resolver contract shape and whether to evolve it now to generic claims vs maintain existing profile/role fields.
- Package boundary enforcement strategy in CI.

## What should be done in the future (overall)

1. Execute Phase 1 resolver extraction.
2. Execute chat package extraction.
3. Relocate runtime stories.
4. Add DAG guard checks to CI.

## Usage Examples

### Reproduce evidence queries

```bash
rg --line-number --no-heading "chatProfiles" go-go-os/packages/engine/src/components/shell/windowing/useDesktopShellController.tsx
rg --line-number --no-heading "@hypercard/hypercard-runtime" go-go-os/packages/engine/src -g'*.stories.tsx'
```

### Validate ticket integrity

```bash
docmgr doctor --ticket GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT --stale-after 30
```

## Related

- Design doc in this same ticket.
- Existing split ticket: `GEPA-26-HYPERCARD-RUNTIME-SPLIT`.

## Step 5: Upload bundle to reMarkable and verify remote path

I completed the required upload workflow with a dry-run first, then a real bundle upload. I included index, tasks, changelog, design doc, and diary in one PDF bundle so the intern receives one coherent packet with table of contents.

There was one verification hiccup: initial directory listing returned `Error: no matches for 'GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT'`. I resolved it by listing the parent directory and then listing the specific folder with a trailing slash.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Publish the final ticket documentation to reMarkable and verify delivery.

**Inferred user intent:** Ensure docs are accessible on-device, not only present in the repo.

**Commit (code):** N/A

### What I did

- Ran `remarquee status` and `remarquee cloud account --non-interactive`.
- Ran `remarquee upload bundle --dry-run ...` with target `/ai/2026/02/28/GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT`.
- Ran real `remarquee upload bundle ...`.
- Verified parent path and final file listing.

### Why

- Skill guardrails require dry-run first and post-upload verification.

### What worked

- Bundle upload succeeded and produced `GEPA-27 Engine Chat Runtime Split Design` in the target folder.

### What didn't work

- First verification command returned:
  - `Error: no matches for 'GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT'`
- Follow-up commands confirmed folder existed and listing works with explicit trailing slash.

### What I learned

- `remarquee cloud ls` path matching can be sensitive to exact path formatting/timing; parent listing is a reliable fallback check.

### What was tricky to build

- Ensuring verification was both strict and resilient to path-listing quirks without skipping proof of upload.

### What warrants a second pair of eyes

- None beyond standard confirmation on-device.

### What should be done in the future

- Add a small helper script in ticket `scripts/` that performs parent-and-child listing automatically after upload.

### Code review instructions

- Re-run the same upload verification commands and compare resulting folder listing.

### Technical details

Commands run:

```bash
remarquee status
remarquee cloud account --non-interactive
remarquee upload bundle --dry-run <docs...> --name "GEPA-27 Engine Chat Runtime Split Design" --remote-dir "/ai/2026/02/28/GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT" --toc-depth 2
remarquee upload bundle <docs...> --name "GEPA-27 Engine Chat Runtime Split Design" --remote-dir "/ai/2026/02/28/GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT" --toc-depth 2
remarquee cloud ls /ai/2026/02/28 --long --non-interactive
remarquee cloud ls /ai/2026/02/28/GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT/ --long --non-interactive
```

## Step 6: Implement Phase 1 visibility resolver API and decouple shell from chatProfiles schema

I started implementation by landing the smallest high-leverage boundary cut first: remove chat profile schema parsing from engine shell controller and replace it with a generic injected resolver. This keeps behavior extensible and allows chat policy logic to live in chat runtime package later.

I introduced new engine-level types for resolver arguments and resolver function, wired a new optional prop into `DesktopShellProps`, and changed the context-menu visibility pipeline to call either the injected resolver or a default resolver that only returns `{ target }`.

### Prompt Context

**User prompt (verbatim):** "Create detailed tasks to do it. Implement it task by task, checking them off as you go, committing at appropriate intervals, keeping a detailed diary. When done, upload the diary to remarkable, as well as the one from the previous refactor ticket."

**Assistant interpretation:** Execute GEPA-27 implementation in phases with task tracking, code commits, and diary updates, then upload diary artifacts.

**Inferred user intent:** Convert the design into real code incrementally with auditable progress.

**Commit (code):** `8955530` — "engine: inject visibility context resolver into desktop shell"

### What I did

- Added new types in engine windowing `types.ts`:
  - `DesktopVisibilityContextResolverArgs`
  - `DesktopVisibilityContextResolver`
- Added optional `visibilityContextResolver` prop in `desktopShellTypes.ts`.
- Refactored `useDesktopShellController.tsx`:
  - removed internal `chatProfiles` parsing helpers,
  - added `resolveDefaultActionVisibilityContext`,
  - routed visibility through `(visibilityContextResolver ?? defaultResolver)`.
- Exposed new types in public exports:
  - `components/shell/windowing/index.ts`
  - `desktop/react/index.ts`
- Added a context-menu test path for externally supplied visibility context in `DesktopShell.contextMenu.test.tsx`.
- Ran package typecheck and targeted test command.

### Why

- This is the prerequisite for severing engine from chat-state assumptions.
- It enables later extraction where chat package provides resolver logic instead of engine reading `state.chatProfiles`.

### What worked

- `npm run typecheck` in `packages/engine` passed.
- Existing `contextActionVisibility.test.ts` passed.
- Commit was clean and isolated to Phase 1 files.

### What didn't work

- Running Vitest directly for `DesktopShell.contextMenu.test.tsx` returned `No test files found` because current Vitest include pattern is `src/**/*.test.ts` and does not include `.test.tsx` files.

### What I learned

- There are existing `.test.tsx` files in engine not covered by current Vitest include configuration; this is a pre-existing test-discovery gap.

### What was tricky to build

- The tricky part was preserving invocation-time safety checks (`isContextCommandAllowed`) while threading optional `invocation` into resolver calls without changing existing call sites.

### What warrants a second pair of eyes

- Whether to standardize test discovery to include `.test.tsx` now, or defer to a dedicated testing hygiene ticket.

### What should be done in the future

- Add/update Vitest include patterns to run `.test.tsx` files consistently.

### Code review instructions

- Review these files in order:
  1. `packages/engine/src/components/shell/windowing/types.ts`
  2. `packages/engine/src/components/shell/windowing/desktopShellTypes.ts`
  3. `packages/engine/src/components/shell/windowing/useDesktopShellController.tsx`
  4. `packages/engine/src/components/shell/windowing/index.ts`
  5. `packages/engine/src/desktop/react/index.ts`
- Validate with:

```bash
cd go-go-os/packages/engine
npm run typecheck
npx vitest run src/components/shell/windowing/contextActionVisibility.test.ts
```

### Technical details

Commands run:

```bash
cd go-go-os/packages/engine
npm run typecheck
npx vitest run src/components/shell/windowing/contextActionVisibility.test.ts
npx vitest run "src/components/shell/windowing/DesktopShell.contextMenu.test.tsx"  # no test files found due include pattern
```

## Step 5: Phase 1 implementation landed in engine (visibility resolver injection)

I implemented the resolver injection seam in the engine shell controller and committed it as the first execution-phase checkpoint.

### What changed

- Added `DesktopVisibilityContextResolver` API types in windowing contracts.
- Added optional `visibilityContextResolver` to `DesktopShellProps`.
- Replaced hardcoded `chatProfiles`-state derivation in `useDesktopShellController` with:
  - injected resolver when provided,
  - default resolver that only passes target.
- Added test coverage for custom resolver path.

### Commit

- `go-go-os`: `8955530`

### Validation

- `pnpm -C go-go-os -r --filter @hypercard/engine run typecheck`
- `vitest` context-action visibility tests passed.

## Step 6: Phase 2+3 package extraction and import rewires

I moved chat runtime source out of engine and rewired package/app consumers to the new package.

### What changed

- Created new package: `packages/chat-runtime`.
- Moved `packages/engine/src/chat/*` -> `packages/chat-runtime/src/chat/*`.
- Removed chat barrel export from `packages/engine/src/index.ts`.
- Rewired `hypercard-runtime` chat-related imports to `@hypercard/chat-runtime`.
- Rewired app stores/consumers (`crm`, `apps-browser`, `inventory`, `os-launcher`) to consume chat reducers/apis from `@hypercard/chat-runtime`.
- Updated workspace refs/path aliases where needed.

### Commits

- `go-go-os`: `d36a5ee` (`runtime: extract chat runtime package and rewire consumers`)
- `go-go-app-inventory`: `d5e9e46` (`inventory: consume chat-runtime package APIs`)
- `go-go-app-arc-agi-3`: `0dea6a1` (`arc-agi: fix bridge dispatch typing for runtime intent mirror`)
- `wesen-os`: `d4231c4` (`os-launcher: wire chat-runtime aliases and reducers`)

### Validation

- `pnpm -C go-go-os -r run typecheck` passed.
- `pnpm -C wesen-os/apps/os-launcher run typecheck` passed.
- `pnpm -C wesen-os/apps/os-launcher run build` passed.
- `npm run typecheck` in `go-go-app-inventory` passed.

## Step 7: Phase 4 story relocation + Phase 6 guardrails

I moved runtime-coupled stories out of `engine` and introduced explicit dependency-boundary checks.

### What changed

- Moved runtime-coupled stories from engine to owning packages:
  - chat stories -> `packages/chat-runtime/src/chat/...`
  - hypercard runtime stories -> `packages/hypercard-runtime/src/hypercard/...`
- Deleted `packages/engine/src/components/widgets/BookTracker.stories.tsx` (app-level `book-tracker-debug` story already exists).
- Replaced `DesktopShell.stories.tsx` with pure engine/windowing scenarios and generic visibility-context demo.
- Added local engine story fixture helper (`storyFakeResponses.ts`) to avoid importing moved chat mocks.
- Extended Storybook config to include `chat-runtime` and `hypercard-runtime` story directories.
- Updated story taxonomy checker to understand new package prefixes and placement constraints.
- Added `scripts/deps/check-boundaries.mjs` for forbidden imports and shell-state-shape coupling checks.
- Wired boundary checker into root `test` script (`deps:check`).

### Commit

- `go-go-os`: `9ee7efe`

### Validation

- `pnpm -C go-go-os -r run typecheck` passed.
- `pnpm -C go-go-os run storybook:check` passed.
- `pnpm -C go-go-os run build-storybook` passed.

## Step 8: Phase 7 full test run and post-split regression fix

I ran the full root `go-go-os` test pipeline. It surfaced a post-split regression in one chat-runtime test import path.

### Failure observed

- `packages/chat-runtime/src/chat/debug/yamlFormat.test.ts` imported `toYaml` from `@hypercard/engine`.
- After split, this symbol no longer exists in `engine`, causing runtime test failures.

### Fix

- Switched test import to local module:
  - `import { toYaml } from './yamlFormat';`

### Commit

- `go-go-os`: `10c7b6b` (`chat-runtime: fix yamlFormat test import after package split`)

### Final validation evidence

- `pnpm -C go-go-os run test` passed end-to-end:
  - taxonomy check,
  - boundary check,
  - engine/chat-runtime/hypercard-runtime/desktop-os test suites.
- `pnpm -C wesen-os/apps/os-launcher run typecheck` passed.
- `npm run typecheck` in `go-go-app-inventory` passed.

## Step 9: Upload final diaries to reMarkable (GEPA-27 + GEPA-26)

After implementation and validation were complete, I uploaded both requested diary documents.

### Upload commands

```bash
remarquee upload md /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT--engine-chat-runtime-package-dag-cleanup-and-visibility-context-redesign/reference/01-investigation-diary-engine-chat-runtime-split-and-visibility-context.md --remote-dir /ai/2026/02/28/GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT

remarquee upload md /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-26-HYPERCARD-RUNTIME-SPLIT--split-hypercard-runtime-plugin-architecture-into-dedicated-package-separate-from-desktop-engine/reference/01-investigation-diary-hypercard-runtime-package-split.md --remote-dir /ai/2026/02/28/GEPA-26-HYPERCARD-RUNTIME-SPLIT
```

### Verification

```bash
remarquee cloud ls /ai/2026/02/28/GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT --long --non-interactive
remarquee cloud ls /ai/2026/02/28/GEPA-26-HYPERCARD-RUNTIME-SPLIT --long --non-interactive
```

Verified entries include:

- `01-investigation-diary-engine-chat-runtime-split-and-visibility-context`
- `01-investigation-diary-hypercard-runtime-package-split`

