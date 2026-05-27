package lxruntime

import (
	"fmt"
	"time"

	"github.com/dop251/goja"
)

func awaitPromise(vm *goja.Runtime, promise goja.Value, timeout time.Duration) (goja.Value, error) {
	if promise == nil || goja.IsUndefined(promise) || goja.IsNull(promise) {
		return goja.Undefined(), nil
	}

	obj := promise.ToObject(vm)
	if obj == nil {
		return promise, nil
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
			return nil, err
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

	return promise, nil
}

type promiseResult struct {
	value goja.Value
	err   error
}
