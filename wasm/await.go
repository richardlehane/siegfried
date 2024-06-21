//go:build js

package main

import (
	"errors"
	"syscall/js"
)

func await(promise js.Value) (js.Value, error) {
	successCh, errCh := make(chan js.Value), make(chan js.Value)
	succFn := js.FuncOf(func(_ js.Value, x []js.Value) any {
		successCh <- x[0]
		return nil
	})
	defer succFn.Release()
	errFn := js.FuncOf(func(_ js.Value, x []js.Value) any {
		errCh <- x[0]
		return nil
	})
	defer errFn.Release()

	go promise.Call("then", succFn, errFn)

	for {
		select {
		case val := <-successCh:
			return val, nil
		case err := <-errCh:
			reason := err.Get("reason")
			if reason.Type() == js.TypeObject {
				if message := reason.Get("message"); message.Type() == js.TypeString {
					return js.Undefined(), errors.New(message.String())
				}
			}
			return js.Undefined(), errors.New(js.Global().Call("String", reason).String())
		}
	}
}
