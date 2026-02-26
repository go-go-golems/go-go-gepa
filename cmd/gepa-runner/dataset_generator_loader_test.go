package main

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

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
	plugin, meta, err := loadDatasetGeneratorPlugin(rt, scriptPath, hostContext)
	if err != nil {
		t.Fatalf("loadDatasetGeneratorPlugin failed: %v", err)
	}
	if meta.RegistryIdentifier != defaultPluginRegistryIdentifier {
		t.Fatalf("expected default registry identifier %q, got %q", defaultPluginRegistryIdentifier, meta.RegistryIdentifier)
	}
	if got, _ := hostContext["pluginRegistryIdentifier"].(string); got != defaultPluginRegistryIdentifier {
		t.Fatalf("host context missing pluginRegistryIdentifier: %q", got)
	}

	rng := &jsRNG{rng: rand.New(rand.NewSource(42))}
	row, metadata, err := plugin.GenerateOne(map[string]any{"index": 0}, datasetGeneratorGenerateOptions{
		Seed: 42,
		RNG:  rng,
	})
	if err != nil {
		t.Fatalf("GenerateOne failed: %v", err)
	}
	if row["index"] != int64(0) && row["index"] != 0 {
		t.Fatalf("unexpected index value: %#v", row["index"])
	}
	if row["registry"] != defaultPluginRegistryIdentifier {
		t.Fatalf("unexpected registry value: %#v", row["registry"])
	}
	if metadata["seed"] != int64(42) && metadata["seed"] != float64(42) {
		t.Fatalf("unexpected metadata seed: %#v", metadata["seed"])
	}
}
