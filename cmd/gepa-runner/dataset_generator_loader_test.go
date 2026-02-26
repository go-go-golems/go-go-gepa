package main

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	datasetgen "github.com/go-go-golems/go-go-gepa/pkg/dataset/generator"
)

type loaderTestRNG struct {
	rng *rand.Rand
}

func (r *loaderTestRNG) IntN(max int) int {
	if r == nil || r.rng == nil || max <= 0 {
		return 0
	}
	return r.rng.Intn(max)
}

func (r *loaderTestRNG) Float64() float64 {
	if r == nil || r.rng == nil {
		return 0
	}
	return r.rng.Float64()
}

func (r *loaderTestRNG) Choice(values []any) any {
	if r == nil || r.rng == nil || len(values) == 0 {
		return nil
	}
	return values[r.rng.Intn(len(values))]
}

func (r *loaderTestRNG) Shuffle(values []any) {
	if r == nil || r.rng == nil || len(values) < 2 {
		return
	}
	r.rng.Shuffle(len(values), func(i, j int) {
		values[i], values[j] = values[j], values[i]
	})
}

func TestLoadDatasetGeneratorPluginAndGenerateOne(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "generator.js")
	script := `
const { defineDatasetGenerator, DATASET_GENERATOR_API_VERSION } = require("gepa/plugins");

module.exports = defineDatasetGenerator({
  apiVersion: DATASET_GENERATOR_API_VERSION,
  kind: "dataset-generator",
  id: "example.generator",
  name: "Example Generator",
  create(hostContext) {
    return {
      generateOne(input, options) {
        const value = options.rng.intN(100);
        return {
          row: {
            index: input.index,
            value: value,
            registry: hostContext.pluginRegistryIdentifier
          },
          metadata: {
            seed: options.seed
          }
        };
      }
    };
  }
});
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}

	rt, err := newJSRuntime(tmpDir)
	if err != nil {
		t.Fatalf("newJSRuntime failed: %v", err)
	}
	defer rt.Close()

	hostContext := map[string]any{
		"app": "test",
	}
	plugin, meta, err := datasetgen.LoadPlugin(rt.vm, rt.reqMod, scriptPath, hostContext)
	if err != nil {
		t.Fatalf("LoadPlugin failed: %v", err)
	}
	if meta.RegistryIdentifier != datasetgen.DefaultRegistryIdentifier {
		t.Fatalf("expected default registry identifier %q, got %q", datasetgen.DefaultRegistryIdentifier, meta.RegistryIdentifier)
	}
	if got, _ := hostContext["pluginRegistryIdentifier"].(string); got != datasetgen.DefaultRegistryIdentifier {
		t.Fatalf("host context missing pluginRegistryIdentifier: %q", got)
	}

	rng := &loaderTestRNG{rng: rand.New(rand.NewSource(42))}
	row, metadata, err := plugin.GenerateOne(map[string]any{"index": 0}, datasetgen.PluginGenerateOptions{
		Seed: 42,
		RNG:  rng,
	})
	if err != nil {
		t.Fatalf("GenerateOne failed: %v", err)
	}
	if row["index"] != int64(0) && row["index"] != 0 {
		t.Fatalf("unexpected index value: %#v", row["index"])
	}
	if row["registry"] != datasetgen.DefaultRegistryIdentifier {
		t.Fatalf("unexpected registry value: %#v", row["registry"])
	}
	if metadata["seed"] != int64(42) && metadata["seed"] != float64(42) {
		t.Fatalf("unexpected metadata seed: %#v", metadata["seed"])
	}
}
