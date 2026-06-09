// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// LINQ chains with => lambda syntax and a custom generic helper.

package main

import "linq"

// ScaleBy multiplies each element by k (custom generic LINQ helper).
func ScaleBy[T ~int | ~float64](l linq.Lazy[T], k T) linq.Lazy[T] {
	return linq.LazySelectBy(l, func(x T) T { return x * k })
}

// FirstMatch returns the first element satisfying pred (custom generic helper).
func FirstMatch[T any](s []T, pred func(T) bool) T {
	return linq.FromSlice(s).Where(pred).First()
}

// EvensDouble filters evens and doubles (custom slice extension using => in Where).
func (s []int) EvensDouble() linq.Lazy[int] {
	return s.Where(n => n%2 == 0).Select(func(n int) int { return n * 2 })
}

func main() {
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8}

	// filter → map → first (=> in Where; Select uses func until method-type inference lands)
	first := nums.Where(n => n%2 == 0).Select(func(n int) int { return n * 2 }).First()
	if first != 4 {
		panic(first)
	}

	// filter → filter → take → list (=> throughout predicates)
	out := nums.Where(n => n > 1).Where(n => n%2 == 0).Take(3).ToList()
	if len(out) != 3 || out[0] != 2 || out[2] != 6 {
		panic(out)
	}

	// custom slice extension
	if nums.EvensDouble().First() != 4 {
		panic("EvensDouble")
	}

	// custom generic ScaleBy mid-chain with => predicates
	got := ScaleBy(nums.Where(n => n%2 == 1), 10).Select(func(n int) int { return n + 1 }).First()
	if got != 11 {
		panic(got)
	}

	// any / all with =>
	if !nums.Any(n => n > 7) {
		panic("Any")
	}
	if !nums.All(n => n < 100) {
		panic("All")
	}

	sum := linq.SumLazy(nums.Where(n => n%2 == 0).Select(func(n int) int { return n * n }))
	if sum != 4+16+36+64 {
		panic(sum)
	}

	if FirstMatch(nums, n => n > 5) != 6 {
		panic("FirstMatch")
	}

	// distinct on lazy chain, group by on slice
	dist := linq.DistinctLazy(nums.Select(func(n int) int { return n / 2 })).ToList()
	if len(dist) != 5 {
		panic(dist)
	}
	groups := linq.LazyGroupBy(linq.FromSlice(nums), func(n int) int { return n % 2 }).ToList()
	if len(groups) != 2 {
		panic(groups)
	}
}
