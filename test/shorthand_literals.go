// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "set"

func spreadInts(xs ...int) int { return len(xs) }

func main() {
	a := ["string", "asdf"]
	if len(a) != 2 || a[0] != "string" || a[1] != "asdf" {
		panic("array literal")
	}

	m := {"a": "b"}
	if len(m) != 1 || m["a"] != "b" {
		panic("map literal")
	}

	s := {"a", "b", "c", "c"}
	if s.Len() != 3 || !s.Contains("a") || !s.Contains("c") {
		panic("set literal")
	}

	t := {}string{"x", "y"}
	if t.Len() != 2 {
		panic("typed set literal")
	}

	b := ["fruit", ...a]
	if len(b) != 3 || b[2] != "asdf" {
		panic("array spread")
	}

	set1 := {"a"}
	set2 := {"b", ...set1}
	if set2.Len() != 2 || !set2.Contains("a") || !set2.Contains("b") {
		panic("set spread")
	}

	nums := [1, 2, 3]
	if spreadInts(...nums) != 3 {
		panic("prefix spread call")
	}
}
