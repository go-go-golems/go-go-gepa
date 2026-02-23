package gepa

import "testing"

func TestConfigWithDefaults(t *testing.T) {
	got := (Config{}).withDefaults()

	if got.MaxEvalCalls != 200 {
		t.Fatalf("expected MaxEvalCalls=200, got %d", got.MaxEvalCalls)
	}
	if got.BatchSize != 8 {
		t.Fatalf("expected BatchSize=8, got %d", got.BatchSize)
	}
	if got.FrontierSize != 10 {
		t.Fatalf("expected FrontierSize=10, got %d", got.FrontierSize)
	}
	if got.RandomSeed == 0 {
		t.Fatalf("expected RandomSeed to be set")
	}
	if got.ReflectionSystemPrompt == "" {
		t.Fatalf("expected ReflectionSystemPrompt to be set")
	}
	if got.ReflectionPromptTemplate == "" {
		t.Fatalf("expected ReflectionPromptTemplate to be set")
	}
	if got.MergeSystemPrompt == "" {
		t.Fatalf("expected MergeSystemPrompt to be set")
	}
	if got.MergePromptTemplate == "" {
		t.Fatalf("expected MergePromptTemplate to be set")
	}
	if got.MergeScheduler != "probabilistic" {
		t.Fatalf("expected MergeScheduler=probabilistic, got %q", got.MergeScheduler)
	}
	if got.MaxMergesDue != 2 {
		t.Fatalf("expected MaxMergesDue=2, got %d", got.MaxMergesDue)
	}
	if got.ComponentSelector != "round_robin" {
		t.Fatalf("expected ComponentSelector=round_robin, got %q", got.ComponentSelector)
	}
	if got.Now == nil {
		t.Fatalf("expected Now function to be set")
	}
}

func TestConfigWithDefaultsClampsMergeProbability(t *testing.T) {
	got := (Config{
		MergeProbability: -1.0,
	}).withDefaults()
	if got.MergeProbability != 0 {
		t.Fatalf("expected MergeProbability to clamp to 0, got %f", got.MergeProbability)
	}
}
