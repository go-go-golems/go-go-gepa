# Changelog

## 2026-02-28

- Initial workspace created


## 2026-02-28

Completed intern-focused architecture deep-dive and migration design for splitting HyperCard/plugin runtime from desktop engine; documented current runtime flow, coupling map, target package boundaries, phased migration, and risks.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-26-HYPERCARD-RUNTIME-SPLIT--split-hypercard-runtime-plugin-architecture-into-dedicated-package-separate-from-desktop-engine/design-doc/01-hypercard-runtime-package-split-architecture-and-migration-guide.md — Primary verbose design and onboarding guide
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-26-HYPERCARD-RUNTIME-SPLIT--split-hypercard-runtime-plugin-architecture-into-dedicated-package-separate-from-desktop-engine/reference/01-investigation-diary-hypercard-runtime-package-split.md — Chronological investigation diary with command trace and rationale
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/components/shell/windowing/PluginCardSessionHost.tsx — Key coupling seam analyzed in proposal
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/plugin-runtime/runtimeService.ts — Runtime core analyzed for extraction boundaries


## 2026-02-28

Completed delivery operations: related-file linkage, doctor validation, and reMarkable upload verification for GEPA-26 design bundle.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-26-HYPERCARD-RUNTIME-SPLIT--split-hypercard-runtime-plugin-architecture-into-dedicated-package-separate-from-desktop-engine/reference/01-investigation-diary-hypercard-runtime-package-split.md — Diary updated with final execution evidence
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-26-HYPERCARD-RUNTIME-SPLIT--split-hypercard-runtime-plugin-architecture-into-dedicated-package-separate-from-desktop-engine/tasks.md — Checklist closed after validation and publish


## 2026-02-28

Executed GEPA-26 hard-cut implementation across codebases: introduced `@hypercard/hypercard-runtime`, removed runtime ownership from engine, rewired go-go-os apps + desktop-os store plumbing, migrated runtime tests, and updated arc-agi-player integration.

### Commits

- `go-go-os`: `5f564f4` — `refactor: split hypercard runtime into dedicated package`
- `go-go-app-arc-agi-3`: `c86ea5e` — `refactor: rewire arc-agi-player to hypercard-runtime package`

### Validation

- `go-go-os` root typecheck: pass (`npm run typecheck`)
- `go-go-os/packages/engine` tests: pass (`npx vitest run`)
- `go-go-os/packages/hypercard-runtime` tests: pass (`npm run test -w packages/hypercard-runtime`)
- `go-go-os/packages/desktop-os` tests: pass (`npm run test -w packages/desktop-os`)
- Known unrelated pre-existing issue: Storybook taxonomy gate fails for `apps/apps-browser` when running package/root `npm run test` scripts that invoke `scripts/storybook/check-taxonomy.mjs`.
- Updated implementation diary bundle uploaded and verified on reMarkable:
  - `GEPA-26-runtime-split-implementation-diary-update` in `/ai/2026/02/28/GEPA-26-HYPERCARD-RUNTIME-SPLIT`

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/hypercard-runtime/src/index.ts — New runtime package public API boundary
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/index.ts — Engine runtime exports removed
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/desktop-os/src/store/createLauncherStore.ts — Launcher store now uses runtime package app store factory
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-app-arc-agi-3/apps/arc-agi-player/src/launcher/module.tsx — ARC app launcher now hosts runtime session from new package


## 2026-02-28

Addressed downstream tmux launcher regression discovered post-cut: added `@hypercard/hypercard-runtime` alias wiring in `wesen-os` launcher and migrated remaining inventory app runtime imports off removed engine adapter exports.

### Commits

- `wesen-os`: `138d455` — `fix: wire hypercard-runtime alias in os-launcher configs`
- `go-go-app-inventory`: `e5710f2` — `refactor: migrate inventory app runtime imports to hypercard-runtime`

### Validation

- `wesen-os/apps/os-launcher`: `npm run build` passes after alias + inventory rewires.
