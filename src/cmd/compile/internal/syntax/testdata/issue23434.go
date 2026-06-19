// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test case for go.dev/issue/23434: Better synchronization of
// parser after missing type. There should be exactly
// one error each time, with now follow errors.

package p

type T // ERROR unexpected newline in type declaration

type Map map[int] /* ERROR missing map value type */

// Examples from go.dev/issue/23434:

func g() {
	m := make(map[string] /* ERROR missing map value type */ !)
	for {
		x := 1
		print(x)
	}
}

func f() {
	m := make(map[string] /* ERROR missing map value type */ )
	for {
		x := 1
		print(x)
	}
}
