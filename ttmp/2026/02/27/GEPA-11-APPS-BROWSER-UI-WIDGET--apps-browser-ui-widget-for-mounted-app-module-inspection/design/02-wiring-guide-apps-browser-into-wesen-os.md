---
Title: 'Wiring guide: Apps Browser into wesen-os desktop launcher'
Ticket: GEPA-11-APPS-BROWSER-UI-WIDGET
Status: active
Topics:
    - frontend
    - integration
    - go-go-os
    - wesen-os
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../go-go-os/apps/apps-browser/src/launcher/module.tsx
      Note: The LaunchableAppModule export and window content adapter
    - Path: ../../../../../../../go-go-os/apps/apps-browser/src/launcher/public.ts
      Note: Public exports for launcher integration
    - Path: ../../../../../../../wesen-os/apps/os-launcher/src/app/modules.tsx
      Note: Where all desktop app modules are registered
    - Path: ../../../../../../../wesen-os/apps/os-launcher/src/app/store.ts
      Note: Where shared reducers are merged into the launcher store
    - Path: ../../../../../../../wesen-os/apps/os-launcher/src/app/registry.ts
      Note: App registry creation from module array
    - Path: ../../../../../../../go-go-os/packages/desktop-os/src/contracts/launchableAppModule.ts
      Note: The LaunchableAppModule contract interface
    - Path: ../../../../../../../go-go-os/packages/desktop-os/src/store/createLauncherStore.ts
      Note: Store factory — does NOT support RTK Query middleware
    - Path: ../../../../../../../go-go-os/packages/engine/src/app/createAppStore.ts
      Note: Base store factory used by createLauncherStore
Summary: Step-by-step guide for a fullstack engineer to wire the completed Apps Browser frontend package into the wesen-os desktop launcher.
LastUpdated: 2026-02-27
WhatFor: Enable a fullstack engineer to integrate the apps-browser package with minimal context.
WhenToUse: When ready to ship apps-browser to the running desktop environment.
---


# Wiring guide: Apps Browser into wesen-os desktop launcher

## What was built

The `@hypercard/apps-browser` package (`go-go-os/apps/apps-browser/`) is a complete frontend widget for inspecting mounted backend modules. It provides four window types:

| Window | Component | Purpose |
|--------|-----------|---------|
| **Mounted Apps** (folder) | `AppsFolderWindow` | Icon grid of all mounted modules with health/required/reflection badges |
| **Module Browser** | `ModuleBrowserWindow` | Three-column Smalltalk-style inspector with cascading selection |
| **Get Info** | `GetInfoWindow` | Single-module inspector panel (one per module) |
| **Health Dashboard** | `HealthDashboardWindow` | Summary cards + module health list with degraded warning |

All components use **RTK Query** for data fetching (not Redux thunks). The package manages its own Redux store with RTK Query middleware.

### Storybook verification

Run `pnpm storybook` from go-go-os root. All stories are under **Apps / AppsBrowser**. Each window type has multiple MSW-backed stories showing healthy, degraded, loading, and empty states.


## Architecture: how it differs from CRM/Todo

CRM and Todo use the **HyperCard stack/card pattern**: their windows render `PluginCardSessionHost` via card-based content adapters. They don't need extra middleware.

Apps Browser uses **RTK Query** to fetch from `/api/os/apps` and `/api/os/apps/:appId/reflection`. This requires RTK Query middleware that `createAppStore()` / `createLauncherStore()` do not support. The solution:

- Apps Browser creates its **own isolated Redux store** (`createAppsBrowserStore()`) with RTK Query middleware.
- The launcher module wraps window content in a `<Provider store={appsBrowserStore}>` internally.
- **No shared reducers** need to be added to the launcher store — apps-browser is fully self-contained.

This is the same pattern used by any app that wraps its own store in `renderWindow()` (CRM does this too with `CrmLauncherAppHost`).


## Backend endpoints consumed

Apps Browser fetches from two endpoints that already exist in go-go-os backendhost:

| Method | Path | Response | Source |
|--------|------|----------|--------|
| `GET` | `/api/os/apps` | `{ "apps": AppManifestDocument[] }` | `go-go-os/pkg/backendhost/manifest_endpoint.go` |
| `GET` | `/api/os/apps/:appId/reflection` | `ModuleReflectionDocument` or `501` | `go-go-os/pkg/backendhost/manifest_endpoint.go` |

**No new backend work is needed.** These endpoints are already served by the running wesen-os launcher.


## Step-by-step wiring

### Step 1: Register the module in `modules.tsx`

**File to edit:** `wesen-os/apps/os-launcher/src/app/modules.tsx`

```typescript
import { bookTrackerLauncherModule } from '@hypercard/book-tracker-debug/src/launcher/module';
import { crmLauncherModule } from '@hypercard/crm/src/launcher/module';
import type { LaunchableAppModule } from '@hypercard/desktop-os';
import { inventoryLauncherModule } from '@hypercard/inventory/launcher';
import { todoLauncherModule } from '@hypercard/todo/src/launcher/module';
// ADD THIS IMPORT:
import { appsBrowserLauncherModule } from '@hypercard/apps-browser/launcher';

export const launcherModules: LaunchableAppModule[] = [
  inventoryLauncherModule,
  todoLauncherModule,
  crmLauncherModule,
  bookTrackerLauncherModule,
  appsBrowserLauncherModule,  // ADD THIS LINE
];
```

The import path `@hypercard/apps-browser/launcher` resolves via the package.json exports map:

```json
{
  "exports": {
    ".": "./src/index.ts",
    "./launcher": "./src/launcher/public.ts"
  }
}
```

### Step 2: Verify pnpm can resolve the package

From `wesen-os/` root:

```bash
pnpm install
```

If `@hypercard/apps-browser` is not already resolvable from `wesen-os`, add it to `wesen-os/apps/os-launcher/package.json` dependencies:

```json
{
  "dependencies": {
    "@hypercard/apps-browser": "workspace:*"
  }
}
```

Or if wesen-os references go-go-os via a different mechanism (git submodule, linked workspace), adjust the pnpm workspace config accordingly.

### Step 3: No store changes needed

**You do NOT need to edit `store.ts`.** Apps Browser does not export shared reducers. It manages its own store internally. The launcher store only needs the module registered in `modules.tsx`.

For clarity, here is what happens automatically when you add the module:

1. `createAppRegistry(launcherModules)` indexes the module for lookup.
2. `createLauncherStore(launcherModules)` sees `module.state` is undefined, so no reducer is added.
3. `buildLauncherContributions()` creates a desktop icon (order: 90) and collects the window content adapter.
4. When the user clicks the icon, `buildLaunchWindow()` opens the "Mounted Apps" folder window.
5. The content adapter routes all `apps-browser:*` appKeys to the correct component.

### Step 4: Verify it works

```bash
# From wesen-os root
pnpm dev

# Or from go-go-os root for just the frontend
pnpm storybook
```

Expected behavior:
- A new desktop icon appears (folder icon, last in row due to order: 90).
- Clicking it opens the "Mounted Apps" folder window showing module icons.
- Double-clicking a module icon should (once wired with openWindow callbacks) open the Module Browser.


## Window content adapter routing

The apps-browser content adapter handles all windows internally using the `appKey` string encoded in `WindowContent`:

| appKey pattern | Window rendered |
|----------------|-----------------|
| `apps-browser:folder` | `AppsFolderWindow` |
| `apps-browser:browser` | `ModuleBrowserWindow` (no initial selection) |
| `apps-browser:browser:gepa` | `ModuleBrowserWindow` (pre-selects gepa) |
| `apps-browser:health` | `HealthDashboardWindow` |
| `apps-browser:get-info:gepa` | `GetInfoWindowByAppId` (fetches gepa from API) |

Each window type is singleton via `dedupeKey`. Get Info windows are singleton per module (`dedupeKey: 'apps-browser:get-info:{appId}'`).


## Exported window payload builders

For programmatic window opening from other parts of the launcher (menus, commands, keyboard shortcuts), import from `@hypercard/apps-browser/launcher`:

```typescript
import {
  appsBrowserLauncherModule,       // The LaunchableAppModule
  buildBrowserWindowPayload,       // (initialAppId?: string) => OpenWindowPayload
  buildHealthWindowPayload,        // () => OpenWindowPayload
  buildGetInfoWindowPayload,       // (appId: string, appName?: string) => OpenWindowPayload
} from '@hypercard/apps-browser/launcher';
```

Example — opening Health Dashboard from a menu command:

```typescript
hostContext.openWindow(buildHealthWindowPayload());
```

Example — opening Get Info for a specific module:

```typescript
hostContext.openWindow(buildGetInfoWindowPayload('gepa', 'GEPA'));
```


## Future wiring work (not blocking)

These are optional enhancements beyond the basic wiring:

### Desktop menu contributions

The module currently registers only a window content adapter. To add menu items:

```typescript
// In createContributions():
{
  id: 'apps-browser.menus',
  menus: [
    {
      id: 'apps-browser.window-menu',
      menuId: 'Window',
      items: [
        { id: 'open-health', label: 'Health Dashboard', command: 'apps-browser:open-health' },
        { id: 'open-browser', label: 'Module Browser', command: 'apps-browser:open-browser' },
      ],
    },
  ],
  commands: [
    {
      id: 'apps-browser:open-health',
      execute: (ctx) => ctx.openWindow(buildHealthWindowPayload()),
    },
    {
      id: 'apps-browser:open-browser',
      execute: (ctx) => ctx.openWindow(buildBrowserWindowPayload()),
    },
  ],
}
```

### Context menu on folder icons

The `AppsFolderWindow` and `AppIcon` components accept `onSelectApp`, `onOpenApp`, and `onContextMenu` callbacks. To wire context menus, the launcher module's content adapter would need to pass callbacks that dispatch `openWindow()` calls via the host context.

### Cross-window navigation

To enable "Open in Browser" from Get Info, or "Get Info" from right-clicking a folder icon, the components need access to the host context's `openWindow()`. This can be done via React context provided by `AppsBrowserHost`.

### Extend `createLauncherStore` for RTK Query

Long-term, if apps-browser should share a single store with the launcher (for cross-app state access), `createAppStore`/`createLauncherStore` would need an `apiSlices` parameter to inject RTK Query middleware. This is tracked as a future engine improvement.


## File reference

### Package structure

```
go-go-os/apps/apps-browser/
├── package.json                          # @hypercard/apps-browser
├── tsconfig.json
└── src/
    ├── index.ts                          # Public barrel exports
    ├── api/
    │   └── appsApi.ts                    # RTK Query: getApps, getReflection
    ├── app/
    │   ├── store.ts                      # Self-contained store with RTK Query middleware
    │   └── stories/
    │       └── AppsBrowserApp.stories.tsx # Full-app stories (6)
    ├── components/
    │   ├── AppIcon.tsx + .css             # Module icon with badge composition
    │   ├── AppIcon.stories.tsx            # (7 stories)
    │   ├── AppsFolderWindow.tsx + .css    # Sorted icon grid + status toolbar
    │   ├── AppsFolderWindow.stories.tsx   # (5 stories, MSW-backed)
    │   ├── BrowserColumns.tsx             # ModuleListPane, APIListPane, SchemaListPane
    │   ├── BrowserDetailPanel.tsx         # Contextual detail (module/API/schema)
    │   ├── ModuleBrowserWindow.tsx + .css # Three-column browser
    │   ├── ModuleBrowserWindow.stories.tsx # (6 stories, MSW-backed)
    │   ├── GetInfoWindow.tsx + .css       # Single-module inspector
    │   ├── GetInfoWindow.stories.tsx      # (4 stories, MSW-backed)
    │   ├── GetInfoWindowByAppId.tsx       # Wrapper that fetches app by ID
    │   ├── HealthDashboardWindow.tsx + .css # Summary + module health list
    │   └── HealthDashboardWindow.stories.tsx # (4 stories, MSW-backed)
    ├── domain/
    │   ├── types.ts                       # AppManifestDocument, ModuleReflectionDocument, etc.
    │   ├── sorting.ts                     # sortApps, computeSummaryStats
    │   └── selectors.ts                   # getCrossRefSchemaIds, findApi, findSchema
    ├── features/
    │   └── appsBrowser/
    │       └── appsBrowserSlice.ts        # Selection state (selectedAppId, etc.)
    ├── launcher/
    │   ├── module.tsx                     # LaunchableAppModule + window adapter
    │   └── public.ts                      # Re-exports for @hypercard/apps-browser/launcher
    └── mocks/
        ├── fixtures/
        │   └── apps.ts                    # Mock app manifests + reflection data
        └── msw/
            ├── createAppsHandlers.ts      # MSW handler factory
            └── defaultHandlers.ts         # Default MSW wiring

```

### Key files in wesen-os (to edit)

| File | Action |
|------|--------|
| `wesen-os/apps/os-launcher/src/app/modules.tsx` | Add import + array entry (Step 1) |
| `wesen-os/apps/os-launcher/package.json` | Add dependency if needed (Step 2) |

### Key files in go-go-os (read-only reference)

| File | What it does |
|------|-------------|
| `packages/desktop-os/src/contracts/launchableAppModule.ts` | Module registration contract |
| `packages/desktop-os/src/store/createLauncherStore.ts` | Store factory (no RTK Query support) |
| `packages/desktop-os/src/registry/createAppRegistry.ts` | Module indexing |
| `packages/desktop-os/src/runtime/buildLauncherContributions.ts` | Icon + command generation |


## Commits

All implementation work is on branch `task/add-gepa-optimizer` in go-go-os:

| Commit | Phase | Description |
|--------|-------|-------------|
| `3748972` | 5-7 | Scaffold, domain types, RTK Query, MSW mocks |
| `67331b4` | 8 | AppIcon component + CSS + stories |
| `9393be6` | 9 | AppsFolderWindow + MSW fix |
| `0e9defd` | 10 | ModuleBrowserWindow (3-column browser) |
| `29ca612` | 11 | GetInfoWindow inspector |
| `55bc8bb` | 12 | HealthDashboardWindow + degraded state |
| `7ddc75b` | 13 | Launcher module + full-app stories |
| `ff9ae78` | 14 | Lint fixes |
