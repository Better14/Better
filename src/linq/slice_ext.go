// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

import (
	"cmp"
	"slices"
)

// Extension methods on []T provide LINQ syntax when import "linq" is present.
// Methods must restate the receiver type parameter T (e.g. Where[T any]).

func (s []T) Where[T any](pred func(T) bool) Lazy[T] {
	return LazyWhere(FromSlice(s), pred)
}

func (s []T) Select[T, U any](fn func(T) U) Lazy[U] {
	return LazySelectBy(FromSlice(s), fn)
}

func (s []T) OrderBy[T, K cmp.Ordered](key func(T) K) Lazy[T] {
	return LazyOrderBy(FromSlice(s), key)
}

func (s []T) OrderByDescending[T, K cmp.Ordered](key func(T) K) Lazy[T] {
	return LazyOrderByDescending(FromSlice(s), key)
}

func (s []T) Take[T any](n int) Lazy[T] {
	return LazyTake(FromSlice(s), n)
}

func (s []T) Skip[T any](n int) Lazy[T] {
	return LazySkip(FromSlice(s), n)
}

func (s []T) ToList[T any]() []T {
	return slices.Clone(s)
}

func (s []T) First[T any]() T {
	v, ok := FirstValue(s)
	if !ok {
		panic("linq: sequence contains no elements")
	}
	return v
}

func (s []T) FirstOrDefault[T any]() T {
	v, ok := FirstValue(s)
	if ok {
		return v
	}
	var z T
	return z
}

func (s []T) Any[T any](pred func(T) bool) bool {
	return LazyAny(FromSlice(s), pred)
}

func (s []T) All[T any](pred func(T) bool) bool {
	return LazyAll(FromSlice(s), pred)
}

func (s []T) Aggregate[T any](fn func(T, T) T) T {
	return LazyAggregate(FromSlice(s), fn)
}

func (s []T) Distinct[T comparable]() Lazy[T] {
	return LazyDistinct(FromSlice(s))
}

func (s []T) GroupBy[T, K comparable](keyFn func(T) K) Lazy[Group[K, T]] {
	return LazyGroupBy(FromSlice(s), keyFn)
}

func (s []T) Sum[T Number]() T {
	return LazySum(FromSlice(s))
}
