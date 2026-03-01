package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadCandidateRunConfigValid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "candidate-run.yaml")
	blob := `
apiVersion: gepa.candidate-run/v2
candidate:
  prompt: "Solve"
metadata:
  candidate_id: cand-1
  reflection_used: refl-1
  tags:
    suite: smoke
runtime:
  profile: default
  engine_overrides:
    temperature: 0.1
`
	if err := os.WriteFile(path, []byte(blob), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, raw, err := loadCandidateRunConfig(path)
	if err != nil {
		t.Fatalf("loadCandidateRunConfig failed: %v", err)
	}
	if cfg.APIVersion != candidateRunConfigAPIVersion {
		t.Fatalf("unexpected apiVersion: %q", cfg.APIVersion)
	}
	if cfg.Metadata.CandidateID != "cand-1" {
		t.Fatalf("unexpected candidate_id: %q", cfg.Metadata.CandidateID)
	}
	if !strings.Contains(raw, "apiVersion") {
		t.Fatalf("expected raw yaml to contain apiVersion")
	}
}

func TestLoadCandidateRunConfigRejectsForbiddenScriptKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "candidate-run.yaml")
	blob := `
apiVersion: gepa.candidate-run/v2
candidate:
  prompt: "x"
script: ./plugin.js
`
	if err := os.WriteFile(path, []byte(blob), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, _, err := loadCandidateRunConfig(path)
	if err == nil {
		t.Fatalf("expected error for forbidden script key")
	}
	if !strings.Contains(err.Error(), "not allowed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveCandidateRunConfigMergesCLIOverrides(t *testing.T) {
	cfg := candidateRunConfig{
		APIVersion: candidateRunConfigAPIVersion,
		Candidate: map[string]any{
			"prompt": "Solve",
		},
		Metadata: candidateRunMetadataConfig{
			CandidateID:    "cand-config",
			ReflectionUsed: "refl-config",
			Tags: map[string]string{
				"suite": "smoke",
			},
		},
	}

	resolved, err := resolveCandidateRunConfig(cfg, "", candidateRunResolveOptions{
		CandidateID:    "cand-cli",
		ReflectionUsed: "refl-cli",
		Tags:           "suite=regression,branch=main",
	})
	if err != nil {
		t.Fatalf("resolveCandidateRunConfig failed: %v", err)
	}

	if resolved.CandidateID != "cand-cli" {
		t.Fatalf("expected candidate id override, got %q", resolved.CandidateID)
	}
	if resolved.ReflectionUsed != "refl-cli" {
		t.Fatalf("expected reflection override, got %q", resolved.ReflectionUsed)
	}
	if resolved.Tags["suite"] != "regression" {
		t.Fatalf("expected suite tag override, got %q", resolved.Tags["suite"])
	}
	if resolved.Tags["branch"] != "main" {
		t.Fatalf("expected branch tag from CLI, got %q", resolved.Tags["branch"])
	}
}

func TestLoadCandidateRunInputFileJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.json")
	blob := `{"question":"2+2","answer":"4"}`
	if err := os.WriteFile(path, []byte(blob), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	m, err := loadCandidateRunInputFile(path)
	if err != nil {
		t.Fatalf("loadCandidateRunInputFile failed: %v", err)
	}
	if m["question"] != "2+2" {
		t.Fatalf("unexpected question: %#v", m["question"])
	}
}
