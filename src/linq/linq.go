// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

import "cmp"

// Lazy is a deferred sequence: chain package functions until a terminal call.
type Lazy[T any] struct {
	next        func() (T, bool)
	knownLen    int
	hasKnownLen bool
}

// From returns a lazy iterator over s.
func From[T any](s []T) Lazy[T] {
	i := 0
	n := len(s)
	return Lazy[T]{
		knownLen:    n,
		hasKnownLen: true,
		next: func() (T, bool) {
			if i >= n {
				var z T
				return z, false
			}
			v := s[i]
			i++
			return v, true
		},
	}
}

// Where filters l and returns a lazy sequence.
func (l Lazy[T]) Where(pred func(T) bool) Lazy[T] {
	return lazyWhere(l, pred)
}

// Select projects each element to type U.
func (l Lazy[T]) Select[U any](fn func(T) U) Lazy[U] {
	return lazySelectBy(l, fn)
}

// SelectMany flattens each element to a lazy sequence.
func (l Lazy[T]) SelectMany[U any](fn func(T) Lazy[U]) Lazy[U] {
	return lazySelectMany(l, fn)
}

// OrderBy sorts by key when the sequence is enumerated.
func (l Lazy[T]) OrderBy[K cmp.Ordered](key func(T) K) Ordered[T] {
	return lazyOrderBy(l, key)
}

// OrderByDescending sorts descending by key when enumerated.
func (l Lazy[T]) OrderByDescending[K cmp.Ordered](key func(T) K) Ordered[T] {
	return lazyOrderByDescending(l, key)
}

// Order sorts ascending by element value.
func (l Lazy[T]) Order[T cmp.Ordered]() Ordered[T] {
	return lazyOrder(l)
}

// OrderDescending sorts descending by element value.
func (l Lazy[T]) OrderDescending[T cmp.Ordered]() Ordered[T] {
	return lazyOrderDescending(l)
}

// Take returns at most n elements.
func (l Lazy[T]) Take(n int) Lazy[T] {
	return lazyTake(l, n)
}

// Skip skips the first n elements.
func (l Lazy[T]) Skip(n int) Lazy[T] {
	return lazySkip(l, n)
}

// TakeWhile takes elements while pred is true.
func (l Lazy[T]) TakeWhile(pred func(T) bool) Lazy[T] {
	return lazyTakeWhile(l, pred)
}

// SkipWhile skips elements while pred is true.
func (l Lazy[T]) SkipWhile(pred func(T) bool) Lazy[T] {
	return lazySkipWhile(l, pred)
}

// TakeLast takes the last count elements.
func (l Lazy[T]) TakeLast(count int) Lazy[T] {
	return lazyTakeLast(l, count)
}

// SkipLast skips the last count elements.
func (l Lazy[T]) SkipLast(count int) Lazy[T] {
	return lazySkipLast(l, count)
}

// Append appends element to the end of the sequence.
func (l Lazy[T]) Append(element T) Lazy[T] {
	return lazyAppend(l, element)
}

// Prepend prepends element to the start of the sequence.
func (l Lazy[T]) Prepend(element T) Lazy[T] {
	return lazyPrepend(l, element)
}

// Concat concatenates l with other.
func (l Lazy[T]) Concat(other Lazy[T]) Lazy[T] {
	return lazyConcat(l, other)
}

// Zip pairs elements with other until either is exhausted.
func (l Lazy[T]) Zip[U, R any](other Lazy[U], fn func(T, U) R) Lazy[R] {
	return lazyZip(l, other, fn)
}

// Reverse returns elements in reverse order.
func (l Lazy[T]) Reverse() Lazy[T] {
	return lazyReverse(l)
}

// Chunk splits the sequence into chunks of size.
func (l Lazy[T]) Chunk(size int) Lazy[[]T] {
	return lazyChunk(l, size)
}

// DefaultIfEmpty returns l, or a single defaultValue if l is empty.
func (l Lazy[T]) DefaultIfEmpty(defaultValue T) Lazy[T] {
	return lazyDefaultIfEmpty(l, defaultValue)
}

// Index pairs each element with its zero-based index.
func (l Lazy[T]) Index() Lazy[Indexed[T]] {
	return lazyIndex(l)
}

// Shuffle returns elements in random order.
func (l Lazy[T]) Shuffle() Lazy[T] {
	return lazyShuffle(l)
}

// Distinct returns distinct elements (comparable T).
func (l Lazy[T]) Distinct[T comparable]() Lazy[T] {
	return lazyDistinct(l)
}

// DistinctBy returns distinct elements by key.
func (l Lazy[T]) DistinctBy[K comparable](keyFn func(T) K) Lazy[T] {
	return lazyDistinctBy(l, keyFn)
}

// Except returns elements not present in other.
func (l Lazy[T]) Except[T comparable](other Lazy[T]) Lazy[T] {
	return lazyExcept(l, other)
}

// ExceptBy returns elements whose key is not in other.
func (l Lazy[T]) ExceptBy[K comparable](other Lazy[T], keyFn func(T) K) Lazy[T] {
	return lazyExceptBy(l, other, keyFn)
}

// Intersect returns elements present in both sequences.
func (l Lazy[T]) Intersect[T comparable](other Lazy[T]) Lazy[T] {
	return lazyIntersect(l, other)
}

// IntersectBy returns elements whose key appears in both sequences.
func (l Lazy[T]) IntersectBy[K comparable](other Lazy[T], keyFn func(T) K) Lazy[T] {
	return lazyIntersectBy(l, other, keyFn)
}

// Union returns distinct elements from l and other.
func (l Lazy[T]) Union[T comparable](other Lazy[T]) Lazy[T] {
	return lazyUnion(l, other)
}

// UnionBy returns distinct elements from l and other by key.
func (l Lazy[T]) UnionBy[K comparable](other Lazy[T], keyFn func(T) K) Lazy[T] {
	return lazyUnionBy(l, other, keyFn)
}

// Contains reports whether l contains value.
func (l Lazy[T]) Contains[T comparable](value T) bool {
	return lazyContains(l, value)
}

// SequenceEqual reports whether l and other are equal in order.
func (l Lazy[T]) SequenceEqual[T comparable](other Lazy[T]) bool {
	return lazySequenceEqual(l, other)
}

// ToList materializes the sequence.
func (l Lazy[T]) ToList() []T {
	return toListLazy(l)
}

// ToArray materializes the sequence to a slice.
func (l Lazy[T]) ToArray() []T {
	return lazyToArray(l)
}

// ToHashSet materializes comparable elements into a set.
func (l Lazy[T]) ToHashSet[T comparable]() HashSet[T] {
	return lazyToHashSet(l)
}

// ToDictionary builds a map; panics on duplicate keys.
func (l Lazy[T]) ToDictionary[K comparable, V any](keyFn func(T) K, valueFn func(T) V) map[K]V {
	return lazyToDictionary(l, keyFn, valueFn)
}

// ToLookup builds a Lookup from key and value selectors.
func (l Lazy[T]) ToLookup[K comparable, V any](keyFn func(T) K, valueFn func(T) V) Lookup[K, V] {
	return lazyToLookup(l, keyFn, valueFn)
}

// First returns the first element, or panics if empty.
func (l Lazy[T]) First() T {
	return firstLazy(l)
}

// FirstOrDefault returns the first element or the zero value.
func (l Lazy[T]) FirstOrDefault() T {
	return firstOrDefaultLazy(l)
}

// Last returns the last element, or panics if empty.
func (l Lazy[T]) Last() T {
	return lazyLast(l)
}

// LastOrDefault returns the last element or the zero value.
func (l Lazy[T]) LastOrDefault() T {
	return lazyLastOrDefault(l)
}

// Single returns the only element, or panics.
func (l Lazy[T]) Single() T {
	return lazySingle(l)
}

// SingleOrDefault returns the only element or defaultValue.
func (l Lazy[T]) SingleOrDefault(defaultValue T) T {
	return lazySingleOrDefault(l, defaultValue)
}

// ElementAt returns the element at index, or panics.
func (l Lazy[T]) ElementAt(index int) T {
	return lazyElementAt(l, index)
}

// ElementAtOrDefault returns the element at index or defaultValue.
func (l Lazy[T]) ElementAtOrDefault(index int, defaultValue T) T {
	return lazyElementAtOrDefault(l, index, defaultValue)
}

// Any reports whether any element satisfies pred.
func (l Lazy[T]) Any(pred func(T) bool) bool {
	return lazyAny(l, pred)
}

// All reports whether all elements satisfy pred.
func (l Lazy[T]) All(pred func(T) bool) bool {
	return lazyAll(l, pred)
}

// Count returns the number of elements.
func (l Lazy[T]) Count() int {
	return lazyCount(l)
}

// LongCount returns the number of elements as int64.
func (l Lazy[T]) LongCount() int64 {
	return lazyLongCount(l)
}

// CountBy counts elements by key.
func (l Lazy[T]) CountBy[K comparable](keyFn func(T) K) Lazy[KeyValue[K, int]] {
	return lazyCountBy(l, keyFn)
}

// Aggregate applies fn pairwise (first element is the seed).
func (l Lazy[T]) Aggregate(fn func(T, T) T) T {
	return lazyAggregate(l, fn)
}

// AggregateWithSeed folds starting with seed.
func (l Lazy[T]) AggregateWithSeed[U any](seed U, fn func(U, T) U) U {
	return lazyAggregateWithSeed(l, seed, fn)
}

// AggregateBy groups by key and aggregates each group.
func (l Lazy[T]) AggregateBy[K comparable, A any, R any](keyFn func(T) K, seed A, fn func(A, T) A, resultFn func(A) R) Lazy[KeyValue[K, R]] {
	return lazyAggregateBy(l, keyFn, seed, fn, resultFn)
}

// Sum returns the sum of a numeric sequence.
func (l Lazy[T]) Sum[T Number]() T {
	return lazySum(l)
}

// Average returns the arithmetic mean as float64.
func (l Lazy[T]) Average[T Number]() float64 {
	return lazyAverage(l)
}

// Max returns the maximum element.
func (l Lazy[T]) Max[T cmp.Ordered]() T {
	return lazyMax(l)
}

// MaxBy returns the element with the maximum key.
func (l Lazy[T]) MaxBy[K cmp.Ordered](key func(T) K) T {
	return lazyMaxBy(l, key)
}

// Min returns the minimum element.
func (l Lazy[T]) Min[T cmp.Ordered]() T {
	return lazyMin(l)
}

// MinBy returns the element with the minimum key.
func (l Lazy[T]) MinBy[K cmp.Ordered](key func(T) K) T {
	return lazyMinBy(l, key)
}

// GroupBy groups by key.
func (l Lazy[T]) GroupBy[K comparable](keyFn func(T) K) Lazy[Group[K, T]] {
	return lazyGroupBy(l, keyFn)
}

// Join inner-joins with inner.
func (l Lazy[T]) Join[U, K comparable, R any](inner Lazy[U], outerKey func(T) K, innerKey func(U) K, resultFn func(T, U) R) Lazy[R] {
	return lazyJoin(l, inner, outerKey, innerKey, resultFn)
}

// GroupJoin groups inner and joins with outer.
func (l Lazy[T]) GroupJoin[U, K comparable, R any](inner Lazy[U], outerKey func(T) K, innerKey func(U) K, resultFn func(T, Lazy[U]) R) Lazy[R] {
	return lazyGroupJoin(l, inner, outerKey, innerKey, resultFn)
}

// LeftJoin left-joins with inner.
func (l Lazy[T]) LeftJoin[U, K comparable, R any](inner Lazy[U], outerKey func(T) K, innerKey func(U) K, resultFn func(T, U) R, defaultInner U) Lazy[R] {
	return lazyLeftJoin(l, inner, outerKey, innerKey, resultFn, defaultInner)
}

// RightJoin right-joins with inner.
func (l Lazy[T]) RightJoin[U, K comparable, R any](inner Lazy[U], outerKey func(T) K, innerKey func(U) K, resultFn func(T, U) R, defaultOuter T) Lazy[R] {
	return lazyRightJoin(l, inner, outerKey, innerKey, resultFn, defaultOuter)
}

// AsEnumerable returns l unchanged.
func (l Lazy[T]) AsEnumerable() Lazy[T] {
	return asEnumerable(l)
}

// TryGetNonEnumeratedCount reports a known length without enumerating.
func (l Lazy[T]) TryGetNonEnumeratedCount() (int, bool) {
	return tryGetNonEnumeratedCount(l)
}

// Cast casts each element to U (for Lazy[any]).
func (l Lazy[any]) Cast[U any]() Lazy[U] {
	return lazyCast[U](l)
}

// OfType filters elements assignable to U (for Lazy[any]).
func (l Lazy[any]) OfType[U any]() Lazy[U] {
	return lazyOfType[U](l)
}

func lazyWhere[T any](l Lazy[T], pred func(T) bool) Lazy[T] {
	src := l.next
	return Lazy[T]{next: func() (T, bool) {
		for {
			v, ok := src()
			if !ok {
				var z T
				return z, false
			}
			if pred(v) {
				return v, true
			}
		}
	}}
}

func lazySelect[T any](l Lazy[T], fn func(T) T) Lazy[T] {
	src := l.next
	return Lazy[T]{next: func() (T, bool) {
		v, ok := src()
		if !ok {
			var z T
			return z, false
		}
		return fn(v), true
	}}
}

func lazySelectBy[T, U any](l Lazy[T], fn func(T) U) Lazy[U] {
	src := l.next
	return Lazy[U]{next: func() (U, bool) {
		v, ok := src()
		if !ok {
			var z U
			return z, false
		}
		return fn(v), true
	}}
}

func lazySkip[T any](l Lazy[T], n int) Lazy[T] {
	src := l.next
	return Lazy[T]{next: func() (T, bool) {
		for i := 0; i < n; i++ {
			if _, ok := src(); !ok {
				var z T
				return z, false
			}
		}
		return src()
	}}
}

func lazyTake[T any](l Lazy[T], n int) Lazy[T] {
	src := l.next
	left := n
	return Lazy[T]{next: func() (T, bool) {
		if left <= 0 {
			var z T
			return z, false
		}
		left--
		return src()
	}}
}

func lazyToSlice[T any](l Lazy[T]) []T {
	var out []T
	for {
		v, ok := l.next()
		if !ok {
			break
		}
		out = append(out, v)
	}
	return out
}

func lazyFirst[T any](l Lazy[T]) (T, bool) {
	return l.next()
}

func lazyFirstOrDefault[T any](l Lazy[T], def T) T {
	v, ok := l.next()
	if ok {
		return v
	}
	return def
}

func lazyAny[T any](l Lazy[T], pred func(T) bool) bool {
	for {
		v, ok := l.next()
		if !ok {
			return false
		}
		if pred(v) {
			return true
		}
	}
}

func lazyAll[T any](l Lazy[T], pred func(T) bool) bool {
	for {
		v, ok := l.next()
		if !ok {
			return true
		}
		if !pred(v) {
			return false
		}
	}
}

func lazyCount[T any](l Lazy[T]) int {
	n := 0
	for {
		if _, ok := l.next(); !ok {
			return n
		}
		n++
	}
}
