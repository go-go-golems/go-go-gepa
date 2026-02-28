# Changelog

## 2026-02-27

- Initial workspace created.
- Created primary design document: `design-doc/01-apps-browser-ux-and-technical-reference.md`.
- Created detailed diary document: `reference/01-implementation-diary.md`.
- Mapped base backend module-discovery contracts (`/api/os/apps`, reflection route) from `go-go-os/pkg/backendhost`.
- Mapped runtime mount sequence and namespaced routing behavior from `wesen-os` launcher.
- Mapped module capability surfaces for inventory and GEPA, including reflection asymmetry (`inventory` 501, `gepa` 200).
- Captured live payload examples from local runtime for module list, reflection, schema, and profile endpoints.
- Authored 5+ page UX-facing reference with endpoint catalog, data model, UI state guidance, interaction flows, pseudocode, and diagrams.
- Ran `docmgr doctor --ticket GEPA-11-APPS-BROWSER-UI-WIDGET --stale-after 30` with all checks passing.
- Uploaded final bundle to reMarkable after dry-run:
  - `GEPA-11 Apps Browser UX Packet.pdf`
  - remote path: `/ai/2026/02/27/GEPA-11-APPS-BROWSER-UI-WIDGET`
- Backfilled integration work for wiring apps-browser into `wesen-os`, including module registration/alias updates and adapter provider fix for RTK Query.
- Wired Apps Browser interactions:
  - double-click module icon -> open Module Browser preselected to module
  - right-click module icon -> context actions (`Open in Browser`, `Get Info`, `Open Health Dashboard`, `Launch App`)
  - Get Info footer action -> opens Browser window
  - Health Dashboard row click -> opens Get Info window
- Added apps-browser launcher command handlers (`apps-browser.open-browser`, `apps-browser.get-info`, `apps-browser.open-health`).
- Added launcher host regression test in `wesen-os` for apps-browser command routing and payload assertions.
- Re-ran integration validation in `wesen-os/apps/os-launcher`:
  - `npm run test -w apps/os-launcher` (pass)
  - `npm run typecheck -w apps/os-launcher` (pass)
  - `npm run build -w apps/os-launcher` (pass)
- Implemented schema-detail on-demand fetching in Apps Browser:
  - Added `getSchemaDocument` RTK Query endpoint for schema URI fetch.
  - Replaced static placeholder text with actionable fetch button and inline error handling.
  - Added fallback URI resolution (`/api/apps/{appId}/schemas/{schemaId}`) when reflection schema URI is missing.
- Refined schema behavior after runtime verification:
  - schema detail now auto-fetches full payload on schema selection
  - manual retry button remains available (`Fetch schema again`)
- Fixed `wesen-os` frontend dev proxy for generic module routes:
  - added `/api/apps` proxy in `apps/os-launcher/vite.config.ts`
  - resolves schema-fetch responses returning `index.html` instead of JSON
- Extended API method detail panel:
  - shows request/response/error schema blocks at the bottom of method info
  - each block auto-fetches and renders full schema payload (with retry/error state)
- Updated schema previews for readability:
  - schema sections are folded/collapsed by default
  - payload rendering now uses YAML + shared syntax highlighting (`SyntaxHighlight`, `toYaml`) like debug windows
- Added documentation visibility in frontend:
  - `Get Info` reflection panel now lists `docs` entries from reflection
  - module detail panel now shows a `Documentation` section
- Added Module Browser right-click support on module rows:
  - context actions now available directly in Module Browser module list (`Open in Browser`, `Get Info`, `Open Health Dashboard`, `Launch App`)
- Expanded `ReflectionDocLink` frontend typing to include `path` and `description` fields from backend contract.

## 2026-02-28

Cleanup: all ticket tasks complete; closing ticket.

