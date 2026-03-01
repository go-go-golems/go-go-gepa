# Tasks

## TODO

- [x] Implement shared JS Promise settlement helper for plugin call return values.
- [x] Add unit tests for settlement helper: immediate resolve, delayed resolve, reject, timeout, cancellation.
- [x] Integrate settlement helper into optimizer plugin loader (`cmd/gepa-runner/plugin_loader.go`).
- [x] Integrate settlement helper into dataset generator plugin loader (`pkg/dataset/generator/plugin_loader.go`).
- [x] Preserve backward compatibility for synchronous plugin returns (existing tests must remain green).
- [x] Add plugin option event sink API (`emitEvent` or `options.events.emit`) for runtime event forwarding.
- [x] Add safe host-side event envelope/validation (sequence, timestamp, plugin id, run id where available).
- [x] Add `--stream` CLI output mode for `candidate run` and `dataset generate`.
- [x] Add integration tests for async Promise plugins on both candidate and dataset paths.
- [x] Add integration tests for emitted streaming events in CLI mode.
- [x] Add example async plugin scripts in ticket `scripts/` directory.
- [x] Document async plugin contract and streaming behavior in command docs/README.
- [x] Record experiment outputs and failure cases in `reference/01-implementation-diary.md`.
- [x] Update changelog per milestone completion.

## Suggested Milestones

- [x] M1: Promise settlement helper + tests complete.
- [x] M2: Loader integrations complete and passing tests.
- [x] M3: Event sink + `--stream` output complete.
- [x] M4: Docs/examples complete and ready for handoff.
