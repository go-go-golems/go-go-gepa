# Tasks

## Completed in this ticket

- [x] Create GEPA-27 ticket workspace and base documents
- [x] Gather line-referenced architecture evidence across engine/runtime/chat/story ownership
- [x] Write detailed 8+ page design and implementation proposal for intern onboarding
- [x] Write chronological investigation diary
- [x] Update ticket index/changelog and prepare delivery bundle

## Implementation Plan (Detailed)

### Phase 1: Visibility Resolver API in Engine

- [x] Add `DesktopVisibilityContextResolver` contract to engine windowing types
- [x] Add optional `visibilityContextResolver` prop to `DesktopShellProps`
- [x] Replace chat-specific resolver logic in `useDesktopShellController` with injected/default resolver
- [x] Remove engine-internal `chatProfiles` state introspection helpers from shell controller
- [x] Add/adjust tests for default + custom resolver behavior
- [x] Verify engine builds and tests pass

### Phase 2: Create `@hypercard/chat-runtime` Package

- [x] Scaffold `packages/chat-runtime` (`package.json`, `tsconfig.json`, `src/index.ts`)
- [x] Move `packages/engine/src/chat/*` into `packages/chat-runtime/src/*`
- [x] Update imports inside moved files to use correct package-relative paths
- [x] Move chat tests with source and keep them passing in new package
- [x] Update workspace `go-go-os/tsconfig.json` references and root build/test scripts
- [x] Verify `chat-runtime` package typecheck/test/build

### Phase 3: Rewire Package Imports to New Chat Package

- [x] Update `desktop-os` and app imports to use `@hypercard/chat-runtime`
- [x] Update `hypercard-runtime` references that currently import chat slices/types from `@hypercard/engine`
- [x] Ensure store wiring in apps that need chat reducers still composes correctly
- [x] Verify affected apps compile (`todo`, `crm`, `book-tracker-debug`, `apps-browser`, `hypercard-tools` as needed)

### Phase 4: Relocate Runtime-Coupled Stories

- [x] Remove/move `engine` runtime-coupled stories:
  - [x] `BookTracker.stories.tsx`
  - [x] `RuntimeCardDebugWindow.stories.tsx`
  - [x] `CodeEditorWindow.stories.tsx`
  - [x] `ChatConversationWindow.stories.tsx`
  - [x] Runtime/chat-heavy story variants in `DesktopShell.stories.tsx`
- [x] Add equivalent stories in owning packages/apps:
  - [x] `apps/book-tracker-debug`
  - [x] `packages/hypercard-runtime`
  - [x] `packages/chat-runtime`
- [x] Keep only pure engine/windowing stories in engine taxonomy
- [x] Fix story imports and taxonomy checks

### Phase 5: Final Engine Boundary Cleanup

- [x] Remove `export * from './chat'` from `packages/engine/src/index.ts`
- [x] Ensure no engine source imports from `@hypercard/hypercard-runtime`
- [x] Ensure no engine windowing code depends on chat state shape
- [x] Verify engine package dependency list reflects boundary intent

### Phase 6: DAG Guardrails and CI Checks

- [x] Add forbidden-import checks (script or test) for package boundaries
- [x] Guard examples:
  - [x] engine must not import `@hypercard/hypercard-runtime`
  - [x] engine must not import `@hypercard/chat-runtime`
  - [x] engine shell must not hardcode `chatProfiles` schema
- [x] Wire guard checks into package/root test scripts

### Phase 7: Full Validation and Documentation

- [x] Run full build across packages/apps
- [x] Run full tests and storybook taxonomy checks
- [x] Check off all completed tasks in this file
- [x] Update GEPA-27 diary with per-phase notes + command evidence
- [x] Update GEPA-27 changelog with implementation milestones
- [x] Upload updated GEPA-27 diary to reMarkable
- [x] Upload GEPA-26 diary to reMarkable
