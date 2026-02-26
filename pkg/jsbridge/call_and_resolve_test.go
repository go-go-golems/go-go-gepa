package jsbridge

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"
	"github.com/go-go-golems/go-go-goja/pkg/runtimeowner"
)

func newTestRunner(t *testing.T) (*eventloop.EventLoop, runtimeowner.Runner) {
	t.Helper()
	loop := eventloop.NewEventLoop()
	go loop.Start()
	vm := goja.New()
	runner := runtimeowner.NewRunner(vm, loop, runtimeowner.Options{
		Name:          "jsbridge-test",
		RecoverPanics: true,
	})
	return loop, runner
}

func TestCallAndResolveValue(t *testing.T) {
	loop, runner := newTestRunner(t)
	defer loop.Stop()

	got, err := CallAndResolve(context.Background(), CallAndResolveOptions{
		Op:             "test.value",
		Runner:         runner,
		DefaultTimeout: time.Second,
	}, func(vm *goja.Runtime) (goja.Value, error) {
		return vm.ToValue(map[string]any{"ok": true}), nil
	})
	if err != nil {
		t.Fatalf("CallAndResolve returned error: %v", err)
	}
	m, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", got)
	}
	if m["ok"] != true {
		t.Fatalf("expected ok=true, got %#v", m["ok"])
	}
}

func TestCallAndResolvePromiseFulfilledAsync(t *testing.T) {
	loop, runner := newTestRunner(t)
	defer loop.Stop()

	got, err := CallAndResolve(context.Background(), CallAndResolveOptions{
		Op:             "test.promise.fulfilled",
		Runner:         runner,
		DefaultTimeout: time.Second,
	}, func(vm *goja.Runtime) (goja.Value, error) {
		p, resolve, _ := vm.NewPromise()
		go func() {
			time.Sleep(20 * time.Millisecond)
			loop.RunOnLoop(func(vm2 *goja.Runtime) {
				_ = resolve(vm2.ToValue("done"))
			})
		}()
		return vm.ToValue(p), nil
	})
	if err != nil {
		t.Fatalf("CallAndResolve returned error: %v", err)
	}
	if got != "done" {
		t.Fatalf("expected done, got %#v", got)
	}
}

func TestCallAndResolvePromiseRejectedAsync(t *testing.T) {
	loop, runner := newTestRunner(t)
	defer loop.Stop()

	_, err := CallAndResolve(context.Background(), CallAndResolveOptions{
		Op:             "test.promise.rejected",
		Runner:         runner,
		DefaultTimeout: time.Second,
	}, func(vm *goja.Runtime) (goja.Value, error) {
		p, _, reject := vm.NewPromise()
		go func() {
			time.Sleep(20 * time.Millisecond)
			loop.RunOnLoop(func(vm2 *goja.Runtime) {
				_ = reject(vm2.ToValue("boom"))
			})
		}()
		return vm.ToValue(p), nil
	})
	if err == nil {
		t.Fatalf("expected rejection error")
	}
	if !strings.Contains(err.Error(), "promise rejected") || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCallAndResolvePromiseTimeout(t *testing.T) {
	loop, runner := newTestRunner(t)
	defer loop.Stop()

	_, err := CallAndResolve(context.Background(), CallAndResolveOptions{
		Op:             "test.promise.timeout",
		Runner:         runner,
		DefaultTimeout: 40 * time.Millisecond,
	}, func(vm *goja.Runtime) (goja.Value, error) {
		p, _, _ := vm.NewPromise()
		return vm.ToValue(p), nil
	})
	if err == nil {
		t.Fatalf("expected timeout error")
	}
	if !strings.Contains(err.Error(), "did not settle before deadline") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCallAndResolvePromiseCanceled(t *testing.T) {
	loop, runner := newTestRunner(t)
	defer loop.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	_, err := CallAndResolve(ctx, CallAndResolveOptions{
		Op:             "test.promise.cancel",
		Runner:         runner,
		DefaultTimeout: time.Second,
	}, func(vm *goja.Runtime) (goja.Value, error) {
		p, _, _ := vm.NewPromise()
		return vm.ToValue(p), nil
	})
	if err == nil {
		t.Fatalf("expected cancel error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled error, got %v", err)
	}
}
