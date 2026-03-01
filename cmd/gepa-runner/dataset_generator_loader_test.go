package main

import (
	"context"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	datasetgen "github.com/go-go-golems/go-go-gepa/pkg/dataset/generator"
	"github.com/go-go-golems/go-go-gepa/pkg/jsbridge"
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
	plugin, meta, err := datasetgen.LoadPlugin(rt.vm, rt.runner, rt.reqMod, scriptPath, hostContext)
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
	row, metadata, err := plugin.GenerateOne(context.Background(), map[string]any{"index": 0}, datasetgen.PluginGenerateOptions{
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

func TestLoadDatasetGeneratorPluginSupportsPromiseGenerateAndStreaming(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "generator-promise.js")
	script := `
const { defineDatasetGenerator, DATASET_GENERATOR_API_VERSION } = require("gepa/plugins");

module.exports = defineDatasetGenerator({
  apiVersion: DATASET_GENERATOR_API_VERSION,
  kind: "dataset-generator",
  id: "example.generator.promise",
  name: "Example Generator Promise",
  create() {
    return {
      generateOne(input, options) {
        return Promise.resolve().then(() => {
          if (options && typeof options.emitEvent === "function") {
            options.emitEvent({ type: "row-start", data: { index: input.index } });
          }
          if (options && options.events && typeof options.events.emit === "function") {
            options.events.emit({ type: "row-progress", message: "building row" });
          }
          return {
            row: {
              index: input.index,
              value: "ok"
            },
            metadata: {
              mode: "async"
            }
          };
        });
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

	plugin, _, err := datasetgen.LoadPlugin(rt.vm, rt.runner, rt.reqMod, scriptPath, map[string]any{"app": "test"})
	if err != nil {
		t.Fatalf("LoadPlugin failed: %v", err)
	}

	eventCh := make(chan jsbridge.Event, 8)
	row, metadata, err := plugin.GenerateOne(context.Background(), map[string]any{"index": 1}, datasetgen.PluginGenerateOptions{
		EventSink: func(event jsbridge.Event) {
			eventCh <- event
		},
	})
	if err != nil {
		t.Fatalf("GenerateOne failed: %v", err)
	}
	if row["value"] != "ok" {
		t.Fatalf("unexpected row value: %#v", row["value"])
	}
	if metadata["mode"] != "async" {
		t.Fatalf("unexpected metadata mode: %#v", metadata["mode"])
	}

	events := make([]jsbridge.Event, 0, 2)
	deadline := time.After(1 * time.Second)
	for len(events) < 2 {
		select {
		case event := <-eventCh:
			events = append(events, event)
		case <-deadline:
			t.Fatalf("timed out waiting for streamed events, got=%d", len(events))
		}
	}
	if events[0].PluginMethod != "generateOne" || events[1].PluginMethod != "generateOne" {
		t.Fatalf("expected plugin method generateOne in streamed events")
	}
	if events[0].Sequence != 1 || events[1].Sequence != 2 {
		t.Fatalf("expected event sequences 1,2 got %d,%d", events[0].Sequence, events[1].Sequence)
	}
}

func TestLoadDatasetGeneratorGenerateOneHonorsCanceledContext(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "generator-pending.js")
	script := `
const { defineDatasetGenerator, DATASET_GENERATOR_API_VERSION } = require("gepa/plugins");

module.exports = defineDatasetGenerator({
  apiVersion: DATASET_GENERATOR_API_VERSION,
  kind: "dataset-generator",
  id: "example.generator.pending",
  name: "Example Generator Pending",
  create() {
    return {
      generateOne() {
        return new Promise(() => {});
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

	plugin, _, err := datasetgen.LoadPlugin(rt.vm, rt.runner, rt.reqMod, scriptPath, map[string]any{"app": "test"})
	if err != nil {
		t.Fatalf("LoadPlugin failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()

	_, _, err = plugin.GenerateOne(ctx, map[string]any{"index": 1}, datasetgen.PluginGenerateOptions{})
	if err == nil {
		t.Fatalf("expected canceled-context error")
	}
	if !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("expected context canceled in error, got: %v", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("expected fast cancellation, took %s", elapsed)
	}
}
