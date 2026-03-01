---
Title: Async and Streaming Troubleshooting
Slug: gepa-runner-async-streaming-troubleshooting
Short: |
  Diagnose Promise settlement failures, stream emission issues, and common plugin contract mistakes.
Topics:
- gepa
- troubleshooting
- promises
- streaming
Commands:
- candidate run
- dataset generate
- eval
- optimize
Flags:
- stream
- debug
- print-parsed-fields
IsTemplate: false
IsTopLevel: false
ShowPerDefault: false
SectionType: GeneralTopic
---
This guide helps you debug failures that appear only after moving plugins from synchronous returns to Promise-based flows.

## Failure Triage Strategy

Use this order to isolate issues:

1. Confirm descriptor validity (`apiVersion`, `kind`, `create`).
2. Confirm method contract shape (`run`, `evaluate`, `generateOne`).
3. Confirm Promise settles (`resolve` or `reject`) on all paths.
4. Confirm emitted event payloads are valid objects/strings.
5. Confirm final output shape matches command expectations.

This order avoids spending time on stream formatting before contract errors are fixed.

## Diagnostics You Should Enable

Use these flags when debugging:

```bash
--stream
--debug
--print-parsed-fields
```

`--stream` reveals real-time plugin events. `--debug` and `--print-parsed-fields` reveal parsed runtime/profile settings so configuration mismatches are visible.

## Common Error Patterns

| Error Pattern | What It Usually Means | Immediate Next Step |
|---|---|---|
| `promise rejected: ...` | Plugin explicitly rejected Promise | Add event emission before rejection and inspect payload |
| `promise did not settle before deadline` | Promise never resolved/rejected | Audit every branch; ensure timer/network branches settle |
| `invalid return value` | Fulfilled Promise returned wrong shape | Compare returned object against method contract |
| `plugin ... not initialized` | Loader/runtime state invalid | Re-check command inputs and script loading path |

## Event Payload Guidance

Recommended event shape:

```json
{
  "type": "row-progress",
  "level": "info",
  "message": "building entity timeline",
  "data": {
    "row_index": 3
  }
}
```

Keep payloads small and focused. Large payloads make stream logs noisy and difficult to inspect.

## See Also

1. `glaze help gepa-runner-async-plugin-contract`
2. `glaze help gepa-runner-candidate-run-streaming-example`
3. `glaze help gepa-runner-dataset-generate-streaming-example`
