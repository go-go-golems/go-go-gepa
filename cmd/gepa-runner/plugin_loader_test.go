package main

import (
	"strings"
	"testing"

	"github.com/dop251/goja"
)

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
