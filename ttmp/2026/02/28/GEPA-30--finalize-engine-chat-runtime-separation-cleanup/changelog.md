# Changelog

## 2026-02-28

- Initial workspace created


## 2026-02-28

Enabled engine TSX test execution, replaced stale DesktopShell context-menu test import with local fixture wiring, and replaced chat.message.copy with generic clipboard.copy-text command handling (commit 33592e4).

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/DesktopShell.contextMenu.test.tsx — Stale import removed and fixture stabilized
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/useDesktopShellController.tsx — Generic clipboard command branch
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/vitest.config.ts — TSX tests now included


## 2026-02-28

Moved chat context-action registration wrappers to chat-runtime, generalized engine context kind unions, removed chat/ai part leftovers, pruned unused engine deps, and updated engine barrel usage comment (commit a04fafd).

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/chat-runtime/src/chat/runtime/contextActions.ts — Chat-owned wrappers
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/package.json — Removed unused heavy dependencies
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/types.ts — Generalized context kind union
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/parts.ts — Removed chat/ai constants


## 2026-02-28

GEPA-30 complete: final engine/chat-runtime decoupling tasks delivered and validated.

