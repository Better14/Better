// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

import (
	"cmp"
	"slices"
)

func orderedFromLess[T any](items []T, less func(a, b T) int) Ordered[T] {
	out := slices.Clone(items)
	slices.SortFunc(out, less)
	return Ordered[T]{items: out, less: less}
}

func orderedFromLazy[T any](l Lazy[T], less func(a, b T) int) Ordered[T] {
	return orderedFromLess(lazyToSlice(l), less)
}

func lazyOrderBy[T any, K cmp.Ordered](l Lazy[T], key func(T) K) Ordered[T] {
	return orderedFromLazy(l, func(a, b T) int { return cmp.Compare(key(a), key(b)) })
}

func lazyOrderByDescending[T any, K cmp.Ordered](l Lazy[T], key func(T) K) Ordered[T] {
	return orderedFromLazy(l, func(a, b T) int { return cmp.Compare(key(b), key(a)) })
}

func lazyOrder[T cmp.Ordered](l Lazy[T]) Ordered[T] {
	return orderedFromLazy(l, func(a, b T) int { return cmp.Compare(a, b) })
}

func lazyOrderDescending[T cmp.Ordered](l Lazy[T]) Ordered[T] {
	return orderedFromLazy(l, func(a, b T) int { return cmp.Compare(b, a) })
}

func (o Ordered[T]) thenBy(keyLess func(a, b T) int) Ordered[T] {
	less := o.less
	combined := func(a, b T) int {
		if c := less(a, b); c != 0 {
			return c
		}
		return keyLess(a, b)
	}
	out := slices.Clone(o.items)
	slices.SortFunc(out, combined)
	return Ordered[T]{items: out, less: combined}
}

// ThenBy sorts with a secondary key ascending.
func (o Ordered[T]) ThenBy[K cmp.Ordered](key func(T) K) Ordered[T] {
	return o.thenBy(func(a, b T) int { return cmp.Compare(key(a), key(b)) })
}

// ThenByDescending sorts with a secondary key descending.
func (o Ordered[T]) ThenByDescending[K cmp.Ordered](key func(T) K) Ordered[T] {
	return o.thenBy(func(a, b T) int { return cmp.Compare(key(b), key(a)) })
}

// Lazy returns the sorted sequence as a lazy iterator.
func (o Ordered[T]) Lazy() Lazy[T] {
	return From(o.items)
}

// ToList materializes the ordered sequence.
func (o Ordered[T]) ToList() []T {
	return slices.Clone(o.items)
}

// Where filters the ordered sequence.
func (o Ordered[T]) Where(pred func(T) bool) Lazy[T] {
	return lazyWhere(From(o.items), pred)
}

// Select projects the ordered sequence.
func (o Ordered[T]) Select[U any](fn func(T) U) Lazy[U] {
	return lazySelectBy(From(o.items), fn)
}

// Take returns at most n elements.
func (o Ordered[T]) Take(n int) Lazy[T] {
	return lazyTake(From(o.items), n)
}

// Skip skips the first n elements.
func (o Ordered[T]) Skip(n int) Lazy[T] {
	return lazySkip(From(o.items), n)
}

// First returns the first element, or panics if empty.
func (o Ordered[T]) First() T {
	return firstLazy(From(o.items))
}

// FirstOrDefault returns the first element or the zero value.
func (o Ordered[T]) FirstOrDefault() T {
	return firstOrDefaultLazy(From(o.items))
}
