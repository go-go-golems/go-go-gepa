---
Title: 'Implementation Plan: Engine Chat UI Extraction'
Ticket: GEPA-29
Status: active
Topics:
    - architecture
    - frontend
    - go-go-os
    - chat-runtime
    - hypercard
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/chat-runtime/package.json
      Note: Theme subpath export for explicit style ownership
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/chat-runtime/src/chat/components/ChatConversationWindow.tsx
      Note: Primary importer that needed ChatWindow ownership change
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/chat-runtime/src/chat/index.ts
      Note: Chat runtime public API target surface
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/widgets/index.ts
      Note: Engine widget barrel cleanup scope
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/desktop/react/index.ts
      Note: Engine desktop-react export boundary
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/theme/index.ts
      Note: Removal of implicit chat css import
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-28T19:49:39.584855769-05:00
WhatFor: Plan and rationale for moving chat UI primitives out of @hypercard/engine into @hypercard/chat-runtime to keep a clean package DAG.
WhenToUse: Use when implementing or reviewing the GEPA-29 package-boundary cleanup.
---


# Goal

Move chat-specific UI primitives and chat-only styling out of `@hypercard/engine` into `@hypercard/chat-runtime`, while keeping the desktop/windowing visibility system generic in engine.

# Current State

`@hypercard/engine` still contains chat UI surface:

- `components/widgets/ChatView.tsx`
- `components/widgets/ChatWindow.tsx`
- `components/widgets/StreamingChatView.tsx`
- `components/shell/ChatSidebar.tsx`
- `theme/desktop/chat.css`
- exports from engine barrels (`components/widgets/index.ts`, `desktop/react/index.ts`, `components/shell/index.ts`)

`@hypercard/chat-runtime` currently consumes `ChatWindow` from engine in `ChatConversationWindow.tsx`.

# Target State

## Dependency DAG

After this ticket:

- `@hypercard/engine`: generic desktop shell, windowing, widget primitives, no chat-specific widgets/styles
- `@hypercard/chat-runtime`: chat widgets, chat sidebar, chat conversation window, chat styling, chat timeline renderers/state/runtime
- `@hypercard/hypercard-runtime`: depends on engine + chat-runtime

No package in `engine` should import from `chat-runtime` or `hypercard-runtime`.

## Ownership

- Move chat widget implementation files and stories into `packages/chat-runtime/src/chat/components/`.
- Keep generic UI primitives (`Btn`, `Chip`, etc.) in engine.
- Create chat-runtime theme entrypoint and move `chat.css` there.
- Remove chat exports from engine public API and types where no longer needed.

# Design Choices

## 1) Keep visibility context generic in engine

The menu visibility context (`allowedRoles`, `allowedProfiles`, `when`, `unauthorized`) remains in engine because it is domain-agnostic and useful for non-chat plugins too.

## 2) Keep shared primitives in engine

Chat widgets may continue to use generic primitives (`Btn`, `Chip`) imported from engine. This keeps tokenized visual primitives centralized while moving chat-specific compositions out.

## 3) Introduce explicit chat-runtime theme import

Engine no longer loads chat CSS by default. Apps using chat runtime UI should import:

```ts
import '@hypercard/chat-runtime/theme';
```

This prevents silent cross-domain styling and makes ownership explicit.

# Implementation Plan

## Phase 1: Move component files

1. Add chat message types in chat-runtime (`ChatMessage`, `ChatMessageStatus`).
2. Move `ChatView`, `ChatWindow`, `StreamingChatView`, `ChatSidebar` to `chat-runtime/src/chat/components`.
3. Move chat widget stories to chat-runtime and update taxonomy titles to `ChatRuntime/Components/*`.

## Phase 2: Wire exports and imports

1. Export moved components/types from `chat-runtime/src/chat/index.ts`.
2. Update `ChatConversationWindow` to import local `ChatWindow`.
3. Remove chat exports and files from engine barrels.
4. Remove chat message type from engine `types.ts`.

## Phase 3: Move style ownership

1. Move `engine/src/theme/desktop/chat.css` to `chat-runtime/src/chat/theme/chat.css`.
2. Add `chat-runtime/src/theme/index.ts` that imports chat CSS.
3. Add chat-runtime package export `"./theme": "./src/theme/index.ts"`.
4. Remove chat import from engine theme index.
5. Add `@hypercard/chat-runtime/theme` imports in app entrypoints that use chat runtime (at least CRM and apps-browser storybook-facing entrypoints if needed).

## Phase 4: Validation and cleanup

1. Run `storybook:check`, package `typecheck`, and package tests.
2. Update docs/changelog/diary with exact commit references.
3. Close ticket after all tasks are checked.

# Risks and Mitigations

- Risk: visual regression if chat CSS not imported in app entrypoints.
  - Mitigation: add explicit theme imports where chat runtime is mounted; verify Storybook chat stories.
- Risk: barrel API break for engine consumers.
  - Mitigation: this ticket intentionally allows non-backward-compatible cleanup; update internal call-sites in same change.
- Risk: taxonomy check failures for moved stories.
  - Mitigation: place stories under `packages/chat-runtime/src/chat/components` and use `ChatRuntime/Components/*` titles.

# Task Checklist

Execution tasks are tracked in ticket task list (`tasks.md`) and checked as code lands.
