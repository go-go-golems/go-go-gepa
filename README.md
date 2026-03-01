# gepa-runner

`gepa-runner` is a standalone GEPA-style optimizer CLI extracted from `geppetto`.
It keeps inference/tooling integration through geppetto's JS runtime while moving
optimization ownership into this repository.

## What it does

- Runs reflective prompt optimization (`optimize`) over JS-defined benchmark tasks.
- Runs one-shot benchmark evaluation (`eval`) for a fixed prompt/candidate.
- Supports optional multi-objective Pareto scoring from plugin results.
- Supports optional SQLite run recording (`--record`, `--record-db`).
- Uses a JS plugin contract helper via `./cmd/gepa-runner/scripts/lib/gepa_plugin_contract.js`.

## Repository layout

- `cmd/gepa-runner/`: CLI commands and JS runtime/plugin loader.
- `pkg/optimizer/gepa/`: GEPA-inspired optimizer primitives.
- `cmd/gepa-runner/scripts/`: starter evaluator plugins and smoke scripts.

Detailed command usage and plugin contract:

- `cmd/gepa-runner/README.md`

## Build

```bash
cd go-go-gepa
go test ./... -count=1
go build ./cmd/gepa-runner
```

## Quick start

```bash
cd go-go-gepa
go build -o ./gepa-runner ./cmd/gepa-runner
```

Optimize with the included toy script:

```bash
./gepa-runner optimize \
  --script ./cmd/gepa-runner/scripts/toy_math_optimizer.js \
  --seed "Answer the question. Respond with only the final answer." \
  --max-evals 50 \
  --batch-size 8 \
  --out-prompt ./best_prompt.txt \
  --out-report ./optimize_report.json \
  --profile 4o-mini
```

Evaluate a prompt:

```bash
./gepa-runner eval \
  --script ./cmd/gepa-runner/scripts/toy_math_optimizer.js \
  --prompt-file ./best_prompt.txt \
  --profile 4o-mini
```

## Development

```bash
cd go-go-gepa
make lint
make test
make build
```

Install local hooks:

```bash
lefthook install
```

## Notes

- This module intentionally depends on `github.com/go-go-golems/geppetto` for
  inference/runtime APIs.
- Geppetto keeps generic plugin helpers; GEPA-specific optimizer/runner code is
  owned here.

## License

MIT. See `LICENSE`.
