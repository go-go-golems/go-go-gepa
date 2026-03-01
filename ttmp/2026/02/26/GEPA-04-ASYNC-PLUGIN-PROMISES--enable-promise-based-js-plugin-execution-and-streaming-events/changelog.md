# Changelog

## 2026-02-26

- Initial workspace created


## 2026-02-26 - Initial scoping and intern implementation plan

Created GEPA-04 ticket with detailed implementation plan for Promise-returning JS plugin support and streaming event propagation, added explicit task breakdown, and initialized implementation diary.

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-04-ASYNC-PLUGIN-PROMISES--enable-promise-based-js-plugin-execution-and-streaming-events/index.md — Ticket overview updated with links and concrete file scope
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-04-ASYNC-PLUGIN-PROMISES--enable-promise-based-js-plugin-execution-and-streaming-events/planning/01-promise-aware-plugin-bridge-and-streaming-events-implementation-plan.md — Intern-facing detailed implementation and scoping document
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-04-ASYNC-PLUGIN-PROMISES--enable-promise-based-js-plugin-execution-and-streaming-events/reference/01-implementation-diary.md — Chronological diary initialized
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-04-ASYNC-PLUGIN-PROMISES--enable-promise-based-js-plugin-execution-and-streaming-events/tasks.md — Detailed milestone and engineering task checklist

## 2026-02-26 - Promise settlement + loader async support

Implemented Promise-aware plugin invocation and event emission infrastructure in `go-go-gepa`.

### Related Commits

- `462e0a8` `feat(jsbridge): add promise settlement helper and plugin event emitter`
- `00c4063` `feat(plugins): support promise-returning JS plugin methods`
- `4b49213` `feat(runner): add --stream output for candidate and dataset commands`

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/pkg/jsbridge/call_and_resolve.go — Shared Promise settlement helper
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/pkg/jsbridge/emitter.go — Plugin event sink with host envelope metadata
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/plugin_loader.go — Async-aware optimizer plugin loader
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/pkg/dataset/generator/plugin_loader.go — Async-aware dataset plugin loader
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/candidate_run_command.go — `--stream` support for candidate run
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/dataset_generate_command.go — `--stream` support for dataset generate

## 2026-02-26 - Help docs, stream CLI integration tests, and runnable GEPA-04 scripts

Added Glazed help pages for async contract and streaming examples, wired embedded docs into CLI help, added stream integration tests, and completed GEPA-04 scripts for candidate and dataset flows.

### Related Commits

- `85e8b58` `docs(runner): add glazed async streaming help and CLI tests`
- `36e98b9` `docs(gepa-04): add detailed diary, tasks, changelog, and scripts`

### Related Files

- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/doc/doc.go — Embedded Glazed help loader
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/doc/01-async-plugin-contract.md — Async plugin contract documentation
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/doc/02-candidate-run-streaming-example.md — Candidate run streaming example
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/doc/03-dataset-generate-streaming-example.md — Dataset generate streaming example
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/doc/04-async-streaming-troubleshooting.md — Troubleshooting reference
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/cmd/gepa-runner/stream_cli_integration_test.go — End-to-end stream output assertions
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-04-ASYNC-PLUGIN-PROMISES--enable-promise-based-js-plugin-execution-and-streaming-events/scripts/exp-02-dataset-config.yaml — Dataset stream experiment config
- /home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa/ttmp/2026/02/26/GEPA-04-ASYNC-PLUGIN-PROMISES--enable-promise-based-js-plugin-execution-and-streaming-events/scripts/exp-02-run-dataset-stream.sh — Dataset stream experiment runner

## 2026-02-28

Cleanup: all ticket tasks complete; closing ticket.

