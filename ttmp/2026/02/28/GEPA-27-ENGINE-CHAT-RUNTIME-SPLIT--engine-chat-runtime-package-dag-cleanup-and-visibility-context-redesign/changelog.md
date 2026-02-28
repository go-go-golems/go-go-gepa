# Changelog

## 2026-02-28

- Initial workspace created


## 2026-02-28

Completed architecture research and produced detailed design+implementation proposal for engine/chat/runtime split, visibility-context redesign, and story ownership migration plan; added execution-ready tasks and investigation diary.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT--engine-chat-runtime-package-dag-cleanup-and-visibility-context-redesign/design-doc/01-design-and-implementation-plan-split-chat-runtime-out-of-engine-and-introduce-generic-visibility-context.md — Primary architecture and implementation plan
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT--engine-chat-runtime-package-dag-cleanup-and-visibility-context-redesign/reference/01-investigation-diary-engine-chat-runtime-split-and-visibility-context.md — Chronological research diary


## 2026-02-28

Delivered GEPA-27 document bundle to reMarkable under /ai/2026/02/28/GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT and verified uploaded file listing.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT--engine-chat-runtime-package-dag-cleanup-and-visibility-context-redesign/reference/01-investigation-diary-engine-chat-runtime-split-and-visibility-context.md — Added upload verification details and command log


## 2026-02-28

Phase 1 implemented: engine now supports injected visibility context resolver and no longer hardcodes chatProfiles schema in DesktopShell controller (commit 8955530).

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/desktopShellTypes.ts — Added visibilityContextResolver integration point
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/useDesktopShellController.tsx — Removed chat-specific visibility context derivation from engine shell


## 2026-02-28

Implemented Phase 2+3: extracted chat runtime to new package, rewired runtime/app consumers, and updated cross-repo aliases/store wiring.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/chat-runtime/package.json — New package scaffold
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/hypercard-runtime/src/hypercard/timeline/registerHypercardTimeline.ts — Chat import rewires to `@hypercard/chat-runtime`
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-app-inventory/apps/inventory/src/launcher/renderInventoryApp.tsx — App-level chat API rewires
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/wesen-os/apps/os-launcher/src/app/store.ts — Launcher reducers now sourced from `@hypercard/chat-runtime`

### Commits

- go-go-os: `d36a5ee`
- go-go-app-inventory: `d5e9e46`
- go-go-app-arc-agi-3: `0dea6a1`
- wesen-os: `d4231c4`


## 2026-02-28

Implemented Phase 4+6: relocated runtime-coupled stories out of engine, simplified engine DesktopShell story surface, and added boundary guardrails (`deps:check`) to root test flow.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/chat-runtime/src/chat/components/ChatConversationWindow.stories.tsx — Story moved from engine
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/hypercard-runtime/src/hypercard/editor/CodeEditorWindow.stories.tsx — Story moved from engine
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/DesktopShell.stories.tsx — Rewritten to pure engine/windowing stories
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/scripts/deps/check-boundaries.mjs — Boundary guard script
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/scripts/storybook/check-taxonomy.mjs — Taxonomy check expanded for new package story roots

### Commit

- go-go-os: `9ee7efe`


## 2026-02-28

Fixed post-split chat-runtime test regression and completed full package test run.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/chat-runtime/src/chat/debug/yamlFormat.test.ts — Import corrected to local module

### Commit

- go-go-os: `10c7b6b`


## 2026-02-28

Uploaded final implementation diaries to reMarkable for both current ticket (GEPA-27) and previous refactor ticket (GEPA-26), then verified cloud listing.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-27-ENGINE-CHAT-RUNTIME-SPLIT--engine-chat-runtime-package-dag-cleanup-and-visibility-context-redesign/reference/01-investigation-diary-engine-chat-runtime-split-and-visibility-context.md
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-26-HYPERCARD-RUNTIME-SPLIT--split-hypercard-runtime-plugin-architecture-into-dedicated-package-separate-from-desktop-engine/reference/01-investigation-diary-hypercard-runtime-package-split.md


## 2026-02-28

Closed: scope implemented (engine/chat/runtime split cleanup delivered).

