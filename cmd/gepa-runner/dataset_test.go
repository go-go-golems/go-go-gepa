package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadSeedCandidateFileJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "seed.json")
	blob := `{
  "prompt": "Be concise",
  "retries": 3,
  "enabled": true,
  "nested": {"a": 1}
}`
	if err := os.WriteFile(path, []byte(blob), 0o644); err != nil {
		t.Fatalf("write seed file: %v", err)
	}

	cand, err := loadSeedCandidateFile(path)
	if err != nil {
		t.Fatalf("loadSeedCandidateFile returned error: %v", err)
	}

	if cand["prompt"] != "Be concise" {
		t.Fatalf("expected prompt value, got %q", cand["prompt"])
	}
	if cand["retries"] != "3" {
		t.Fatalf("expected retries to coerce to \"3\", got %q", cand["retries"])
	}
	if cand["enabled"] != "true" {
		t.Fatalf("expected enabled to coerce to \"true\", got %q", cand["enabled"])
	}
	if cand["nested"] != `{"a":1}` {
		t.Fatalf("expected nested object to coerce to compact JSON, got %q", cand["nested"])
	}
}

func TestLoadSeedCandidateFileYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "seed.yaml")
	blob := "prompt: YAML prompt\nretries: 7\n"
	if err := os.WriteFile(path, []byte(blob), 0o644); err != nil {
		t.Fatalf("write seed file: %v", err)
	}

	cand, err := loadSeedCandidateFile(path)
	if err != nil {
		t.Fatalf("loadSeedCandidateFile returned error: %v", err)
	}
	if cand["prompt"] != "YAML prompt" {
		t.Fatalf("expected prompt value, got %q", cand["prompt"])
	}
	if cand["retries"] != "7" {
		t.Fatalf("expected retries to coerce to \"7\", got %q", cand["retries"])
	}
}

func TestLoadSeedCandidateFileRejectsNonMap(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "seed.json")
	if err := os.WriteFile(path, []byte(`[1,2,3]`), 0o644); err != nil {
		t.Fatalf("write seed file: %v", err)
	}

	_, err := loadSeedCandidateFile(path)
	if err == nil {
		t.Fatalf("expected error for non-map seed candidate")
	}
	if !strings.Contains(err.Error(), "object/map") {
		t.Fatalf("expected object/map error, got: %v", err)
	}
}
