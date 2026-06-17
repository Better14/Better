// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// LINQ chains with => lambda syntax and custom extension helpers.

package main

import (
	"iter"
	"linq"
)

// ScaleBy multiplies each element by k (extension on iter.Seq[T]).
func (seq iter.Seq[T]) ScaleBy[T ~int | ~float64](k T) iter.Seq[T] {
	return seq.Select(x => x * k)
}

// FirstMatch returns the first element satisfying pred.
func FirstMatch[T any](s []T, pred func(T) bool) T {
	return s.Where(pred).First()
}

// EvensDouble filters evens and doubles (extension on []int).
func EvensDouble(s []int) iter.Seq[int] {
	return s.Where(n => n%2 == 0).Select(n => n * 2)
}

func main() {
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8}

	// filter → map → first
	first := nums.Where(n => n%2 == 0).Select(n => n * 2).First()
	if first != 4 {
		panic(first)
	}

	// filter → filter → take → list
	out := nums.Where(n => n > 1).Where(n => n%2 == 0).Take(3).ToList()
	if len(out) != 3 || out[0] != 2 || out[2] != 6 {
		panic(out)
	}

	// custom helper
	if EvensDouble(nums).First() != 4 {
		panic("EvensDouble")
	}

	// custom ScaleBy extension mid-chain
	got := nums.Where(n => n%2 == 1).ScaleBy(10).Select(n => n + 1).First()
	if got != 11 {
		panic(got)
	}

	// any / all
	if !nums.Any(n => n > 7) {
		panic("Any")
	}
	if !nums.All(n => n < 100) {
		panic("All")
	}

	sum := nums.Where(n => n%2 == 0).Select(n => n * n).Sum()
	if sum != 4+16+36+64 {
		panic(sum)
	}

	if FirstMatch(nums, n => n > 5) != 6 {
		panic("FirstMatch")
	}

	// distinct on lazy chain, group by on slice
	dist := nums.Select(n => n / 2).Distinct().ToList()
	if len(dist) != 5 {
		panic(dist)
	}
	groups := nums.GroupBy(n => n%2).ToList()
	if len(groups) != 2 {
		panic(groups)
	}
}
