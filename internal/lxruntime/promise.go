package lxruntime

import (
	"fmt"
	"time"

	"github.com/dop251/goja"
)

func awaitPromise(vm *goja.Runtime, promise goja.Value, timeout time.Duration) (goja.Value, error) {
	value, resultCh, err := preparePromiseAwait(vm, promise)
	if err != nil {
		return nil, err
	}
	if resultCh == nil {
		return value, nil
	}

	select {
	case result := <-resultCh:
		if result.err != nil {
			return nil, result.err
		}
		return result.value, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("promise timeout")
	}
}

func preparePromiseAwait(vm *goja.Runtime, promise goja.Value) (goja.Value, <-chan promiseResult, error) {
	if promise == nil || goja.IsUndefined(promise) || goja.IsNull(promise) {
		return goja.Undefined(), nil, nil
	}

	obj := promise.ToObject(vm)
	if obj == nil {
		return promise, nil, nil
	}

	if thenFn, ok := goja.AssertFunction(obj.Get("then")); ok {
		resultCh := make(chan promiseResult, 1)
		resolve := vm.ToValue(func(call goja.FunctionCall) goja.Value {
			value := goja.Undefined()
			if len(call.Arguments) > 0 {
				value = call.Arguments[0]
			}
			resultCh <- promiseResult{value: value}
			return goja.Undefined()
		})
		reject := vm.ToValue(func(call goja.FunctionCall) goja.Value {
			message := "promise rejected"
			if len(call.Arguments) > 0 {
				message = call.Arguments[0].String()
			}
			resultCh <- promiseResult{err: fmt.Errorf("%s", message)}
			return goja.Undefined()
		})

		if _, err := thenFn(obj, resolve, reject); err != nil {
			return nil, nil, err
		}
		return goja.Undefined(), resultCh, nil
	}

	return promise, nil, nil
}

type promiseResult struct {
	value goja.Value
	err   error
}
