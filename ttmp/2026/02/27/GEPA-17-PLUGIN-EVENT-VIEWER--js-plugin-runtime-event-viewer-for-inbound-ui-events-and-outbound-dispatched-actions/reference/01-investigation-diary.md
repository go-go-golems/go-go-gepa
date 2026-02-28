---
Title: Investigation diary
Ticket: GEPA-17-PLUGIN-EVENT-VIEWER
Status: active
Topics:
    - js-vm
    - event-streaming
    - go-go-os
    - hypercard
    - inventory-app
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/27/GEPA-17-PLUGIN-EVENT-VIEWER--js-plugin-runtime-event-viewer-for-inbound-ui-events-and-outbound-dispatched-actions/design-doc/01-plugin-runtime-event-viewer-architecture-and-implementation-plan.md
      Note: Primary design deliverable authored from this diary chronology
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/chat/debug/EventViewerWindow.tsx
      Note: Diary evidence for reusable stream-viewer UX patterns
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx
      Note: Diary evidence for runtime host event lifecycle tracing
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/desktopContributions.ts
      Note: Diary evidence for adapter/command contribution-based launch strategy
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/pluginIntentRouting.ts
      Note: Diary evidence for routing decisions and potential missing-session mismatch
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/debug/debugSlice.ts
      Note: Diary evidence for debug event capacity/filter behavior
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/features/pluginCardRuntime/pluginCardRuntimeSlice.ts
      Note: Diary evidence for timeline outcome/pending queue behavior
ExternalSources: []
Summary: Chronological command-and-findings diary for GEPA-17 plugin runtime event viewer investigation and delivery.
LastUpdated: 2026-02-27T23:08:12-05:00
WhatFor: Preserve reproducible investigation context, decisions, and validation evidence for intern onboarding and follow-up implementation.
WhenToUse: Use when implementing GEPA-17 or reviewing why specific event-viewer architecture decisions were made.
---


# Investigation diary

## Goal

Create ticket `GEPA-17-PLUGIN-EVENT-VIEWER` and produce an evidence-backed deep investigation for adding a running JS-plugin event viewer that shows inbound UI events and outbound dispatched actions, then validate and deliver the documentation bundle to reMarkable.

## Step 1: Resume context and lock skill workflow

This step re-established execution context after a partial earlier run and explicitly reloaded the required skills (`ticket-research-docmgr-remarkable`, `diary`) before making more changes. The focus was to avoid drifting from expected docmgr + diary + reMarkable workflow constraints.

I also validated where the ticket workspace and files currently stood so subsequent edits would not overwrite unexpected paths.

### Prompt Context

**User prompt (verbatim):** "Create a new ticket that exmaines how to add an event viewer to a running JS plugin to list all the events going in and the dispatched actions going out. 

Do an in depth investigation 5+ pages and store in the document and upload to remarkable. 

Keep a frequent diary."

**Assistant interpretation:** Produce an implementation-grade research package for plugin runtime event observability, including a long-form design document, frequent chronological diary, and reMarkable upload.

**Inferred user intent:** Enable a new intern to implement and debug plugin event tracing without reverse engineering runtime internals from scratch.

**Commit (code):** N/A (documentation/investigation turn)

### What I did

- Loaded required skill instructions and references:

```bash
sed -n '1,220p' /home/manuel/.codex/skills/ticket-research-docmgr-remarkable/SKILL.md
sed -n '1,260p' /home/manuel/.codex/skills/diary/SKILL.md
sed -n '1,260p' /home/manuel/.codex/skills/ticket-research-docmgr-remarkable/references/writing-style.md
sed -n '1,260p' /home/manuel/.codex/skills/ticket-research-docmgr-remarkable/references/deliverable-checklist.md
sed -n '1,320p' /home/manuel/.codex/skills/diary/references/diary.md
```

- Confirmed ticket workspace content and current doc stub state:

```bash
ls -la .../GEPA-17-PLUGIN-EVENT-VIEWER--js-plugin-runtime-event-viewer-for-inbound-ui-events-and-outbound-dispatched-actions
rg --files .../GEPA-17-PLUGIN-EVENT-VIEWER--js-plugin-runtime-event-viewer-for-inbound-ui-events-and-outbound-dispatched-actions
sed -n '1,220p' .../design-doc/01-plugin-runtime-event-viewer-architecture-and-implementation-plan.md
sed -n '1,260p' .../reference/01-investigation-diary.md
```

### Why

I needed to ensure workflow compliance and verify whether existing docs were placeholders before deep research continued.

### What worked

- Skill and reference docs were available and readable.
- Ticket docs were confirmed to still be templates (clear need for full authoring).

### What didn't work

- N/A in this step.

### What I learned

- The skill's strict diary format requirements are precise enough that it is easier to author from the template than patch incrementally.

### What was tricky to build

- The trickiness was procedural: blending skill constraints with previous partial progress without losing prior evidence extraction.

### What warrants a second pair of eyes

- N/A (process/setup step).

### What should be done in the future

- Keep a small reusable snippet for "skill load + checklist read + ticket health check" to reduce warm-up overhead.

### Code review instructions

- Verify only that skill/reference files were loaded and ticket docs were confirmed as stubs.

### Technical details

- Ticket path used throughout:
  - `/home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/27/GEPA-17-PLUGIN-EVENT-VIEWER--js-plugin-runtime-event-viewer-for-inbound-ui-events-and-outbound-dispatched-actions`

## Step 2: Resolve file-path drift and rebuild evidence map

The next step was an evidence-collection pass that initially hit path mismatches (`src/hypercard/runtime/...` vs current `src/components/shell/windowing/...` and `src/features/pluginCardRuntime/...`). I intentionally recorded those failures because they explain why the final references use a different path set than earlier assumptions.

After resolving the actual file locations, I rebuilt the runtime event-flow map with line-anchored reads.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Gather concrete architecture evidence for how plugin events move from UI interaction through VM and intent routing.

**Inferred user intent:** Remove guesswork and produce a reliable implementation baseline.

**Commit (code):** N/A

### What I did

- Initial (failed) path-based grep:

```bash
rg -n "emitRuntimeEvent|eventCard|dispatchRuntimeIntent|ingestRuntimeIntent|timeline|pendingDomainIntents|pendingSystemIntents" \
  .../src/hypercard/runtime/PluginCardSessionHost.tsx \
  .../src/hypercard/runtime/pluginIntentRouting.ts \
  .../src/store/pluginCardRuntimeSlice.ts \
  .../src/store/pluginCardRuntime/selectors.ts
```

- Observed errors:

```text
No such file or directory (os error 2)
```

- Resolved current paths and re-ran evidence search:

```bash
rg --files .../go-go-os/packages/engine/src | rg -n "PluginCardSessionHost|pluginIntentRouting|pluginCardRuntimeSlice|pluginCardRuntime/selectors"
rg -n "eventCard\(|dispatchRuntimeIntent\(|ingestRuntimeIntent\(|emitRuntimeEvent" .../go-go-os/packages/engine/src -S
```

- Extracted key line ranges:

```bash
nl -ba .../PluginCardSessionHost.tsx | sed -n '1,220p'
nl -ba .../PluginCardSessionHost.tsx | sed -n '220,373p'
nl -ba .../pluginIntentRouting.ts | sed -n '1,220p'
nl -ba .../pluginCardRuntimeSlice.ts | sed -n '1,220p'
nl -ba .../pluginCardRuntimeSlice.ts | sed -n '220,420p'
nl -ba .../selectors.ts | sed -n '1,220p'
```

### Why

The design doc needed exact line-level evidence for each major claim and to avoid stale path assumptions.

### What worked

- Runtime host flow was confirmed around `emitRuntimeEvent` and `runtimeService.eventCard` call path.
- Routing and reducer lifecycle were reconstructed with concrete line anchors.

### What didn't work

- Early path assumptions were stale and produced multiple `No such file or directory` errors.

### What I learned

- In this branch, plugin runtime host/routing is under `components/shell/windowing` and reducer logic under `features/pluginCardRuntime`.

### What was tricky to build

- The first search batch mixed old and new folder conventions, causing noisy failures and forcing a clean restart of the file map.

### What warrants a second pair of eyes

- Validate that cited paths/line ranges match the branch used for implementation (if rebased, line numbers may move).

### What should be done in the future

- Start every ticket with `rg --files` path discovery before issuing deep pattern queries.

### Code review instructions

- Start with `PluginCardSessionHost.tsx`, then `pluginIntentRouting.ts`, then `pluginCardRuntimeSlice.ts`.

### Technical details

- File sizes snapshot used to scope reading effort:

```bash
wc -l .../PluginCardSessionHost.tsx .../pluginIntentRouting.ts .../pluginCardRuntimeSlice.ts .../selectors.ts
```

## Step 3: Extract debug infrastructure and existing viewer capabilities

This step investigated whether a reusable debug stack already existed. The answer was yes: debug event contracts, reducer, hooks, and pane UI are already present and production-ready enough for a plugin event viewer foundation.

I focused on confirming both data structures and UI interaction behavior (filters, selected detail, state inspector) to avoid reinventing these pieces.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Verify whether plugin-event observability can reuse existing debug plumbing instead of introducing a parallel subsystem.

**Inferred user intent:** Minimize implementation risk and onboarding complexity.

**Commit (code):** N/A

### What I did

- Read debug contracts/hooks/slice/panes:

```bash
nl -ba .../cards/runtime.ts | sed -n '1,220p'
nl -ba .../debug/useStandardDebugHooks.ts | sed -n '1,220p'
nl -ba .../debug/debugSlice.ts | sed -n '1,220p'
nl -ba .../debug/StandardDebugPane.tsx | sed -n '1,220p'
nl -ba .../components/shell/RuntimeDebugPane.tsx | sed -n '1,260p'
```

- Read runtime card debug window:

```bash
nl -ba .../hypercard/debug/RuntimeCardDebugWindow.tsx | sed -n '1,280p'
```

- Verified API docs mention for debug hooks/pane:

```bash
wc -l .../docs/js-api-user-guide-reference.md
nl -ba .../docs/js-api-user-guide-reference.md | sed -n '200,280p'
```

### Why

I needed to determine if event viewer work should extend existing debug architecture or create a new event stack.

### What worked

- Confirmed `RuntimeDebugEvent` + `emitRuntimeDebugEvent` helper exists.
- Confirmed `debugSlice` capacity/filter/select behavior is already implemented.
- Confirmed UI pane can already list events and inspect payload/state.

### What didn't work

- One range query against docs used an out-of-range slice and returned no content initially.

### What I learned

- The infrastructure exists but plugin runtime flow is not currently wired into it.

### What was tricky to build

- Distinguishing "available infrastructure" from "currently integrated runtime paths" required both grep usage and line reads.

### What warrants a second pair of eyes

- Check if any app-level runtime path already wires these hooks indirectly (none found in current search, but worth a final verification during implementation).

### What should be done in the future

- Add a short architecture diagram in engine docs linking debug hooks to expected instrumentation points.

### Code review instructions

- Review `cards/runtime.ts` and `useStandardDebugHooks.ts` first, then `debugSlice.ts`, then pane components.

### Technical details

- Important baseline constants:
  - debug event capacity default: `300` (`debugSlice.ts`).
  - sanitization/truncation/redaction is already implemented in `sanitizeDebugValue`.

## Step 4: Investigate event-viewer precedents and pattern reuse

To avoid UI/behavior design from scratch, this step examined the existing chat `EventViewerWindow` and `eventBus` architecture. I treated this as a pattern library: filter controls, pause/resume, auto-scroll, payload expansion, and YAML export are all directly relevant to plugin viewer requirements.

The outcome was not "reuse chat bus directly," but "reuse its interaction patterns while keeping plugin event data in debug slice."

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Identify proven event-viewer UX patterns and retention behavior already exercised in tests.

**Inferred user intent:** Reduce implementation risk by using familiar, validated interfaces.

**Commit (code):** N/A

### What I did

- Read event viewer and bus implementation/tests:

```bash
nl -ba .../chat/debug/EventViewerWindow.tsx | sed -n '1,280p'
nl -ba .../chat/debug/EventViewerWindow.tsx | sed -n '280,520p'
nl -ba .../chat/debug/eventBus.ts | sed -n '1,260p'
nl -ba .../chat/debug/EventViewerWindow.test.ts | sed -n '1,320p'
nl -ba .../chat/debug/eventBus.test.ts | sed -n '1,260p'
```

- Read related stories for behavior examples:

```bash
nl -ba .../components/widgets/EventViewerWindow.stories.tsx | sed -n '1,260p'
nl -ba .../components/shell/RuntimeDebugPane.stories.tsx | sed -n '1,260p'
```

### Why

Existing tested behavior is better than ad-hoc viewer design, especially for intern onboarding.

### What worked

- Confirmed reusable control patterns and bounded-history semantics.
- Confirmed export/filter tests provide acceptance criteria templates.

### What didn't work

- N/A in this step.

### What I learned

- Chat event viewer is a strong UX reference, but data origin differs (chat bus vs plugin debug slice).

### What was tricky to build

- Separating reusable UI behaviors from chat-specific family semantics (`llm/tool/timeline/ws`) to avoid overfitting plugin viewer taxonomy.

### What warrants a second pair of eyes

- Whether to directly generalize `EventViewerWindow` into a shared component versus creating a plugin-specific window first.

### What should be done in the future

- Consider extracting a generic event-stream viewer shell component once plugin viewer ships and stabilizes.

### Code review instructions

- Compare `EventViewerWindow.tsx` controls with `RuntimeDebugPane.tsx` controls to decide convergence strategy.

### Technical details

- Chat bus history cap observed: `MAX_EVENT_HISTORY = 1000` (`eventBus.ts`).

## Step 5: Cross-reference recent tickets and extension-point realism

This step consulted nearby ticket context and validated actual extension surfaces for opening viewer UI in desktop runtime. The major outcome was confirming that adapters/commands/startup windows are first-class extension points and already used in app modules.

I also validated that older assumptions about `widget` window content are outdated in this branch (`WindowContentKind` is `card | app | dialog`), so integration should target `app` windows/adapters or existing debug windows.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Align this ticket with recent VM/runtime work and ensure integration recommendations match real extension APIs.

**Inferred user intent:** Avoid proposing architectures that cannot be wired in current code.

**Commit (code):** N/A

### What I did

- Checked recent tickets and GEPA-14 artifacts:

```bash
ls -1 .../go-go-gepa/ttmp/2026/02/27
rg --files .../GEPA-14-VM-JS-PROGRAMS--js-vm-programs-integration-for-hypercard-in-inventory-go-go-os
sed -n '1,240p' .../GEPA-14-VM-JS-PROGRAMS.../reference/01-investigation-diary.md
```

- Verified windowing/content integration points:

```bash
nl -ba .../desktop/core/state/types.ts | sed -n '1,320p'
nl -ba .../components/shell/windowing/windowContentAdapter.ts | sed -n '1,320p'
nl -ba .../components/shell/windowing/defaultWindowContentAdapters.tsx | sed -n '1,360p'
nl -ba .../components/shell/windowing/desktopContributions.ts | sed -n '1,260p'
nl -ba .../components/shell/windowing/useDesktopShellController.tsx | sed -n '840,1088p'
nl -ba .../components/shell/windowing/useDesktopShellController.tsx | sed -n '1152,1176p'
```

- Verified real app module adapter usage:

```bash
nl -ba .../apps/arc-agi-player/src/launcher/module.tsx | sed -n '1,260p'
nl -ba .../apps/apps-browser/src/launcher/module.tsx | sed -n '140,280p'
```

### Why

Implementation guidance needed to be realistic for current desktop shell architecture.

### What worked

- Confirmed contribution-based adapter/command pattern is live and used.
- Confirmed dedicated viewer can be opened as `content.kind='app'` with an `appKey` adapter.

### What didn't work

- Initial assumption that there may be a `windowContentAdapter.tsx` file was wrong in this branch (`windowContentAdapter.ts` is correct path).

### What I learned

- Runtime context targets still include `widget` for context menus, but window content rendering is adapter-based over `card|app|dialog`.

### What was tricky to build

- Reconciling story-only constructs (e.g., runtime story context actions) with production extension points without overstating availability.

### What warrants a second pair of eyes

- Decide whether v1 should embed viewer in `RuntimeCardDebugWindow` or ship directly as an app-window adapter path.

### What should be done in the future

- Add an explicit engine doc section for "debug tool windows as app adapters" to reduce confusion with widget context terminology.

### Code review instructions

- Start with `desktopContributions.ts`, then adapter composition in `useDesktopShellController.tsx`, then app launcher module examples.

### Technical details

- `WindowContentKind` currently: `'card' | 'app' | 'dialog'`.

## Step 6: Synthesize findings into architecture and identify correctness risk

Once file evidence was collected, I synthesized design options and selected the recommended approach: instrument plugin runtime flow via existing RuntimeDebugEvent infrastructure and build a plugin-specific viewer over debug slice.

During synthesis, I identified a potential correctness mismatch: if runtime session lookup fails in router, forwarding paths can still proceed because capability checks are conditional on runtimeSession?.capabilities.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Convert evidence into actionable architecture and surface implementation risks clearly.

**Inferred user intent:** Produce a document that both educates and protects implementation quality.

**Commit (code):** N/A

### What I did

- Re-read key routing and reducer lines to verify the risk hypothesis:

    nl -ba .../components/shell/windowing/pluginIntentRouting.ts | sed -n '81,119p'
    nl -ba .../features/pluginCardRuntime/pluginCardRuntimeSlice.ts | sed -n '235,310p'

- Searched for pending queue consumption paths:

    rg -n "selectPendingDomainIntents|selectPendingSystemIntents|selectPendingNavIntents|dequeuePending" .../go-go-os/packages/engine/src -S

### Why

I needed high-confidence findings before writing recommendations and pseudocode.

### What worked

- Confirmed concrete evidence for recommendation and identified possible missing-session forwarding inconsistency.
- Confirmed pending queues are produced, but no active consumption path is visible in the same engine surface.

### What didn't work

- N/A in this step.

### What I learned

- Observability work should include both instrumentation and route-hardening to avoid misleading event streams.

### What was tricky to build

- Avoiding overclaiming a bug without tests; I documented it as a risk requiring targeted test validation in implementation.

### What warrants a second pair of eyes

- Confirm router missing-session behavior with an explicit regression test before shipping instrumentation semantics that assume denial.

### What should be done in the future

- Add explicit contract tests for missing-session and malformed system payload routing outcomes.

### Code review instructions

- Verify missing-session branch assumptions directly in pluginIntentRouting.ts and confirm with a focused new test.

### Technical details

- Risk anchor:
  - Router lookup runtimeSession at pluginIntentRouting.ts:90-92.
  - Capability checks only when runtimeSession?.capabilities at pluginIntentRouting.ts:94-106.

## Step 7: Author primary design document and ticket scaffolding updates

This step produced the full 5+ page design document and updated ticket scaffold docs (index/tasks/changelog) to reflect current research status. The design doc includes architecture mapping, gap analysis, options, recommendation, pseudocode, phased implementation, testing strategy, and risks.

I wrote this with intern onboarding in mind: definitions, boundaries, and flow sequencing are all explicit.

### Prompt Context

**User prompt (verbatim):** (same as Step 1; resumed by a later "continue")

**Assistant interpretation:** Deliver the full research artifact, not just notes.

**Inferred user intent:** Hand over implementation-ready documentation for immediate engineering execution.

**Commit (code):** N/A

### What I did

- Authored design doc:

    cat > .../design-doc/01-plugin-runtime-event-viewer-architecture-and-implementation-plan.md

- Updated ticket scaffold docs:

    cat > .../index.md
    cat > .../tasks.md
    cat > .../changelog.md

### Why

The ticket required a complete, intern-friendly deliverable with explicit implementation guidance.

### What worked

- Design doc was fully authored from evidence gathered in prior steps.
- Index/tasks/changelog now reflect non-stub state.

### What didn't work

- N/A in this step.

### What I learned

- Reusing existing debug infrastructure is materially simpler than introducing a parallel plugin-event store.

### What was tricky to build

- Balancing depth (intern onboarding) with implementation realism (branch-specific extension points).

### What warrants a second pair of eyes

- Validate that proposed event taxonomy (plugin.*) aligns with any existing debug-kind naming conventions outside this ticket.

### What should be done in the future

- If adopted, codify event-kind taxonomy in a shared debug schema doc to prevent drift.

### Code review instructions

- Review design doc sections in order: current-state evidence, gap analysis, recommended architecture, pseudocode, phase plan.

### Technical details

- Timestamp captured for doc frontmatter updates:

    date -Iseconds
    # 2026-02-27T23:08:12-05:00

## Step 8: Validation, relate, and reMarkable delivery

This final step completed ticket bookkeeping and publication. I related evidence files to both docs, updated changelog entries, ran doctor validation, and then performed reMarkable dry-run/upload/list verification.

The result is a fully closed research delivery with concrete validation output and confirmed remote artifact presence.

### Prompt Context

**User prompt (verbatim):** (same as Step 1; resumed by a later "continue")

**Assistant interpretation:** Complete end-to-end delivery with validation and upload proof.

**Inferred user intent:** Ensure the documentation is finalized, validated, and accessible on reMarkable.

**Commit (code):** N/A

### What I did

- Executed operational pre-checks:

    docmgr status --summary-only
    remarquee status
    remarquee cloud account --non-interactive

- Observed outputs:

    root=/home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp ... tickets=22 stale=0 docs=59 ...
    remarquee: ok
    user=wesen@ruinwesen.com sync_version=1.5

- Related files to docs:

    docmgr doc relate --doc .../design-doc/01-plugin-runtime-event-viewer-architecture-and-implementation-plan.md --file-note ...
    docmgr doc relate --doc .../reference/01-investigation-diary.md --file-note ...

- Updated ticket changelog with final delivery entry:

    docmgr changelog update --ticket GEPA-17-PLUGIN-EVENT-VIEWER --entry \"Completed deep architecture investigation and authored implementation-grade plugin runtime event viewer design + diary deliverables; validated docs and prepared reMarkable handoff.\" --file-note ...

- Ran doc validation:

    docmgr doctor --ticket GEPA-17-PLUGIN-EVENT-VIEWER --stale-after 30

- Ran bundle upload workflow:

    remarquee upload bundle --dry-run ... --name \"GEPA-17 Plugin Runtime Event Viewer Research\" --remote-dir \"/ai/2026/02/27/GEPA-17-PLUGIN-EVENT-VIEWER\" --toc-depth 2
    remarquee upload bundle ... --name \"GEPA-17 Plugin Runtime Event Viewer Research\" --remote-dir \"/ai/2026/02/27/GEPA-17-PLUGIN-EVENT-VIEWER\" --toc-depth 2
    remarquee cloud ls /ai/2026/02/27/GEPA-17-PLUGIN-EVENT-VIEWER --long --non-interactive

- Observed final outputs:

    ## Doctor Report (1 findings)
    ### GEPA-17-PLUGIN-EVENT-VIEWER
    - ✅ All checks passed

    OK: uploaded GEPA-17 Plugin Runtime Event Viewer Research.pdf -> /ai/2026/02/27/GEPA-17-PLUGIN-EVENT-VIEWER
    [f]    GEPA-17 Plugin Runtime Event Viewer Research

### Why

This is required to close the ticket with traceable quality gates and publication evidence.

### What worked

- Tooling status checks passed.
- `docmgr doc relate` updated both docs successfully.
- `docmgr doctor` passed cleanly.
- reMarkable dry-run and real upload both succeeded.
- Remote listing confirmed artifact availability.

### What didn't work

- N/A in this step.

### What I learned

- Dry-run before real upload catches packaging issues with zero remote side effects and should stay mandatory.

### What was tricky to build

- Maintaining strict chronology while substantial writing and operational steps occur in interleaved order.

### What warrants a second pair of eyes

- Confirm whether reMarkable naming should be standardized across tickets (current bundle uses human-readable title without `.pdf` in listing).

### What should be done in the future

- Keep a reusable shell snippet for bundle dry-run + upload + list verification to reduce typo risk in long commands.

### Code review instructions

- Verify `docmgr doctor` output is clean, then verify the exact remote path and filename in `remarquee cloud ls`.

### Technical details

- Final remote destination:

    /ai/2026/02/27/GEPA-17-PLUGIN-EVENT-VIEWER
