# Tasks

## Phase 1 - Ticket Setup

- [x] `GEPA11-01` Create ticket workspace and baseline docs.
- [x] `GEPA11-02` Add primary design-doc and detailed implementation diary docs.

## Phase 2 - Endpoint and Data Model Research

- [x] `GEPA11-10` Map base module discovery contracts in `go-go-os/pkg/backendhost`.
- [x] `GEPA11-11` Map runtime mounting sequence in `wesen-os` launcher.
- [x] `GEPA11-12` Map inventory and GEPA module-specific manifests/routes/reflection behaviors.
- [x] `GEPA11-13` Capture live endpoint payload examples from running local stack.

## Phase 3 - UX-Facing Documentation

- [x] `GEPA11-20` Write 5+ page non-technical Apps Browser reference with endpoint tables, model reference, diagrams, and pseudocode.
- [x] `GEPA11-21` Keep detailed chronological diary of commands/findings/decisions.
- [x] `GEPA11-22` Update ticket changelog and related-file metadata.

## Phase 4 - Validation and Delivery

- [x] `GEPA11-30` Run `docmgr doctor --ticket GEPA-11-APPS-BROWSER-UI-WIDGET --stale-after 30` and resolve warnings.
- [x] `GEPA11-31` Upload final doc bundle to reMarkable with dry-run first and verify cloud listing.

## Phase 5 - Package Scaffold

- [x] `GEPA11-40` Create `go-go-os/apps/apps-browser/` package: `package.json`, `tsconfig.json`, `src/index.ts`.
- [x] `GEPA11-41` Add MSW + msw-storybook-addon devDependencies; add story glob to `.storybook/main.ts`; wire MSW `initialize()` + `mswLoader` in `.storybook/preview.ts`.
- [x] `GEPA11-42` Add `apps/apps-browser` to root `tsconfig.json` references.

## Phase 6 - Domain Types, RTK Query API, Selection Slice

- [x] `GEPA11-50` Write `src/domain/types.ts` with `AppManifestDocument`, `ModuleReflectionDocument`, and related interfaces.
- [x] `GEPA11-51` Write `src/api/appsApi.ts` with RTK Query `createApi`: `getApps` and `getReflection` endpoints.
- [x] `GEPA11-52` Write `src/features/appsBrowser/appsBrowserSlice.ts` with selection-only state (`selectedAppId`, `selectedApiId`, `selectedSchemaId`).
- [x] `GEPA11-53` Write `src/app/store.ts` combining engine reducers + appsApi + appsBrowser slice.
- [x] `GEPA11-54` Write `src/domain/selectors.ts` (sortedApps, summaryStats, crossRefSchemaIds, etc.).
- [x] `GEPA11-55` Write `src/domain/sorting.ts` (unhealthy-first, required-first, name-asc).

## Phase 7 - MSW Mock Layer

- [x] `GEPA11-60` Write `src/mocks/fixtures/apps.ts` with mock data constants (MOCK_INVENTORY, MOCK_GEPA, MOCK_GEPA_REFLECTION, etc.).
- [x] `GEPA11-61` Write `src/mocks/msw/createAppsHandlers.ts` handler factory.
- [x] `GEPA11-62` Write `src/mocks/msw/defaultHandlers.ts` default wiring.

## Phase 8 - AppIcon Component + Stories

- [x] `GEPA11-70` Write `src/components/AppIcon.tsx` + `AppIcon.css`.
- [x] `GEPA11-71` Write `src/components/AppIcon.stories.tsx` (7 stories: HealthyOptional, HealthyRequired, HealthyReflective, HealthyRequiredReflective, UnhealthyRequired, Selected, IconGrid).
- [x] `GEPA11-72` Verify stories render in Storybook.

## Phase 9 - AppsFolderWindow + Stories

- [x] `GEPA11-80` Write `src/components/AppsFolderWindow.tsx` + `AppsFolderWindow.css`.
- [x] `GEPA11-81` Write `src/components/AppsFolderWindow.stories.tsx` (5 stories: Default, WithUnhealthy, ManyModules, Loading, Empty — all MSW-backed).
- [x] `GEPA11-82` Verify stories render in Storybook. Also fixed MSW setup (mockServiceWorker.js + staticDirs).

## Phase 10 - ModuleBrowserWindow + Stories

- [x] `GEPA11-90` Write `src/components/BrowserColumns.tsx` (ModuleListPane, APIListPane, SchemaListPane).
- [x] `GEPA11-91` Write `src/components/BrowserDetailPanel.tsx` (ModuleDetail, APIDetail, SchemaDetail, EmptyDetail).
- [x] `GEPA11-92` Write `src/components/ModuleBrowserWindow.tsx` + `ModuleBrowserWindow.css`.
- [x] `GEPA11-93` Write `src/components/ModuleBrowserWindow.stories.tsx` (6 stories — MSW-backed).
- [x] `GEPA11-94` Verify stories render in Storybook.

## Phase 11 - GetInfoWindow + Stories

- [x] `GEPA11-100` Write `src/components/GetInfoWindow.tsx` + `GetInfoWindow.css`.
- [x] `GEPA11-101` Write `src/components/GetInfoWindow.stories.tsx` (4 stories — MSW-backed).
- [x] `GEPA11-102` Verify stories render in Storybook.

## Phase 12 - HealthDashboardWindow + Stories

- [x] `GEPA11-110` Write SummaryCards, DegradedBanner, HealthModuleRow as sub-components within `HealthDashboardWindow.tsx`.
- [x] `GEPA11-111` Write `src/components/HealthDashboardWindow.tsx` + `HealthDashboardWindow.css`.
- [x] `GEPA11-112` Write `src/components/HealthDashboardWindow.stories.tsx` (4 stories — MSW-backed).
- [x] `GEPA11-113` Verify stories render in Storybook.

## Phase 13 - Launcher Module + Registration

- [x] `GEPA11-120` Write `src/launcher/module.tsx` (LaunchableAppModule, window adapter, window payload builders).
- [x] `GEPA11-121` Write `src/launcher/public.ts` exports.
- [x] `GEPA11-122` Write full-app story `src/app/stories/AppsBrowserApp.stories.tsx` (6 stories).

## Phase 14 - Final Validation

- [x] `GEPA11-130` Run `tsc --build` from go-go-os root — passes clean.
- [x] `GEPA11-131` Run `biome check` and fix all errors (21 auto-fixes, 1 manual fix).
- [x] `GEPA11-132` Run Storybook build — passes clean, all stories included.
- [x] `GEPA11-133` Update diary with implementation notes (Steps 6-7).

## Phase 15 - Interaction Wiring Follow-up

- [x] `GEPA11-140` Wire Apps Folder double-click behavior to open Module Browser preselected to clicked module.
- [x] `GEPA11-141` Wire per-module right-click context menu actions (`Open in Browser`, `Get Info`, `Open Health Dashboard`, `Launch App`).
- [x] `GEPA11-142` Add launcher command handlers for `apps-browser.open-browser`, `apps-browser.get-info`, `apps-browser.open-health`.
- [x] `GEPA11-143` Wire cross-window navigation callbacks (`Get Info -> Open in Browser`, `Health row -> Get Info`).
- [x] `GEPA11-144` Add/extend launcher host tests in `wesen-os` to lock command routing and window payloads.
- [x] `GEPA11-145` Run integration validation (`test`, `typecheck`, `build`) for `wesen-os/apps/os-launcher`.
- [x] `GEPA11-146` Backfill GEPA-11 diary with prior integration commits and current interaction changes.
- [x] `GEPA11-147` Implement schema detail fetch action (`Fetch from ... to view the full schema as well`) with on-demand RTK Query loading and fallback endpoint resolution.
- [x] `GEPA11-148` Auto-fetch schema payload when a schema row is selected (keep manual retry action visible).
- [x] `GEPA11-149` Fix frontend dev proxy for generic `/api/apps/*` so schema fetch routes do not fall through to SPA HTML.
- [x] `GEPA11-150` Show request/response/error schema previews at the bottom of selected API method detail panel.
- [x] `GEPA11-151` Fold API schema previews by default and render schema payloads as syntax-highlighted YAML.
- [x] `GEPA11-152` Surface reflection docs in frontend module/get-info detail views.
- [x] `GEPA11-153` Add right-click context menu support for Module Browser module rows.
