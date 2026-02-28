# Changelog

## 2026-02-28

- Initial workspace created


## 2026-02-28

Initialized GEPA-20 ticket, wrote detailed implementation plan, seeded granular task list, and started implementation diary.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-20-HYPERCARD-TOOLS-APP--dedicated-hypercard-tools-app-and-typed-runtime-card-editor-routing/design-doc/01-implementation-plan-dedicated-hypercard-tools-app-and-typed-runtime-card-editor-routing.md — Primary architecture and rollout plan
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-20-HYPERCARD-TOOLS-APP--dedicated-hypercard-tools-app-and-typed-runtime-card-editor-routing/reference/01-implementation-diary.md — Chronological implementation log
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-20-HYPERCARD-TOOLS-APP--dedicated-hypercard-tools-app-and-typed-runtime-card-editor-routing/tasks.md — Granular task execution checklist


## 2026-02-28

Removed inventory-owned code editor window rendering branch for hard cutover (commit 9edbedd).

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-app-inventory/apps/inventory/src/launcher/renderInventoryApp.tsx — Deleted legacy code-editor instance branch and imports


## 2026-02-28

Registered hypercard-tools in OS launcher composition and added regression tests/config alias wiring (commit af56170).

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/wesen-os/apps/os-launcher/src/__tests__/launcherHost.test.tsx — Encoded editor instance render/fallback coverage
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/wesen-os/apps/os-launcher/src/app/modules.tsx — Launcher module registration
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/wesen-os/apps/os-launcher/vite.config.ts — Vite alias for new workspace app
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/wesen-os/apps/os-launcher/vitest.config.ts — Vitest alias for new workspace app


## 2026-02-28

Implemented typed runtime-card editor identity and dedicated hypercard-tools app module (commit b25e276).

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/apps/hypercard-tools/src/launcher/module.tsx — New HyperCard Tools launcher module rendering editor windows
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/hypercard/editor/editorLaunch.ts — Hard-cutover editor app key and payload routing
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/hypercard/editor/runtimeCardRef.test.ts — Helper contract tests
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/hypercard/editor/runtimeCardRef.ts — Added RuntimeCardRef and editor instance codec helpers


## 2026-02-28

Validation completed for automated checks (typecheck + unit/integration tests); manual smoke remains pending task 7.3.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-20-HYPERCARD-TOOLS-APP--dedicated-hypercard-tools-app-and-typed-runtime-card-editor-routing/tasks.md — Task status reflects only manual smoke pending
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/engine/src/hypercard/editor/runtimeCardRef.test.ts — Targeted helper tests pass
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/wesen-os/apps/os-launcher/src/__tests__/launcherHost.test.tsx — Launcher host tests include hypercard-tools rendering paths

