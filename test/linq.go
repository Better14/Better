// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "linq"

func main() {
	nums := []int{1, 2, 3, 4, 5}
	out := nums.Where(func(n int) bool { return n < 5 }).Select(func(n int) int { return n + 1 }).ToList()
	if len(out) != 4 || out[0] != 2 || out[3] != 5 {
		panic(out)
	}
	first := nums.Where(func(n int) bool { return n%2 == 0 }).Select(func(n int) int { return n * 2 }).First()
	if first != 4 {
		panic(first)
	}
}
