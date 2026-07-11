// run

// Verify extension methods on iter.Seq and foreign types.

package main

import "linq"

func main() {
	nums := []int{1, 2, 3, 4, 5}
	out := linq.From(nums).Where(func(n int) bool { return n < 5 }).Select(func(n int) int { return n + 1 }).ToList()
	if len(out) != 4 || out[0] != 2 || out[3] != 5 {
		panic(out)
	}
	seq := linq.From(nums).Where(func(n int) bool { return n%2 == 0 }).Select(func(n int) int { return n * 2 })
	if seq.First() != 4 {
		panic("seq chain failed")
	}
}
