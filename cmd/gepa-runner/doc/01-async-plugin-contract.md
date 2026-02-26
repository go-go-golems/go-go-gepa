---
Title: Async Plugin Contract and Promise Semantics
Slug: gepa-runner-async-plugin-contract
Short: |
  Understand how `gepa-runner` executes sync and Promise-returning JS plugin methods, and how plugin events are streamed.
Topics:
- gepa
- plugins
- async
- promises
- streaming
Commands:
- optimize
- eval
- candidate run
- dataset generate
Flags:
- stream
- profile
- profile-registries
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---
`gepa-runner` supports two plugin return styles for method calls such as `run()`, `evaluate()`, and `generateOne()`:

1. Immediate object/string return (synchronous behavior).
2. Promise return (asynchronous behavior).

This matters because plugin authors can now compose `geppetto` asynchronous APIs like `session.start(...).promise` without building custom wrapper scripts.

## How Execution Works

Plugin calls are executed on the Goja runtime owner, then normalized:

1. If the JS method returns a non-Promise value, it is decoded immediately.
2. If it returns a Promise:
   - fulfilled Promise values are decoded and returned,
   - rejected Promises fail the command with rejection context,
   - pending Promises are awaited with a timeout guard.

This ensures command behavior is deterministic and avoids deadlocks caused by waiting on the wrong goroutine.

## Event Emission Contract

Plugin methods receive optional event emit hooks via options:

1. `options.emitEvent(payload)`
2. `options.events.emit(payload)`

Both hooks are equivalent. The host wraps each event with metadata such as sequence number, plugin identifiers, and method name.

When `--stream` is enabled, emitted events are printed as they arrive:

```text
stream-event {"kind":"plugin_stream","command":"candidate_run","event":{...}}
```

## Minimal Async Plugin Patterns

Promise-returning candidate run:

```javascript
run(input, options) {
  return Promise.resolve().then(() => {
    options.emitEvent({ type: "run-start", data: { id: input.id } });
    return { output: { ok: true } };
  });
}
```

Promise-returning dataset row generation:

```javascript
generateOne(input, options) {
  return Promise.resolve().then(() => {
    options.events.emit({ type: "row-progress", data: { index: input.index } });
    return { row: { value: "ok" }, metadata: { source: "async" } };
  });
}
```

## Output Compatibility Expectations

Async support does not change the shape of final command output:

1. `candidate run` still emits the same `runId/plugin/input/output/...` payload.
2. `dataset generate` still emits the same generation summary and output file/db metadata.
3. Streaming lines are additive and only shown when `--stream` is set.

## Troubleshooting

| Problem | Cause | Solution |
|---|---|---|
| Command exits with "promise rejected" | Plugin Promise rejected without handling | Add `catch` logic in plugin and emit contextual event before rethrowing |
| No stream lines visible | `--stream` not enabled or plugin does not emit events | Pass `--stream` and call `options.emitEvent(...)` in plugin method |
| Promise never resolves | Plugin Promise not settled on some path | Ensure every branch resolves or rejects; add timeout-safe guards in plugin logic |
| Plugin works sync but fails async | Returned shape differs after Promise chain | Ensure fulfilled Promise resolves to same object contract as sync path |

## See Also

1. `glaze help gepa-runner-candidate-run-streaming-example`
2. `glaze help gepa-runner-dataset-generate-streaming-example`
3. `glaze help gepa-runner-async-streaming-troubleshooting`
