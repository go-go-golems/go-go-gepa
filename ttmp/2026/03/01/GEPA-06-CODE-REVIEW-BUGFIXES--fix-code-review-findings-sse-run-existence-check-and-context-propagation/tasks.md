# Tasks

## TODO

- [x] Fix SSE run-events endpoint to verify run existence before writing event-stream headers/body and add regression test for unknown run IDs.
- [x] Thread caller context through optimizer plugin execution path (callPluginFunction + Dataset/Evaluate/Run/Merge/InitialCandidate/SelectComponents/ComponentSideInfo and call sites).
- [x] Thread caller context through dataset generator execution path (RunWithRuntime/GenerateRows/GenerateOne and command/test call sites).
- [x] Scan remaining production context.Background() uses and keep only intentional fallback behavior.
- [x] Run focused and package-level tests for backendmodule, runner plugin loader, and dataset generator paths.
- [x] Update plan/diary/changelog, mark tasks complete, and close the ticket.
