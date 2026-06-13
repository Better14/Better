// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test err! early-return on plain error values.

package main

import (
	"errors"
	"fmt"
)

var errBoom = errors.New("boom")

func possibleError() error {
	return nil
}

func possibleErrorBoom() error {
	return errBoom
}

func withErrForce() int! {
	err := possibleError()
	err!
	return 42
}

func withErrForceFail() int! {
	err := possibleErrorBoom()
	err!
	return 42
}

func withCallErrForce() int! {
	possibleError()!
	return 7
}

func withTupleErrForce() (int, error) {
	err := possibleErrorBoom()
	err!
	return 1, nil
}

func main() {
	v, err := withErrForce()
	if err != nil || v != 42 {
		panic(fmt.Sprintf("withErrForce: v=%d err=%v", v, err))
	}

	v, err = withErrForceFail()
	if err != errBoom || v != 0 {
		panic(fmt.Sprintf("withErrForceFail: v=%d err=%v", v, err))
	}

	v, err = withCallErrForce()
	if err != nil || v != 7 {
		panic(fmt.Sprintf("withCallErrForce: v=%d err=%v", v, err))
	}

	v, err = withTupleErrForce()
	if err != errBoom || v != 0 {
		panic(fmt.Sprintf("withTupleErrForce: v=%d err=%v", v, err))
	}

	fmt.Println("ok")
}
