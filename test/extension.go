// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify extension methods on slices and foreign types.

package main

import "linq"

func main() {
	nums := []int{1, 2, 3, 4, 5}
	out := nums.Where(func(n int) bool { return n < 5 }).Select(func(n int) int { return n + 1 }).ToList()
	if len(out) != 4 || out[0] != 2 || out[3] != 5 {
		panic(out)
	}
	lazy := linq.FromSlice(nums).Where(func(n int) bool { return n%2 == 0 }).Select(func(n int) int { return n * 2 })
	if lazy.First() != 4 {
		panic("lazy chain failed")
	}
}
