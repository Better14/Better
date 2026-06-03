// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test the T? result type and the postfix ? unwrap operator.
//
// Rules being checked:
//   * A function declared with result type T? has effective results (T, error).
//   * Inside such a function, expr? is allowed when expr has type (T2, error).
//     - If the error is non-nil, the enclosing function returns (zeroT, err).
//     - Otherwise, expr? evaluates to the T2 value.
//   * `return v, nil` still works for a T? function (it is just (T, error)).

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
func unwrapOK() int? {
	v := okPair()?
	return v, nil
}

// unwrapErr early-returns errBoom.
func unwrapErr() int? {
	v := errPair()?
	return v, nil
}

// callsTQuestion calls a T? function and unwraps it.
func callsTQuestion() int? {
	v := unwrapOK()?
	return v * 2, nil
}

// callsTQuestionErr propagates an error through two ?-using calls.
func callsTQuestionErr() int? {
	v := unwrapErr()?
	return v, nil
}

// usesStringResult ensures T? works for non-int Ts too.
func usesStringResult() string? {
	s := okString()?
	return s + " world", nil
}

func check(name string, got, want int, gotErr, wantErr error) {
	if got != want {
		panic(fmt.Sprintf("%s: got value %d, want %d", name, got, want))
	}
	switch {
	case wantErr == nil && gotErr != nil:
		panic(fmt.Sprintf("%s: unexpected error %v", name, gotErr))
	case wantErr != nil && gotErr == nil:
		panic(fmt.Sprintf("%s: missing error, want %v", name, wantErr))
	case wantErr != nil && !errors.Is(gotErr, wantErr):
		panic(fmt.Sprintf("%s: got error %v, want %v", name, gotErr, wantErr))
	}
}

func main() {
	v, err := unwrapOK()
	check("unwrapOK", v, 7, err, nil)

	v, err = unwrapErr()
	check("unwrapErr", v, 0, err, errBoom)

	v, err = callsTQuestion()
	check("callsTQuestion", v, 14, err, nil)

	v, err = callsTQuestionErr()
	check("callsTQuestionErr", v, 0, err, errBoom)

	s, err := usesStringResult()
	if s != "hello world" || err != nil {
		panic(fmt.Sprintf("usesStringResult: got %q, %v; want %q, nil", s, err, "hello world"))
	}
}
