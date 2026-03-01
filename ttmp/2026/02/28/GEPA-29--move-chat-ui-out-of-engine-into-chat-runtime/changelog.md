# Changelog

## 2026-02-28

- Initial workspace created


## 2026-02-28

Moved chat UI component ownership from engine to chat-runtime (ChatView/ChatWindow/StreamingChatView/ChatSidebar), added chat-local message types, rewired exports/imports, and removed engine chat exports (commit 360042f).

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/chat-runtime/src/chat/components/ChatSidebar.tsx — New owner of ChatSidebar
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/chat-runtime/src/chat/components/ChatWindow.tsx — New owner of ChatWindow
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/widgets/index.ts — Removed chat widget exports


## 2026-02-28

Moved chat stylesheet ownership to chat-runtime, added @hypercard/chat-runtime/theme export, removed implicit engine chat CSS import, and added explicit theme imports in Storybook + CRM entrypoint (commit 2ea378b).

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/chat-runtime/package.json — Added ./theme export
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/chat-runtime/src/chat/theme/chat.css — Chat style ownership moved here
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/theme/index.ts — Removed chat.css import


## 2026-02-28

GEPA-29 completed: chat UI ownership moved from engine to chat-runtime with explicit chat theme import boundary and passing validation matrix.

