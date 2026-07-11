// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	items := []string{"a", "b", "c"}
	var sum string
	for item in items {
		sum += item
	}
	if sum != "abc" {
		panic("for item in failed")
	}
	var n int
	for i, item in items {
		if item != items[i] {
			panic("for i, item in failed")
		}
		n++
	}
	if n != 3 {
		panic("for i, item in count failed")
	}
	// legacy range still works
	for _, item := range items {
		_ = item
	}
}
