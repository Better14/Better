// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test the T! result type and the postfix !.value operator.
//
// Rules being checked:
//   * A function declared with result type T! has effective results (T, error).
//   * Inside such a function, expr!.value is allowed when expr has type (T2, error).
//     - If the error is non-nil, the enclosing function returns (zeroT, err).
//     - Otherwise, expr!.value evaluates to the T2 value.
//   * `return v, nil` still works for a T! function (it is just (T, error)).

package main

import (
	"errors"
	"fmt"
)

var errBoom = errors.New("boom")

func okPair() (int, error)   { return 7, nil }
func errPair() (int, error)  { return 0, errBoom }
func okString() (string, error)  { return "hello", nil }

// unwrapOK returns 7 successfully.
func unwrapOK() int! {
	v := okPair()!.value
	return v, nil
}

// unwrapErr early-returns errBoom.
func unwrapErr() int! {
	v := errPair()!.value
	return v, nil
}

// callsTQuestion calls a T! function and unwraps it.
func callsTQuestion() int! {
	v := unwrapOK()!.value
	return v * 2, nil
}

// callsTQuestionErr propagates an error through two !-using calls.
func callsTQuestionErr() int! {
	v := unwrapErr()!.value
	return v, nil
}

// forceOnPair is not supported: standalone ! is invalid.
// Use !.value instead (see unwrapOK).

func main() {
	v, err := unwrapOK()
	if err != nil || v != 7 {
		panic("unwrapOK")
	}
	if _, err := unwrapErr(); err != errBoom {
		panic("unwrapErr")
	}
	v, err = callsTQuestion()
	if err != nil || v != 14 {
		panic("callsTQuestion")
	}
	if _, err := callsTQuestionErr(); err != errBoom {
		panic("callsTQuestionErr")
	}
	fmt.Println("ok")
}
