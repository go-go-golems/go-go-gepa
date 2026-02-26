#!/usr/bin/env bash
set -euo pipefail

ROOT="${1:-/home/manuel/workspaces/2026-02-22/add-gepa-optimizer}"
TICKET_DIR="$ROOT/gepa/ttmp/2026/02/23/GP-04-GEPA-MERGE-MULTIPARAM-ALIGN--integrate-gepa-merge-and-multi-parameter-optimization-in-go-go-gepa-and-align-with-python-original"
OUT="$TICKET_DIR/sources"

mkdir -p "$OUT"

echo "[1/6] imported merge/multi-param patch"
git -C "$ROOT/imported/geppetto-main" diff bb5b37b..7b488b9 \
  -- pkg/optimizer/gepa/config.go \
     pkg/optimizer/gepa/format.go \
     pkg/optimizer/gepa/optimizer.go \
     pkg/optimizer/gepa/reflector.go \
     cmd/gepa-runner/main.go \
     cmd/gepa-runner/dataset.go \
     cmd/gepa-runner/plugin_loader.go \
     cmd/gepa-runner/scripts/toy_math_optimizer.js \
     cmd/gepa-runner/README.md \
  > "$OUT/01-imported-merge-multiparam.patch"

echo "[2/6] file-level go-go-gepa vs imported diffs"
for f in \
  pkg/optimizer/gepa/config.go \
  pkg/optimizer/gepa/format.go \
  pkg/optimizer/gepa/optimizer.go \
  pkg/optimizer/gepa/reflector.go \
  cmd/gepa-runner/main.go \
  cmd/gepa-runner/dataset.go \
  cmd/gepa-runner/plugin_loader.go \
  cmd/gepa-runner/scripts/toy_math_optimizer.js
do
  out="$(echo "$f" | tr '/' '_').diff"
  git diff --no-index -- "$ROOT/imported/geppetto-main/$f" "$ROOT/go-go-gepa/$f" > "$OUT/02-$out" || true
done

echo "[3/6] python and go symbol index"
rg -n "class ReflectionConfig|class MergeConfig|class GEPAConfig|def optimize_anything\(|def optimize\(|class MergeProposer" \
  "$ROOT/gepa/src/gepa/optimize_anything.py" "$ROOT/gepa/src/gepa/api.py" "$ROOT/gepa/src/gepa/proposer/merge.py" \
  > "$OUT/03-python-symbol-index.txt"

rg -n "type Config struct|type Optimizer struct|type MergeInput|type MergeFunc|func \(o \*Optimizer\) Optimize|func \(o \*Optimizer\) SetMergeFunc|HasMerge|Merge\(" \
  "$ROOT/go-go-gepa/pkg/optimizer/gepa/"*.go "$ROOT/go-go-gepa/cmd/gepa-runner/"*.go \
  > "$OUT/04-go-symbol-index.txt"

echo "[4/6] python core excerpts"
nl -ba "$ROOT/gepa/src/gepa/optimize_anything.py" | sed -n '680,940p' > "$OUT/05-python-optimize_anything-config.txt"
nl -ba "$ROOT/gepa/src/gepa/optimize_anything.py" | sed -n '980,1160p' > "$OUT/06-python-optimize_anything-signature.txt"
nl -ba "$ROOT/gepa/src/gepa/optimize_anything.py" | sed -n '1400,1495p' > "$OUT/07-python-optimize_anything-merge-wiring.txt"
nl -ba "$ROOT/gepa/src/gepa/proposer/merge.py" | sed -n '200,420p' > "$OUT/08-python-merge-proposer-core.txt"
nl -ba "$ROOT/gepa/src/gepa/core/engine.py" | sed -n '380,560p' > "$OUT/09-python-engine-merge-flow.txt"
nl -ba "$ROOT/gepa/src/gepa/api.py" | sed -n '40,220p' > "$OUT/10-python-api-optimize-signature.txt"
nl -ba "$ROOT/gepa/src/gepa/strategies/component_selector.py" | sed -n '1,260p' > "$OUT/18-python-component-selector.txt"
nl -ba "$ROOT/gepa/src/gepa/proposer/reflective_mutation/reflective_mutation.py" | sed -n '70,360p' > "$OUT/19-python-reflective-mutation-components.txt"
nl -ba "$ROOT/gepa/src/gepa/adapters/optimize_anything_adapter/optimize_anything_adapter.py" | sed -n '480,620p' > "$OUT/20-python-optimize-anything-adapter-components.txt"

echo "[5/6] go current and imported excerpts"
nl -ba "$ROOT/go-go-gepa/pkg/optimizer/gepa/config.go" | sed -n '1,220p' > "$OUT/11-go-current-config.txt"
nl -ba "$ROOT/go-go-gepa/pkg/optimizer/gepa/optimizer.go" | sed -n '1,320p' > "$OUT/12-go-current-optimizer.txt"
nl -ba "$ROOT/go-go-gepa/cmd/gepa-runner/main.go" | sed -n '1,280p' > "$OUT/13-go-current-runner-main.txt"
nl -ba "$ROOT/imported/geppetto-main/pkg/optimizer/gepa/config.go" | sed -n '1,260p' > "$OUT/14-imported-new-config.txt"
nl -ba "$ROOT/imported/geppetto-main/pkg/optimizer/gepa/optimizer.go" | sed -n '1,380p' > "$OUT/15-imported-new-optimizer-part1.txt"
nl -ba "$ROOT/imported/geppetto-main/pkg/optimizer/gepa/optimizer.go" | sed -n '380,900p' > "$OUT/16-imported-new-optimizer-part2.txt"
nl -ba "$ROOT/imported/geppetto-main/cmd/gepa-runner/main.go" | sed -n '1,340p' > "$OUT/17-imported-new-runner-main.txt"
nl -ba "$ROOT/geppetto/pkg/js/modules/geppetto/plugins_module.go" | sed -n '1,240p' > "$OUT/21-geppetto-plugins-module.txt"
nl -ba "$ROOT/go-go-gepa/cmd/gepa-runner/plugin_loader.go" | sed -n '1,420p' > "$OUT/22-go-plugin-loader-current.txt"
nl -ba "$ROOT/imported/geppetto-main/cmd/gepa-runner/plugin_loader.go" | sed -n '1,520p' > "$OUT/23-imported-plugin-loader-with-merge.txt"

echo "[6/6] ticket doctor"
(cd "$ROOT" && docmgr doctor --ticket GP-04-GEPA-MERGE-MULTIPARAM-ALIGN --stale-after 30) > "$OUT/24-gp04-doctor.txt" 2>&1

echo "done: artifacts written to $OUT"
