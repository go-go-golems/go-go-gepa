---
Title: Dataset Generate Streaming Example
Slug: gepa-runner-dataset-generate-streaming-example
Short: |
  Use an async dataset generator plugin that emits row-level progress events with `dataset generate --stream`.
Topics:
- gepa
- dataset
- plugins
- streaming
Commands:
- dataset generate
Flags:
- script
- config
- stream
- dry-run
- output-dir
- output-db
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Example
---
This example demonstrates row-by-row event streaming from a Promise-based dataset generator plugin. Use this during synthetic data prompt tuning to inspect generation behavior before writing files or sqlite rows.

## Plugin Script

Create `dataset-plugin.js`:

```javascript
const { defineDatasetGenerator, DATASET_GENERATOR_API_VERSION } = require("gepa/plugins");

module.exports = defineDatasetGenerator({
  apiVersion: DATASET_GENERATOR_API_VERSION,
  kind: "dataset-generator",
  id: "examples.dataset-stream",
  name: "Dataset Stream Example",
  create() {
    return {
      generateOne(input, options) {
        return Promise.resolve().then(() => {
          options.events.emit({ type: "row-start", data: { index: input.index } });
          return {
            row: {
              id: `row-${input.index}`,
              value: "ok"
            },
            metadata: {
              row_index: input.index,
              mode: "async-demo"
            }
          };
        });
      }
    };
  }
});
```

## Config File

Create `dataset-config.yaml`:

```yaml
apiVersion: gepa.dataset-generate/v2
name: stream-demo
count: 2
prompting:
  user_template: "unused in this plugin"
validation:
  required_fields:
    - id
    - value
  max_retries: 0
  drop_invalid: false
```

## Commands

Dry-run stream check:

```bash
gepa-runner dataset generate \
  --script ./dataset-plugin.js \
  --config ./dataset-config.yaml \
  --stream \
  --dry-run
```

Persist outputs:

```bash
gepa-runner dataset generate \
  --script ./dataset-plugin.js \
  --config ./dataset-config.yaml \
  --stream \
  --output-dir ./out \
  --output-db ./generated.sqlite
```

## Expected Behavior

1. `stream-event ...` lines appear for each row while generation is running.
2. Final summary still prints generated count and output paths.
3. In `--dry-run`, no JSONL or sqlite output is written.

## Troubleshooting

| Problem | Cause | Solution |
|---|---|---|
| Generation fails with missing required fields | Plugin row shape does not satisfy `validation.required_fields` | Align plugin row keys with config validation keys |
| Stream lines missing for some rows | Plugin emits conditionally | Emit at deterministic lifecycle points in `generateOne` |
| Dry-run still expected files | `--dry-run` intentionally skips persistence | Remove `--dry-run` and provide `--output-dir` and/or `--output-db` |
| Async plugin hangs | Promise never settles | Ensure all code paths resolve or reject Promise |

## See Also

1. `glaze help gepa-runner-async-plugin-contract`
2. `glaze help gepa-runner-candidate-run-streaming-example`
