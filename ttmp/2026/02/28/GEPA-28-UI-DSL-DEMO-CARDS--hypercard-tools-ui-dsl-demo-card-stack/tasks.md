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
