#!/usr/bin/env bash
set -euo pipefail

ROOT="/home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa"
TICKET_SCRIPTS="$ROOT/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/scripts"
CONFIG="$TICKET_SCRIPTS/exp-07-dataset-generate-gpt5nano.yaml"
REGISTRY="$TICKET_SCRIPTS/exp-07-profile-registry-gpt5nano.yaml"
GENERATOR="$ROOT/cmd/gepa-runner/scripts/arithmetic_dataset_generator.js"
DB="$TICKET_SCRIPTS/exp-07-generated.sqlite"
OUT1="$TICKET_SCRIPTS/exp-07-out-1"
OUT2="$TICKET_SCRIPTS/exp-07-out-2"
LOG1="$TICKET_SCRIPTS/exp-07-run-1.txt"
LOG2="$TICKET_SCRIPTS/exp-07-run-2.txt"
SQL_SUMMARY="$TICKET_SCRIPTS/exp-07-sql-summary.txt"

rm -f "$DB" "$LOG1" "$LOG2" "$SQL_SUMMARY"
rm -rf "$OUT1" "$OUT2"

cd "$ROOT"

go run ./cmd/gepa-runner dataset generate \
  --profile gpt-5-nano \
  --profile-registries "$REGISTRY" \
  --script "$GENERATOR" \
  --config "$CONFIG" \
  --count 3 \
  --output-dir "$OUT1" \
  --output-db "$DB" \
  > "$LOG1" 2>&1

go run ./cmd/gepa-runner dataset generate \
  --profile gpt-5-nano \
  --profile-registries "$REGISTRY" \
  --script "$GENERATOR" \
  --config "$CONFIG" \
  --count 5 \
  --output-dir "$OUT2" \
  --output-db "$DB" \
  > "$LOG2" 2>&1

{
  echo ".tables"
  echo "SELECT COUNT(*) AS dataset_count FROM gepa_generated_datasets;"
  echo "SELECT COUNT(*) AS row_count FROM gepa_generated_dataset_rows;"
  echo "SELECT dataset_id, name, requested_count, generated_count, seed, plugin_id, plugin_registry_identifier FROM gepa_generated_datasets ORDER BY created_at_ms;"
} | sqlite3 "$DB" > "$SQL_SUMMARY"
