package gepa

import "time"

// Config controls the GEPA-style optimization loop.
type Config struct {
	// Maximum number of evaluator calls. In prompt optimization, a “rollout”
	// usually corresponds to one (candidate, example) evaluation.
	MaxEvalCalls int

	// BatchSize is the number of examples sampled per iteration for mutation testing.
	BatchSize int

	// FrontierSize caps the number of candidates kept on the Pareto frontier
	// in the single-objective fallback path.
	FrontierSize int

	// RandomSeed controls stochastic selection and batching.
	// If 0, the optimizer will use time-based entropy.
	RandomSeed int64

	// ReflectionSystemPrompt is the system prompt used for the reflection LLM.
	ReflectionSystemPrompt string

	// ReflectionPromptTemplate must include "<curr_param>" and "<side_info>" placeholders.
	ReflectionPromptTemplate string

	// MergeProbability is the probability of attempting a merge (crossover) step
	// instead of a standard reflective mutation.
	//
	// 0 disables merging (default).
	MergeProbability float64

	// MergeSystemPrompt is the system prompt used for the merge proposer LLM.
	// If empty, the ReflectionSystemPrompt (or the reflector default) is used.
	MergeSystemPrompt string

	// MergePromptTemplate must include "<param_a>", "<param_b>",
	// "<side_info_a>", and "<side_info_b>" placeholders.
	MergePromptTemplate string

	// MergeScheduler controls when merge should be attempted.
	//
	// Supported values:
	//  - "probabilistic" (default): use MergeProbability each iteration.
	//  - "stagnation_due": maintain a merges_due counter that increases when no child is accepted,
	//    and spend due merge attempts when possible.
	//
	// Any other value falls back to "probabilistic".
	MergeScheduler string

	// MaxMergesDue caps the internal merges_due counter used by MergeScheduler=stagnation_due.
	MaxMergesDue int

	// OptimizableKeys optionally restricts which candidate keys are eligible for mutation/merge.
	//
	// If empty, the optimizer defaults to using the keys present in the seed candidate.
	//
	// This is the primary mechanism to support multi-parameter optimization (e.g., multiple
	// module prompts in a larger system).
	OptimizableKeys []string

	// ComponentSelector controls how many and which keys are updated per iteration.
	//
	// Supported values:
	//  - "round_robin" (default): update one key per iteration, cycling through OptimizableKeys.
	//  - "all": update all OptimizableKeys every iteration.
	//
	// Any other value falls back to "round_robin".
	ComponentSelector string

	// Objective is an optional natural-language description of what we are optimizing for.
	Objective string

	// MaxSideInfoChars caps the amount of formatted side-info passed to the reflector.
	// 0 means “no explicit cap”.
	MaxSideInfoChars int

	// Epsilon is the minimum improvement required to accept a child over its parent
	// in the single-objective setting. 0 is fine for most use.
	Epsilon float64

	// Now is injectable for tests. If nil, time.Now is used.
	Now func() time.Time
}

func (c Config) withDefaults() Config {
	out := c
	if out.MaxEvalCalls <= 0 {
		out.MaxEvalCalls = 200
	}
	if out.BatchSize <= 0 {
		out.BatchSize = 8
	}
	if out.FrontierSize <= 0 {
		out.FrontierSize = 10
	}
	if out.RandomSeed == 0 {
		out.RandomSeed = time.Now().UnixNano()
	}
	if out.ReflectionSystemPrompt == "" {
		out.ReflectionSystemPrompt = "You are an expert prompt engineer."
	}
	if out.ReflectionPromptTemplate == "" {
		out.ReflectionPromptTemplate = DefaultReflectionPromptTemplate
	}
	if out.MergeProbability < 0 {
		out.MergeProbability = 0
	}
	if out.MergeSystemPrompt == "" {
		out.MergeSystemPrompt = out.ReflectionSystemPrompt
	}
	if out.MergePromptTemplate == "" {
		out.MergePromptTemplate = DefaultMergePromptTemplate
	}
	if out.MergeScheduler == "" {
		out.MergeScheduler = "probabilistic"
	}
	if out.MaxMergesDue <= 0 {
		out.MaxMergesDue = 2
	}
	if out.ComponentSelector == "" {
		out.ComponentSelector = "round_robin"
	}
	if out.Now == nil {
		out.Now = time.Now
	}
	return out
}
