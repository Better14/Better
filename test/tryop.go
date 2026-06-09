// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test the T! result type and the postfix ! operator.
//
// Rules being checked:
//   * A function declared with result type T! has effective results (T, error).
//   * Inside such a function, expr! is allowed when expr has type (T2, error).
//     - If the error is non-nil, the enclosing function returns (zeroT, err).
//     - Otherwise, expr! evaluates to the T2 value.
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
	v := okPair()!
	return v, nil
}

// unwrapErr early-returns errBoom.
func unwrapErr() int! {
	v := errPair()!
	return v, nil
}

// callsTQuestion calls a T! function and unwraps it.
func callsTQuestion() int! {
	v := unwrapOK()!
	return v * 2, nil
}

// callsTQuestionErr propagates an error through two !-using calls.
func callsTQuestionErr() int! {
	v := unwrapErr()!
	return v, nil
}

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
