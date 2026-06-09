// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// LINQ chains with => lambda syntax and custom extension helpers.

package main

import "linq"

// ScaleBy multiplies each element by k (extension on linq.Lazy[T]).
func (l linq.Lazy[T]) ScaleBy[T ~int | ~float64](k T) linq.Lazy[T] {
	return linq.LazySelectBy(l, x => x*k)
}

// FirstMatch returns the first element satisfying pred (extension on []T).
func (s []T) FirstMatch[T any](pred func(T) bool) T {
	return linq.FromSlice(s).Where(pred).First()
}

// EvensDouble filters evens and doubles (extension on []int).
func (s []int) EvensDouble() linq.Lazy[int] {
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

	// custom slice extension
	if nums.EvensDouble().First() != 4 {
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

	sum := linq.SumLazy(nums.Where(n => n%2 == 0).Select(n => n * n))
	if sum != 4+16+36+64 {
		panic(sum)
	}

	if nums.FirstMatch(n => n > 5) != 6 {
		panic("FirstMatch")
	}

	// distinct on lazy chain, group by on slice
	dist := linq.DistinctLazy(nums.Select(n => n / 2)).ToList()
	if len(dist) != 5 {
		panic(dist)
	}
	groups := linq.LazyGroupBy(linq.FromSlice(nums), n => n%2).ToList()
	if len(groups) != 2 {
		panic(groups)
	}
}
