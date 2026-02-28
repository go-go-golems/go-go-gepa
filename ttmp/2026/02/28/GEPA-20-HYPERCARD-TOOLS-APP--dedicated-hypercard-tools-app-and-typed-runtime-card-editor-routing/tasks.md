# Tasks

## TODO

- [x] 1. Establish ticket artifacts
- [x] 1.1 Write implementation plan document with architecture, API contracts, rollout, and validation strategy
- [x] 1.2 Create implementation diary and start chronological logging
- [x] 1.3 Relate key code files and docs with `docmgr doc relate`
- [x] 2. Introduce typed runtime-card identity contract
- [x] 2.1 Add `RuntimeCardRef` type in engine hypercard editor module
- [x] 2.2 Add helper codecs for editor instance IDs (`encodeRuntimeCardEditorInstanceId`, `decodeRuntimeCardEditorInstanceId`)
- [x] 2.3 Add helper for runtime-card editor app key construction
- [x] 2.4 Add tests for encode/decode and app-key helpers
- [x] 3. Build dedicated HyperCard Tools launcher module
- [x] 3.1 Scaffold `go-go-os/apps/hypercard-tools` package (`package.json`, `tsconfig`, `src/index.ts`, launcher exports)
- [x] 3.2 Implement launchable app module with manifest id `hypercard-tools`
- [x] 3.3 Implement runtime-card editor window renderer delegating to `CodeEditorWindow`
- [x] 3.4 Add command/open-window helpers for future tools windows
- [x] 3.5 Add module tests for instance-id parsing and unknown-instance fallback
- [x] 4. Hard-cutover editor launch path to `hypercard-tools`
- [x] 4.1 Update `editorLaunch.ts` to emit `hypercard-tools:<encodedInstanceId>` app keys
- [x] 4.2 Change payload/window IDs and dedupe keys to new convention
- [x] 4.3 Preserve initial-code stash semantics by card identity
- [x] 4.4 Update all call sites (`RuntimeCardDebugWindow`, timeline card/widget renderers)
- [x] 5. Integrate app module into launcher composition
- [x] 5.1 Register `hypercardToolsLauncherModule` in `wesen-os/apps/os-launcher/src/app/modules.tsx`
- [x] 5.2 Add TS path aliases in `wesen-os/apps/os-launcher/tsconfig.json`
- [x] 5.3 Validate launcher registry resolves `hypercard-tools:*`
- [x] 6. Remove inventory-owned editor routing
- [x] 6.1 Remove `CodeEditorWindow` and `getEditorInitialCode` imports from inventory launcher
- [x] 6.2 Remove code-editor instance-prefix routing branch in inventory window renderer
- [x] 6.3 Verify no remaining `code-editor:<cardId>` app-key emission paths
- [x] 7. Validate end-to-end behavior
- [x] 7.1 Run targeted unit tests in engine/desktop-os/os-launcher
- [x] 7.2 Run TS typechecks for modified workspaces
- [ ] 7.3 Manual smoke: runtime debug -> Edit opens editor, save inject path still works
- [ ] 8. Commit and documentation closure
- [x] 8.1 Commit code changes in focused commits
- [x] 8.2 Check off tasks and update changelog per step
- [x] 8.3 Final diary entry with validation evidence and follow-ups
