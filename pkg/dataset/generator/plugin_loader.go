package generator

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/pkg/errors"
)

const PluginAPIVersion = "gepa.dataset-generator/v1"
const DefaultRegistryIdentifier = "local"

type PluginMeta struct {
	APIVersion         string
	Kind               string
	ID                 string
	Name               string
	RegistryIdentifier string
}

type PluginGenerateOptions struct {
	Profile       string
	EngineOptions map[string]any
	Tags          map[string]any
	Seed          int64
	RNG           any
	Config        any
}

type RNG interface {
	IntN(max int) int
	Float64() float64
	Choice(values []any) any
	Shuffle(values []any)
}

type Plugin struct {
	vm         *goja.Runtime
	meta       PluginMeta
	instance   *goja.Object
	generateFn goja.Callable
}

func LoadPlugin(vm *goja.Runtime, req *require.RequireModule, absScriptPath string, hostContext map[string]any) (*Plugin, PluginMeta, error) {
	if vm == nil || req == nil {
		return nil, PluginMeta{}, errors.New("dataset generator loader: runtime is nil")
	}
	if strings.TrimSpace(absScriptPath) == "" {
		return nil, PluginMeta{}, errors.New("dataset generator loader: script path is empty")
	}

	exported, err := req.Require(absScriptPath)
	if err != nil {
		base := filepath.Base(absScriptPath)
		baseNoExt := strings.TrimSuffix(base, filepath.Ext(base))
		fallbacks := []string{base, baseNoExt, "./" + base, "./" + baseNoExt}
		for _, mod := range fallbacks {
			exported, err = req.Require(mod)
			if err == nil {
				break
			}
		}
		if err != nil {
			return nil, PluginMeta{}, errors.Wrap(err, "dataset generator loader: require script module")
		}
	}

	descriptorObj := exported.ToObject(vm)
	if descriptorObj == nil {
		return nil, PluginMeta{}, fmt.Errorf("dataset generator loader: script module did not export an object descriptor")
	}

	meta, err := decodePluginMeta(descriptorObj)
	if err != nil {
		return nil, PluginMeta{}, err
	}

	createVal := descriptorObj.Get("create")
	createFn, ok := goja.AssertFunction(createVal)
	if !ok {
		return nil, PluginMeta{}, fmt.Errorf("dataset generator loader: descriptor.create must be a function")
	}

	if hostContext == nil {
		hostContext = map[string]any{}
	}
	if strings.TrimSpace(meta.RegistryIdentifier) == "" {
		meta.RegistryIdentifier = DefaultRegistryIdentifier
	}
	hostContext["pluginRegistryIdentifier"] = meta.RegistryIdentifier

	instanceVal, err := createFn(descriptorObj, vm.ToValue(hostContext))
	if err != nil {
		return nil, PluginMeta{}, errors.Wrap(err, "dataset generator loader: descriptor.create failed")
	}
	instanceObj := instanceVal.ToObject(vm)
	if instanceObj == nil {
		return nil, PluginMeta{}, fmt.Errorf("dataset generator loader: descriptor.create must return an object instance")
	}

	generateFn := findOptionalCallable(instanceObj, "generateOne", "generate")
	if generateFn == nil {
		return nil, PluginMeta{}, fmt.Errorf("dataset generator loader: plugin instance.generateOne must be a function")
	}

	p := &Plugin{
		vm:         vm,
		meta:       meta,
		instance:   instanceObj,
		generateFn: generateFn,
	}
	return p, meta, nil
}

func (p *Plugin) Meta() PluginMeta {
	if p == nil {
		return PluginMeta{}
	}
	return p.meta
}

func (p *Plugin) GenerateOne(input map[string]any, opts PluginGenerateOptions) (map[string]any, map[string]any, error) {
	if p == nil || p.vm == nil || p.instance == nil || p.generateFn == nil {
		return nil, nil, fmt.Errorf("dataset generator: plugin not initialized")
	}

	options := map[string]any{
		"profile":       strings.TrimSpace(opts.Profile),
		"engineOptions": opts.EngineOptions,
		"tags":          opts.Tags,
		"seed":          opts.Seed,
		"rng":           buildRNGBridge(p.vm, opts.RNG),
		"config":        opts.Config,
	}

	ret, err := p.generateFn(p.instance, p.vm.ToValue(input), p.vm.ToValue(options))
	if err != nil {
		return nil, nil, errors.Wrap(err, "dataset generator: call failed")
	}
	decoded, err := decodeJSReturnValue(ret)
	if err != nil {
		return nil, nil, errors.Wrap(err, "dataset generator: invalid return value")
	}
	return decodeGeneratorOutput(decoded)
}

func decodePluginMeta(descriptorObj *goja.Object) (PluginMeta, error) {
	apiVersion := decodeOptionalJSString(descriptorObj.Get("apiVersion"))
	kind := decodeOptionalJSString(descriptorObj.Get("kind"))
	id := decodeOptionalJSString(descriptorObj.Get("id"))
	name := decodeOptionalJSString(descriptorObj.Get("name"))
	registryIdentifier := decodeOptionalJSString(descriptorObj.Get("registryIdentifier"))

	if apiVersion == "" {
		return PluginMeta{}, fmt.Errorf("dataset generator loader: descriptor.apiVersion is required")
	}
	if apiVersion != PluginAPIVersion {
		return PluginMeta{}, fmt.Errorf("dataset generator loader: unsupported apiVersion %q (expected %q)", apiVersion, PluginAPIVersion)
	}
	if kind != "dataset-generator" {
		return PluginMeta{}, fmt.Errorf("dataset generator loader: descriptor.kind must be %q", "dataset-generator")
	}
	if id == "" {
		return PluginMeta{}, fmt.Errorf("dataset generator loader: descriptor.id is required")
	}
	if name == "" {
		return PluginMeta{}, fmt.Errorf("dataset generator loader: descriptor.name is required")
	}

	if cv := descriptorObj.Get("create"); cv == nil || goja.IsUndefined(cv) || goja.IsNull(cv) {
		return PluginMeta{}, fmt.Errorf("dataset generator loader: descriptor.create is required")
	}
	if _, ok := goja.AssertFunction(descriptorObj.Get("create")); !ok {
		return PluginMeta{}, fmt.Errorf("dataset generator loader: descriptor.create must be a function")
	}

	return PluginMeta{
		APIVersion:         apiVersion,
		Kind:               kind,
		ID:                 id,
		Name:               name,
		RegistryIdentifier: firstNonEmpty(registryIdentifier, DefaultRegistryIdentifier),
	}, nil
}

func decodeGeneratorOutput(v any) (map[string]any, map[string]any, error) {
	out, ok := normalizeStringAnyMap(v)
	if !ok {
		return nil, nil, fmt.Errorf("dataset generator must return an object or {row, metadata}, got %T", v)
	}

	if rowRaw, hasRow := out["row"]; hasRow {
		row, ok := normalizeStringAnyMap(rowRaw)
		if !ok {
			return nil, nil, fmt.Errorf("dataset generator output.row must be an object, got %T", rowRaw)
		}
		if len(row) == 0 {
			return nil, nil, fmt.Errorf("dataset generator output.row is empty")
		}
		metadata := map[string]any(nil)
		if metadataRaw, hasMetadata := out["metadata"]; hasMetadata && metadataRaw != nil {
			if mm, ok := normalizeStringAnyMap(metadataRaw); ok {
				metadata = mm
			} else {
				metadata = map[string]any{"value": metadataRaw}
			}
		}
		return row, metadata, nil
	}

	if len(out) == 0 {
		return nil, nil, fmt.Errorf("dataset generator returned an empty object")
	}
	return out, nil, nil
}

func buildRNGBridge(vm *goja.Runtime, rng any) any {
	if vm == nil || rng == nil {
		return rng
	}
	bridge, ok := rng.(RNG)
	if !ok {
		return rng
	}
	obj := vm.NewObject()
	_ = obj.Set("intN", func(max int) int {
		return bridge.IntN(max)
	})
	_ = obj.Set("float64", func() float64 {
		return bridge.Float64()
	})
	_ = obj.Set("choice", func(values []any) any {
		return bridge.Choice(values)
	})
	_ = obj.Set("shuffle", func(values []any) {
		bridge.Shuffle(values)
	})
	return obj
}

func normalizeStringAnyMap(v any) (map[string]any, bool) {
	switch x := v.(type) {
	case map[string]any:
		return x, true
	case map[any]any:
		out := make(map[string]any, len(x))
		for k, vv := range x {
			out[fmt.Sprintf("%v", k)] = vv
		}
		return out, true
	default:
		return nil, false
	}
}

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
