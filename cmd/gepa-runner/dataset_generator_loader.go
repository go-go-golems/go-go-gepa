package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
	"github.com/pkg/errors"
)

type datasetGeneratorPluginMeta struct {
	APIVersion         string
	Kind               string
	ID                 string
	Name               string
	RegistryIdentifier string
}

type datasetGeneratorGenerateOptions struct {
	Profile       string
	EngineOptions map[string]any
	Tags          map[string]any
	Seed          int64
	RNG           any
	Config        any
}

type datasetGeneratorPlugin struct {
	rt       *jsRuntime
	meta     datasetGeneratorPluginMeta
	instance *goja.Object

	generateOneFn goja.Callable
}

func loadDatasetGeneratorPlugin(rt *jsRuntime, absScriptPath string, hostContext map[string]any) (*datasetGeneratorPlugin, datasetGeneratorPluginMeta, error) {
	if rt == nil || rt.vm == nil || rt.reqMod == nil {
		return nil, datasetGeneratorPluginMeta{}, errors.New("dataset generator loader: runtime is nil")
	}
	if strings.TrimSpace(absScriptPath) == "" {
		return nil, datasetGeneratorPluginMeta{}, errors.New("dataset generator loader: script path is empty")
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
			return nil, datasetGeneratorPluginMeta{}, errors.Wrap(err, "dataset generator loader: require script module")
		}
	}

	descriptorObj := exported.ToObject(rt.vm)
	if descriptorObj == nil {
		return nil, datasetGeneratorPluginMeta{}, fmt.Errorf("dataset generator loader: script module did not export an object descriptor")
	}

	meta, err := decodeDatasetGeneratorMeta(descriptorObj)
	if err != nil {
		return nil, datasetGeneratorPluginMeta{}, err
	}

	createVal := descriptorObj.Get("create")
	createFn, ok := goja.AssertFunction(createVal)
	if !ok {
		return nil, datasetGeneratorPluginMeta{}, fmt.Errorf("dataset generator loader: descriptor.create must be a function")
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
		return nil, datasetGeneratorPluginMeta{}, errors.Wrap(err, "dataset generator loader: descriptor.create failed")
	}
	instanceObj := instanceVal.ToObject(rt.vm)
	if instanceObj == nil {
		return nil, datasetGeneratorPluginMeta{}, fmt.Errorf("dataset generator loader: descriptor.create must return an object instance")
	}

	generateOneFn := findOptionalCallable(instanceObj, "generateOne", "generate")
	if generateOneFn == nil {
		return nil, datasetGeneratorPluginMeta{}, fmt.Errorf("dataset generator loader: plugin instance.generateOne must be a function")
	}

	p := &datasetGeneratorPlugin{
		rt:            rt,
		meta:          meta,
		instance:      instanceObj,
		generateOneFn: generateOneFn,
	}

	return p, meta, nil
}

func decodeDatasetGeneratorMeta(descriptorObj *goja.Object) (datasetGeneratorPluginMeta, error) {
	apiVersion := decodeOptionalJSString(descriptorObj.Get("apiVersion"))
	kind := decodeOptionalJSString(descriptorObj.Get("kind"))
	id := decodeOptionalJSString(descriptorObj.Get("id"))
	name := decodeOptionalJSString(descriptorObj.Get("name"))
	registryIdentifier := decodeOptionalJSString(descriptorObj.Get("registryIdentifier"))

	if apiVersion == "" {
		return datasetGeneratorPluginMeta{}, fmt.Errorf("dataset generator loader: descriptor.apiVersion is required")
	}
	if apiVersion != datasetGeneratorPluginAPIVersion {
		return datasetGeneratorPluginMeta{}, fmt.Errorf("dataset generator loader: unsupported apiVersion %q (expected %q)", apiVersion, datasetGeneratorPluginAPIVersion)
	}
	if kind != "dataset-generator" {
		return datasetGeneratorPluginMeta{}, fmt.Errorf("dataset generator loader: descriptor.kind must be %q", "dataset-generator")
	}
	if id == "" {
		return datasetGeneratorPluginMeta{}, fmt.Errorf("dataset generator loader: descriptor.id is required")
	}
	if name == "" {
		return datasetGeneratorPluginMeta{}, fmt.Errorf("dataset generator loader: descriptor.name is required")
	}

	if cv := descriptorObj.Get("create"); cv == nil || goja.IsUndefined(cv) || goja.IsNull(cv) {
		return datasetGeneratorPluginMeta{}, fmt.Errorf("dataset generator loader: descriptor.create is required")
	}
	if _, ok := goja.AssertFunction(descriptorObj.Get("create")); !ok {
		return datasetGeneratorPluginMeta{}, fmt.Errorf("dataset generator loader: descriptor.create must be a function")
	}

	return datasetGeneratorPluginMeta{
		APIVersion:         apiVersion,
		Kind:               kind,
		ID:                 id,
		Name:               name,
		RegistryIdentifier: firstNonEmpty(registryIdentifier, defaultPluginRegistryIdentifier),
	}, nil
}

func (p *datasetGeneratorPlugin) GenerateOne(input map[string]any, opts datasetGeneratorGenerateOptions) (map[string]any, map[string]any, error) {
	if p == nil || p.rt == nil || p.instance == nil || p.generateOneFn == nil {
		return nil, nil, fmt.Errorf("dataset generator: plugin not initialized")
	}

	options := map[string]any{
		"profile":       strings.TrimSpace(opts.Profile),
		"engineOptions": opts.EngineOptions,
		"tags":          opts.Tags,
		"seed":          opts.Seed,
		"rng":           buildDatasetGeneratorRNGBridge(p.rt.vm, opts.RNG),
		"config":        opts.Config,
	}
	ret, err := p.generateOneFn(p.instance, p.rt.vm.ToValue(input), p.rt.vm.ToValue(options))
	if err != nil {
		return nil, nil, errors.Wrap(err, "dataset generator: call failed")
	}
	decoded, err := decodeJSReturnValue(ret)
	if err != nil {
		return nil, nil, errors.Wrap(err, "dataset generator: invalid return value")
	}
	return decodeDatasetGeneratorOutput(decoded)
}

func decodeDatasetGeneratorOutput(v any) (map[string]any, map[string]any, error) {
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

func buildDatasetGeneratorRNGBridge(vm *goja.Runtime, rng any) any {
	if vm == nil || rng == nil {
		return rng
	}
	jsRng, ok := rng.(*jsRNG)
	if !ok || jsRng == nil {
		return rng
	}
	obj := vm.NewObject()
	_ = obj.Set("intN", func(max int) int {
		return jsRng.IntN(max)
	})
	_ = obj.Set("float64", func() float64 {
		return jsRng.Float64()
	})
	_ = obj.Set("choice", func(values []any) any {
		return jsRng.Choice(values)
	})
	_ = obj.Set("shuffle", func(values []any) {
		jsRng.Shuffle(values)
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
