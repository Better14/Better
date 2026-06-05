// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

import (
	"cmp"
	"slices"
)

// --- slice/array entry (lazy pipelines) ---
//
// Go does not allow methods on []T. Slice LINQ syntax (nums.Where(...)) is
// desugared by the compiler to these package functions when "linq" is imported.
// On linq.Lazy[T], use the receiver methods in lazy.go (e.g. l.Where(pred)).

// Where filters s and returns a lazy sequence.
func Where[T any](s []T, pred func(T) bool) Lazy[T] {
	return LazyWhere(FromSlice(s), pred)
}

// Select projects s and returns a lazy sequence.
func Select[T, U any](s []T, fn func(T) U) Lazy[U] {
	return LazySelectBy(FromSlice(s), fn)
}

// OrderBy sorts by key when the sequence is enumerated.
func OrderBy[T any, K cmp.Ordered](s []T, key func(T) K) Lazy[T] {
	return LazyOrderBy(FromSlice(s), key)
}

// OrderByDescending sorts descending by key when enumerated.
func OrderByDescending[T any, K cmp.Ordered](s []T, key func(T) K) Lazy[T] {
	return LazyOrderByDescending(FromSlice(s), key)
}

// Take returns at most n elements from s.
func Take[T any](s []T, n int) Lazy[T] {
	return LazyTake(FromSlice(s), n)
}

// Skip skips the first n elements of s.
func Skip[T any](s []T, n int) Lazy[T] {
	return LazySkip(FromSlice(s), n)
}

// Distinct returns distinct elements (comparable T).
func Distinct[T comparable](s []T) Lazy[T] {
	return LazyDistinct(FromSlice(s))
}

// GroupBy groups s by key.
func GroupBy[T any, K comparable](s []T, key func(T) K) Lazy[Group[K, T]] {
	return LazyGroupBy(FromSlice(s), key)
}

// ToList returns a copy of s.
func ToList[T any](s []T) []T {
	return slices.Clone(s)
}

// First returns the first element of s, or panics if s is empty.
func First[T any](s []T) T {
	v, ok := FirstValue(s)
	if !ok {
		panic("linq: sequence contains no elements")
	}
	return v
}

// FirstValue returns the first element of s and whether it exists.
func FirstValue[T any](s []T) (T, bool) {
	if len(s) == 0 {
		var z T
		return z, false
	}
	return s[0], true
}

// FirstOrDefault returns the first element of s, or the zero value.
func FirstOrDefault[T any](s []T) T {
	v, ok := FirstValue(s)
	if ok {
		return v
	}
	var z T
	return z
}

// Any reports whether any element satisfies pred.
func Any[T any](s []T, pred func(T) bool) bool {
	return LazyAny(FromSlice(s), pred)
}

// All reports whether all elements satisfy pred.
func All[T any](s []T, pred func(T) bool) bool {
	return LazyAll(FromSlice(s), pred)
}

// Aggregate applies fn pairwise over s (first element is the seed).
func Aggregate[T any](s []T, fn func(T, T) T) T {
	return LazyAggregate(FromSlice(s), fn)
}

// Sum returns the sum of numeric elements in s.
func Sum[U Number](s []U) U {
	return LazySum(FromSlice(s))
}

// --- lazy chain continuations (compiler maps method calls here) ---

// LazyOrderBy sorts by key when enumerated.
func LazyOrderBy[T any, K cmp.Ordered](l Lazy[T], key func(T) K) Lazy[T] {
	return FromSlice(sortOrderBy(LazyToSlice(l), key))
}

// LazyOrderByDescending sorts descending by key when enumerated.
func LazyOrderByDescending[T any, K cmp.Ordered](l Lazy[T], key func(T) K) Lazy[T] {
	return FromSlice(sortOrderByDescending(LazyToSlice(l), key))
}

// LazyDistinct returns distinct elements (comparable T).
func LazyDistinct[T comparable](l Lazy[T]) Lazy[T] {
	seen := make(map[T]struct{})
	src := l.next
	return Lazy[T]{next: func() (T, bool) {
		for {
			v, ok := src()
			if !ok {
				var z T
				return z, false
			}
			if _, dup := seen[v]; dup {
				continue
			}
			seen[v] = struct{}{}
			return v, true
		}
	}}
}

// LazyGroupBy groups by key.
func LazyGroupBy[T any, K comparable](l Lazy[T], keyFn func(T) K) Lazy[Group[K, T]] {
	groups := map[K][]T{}
	var keys []K
	for v, ok := l.next(); ok; v, ok = l.next() {
		k := keyFn(v)
		if _, seen := groups[k]; !seen {
			keys = append(keys, k)
			groups[k] = nil
		}
		groups[k] = append(groups[k], v)
	}
	idx := 0
	return Lazy[Group[K, T]]{next: func() (Group[K, T], bool) {
		if idx >= len(keys) {
			var z Group[K, T]
			return z, false
		}
		k := keys[idx]
		g := Group[K, T]{Key: k, items: groups[k]}
		idx++
		return g, true
	}}
}

// LazyAggregate applies fn pairwise (first element is the seed).
func LazyAggregate[T any](l Lazy[T], fn func(T, T) T) T {
	v, ok := l.next()
	if !ok {
		panic("linq: sequence contains no elements")
	}
	acc := v
	for {
		v, ok = l.next()
		if !ok {
			return acc
		}
		acc = fn(acc, v)
	}
}

// LazySum returns the sum of a lazy numeric sequence.
func LazySum[U Number](l Lazy[U]) U {
	var acc U
	for {
		v, ok := l.next()
		if !ok {
			return acc
		}
		acc += v
	}
}

// ToList materializes a lazy sequence.
func ToListLazy[T any](l Lazy[T]) []T {
	return LazyToSlice(l)
}

// First returns the first element of a lazy sequence, or panics if empty.
func FirstLazy[T any](l Lazy[T]) T {
	v, ok := LazyFirst(l)
	if !ok {
		panic("linq: sequence contains no elements")
	}
	return v
}

// FirstOrDefaultLazy returns the first element or the zero value.
func FirstOrDefaultLazy[T any](l Lazy[T]) T {
	v, ok := LazyFirst(l)
	if ok {
		return v
	}
	var z T
	return z
}

// sortOrderByDescending sorts a slice descending by key (eager helper).
func sortOrderByDescending[T any, K cmp.Ordered](s []T, key func(T) K) []T {
	out := slices.Clone(s)
	slices.SortFunc(out, func(a, b T) int { return cmp.Compare(key(b), key(a)) })
	return out
}
