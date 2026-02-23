package gepa

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"
)

// EvaluateFunc evaluates a candidate on a single example.
// The exampleIndex is provided for caching and traceability.
type EvaluateFunc func(ctx context.Context, candidate Candidate, exampleIndex int, example any) (EvalResult, error)

// MergeInput is passed to an optional MergeFunc hook.
// It is intentionally string-first (for the prompt-like primary param), while still
// providing full candidate maps for advanced merge logic.
type MergeInput struct {
	ParentA   Candidate `json:"parent_a"`
	ParentB   Candidate `json:"parent_b"`
	ParamKey  string    `json:"param_key"`
	ParamA    string    `json:"param_a"`
	ParamB    string    `json:"param_b"`
	SideInfoA string    `json:"side_info_a"`
	SideInfoB string    `json:"side_info_b"`
}

// MergeFunc can be provided by the caller to override how merges are performed.
// It should return a new value for the optimizable parameter, plus a raw/debug string.
//
// If nil, the optimizer falls back to Reflector.Merge.
type MergeFunc func(ctx context.Context, in MergeInput) (mergedParam string, raw string, err error)

// Optimizer runs a GEPA-style reflective evolutionary loop.
//
// This is intentionally “GEPA-inspired” rather than a 1:1 port of the Python reference.
// The goal is a small, composable core that can sit on top of Geppetto and JS evaluators.
type Optimizer struct {
	cfg       Config
	eval      EvaluateFunc
	reflector *Reflector
	mergeFn   MergeFunc

	rng       *rand.Rand
	cache     map[string]map[int]EvalResult // candidateHash -> exampleIndex -> result
	callsUsed int

	// pool is append-only.
	pool []*candidateNode

	// paramKeys are the ordered candidate keys that can be optimized.
	// Determined at Optimize() time from cfg.OptimizableKeys or the seed candidate.
	paramKeys []string
}

type candidateNode struct {
	ID        CandidateID
	ParentID  CandidateID
	Parent2ID CandidateID
	Candidate Candidate
	CreatedAt time.Time
	Operation string

	// NextParamIndex tracks which parameter will be updated next under the
	// "round_robin" component selector.
	NextParamIndex int

	// LastUpdated records, per candidate key, the node id that last updated that key.
	// This enables simple system-aware merges that can pick the newest version of each key
	// from multiple parents.
	LastUpdated map[string]CandidateID

	// UpdatedKeys records which candidate keys were modified to create this node.
	UpdatedKeys []string

	// Evaluations is the subset of cached evaluations we’ve pulled into the node.
	// The optimizer cache is the source of truth; this field is a convenience view.
	Evaluations map[int]EvalResult

	// ReflectionRaw is the full raw LLM response from the last mutation step (debug).
	ReflectionRaw string
}

type Result struct {
	BestCandidate Candidate        `json:"best_candidate"`
	BestStats     CandidateStats   `json:"best_stats"`
	CallsUsed     int              `json:"calls_used"`
	Candidates    []CandidateEntry `json:"candidates"`
}

type CandidateEntry struct {
	ID            int            `json:"id"`
	ParentID      int            `json:"parent_id"`
	Parent2ID     int            `json:"parent2_id"` // optional (only set for merge children)
	Operation     string         `json:"operation"`  // seed|mutate|merge
	Hash          string         `json:"hash"`
	CreatedAt     time.Time      `json:"created_at"`
	Candidate     Candidate      `json:"candidate"`
	GlobalStats   CandidateStats `json:"global_stats"`
	EvalsCached   int            `json:"evals_cached"`
	ReflectionRaw string         `json:"reflection_raw,omitempty"`
	UpdatedKeys   []string       `json:"updated_keys,omitempty"`
}

// NewOptimizer constructs an optimizer.
func NewOptimizer(cfg Config, eval EvaluateFunc, reflector *Reflector) *Optimizer {
	c := cfg.withDefaults()
	r := rand.New(rand.NewSource(c.RandomSeed))
	// Plumb config defaults into the reflector unless the caller overrides.
	if reflector != nil {
		if reflector.System == "" {
			reflector.System = c.ReflectionSystemPrompt
		}
		if reflector.Template == "" {
			reflector.Template = c.ReflectionPromptTemplate
		}
		if reflector.MergeSystem == "" {
			reflector.MergeSystem = c.MergeSystemPrompt
		}
		if reflector.MergeTemplate == "" {
			reflector.MergeTemplate = c.MergePromptTemplate
		}
	}
	return &Optimizer{
		cfg:       c,
		eval:      eval,
		reflector: reflector,
		rng:       r,
		cache:     map[string]map[int]EvalResult{},
	}
}

// SetMergeFunc installs an optional merge hook.
// If not set (or set to nil), merges use Reflector.Merge.
func (o *Optimizer) SetMergeFunc(fn MergeFunc) {
	if o == nil {
		return
	}
	o.mergeFn = fn
}

// CallsUsed returns the number of evaluator calls consumed so far.
func (o *Optimizer) CallsUsed() int {
	if o == nil {
		return 0
	}
	return o.callsUsed
}

func (o *Optimizer) Optimize(ctx context.Context, seed Candidate, examples []any) (*Result, error) {
	if o == nil {
		return nil, fmt.Errorf("optimizer is nil")
	}
	if o.eval == nil {
		return nil, fmt.Errorf("optimizer: evaluator is nil")
	}
	if o.reflector == nil || o.reflector.Engine == nil {
		return nil, fmt.Errorf("optimizer: reflector is nil")
	}
	if len(seed) == 0 {
		return nil, fmt.Errorf("optimizer: seed candidate is empty")
	}
	if len(examples) == 0 {
		return nil, fmt.Errorf("optimizer: dataset is empty")
	}

	keys, err := deriveOptimizableKeys(o.cfg, seed)
	if err != nil {
		return nil, err
	}
	o.paramKeys = keys

	// Initialize pool with seed.
	seedLastUpdated := map[string]CandidateID{}
	for k := range seed {
		seedLastUpdated[k] = 0
	}
	seedNode := &candidateNode{
		ID:             0,
		ParentID:       -1,
		Parent2ID:      -1,
		Candidate:      cloneCandidate(seed),
		CreatedAt:      o.cfg.Now(),
		Evaluations:    map[int]EvalResult{},
		Operation:      "seed",
		NextParamIndex: 0,
		LastUpdated:    seedLastUpdated,
	}
	o.pool = append(o.pool, seedNode)

	// Evaluate seed on an initial batch (to get some ASI).
	initIdx := o.sampleBatchIndices(len(examples), o.cfg.BatchSize, o.remainingBudget())
	if len(initIdx) == 0 {
		return nil, fmt.Errorf("optimizer: insufficient budget to evaluate seed")
	}
	if _, err := o.ensureEvaluated(ctx, seedNode, examples, initIdx); err != nil {
		return nil, err
	}

	bestNode := seedNode

	for o.callsUsed < o.cfg.MaxEvalCalls {
		callsAtIterStart := o.callsUsed
		remaining := o.remainingBudget()
		if remaining <= 0 {
			break
		}

		parent := o.selectParent()
		if parent == nil {
			break
		}

		useMerge := o.cfg.MergeProbability > 0 && len(o.pool) >= 2 && o.rng.Float64() < o.cfg.MergeProbability
		var parent2 *candidateNode
		if useMerge {
			parent2 = o.selectParentDistinct(parent.ID)
			if parent2 == nil {
				useMerge = false
			}
		}

		// Plan evaluations within remaining budget.
		// Worst case:
		//  - mutation: evaluate parent + child on the same minibatch  => 2 * batchSize
		//  - merge: evaluate parentA + parentB + child on minibatch   => 3 * batchSize
		mult := 2
		if useMerge {
			mult = 3
		}
		batchSize := o.cfg.BatchSize
		if batchSize*mult > remaining {
			batchSize = remaining / mult
		}
		if batchSize <= 0 {
			break
		}

		batchIdx := o.sampleBatchIndices(len(examples), batchSize, remaining)
		if len(batchIdx) == 0 {
			break
		}

		parentEvals, err := o.ensureEvaluated(ctx, parent, examples, batchIdx)
		if err != nil {
			return nil, err
		}

		var parent2Stats CandidateStats

		childID := CandidateID(len(o.pool))
		components := o.selectComponents(parent)
		if len(components) == 0 {
			break
		}

		var childCand Candidate
		var childLastUpdated map[string]CandidateID
		var rawReflection string
		var operation string
		var updatedKeys []string

		if useMerge {
			parent2Evals, err := o.ensureEvaluated(ctx, parent2, examples, batchIdx)
			if err != nil {
				return nil, err
			}
			parent2Stats = AggregateStats(parent2Evals)

			// Start with a simple system-aware merge that picks the newest version of each key
			// from the two parents. Then (optionally) run an LLM-based merge for the selected
			// component(s).
			childCand, childLastUpdated = o.systemAwareMerge(parent, parent2)
			if childCand == nil {
				childCand = cloneCandidate(parent.Candidate)
			}
			if childLastUpdated == nil {
				childLastUpdated = cloneLastUpdated(parent.LastUpdated)
			}

			rawByKey := map[string]string{}
			for _, key := range components {
				sideInfoA := FormatSideInfoForKey(examples, parentEvals, key, o.cfg.MaxSideInfoChars)
				sideInfoB := FormatSideInfoForKey(examples, parent2Evals, key, o.cfg.MaxSideInfoChars)

				mergedText, mergeRaw, err := o.proposeMerge(ctx, MergeInput{
					ParentA:   cloneCandidate(parent.Candidate),
					ParentB:   cloneCandidate(parent2.Candidate),
					ParamKey:  key,
					ParamA:    parent.Candidate[key],
					ParamB:    parent2.Candidate[key],
					SideInfoA: sideInfoA,
					SideInfoB: sideInfoB,
				})
				if err != nil {
					return nil, err
				}
				rawByKey[key] = mergeRaw
				childCand[key] = mergedText
				childLastUpdated[key] = childID
				updatedKeys = append(updatedKeys, key)
			}
			rawReflection = encodeRawByKey(rawByKey)
			operation = "merge"
		} else {
			childCand = cloneCandidate(parent.Candidate)
			childLastUpdated = cloneLastUpdated(parent.LastUpdated)
			rawByKey := map[string]string{}
			for _, key := range components {
				sideInfo := FormatSideInfoForKey(examples, parentEvals, key, o.cfg.MaxSideInfoChars)
				current := parent.Candidate[key]

				childText, raw, err := o.reflector.Propose(ctx, current, sideInfo)
				if err != nil {
					return nil, err
				}
				rawByKey[key] = raw
				childCand[key] = childText
				childLastUpdated[key] = childID
				updatedKeys = append(updatedKeys, key)
			}
			rawReflection = encodeRawByKey(rawByKey)
			operation = "mutate"
		}

		childNode := &candidateNode{
			ID:             childID,
			ParentID:       parent.ID,
			Parent2ID:      -1,
			Candidate:      childCand,
			CreatedAt:      o.cfg.Now(),
			Evaluations:    map[int]EvalResult{},
			ReflectionRaw:  rawReflection,
			Operation:      operation,
			LastUpdated:    childLastUpdated,
			UpdatedKeys:    append([]string(nil), updatedKeys...),
			NextParamIndex: parent.NextParamIndex,
		}
		if useMerge {
			childNode.Parent2ID = parent2.ID
			childNode.NextParamIndex = maxInt(parent.NextParamIndex, parent2.NextParamIndex)
		}

		childEvals, err := o.ensureEvaluated(ctx, childNode, examples, batchIdx)
		if err != nil {
			return nil, err
		}

		parentStats := AggregateStats(parentEvals)
		childStats := AggregateStats(childEvals)
		baselineStats := parentStats
		if useMerge {
			// For merge, accept relative to the better of the two parents on this batch.
			if parent2Stats.MeanScore > baselineStats.MeanScore {
				baselineStats = parent2Stats
			}
		}

		accepted := o.acceptChild(baselineStats, childStats)
		if accepted {
			o.pool = append(o.pool, childNode)
			// Update best based on global stats available so far.
			if childGlobal := o.globalStats(childNode); childGlobal.MeanScore > o.globalStats(bestNode).MeanScore {
				bestNode = childNode
			}
		}

		// Guard against stagnation: when all evals come from cache and no candidate is accepted,
		// the loop would otherwise spin forever.
		if o.callsUsed == callsAtIterStart && !accepted {
			break
		}
	}

	bestStats := o.globalStats(bestNode)

	entries := make([]CandidateEntry, 0, len(o.pool))
	for _, n := range o.pool {
		entries = append(entries, CandidateEntry{
			ID:            int(n.ID),
			ParentID:      int(n.ParentID),
			Parent2ID:     int(n.Parent2ID),
			Operation:     n.Operation,
			Hash:          candidateHash(n.Candidate),
			CreatedAt:     n.CreatedAt,
			Candidate:     cloneCandidate(n.Candidate),
			GlobalStats:   o.globalStats(n),
			EvalsCached:   len(o.cache[candidateHash(n.Candidate)]),
			ReflectionRaw: n.ReflectionRaw,
			UpdatedKeys:   append([]string(nil), n.UpdatedKeys...),
		})
	}

	return &Result{
		BestCandidate: cloneCandidate(bestNode.Candidate),
		BestStats:     bestStats,
		CallsUsed:     o.callsUsed,
		Candidates:    entries,
	}, nil
}

func (o *Optimizer) remainingBudget() int {
	return o.cfg.MaxEvalCalls - o.callsUsed
}

func (o *Optimizer) acceptChild(parent, child CandidateStats) bool {
	// Multi-objective: accept if child dominates parent.
	if len(parent.MeanObjectives) > 1 || len(child.MeanObjectives) > 1 {
		if Dominates(child.MeanObjectives, parent.MeanObjectives) {
			return true
		}
		// Fallback: accept if scalar score improved.
	}

	return child.MeanScore > parent.MeanScore+o.cfg.Epsilon
}

func (o *Optimizer) selectParent() *candidateNode {
	if len(o.pool) == 0 {
		return nil
	}

	// Build objective vectors for selection.
	obj := make([]ObjectiveScores, 0, len(o.pool))
	scalars := make([]float64, 0, len(o.pool))
	for _, n := range o.pool {
		stats := o.globalStats(n)
		vec := stats.MeanObjectives
		if len(vec) == 0 {
			vec = ObjectiveScores{"score": stats.MeanScore}
		}
		obj = append(obj, vec)
		scalars = append(scalars, stats.MeanScore)
	}

	// If we have multiple objectives overall, use Pareto front.
	keys := unionObjectiveKeys(obj)
	var candIdx []int
	if len(keys) > 1 {
		candIdx = ParetoFront(obj)
	} else {
		candIdx = TopKByScore(scalars, o.cfg.FrontierSize)
	}

	if len(candIdx) == 0 {
		// Fallback to uniform.
		return o.pool[o.rng.Intn(len(o.pool))]
	}

	// Weighted random selection by scalar score (shifted to positive).
	minS := math.Inf(1)
	for _, i := range candIdx {
		if scalars[i] < minS {
			minS = scalars[i]
		}
	}
	weights := make([]float64, 0, len(candIdx))
	sum := 0.0
	for _, i := range candIdx {
		w := scalars[i] - minS + 1e-9
		if w < 0 {
			w = 0
		}
		weights = append(weights, w)
		sum += w
	}
	var chosen int
	if sum <= 0 {
		chosen = candIdx[o.rng.Intn(len(candIdx))]
	} else {
		r := o.rng.Float64() * sum
		acc := 0.0
		for j, i := range candIdx {
			acc += weights[j]
			if r <= acc {
				chosen = i
				break
			}
		}
	}
	return o.pool[chosen]
}

func (o *Optimizer) selectParentDistinct(exclude CandidateID) *candidateNode {
	if o == nil || len(o.pool) < 2 {
		return nil
	}
	// Try a few times using the same selection policy.
	for i := 0; i < 10; i++ {
		n := o.selectParent()
		if n != nil && n.ID != exclude {
			return n
		}
	}
	// Fallback: pick uniformly among all nodes except exclude.
	idx := o.rng.Intn(len(o.pool) - 1)
	for _, n := range o.pool {
		if n.ID == exclude {
			continue
		}
		if idx == 0 {
			return n
		}
		idx--
	}
	return nil
}

func (o *Optimizer) proposeMerge(ctx context.Context, in MergeInput) (string, string, error) {
	if o == nil {
		return "", "", fmt.Errorf("proposeMerge: optimizer is nil")
	}
	if o.mergeFn != nil {
		return o.mergeFn(ctx, in)
	}
	if o.reflector == nil {
		return "", "", fmt.Errorf("proposeMerge: reflector is nil")
	}
	return o.reflector.Merge(ctx, in.ParamA, in.ParamB, in.SideInfoA, in.SideInfoB)
}

func unionObjectiveKeys(vecs []ObjectiveScores) map[string]struct{} {
	out := map[string]struct{}{}
	for _, v := range vecs {
		for k := range v {
			out[k] = struct{}{}
		}
	}
	return out
}

func (o *Optimizer) sampleBatchIndices(n, batchSize, budget int) []int {
	if n <= 0 || batchSize <= 0 || budget <= 0 {
		return nil
	}
	if batchSize > n {
		batchSize = n
	}
	if batchSize > budget {
		batchSize = budget
	}
	if batchSize <= 0 {
		return nil
	}
	if batchSize == n {
		out := make([]int, n)
		for i := 0; i < n; i++ {
			out[i] = i
		}
		return out
	}

	// Sample without replacement.
	perm := o.rng.Perm(n)
	return append([]int(nil), perm[:batchSize]...)
}

func (o *Optimizer) ensureEvaluated(ctx context.Context, n *candidateNode, examples []any, indices []int) ([]ExampleEval, error) {
	if n == nil {
		return nil, fmt.Errorf("ensureEvaluated: node is nil")
	}
	h := candidateHash(n.Candidate)
	if h == "" {
		return nil, fmt.Errorf("ensureEvaluated: candidate hash is empty")
	}
	if _, ok := o.cache[h]; !ok {
		o.cache[h] = map[int]EvalResult{}
	}

	out := make([]ExampleEval, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(examples) {
			continue
		}
		if cached, ok := o.cache[h][idx]; ok {
			n.Evaluations[idx] = cached
			out = append(out, ExampleEval{ExampleIndex: idx, Result: cached})
			continue
		}
		if o.callsUsed >= o.cfg.MaxEvalCalls {
			break
		}
		res, err := o.eval(ctx, n.Candidate, idx, examples[idx])
		if err != nil {
			return nil, fmt.Errorf("evaluator failed for example %d: %w", idx, err)
		}
		// Ensure objectives has a "score" dimension if none were provided.
		if len(res.Objectives) == 0 {
			res.Objectives = ObjectiveScores{"score": res.Score}
		}
		o.cache[h][idx] = res
		n.Evaluations[idx] = res
		o.callsUsed++
		out = append(out, ExampleEval{ExampleIndex: idx, Result: res})
	}
	return out, nil
}

func (o *Optimizer) globalStats(n *candidateNode) CandidateStats {
	if n == nil {
		return CandidateStats{}
	}
	evals := make([]ExampleEval, 0, len(n.Evaluations))
	for idx, res := range n.Evaluations {
		evals = append(evals, ExampleEval{ExampleIndex: idx, Result: res})
	}
	return AggregateStats(evals)
}

// AggregateStats computes mean score + mean objectives for a slice of evaluations.
func AggregateStats(evals []ExampleEval) CandidateStats {
	if len(evals) == 0 {
		return CandidateStats{}
	}
	sumScore := 0.0
	count := 0

	sumObj := map[string]float64{}
	cntObj := map[string]int{}

	for _, e := range evals {
		sumScore += e.Result.Score
		count++

		vec := e.Result.Objectives
		if len(vec) == 0 {
			vec = ObjectiveScores{"score": e.Result.Score}
		}
		for k, v := range vec {
			sumObj[k] += v
			cntObj[k]++
		}
	}

	meanObj := ObjectiveScores{}
	for k, s := range sumObj {
		if cntObj[k] > 0 {
			meanObj[k] = s / float64(cntObj[k])
		}
	}

	return CandidateStats{
		MeanScore:      sumScore / float64(count),
		MeanObjectives: meanObj,
		N:              count,
	}
}

func cloneCandidate(c Candidate) Candidate {
	if c == nil {
		return nil
	}
	out := make(Candidate, len(c))
	for k, v := range c {
		out[k] = v
	}
	return out
}

func cloneLastUpdated(m map[string]CandidateID) map[string]CandidateID {
	if m == nil {
		return map[string]CandidateID{}
	}
	out := make(map[string]CandidateID, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func encodeRawByKey(rawByKey map[string]string) string {
	if len(rawByKey) == 0 {
		return ""
	}
	if len(rawByKey) == 1 {
		for _, v := range rawByKey {
			return v
		}
	}
	blob, err := json.MarshalIndent(rawByKey, "", "  ")
	if err != nil {
		// best-effort fallback
		parts := make([]string, 0, len(rawByKey))
		for k, v := range rawByKey {
			parts = append(parts, fmt.Sprintf("[%s]\n%s", k, v))
		}
		sort.Strings(parts)
		return strings.Join(parts, "\n\n")
	}
	return string(blob)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func deriveOptimizableKeys(cfg Config, seed Candidate) ([]string, error) {
	// Explicit list provided.
	if len(cfg.OptimizableKeys) > 0 {
		seen := map[string]struct{}{}
		out := make([]string, 0, len(cfg.OptimizableKeys))
		for _, raw := range cfg.OptimizableKeys {
			k := strings.TrimSpace(raw)
			if k == "" {
				continue
			}
			if _, ok := seen[k]; ok {
				continue
			}
			if _, ok := seed[k]; !ok {
				return nil, fmt.Errorf("optimizer: optimizable key %q is not present in seed candidate", k)
			}
			seen[k] = struct{}{}
			out = append(out, k)
		}
		if len(out) == 0 {
			return nil, fmt.Errorf("optimizer: OptimizableKeys was provided but no valid keys were found")
		}
		return out, nil
	}

	// Default: deterministically use seed keys (prefer "prompt" first when present).
	keys := make([]string, 0, len(seed))
	for k := range seed {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return nil, fmt.Errorf("optimizer: seed candidate has no keys")
	}
	if _, ok := seed["prompt"]; ok {
		// Move "prompt" to the front for backwards-compatible behavior.
		out := make([]string, 0, len(keys))
		out = append(out, "prompt")
		for _, k := range keys {
			if k == "prompt" {
				continue
			}
			out = append(out, k)
		}
		return out, nil
	}
	return keys, nil
}

func (o *Optimizer) isOptimizableKey(k string) bool {
	if o == nil {
		return false
	}
	for _, kk := range o.paramKeys {
		if kk == k {
			return true
		}
	}
	return false
}

// selectComponents chooses which candidate key(s) to update this iteration.
//
// Under the "round_robin" strategy, this method advances the parent's NextParamIndex.
func (o *Optimizer) selectComponents(parent *candidateNode) []string {
	if o == nil || parent == nil {
		return nil
	}
	if len(o.paramKeys) == 0 {
		k := primaryParamKey(parent.Candidate)
		if strings.TrimSpace(k) == "" {
			return nil
		}
		return []string{k}
	}

	switch strings.ToLower(strings.TrimSpace(o.cfg.ComponentSelector)) {
	case "all":
		out := make([]string, 0, len(o.paramKeys))
		for _, k := range o.paramKeys {
			if _, ok := parent.Candidate[k]; ok {
				out = append(out, k)
			}
		}
		return out
	default: // round_robin
		idx := parent.NextParamIndex
		if idx < 0 {
			idx = 0
		}
		for tries := 0; tries < len(o.paramKeys); tries++ {
			k := o.paramKeys[(idx+tries)%len(o.paramKeys)]
			if _, ok := parent.Candidate[k]; ok {
				parent.NextParamIndex = (idx + tries + 1) % len(o.paramKeys)
				return []string{k}
			}
		}
		return nil
	}
}

// systemAwareMerge performs a simple component-wise merge between two parents.
// For optimizable keys, it chooses the parent's value whose key was updated more recently.
//
// This mirrors the idea in GEPA's "system-aware merge": if a component has evolved in one
// parent more recently than the other, keep that newer component.
func (o *Optimizer) systemAwareMerge(a, b *candidateNode) (Candidate, map[string]CandidateID) {
	if o == nil || a == nil || b == nil {
		return nil, nil
	}

	child := cloneCandidate(a.Candidate)
	last := cloneLastUpdated(a.LastUpdated)

	for k, vb := range b.Candidate {
		// Always copy missing keys (even if not optimizable) to preserve candidate completeness.
		_, okA := a.Candidate[k]
		if !okA {
			child[k] = vb
			if lu, ok := b.LastUpdated[k]; ok {
				last[k] = lu
			}
			continue
		}

		if !o.isOptimizableKey(k) {
			// Keep A's value for non-optimizable keys.
			continue
		}

		luA := a.LastUpdated[k]
		luB := b.LastUpdated[k]
		if luB > luA {
			child[k] = vb
			last[k] = luB
		} else {
			last[k] = luA
		}
	}

	return child, last
}

func primaryParamKey(c Candidate) string {
	if c == nil {
		return "prompt"
	}
	if _, ok := c["prompt"]; ok {
		return "prompt"
	}
	// deterministic fallback: smallest key.
	var best string
	for k := range c {
		if best == "" || k < best {
			best = k
		}
	}
	if best == "" {
		return "prompt"
	}
	return best
}
