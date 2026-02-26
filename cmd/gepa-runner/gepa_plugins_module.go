package main

import (
	"fmt"
	"strings"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
)

const gepaPluginsModuleName = "gepa/plugins"

func registerGepaPluginsModule(reg *require.Registry) {
	if reg == nil {
		return
	}
	reg.RegisterNativeModule(gepaPluginsModuleName, loadGepaPluginsModule)
}

func loadGepaPluginsModule(vm *goja.Runtime, moduleObj *goja.Object) {
	exports := moduleObj.Get("exports").(*goja.Object)
	mustSet := func(key string, v any) {
		if err := exports.Set(key, v); err != nil {
			panic(vm.NewGoError(fmt.Errorf("set %s: %w", key, err)))
		}
	}

	mustSet("OPTIMIZER_PLUGIN_API_VERSION", optimizerPluginAPIVersion)
	mustSet("defineOptimizerPlugin", func(call goja.FunctionCall) goja.Value {
		descriptor := call.Argument(0)
		if descriptor == nil || goja.IsUndefined(descriptor) || goja.IsNull(descriptor) {
			panic(vm.NewTypeError("plugin descriptor must be an object"))
		}
		descriptorObj := descriptor.ToObject(vm)
		if descriptorObj == nil || descriptorObj.ClassName() != "Object" {
			panic(vm.NewTypeError("plugin descriptor must be an object"))
		}

		apiVersion := readJSStringField(vm, descriptorObj, "apiVersion", false)
		if apiVersion == "" {
			apiVersion = optimizerPluginAPIVersion
		}
		if apiVersion != optimizerPluginAPIVersion {
			panic(vm.NewTypeError(
				"unsupported plugin descriptor apiVersion %q (expected %q)",
				apiVersion, optimizerPluginAPIVersion,
			))
		}

		kind := readJSStringField(vm, descriptorObj, "kind", false)
		if kind == "" {
			kind = "optimizer"
		}
		if kind != "optimizer" {
			panic(vm.NewTypeError("plugin descriptor kind must be %q, got %q", "optimizer", kind))
		}

		id := readJSStringField(vm, descriptorObj, "id", true)
		name := readJSStringField(vm, descriptorObj, "name", true)
		registryIdentifier := readJSStringField(vm, descriptorObj, "registryIdentifier", false)
		if registryIdentifier == "" {
			registryIdentifier = defaultPluginRegistryIdentifier
		}

		createVal := descriptorObj.Get("create")
		if _, ok := goja.AssertFunction(createVal); !ok {
			panic(vm.NewTypeError("plugin descriptor create must be a function"))
		}

		out := vm.NewObject()
		_ = out.Set("apiVersion", apiVersion)
		_ = out.Set("kind", kind)
		_ = out.Set("id", id)
		_ = out.Set("name", name)
		_ = out.Set("registryIdentifier", registryIdentifier)
		_ = out.Set("create", createVal)
		return freezeJSObject(vm, out)
	})
}

func readJSStringField(vm *goja.Runtime, obj *goja.Object, key string, required bool) string {
	v := obj.Get(key)
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		if required {
			panic(vm.NewTypeError("plugin descriptor %s is required", key))
		}
		return ""
	}
	s, ok := v.Export().(string)
	if !ok {
		panic(vm.NewTypeError("plugin descriptor %s must be a string", key))
	}
	out := strings.TrimSpace(s)
	if required && out == "" {
		panic(vm.NewTypeError("plugin descriptor %s is required", key))
	}
	return out
}

func freezeJSObject(vm *goja.Runtime, obj *goja.Object) goja.Value {
	objectVal := vm.Get("Object")
	objectObj := objectVal.ToObject(vm)
	freezeVal := objectObj.Get("freeze")
	freezeFn, ok := goja.AssertFunction(freezeVal)
	if !ok {
		return obj
	}
	ret, err := freezeFn(objectVal, obj)
	if err != nil {
		panic(err)
	}
	return ret
}
