package main

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-gepa/pkg/jsbridge"
	gepaopt "github.com/go-go-golems/go-go-gepa/pkg/optimizer/gepa"
	"github.com/pkg/errors"
)

const optimizerPluginAPIVersion = "gepa.optimizer/v1"
const defaultPluginRegistryIdentifier = "local"

type optimizerPluginMeta struct {
	APIVersion         string
	Kind               string
	ID                 string
	Name               string
	RegistryIdentifier string
}

type pluginEvaluateOptions struct {
	Profile       string
	EngineOptions map[string]any
	Tags          map[string]any
	EventSink     jsbridge.EventSink
}

type optimizerPlugin struct {
	rt       *jsRuntime
	meta     optimizerPluginMeta
	instance *goja.Object

	evaluateFn          goja.Callable
	runFn               goja.Callable
	datasetFn           goja.Callable
	mergeFn             goja.Callable
	initialCandidateFn  goja.Callable
	selectComponentsFn  goja.Callable
	componentSideInfoFn goja.Callable
}

func loadOptimizerPlugin(rt *jsRuntime, absScriptPath string, hostContext map[string]any) (*optimizerPlugin, optimizerPluginMeta, error) {
	if rt == nil || rt.vm == nil || rt.reqMod == nil {
		return nil, optimizerPluginMeta{}, errors.New("plugin loader: runtime is nil")
	}
	if strings.TrimSpace(absScriptPath) == "" {
		return nil, optimizerPluginMeta{}, errors.New("plugin loader: script path is empty")
	}

	exported, err := rt.reqMod.Require(absScriptPath)
	if err != nil {
		base := filepath.Base(absScriptPath)
		baseNoExt := strings.TrimSuffix(base, filepath.Ext(base))
		fallbacks := []string{
			base,
			baseNoExt,
			"./" + base,
			"./" + baseNoExt,
		}
		for _, mod := range fallbacks {
			exported, err = rt.reqMod.Require(mod)
			if err == nil {
				break
			}
		}
		if err != nil {
			return nil, optimizerPluginMeta{}, errors.Wrap(err, "plugin loader: require script module")
		}
	}

	descriptorObj := exported.ToObject(rt.vm)
	if descriptorObj == nil {
		return nil, optimizerPluginMeta{}, fmt.Errorf("plugin loader: script module did not export an object descriptor")
	}

	meta, err := decodeOptimizerPluginMeta(descriptorObj)
	if err != nil {
		return nil, optimizerPluginMeta{}, err
	}

	createVal := descriptorObj.Get("create")
	createFn, ok := goja.AssertFunction(createVal)
	if !ok {
		return nil, optimizerPluginMeta{}, fmt.Errorf("plugin loader: descriptor.create must be a function")
	}

	if hostContext == nil {
		hostContext = map[string]any{}
	}
	if strings.TrimSpace(meta.RegistryIdentifier) == "" {
		meta.RegistryIdentifier = defaultPluginRegistryIdentifier
	}
	hostContext["pluginRegistryIdentifier"] = meta.RegistryIdentifier

	instanceVal, err := createFn(descriptorObj, rt.vm.ToValue(hostContext))
	if err != nil {
		return nil, optimizerPluginMeta{}, errors.Wrap(err, "plugin loader: descriptor.create failed")
	}
	instanceObj := instanceVal.ToObject(rt.vm)
	if instanceObj == nil {
		return nil, optimizerPluginMeta{}, fmt.Errorf("plugin loader: descriptor.create must return an object instance")
	}

	evaluateFn := findOptionalCallable(instanceObj, "evaluate")
	runFn := findOptionalCallable(instanceObj, "run", "runCandidate")
	if evaluateFn == nil && runFn == nil {
		return nil, optimizerPluginMeta{}, fmt.Errorf("plugin loader: plugin instance must expose evaluate() and/or run()")
	}

	mergeFn := findOptionalCallable(instanceObj, "merge", "mergeCandidate", "mergePrompt")
	datasetFn := findOptionalCallable(instanceObj, "dataset", "getDataset")
	initialCandidateFn := findOptionalCallable(instanceObj, "initialCandidate", "getInitialCandidate")
	selectComponentsFn := findOptionalCallable(instanceObj, "selectComponents", "chooseComponents")
	componentSideInfoFn := findOptionalCallable(instanceObj, "componentSideInfo", "sideInfoForComponent", "buildSideInfo")

	p := &optimizerPlugin{
		rt:                  rt,
		meta:                meta,
		instance:            instanceObj,
		evaluateFn:          evaluateFn,
		runFn:               runFn,
		datasetFn:           datasetFn,
		mergeFn:             mergeFn,
		initialCandidateFn:  initialCandidateFn,
		selectComponentsFn:  selectComponentsFn,
		componentSideInfoFn: componentSideInfoFn,
	}

	return p, meta, nil
}

func findOptionalCallable(obj *goja.Object, keys ...string) goja.Callable {
	if obj == nil {
		return nil
	}
	for _, key := range keys {
		v := obj.Get(key)
		if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
			continue
		}
		if fn, ok := goja.AssertFunction(v); ok {
			return fn
		}
	}
	return nil
}

func decodeOptimizerPluginMeta(descriptorObj *goja.Object) (optimizerPluginMeta, error) {
	apiVersion := decodeOptionalJSString(descriptorObj.Get("apiVersion"))
	kind := decodeOptionalJSString(descriptorObj.Get("kind"))
	id := decodeOptionalJSString(descriptorObj.Get("id"))
	name := decodeOptionalJSString(descriptorObj.Get("name"))
	registryIdentifier := decodeOptionalJSString(descriptorObj.Get("registryIdentifier"))

	if apiVersion == "" {
		return optimizerPluginMeta{}, fmt.Errorf("plugin loader: descriptor.apiVersion is required")
	}
	if apiVersion != optimizerPluginAPIVersion {
		return optimizerPluginMeta{}, fmt.Errorf("plugin loader: unsupported apiVersion %q (expected %q)", apiVersion, optimizerPluginAPIVersion)
	}
	if kind != "optimizer" {
		return optimizerPluginMeta{}, fmt.Errorf("plugin loader: descriptor.kind must be %q", "optimizer")
	}
	if id == "" {
		return optimizerPluginMeta{}, fmt.Errorf("plugin loader: descriptor.id is required")
	}
	if name == "" {
		return optimizerPluginMeta{}, fmt.Errorf("plugin loader: descriptor.name is required")
	}

	if cv := descriptorObj.Get("create"); cv == nil || goja.IsUndefined(cv) || goja.IsNull(cv) {
		return optimizerPluginMeta{}, fmt.Errorf("plugin loader: descriptor.create is required")
	}
	if _, ok := goja.AssertFunction(descriptorObj.Get("create")); !ok {
		return optimizerPluginMeta{}, fmt.Errorf("plugin loader: descriptor.create must be a function")
	}

	return optimizerPluginMeta{
		APIVersion:         apiVersion,
		Kind:               kind,
		ID:                 id,
		Name:               name,
		RegistryIdentifier: firstNonEmpty(registryIdentifier, defaultPluginRegistryIdentifier),
	}, nil
}

func decodeOptionalJSString(v goja.Value) string {
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return ""
	}
	return strings.TrimSpace(v.String())
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func (p *optimizerPlugin) callPluginFunction(ctx context.Context, method string, fn goja.Callable, args ...any) (any, error) {
	if p == nil || p.rt == nil || p.rt.vm == nil || p.instance == nil {
		return nil, fmt.Errorf("plugin %s: plugin not initialized", method)
	}
	if fn == nil {
		return nil, fmt.Errorf("plugin %s: method is not available", method)
	}

	op := fmt.Sprintf("optimizer.%s.%s", strings.TrimSpace(p.meta.ID), strings.TrimSpace(method))
	resolved, err := jsbridge.CallAndResolve(ctx, jsbridge.CallAndResolveOptions{
		Op:             op,
		VM:             p.rt.vm,
		Runner:         p.rt.runner,
		DefaultTimeout: jsbridge.DefaultPromiseTimeout,
	}, func(vm *goja.Runtime) (goja.Value, error) {
		jsArgs := make([]goja.Value, 0, len(args))
		for _, arg := range args {
			jsArgs = append(jsArgs, vm.ToValue(arg))
		}
		return fn(p.instance, jsArgs...)
	})
	if err != nil {
		return nil, err
	}

	decoded, err := decodeJSReturnValue(resolved)
	if err != nil {
		return nil, err
	}
	return decoded, nil
}

func (p *optimizerPlugin) makeEventHooks(method string, sink jsbridge.EventSink) (any, map[string]any) {
	if sink == nil {
		return nil, nil
	}
	emitter := jsbridge.NewEmitter(jsbridge.EmitterOptions{
		PluginMethod:             method,
		PluginID:                 p.meta.ID,
		PluginName:               p.meta.Name,
		PluginRegistryIdentifier: p.meta.RegistryIdentifier,
		Sink:                     sink,
	})
	emit := func(payload any) {
		emitter.Emit(payload)
	}
	return emit, map[string]any{"emit": emit}
}

func (p *optimizerPlugin) Dataset(ctx context.Context) ([]any, error) {
	if p == nil || p.rt == nil || p.instance == nil {
		return nil, fmt.Errorf("plugin dataset: plugin not initialized")
	}
	if p.datasetFn == nil {
		return nil, fmt.Errorf("plugin dataset: instance.dataset() not found (provide --dataset)")
	}

	decoded, err := p.callPluginFunction(ctx, "dataset", p.datasetFn, goja.Undefined())
	if err != nil {
		return nil, errors.Wrap(err, "plugin dataset: call failed")
	}
	arr, ok := decoded.([]any)
	if ok {
		return arr, nil
	}
	return nil, fmt.Errorf("plugin dataset: expected array, got %T", decoded)
}

func (p *optimizerPlugin) Evaluate(ctx context.Context, candidate gepaopt.Candidate, exampleIndex int, example any, opts pluginEvaluateOptions) (gepaopt.EvalResult, error) {
	if p == nil || p.rt == nil || p.instance == nil {
		return gepaopt.EvalResult{}, fmt.Errorf("plugin evaluate: plugin not initialized")
	}
	if p.evaluateFn == nil {
		return gepaopt.EvalResult{}, fmt.Errorf("plugin evaluate: evaluate() not available")
	}

	input := map[string]any{
		"candidate":    candidate,
		"example":      example,
		"exampleIndex": exampleIndex,
	}
	options := map[string]any{
		"profile":       strings.TrimSpace(opts.Profile),
		"engineOptions": opts.EngineOptions,
		"tags":          opts.Tags,
	}
	if emit, events := p.makeEventHooks("evaluate", opts.EventSink); emit != nil {
		options["emitEvent"] = emit
		options["events"] = events
	}

	decoded, err := p.callPluginFunction(ctx, "evaluate", p.evaluateFn, input, options)
	if err != nil {
		return gepaopt.EvalResult{}, errors.Wrap(err, "plugin evaluate: call failed")
	}

	er, err := decodeEvalResult(decoded)
	if err != nil {
		return gepaopt.EvalResult{}, err
	}
	er.Raw = decoded
	return er, nil
}

func (p *optimizerPlugin) HasEvaluate() bool {
	return p != nil && p.evaluateFn != nil
}

func (p *optimizerPlugin) HasRun() bool {
	return p != nil && p.runFn != nil
}

func (p *optimizerPlugin) Run(ctx context.Context, input map[string]any, candidate gepaopt.Candidate, opts pluginEvaluateOptions) (any, error) {
	if p == nil || p.rt == nil || p.instance == nil {
		return nil, fmt.Errorf("plugin run: plugin not initialized")
	}
	if p.runFn == nil {
		return nil, fmt.Errorf("plugin run: run() not available")
	}

	options := map[string]any{
		"candidate":     candidate,
		"profile":       strings.TrimSpace(opts.Profile),
		"engineOptions": opts.EngineOptions,
		"tags":          opts.Tags,
	}
	if emit, events := p.makeEventHooks("run", opts.EventSink); emit != nil {
		options["emitEvent"] = emit
		options["events"] = events
	}

	decoded, err := p.callPluginFunction(ctx, "run", p.runFn, input, options)
	if err != nil {
		return nil, errors.Wrap(err, "plugin run: call failed")
	}
	return decoded, nil
}

func (p *optimizerPlugin) HasMerge() bool {
	return p != nil && p.mergeFn != nil
}

func (p *optimizerPlugin) Merge(ctx context.Context, in gepaopt.MergeInput, opts pluginEvaluateOptions) (string, string, error) {
	if p == nil || p.rt == nil || p.instance == nil || p.mergeFn == nil {
		return "", "", fmt.Errorf("plugin merge: merge() not available")
	}

	input := map[string]any{
		"candidateA": in.ParentA,
		"candidateB": in.ParentB,
		"paramKey":   in.ParamKey,
		"paramA":     in.ParamA,
		"paramB":     in.ParamB,
		"sideInfoA":  in.SideInfoA,
		"sideInfoB":  in.SideInfoB,
	}
	options := map[string]any{
		"profile":       strings.TrimSpace(opts.Profile),
		"engineOptions": opts.EngineOptions,
		"tags":          opts.Tags,
	}
	if emit, events := p.makeEventHooks("merge", opts.EventSink); emit != nil {
		options["emitEvent"] = emit
		options["events"] = events
	}

	decoded, err := p.callPluginFunction(ctx, "merge", p.mergeFn, input, options)
	if err != nil {
		return "", "", errors.Wrap(err, "plugin merge: call failed")
	}

	merged, err := decodeMergeOutput(decoded, in.ParamKey)
	if err != nil {
		return "", "", err
	}

	raw := ""
	switch x := decoded.(type) {
	case string:
		raw = x
	default:
		blob, _ := json.MarshalIndent(decoded, "", "  ")
		raw = string(blob)
	}

	return merged, raw, nil
}

func (p *optimizerPlugin) HasInitialCandidate() bool {
	return p != nil && p.initialCandidateFn != nil
}

func (p *optimizerPlugin) InitialCandidate(ctx context.Context, opts pluginEvaluateOptions) (gepaopt.Candidate, error) {
	if p == nil || p.rt == nil || p.instance == nil || p.initialCandidateFn == nil {
		return nil, fmt.Errorf("plugin initialCandidate: initialCandidate() not available")
	}
	options := map[string]any{
		"profile":       strings.TrimSpace(opts.Profile),
		"engineOptions": opts.EngineOptions,
		"tags":          opts.Tags,
	}
	if emit, events := p.makeEventHooks("initialCandidate", opts.EventSink); emit != nil {
		options["emitEvent"] = emit
		options["events"] = events
	}
	decoded, err := p.callPluginFunction(ctx, "initialCandidate", p.initialCandidateFn, options)
	if err != nil {
		return nil, errors.Wrap(err, "plugin initialCandidate: call failed")
	}
	return decodeCandidate(decoded)
}

func (p *optimizerPlugin) HasSelectComponents() bool {
	return p != nil && p.selectComponentsFn != nil
}

func (p *optimizerPlugin) SelectComponents(ctx context.Context, in gepaopt.ComponentSelectionInput, opts pluginEvaluateOptions) ([]string, error) {
	if p == nil || p.rt == nil || p.instance == nil || p.selectComponentsFn == nil {
		return nil, fmt.Errorf("plugin selectComponents: selectComponents() not available")
	}

	input := map[string]any{
		"operation":      in.Operation,
		"parentId":       int(in.ParentID),
		"parent2Id":      int(in.Parent2ID),
		"candidate":      in.Candidate,
		"availableKeys":  in.AvailableKeys,
		"nextParamIndex": in.NextParamIndex,
	}
	options := map[string]any{
		"profile":       strings.TrimSpace(opts.Profile),
		"engineOptions": opts.EngineOptions,
		"tags":          opts.Tags,
	}
	if emit, events := p.makeEventHooks("selectComponents", opts.EventSink); emit != nil {
		options["emitEvent"] = emit
		options["events"] = events
	}
	decoded, err := p.callPluginFunction(ctx, "selectComponents", p.selectComponentsFn, input, options)
	if err != nil {
		return nil, errors.Wrap(err, "plugin selectComponents: call failed")
	}
	components, err := decodeStringList(decoded)
	if err != nil {
		return nil, err
	}
	return components, nil
}

func (p *optimizerPlugin) HasComponentSideInfo() bool {
	return p != nil && p.componentSideInfoFn != nil
}

func (p *optimizerPlugin) ComponentSideInfo(ctx context.Context, in gepaopt.SideInfoInput, opts pluginEvaluateOptions) (string, error) {
	if p == nil || p.rt == nil || p.instance == nil || p.componentSideInfoFn == nil {
		return "", fmt.Errorf("plugin componentSideInfo: componentSideInfo() not available")
	}

	input := map[string]any{
		"operation": in.Operation,
		"paramKey":  in.ParamKey,
		"examples":  in.Examples,
		"evals":     in.Evals,
		"maxChars":  in.MaxChars,
		"default":   in.Default,
	}
	options := map[string]any{
		"profile":       strings.TrimSpace(opts.Profile),
		"engineOptions": opts.EngineOptions,
		"tags":          opts.Tags,
	}
	if emit, events := p.makeEventHooks("componentSideInfo", opts.EventSink); emit != nil {
		options["emitEvent"] = emit
		options["events"] = events
	}
	decoded, err := p.callPluginFunction(ctx, "componentSideInfo", p.componentSideInfoFn, input, options)
	if err != nil {
		return "", errors.Wrap(err, "plugin componentSideInfo: call failed")
	}
	return decodeSideInfoOutput(decoded)
}

// decodeJSReturnValue mirrors the cozo runner behavior:
// - if JS returns a string, attempt JSON parsing
// - if JS returns bytes, attempt JSON parsing
// - otherwise, return exported value
func decodeJSReturnValue(ret any) (any, error) {
	if ret == nil {
		return nil, fmt.Errorf("returned null/undefined")
	}
	if raw, ok := ret.(string); ok {
		if strings.TrimSpace(raw) == "" {
			return nil, fmt.Errorf("returned empty string")
		}
		var jsonValue any
		if err := json.Unmarshal([]byte(raw), &jsonValue); err == nil {
			return jsonValue, nil
		}
		return raw, nil
	}
	if bytes, ok := ret.([]uint8); ok {
		var jsonValue any
		if err := json.Unmarshal(bytes, &jsonValue); err == nil {
			return jsonValue, nil
		}
		return bytes, nil
	}
	return ret, nil
}

func decodeEvalResult(v any) (gepaopt.EvalResult, error) {
	switch x := v.(type) {
	case map[string]any:
		return decodeEvalResultFromMap(x)
	case float64:
		return gepaopt.EvalResult{Score: x}, nil
	case int:
		return gepaopt.EvalResult{Score: float64(x)}, nil
	default:
		return gepaopt.EvalResult{}, fmt.Errorf("evaluator must return an object with {score}, got %T", v)
	}
}

func decodeMergeOutput(v any, paramKey string) (string, error) {
	paramKey = strings.TrimSpace(paramKey)
	if paramKey == "" {
		paramKey = "prompt"
	}

	readString := func(m map[string]any, key string) string {
		vv, ok := m[key]
		if !ok || vv == nil {
			return ""
		}
		s, ok := vv.(string)
		if !ok {
			return ""
		}
		return strings.TrimSpace(s)
	}

	switch x := v.(type) {
	case string:
		out := strings.TrimSpace(x)
		if out == "" {
			return "", fmt.Errorf("merge returned empty string")
		}
		return out, nil
	case map[string]any:
		if candRaw, ok := x["candidate"]; ok && candRaw != nil {
			if cm, ok := candRaw.(map[string]any); ok {
				if s := readString(cm, paramKey); s != "" {
					return s, nil
				}
			}
		}

		for _, key := range []string{paramKey, "prompt", "merged", "mergedPrompt", "text"} {
			if s := readString(x, key); s != "" {
				return s, nil
			}
		}
		return "", fmt.Errorf("merge must return a string or an object containing %q (or {candidate:{%q:...}}); got keys=%v", paramKey, paramKey, keysOf(x))
	default:
		return "", fmt.Errorf("merge must return a string or object, got %T", v)
	}
}

func decodeCandidate(v any) (gepaopt.Candidate, error) {
	switch x := v.(type) {
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return nil, fmt.Errorf("initial candidate string is empty")
		}
		return gepaopt.Candidate{"prompt": s}, nil
	case map[string]any:
		out := gepaopt.Candidate{}
		for k, vv := range x {
			key := strings.TrimSpace(k)
			if key == "" {
				continue
			}
			out[key] = toStringLossy(vv)
		}
		if len(out) == 0 {
			return nil, fmt.Errorf("initial candidate map is empty")
		}
		return out, nil
	default:
		return nil, fmt.Errorf("initial candidate must be string or object, got %T", v)
	}
}

func decodeStringList(v any) ([]string, error) {
	switch x := v.(type) {
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return nil, nil
		}
		return []string{s}, nil
	case []string:
		out := make([]string, 0, len(x))
		for _, item := range x {
			if s := strings.TrimSpace(item); s != "" {
				out = append(out, s)
			}
		}
		return out, nil
	case []any:
		out := make([]string, 0, len(x))
		for _, item := range x {
			s := strings.TrimSpace(toStringLossy(item))
			if s != "" {
				out = append(out, s)
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("component list must be string or array, got %T", v)
	}
}

func decodeSideInfoOutput(v any) (string, error) {
	switch x := v.(type) {
	case string:
		return x, nil
	case map[string]any:
		for _, key := range []string{"sideInfo", "text", "value", "default"} {
			if vv, ok := x[key]; ok {
				return toStringLossy(vv), nil
			}
		}
		return "", nil
	default:
		return toStringLossy(v), nil
	}
}

func toStringLossy(v any) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	case json.Number:
		return x.String()
	case bool:
		if x {
			return "true"
		}
		return "false"
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(x), 'f', -1, 32)
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case int32:
		return strconv.FormatInt(int64(x), 10)
	case int16:
		return strconv.FormatInt(int64(x), 10)
	case int8:
		return strconv.FormatInt(int64(x), 10)
	case uint:
		return strconv.FormatUint(uint64(x), 10)
	case uint64:
		return strconv.FormatUint(x, 10)
	case uint32:
		return strconv.FormatUint(uint64(x), 10)
	case uint16:
		return strconv.FormatUint(uint64(x), 10)
	case uint8:
		return strconv.FormatUint(uint64(x), 10)
	default:
		if blob, err := json.Marshal(x); err == nil {
			return string(blob)
		}
		return fmt.Sprintf("%v", x)
	}
}

func keysOf(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func decodeEvalResultFromMap(m map[string]any) (gepaopt.EvalResult, error) {
	scoreRaw, ok := m["score"]
	if !ok {
		scoreRaw, ok = m["value"]
	}
	if !ok {
		return gepaopt.EvalResult{}, fmt.Errorf("evaluator return value missing required field: score")
	}

	score, err := toFloat(scoreRaw)
	if err != nil {
		return gepaopt.EvalResult{}, fmt.Errorf("invalid score: %w", err)
	}

	var objScores gepaopt.ObjectiveScores
	for _, key := range []string{"objectiveScores", "objectives"} {
		if v, ok := m[key]; ok && v != nil {
			objScores, _ = decodeObjectiveScores(v)
			break
		}
	}

	out := gepaopt.EvalResult{
		Score:      score,
		Objectives: objScores,
		Output:     m["output"],
		Feedback:   m["feedback"],
		Trace:      m["trace"],
	}

	if notes, ok := m["notes"].(string); ok {
		out.EvaluatorNotes = notes
	} else if notes, ok := m["evaluatorNotes"].(string); ok {
		out.EvaluatorNotes = notes
	}

	return out, nil
}

func decodeObjectiveScores(v any) (gepaopt.ObjectiveScores, error) {
	out := gepaopt.ObjectiveScores{}
	switch x := v.(type) {
	case map[string]any:
		for k, vv := range x {
			f, err := toFloat(vv)
			if err != nil {
				continue
			}
			out[k] = f
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("empty objective scores")
	}
	return out, nil
}

func toFloat(v any) (float64, error) {
	switch x := v.(type) {
	case float64:
		return x, nil
	case float32:
		return float64(x), nil
	case int:
		return float64(x), nil
	case int64:
		return float64(x), nil
	case json.Number:
		return x.Float64()
	case string:
		if strings.TrimSpace(x) == "" {
			return 0, fmt.Errorf("empty string")
		}
		num := json.Number(strings.TrimSpace(x))
		return num.Float64()
	default:
		return 0, fmt.Errorf("unsupported numeric type %T", v)
	}
}
