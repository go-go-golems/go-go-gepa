# Tasks

## Phase 0: Ticket and research setup

- [x] Create ticket workspace `GEPA-28-UI-DSL-DEMO-CARDS`
- [x] Create primary design doc and implementation diary docs
- [x] Close `GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT`

## Phase 1: Architecture investigation for intern handoff

- [x] Map HyperCard Tools launcher flow and current launch behavior
- [x] Map runtime UI DSL contract and supported widget kinds
- [x] Map renderer and runtime intent dispatch path
- [x] Write detailed intern onboarding and implementation guide
- [x] Upload research/guide bundle to reMarkable

## Phase 2: HyperCard Tools demo stack implementation

- [x] Add `apps/hypercard-tools/src/domain` stack scaffold (`stack.ts`, plugin bundle, authoring types)
- [x] Implement home/folder card with navigable demo catalog
- [x] Implement widget showcase cards for all active UI DSL widgets
- [x] Implement interactive state/demo handlers (card/session/system intents)
- [x] Wire launcher to open demo stack by default when clicking HyperCard Tools icon
- [x] Preserve runtime-card editor window behavior for encoded editor app keys

## Phase 3: Validation and tests

- [x] Update tests that assume HyperCard Tools launch content kind is `app`
- [x] Add/adjust tests covering new HyperCard Tools launch and render behavior
- [x] Run targeted test suite(s) in `wesen-os` and `go-go-os`
- [x] Run typecheck/build validation for touched packages/apps

## Phase 4: Bookkeeping, diary, and commits

- [x] Update diary with chronological implementation steps and command evidence
- [x] Update ticket changelog with implementation milestones
- [x] Commit go-go-os implementation changes
- [x] Commit go-go-gepa documentation/task updates
- [x] Upload final docs bundle (guide + diary) to reMarkable

## Phase 5: DSL and demo coverage expansion (follow-up)

- [x] Extend UI DSL contract with `ui.dropdown`
- [x] Extend UI DSL contract with `ui.selectableTable`
- [x] Extend UI DSL contract with `ui.gridBoard`
- [x] Wire schema validation and runtime renderer support for new node kinds
- [x] Add HyperCard Tools demo cards for dropdown, selectable table, and grid board
- [x] Add missing examples: button variants, input placeholder, event-args merge payload
- [x] Add domain-intent example card and backing app domain reducer wiring
- [x] Run targeted typecheck/tests for `go-go-os` and `wesen-os`
- [x] Update GEPA-28 diary/changelog and check off follow-up tasks
