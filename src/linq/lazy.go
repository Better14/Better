// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

import (
	"cmp"
	"slices"
)

// --- lazy chain implementations ---

// lazyDistinct returns distinct elements (comparable T).
func lazyDistinct[T comparable](l Lazy[T]) Lazy[T] {
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

// lazyGroupBy groups by key.
func lazyGroupBy[T any, K comparable](l Lazy[T], keyFn func(T) K) Lazy[Group[K, T]] {
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

// lazyAggregate applies fn pairwise (first element is the seed).
func lazyAggregate[T any](l Lazy[T], fn func(T, T) T) T {
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

// lazySum returns the sum of a lazy numeric sequence.
func lazySum[U Number](l Lazy[U]) U {
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
func toListLazy[T any](l Lazy[T]) []T {
	return lazyToSlice(l)
}

// First returns the first element of a lazy sequence, or panics if empty.
func firstLazy[T any](l Lazy[T]) T {
	v, ok := lazyFirst(l)
	if !ok {
		panic("linq: sequence contains no elements")
	}
	return v
}

// firstOrDefaultLazy returns the first element or the zero value.
func firstOrDefaultLazy[T any](l Lazy[T]) T {
	v, ok := lazyFirst(l)
	if ok {
		return v
	}
	var z T
	return z
}

// FirstValue returns the first element of s and whether it exists.
func FirstValue[T any](s []T) (T, bool) {
	if len(s) == 0 {
		var z T
		return z, false
	}
	return s[0], true
}

// sortOrderBy sorts s by key (eager materialization helper).
func sortOrderBy[T any, K cmp.Ordered](s []T, key func(T) K) []T {
	out := slices.Clone(s)
	slices.SortFunc(out, func(a, b T) int { return cmp.Compare(key(a), key(b)) })
	return out
}

// sortOrderByDescending sorts a slice descending by key (eager helper).
func sortOrderByDescending[T any, K cmp.Ordered](s []T, key func(T) K) []T {
	out := slices.Clone(s)
	slices.SortFunc(out, func(a, b T) int { return cmp.Compare(key(b), key(a)) })
	return out
}

// DistinctComparable returns distinct elements of s (eager).
func DistinctComparable[T comparable](s []T) []T {
	seen := make(map[T]struct{})
	var out []T
	for _, v := range s {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
