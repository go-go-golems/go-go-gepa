#!/usr/bin/env bash
set -euo pipefail

ROOT="/home/manuel/workspaces/2026-02-22/add-gepa-optimizer/go-go-gepa"
SCRIPTS="$ROOT/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/scripts"
REGISTRY="$SCRIPTS/exp-07-profile-registry-gpt5nano.yaml"
PLUGIN="$SCRIPTS/exp-10-candidate-run-plugin.js"
CONFIG="$SCRIPTS/exp-10-candidate-run-config.yaml"
INPUT="$SCRIPTS/exp-10-candidate-run-input.json"
DB="$SCRIPTS/exp-10-candidate-runs.sqlite"
OUT_JSON="$SCRIPTS/exp-10-run-result.json"
LOG="$SCRIPTS/exp-10-run.txt"
SQL_SUMMARY="$SCRIPTS/exp-10-sql-summary.txt"

rm -f "$DB" "$OUT_JSON" "$LOG" "$SQL_SUMMARY"

cd "$ROOT"
go run ./cmd/gepa-runner candidate run \
  --profile gpt-5-nano \
  --profile-registries "$REGISTRY" \
  --script "$PLUGIN" \
  --config "$CONFIG" \
  --input-file "$INPUT" \
  --output-format json \
  --record \
  --record-db "$DB" \
  --out-result "$OUT_JSON" \
  > "$LOG" 2>&1

{
  echo ".tables"
  echo "SELECT COUNT(*) AS run_count FROM gepa_candidate_runs;"
  echo "SELECT run_id, plugin_id, candidate_id, reflection_used, status FROM gepa_candidate_runs ORDER BY timestamp_ms DESC;"
} | sqlite3 "$DB" > "$SQL_SUMMARY"
