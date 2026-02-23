package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	gepaopt "github.com/go-go-golems/go-go-gepa/pkg/optimizer/gepa"
	"github.com/pkg/errors"
)

const optimizerPluginAPIVersion = "gepa.optimizer/v1"

type optimizerPluginMeta struct {
	APIVersion string
	Kind       string
	ID         string
	Name       string
}

type pluginEvaluateOptions struct {
	Profile       string
	EngineOptions map[string]any
	Tags          map[string]any
}

type optimizerPlugin struct {
	rt       *jsRuntime
	meta     optimizerPluginMeta
	instance *goja.Object

	evaluateFn goja.Callable
	datasetFn  goja.Callable
	mergeFn    goja.Callable
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
		return nil, optimizerPluginMeta{}, errors.Wrap(err, "plugin loader: require script module")
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

	instanceVal, err := createFn(descriptorObj, rt.vm.ToValue(hostContext))
	if err != nil {
		return nil, optimizerPluginMeta{}, errors.Wrap(err, "plugin loader: descriptor.create failed")
	}
	instanceObj := instanceVal.ToObject(rt.vm)
	if instanceObj == nil {
		return nil, optimizerPluginMeta{}, fmt.Errorf("plugin loader: descriptor.create must return an object instance")
	}

	evaluateVal := instanceObj.Get("evaluate")
	evaluateFn, ok := goja.AssertFunction(evaluateVal)
	if !ok {
		return nil, optimizerPluginMeta{}, fmt.Errorf("plugin loader: plugin instance.evaluate must be a function")
	}

	var mergeFn goja.Callable
	for _, key := range []string{"merge", "mergeCandidate", "mergePrompt"} {
		if mv := instanceObj.Get(key); mv != nil && !goja.IsUndefined(mv) && !goja.IsNull(mv) {
			if fn, ok := goja.AssertFunction(mv); ok {
				mergeFn = fn
				break
			}
		}
	}

	var datasetFn goja.Callable
	if dv := instanceObj.Get("dataset"); dv != nil && !goja.IsUndefined(dv) && !goja.IsNull(dv) {
		if fn, ok := goja.AssertFunction(dv); ok {
			datasetFn = fn
		}
	}
	if datasetFn == nil {
		if dv := instanceObj.Get("getDataset"); dv != nil && !goja.IsUndefined(dv) && !goja.IsNull(dv) {
			if fn, ok := goja.AssertFunction(dv); ok {
				datasetFn = fn
			}
		}
	}

	p := &optimizerPlugin{
		rt:         rt,
		meta:       meta,
		instance:   instanceObj,
		evaluateFn: evaluateFn,
		datasetFn:  datasetFn,
		mergeFn:    mergeFn,
	}

	return p, meta, nil
}

func decodeOptimizerPluginMeta(descriptorObj *goja.Object) (optimizerPluginMeta, error) {
	apiVersion := strings.TrimSpace(descriptorObj.Get("apiVersion").String())
	kind := strings.TrimSpace(descriptorObj.Get("kind").String())
	id := strings.TrimSpace(descriptorObj.Get("id").String())
	name := strings.TrimSpace(descriptorObj.Get("name").String())

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
		APIVersion: apiVersion,
		Kind:       kind,
		ID:         id,
		Name:       name,
	}, nil
}

func (p *optimizerPlugin) Dataset() ([]any, error) {
	if p == nil || p.rt == nil || p.instance == nil {
		return nil, fmt.Errorf("plugin dataset: plugin not initialized")
	}
	if p.datasetFn == nil {
		return nil, fmt.Errorf("plugin dataset: instance.dataset() not found (provide --dataset)")
	}

	ret, err := p.datasetFn(p.instance, goja.Undefined())
	if err != nil {
		return nil, errors.Wrap(err, "plugin dataset: call failed")
	}

	decoded, err := decodeJSReturnValue(ret)
	if err != nil {
		return nil, errors.Wrap(err, "plugin dataset: invalid return value")
	}
	arr, ok := decoded.([]any)
	if ok {
		return arr, nil
	}
	return nil, fmt.Errorf("plugin dataset: expected array, got %T", decoded)
}

func (p *optimizerPlugin) Evaluate(
	candidate gepaopt.Candidate,
	exampleIndex int,
	example any,
	opts pluginEvaluateOptions,
) (gepaopt.EvalResult, error) {
	if p == nil || p.rt == nil || p.instance == nil || p.evaluateFn == nil {
		return gepaopt.EvalResult{}, fmt.Errorf("plugin evaluate: plugin not initialized")
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

	ret, err := p.evaluateFn(p.instance, p.rt.vm.ToValue(input), p.rt.vm.ToValue(options))
	if err != nil {
		return gepaopt.EvalResult{}, errors.Wrap(err, "plugin evaluate: call failed")
	}

	decoded, err := decodeJSReturnValue(ret)
	if err != nil {
		return gepaopt.EvalResult{}, errors.Wrap(err, "plugin evaluate: invalid return value")
	}

	er, err := decodeEvalResult(decoded)
	if err != nil {
		return gepaopt.EvalResult{}, err
	}
	er.Raw = decoded
	return er, nil
}

func (p *optimizerPlugin) HasMerge() bool {
	return p != nil && p.mergeFn != nil
}

func (p *optimizerPlugin) Merge(in gepaopt.MergeInput, opts pluginEvaluateOptions) (string, string, error) {
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

	ret, err := p.mergeFn(p.instance, p.rt.vm.ToValue(input), p.rt.vm.ToValue(options))
	if err != nil {
		return "", "", errors.Wrap(err, "plugin merge: call failed")
	}

	decoded, err := decodeJSReturnValue(ret)
	if err != nil {
		return "", "", errors.Wrap(err, "plugin merge: invalid return value")
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

// decodeJSReturnValue mirrors the cozo runner behavior:
// - if JS returns a string, attempt JSON parsing
// - if JS returns bytes, attempt JSON parsing
// - otherwise, return exported value
func decodeJSReturnValue(ret goja.Value) (any, error) {
	if ret == nil || goja.IsUndefined(ret) || goja.IsNull(ret) {
		return nil, fmt.Errorf("returned null/undefined")
	}
	if raw, ok := ret.Export().(string); ok {
		if strings.TrimSpace(raw) == "" {
			return nil, fmt.Errorf("returned empty string")
		}
		var jsonValue any
		if err := json.Unmarshal([]byte(raw), &jsonValue); err == nil {
			return jsonValue, nil
		}
		return raw, nil
	}
	if bytes, ok := ret.Export().([]uint8); ok {
		var jsonValue any
		if err := json.Unmarshal(bytes, &jsonValue); err == nil {
			return jsonValue, nil
		}
		return bytes, nil
	}
	return ret.Export(), nil
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
		var num json.Number = json.Number(strings.TrimSpace(x))
		return num.Float64()
	default:
		return 0, fmt.Errorf("unsupported numeric type %T", v)
	}
}

// Ensure the require package is linked (some Go compilers prune unused imports).
var _ = require.Require
