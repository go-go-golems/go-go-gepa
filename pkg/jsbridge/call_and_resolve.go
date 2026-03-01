package jsbridge

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/pkg/runtimeowner"
)

const DefaultPromiseTimeout = 10 * time.Minute

type CallAndResolveOptions struct {
	Op             string
	VM             *goja.Runtime
	Runner         runtimeowner.Runner
	DefaultTimeout time.Duration
}

type CallFunc func(vm *goja.Runtime) (goja.Value, error)

type promiseOutcome struct {
	value any
}

type pendingPromise struct{}

type settledResult struct {
	value any
	err   error
}

func CallAndResolve(ctx context.Context, options CallAndResolveOptions, fn CallFunc) (any, error) {
	if fn == nil {
		return nil, fmt.Errorf("jsbridge: call function is nil")
	}
	op := strings.TrimSpace(options.Op)
	if op == "" {
		op = "jsbridge.call"
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		timeout := options.DefaultTimeout
		if timeout <= 0 {
			timeout = DefaultPromiseTimeout
		}
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		ctx = timeoutCtx
	}

	if options.Runner == nil {
		return callAndResolveWithoutRunner(op, options.VM, fn)
	}

	settleCh := make(chan settledResult, 1)
	ret, err := options.Runner.Call(ctx, op+".invoke", func(_ context.Context, vm *goja.Runtime) (any, error) {
		value, callErr := fn(vm)
		if callErr != nil {
			return nil, callErr
		}
		return preparePromiseOrValue(op, vm, value, settleCh)
	})
	if err != nil {
		return nil, fmt.Errorf("%s: invoke failed: %w", op, err)
	}

	switch x := ret.(type) {
	case promiseOutcome:
		return x.value, nil
	case pendingPromise:
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%s: promise did not settle before deadline: %w", op, ctx.Err())
		case settled := <-settleCh:
			if settled.err != nil {
				return nil, settled.err
			}
			return settled.value, nil
		}
	default:
		return x, nil
	}
}

func callAndResolveWithoutRunner(op string, vm *goja.Runtime, fn CallFunc) (any, error) {
	value, err := fn(vm)
	if err != nil {
		return nil, fmt.Errorf("%s: invoke failed: %w", op, err)
	}
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return nil, fmt.Errorf("%s: returned null/undefined", op)
	}
	promise, isPromise := value.Export().(*goja.Promise)
	if !isPromise {
		return value.Export(), nil
	}

	switch promise.State() {
	case goja.PromiseStateFulfilled:
		return exportValue(promise.Result()), nil
	case goja.PromiseStateRejected:
		return nil, rejectionError(op, exportValue(promise.Result()))
	case goja.PromiseStatePending:
		return nil, fmt.Errorf("%s: pending promise requires runtime runner", op)
	}
	return nil, fmt.Errorf("%s: unknown promise state", op)
}

func preparePromiseOrValue(op string, vm *goja.Runtime, value goja.Value, settleCh chan<- settledResult) (any, error) {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return nil, fmt.Errorf("%s: returned null/undefined", op)
	}

	promise, isPromise := value.Export().(*goja.Promise)
	if !isPromise {
		return promiseOutcome{value: value.Export()}, nil
	}

	switch promise.State() {
	case goja.PromiseStateFulfilled:
		return promiseOutcome{value: exportValue(promise.Result())}, nil
	case goja.PromiseStateRejected:
		return nil, rejectionError(op, exportValue(promise.Result()))
	case goja.PromiseStatePending:
		// Handlers are attached below and completion is reported via settleCh.
	}

	promiseObj := value.ToObject(vm)
	thenValue := promiseObj.Get("then")
	thenFn, ok := goja.AssertFunction(thenValue)
	if !ok {
		return nil, fmt.Errorf("%s: promise object missing callable then()", op)
	}

	deliver := func(result settledResult) {
		select {
		case settleCh <- result:
		default:
		}
	}

	onFulfilled := func(call goja.FunctionCall) goja.Value {
		deliver(settledResult{value: exportValue(call.Argument(0))})
		return goja.Undefined()
	}
	onRejected := func(call goja.FunctionCall) goja.Value {
		deliver(settledResult{err: rejectionError(op, exportValue(call.Argument(0)))})
		return goja.Undefined()
	}

	if _, err := thenFn(promiseObj, vm.ToValue(onFulfilled), vm.ToValue(onRejected)); err != nil {
		return nil, fmt.Errorf("%s: failed to attach promise handlers: %w", op, err)
	}

	return pendingPromise{}, nil
}

func exportValue(v goja.Value) any {
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return nil
	}
	return v.Export()
}

func rejectionError(op string, value any) error {
	if value == nil {
		return fmt.Errorf("%s: promise rejected", op)
	}
	switch x := value.(type) {
	case error:
		return fmt.Errorf("%s: promise rejected: %w", op, x)
	case string:
		if strings.TrimSpace(x) == "" {
			return fmt.Errorf("%s: promise rejected", op)
		}
		return fmt.Errorf("%s: promise rejected: %s", op, x)
	default:
		if blob, err := json.Marshal(x); err == nil {
			return fmt.Errorf("%s: promise rejected: %s", op, string(blob))
		}
		return fmt.Errorf("%s: promise rejected: %v", op, x)
	}
}
