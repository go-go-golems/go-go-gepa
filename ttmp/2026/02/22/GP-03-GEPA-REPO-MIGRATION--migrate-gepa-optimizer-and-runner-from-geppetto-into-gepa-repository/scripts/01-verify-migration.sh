#!/usr/bin/env bash
set -euo pipefail

ROOT="${1:-/home/manuel/workspaces/2026-02-22/add-gepa-optimizer}"
TICKET_DIR="$ROOT/gepa/ttmp/2026/02/22/GP-03-GEPA-REPO-MIGRATION--migrate-gepa-optimizer-and-runner-from-geppetto-into-gepa-repository"
SOURCES_DIR="$TICKET_DIR/sources"

mkdir -p "$SOURCES_DIR"

echo "[1/5] geppetto reference scan"
{
  echo "# scan timestamp"
  date -Iseconds
  echo
  echo "# geppetto references to removed GEPA paths"
  (cd "$ROOT/geppetto" && rg -n "pkg/optimizer/gepa|cmd/gepa-runner" pkg cmd || true)
} >"$SOURCES_DIR/01-geppetto-reference-scan.txt"

echo "[2/5] geppetto full test"
(
  cd "$ROOT/geppetto"
  go test ./... -count=1
) >"$SOURCES_DIR/02-geppetto-go-test.txt" 2>&1

echo "[3/5] go-gepa-runner tests"
(
  cd "$ROOT/gepa/go-gepa-runner"
  go test ./... -count=1
) >"$SOURCES_DIR/03-go-gepa-runner-go-test.txt" 2>&1

echo "[4/5] go-gepa-runner build"
(
  cd "$ROOT/gepa/go-gepa-runner"
  go build ./cmd/gepa-runner
) >"$SOURCES_DIR/04-go-gepa-runner-go-build.txt" 2>&1

echo "[5/5] docmgr doctor (migration ticket)"
(
  cd "$ROOT"
  docmgr doctor --ticket GP-03-GEPA-REPO-MIGRATION --stale-after 30
) >"$SOURCES_DIR/05-gp03-doctor.txt" 2>&1

echo "done: artifacts written to $SOURCES_DIR"
