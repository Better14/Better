// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:nilable_pointers enable

// Test nilable pointer types (*T vs *T?) and nil-check narrowing.

package main

type node struct {
	value int
}

func useRequired(p *node) int {
	return p.value
}

func useNilable(p *node?) int {
	if p != nil {
		return useRequired(p)
	}
	return -1
}

func main() {
	var optional *node? = nil
	var required *node = &node{value: 1}
	_ = optional
	_ = required

	if optional != nil {
		_ = useRequired(optional)
	}

	_ = useNilable(&node{value: 2})
}
