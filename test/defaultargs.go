// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

const defaultC = 5

func f(a, b int, c int = defaultC, d int = 7) int {
	return a + b + c + d
}

func main() {
	if got := f(1, 2); got != 15 {
		panic(got)
	}
	if got := f(1, 2, 3); got != 13 {
		panic(got)
	}
	if got := f(1, 2, 3, 4); got != 10 {
		panic(got)
	}
}
