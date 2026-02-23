package gepa

import (
	"context"
	"testing"

	"github.com/go-go-golems/geppetto/pkg/turns"
)

func TestParetoFrontAndDominance(t *testing.T) {
	a := ObjectiveScores{"accuracy": 0.9, "latency": -0.2}
	b := ObjectiveScores{"accuracy": 0.8, "latency": -0.3}
	if !Dominates(a, b) {
		t.Fatalf("expected a to dominate b")
	}

	points := []ObjectiveScores{
		{"accuracy": 0.9, "cost": -0.4}, // non-dominated
		{"accuracy": 0.8, "cost": -0.3}, // non-dominated
		{"accuracy": 0.7, "cost": -0.6}, // dominated by #0
	}
	front := ParetoFront(points)
	if len(front) != 2 {
		t.Fatalf("expected 2 points on pareto front, got %d", len(front))
	}
	if !containsInt(front, 0) || !containsInt(front, 1) {
		t.Fatalf("expected indices 0 and 1 on front, got %v", front)
	}
}

func TestAggregateStats(t *testing.T) {
	evals := []ExampleEval{
		{ExampleIndex: 0, Result: EvalResult{Score: 1, Objectives: ObjectiveScores{"score": 1, "accuracy": 1}}},
		{ExampleIndex: 1, Result: EvalResult{Score: 0, Objectives: ObjectiveScores{"score": 0, "accuracy": 0.5}}},
	}
	got := AggregateStats(evals)
	if got.N != 2 {
		t.Fatalf("expected N=2, got %d", got.N)
	}
	if got.MeanScore != 0.5 {
		t.Fatalf("expected MeanScore=0.5, got %f", got.MeanScore)
	}
	if got.MeanObjectives["accuracy"] != 0.75 {
		t.Fatalf("expected mean accuracy=0.75, got %f", got.MeanObjectives["accuracy"])
	}
}

type constantEngine struct {
	text string
}

func (e *constantEngine) RunInference(_ context.Context, _ *turns.Turn) (*turns.Turn, error) {
	return &turns.Turn{
		Blocks: []turns.Block{
			turns.NewAssistantTextBlock(e.text),
		},
	}, nil
}

func TestOptimizerStopsOnNoProgressAndReusesCache(t *testing.T) {
	examples := []any{
		map[string]any{"x": 1},
		map[string]any{"x": 2},
		map[string]any{"x": 3},
	}

	evalCalls := 0
	evalFn := func(_ context.Context, _ Candidate, _ int, _ any) (EvalResult, error) {
		evalCalls++
		return EvalResult{Score: 0}, nil
	}

	// Reflector always proposes exactly the same prompt as the seed.
	// After seed init eval, parent and child evaluations should hit cache only.
	reflector := &Reflector{
		Engine: &constantEngine{text: "```Base prompt```"},
	}

	cfg := Config{
		MaxEvalCalls: 12,
		BatchSize:    2,
		RandomSeed:   1234,
	}
	opt := NewOptimizer(cfg, evalFn, reflector)

	res, err := opt.Optimize(context.Background(), Candidate{"prompt": "Base prompt"}, examples)
	if err != nil {
		t.Fatalf("Optimize returned error: %v", err)
	}
	if res == nil {
		t.Fatalf("expected non-nil result")
	}

	if res.CallsUsed >= cfg.MaxEvalCalls {
		t.Fatalf("expected no-progress guard to break early, got calls=%d max=%d", res.CallsUsed, cfg.MaxEvalCalls)
	}
	if res.CallsUsed > len(examples) {
		t.Fatalf("expected same candidate to be cached after one eval per example, got calls=%d examples=%d", res.CallsUsed, len(examples))
	}
	if evalCalls != res.CallsUsed {
		t.Fatalf("expected evalCalls (%d) to match calls used (%d)", evalCalls, res.CallsUsed)
	}
}

func containsInt(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

func TestSelectComponentsRoundRobin(t *testing.T) {
	opt := &Optimizer{
		cfg:       Config{ComponentSelector: "round_robin"},
		paramKeys: []string{"a", "b"},
	}
	parent := &candidateNode{
		Candidate:      Candidate{"a": "x", "b": "y"},
		NextParamIndex: 0,
	}

	c1 := opt.selectComponents(parent)
	c2 := opt.selectComponents(parent)
	c3 := opt.selectComponents(parent)

	if len(c1) != 1 || c1[0] != "a" {
		t.Fatalf("expected first key to be a, got %v", c1)
	}
	if len(c2) != 1 || c2[0] != "b" {
		t.Fatalf("expected second key to be b, got %v", c2)
	}
	if len(c3) != 1 || c3[0] != "a" {
		t.Fatalf("expected third key to cycle back to a, got %v", c3)
	}
}

func TestSelectComponentsAll(t *testing.T) {
	opt := &Optimizer{
		cfg:       Config{ComponentSelector: "all"},
		paramKeys: []string{"a", "b"},
	}
	parent := &candidateNode{
		Candidate: Candidate{"a": "x", "b": "y"},
	}

	components := opt.selectComponents(parent)
	if len(components) != 2 {
		t.Fatalf("expected 2 components, got %v", components)
	}
	if components[0] != "a" || components[1] != "b" {
		t.Fatalf("expected [a b], got %v", components)
	}
}

func TestMergeChildMustBeatBestParent(t *testing.T) {
	examples := []any{map[string]any{"x": 1}}

	evalFn := func(_ context.Context, c Candidate, _ int, _ any) (EvalResult, error) {
		switch c["prompt"] {
		case "p0":
			return EvalResult{Score: 0.50}, nil
		case "p1":
			return EvalResult{Score: 1.00}, nil
		case "p06":
			return EvalResult{Score: 0.60}, nil
		default:
			return EvalResult{Score: 0.10}, nil
		}
	}

	reflector := &Reflector{
		Engine: &constantEngine{text: "```p1```"},
	}
	cfg := Config{
		MaxEvalCalls:     8,
		BatchSize:        1,
		RandomSeed:       7,
		MergeProbability: 1.0,
	}
	opt := NewOptimizer(cfg, evalFn, reflector)
	opt.SetMergeFunc(func(_ context.Context, _ MergeInput) (string, string, error) {
		return "p06", "merge-raw", nil
	})

	res, err := opt.Optimize(context.Background(), Candidate{"prompt": "p0"}, examples)
	if err != nil {
		t.Fatalf("Optimize returned error: %v", err)
	}
	if res == nil {
		t.Fatalf("expected non-nil result")
	}

	if !hasCandidatePrompt(res.Candidates, "p1") {
		t.Fatalf("expected accepted mutation candidate p1 in pool")
	}
	if hasCandidatePrompt(res.Candidates, "p06") {
		t.Fatalf("merge child p06 should be rejected because it does not beat best parent")
	}
}

func hasCandidatePrompt(cands []CandidateEntry, prompt string) bool {
	for _, c := range cands {
		if c.Candidate["prompt"] == prompt {
			return true
		}
	}
	return false
}
