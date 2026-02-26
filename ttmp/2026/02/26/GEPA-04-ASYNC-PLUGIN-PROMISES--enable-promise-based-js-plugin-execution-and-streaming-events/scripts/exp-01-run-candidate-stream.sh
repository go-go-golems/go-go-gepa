#!/usr/bin/env bash
set -euo pipefail

ROOT="/home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa"
SCRIPTS="$ROOT/ttmp/2026/02/26/GEPA-04-ASYNC-PLUGIN-PROMISES--enable-promise-based-js-plugin-execution-and-streaming-events/scripts"

PLUGIN="$SCRIPTS/exp-01-candidate-stream-plugin.js"
CONFIG="$SCRIPTS/exp-01-candidate-config.yaml"
INPUT="$SCRIPTS/exp-01-candidate-input.json"
REGISTRY="$SCRIPTS/exp-01-profile-registry-gpt5nano.yaml"
LOG="$SCRIPTS/exp-01-candidate-stream-run.txt"

cd "$ROOT"
go run ./cmd/gepa-runner candidate run \
  --profile gpt-5-nano \
  --profile-registries "$REGISTRY" \
  --script "$PLUGIN" \
  --config "$CONFIG" \
  --input-file "$INPUT" \
  --stream \
  --output-format json \
  > "$LOG" 2>&1

echo "Wrote: $LOG"
