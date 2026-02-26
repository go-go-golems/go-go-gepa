package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	datasetgen "github.com/go-go-golems/go-go-gepa/pkg/dataset/generator"
)

func TestLoadDatasetGenerateConfigValid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dataset.yaml")
	blob := `
apiVersion: gepa.dataset-generate/v2
name: arithmetic
count: 3
seed: 42
prompting:
  system: "You generate rows"
  user_template: "Generate {{difficulty}}"
  variables:
    difficulty:
      - easy
      - hard
validation:
  required_fields: ["question", "answer"]
  max_retries: 2
  drop_invalid: true
`
	if err := os.WriteFile(path, []byte(blob), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, raw, err := datasetgen.LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.APIVersion != datasetgen.ConfigAPIVersion {
		t.Fatalf("unexpected apiVersion: %q", cfg.APIVersion)
	}
	if cfg.Count != 3 {
		t.Fatalf("unexpected count: %d", cfg.Count)
	}
	if cfg.Seed == nil || *cfg.Seed != 42 {
		t.Fatalf("unexpected seed: %v", cfg.Seed)
	}
	if !strings.Contains(raw, "apiVersion") {
		t.Fatalf("expected raw yaml to contain apiVersion")
	}
}

func TestLoadDatasetGenerateConfigRejectsForbiddenScriptKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dataset.yaml")
	blob := `
apiVersion: gepa.dataset-generate/v2
count: 2
script: ./scripts/generator.js
`
	if err := os.WriteFile(path, []byte(blob), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	_, _, err := datasetgen.LoadConfig(path)
	if err == nil {
		t.Fatalf("expected error for forbidden script key")
	}
	if !strings.Contains(err.Error(), "not allowed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadDatasetGenerateConfigRejectsForbiddenOutputSection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dataset.yaml")
	blob := `
apiVersion: gepa.dataset-generate/v2
count: 2
output:
  dir: ./out
`
	if err := os.WriteFile(path, []byte(blob), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	_, _, err := datasetgen.LoadConfig(path)
	if err == nil {
		t.Fatalf("expected error for forbidden output key")
	}
	if !strings.Contains(err.Error(), "output") {
		t.Fatalf("unexpected error: %v", err)
	}
}
