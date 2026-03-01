---
Title: 'Implementation Plan: Engine Final Chat Decoupling'
Ticket: GEPA-30
Status: active
Topics:
    - architecture
    - frontend
    - go-go-os
    - engine
    - chat-runtime
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/chat-runtime/src/chat/runtime/contextActions.ts
      Note: Chat-owned context action registration hooks
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/package.json
      Note: Dependency pruning scope
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/types.ts
      Note: Open-ended context target kind union
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/useDesktopShellController.tsx
      Note: Generic clipboard command routing in engine
    - Path: workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/parts.ts
      Note: Removed chat/ai part leftovers
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-28T20:09:40.828476559-05:00
WhatFor: Finalize removal of chat-specific ownership from @hypercard/engine and complete runtime boundary cleanup.
WhenToUse: Use when implementing or reviewing post-GEPA-29 cleanup tasks for engine/chat-runtime separation.
---


# Goal

Finish the chat/runtime extraction by removing remaining chat-coupled behaviors from `@hypercard/engine`, keeping engine APIs domain-agnostic, and ensuring tests cover the new boundaries.

# Scope

1. Ensure engine context-menu TSX tests are actually executed.
2. Remove hardcoded chat command behavior from engine (`chat.message.copy`).
3. Move chat context-action registration helpers to `@hypercard/chat-runtime`.
4. Generalize engine target kind unions to avoid hardcoded chat literals.
5. Remove stale chat/AI part constants from engine.
6. Remove unused heavyweight runtime dependencies from engine package metadata.
7. Update stale engine barrel usage comments.
8. Validate all impacted packages with typecheck/test/taxonomy checks.

# Implementation Decisions

## Generic Clipboard Command

- Replaced engine-only chat copy branch with generic `clipboard.copy-text` handling in the desktop shell controller.
- Chat runtime message renderer now emits `clipboard.copy-text` with `payload.text`.

Rationale: engine should provide generic platform commands, not chat-specific command IDs.

## Chat Context Registration Ownership

- Removed `useRegisterConversationContextActions` and `useRegisterMessageContextActions` from engine runtime hooks.
- Added equivalents in chat-runtime (`chat/runtime/contextActions.ts`) that compose generic engine APIs.

Rationale: registration semantics for chat entities belong to chat-runtime.

## Context Kind Generalization

- Updated engine target kind unions to:
  - `'window' | 'icon' | 'widget' | (string & {})`

Rationale: allow domain modules to define namespaced kinds (e.g. `chat.message`) without changing engine unions.

## Test Coverage Fix

- Enabled TSX tests in engine Vitest config.
- Rewrote `DesktopShell.contextMenu.test.tsx` to avoid stale import coupling to removed engine chat renderer path.
- Stabilized test fixture registration via `useMemo` to prevent re-registration loops.

# Validation Plan

Run all:

```bash
pnpm -C go-go-os storybook:check
pnpm -C go-go-os --filter @hypercard/engine typecheck
pnpm -C go-go-os --filter @hypercard/chat-runtime typecheck
pnpm -C go-go-os --filter @hypercard/hypercard-runtime typecheck
pnpm -C go-go-os --filter @hypercard/engine test
pnpm -C go-go-os --filter @hypercard/chat-runtime test
pnpm -C go-go-os --filter @hypercard/hypercard-runtime test
```

# Risks

- External consumers may still emit `chat.message.copy` and lose built-in copy behavior.
  - Accepted for this cutover; no compatibility wrapper added.
- Engine docs mentioning old chat helper hooks are stale after this change.
  - Follow-up doc cleanup can happen separately.
