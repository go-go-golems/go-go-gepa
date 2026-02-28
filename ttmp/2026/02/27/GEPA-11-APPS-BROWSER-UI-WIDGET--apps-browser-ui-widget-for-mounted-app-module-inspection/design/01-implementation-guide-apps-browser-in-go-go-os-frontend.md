---
Title: 'Implementation guide: Apps Browser in go-go-os frontend'
Ticket: GEPA-11-APPS-BROWSER-UI-WIDGET
Status: active
Topics:
    - frontend
    - ui
    - backend
    - architecture
    - go-go-os
    - wesen-os
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../go-go-os/packages/engine/src/app/createAppStore.ts
      Note: createAppStore factory — apps-browser store must follow this pattern
    - Path: ../../../../../../../go-go-os/packages/engine/src/app/generateCardStories.tsx
      Note: createStoryHelpers — pattern for full-app Storybook stories
    - Path: ../../../../../../../go-go-os/packages/engine/src/desktop/core/state/windowingSlice.ts
      Note: openWindow/closeWindow actions used by module launcher
    - Path: ../../../../../../../go-go-os/packages/engine/src/components/shell/windowing/desktopContributions.ts
      Note: DesktopContribution type and merging logic
    - Path: ../../../../../../../go-go-os/packages/engine/src/components/widgets/ContextMenu.tsx
      Note: Context menu component pattern
    - Path: ../../../../../../../go-go-os/packages/engine/src/components/widgets/DataTable.tsx
      Note: DataTable widget for API table in Module Browser
    - Path: ../../../../../../../go-go-os/packages/engine/src/components/widgets/Chip.tsx
      Note: Chip widget for capability/schema badges
    - Path: ../../../../../../../go-go-os/packages/engine/src/components/widgets/TabControl.tsx
      Note: TabControl widget for potential tab-based layouts
    - Path: ../../../../../../../go-go-os/packages/engine/src/theme/desktop/tokens.css
      Note: CSS token system — all new styles must use --hc-* variables
    - Path: ../../../../../../../go-go-os/packages/engine/src/parts.ts
      Note: data-part constants for CSS targeting
    - Path: ../../../../../../../go-go-os/packages/desktop-os/src/contracts/launcherHostContext.ts
      Note: LauncherHostContext interface (dispatch, openWindow, resolveApiBase)
    - Path: ../../../../../../../go-go-os/packages/desktop-os/src/contracts/launchableAppModule.ts
      Note: LaunchableAppModule interface — the module registration contract
    - Path: ../../../../../../../go-go-app-inventory/apps/inventory/src/launcher/module.tsx
      Note: Reference implementation of LaunchableAppModule (inventory)
    - Path: ../../../../../../../go-go-app-inventory/apps/inventory/src/launcher/renderInventoryApp.tsx
      Note: Contribution creation pattern (menus, commands, adapters)
    - Path: ../../../../../../../go-go-os/apps/crm/src/launcher/module.tsx
      Note: CRM module — simpler LaunchableAppModule reference
    - Path: ../../../../../../../go-go-os/apps/crm/src/app/store.ts
      Note: CRM store — createAppStore usage pattern
    - Path: ../../../../../../../go-go-os/apps/crm/src/app/stories/CrmApp.stories.tsx
      Note: Full-app Storybook story pattern
    - Path: ../../../../../../../go-go-os/packages/engine/src/components/widgets/Btn.stories.tsx
      Note: Widget-level Storybook story pattern
    - Path: ../../../../../../../wesen-os/apps/os-launcher/src/app/modules.tsx
      Note: Module registration array — where apps-browser gets added
    - Path: ../../../../../../../wesen-os/apps/os-launcher/src/App.tsx
      Note: App composition showing hostContext creation
    - Path: ../../../../../../../go-go-os/.storybook/main.ts
      Note: Storybook config — needs new story glob for apps-browser
    - Path: ../../../../../../../pinocchio/cmd/web-chat/web/src/debug-ui/api/debugApi.ts
      Note: RTK Query createApi reference implementation (debugApi)
    - Path: ../../../../../../../pinocchio/cmd/web-chat/web/src/store/profileApi.ts
      Note: RTK Query createApi reference implementation (profileApi)
    - Path: ../../../../../../../pinocchio/cmd/web-chat/web/src/debug-ui/mocks/msw/createDebugHandlers.ts
      Note: MSW handler factory pattern for Storybook mock API
    - Path: ../../../../../../../pinocchio/cmd/web-chat/web/src/debug-ui/mocks/msw/defaultHandlers.ts
      Note: Default MSW handler wiring with fixture data
    - Path: ../../../../../../../pinocchio/cmd/web-chat/web/.storybook/preview.tsx
      Note: Storybook preview with MSW initialize() + mswLoader
    - Path: ../design-doc/01-apps-browser-ux-and-technical-reference.md
      Note: Design doc with endpoint contracts and data models
    - Path: ../sources/ui.md
      Note: UI wireframes and YAML widget DSL
ExternalSources: []
Summary: Step-by-step implementation guide for building the Apps Browser widget in the go-go-os frontend, covering Redux state, Storybook stories, component architecture, and module registration.
LastUpdated: 2026-02-27T20:10:00-05:00
WhatFor: Give an implementation engineer a concrete, file-by-file build plan that follows existing go-go-os patterns for Redux, Storybook, theming, and window management.
WhenToUse: Use when starting implementation of GEPA-11. Follow phases in order; each phase produces working Storybook stories before the next phase begins.
---


# Implementation Guide: Apps Browser in go-go-os Frontend

## Overview

This guide translates the GEPA-11 design doc and UI wireframes into a concrete implementation plan within the go-go-os frontend architecture. It follows existing patterns from the CRM and Inventory modules: Redux Toolkit slices, `data-part` CSS theming, `LaunchableAppModule` registration, and Storybook-first development.

The Apps Browser is three windows backed by one shared Redux slice:

1. **Mounted Apps** (icon folder) — default entry point
2. **Module Browser** (Smalltalk inspector) — three-column drill-down
3. **Health Dashboard** — operational overview

All three windows read from `GET /api/os/apps` and optionally `GET /api/os/apps/{app_id}/reflection`.

**Where this code lives:** `go-go-os/apps/apps-browser/` — inside the go-go-os monorepo, not a separate repo. See **Part 1a: Package Decision** below.


---

## Part 1a: Package Decision — Where Does Apps Browser Live?

### Options considered

| Option | Location | Storybook | Import from wesen-os |
|---|---|---|---|
| **A. Inside go-go-os monorepo** | `go-go-os/apps/apps-browser/` | Runs in go-go-os Storybook (shared with CRM, Todo, engine) | Vite alias or workspace link |
| B. Separate repo (like go-go-app-inventory) | `go-go-app-apps-browser/apps/apps-browser/` | Needs own Storybook config | npm workspace link |
| C. Inside wesen-os | `wesen-os/apps/os-launcher/src/features/apps-browser/` | Runs in wesen-os Storybook (if one exists) | Direct import |

### Recommendation: Option A — inside go-go-os

**Rationale:**

1. **Apps Browser is an OS-level concern, not an application module.** It inspects the module registry itself — it is infrastructure UI, not domain UI like Inventory or GEPA. It belongs with the OS primitives.

2. **It depends heavily on engine internals.** The DesktopShell windowing system, context menus, data-part tokens, `createAppStore` — all live in `@hypercard/engine`. Co-location means zero version-pinning friction.

3. **The CRM and Todo apps already live here.** `go-go-os/apps/crm/` and `go-go-os/apps/todo/` are the established pattern for apps that ship with the OS. Apps Browser is the same kind of thing.

4. **Storybook is already configured.** Adding one glob entry to `.storybook/main.ts` is all that's needed. A separate repo would need its own Storybook toolchain.

5. **go-go-app-inventory is separate because it has Go backend code.** Its separation is driven by the Go module boundary, not a frontend preference. Apps Browser has no Go code — it is pure frontend consuming existing `go-go-os` endpoints.

**If this decision changes later** (e.g., to support third-party OS shell plugins), the module.tsx / LaunchableAppModule contract already provides a clean extraction seam. The code can be moved to a separate repo at that point with minimal refactoring.

### Package.json entry

The go-go-os monorepo uses npm workspaces (`"workspaces": ["packages/*", "apps/*"]`), so `apps/apps-browser/package.json` is automatically discovered:

```json
{
  "name": "@hypercard/apps-browser",
  "version": "0.1.0",
  "private": true,
  "type": "module",
  "exports": {
    ".": "./src/index.ts",
    "./launcher": "./src/launcher/public.ts"
  },
  "main": "./src/index.ts",
  "types": "./src/index.ts",
  "dependencies": {
    "@reduxjs/toolkit": "^2.5.0",
    "react": "^19.0.0",
    "react-dom": "^19.0.0",
    "react-redux": "^9.2.0"
  },
  "devDependencies": {
    "msw": "^2.4.9",
    "msw-storybook-addon": "^2.0.4"
  }
}
```


---

## Part 1b: File Layout

```
go-go-os/apps/apps-browser/
  package.json
  src/
    domain/
      types.ts                     # TypeScript interfaces for API responses
      selectors.ts                 # derived selectors (crossRefSchemas, summaryStats, etc.)
      sorting.ts                   # sort helpers (unhealthy-first, required-first, name)

    api/
      appsApi.ts                   # RTK Query createApi — apps manifest + reflection
      appsApi.test.ts              # API slice tests (optional, MSW covers behavior)

    features/
      appsBrowser/
        appsBrowserSlice.ts        # Redux slice: selection state only (no fetch logic)
        appsBrowserSlice.test.ts   # Slice unit tests

    mocks/
      msw/
        createAppsHandlers.ts      # MSW handler factory (like pinocchio's createDebugHandlers)
        defaultHandlers.ts         # Default wiring with fixture data
      fixtures/
        apps.ts                    # Mock AppManifestDocument[] data
        reflection.ts              # Mock ModuleReflectionDocument data
      browser.ts                   # setupWorker(...handlers) for Storybook

    components/
      AppIcon.tsx                  # Icon with health dot, required diamond, reflection star
      AppIcon.css
      AppIcon.stories.tsx

      AppsFolderWindow.tsx         # Icon grid window
      AppsFolderWindow.css
      AppsFolderWindow.stories.tsx

      ModuleBrowserWindow.tsx      # Three-column inspector + detail panel
      ModuleBrowserWindow.css
      ModuleBrowserWindow.stories.tsx

      HealthDashboardWindow.tsx    # Summary cards + module list
      HealthDashboardWindow.css
      HealthDashboardWindow.stories.tsx

      GetInfoWindow.tsx            # Per-module info inspector
      GetInfoWindow.css
      GetInfoWindow.stories.tsx

      BrowserColumns.tsx           # Top half of Module Browser
      BrowserDetailPanel.tsx       # Bottom half of Module Browser

      SummaryCards.tsx             # Three stat boxes for Health Dashboard
      HealthModuleRow.tsx          # Row in Health Dashboard with expandable error

    launcher/
      module.tsx                   # LaunchableAppModule definition
      public.ts                    # Public exports for launcher registration

    app/
      store.ts                     # createAppStore with RTK Query middleware
      stories/
        AppsBrowserApp.stories.tsx # Full-app story with all windows

    index.ts                       # Public exports
```

### Storybook config update

Add to `go-go-os/.storybook/main.ts` `stories` array:

```typescript
{ directory: '../apps/apps-browser/src', files: '**/*.stories.@(ts|tsx)' },
```

Update `go-go-os/.storybook/preview.ts` to add MSW initialization (see Part 4b).


---

## Part 2: Domain Types (`domain/types.ts`)

These interfaces match the live API payloads documented in the design doc.

```typescript
// --- API response types ---

export interface AppsManifestResponse {
  apps: AppManifestDocument[];
}

export interface AppManifestDocument {
  app_id: string;
  name: string;
  description?: string;
  required: boolean;
  capabilities?: string[];
  reflection?: {
    available: boolean;
    url?: string;
    version?: string;
  };
  healthy: boolean;
  health_error?: string;
}

export interface ModuleReflectionDocument {
  app_id: string;
  name: string;
  version?: string;
  summary?: string;
  capabilities?: ReflectionCapability[];
  docs?: ReflectionDocLink[];
  apis?: ReflectionAPI[];
  schemas?: ReflectionSchemaRef[];
}

export interface ReflectionCapability {
  id: string;
  stability?: string;
  description?: string;
}

export interface ReflectionDocLink {
  id: string;
  title: string;
  url?: string;
}

export interface ReflectionAPI {
  id: string;
  method: string;
  path: string;
  summary?: string;
  request_schema?: string;
  response_schema?: string;
  error_schema?: string;
  tags?: string[];
}

export interface ReflectionSchemaRef {
  id: string;
  format: string;
  uri?: string;
  embedded?: unknown;
}

// --- Redux state types ---

export type LoadState = 'idle' | 'loading' | 'loaded' | 'error';

export type ReflectionState =
  | { kind: 'idle' }
  | { kind: 'loading'; appId: string }
  | { kind: 'loaded'; appId: string; doc: ModuleReflectionDocument }
  | { kind: 'unsupported'; appId: string; reason: string }
  | { kind: 'error'; appId: string; message: string };

export interface AppsBrowserState {
  loadState: LoadState;
  apps: AppManifestDocument[];
  fetchError?: string;

  selectedAppId?: string;
  selectedApiId?: string;
  selectedSchemaId?: string;

  reflectionCache: Record<string, ReflectionState>;

  lastRefreshedAt?: string;
}
```

**Storybook checkpoint:** After writing `types.ts`, no story yet but this is the foundation. Move to the slice.


---

## Part 3: RTK Query API Slice (`api/appsApi.ts`)

Instead of manual `createAsyncThunk` + fetch wrappers, use **RTK Query** (`createApi` + `fetchBaseQuery`) following the pattern established in pinocchio's `debugApi.ts` and `profileApi.ts`. RTK Query gives us:

- Automatic caching, deduplication, and refetching
- Loading/error/success states without manual state management
- Generated hooks (`useGetAppsQuery`, `useGetReflectionQuery`)
- Tag-based cache invalidation for refresh
- Storybook-compatible via MSW (no mock state seeding needed)

### Reference implementations

The two RTK Query APIs already in the codebase:

| API | File | Pattern |
|---|---|---|
| `profileApi` | `pinocchio/cmd/web-chat/web/src/store/profileApi.ts` | Simple CRUD, `tagTypes: ['Profile']`, dynamic `baseUrl` via `basePrefixFromLocation()` |
| `debugApi` | `pinocchio/cmd/web-chat/web/src/debug-ui/api/debugApi.ts` | Read-heavy, many endpoints, `transformResponse` for data normalization, query params |

Apps Browser follows `debugApi` more closely (read-heavy, multiple endpoints, response transformation).

### Implementation

```typescript
// api/appsApi.ts

import type { BaseQueryFn, FetchArgs, FetchBaseQueryError } from '@reduxjs/toolkit/query';
import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type {
  AppManifestDocument,
  AppsManifestResponse,
  ModuleReflectionDocument,
} from '../domain/types';

// Base query that resolves URL at call time (supports both
// Storybook MSW interception and runtime proxy).
const rawBaseQuery = fetchBaseQuery({ baseUrl: '' });

const baseQuery: BaseQueryFn<string | FetchArgs, unknown, FetchBaseQueryError> = async (
  args,
  api,
  extraOptions
) => {
  // In wesen-os runtime, /api/os/apps is served by the Go backend directly.
  // No prefix rewriting needed. MSW intercepts these same URLs in Storybook.
  return rawBaseQuery(args, api, extraOptions);
};

export const appsApi = createApi({
  reducerPath: 'appsApi',
  baseQuery,
  tagTypes: ['AppsList', 'Reflection'],
  endpoints: (builder) => ({

    // GET /api/os/apps
    getApps: builder.query<AppManifestDocument[], void>({
      query: () => '/api/os/apps',
      providesTags: ['AppsList'],
      transformResponse: (response: AppsManifestResponse) => response.apps ?? [],
    }),

    // GET /api/os/apps/{appId}/reflection
    // Returns the reflection doc, or a typed error for 501/404.
    getReflection: builder.query<ModuleReflectionDocument, string>({
      query: (appId) => ({
        url: `/api/os/apps/${appId}/reflection`,
        validateStatus: (response) => {
          // Treat 501 as a "successful" query that returns a specific shape,
          // rather than letting RTK Query treat it as a generic error.
          return response.status === 200 || response.status === 501;
        },
      }),
      providesTags: (result, error, appId) => [{ type: 'Reflection', id: appId }],
      transformResponse: (response: ModuleReflectionDocument | string, meta) => {
        if (meta?.response?.status === 501) {
          // Return a sentinel object so components can distinguish
          // "unsupported" from "error" without checking error state.
          return { app_id: '', name: '', _unsupported: true } as ModuleReflectionDocument & { _unsupported: true };
        }
        return response as ModuleReflectionDocument;
      },
    }),

  }),
});

export const {
  useGetAppsQuery,
  useGetReflectionQuery,
} = appsApi;
```

### Key design choices

1. **`validateStatus` for 501:** The reflection endpoint returns `501 Not Implemented` for non-reflective modules. This is not an error — it is a normal expected state. By handling it in `validateStatus` + `transformResponse`, we avoid RTK Query's error path and keep the component logic clean.

2. **Empty `baseUrl`:** All URLs are absolute paths (`/api/os/apps`). MSW intercepts them identically in Storybook and the real backend serves them in production. No prefix logic needed (unlike pinocchio which needs `basePrefixFromLocation()`).

3. **Tag-based refresh:** `providesTags: ['AppsList']` means calling `dispatch(appsApi.util.invalidateTags(['AppsList']))` triggers a re-fetch — this powers the Refresh button.

4. **Lazy reflection loading:** `useGetReflectionQuery(appId, { skip: !appId })` with the `skip` option means reflection is only fetched when a module is actually selected and has `reflection.available`.

### Usage in components

```typescript
// In any component:
const { data: apps, isLoading, isError, error, refetch } = useGetAppsQuery();

// Lazy reflection fetch — only when module is selected:
const { data: reflection, isLoading: reflLoading } = useGetReflectionQuery(selectedAppId!, {
  skip: !selectedAppId || !selectedApp?.reflection?.available,
});

// Refresh button:
const handleRefresh = () => {
  dispatch(appsApi.util.invalidateTags(['AppsList']));
};
```


---

## Part 3b: Selection State Slice (`features/appsBrowser/appsBrowserSlice.ts`)

With RTK Query handling all fetch state, the Redux slice becomes minimal — it only manages **UI selection state** that is shared across the three windows.

```typescript
import { createSlice, type PayloadAction } from '@reduxjs/toolkit';

interface AppsBrowserUIState {
  selectedAppId?: string;
  selectedApiId?: string;
  selectedSchemaId?: string;
}

const initialState: AppsBrowserUIState = {
  selectedAppId: undefined,
  selectedApiId: undefined,
  selectedSchemaId: undefined,
};

export const appsBrowserSlice = createSlice({
  name: 'appsBrowser',
  initialState,
  reducers: {
    selectModule(state, action: PayloadAction<string | undefined>) {
      state.selectedAppId = action.payload;
      state.selectedApiId = undefined;
      state.selectedSchemaId = undefined;
    },
    selectApi(state, action: PayloadAction<string | undefined>) {
      state.selectedApiId = action.payload;
      state.selectedSchemaId = undefined;
    },
    selectSchema(state, action: PayloadAction<string | undefined>) {
      state.selectedSchemaId = action.payload;
    },
    clearSelection(state) {
      state.selectedAppId = undefined;
      state.selectedApiId = undefined;
      state.selectedSchemaId = undefined;
    },
  },
});

export const appsBrowserReducer = appsBrowserSlice.reducer;
export const { selectModule, selectApi, selectSchema, clearSelection } = appsBrowserSlice.actions;
```

### Store creation (`app/store.ts`)

The store must include both the RTK Query API reducer and middleware alongside the UI slice:

```typescript
import { createAppStore } from '@hypercard/engine';
import { appsBrowserReducer } from '../features/appsBrowser/appsBrowserSlice';
import { appsApi } from '../api/appsApi';

// createAppStore uses configureStore internally. We need to extend it
// to include RTK Query's reducer and middleware.
//
// OPTION A: If createAppStore is extended to accept apiSlices parameter:
//
//   export const { store, createStore } = createAppStore(
//     { appsBrowser: appsBrowserReducer },
//     { apiSlices: [appsApi] }
//   );
//
// OPTION B: If createAppStore is NOT extended, use configureStore directly
// (matching pinocchio's store.ts pattern):

import { configureStore } from '@reduxjs/toolkit';
import { windowingReducer } from '@hypercard/engine/desktop-core';
import { pluginCardRuntimeReducer, notificationsReducer, debugReducer, hypercardArtifactsReducer } from '@hypercard/engine';

function createStore() {
  return configureStore({
    reducer: {
      // Engine built-ins (same as createAppStore provides)
      pluginCardRuntime: pluginCardRuntimeReducer,
      windowing: windowingReducer,
      notifications: notificationsReducer,
      debug: debugReducer,
      hypercardArtifacts: hypercardArtifactsReducer,
      // Domain
      appsBrowser: appsBrowserReducer,
      // RTK Query
      [appsApi.reducerPath]: appsApi.reducer,
    },
    middleware: (getDefaultMiddleware) =>
      getDefaultMiddleware().concat(appsApi.middleware),
  });
}

export const store = createStore();
export { createStore as createAppsBrowserStore };

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
```

**Recommendation:** Extend `createAppStore` to accept an optional `apiSlices` parameter so apps-browser (and future modules) can use RTK Query without duplicating the engine reducer list. This is a small change to `packages/engine/src/app/createAppStore.ts`:

```typescript
// Proposed extension to createAppStore:
export function createAppStore<T extends Record<string, Reducer>>(
  domainReducers: T,
  options: CreateAppStoreOptions & {
    apiSlices?: Array<{ reducerPath: string; reducer: Reducer; middleware: Middleware }>;
  } = {},
) {
  const apiReducers = Object.fromEntries(
    (options.apiSlices ?? []).map(api => [api.reducerPath, api.reducer])
  );
  const apiMiddlewares = (options.apiSlices ?? []).map(api => api.middleware);

  const reducer = {
    ...builtInReducers,
    ...domainReducers,
    ...apiReducers,
  };

  function createStore() {
    return configureStore({
      reducer,
      middleware: (getDefault) =>
        getDefault().concat(...apiMiddlewares, ...otherMiddlewares),
    });
  }
  // ...
}
```

This keeps the pattern clean and avoids Option B's manual reducer duplication.

### Selectors (`domain/selectors.ts`)

These selectors combine RTK Query cache data with UI selection state:

| Selector | Returns |
|---|---|
| `selectSelectedAppId` | `state.appsBrowser.selectedAppId` |
| `selectSelectedApiId` | `state.appsBrowser.selectedApiId` |
| `selectSelectedSchemaId` | `state.appsBrowser.selectedSchemaId` |
| `selectSelectedApp(apps, selectedAppId)` | The full `AppManifestDocument` for selected ID |
| `selectSelectedApi(reflection, selectedApiId)` | The `ReflectionAPI` for selected API ID |
| `selectSelectedSchema(reflection, selectedSchemaId)` | The `ReflectionSchemaRef` for selected schema ID |
| `selectCrossRefSchemaIds(reflection, selectedApiId)` | Set of schema IDs referenced by selected API's `request_schema`/`response_schema` |
| `selectSortedApps(apps)` | Apps sorted: unhealthy first, required first, then name asc |
| `selectSummaryStats(apps)` | `{ mounted, healthy, unhealthy, required, reflective }` counts |
| `selectHasUnhealthyRequired(apps)` | Boolean: any app where `required && !healthy` |

Note: Most selectors take RTK Query hook data as input (e.g., `apps` from `useGetAppsQuery().data`) rather than reading from Redux state directly. This is the idiomatic RTK Query pattern — the cache is the source of truth for server data, the slice is the source of truth for UI state.


---

## Part 4: MSW Handlers for Storybook (`mocks/`)

MSW (Mock Service Worker) intercepts fetch requests at the network level, so RTK Query works identically in Storybook and production — no mock state seeding, no conditional fetch logic. This follows the pattern from `pinocchio/cmd/web-chat/web/src/debug-ui/mocks/`.

### 4a. Handler factory (`mocks/msw/createAppsHandlers.ts`)

```typescript
import { HttpResponse, http } from 'msw';
import type { AppManifestDocument, ModuleReflectionDocument } from '../../domain/types';

export interface AppsHandlerData {
  apps: AppManifestDocument[];
  reflections: Record<string, ModuleReflectionDocument>;
  // app_ids that should return 501 (non-reflective)
  unsupportedReflection: string[];
}

export interface CreateAppsHandlersOptions {
  data: AppsHandlerData;
  delayMs?: number;
}

export function createAppsHandlers(options: CreateAppsHandlersOptions) {
  const { data, delayMs = 0 } = options;

  return [
    // GET /api/os/apps
    http.get('/api/os/apps', async () => {
      if (delayMs > 0) await delay(delayMs);
      return HttpResponse.json({ apps: data.apps });
    }),

    // GET /api/os/apps/:appId/reflection
    http.get('/api/os/apps/:appId/reflection', async ({ params }) => {
      if (delayMs > 0) await delay(delayMs);
      const appId = String(params.appId);

      // Unknown app → 404
      if (!data.apps.some(a => a.app_id === appId)) {
        return HttpResponse.text('app not found', { status: 404 });
      }

      // Non-reflective → 501
      if (data.unsupportedReflection.includes(appId)) {
        return HttpResponse.text('reflection not implemented', { status: 501 });
      }

      // Reflective → 200
      const doc = data.reflections[appId];
      if (!doc) {
        return HttpResponse.text('reflection not implemented', { status: 501 });
      }
      return HttpResponse.json(doc);
    }),
  ];
}

function delay(ms: number) {
  return new Promise(resolve => setTimeout(resolve, ms));
}
```

### 4b. Default handlers + fixtures (`mocks/msw/defaultHandlers.ts`)

```typescript
import { MOCK_APPS, MOCK_GEPA_REFLECTION } from '../fixtures/apps';
import { createAppsHandlers, type AppsHandlerData } from './createAppsHandlers';

export const defaultAppsHandlerData: AppsHandlerData = {
  apps: MOCK_APPS,
  reflections: { gepa: MOCK_GEPA_REFLECTION },
  unsupportedReflection: ['inventory'],
};

export function createDefaultAppsHandlers(
  overrides: Partial<AppsHandlerData> = {},
  options: { delayMs?: number } = {},
) {
  return createAppsHandlers({
    data: { ...defaultAppsHandlerData, ...overrides },
    delayMs: options.delayMs,
  });
}

export const defaultHandlers = createDefaultAppsHandlers();
```

### 4c. Storybook integration

**Option 1 (per-story handlers via `msw-storybook-addon`):**

```typescript
// In a story file:
import { createDefaultAppsHandlers } from '../../mocks/msw/defaultHandlers';

export const Default: Story = {
  parameters: {
    msw: {
      handlers: createDefaultAppsHandlers(),
    },
  },
};

export const WithUnhealthy: Story = {
  parameters: {
    msw: {
      handlers: createDefaultAppsHandlers({
        apps: [MOCK_INVENTORY_UNHEALTHY, MOCK_GEPA],
      }),
    },
  },
};

export const Loading: Story = {
  parameters: {
    msw: {
      handlers: createDefaultAppsHandlers({}, { delayMs: 999999 }),
    },
  },
};
```

**Option 2 (global default + per-story overrides):**

Update `go-go-os/.storybook/preview.ts` to initialize MSW globally:

```typescript
import { initialize, mswLoader } from 'msw-storybook-addon';

// Initialize MSW for all stories (no-op for stories that don't use it)
initialize();

const preview: Preview = {
  loaders: [mswLoader],
  decorators: [
    (Story) => React.createElement(HyperCardTheme, null, React.createElement(Story)),
  ],
  // ...
};
```

This mirrors pinocchio's `.storybook/preview.tsx` pattern exactly. Stories that don't specify `parameters.msw` are unaffected. Stories that do get real network-level mocking.

**Why MSW over seeded Redux state:**

| Approach | Pros | Cons |
|---|---|---|
| Seeded Redux state | No network setup, immediate | Bypasses RTK Query entirely; doesn't test real fetch/cache/error paths |
| MSW handlers | Tests real RTK Query flow; loading/error states happen naturally; same code path as production | Requires msw + msw-storybook-addon deps |

MSW is the right choice because the whole point of RTK Query is to manage server cache state — mocking at the Redux level defeats that purpose.


---

## Part 5: Components — Phase 1 (AppIcon + Apps Folder)

### 5a. AppIcon Component

**Purpose:** Renders one module icon with badge composition for the icon folder view.

**Props:**

```typescript
interface AppIconProps {
  app: AppManifestDocument;
  selected?: boolean;
  onDoubleClick?: () => void;
  onContextMenu?: (e: React.MouseEvent) => void;
}
```

**Anatomy (from UI sketch):**

```
┌──────┐
│ ◈ ●  │   top-left: required diamond (◈), top-right: health dot (● or ○)
│ ┌──┐ │
│ │▦▦│ │   center: module glyph (▦▦ normal, ░░ unhealthy)
│ └──┘ │
│  ★   │   bottom-center: reflection star (★) if available
└──────┘
 Label      app.name, prefixed with "warning " if unhealthy
```

**CSS parts:**

```css
[data-part="app-icon"]              { /* icon container */ }
[data-part="app-icon-badge-health"] { /* ● or ○ dot */ }
[data-part="app-icon-badge-required"] { /* ◈ diamond */ }
[data-part="app-icon-badge-reflection"] { /* ★ star */ }
[data-part="app-icon-glyph"]        { /* center module image */ }
[data-part="app-icon-label"]        { /* text label below icon */ }

[data-part="app-icon"][data-state="selected"]  { background: var(--hc-icon-selected-bg); }
[data-part="app-icon"][data-state="unhealthy"] { /* dimmed glyph, warning prefix */ }
```

**Stories (`AppIcon.stories.tsx`):**

| Story | Description |
|---|---|
| `HealthyOptional` | Basic healthy, non-required, non-reflective |
| `HealthyRequired` | Healthy + required diamond |
| `HealthyReflective` | Healthy + reflection star |
| `HealthyRequiredReflective` | All badges |
| `UnhealthyRequired` | Red dot, dimmed glyph, warning prefix |
| `Selected` | Highlighted selection state |

Story structure (AppIcon is a pure presentational component — no MSW needed, just props):

```typescript
const meta = {
  title: 'Apps/AppsBrowser/AppIcon',
  component: AppIcon,
  tags: ['autodocs'],
  decorators: [(Story) => <HyperCardTheme><Story /></HyperCardTheme>],
} satisfies Meta<typeof AppIcon>;
```

### 5b. AppsFolderWindow Component

**Purpose:** Icon grid with toolbar and status summary.

**Layout:**

```
┌─ Mounted Apps ─────────────────────── ○ □ ┐
│  [back/fwd]  {n} apps · {n} healthy   [⟳]  │
├────────────────────────────────────────────┤
│                                            │
│   [AppIcon]    [AppIcon]    [AppIcon]      │
│   [AppIcon]    [AppIcon]                   │
│                                            │
└────────────────────────────────────────────┘
```

**Props:**

```typescript
interface AppsFolderWindowProps {
  // All state comes from Redux via useSelector
}
```

**Internal structure:**

```
AppsFolderWindow
  ├── Toolbar
  │   ├── BackForwardButtons (disabled phase 1)
  │   ├── StatusSummary (computed from selectSummaryStats)
  │   └── RefreshButton (dispatches fetchApps thunk)
  ├── IconGrid
  │   ├── for each app in selectSortedApps:
  │   │   └── AppIcon
  │   └── (loading skeleton if loadState === 'loading')
  └── (error retry panel if loadState === 'error')
```

**CSS parts:**

```css
[data-part="apps-folder"]            { /* window body */ }
[data-part="apps-folder-toolbar"]    { /* toolbar bar */ }
[data-part="apps-folder-status"]     { /* status summary text */ }
[data-part="apps-folder-grid"]       { /* icon grid container, CSS grid */ }
[data-part="apps-folder-empty"]      { /* loading/error state */ }
```

**Stories (`AppsFolderWindow.stories.tsx`):**

| Story | Description |
|---|---|
| `Default` | 2 apps (Inventory + GEPA), both healthy |
| `WithUnhealthy` | Inventory unhealthy, GEPA healthy |
| `Loading` | Skeleton loading state |
| `Error` | Error state with retry button |
| `ManyModules` | 6+ modules to test grid layout |

Each story uses MSW handlers to control what the API returns. RTK Query fetches on mount, so loading/error/success states happen naturally:

```typescript
import { createDefaultAppsHandlers } from '../../mocks/msw/defaultHandlers';
import { MOCK_INVENTORY_UNHEALTHY, MOCK_GEPA } from '../../mocks/fixtures/apps';

const meta = {
  title: 'Apps/AppsBrowser/AppsFolderWindow',
  component: AppsFolderWindow,
  decorators: [storeDecorator],
  parameters: {
    layout: 'fullscreen',
    msw: { handlers: createDefaultAppsHandlers() },
  },
} satisfies Meta<typeof AppsFolderWindow>;

export const Default: Story = {};

export const WithUnhealthy: Story = {
  parameters: {
    msw: {
      handlers: createDefaultAppsHandlers({
        apps: [MOCK_INVENTORY_UNHEALTHY, MOCK_GEPA],
      }),
    },
  },
};

export const Loading: Story = {
  parameters: {
    msw: {
      handlers: createDefaultAppsHandlers({}, { delayMs: 999999 }),
    },
  },
};

export const Error: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('/api/os/apps', () => HttpResponse.json({}, { status: 500 })),
      ],
    },
  },
};
```

**Storybook checkpoint:** After this phase, you can see the icon folder with all badge combinations and loading/error states in Storybook. MSW handles all API behavior — no Redux seeding required.


---

## Part 6: Components — Phase 2 (Module Browser)

### 6a. BrowserColumns Component

**Purpose:** The top half of the Module Browser — three resizable column panes.

**Layout:**

```
┌──────────────┬──────────────┬──────────────────┐
│ Modules      │ APIs         │ Schemas          │
│──────────────│──────────────│──────────────────│
│              │              │                  │
│ ● Inventory◈│ GET /scripts │ run-request      │
│«GEPA»     ★ │ POST /run    │ ▸ run-response   │
│              │ GET /events  │ script-def       │
│              │              │                  │
└──────────────┴──────────────┴──────────────────┘
```

**Sub-components:**

**`ModuleListPane`** — left column
- Shows all apps with health dot, required/reflection badges
- Selection highlight with `data-state="selected"` on the `«guillemot»` style
- Dispatches `selectModule` on click
- On module select, if `reflection.available && !cached`, dispatches `fetchReflection`

**`APIListPane`** — middle column
- Shows `reflection.apis[]` as `{method} {path}` rows, monospace
- Empty states: no module selected (blank), reflection unavailable (informational message), loading (skeleton)
- Selection dispatches `selectApi`

**`SchemaListPane`** — right column
- Shows `reflection.schemas[]` as `{schema.id}` rows
- Cross-reference marker `▸` on schemas referenced by selected API
- Uses `selectCrossRefSchemaIds` selector for highlighting
- Selection dispatches `selectSchema`

**CSS parts:**

```css
[data-part="browser-columns"]        { display: flex; /* horizontal split */ }
[data-part="browser-pane"]           { /* single column */ }
[data-part="browser-pane-header"]    { /* column header (Modules, APIs, Schemas) */ }
[data-part="browser-pane-list"]      { /* scrollable list area */ }
[data-part="browser-pane-item"]      { /* single row */ }
[data-part="browser-pane-item"][data-state="selected"] { /* selection highlight */ }
[data-part="browser-pane-item"][data-state="cross-ref"] { /* ▸ marker highlight */ }
[data-part="browser-pane-empty"]     { /* empty/placeholder state */ }
```

### 6b. BrowserDetailPanel Component

**Purpose:** The bottom half — shows detail for the deepest selected item.

**States (from UI sketch):**

| Selection depth | Detail panel shows |
|---|---|
| Nothing selected | "Select a module to inspect." |
| Module selected | Module summary: id, description, required, base URL, reflection status |
| API selected | API detail: full path, summary, tags, request/response/error schema links |
| Schema selected | Schema detail: id, format, URI, embedded JSON (code block) |

**Implementation:** A single component that switches on `selectedSchemaId > selectedApiId > selectedAppId > nothing`:

```typescript
function BrowserDetailPanel() {
  const schema = useSelector(selectSelectedSchema);
  const api = useSelector(selectSelectedApi);
  const app = useSelector(selectSelectedApp);

  if (schema) return <SchemaDetail schema={schema} />;
  if (api) return <APIDetail api={api} appId={app?.app_id} />;
  if (app) return <ModuleDetail app={app} reflection={...} />;
  return <EmptyDetail message="Select a module to inspect." />;
}
```

Each sub-detail view uses the engine's `FieldRow`-style key-value layout with `data-part="field-grid"`.

### 6c. ModuleBrowserWindow Component

**Purpose:** Composes BrowserColumns + BrowserDetailPanel in a vertical split.

```
ModuleBrowserWindow
  ├── Toolbar (refresh button)
  ├── BrowserColumns (flex: 1, min-height: 50%)
  │   ├── ModuleListPane
  │   ├── APIListPane
  │   └── SchemaListPane
  └── BrowserDetailPanel (flex: 1, min-height: 50%)
```

**CSS parts:**

```css
[data-part="module-browser"]         { display: flex; flex-direction: column; }
[data-part="module-browser-top"]     { flex: 1; min-height: 0; }
[data-part="module-browser-bottom"]  { flex: 1; min-height: 0; border-top: 1px solid var(--hc-color-border); }
```

**Stories (`ModuleBrowserWindow.stories.tsx`):**

| Story | Description |
|---|---|
| `NothingSelected` | All three columns empty, placeholder message |
| `GEPASelected` | GEPA selected, APIs and Schemas populated |
| `GEPAApiSelected` | GEPA + POST /run selected, cross-refs highlighted |
| `GEPASchemaSelected` | Schema detail with JSON code block |
| `InventorySelected` | Inventory selected, "reflection not available" in API/Schema columns |
| `ReflectionLoading` | GEPA selected, skeleton in API/Schema columns |

**Storybook checkpoint:** After this phase, the Module Browser with three-column navigation and detail panel is visible in Storybook with all selection states.


---

## Part 7: Components — Phase 3 (Get Info Window)

### GetInfoWindow Component

**Purpose:** Static info panel for one module — opened from context menu "Get Info".

**Layout (from UI sketch):**

```
┌─ GEPA — Get Info ──────────── ○ □ ┐
│    [large icon]  GEPA              │
│                  gepa              │
│                                    │
│  --- General ──────────────────    │
│  Description:  ...                 │
│  Required:     No                  │
│  Base URL:     /api/apps/gepa/     │
│                                    │
│  --- Health ───────────────────    │
│  Status:  ● Healthy                │
│  Last check: 14:32:07             │
│                                    │
│  --- Reflection ───────────────    │
│  Available: ★ Yes (v1)             │
│  URL: /api/os/apps/gepa/reflection │
│                                    │
│  --- APIs (5) ─────────────────    │
│  GET  /scripts    list local ...   │
│  POST /run        execute a ...    │
│  ...                               │
│                                    │
│  --- Schemas (4) ──────────────    │
│  [run-request] [run-response] ...  │
│                                    │
│  [ Open in Browser ]               │
└────────────────────────────────────┘
```

**Props:**

```typescript
interface GetInfoWindowProps {
  appId: string;  // passed via window content
}
```

**Sections:**

1. **Header** — Large AppIcon + app name + app_id (subtitle)
2. **General** — Description, Required badge, Base URL (monospace, uses existing `FieldRow` pattern)
3. **Health** — Status dot + label, last check timestamp; expandable error block if unhealthy
4. **Reflection** — "Available: Yes (v1)" or informational "not available yet" message
5. **APIs** — Compact list: `METHOD /path  summary` (only if reflection loaded)
6. **Schemas** — Chip row using engine `Chip` component (only if reflection loaded)
7. **Footer** — "Open in Browser" button (dispatches `open_browser_view` action)

**CSS parts:**

```css
[data-part="get-info"]               { /* scrollable content area */ }
[data-part="get-info-header"]        { /* icon + name header */ }
[data-part="get-info-section"]       { /* section container */ }
[data-part="get-info-section-title"] { /* --- Section Name ──── */ }
[data-part="get-info-api-row"]       { /* compact API entry */ }
[data-part="get-info-schema-chips"]  { /* chip row for schemas */ }
[data-part="get-info-footer"]        { /* bottom action bar */ }
```

**Stories (`GetInfoWindow.stories.tsx`):**

| Story | Description |
|---|---|
| `GEPAReflective` | GEPA with full reflection data (APIs, schemas) |
| `InventoryNoReflection` | Inventory with "not available yet" message |
| `UnhealthyModule` | Unhealthy required module with expanded error block |
| `ReflectionLoading` | Reflection still loading, APIs/Schemas show skeleton |

**Storybook checkpoint:** Get Info window with all section combinations visible.


---

## Part 8: Components — Phase 4 (Health Dashboard)

### HealthDashboardWindow Component

**Layout (from UI sketch):**

```
┌─ Health Dashboard ───────────── ○ □ ┐
│                   Last check: ...  [⟳] │
├────────────────────────────────────────┤
│                                        │
│  [⚠ System degraded banner]           │ ← only if required module unhealthy
│                                        │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ │
│  │    2    │ │    2    │ │    1    │ │
│  │ mounted │ │ healthy │ │required │ │
│  └─────────┘ └─────────┘ └─────────┘ │
│                                        │
│  Modules                               │
│  ┌────────────────────────────────┐   │
│  │ ● Inventory  healthy ◈ /api/…  │   │
│  │ ● GEPA       healthy ★ /api/…  │   │
│  └────────────────────────────────┘   │
│                                        │
└────────────────────────────────────────┘
```

**Sub-components:**

**`DegradedBanner`** — Warning banner when `selectHasUnhealthyRequired` is true:

```
⚠ System degraded — {n} required module(s) unhealthy
```

Uses `data-part="degraded-banner"` with warning styling.

**`SummaryCards`** — Three stat boxes:

| Card | Value | Warning state |
|---|---|---|
| Mounted | `apps.length` | Never |
| Healthy | `healthy count` (+ "of {total}" subtitle when degraded) | When `< total` |
| Required | `required count` | When any required is unhealthy |

Uses CSS grid or flexbox, `data-part="summary-card"`.

**`HealthModuleRow`** — Row in module list:

```
● ModuleName    healthy/UNHEALTHY    ◈/★    /api/apps/mod/
  └── [expanded error block when unhealthy]
```

- Auto-expands when `!app.healthy` to show `health_error` in monospace block
- Click dispatches `open_get_info`

**CSS parts:**

```css
[data-part="health-dashboard"]       { /* window body */ }
[data-part="degraded-banner"]        { /* warning banner */ }
[data-part="summary-cards"]          { display: flex; gap: ...; }
[data-part="summary-card"]           { /* stat box */ }
[data-part="summary-card"][data-state="warning"] { /* warning color */ }
[data-part="health-module-list"]     { /* module list container */ }
[data-part="health-module-row"]      { /* single row */ }
[data-part="health-error-block"]     { /* expanded error detail, monospace */ }
```

**Stories (`HealthDashboardWindow.stories.tsx`):**

| Story | Description |
|---|---|
| `AllHealthy` | 2 apps, both healthy, no banner |
| `DegradedRequired` | Inventory unhealthy + required, banner visible |
| `ManyModules` | 6 modules with mixed health for layout testing |
| `Loading` | Loading skeleton |

**Storybook checkpoint:** Health Dashboard with summary cards, degraded banner, and expandable error blocks.


---

## Part 9: Context Menu Integration

### Context menu items

Registered via `createContributions` in the launcher module. Uses the `DesktopContribution.commands` pattern from inventory.

**Context actions for app icons:**

| Action | Command ID | Enabled |
|---|---|---|
| Get Info | `apps-browser.get-info` | Always |
| Open in Browser | `apps-browser.open-browser` | Always |
| Inspect Reflection | `apps-browser.inspect-reflection` | `reflection.available` |
| Copy Base URL | `apps-browser.copy-base-url` | Always |
| Copy Reflection URL | `apps-browser.copy-reflection-url` | `reflection.available` |
| Health Check | `apps-browser.health-check` | Always |

**Implementation pattern:**

Each icon in the `AppsFolderWindow` registers context actions using `useRegisterIconContextActions` or a similar hook. The actions are defined in the contributions:

```typescript
createContributions: (ctx): DesktopContribution[] => [{
  id: 'apps-browser.contributions',
  menus: [
    {
      id: 'file',
      label: 'File',
      merge: 'append',
      items: [
        { id: 'apps-browser.refresh', label: 'Refresh All', commandId: 'apps-browser.refresh', shortcut: 'Cmd+R' },
      ],
    },
    {
      id: 'view',
      label: 'View',
      items: [
        { id: 'apps-browser.view-icons', label: 'as Icons', commandId: 'apps-browser.view-icons', shortcut: 'Cmd+1' },
        { id: 'apps-browser.view-browser', label: 'as Browser', commandId: 'apps-browser.view-browser', shortcut: 'Cmd+2' },
      ],
    },
    {
      id: 'module',
      label: 'Module',
      items: [
        { id: 'apps-browser.get-info', label: 'Get Info', commandId: 'apps-browser.get-info', shortcut: 'Cmd+I' },
        { id: 'apps-browser.inspect-reflection', label: 'Inspect Reflection', commandId: 'apps-browser.inspect-reflection' },
        // ... more items
      ],
    },
    {
      id: 'window',
      label: 'Window',
      items: [
        { id: 'apps-browser.window-folder', label: 'Mounted Apps', commandId: 'apps-browser.open-folder' },
        { id: 'apps-browser.window-browser', label: 'Module Browser', commandId: 'apps-browser.open-module-browser' },
        { id: 'apps-browser.window-health', label: 'Health Dashboard', commandId: 'apps-browser.open-health-dashboard' },
      ],
    },
  ],
  commands: [appsBrowserCommandHandler],
  windowContentAdapters: [createAppsBrowserWindowAdapter()],
}],
```


---

## Part 10: Launcher Module Registration

### `launcher/module.tsx`

Follows the CRM/Inventory `LaunchableAppModule` pattern exactly.

```typescript
import type { LaunchableAppModule } from '@hypercard/desktop-os';
import { createSlice, type PayloadAction } from '@reduxjs/toolkit';
import { appsBrowserReducer } from '../features/appsBrowser/appsBrowserSlice';

const launcherStateSlice = createSlice({
  name: 'appsBrowserLauncher',
  initialState: { launchCount: 0 },
  reducers: {
    markLaunched(state) { state.launchCount += 1; },
  },
});

export const appsBrowserLauncherModule: LaunchableAppModule = {
  manifest: {
    id: 'apps-browser',
    name: 'Apps Browser',
    icon: '🔍',      // or suitable icon
    launch: { mode: 'window' },
    desktop: { order: 50 },
  },

  state: {
    stateKey: 'app_apps_browser',
    reducer: launcherStateSlice.reducer,
  },

  buildLaunchWindow: (ctx, reason) => {
    ctx.dispatch(launcherStateSlice.actions.markLaunched());
    return {
      id: `window:apps-browser:folder:${Date.now()}`,
      title: 'Mounted Apps',
      icon: '🔍',
      bounds: { x: 120, y: 80, w: 640, h: 480 },
      content: { kind: 'app', appKey: 'apps-browser-folder' },
      dedupeKey: reason === 'startup' ? 'apps-browser:folder' : undefined,
    };
  },

  createContributions: (ctx) => [/* see Part 9 */],

  renderWindow: ({ windowId, ctx }) => (
    <AppsBrowserHost key={windowId} windowId={windowId} ctx={ctx} />
  ),
};
```

### Register in wesen-os

In `wesen-os/apps/os-launcher/src/app/modules.tsx`:

```typescript
import { appsBrowserLauncherModule } from '@hypercard/apps-browser/src/launcher/module';

export const launcherModules: LaunchableAppModule[] = [
  inventoryLauncherModule,
  todoLauncherModule,
  crmLauncherModule,
  appsBrowserLauncherModule,  // ← add
];
```


---

## Part 11: Window Content Adapters

The Apps Browser spawns multiple window types. Each uses `content.kind === 'app'` with different `appKey` values:

| Window | `appKey` | Content |
|---|---|---|
| Mounted Apps | `'apps-browser-folder'` | `<AppsFolderWindow />` |
| Module Browser | `'apps-browser-module-browser'` | `<ModuleBrowserWindow />` |
| Health Dashboard | `'apps-browser-health-dashboard'` | `<HealthDashboardWindow />` |
| Get Info | `'apps-browser-get-info'` | `<GetInfoWindow appId={appId} />` |

Window adapter:

```typescript
function createAppsBrowserWindowAdapter(): WindowContentAdapter {
  return {
    id: 'apps-browser.window-adapter',
    canRender: (window) =>
      window.content.kind === 'app' &&
      typeof window.content.appKey === 'string' &&
      window.content.appKey.startsWith('apps-browser-'),
    render: (window) => {
      switch (window.content.appKey) {
        case 'apps-browser-folder':
          return <AppsFolderWindow />;
        case 'apps-browser-module-browser':
          return <ModuleBrowserWindow />;
        case 'apps-browser-health-dashboard':
          return <HealthDashboardWindow />;
        case 'apps-browser-get-info':
          return <GetInfoWindow appId={window.content.appParams?.appId} />;
        default:
          return null;
      }
    },
  };
}
```


---

## Part 12: CSS Theming

All styles use the `data-part` + CSS variable pattern. No inline styles for colors/sizing.

### New CSS variables to define

Add to the apps-browser CSS file (scoped under `[data-widget="hypercard"]`):

```css
/* Health status */
--hc-health-ok-color: #2d8a4e;
--hc-health-error-color: #d1242f;
--hc-health-ok-bg: #dafbe1;
--hc-health-error-bg: #ffebe9;

/* Apps Browser specific */
--hc-apps-browser-icon-size: 64px;
--hc-apps-browser-icon-gap: 32px;
--hc-apps-browser-pane-border: var(--hc-color-border);
--hc-apps-browser-badge-size: 10px;

/* Summary cards */
--hc-summary-card-bg: var(--hc-color-bg);
--hc-summary-card-border: var(--hc-color-border);
--hc-summary-card-warning-bg: var(--hc-health-error-bg);
```

### data-part naming convention

All new parts are prefixed with `app-icon-`, `apps-folder-`, `module-browser-`, `browser-`, `get-info-`, `health-`, `summary-`, or `degraded-` to avoid collision with engine parts.


---

## Part 13: Storybook Story Index

Complete story manifest, organized by development order:

### Phase 1: Foundation
| File | Title | Stories |
|---|---|---|
| `AppIcon.stories.tsx` | `Apps/AppsBrowser/AppIcon` | HealthyOptional, HealthyRequired, HealthyReflective, HealthyRequiredReflective, UnhealthyRequired, Selected |
| `AppsFolderWindow.stories.tsx` | `Apps/AppsBrowser/AppsFolderWindow` | Default, WithUnhealthy, Loading, Error, ManyModules |

### Phase 2: Module Browser
| File | Title | Stories |
|---|---|---|
| `ModuleBrowserWindow.stories.tsx` | `Apps/AppsBrowser/ModuleBrowserWindow` | NothingSelected, GEPASelected, GEPAApiSelected, GEPASchemaSelected, InventorySelected, ReflectionLoading |

### Phase 3: Get Info
| File | Title | Stories |
|---|---|---|
| `GetInfoWindow.stories.tsx` | `Apps/AppsBrowser/GetInfoWindow` | GEPAReflective, InventoryNoReflection, UnhealthyModule, ReflectionLoading |

### Phase 4: Health Dashboard
| File | Title | Stories |
|---|---|---|
| `HealthDashboardWindow.stories.tsx` | `Apps/AppsBrowser/HealthDashboardWindow` | AllHealthy, DegradedRequired, ManyModules, Loading |

### Phase 5: Full App
| File | Title | Stories |
|---|---|---|
| `AppsBrowserApp.stories.tsx` | `Apps/AppsBrowser/FullApp` | Default (folder open), AllWindows (desktop with all three) |


---

## Part 14: Mock Data for Stories

Define mock data constants in `src/domain/__fixtures__/mockData.ts`:

```typescript
export const MOCK_INVENTORY: AppManifestDocument = {
  app_id: 'inventory',
  name: 'Inventory',
  description: 'Inventory chat runtime, profiles, timeline, and confirm APIs',
  required: true,
  capabilities: ['chat', 'ws', 'timeline', 'profiles', 'confirm'],
  healthy: true,
};

export const MOCK_GEPA: AppManifestDocument = {
  app_id: 'gepa',
  name: 'GEPA',
  description: 'GEPA script runner backend module',
  required: false,
  capabilities: ['script-runner', 'events', 'timeline', 'schemas', 'reflection'],
  reflection: { available: true, url: '/api/os/apps/gepa/reflection', version: 'v1' },
  healthy: true,
};

export const MOCK_INVENTORY_UNHEALTHY: AppManifestDocument = {
  ...MOCK_INVENTORY,
  healthy: false,
  health_error: 'database connection pool exhausted: dial tcp 127.0.0.1:5432: connect: connection refused',
};

export const MOCK_GEPA_REFLECTION: ModuleReflectionDocument = {
  app_id: 'gepa',
  name: 'GEPA',
  version: 'v1',
  summary: 'GEPA script runner backend module',
  capabilities: [
    { id: 'script-runner', stability: 'stable', description: 'Execute named scripts' },
    { id: 'events', stability: 'stable', description: 'Server-sent event stream' },
  ],
  apis: [
    { id: 'list-scripts', method: 'GET', path: '/scripts', summary: 'List local scripts' },
    { id: 'run-script', method: 'POST', path: '/run', summary: 'Execute a script', request_schema: 'run-request', response_schema: 'run-response' },
    { id: 'events', method: 'GET', path: '/events', summary: 'Event stream' },
    { id: 'get-schema', method: 'GET', path: '/schemas/{id}', summary: 'Schema doc' },
    { id: 'timeline', method: 'GET', path: '/timeline', summary: 'Timeline entries' },
  ],
  schemas: [
    { id: 'run-request', format: 'json-schema', uri: '/api/apps/gepa/schemas/run-request' },
    { id: 'run-response', format: 'json-schema', uri: '/api/apps/gepa/schemas/run-response' },
    { id: 'script-def', format: 'json-schema', uri: '/api/apps/gepa/schemas/script-def' },
    { id: 'event-entry', format: 'json-schema', uri: '/api/apps/gepa/schemas/event-entry' },
  ],
};

// Additional mocks for "many modules" scenario
export const MOCK_AUTH: AppManifestDocument = { /* ... */ };
export const MOCK_BILLING: AppManifestDocument = { /* ... */ };
export const MOCK_SCHEDULER: AppManifestDocument = { /* ... */ };
export const MOCK_TELEMETRY: AppManifestDocument = { /* ... */ };
```


---

## Part 15: Implementation Order Summary

| Phase | Deliverable | Stories produced | Dependencies |
|---|---|---|---|
| **1a** | `types.ts`, `sorting.ts`, `selectors.ts` | (none, pure logic) | — |
| **1b** | `appsApi.ts` (RTK Query), `appsBrowserSlice.ts`, `store.ts` | Slice unit tests | types |
| **1c** | MSW handlers: `createAppsHandlers.ts`, fixtures, `defaultHandlers.ts` | (tested via component stories) | types |
| **1d** | Storybook MSW wiring: update `.storybook/preview.ts` with `initialize()` + `mswLoader` | — | msw, msw-storybook-addon |
| **2a** | `AppIcon` + stories | 6 icon stories | types, theme |
| **2b** | `AppsFolderWindow` + stories | 5 folder stories (MSW-backed) | icon, appsApi, msw handlers |
| **3a** | `BrowserColumns` (3 panes) | (tested via browser stories) | appsApi, slice |
| **3b** | `BrowserDetailPanel` (4 detail views) | (tested via browser stories) | appsApi, slice |
| **3c** | `ModuleBrowserWindow` + stories | 6 browser stories (MSW-backed) | columns, detail |
| **4** | `GetInfoWindow` + stories | 4 get-info stories (MSW-backed) | appsApi, Chip |
| **5a** | `SummaryCards`, `DegradedBanner`, `HealthModuleRow` | (tested via dashboard stories) | appsApi, selectors |
| **5b** | `HealthDashboardWindow` + stories | 4 dashboard stories (MSW-backed) | summary, row |
| **6** | Context menu + menu bar + contributions | (tested in full-app story) | all windows |
| **7** | `module.tsx` + wesen-os registration | Full-app story | all components |
| **8** | Integration with live backend | Manual smoke test | running wesen-os |


---

## Part 16: Key Patterns Cheat Sheet

### Creating a window from command handler

```typescript
ctx.openWindow({
  id: `window:apps-browser:get-info:${appId}`,
  title: `${appName} — Get Info`,
  icon: '🔍',
  bounds: { x: 250, y: 100, w: 480, h: 600 },
  content: { kind: 'app', appKey: 'apps-browser-get-info', appParams: { appId } },
  dedupeKey: `apps-browser:get-info:${appId}`,
});
```

### Fetching data (RTK Query hooks)

```typescript
// Apps list — auto-fetches on mount, caches, provides loading/error states
const { data: apps = [], isLoading, isError, refetch } = useGetAppsQuery();

// Reflection — lazy, only fetches when appId is set and reflection is available
const { data: reflection, isLoading: reflLoading } = useGetReflectionQuery(selectedAppId!, {
  skip: !selectedAppId || !selectedApp?.reflection?.available,
});

// Refresh button — invalidate cache tag to trigger re-fetch
const dispatch = useDispatch<AppDispatch>();
const handleRefresh = useCallback(() => {
  dispatch(appsApi.util.invalidateTags(['AppsList']));
}, [dispatch]);
```

### Selection state (Redux slice)

```typescript
const dispatch = useDispatch<AppDispatch>();
const selectedAppId = useSelector(selectSelectedAppId);

// Select a module (clears API + schema selection)
dispatch(selectModule(appId));

// Derived selectors combine RTK Query data + selection state
const sortedApps = useMemo(() => selectSortedApps(apps), [apps]);
const stats = useMemo(() => selectSummaryStats(apps), [apps]);
const selectedApp = useMemo(() => apps.find(a => a.app_id === selectedAppId), [apps, selectedAppId]);
```

### CSS theming

```css
[data-part="health-module-row"] {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: var(--hc-field-padding);
  border-bottom: 1px solid var(--hc-color-border);
  font-family: var(--hc-font-family);
  font-size: var(--hc-font-size);
}

[data-part="health-module-row"][data-state="unhealthy"] {
  background: var(--hc-health-error-bg);
}
```

### Story with MSW handlers (RTK Query fetches real-style)

```typescript
import { createDefaultAppsHandlers } from '../../mocks/msw/defaultHandlers';
import { MOCK_INVENTORY_UNHEALTHY, MOCK_GEPA } from '../../mocks/fixtures/apps';

const meta = {
  title: 'Apps/AppsBrowser/HealthDashboardWindow',
  component: HealthDashboardWindow,
  decorators: [storeDecorator],
  parameters: {
    layout: 'fullscreen',
    msw: { handlers: createDefaultAppsHandlers() },
  },
} satisfies Meta<typeof HealthDashboardWindow>;

// Override handlers per-story for different API responses:
export const DegradedRequired: Story = {
  parameters: {
    msw: {
      handlers: createDefaultAppsHandlers({
        apps: [MOCK_INVENTORY_UNHEALTHY, MOCK_GEPA],
      }),
    },
  },
};
```


---

## Part 17: Risks and Mitigations

| Risk | Mitigation |
|---|---|
| `createAppStore` doesn't support RTK Query middleware | Either extend `createAppStore` to accept `apiSlices` (recommended) or use `configureStore` directly (pinocchio pattern). See Part 3b store creation. |
| MSW not yet initialized in go-go-os Storybook | Add `initialize()` + `mswLoader` to `.storybook/preview.ts`. This is a one-time setup that doesn't affect existing stories (MSW is no-op for stories without `parameters.msw`). |
| Three windows sharing selection state but rendered independently | UI selection lives in `appsBrowserSlice`; server data lives in RTK Query cache. Both propagate via `useSelector`/`useGetAppsQuery` across all windows. |
| RTK Query cache invalidation across windows | All windows share the same store instance, so `invalidateTags(['AppsList'])` triggers re-fetch visible in all windows simultaneously. |
| Reflection 501 handling in RTK Query | Handled via `validateStatus` in the query definition — 501 is treated as a successful response with a sentinel `_unsupported` flag, not as an error. |
| Context menu items leaking across app modules | All command IDs prefixed with `apps-browser.`; contributions scoped by `id: 'apps-browser.contributions'` |
| CSS collisions with other modules | All `data-part` names prefixed with `app-icon-`, `apps-folder-`, etc. |
| MSW handlers need to match real API exactly | Handler factory (`createAppsHandlers`) mirrors the exact response shapes from the design doc. Fixtures are typed against the same `AppManifestDocument`/`ModuleReflectionDocument` interfaces. |

---

## Open Implementation Questions

1. **Should `createAppStore` be extended to support RTK Query `apiSlices`?** Recommendation: yes, this is a small change that benefits any future module wanting RTK Query. The alternative (raw `configureStore` in apps-browser) works but duplicates the engine reducer list.

2. **Should schema fetching (from `schema.uri`) be done eagerly or lazily?** Recommendation: lazy, on first schema selection. Add a third RTK Query endpoint (`getSchema`) with skip logic. Show "Fetch Live" button as the design spec suggests.

3. **Should the icon folder support drag-and-drop (drag icon = copy base URL)?** Recommendation: defer to phase 2. Phase 1 is right-click context menu only.

4. **Should window positions be persisted?** Recommendation: use `dedupeKey` for singleton windows (folder, browser, health dashboard). Get Info windows are per-app-id.

5. **Should MSW be set up globally in go-go-os Storybook or only for apps-browser stories?** Recommendation: global `initialize()` + `mswLoader` in preview.ts. This is the standard pattern and has zero impact on stories that don't use `parameters.msw`. It also prepares the ground for other modules that may want MSW later.
