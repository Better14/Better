// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "fmt"

func main() {
	a := if 5 < 6 { 1 } else { 2 }
	if a != 1 {
		panic(fmt.Sprintf("if expr: got %d", a))
	}
	b := if false { "no" } else { "yes" }
	if b != "yes" {
		panic(fmt.Sprintf("if expr string: got %q", b))
	}
	fmt.Println("ok")
}
