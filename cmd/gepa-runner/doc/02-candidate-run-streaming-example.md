---
Title: Candidate Run Streaming Example
Slug: gepa-runner-candidate-run-streaming-example
Short: |
  Build and run a Promise-based candidate plugin that emits live stream events during `candidate run`.
Topics:
- gepa
- candidate
- plugins
- streaming
Commands:
- candidate run
Flags:
- script
- config
- input-file
- stream
- output-format
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Example
---
This example shows how to run a single candidate against one input while streaming plugin events in real time. It is useful when developing plugin behavior and validating intermediate states before adding evaluator logic.

## Plugin Script

Create `candidate-plugin.js`:

```javascript
const { defineOptimizerPlugin, OPTIMIZER_PLUGIN_API_VERSION } = require("gepa/plugins");

module.exports = defineOptimizerPlugin({
  apiVersion: OPTIMIZER_PLUGIN_API_VERSION,
  kind: "optimizer",
  id: "examples.candidate-stream",
  name: "Candidate Stream Example",
  create() {
    return {
      run(input, options) {
        return Promise.resolve().then(() => {
          options.emitEvent({ type: "candidate-start", data: { input } });
          options.events.emit({ type: "candidate-progress", message: "assembling result" });
          return {
            output: {
              answer: "ok",
              question: input.question || ""
            },
            metadata: {
              mode: "async-demo"
            }
          };
        });
      }
    };
  }
});
```

## Config and Input Files

Candidate config (`candidate-config.yaml`):

```yaml
apiVersion: gepa.candidate-run/v2
candidate:
  prompt: "Answer concisely"
metadata:
  candidate_id: "demo-candidate-1"
  reflection_used: "none"
```

Input payload (`candidate-input.json`):

```json
{
  "question": "What is the status?"
}
```

## Command

```bash
gepa-runner candidate run \
  --script ./candidate-plugin.js \
  --config ./candidate-config.yaml \
  --input-file ./candidate-input.json \
  --stream \
  --output-format json
```

## Expected Behavior

1. One or more `stream-event ...` lines appear while the Promise resolves.
2. Final JSON result is printed with normal candidate-run structure.
3. Stream output is additive and does not replace the final result payload.

## Troubleshooting

| Problem | Cause | Solution |
|---|---|---|
| No streamed lines | Plugin never calls emit hook | Call `options.emitEvent(...)` or `options.events.emit(...)` |
| Stream lines appear but final output missing | Promise never resolves | Ensure the Promise chain returns a final object |
| Final output exists but missing metadata | Plugin returns only `output` | Return `{ output: ..., metadata: ... }` when metadata is needed |
| Unknown plugin API error | Descriptor mismatch | Confirm `apiVersion` and `kind` in plugin descriptor |

## See Also

1. `glaze help gepa-runner-async-plugin-contract`
2. `glaze help gepa-runner-dataset-generate-streaming-example`
