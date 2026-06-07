// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test unwrapping int! to int via err-check, .value, and ??.

package main

import (
	"errors"
	"fmt"
)

var errBoom = errors.New("boom")

func myPrint(a int) {
	fmt.Println(a)
}

func errCheck() {
	var a int! = 7
	if a.err == nil {
		myPrint(a.value)
	}
}

func coalesce() {
	var a int! = errBoom
	myPrint(a ?? 0)
}

func coalesceOk() {
	var a int! = 9
	myPrint(a ?? 0)
}

func valuePanic() {
	defer func() {
		if recover() == nil {
			panic("expected panic from a.value")
		}
	}()
	var a int! = errBoom
	myPrint(a.value)
}

func main() {
	errCheck()
	coalesce()
	coalesceOk()
	valuePanic()
	fmt.Println("ok")
}
