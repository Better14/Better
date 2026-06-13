// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test single-value return sugar in T! functions:
//   return v   => return v, nil
//   return err => return zero, err

package main

import (
	"errors"
	"fmt"
)

var errBoom = errors.New("boom")

func returnValue() string! {
	return "ok"
}

func returnError() string! {
	return errBoom
}

func returnInt() int! {
	return 42
}

func main() {
	s, err := returnValue()
	if err != nil || s != "ok" {
		panic(fmt.Sprintf("returnValue: s=%q err=%v", s, err))
	}

	s, err = returnError()
	if err != errBoom || s != "" {
		panic(fmt.Sprintf("returnError: s=%q err=%v", s, err))
	}

	n, err := returnInt()
	if err != nil || n != 42 {
		panic(fmt.Sprintf("returnInt: n=%d err=%v", n, err))
	}

	fmt.Println("ok")
}
