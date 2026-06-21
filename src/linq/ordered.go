// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

import (
	"cmp"
	"iter"
	"slices"
)

func orderedFromLess[T any](items []T, less func(a, b T) int) Ordered[T] {
	out := slices.Clone(items)
	slices.SortFunc(out, less)
	return Ordered[T]{items: out, less: less}
}

func orderedFromSeq[T any](seq iter.Seq[T], less func(a, b T) int) Ordered[T] {
	return orderedFromLess(slices.Collect(seq), less)
}

func orderBySeq[T any, K cmp.Ordered](seq iter.Seq[T], key func(T) K) Ordered[T] {
	return orderedFromSeq(seq, func(a, b T) int { return cmp.Compare(key(a), key(b)) })
}

func orderByDescendingSeq[T any, K cmp.Ordered](seq iter.Seq[T], key func(T) K) Ordered[T] {
	return orderedFromSeq(seq, func(a, b T) int { return cmp.Compare(key(b), key(a)) })
}

func orderSeq[T cmp.Ordered](seq iter.Seq[T]) Ordered[T] {
	return orderedFromSeq(seq, func(a, b T) int { return cmp.Compare(a, b) })
}

func orderDescendingSeq[T cmp.Ordered](seq iter.Seq[T]) Ordered[T] {
	return orderedFromSeq(seq, func(a, b T) int { return cmp.Compare(b, a) })
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

// Seq returns the sorted sequence as an iterator.
func (o Ordered[T]) Seq() iter.Seq[T] {
	return slices.Values(o.items)
}

// ToList materializes the ordered sequence.
func (o Ordered[T]) ToList() []T {
	return slices.Clone(o.items)
}

func (o Ordered[T]) Zip[U, R any](other iter.Seq[U], fn func(T, U) R) iter.Seq[R] {
	return Zip(slices.Values(o.items), other, fn)
}

// Where filters the ordered sequence.
func (o Ordered[T]) Where(pred func(T) bool) iter.Seq[T] {
	return whereSeq(slices.Values(o.items), pred)
}

// Select projects the ordered sequence.
func (o Ordered[T]) Select[U any](fn func(T) U) iter.Seq[U] {
	return selectBySeq(slices.Values(o.items), fn)
}

// Take returns at most n elements.
func (o Ordered[T]) Take(n int) iter.Seq[T] {
	return takeSeq(slices.Values(o.items), n)
}

// Skip skips the first n elements.
func (o Ordered[T]) Skip(n int) iter.Seq[T] {
	return skipSeq(slices.Values(o.items), n)
}

// First returns the first element, or panics if empty.
func (o Ordered[T]) First() T {
	return firstFromSeq(slices.Values(o.items))
}

// FirstOrDefault returns the first element or the zero value.
func (o Ordered[T]) FirstOrDefault() T {
	return firstOrDefaultSeq(slices.Values(o.items))
}

func (o Ordered[T]) Aggregate(fn func(T, T) T) T {
	return aggregateSeq(slices.Values(o.items), fn)
}

func (o Ordered[T]) AggregateWithSeed[U any](seed U, fn func(U, T) U) U {
	return aggregateWithSeedSeq(slices.Values(o.items), seed, fn)
}
