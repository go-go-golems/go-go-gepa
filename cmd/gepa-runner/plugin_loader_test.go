package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dop251/goja"
	gepaopt "github.com/go-go-golems/go-go-gepa/pkg/optimizer/gepa"
)

func TestDecodeOptimizerPluginMetaDefaultsRegistryIdentifier(t *testing.T) {
	vm := goja.New()
	descriptor := vm.NewObject()
	_ = descriptor.Set("apiVersion", optimizerPluginAPIVersion)
	_ = descriptor.Set("kind", "optimizer")
	_ = descriptor.Set("id", "example.optimizer")
	_ = descriptor.Set("name", "Example Optimizer")
	_ = descriptor.Set("create", func(goja.FunctionCall) goja.Value { return goja.Undefined() })

	meta, err := decodeOptimizerPluginMeta(descriptor)
	if err != nil {
		t.Fatalf("decodeOptimizerPluginMeta returned error: %v", err)
	}
	if meta.RegistryIdentifier != defaultPluginRegistryIdentifier {
		t.Fatalf("expected default registry identifier %q, got %q", defaultPluginRegistryIdentifier, meta.RegistryIdentifier)
	}
}

func TestDecodeOptimizerPluginMetaUsesExplicitRegistryIdentifier(t *testing.T) {
	vm := goja.New()
	descriptor := vm.NewObject()
	_ = descriptor.Set("apiVersion", optimizerPluginAPIVersion)
	_ = descriptor.Set("kind", "optimizer")
	_ = descriptor.Set("id", "example.optimizer")
	_ = descriptor.Set("name", "Example Optimizer")
	_ = descriptor.Set("registryIdentifier", "registry.example/optimizer")
	_ = descriptor.Set("create", func(goja.FunctionCall) goja.Value { return goja.Undefined() })

	meta, err := decodeOptimizerPluginMeta(descriptor)
	if err != nil {
		t.Fatalf("decodeOptimizerPluginMeta returned error: %v", err)
	}
	if meta.RegistryIdentifier != "registry.example/optimizer" {
		t.Fatalf("unexpected registry identifier: %q", meta.RegistryIdentifier)
	}
}

func TestLoadOptimizerPluginInjectsRegistryIdentifierIntoHostContext(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "plugin.js")
	script := `
const { defineOptimizerPlugin, OPTIMIZER_PLUGIN_API_VERSION } = require("gepa/plugins");

module.exports = defineOptimizerPlugin({
  apiVersion: OPTIMIZER_PLUGIN_API_VERSION,
  kind: "optimizer",
  id: "example.optimizer",
  name: "Example Optimizer",
  registryIdentifier: "registry.example/optimizer",
  create(hostContext) {
    return {
      evaluate() {
        return {
          score: 1,
          objectives: { score: 1 },
          feedback: "ok",
          output: { registry: hostContext.pluginRegistryIdentifier }
        };
      },
      dataset() {
        return [{ prompt: "example" }];
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
	plugin, meta, err := loadOptimizerPlugin(rt, scriptPath, hostContext)
	if err != nil {
		t.Fatalf("loadOptimizerPlugin failed: %v", err)
	}
	if meta.RegistryIdentifier != "registry.example/optimizer" {
		t.Fatalf("unexpected plugin registry identifier in meta: %q", meta.RegistryIdentifier)
	}
	if got, _ := hostContext["pluginRegistryIdentifier"].(string); got != "registry.example/optimizer" {
		t.Fatalf("hostContext pluginRegistryIdentifier mismatch: %q", got)
	}

	res, err := plugin.Evaluate(gepaopt.Candidate{"prompt": "test"}, 0, map[string]any{"prompt": "sample"}, pluginEvaluateOptions{})
	if err != nil {
		t.Fatalf("plugin Evaluate failed: %v", err)
	}
	output, ok := res.Output.(map[string]any)
	if !ok {
		t.Fatalf("expected output map, got %T", res.Output)
	}
	if got, _ := output["registry"].(string); got != "registry.example/optimizer" {
		t.Fatalf("expected registry in output, got %q", got)
	}
}

func TestDecodeMergeOutputString(t *testing.T) {
	got, err := decodeMergeOutput(" merged prompt ", "prompt")
	if err != nil {
		t.Fatalf("decodeMergeOutput returned error: %v", err)
	}
	if got != "merged prompt" {
		t.Fatalf("unexpected merged output: %q", got)
	}
}

func TestDecodeMergeOutputMapParamKey(t *testing.T) {
	got, err := decodeMergeOutput(map[string]any{
		"instruction": "new value",
	}, "instruction")
	if err != nil {
		t.Fatalf("decodeMergeOutput returned error: %v", err)
	}
	if got != "new value" {
		t.Fatalf("unexpected merged output: %q", got)
	}
}

func TestDecodeMergeOutputCandidateMap(t *testing.T) {
	got, err := decodeMergeOutput(map[string]any{
		"candidate": map[string]any{
			"prompt": "merged from candidate",
		},
	}, "prompt")
	if err != nil {
		t.Fatalf("decodeMergeOutput returned error: %v", err)
	}
	if got != "merged from candidate" {
		t.Fatalf("unexpected merged output: %q", got)
	}
}

func TestDecodeMergeOutputMissingKey(t *testing.T) {
	_, err := decodeMergeOutput(map[string]any{
		"foo": "bar",
	}, "prompt")
	if err == nil {
		t.Fatalf("expected decodeMergeOutput to return error")
	}
	if !strings.Contains(err.Error(), "must return a string or an object containing") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOptimizerPluginHasMerge(t *testing.T) {
	p := &optimizerPlugin{}
	if p.HasMerge() {
		t.Fatalf("expected HasMerge false when mergeFn is nil")
	}

	p.mergeFn = func(_ goja.Value, _ ...goja.Value) (goja.Value, error) {
		return goja.Undefined(), nil
	}
	if !p.HasMerge() {
		t.Fatalf("expected HasMerge true when mergeFn is set")
	}
}

func TestDecodeCandidateFromString(t *testing.T) {
	cand, err := decodeCandidate(" hello ")
	if err != nil {
		t.Fatalf("decodeCandidate returned error: %v", err)
	}
	if cand["prompt"] != "hello" {
		t.Fatalf("expected prompt to be trimmed, got %q", cand["prompt"])
	}
}

func TestDecodeCandidateFromMap(t *testing.T) {
	cand, err := decodeCandidate(map[string]any{
		"prompt": "seed",
		"tries":  3,
	})
	if err != nil {
		t.Fatalf("decodeCandidate returned error: %v", err)
	}
	if cand["prompt"] != "seed" {
		t.Fatalf("expected prompt=seed, got %q", cand["prompt"])
	}
	if cand["tries"] != "3" {
		t.Fatalf("expected tries to coerce to string, got %q", cand["tries"])
	}
}

func TestDecodeStringList(t *testing.T) {
	got, err := decodeStringList([]any{" prompt ", "instruction", ""})
	if err != nil {
		t.Fatalf("decodeStringList returned error: %v", err)
	}
	if len(got) != 2 || got[0] != "prompt" || got[1] != "instruction" {
		t.Fatalf("unexpected list: %v", got)
	}
}

func TestDecodeSideInfoOutput(t *testing.T) {
	got, err := decodeSideInfoOutput(map[string]any{"sideInfo": "abc"})
	if err != nil {
		t.Fatalf("decodeSideInfoOutput returned error: %v", err)
	}
	if got != "abc" {
		t.Fatalf("expected abc, got %q", got)
	}
}

func TestOptimizerPluginHookPresence(t *testing.T) {
	p := &optimizerPlugin{}
	if p.HasInitialCandidate() {
		t.Fatalf("expected HasInitialCandidate false when fn is nil")
	}
	if p.HasSelectComponents() {
		t.Fatalf("expected HasSelectComponents false when fn is nil")
	}
	if p.HasComponentSideInfo() {
		t.Fatalf("expected HasComponentSideInfo false when fn is nil")
	}

	p.initialCandidateFn = func(_ goja.Value, _ ...goja.Value) (goja.Value, error) { return goja.Undefined(), nil }
	p.selectComponentsFn = func(_ goja.Value, _ ...goja.Value) (goja.Value, error) { return goja.Undefined(), nil }
	p.componentSideInfoFn = func(_ goja.Value, _ ...goja.Value) (goja.Value, error) { return goja.Undefined(), nil }

	if !p.HasInitialCandidate() {
		t.Fatalf("expected HasInitialCandidate true when fn is set")
	}
	if !p.HasSelectComponents() {
		t.Fatalf("expected HasSelectComponents true when fn is set")
	}
	if !p.HasComponentSideInfo() {
		t.Fatalf("expected HasComponentSideInfo true when fn is set")
	}
}

func TestSelectComponentsMethodFailsWithoutRuntime(t *testing.T) {
	p := &optimizerPlugin{selectComponentsFn: func(_ goja.Value, _ ...goja.Value) (goja.Value, error) {
		return goja.Undefined(), nil
	}}
	_, err := p.SelectComponents(gepaopt.ComponentSelectionInput{}, pluginEvaluateOptions{})
	if err == nil {
		t.Fatalf("expected error for uninitialized runtime")
	}
}
