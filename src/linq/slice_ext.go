// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

import (
	"cmp"
	"slices"
)

// Extension methods on []T provide LINQ syntax: nums.Where(...).Select(...).

func (s []T) Where(pred func(T) bool) Lazy[T] {
	return LazyWhere(FromSlice(s), pred)
}

func (s []T) Select[U any](fn func(T) U) Lazy[U] {
	return LazySelectBy(FromSlice(s), fn)
}

func (s []T) OrderBy[K cmp.Ordered](key func(T) K) Lazy[T] {
	return LazyOrderBy(FromSlice(s), key)
}

func (s []T) OrderByDescending[K cmp.Ordered](key func(T) K) Lazy[T] {
	return LazyOrderByDescending(FromSlice(s), key)
}

func (s []T) Take(n int) Lazy[T] {
	return LazyTake(FromSlice(s), n)
}

func (s []T) Skip(n int) Lazy[T] {
	return LazySkip(FromSlice(s), n)
}

func (s []T) Distinct() Lazy[T] {
	return LazyDistinct(FromSlice(s))
}

func (s []T) GroupBy[K comparable](key func(T) K) Lazy[Group[K, T]] {
	return LazyGroupBy(FromSlice(s), key)
}

func (s []T) ToList() []T {
	return slices.Clone(s)
}

func (s []T) First() T {
	v, ok := FirstValue(s)
	if !ok {
		panic("linq: sequence contains no elements")
	}
	return v
}

func (s []T) FirstOrDefault() T {
	v, ok := FirstValue(s)
	if ok {
		return v
	}
	var z T
	return z
}

func (s []T) Sum() T {
	return LazySum(FromSlice(s))
}

func (s []T) Any(pred func(T) bool) bool {
	return LazyAny(FromSlice(s), pred)
}

func (s []T) All(pred func(T) bool) bool {
	return LazyAll(FromSlice(s), pred)
}

func (s []T) Aggregate(fn func(T, T) T) T {
	return LazyAggregate(FromSlice(s), fn)
}
