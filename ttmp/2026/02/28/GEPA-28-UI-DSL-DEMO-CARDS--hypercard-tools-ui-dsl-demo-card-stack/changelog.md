# Changelog

## 2026-02-28

- Initial workspace created


## 2026-02-28 - Design and onboarding deliverables

Authored the intern-focused architecture/implementation guide, created phased task list, and uploaded the initial guide+diary bundle to reMarkable.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-28-UI-DSL-DEMO-CARDS--hypercard-tools-ui-dsl-demo-card-stack/design-doc/01-ui-dsl-demo-cards-architecture-implementation-plan-and-intern-onboarding-guide.md — Primary design doc
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/28/GEPA-28-UI-DSL-DEMO-CARDS--hypercard-tools-ui-dsl-demo-card-stack/reference/01-investigation-and-implementation-diary.md — Chronological diary


## 2026-02-28 - Implementation and validation

Implemented HyperCard Tools UI DSL demo stack (commit 8c79f49), switched launcher icon launch to card-session workspace, updated launcher host tests in wesen-os (commit f567ed6), and validated with targeted typecheck/test runs.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/apps/hypercard-tools/src/domain/pluginBundle.vm.js — Demo card stack and handlers
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/apps/hypercard-tools/src/launcher/module.tsx — Launcher cutover to card-window default
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/wesen-os/apps/os-launcher/src/__tests__/launcherHost.test.tsx — Launch expectation updates for HyperCard Tools


## 2026-02-28 - Phase 5 DSL expansion

Extended UI DSL and runtime support with dropdown/selectableTable/gridBoard nodes, added schema coverage tests, expanded HyperCard Tools demos for missing examples, and wired app_hypercard_tools domain reducer for domain-intent demonstrations.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/apps/hypercard-tools/src/domain/pluginBundle.vm.js — Expanded demo cards and missing examples
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/apps/hypercard-tools/src/launcher/module.tsx — Domain reducer wiring for domain intent demos
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/hypercard-runtime/src/plugin-runtime/uiSchema.ts — Validation logic for new node kinds
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/hypercard-runtime/src/plugin-runtime/uiTypes.ts — Added new UI node kinds
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-os/packages/hypercard-runtime/src/runtime-host/PluginCardRenderer.tsx — Renderer mapping for dropdown/selectableTable/gridBoard

