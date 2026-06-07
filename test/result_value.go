// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test T! as a value type with .value, .err fields and destructuring.

package main

import (
	"errors"
	"fmt"
)

var errBoom = errors.New("boom")

func doSomething(n int) {}
func handleError(err error) {}

type widget struct{ n int }

func someFunc() (widget, error) { return widget{3}, nil }

func okResult() int! {
	var a int! = 0
	if a.err == nil {
		doSomething(a.value)
	} else {
		handleError(a.err)
	}
	val, err := a
	if err != nil || val != 0 {
		panic(fmt.Sprintf("destructure ok: val=%d err=%v", val, err))
	}
	return val
}

func errResult() int! {
	var a int! = errBoom
	if a.err == nil {
		panic("expected err")
	}
	handleError(a.err)
	val, err := a
	if err != errBoom || val != 0 {
		panic(fmt.Sprintf("destructure err: val=%d err=%v", val, err))
	}
	return a
}

func propagateField() int! {
	y := someFunc()!.n
	return y
}

func main() {
	v, err := okResult()
	if err != nil || v != 0 {
		panic("okResult")
	}
	if _, err := errResult(); err != errBoom {
		panic("errResult")
	}
	n, err := propagateField()
	if err != nil || n != 3 {
		panic(fmt.Sprintf("propagateField: n=%d err=%v", n, err))
	}
	fmt.Println("ok")
}
