#!/usr/bin/env bash
set -euo pipefail

ROOT="/home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa"
SCRIPTS="$ROOT/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/scripts"
REGISTRY="$SCRIPTS/exp-07-profile-registry-gpt5nano.yaml"
PLUGIN="$SCRIPTS/exp-11-coaching-dataset-generator.js"
CONFIG="$SCRIPTS/exp-11-coaching-dataset-config.yaml"
OUT_DIR="$SCRIPTS/exp-11-out"
DB="$SCRIPTS/exp-11-generated.sqlite"
LOG="$SCRIPTS/exp-11-run.txt"
SQL_SUMMARY="$SCRIPTS/exp-11-sql-summary.txt"
ROW_SUMMARY="$SCRIPTS/exp-11-row-summary.txt"

rm -rf "$OUT_DIR"
rm -f "$DB" "$LOG" "$SQL_SUMMARY" "$ROW_SUMMARY"

cd "$ROOT"
go run ./cmd/gepa-runner dataset generate \
  --profile gpt-5-nano \
  --profile-registries "$REGISTRY" \
  --script "$PLUGIN" \
  --config "$CONFIG" \
  --count 2 \
  --output-dir "$OUT_DIR" \
  --output-db "$DB" \
  > "$LOG" 2>&1

{
  echo ".tables"
  echo "SELECT COUNT(*) AS dataset_count FROM gepa_generated_datasets;"
  echo "SELECT COUNT(*) AS row_count FROM gepa_generated_dataset_rows;"
  echo "SELECT dataset_id, name, generated_count, seed, plugin_id FROM gepa_generated_datasets ORDER BY created_at_ms DESC;"
} | sqlite3 "$DB" > "$SQL_SUMMARY"

jq -c '{case_id: .case_id, sessions: (.transcript|length), entity_count: (.ground_truth.entities|length), relationship_count: (.ground_truth.relationships|length)}' "$OUT_DIR/coaching-entity-sentiment-small.jsonl" > "$ROW_SUMMARY"
