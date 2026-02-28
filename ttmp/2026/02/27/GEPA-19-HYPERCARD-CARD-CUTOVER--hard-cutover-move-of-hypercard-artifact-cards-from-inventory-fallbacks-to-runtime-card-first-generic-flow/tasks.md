# Tasks

## Phase A - Planning and Ticket Setup

- [x] A1 Create GEPA-19 ticket workspace via `docmgr ticket create-ticket`
- [x] A2 Add design-doc with hard-cutover implementation plan
- [x] A3 Add reference diary document
- [x] A4 Populate index metadata and key links
- [x] A5 Relate initial evidence files to design doc

## Phase B - Engine Hard Cutover (`go-go-os`)

- [ ] B1 Remove template fallback helpers from `artifactRuntime.ts` (`templateToCardId`, template icon mapping)
- [ ] B2 Make `buildArtifactOpenWindowPayload` runtime-card-first (requires `runtimeCardId`)
- [ ] B3 Remove default `stackId: inventory` fallback from artifact open payload path
- [ ] B4 Update `hypercardWidget.tsx` to remove template-based `Edit` routing
- [ ] B5 Gate widget/card open/edit controls on runtime-card presence
- [ ] B6 Update and pass `artifactRuntime.test.ts`
- [ ] B7 Update and pass `hypercardWidget.test.ts`
- [ ] B8 Run targeted engine tests for touched areas
- [ ] B9 Commit engine changes with task-referenced message

## Phase C - Inventory Fallback Card Removal (`go-go-app-inventory`)

- [ ] C1 Remove `reportViewer` card metadata from `apps/inventory/src/domain/stack.ts`
- [ ] C2 Remove `itemViewer` card metadata from `apps/inventory/src/domain/stack.ts`
- [ ] C3 Remove `reportViewer` implementation block from `pluginBundle.vm.js`
- [ ] C4 Remove `itemViewer` implementation block from `pluginBundle.vm.js`
- [ ] C5 Run inventory validation checks for touched files
- [ ] C6 Commit inventory changes with task-referenced message

## Phase D - Ticket Hygiene and Handoff

- [ ] D1 Update diary with each completed task + command/output summary
- [ ] D2 Update changelog with implementation milestones and related files
- [ ] D3 Update index summary/status and related tickets
- [ ] D4 Relate final touched files to ticket docs
- [ ] D5 Run `docmgr doctor --ticket GEPA-19-HYPERCARD-CARD-CUTOVER --stale-after 30`
- [ ] D6 Final review of remaining references to removed fallback cards

## Done Criteria

- [ ] No runtime path opens `reportViewer` or `itemViewer` as fallback cards
- [ ] Artifact/card open flow is runtime-card-first and template fallback is removed
- [ ] Ticket docs (plan/tasks/diary/changelog/index) are up-to-date and validated
