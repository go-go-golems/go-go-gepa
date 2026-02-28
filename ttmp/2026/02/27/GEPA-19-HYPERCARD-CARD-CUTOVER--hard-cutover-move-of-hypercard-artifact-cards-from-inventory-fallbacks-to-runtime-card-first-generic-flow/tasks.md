# Tasks

## Phase A - Planning and Ticket Setup

- [x] A1 Create GEPA-19 ticket workspace via `docmgr ticket create-ticket`
- [x] A2 Add design-doc with hard-cutover implementation plan
- [x] A3 Add reference diary document
- [x] A4 Populate index metadata and key links
- [x] A5 Relate initial evidence files to design doc

## Phase B - Engine Hard Cutover (`go-go-os`)

- [x] B1 Remove template fallback helpers from `artifactRuntime.ts` (`templateToCardId`, template icon mapping)
- [x] B2 Make `buildArtifactOpenWindowPayload` runtime-card-first (requires `runtimeCardId`)
- [x] B3 Remove default `stackId: inventory` fallback from artifact open payload path
- [x] B4 Update `hypercardWidget.tsx` to remove template-based `Edit` routing
- [x] B5 Gate widget/card open/edit controls on runtime-card presence
- [x] B6 Update and pass `artifactRuntime.test.ts`
- [x] B7 Update and pass `hypercardWidget.test.ts`
- [x] B8 Run targeted engine tests for touched areas
- [x] B9 Commit engine changes with task-referenced message

## Phase C - Inventory Fallback Card Removal (`go-go-app-inventory`)

- [x] C1 Remove `reportViewer` card metadata from `apps/inventory/src/domain/stack.ts`
- [x] C2 Remove `itemViewer` card metadata from `apps/inventory/src/domain/stack.ts`
- [x] C3 Remove `reportViewer` implementation block from `pluginBundle.vm.js`
- [x] C4 Remove `itemViewer` implementation block from `pluginBundle.vm.js`
- [x] C5 Run inventory validation checks for touched files
- [x] C6 Commit inventory changes with task-referenced message

## Phase D - Ticket Hygiene and Handoff

- [x] D1 Update diary with each completed task + command/output summary
- [x] D2 Update changelog with implementation milestones and related files
- [x] D3 Update index summary/status and related tickets
- [x] D4 Relate final touched files to ticket docs
- [x] D5 Run `docmgr doctor --ticket GEPA-19-HYPERCARD-CARD-CUTOVER --stale-after 30`
- [x] D6 Final review of remaining references to removed fallback cards

## Done Criteria

- [x] No runtime path opens `reportViewer` or `itemViewer` as fallback cards
- [x] Artifact/card open flow is runtime-card-first and template fallback is removed
- [x] Ticket docs (plan/tasks/diary/changelog/index) are up-to-date and validated
