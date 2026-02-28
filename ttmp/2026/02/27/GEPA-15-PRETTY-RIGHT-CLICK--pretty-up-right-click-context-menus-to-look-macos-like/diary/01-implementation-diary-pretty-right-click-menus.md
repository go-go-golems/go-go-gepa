---
Title: Implementation Diary - Pretty Right-Click Menus
Ticket: GEPA-15-PRETTY-RIGHT-CLICK
Status: active
Topics:
    - frontend
    - ui
    - context-menu
DocType: diary
Intent: long-term
Owners: []
RelatedFiles:
    - go-go-os/packages/engine/src/components/widgets/ContextMenu.tsx:Core context menu component - restructured for macOS layout
    - go-go-os/packages/engine/src/components/widgets/ContextMenu.stories.tsx:Stories updated with new Finder-style and disabled state demos
    - go-go-os/packages/engine/src/theme/desktop/tokens.css:New CSS custom properties for context menu styling
    - go-go-os/packages/engine/src/theme/desktop/primitives.css:Rewritten context menu and dropdown CSS rules
    - go-go-os/packages/engine/src/theme/desktop/shell.css:Updated windowing menu panel CSS
    - go-go-os/packages/engine/src/parts.ts:New data-part entries for context menu item columns
ExternalSources: []
Summary: "Overhauled right-click context menus, dropdown menus, and menu bar panels to use macOS-like styling with rounded corners, soft shadows, backdrop blur, blue hover highlights, and proper column layout."
LastUpdated: 2026-02-27T22:37:44.223320134-05:00
WhatFor: ""
WhenToUse: ""
---

# Implementation Diary: Pretty Right-Click Context Menus

## 2026-02-27 - Initial Implementation

### Goal
Make the right-click context menus (and related menu panels) look modern and macOS-like instead of the retro System 1 HyperCard aesthetic with hard pixel shadows, inverted black/white hover, and dashed separators.

### What Was Changed

#### 1. New CSS Tokens (`tokens.css`)
Added 14 new CSS custom properties for fine-grained control over context menu appearance:
- `--hc-context-menu-bg`: Translucent white (`rgba(255,255,255,0.88)`) for backdrop blur
- `--hc-context-menu-border`: Subtle 1px border (`rgba(0,0,0,0.12)`)
- `--hc-context-menu-border-radius`: 6px rounded corners
- `--hc-context-menu-shadow`: Multi-layer soft shadow (two shadow layers)
- `--hc-context-menu-item-hover-bg`: Blue highlight (#0a6cff)
- `--hc-context-menu-item-hover-fg`: White text on hover
- `--hc-context-menu-item-disabled-fg`: Grayed-out disabled text
- `--hc-context-menu-separator-color`: Subtle separator line
- Plus matching tokens for dropdown panel and windowing menu panel

Also updated `--hc-menu-hover-bg` from inverted `var(--hc-color-fg)` to blue `#0a6cff`.

#### 2. New Data Parts (`parts.ts`)
Added three new part names for the three-column item layout:
- `contextMenuItemCheck` (`context-menu-item-check`): Checkmark/icon column (18px fixed width)
- `contextMenuItemLabel` (`context-menu-item-label`): Label text (flex: 1)
- `contextMenuItemShortcut` (`context-menu-item-shortcut`): Keyboard shortcut (right-aligned, muted color)

#### 3. Component Restructure (`ContextMenu.tsx`)
- Items now render as `<button>` with three inner `<span>` elements for check, label, shortcut
- Added `activeIndex` state for hover tracking (supports keyboard navigation prep)
- Added viewport clamping: menu repositions to stay within viewport bounds
- Added `data-has-checks` and `data-has-shortcuts` attributes on the root for CSS-level layout control
- String items conditionally render the check column only when other items have checks (visual alignment)

#### 4. CSS Rewrites (`primitives.css`, `shell.css`)

**Context Menu (`primitives.css`):**
- Replaced 2px solid black border with subtle 1px translucent border
- Replaced hard `2px 2px 0 #000` shadow with layered soft shadows
- Added `backdrop-filter: blur(20px) saturate(1.6)` for glassmorphism
- Added `border-radius: 6px` and `padding: 4px`
- Added `animation: hc-context-menu-in 0.12s ease-out` fade+scale entrance
- Items now flex with rounded `border-radius: 4px` hover highlight
- Hover color changed from inverted (black bg) to blue (#0a6cff)
- Separators changed from `1px dashed` to `1px solid rgba(0,0,0,0.10)` with margin
- Disabled items grayed out at 30% opacity
- Shortcut text: muted color that turns white on hover

**Dropdown Menu (`primitives.css`):**
- Panel now matches context menu: rounded, blurred, soft shadow, blue hover
- Trigger updated with lighter border and border-radius

**Windowing Menu Panel (`shell.css`):**
- Menu bar dropdown panels now match: rounded corners, blur, soft shadows
- Menu items have rounded hover highlights
- Separators and shortcuts match context menu styling
- Added disabled state support

#### 5. Stories (`ContextMenu.stories.tsx`)
Added two new stories:
- **FinderStyle**: macOS Finder-like menu with checkmarks, shortcuts, and multiple separator groups
- **MixedWithDisabled**: Edit menu with Undo/Redo/Cut/Copy/Paste showing disabled states
- Updated existing stories to use macOS-style Unicode shortcut symbols

### What Worked
- The `data-part` attribute system made CSS changes completely decoupled from component logic
- CSS custom properties (tokens) make it trivial for themes to override the new defaults
- `backdrop-filter: blur()` works well in modern browsers and degrades gracefully
- The three-column layout (check/label/shortcut) properly aligns even with mixed item types
- All 3 existing context menu tests pass unchanged

### What Was Tricky
- Font size: Initially set to 13px (macOS native size) but the user pointed out it should match the existing `--hc-font-size` (11px) for consistency with the rest of the HyperCard engine. Fixed to use `var(--hc-font-size)`.
- Viewport clamping: The menu needs to reposition if it would overflow the viewport edge, required a `useEffect` to measure and adjust after initial render.

### How to Review/Verify
1. Run Storybook: `cd go-go-os && npx storybook dev`
2. Navigate to Engine > Widgets > ContextMenu
3. Check each story variant: Default, WithSeparators, ManyItems, ActionEntries, FinderStyle, MixedWithDisabled
4. Verify hover highlights are blue with rounded corners
5. Verify disabled items are grayed out and non-interactive
6. Verify shortcuts are right-aligned
7. Run tests: `npx vitest run` (all 541 tests should pass)
