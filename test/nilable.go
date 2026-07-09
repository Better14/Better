// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test nilable types (T?), ?. null-conditional access, and ?? coalescing.

package main

import "fmt"

type node struct {
	name  string
	child *node
}

func intPtr(n int) *int {
	p := new(int)
	*p = n
	return p
}

func main() {
	var count int? = intPtr(5)
	if count == nil {
		panic("count")
	}
	n := count ?? 0
	if n != 5 {
		panic(fmt.Sprintf("coalesce int: got %d", n))
	}

	var absent int? = nil
	if absent != nil {
		panic("absent")
	}
	var z int = absent ?? 42
	if z != 42 {
		panic(fmt.Sprintf("coalesce nil: got %d", z))
	}

	if got := root?.child?.name ?? ""; got != "leaf" {
		panic(fmt.Sprintf("elvis chain: %q", got))
	}
	label := root?.child?.name ?? "missing"
	if label != "leaf" {
		panic(fmt.Sprintf("elvis coalesce: %q", label))
	}

	var noChild string? = root?.child?.child?.name
	if noChild != nil {
		panic("elvis nil chain")
	}
	def := noChild ?? "none"
	if def != "none" {
		panic(fmt.Sprintf("elvis default: %q", def))
	}

	fmt.Println("ok")
}

var root = &node{name: "root", child: &node{name: "leaf"}}
