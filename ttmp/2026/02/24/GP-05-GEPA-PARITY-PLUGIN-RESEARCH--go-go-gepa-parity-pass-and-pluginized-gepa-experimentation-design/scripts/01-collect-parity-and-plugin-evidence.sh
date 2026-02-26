#!/usr/bin/env bash
set -euo pipefail

ROOT="/home/manuel/workspaces/2026-02-22/add-gepa-optimizer"
TICKET_DIR="$ROOT/gepa/ttmp/2026/02/24/GP-05-GEPA-PARITY-PLUGIN-RESEARCH--go-go-gepa-parity-pass-and-pluginized-gepa-experimentation-design"
OUT="$TICKET_DIR/sources"

mkdir -p "$OUT"

echo "# GP-05 evidence collection" > "$OUT/00-summary.txt"
echo "" >> "$OUT/00-summary.txt"
echo "- generated_at: $(date --iso-8601=seconds)" >> "$OUT/00-summary.txt"
echo "- root: $ROOT" >> "$OUT/00-summary.txt"

nl -ba "$ROOT/go-go-gepa/pkg/optimizer/gepa/optimizer.go" | sed -n '1,240p' > "$OUT/01-go-optimizer-hooks-and-types.txt"
nl -ba "$ROOT/go-go-gepa/pkg/optimizer/gepa/optimizer.go" | sed -n '240,620p' > "$OUT/02-go-optimizer-main-loop.txt"
nl -ba "$ROOT/go-go-gepa/pkg/optimizer/gepa/optimizer.go" | sed -n '620,1135p' > "$OUT/03-go-optimizer-selection-batching.txt"
nl -ba "$ROOT/go-go-gepa/pkg/optimizer/gepa/pareto.go" | sed -n '1,220p' > "$OUT/04-go-pareto.txt"
nl -ba "$ROOT/go-go-gepa/pkg/optimizer/gepa/config.go" | sed -n '1,260p' > "$OUT/05-go-config.txt"

nl -ba "$ROOT/go-go-gepa/cmd/gepa-runner/main.go" | sed -n '140,380p' > "$OUT/06-go-runner-wiring.txt"
nl -ba "$ROOT/go-go-gepa/cmd/gepa-runner/plugin_loader.go" | sed -n '1,700p' > "$OUT/07-go-plugin-loader-contract.txt"
nl -ba "$ROOT/go-go-gepa/cmd/gepa-runner/js_runtime.go" | sed -n '1,220p' > "$OUT/08-go-js-runtime.txt"
nl -ba "$ROOT/go-go-gepa/cmd/gepa-runner/README.md" | sed -n '1,280p' > "$OUT/09-go-plugin-docs.txt"
nl -ba "$ROOT/go-go-gepa/cmd/gepa-runner/scripts/lib/gepa_plugin_contract.js" | sed -n '1,220p' > "$OUT/10-go-plugin-contract-js.txt"

nl -ba "$ROOT/gepa/src/gepa/core/state.py" | sed -n '1,980p' > "$OUT/11-py-state.txt"
nl -ba "$ROOT/gepa/src/gepa/core/engine.py" | sed -n '1,620p' > "$OUT/12-py-engine.txt"
nl -ba "$ROOT/gepa/src/gepa/api.py" | sed -n '260,420p' > "$OUT/13-py-api-strategy-wiring.txt"
nl -ba "$ROOT/gepa/src/gepa/strategies/candidate_selector.py" | sed -n '1,260p' > "$OUT/14-py-candidate-selector.txt"
nl -ba "$ROOT/gepa/src/gepa/strategies/batch_sampler.py" | sed -n '1,320p' > "$OUT/15-py-batch-sampler.txt"
nl -ba "$ROOT/gepa/src/gepa/strategies/component_selector.py" | sed -n '1,260p' > "$OUT/16-py-component-selector.txt"
nl -ba "$ROOT/gepa/src/gepa/proposer/reflective_mutation/reflective_mutation.py" | sed -n '1,740p' > "$OUT/17-py-reflective-proposer.txt"
nl -ba "$ROOT/gepa/src/gepa/proposer/merge.py" | sed -n '1,700p' > "$OUT/18-py-merge-proposer.txt"
nl -ba "$ROOT/gepa/src/gepa/gepa_utils.py" | sed -n '1,260p' > "$OUT/19-py-pareto-utils.txt"
nl -ba "$ROOT/gepa/src/gepa/core/adapter.py" | sed -n '1,260p' > "$OUT/20-py-adapter-contract.txt"

(
  cd "$ROOT/gepa"
  docmgr doctor --ticket GP-05-GEPA-PARITY-PLUGIN-RESEARCH --stale-after 30
) > "$OUT/21-doctor-before-docs.txt" 2>&1 || true

echo "wrote evidence files under: $OUT"
