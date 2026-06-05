// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

import (
	"cmp"
	"slices"
)

// Slice LINQ syntax (nums.Where(...)) uses extension methods in slice_ext.go
// when import "linq" is present. Lazy[T] uses receiver methods in lazy.go.

// FirstValue returns the first element of s and whether it exists.
func FirstValue[T any](s []T) (T, bool) {
	if len(s) == 0 {
		var z T
		return z, false
	}
	return s[0], true
}

// --- lazy chain continuations (package functions for direct use) ---

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
