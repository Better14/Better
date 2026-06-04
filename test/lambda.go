// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

var add func(int, int) int = (a, b) => a + b

func main() {
	if add(1, 2) != 3 {
		panic(add(1, 2))
	}
}
