// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

import (
	"cmp"
	"iter"
	"slices"
)

// From returns an iterator over s.
func From[T any](s []T) iter.Seq[T] {
	return slices.Values(s)
}

// AsEnumerable returns seq unchanged.
func AsEnumerable[T any](seq iter.Seq[T]) iter.Seq[T] {
	return seq
}

// Where filters elements.
func Where[T any](seq iter.Seq[T], pred func(T) bool) iter.Seq[T] {
	return whereSeq(seq, pred)
}

// Select projects each element to type U.
func Select[T, U any](seq iter.Seq[T], fn func(T) U) iter.Seq[U] {
	return selectBySeq(seq, fn)
}

// SelectBy is an alias for Select.
func SelectBy[T, U any](seq iter.Seq[T], fn func(T) U) iter.Seq[U] {
	return selectBySeq(seq, fn)
}

// SelectMany flattens each element to a sequence.
func SelectMany[T, U any](seq iter.Seq[T], fn func(T) iter.Seq[U]) iter.Seq[U] {
	return selectManySeq(seq, fn)
}

// OrderBy sorts by key when the sequence is enumerated.
func OrderBy[T any, K cmp.Ordered](seq iter.Seq[T], key func(T) K) Ordered[T] {
	return orderBySeq(seq, key)
}

// OrderByDescending sorts descending by key when enumerated.
func OrderByDescending[T any, K cmp.Ordered](seq iter.Seq[T], key func(T) K) Ordered[T] {
	return orderByDescendingSeq(seq, key)
}

// Order sorts ascending by element value.
func Order[T cmp.Ordered](seq iter.Seq[T]) Ordered[T] {
	return orderSeq(seq)
}

// OrderDescending sorts descending by element value.
func OrderDescending[T cmp.Ordered](seq iter.Seq[T]) Ordered[T] {
	return orderDescendingSeq(seq)
}

// Take returns at most n elements.
func Take[T any](seq iter.Seq[T], n int) iter.Seq[T] {
	return takeSeq(seq, n)
}

// Skip skips the first n elements.
func Skip[T any](seq iter.Seq[T], n int) iter.Seq[T] {
	return skipSeq(seq, n)
}

// TakeWhile takes elements while pred is true.
func TakeWhile[T any](seq iter.Seq[T], pred func(T) bool) iter.Seq[T] {
	return takeWhileSeq(seq, pred)
}

// SkipWhile skips elements while pred is true.
func SkipWhile[T any](seq iter.Seq[T], pred func(T) bool) iter.Seq[T] {
	return skipWhileSeq(seq, pred)
}

// TakeLast takes the last count elements.
func TakeLast[T any](seq iter.Seq[T], count int) iter.Seq[T] {
	return takeLastSeq(seq, count)
}

// SkipLast skips the last count elements.
func SkipLast[T any](seq iter.Seq[T], count int) iter.Seq[T] {
	return skipLastSeq(seq, count)
}

// Append appends element to the end of the sequence.
func Append[T any](seq iter.Seq[T], element T) iter.Seq[T] {
	return appendSeq(seq, element)
}

// Prepend prepends element to the start of the sequence.
func Prepend[T any](seq iter.Seq[T], element T) iter.Seq[T] {
	return prependSeq(seq, element)
}

// Concat concatenates two sequences.
func Concat[T any](seq iter.Seq[T], other iter.Seq[T]) iter.Seq[T] {
	return concatSeq(seq, other)
}

// Zip pairs elements until either sequence is exhausted.
func Zip[T, U, R any](seq iter.Seq[T], other iter.Seq[U], fn func(T, U) R) iter.Seq[R] {
	return zipSeq(seq, other, fn)
}

// Reverse returns elements in reverse order.
func Reverse[T any](seq iter.Seq[T]) iter.Seq[T] {
	return reverseSeq(seq)
}

// Chunk splits the sequence into chunks of size.
func Chunk[T any](seq iter.Seq[T], size int) iter.Seq[[]T] {
	return chunkSeq(seq, size)
}

// DefaultIfEmpty returns the sequence, or a single defaultValue if empty.
func DefaultIfEmpty[T any](seq iter.Seq[T], defaultValue T) iter.Seq[T] {
	return defaultIfEmptySeq(seq, defaultValue)
}

// Index pairs each element with its zero-based index.
func Index[T any](seq iter.Seq[T]) iter.Seq[Indexed[T]] {
	return indexSeq(seq)
}

// Shuffle returns elements in random order.
func Shuffle[T any](seq iter.Seq[T]) iter.Seq[T] {
	return shuffleSeq(seq)
}

// Distinct returns distinct elements (comparable T).
func Distinct[T comparable](seq iter.Seq[T]) iter.Seq[T] {
	return distinctSeq(seq)
}

// DistinctBy returns distinct elements by key.
func DistinctBy[T any, K comparable](seq iter.Seq[T], keyFn func(T) K) iter.Seq[T] {
	return distinctBySeq(seq, keyFn)
}

// Except returns elements not present in other.
func Except[T comparable](seq iter.Seq[T], other iter.Seq[T]) iter.Seq[T] {
	return exceptSeq(seq, other)
}

// ExceptBy returns elements whose key is not in other.
func ExceptBy[T any, K comparable](seq iter.Seq[T], other iter.Seq[T], keyFn func(T) K) iter.Seq[T] {
	return exceptBySeq(seq, other, keyFn)
}

// Intersect returns elements present in both sequences.
func Intersect[T comparable](seq iter.Seq[T], other iter.Seq[T]) iter.Seq[T] {
	return intersectSeq(seq, other)
}

// IntersectBy returns elements whose key appears in both sequences.
func IntersectBy[T any, K comparable](seq iter.Seq[T], other iter.Seq[T], keyFn func(T) K) iter.Seq[T] {
	return intersectBySeq(seq, other, keyFn)
}

// Union returns distinct elements from both sequences.
func Union[T comparable](seq iter.Seq[T], other iter.Seq[T]) iter.Seq[T] {
	return unionSeq(seq, other)
}

// UnionBy returns distinct elements from both sequences by key.
func UnionBy[T any, K comparable](seq iter.Seq[T], other iter.Seq[T], keyFn func(T) K) iter.Seq[T] {
	return unionBySeq(seq, other, keyFn)
}

// Contains reports whether the sequence contains value.
func Contains[T comparable](seq iter.Seq[T], value T) bool {
	return containsSeq(seq, value)
}

// SequenceEqual reports whether two sequences are equal in order.
func SequenceEqual[T comparable](seq iter.Seq[T], other iter.Seq[T]) bool {
	return sequenceEqualSeq(seq, other)
}

// ToList materializes the sequence.
func ToList[T any](seq iter.Seq[T]) []T {
	return toListSeq(seq)
}

// ToArray materializes the sequence to a slice.
func ToArray[T any](seq iter.Seq[T]) []T {
	return toListSeq(seq)
}

// ToHashSet materializes comparable elements into a set.
func ToHashSet[T comparable](seq iter.Seq[T]) HashSet[T] {
	return toHashSetSeq(seq)
}

// ToDictionary builds a map; panics on duplicate keys.
func ToDictionary[T any, K comparable, V any](seq iter.Seq[T], keyFn func(T) K, valueFn func(T) V) map[K]V {
	return toDictionarySeq(seq, keyFn, valueFn)
}

// ToLookup builds a Lookup from key and value selectors.
func ToLookup[T any, K comparable, V any](seq iter.Seq[T], keyFn func(T) K, valueFn func(T) V) Lookup[K, V] {
	return toLookupSeq(seq, keyFn, valueFn)
}

// First returns the first element, or panics if empty.
func First[T any](seq iter.Seq[T]) T {
	return firstFromSeq(seq)
}

// FirstOrDefault returns the first element or the zero value.
func FirstOrDefault[T any](seq iter.Seq[T]) T {
	return firstOrDefaultSeq(seq)
}

// Last returns the last element, or panics if empty.
func Last[T any](seq iter.Seq[T]) T {
	return lastFromSeq(seq)
}

// LastOrDefault returns the last element or the zero value.
func LastOrDefault[T any](seq iter.Seq[T]) T {
	return lastOrDefaultSeq(seq)
}

// Single returns the only element, or panics.
func Single[T any](seq iter.Seq[T]) T {
	return singleSeq(seq)
}

// SingleOrDefault returns the only element or defaultValue.
func SingleOrDefault[T any](seq iter.Seq[T], defaultValue T) T {
	return singleOrDefaultSeq(seq, defaultValue)
}

// ElementAt returns the element at index, or panics.
func ElementAt[T any](seq iter.Seq[T], index int) T {
	return elementAtSeq(seq, index)
}

// ElementAtOrDefault returns the element at index or defaultValue.
func ElementAtOrDefault[T any](seq iter.Seq[T], index int, defaultValue T) T {
	return elementAtOrDefaultSeq(seq, index, defaultValue)
}

// Any reports whether any element satisfies pred.
func Any[T any](seq iter.Seq[T], pred func(T) bool) bool {
	return anySeq(seq, pred)
}

// All reports whether all elements satisfy pred.
func All[T any](seq iter.Seq[T], pred func(T) bool) bool {
	return allSeq(seq, pred)
}

// Count returns the number of elements.
func Count[T any](seq iter.Seq[T]) int {
	return countSeq(seq)
}

// LongCount returns the number of elements as int64.
func LongCount[T any](seq iter.Seq[T]) int64 {
	return longCountSeq(seq)
}

// CountBy counts elements by key.
func CountBy[T any, K comparable](seq iter.Seq[T], keyFn func(T) K) iter.Seq[KeyValue[K, int]] {
	return countBySeq(seq, keyFn)
}

// Aggregate applies fn pairwise (first element is the seed).
func Aggregate[T any](seq iter.Seq[T], fn func(T, T) T) T {
	return aggregateSeq(seq, fn)
}

// AggregateWithSeed folds starting with seed.
func AggregateWithSeed[T any, U any](seq iter.Seq[T], seed U, fn func(U, T) U) U {
	return aggregateWithSeedSeq(seq, seed, fn)
}

// AggregateBy groups by key and aggregates each group.
func AggregateBy[T any, K comparable, A any, R any](seq iter.Seq[T], keyFn func(T) K, seed A, fn func(A, T) A, resultFn func(A) R) iter.Seq[KeyValue[K, R]] {
	return aggregateBySeq(seq, keyFn, seed, fn, resultFn)
}

// Sum returns the sum of a numeric sequence.
func Sum[U Number](seq iter.Seq[U]) U {
	return sumSeq(seq)
}

// Average returns the arithmetic mean as float64.
func Average[U Number](seq iter.Seq[U]) float64 {
	return averageSeq(seq)
}

// Max returns the maximum element.
func Max[T cmp.Ordered](seq iter.Seq[T]) T {
	return maxSeq(seq)
}

// MaxBy returns the element with the maximum key.
func MaxBy[T any, K cmp.Ordered](seq iter.Seq[T], key func(T) K) T {
	return maxBySeq(seq, key)
}

// Min returns the minimum element.
func Min[T cmp.Ordered](seq iter.Seq[T]) T {
	return minSeq(seq)
}

// MinBy returns the element with the minimum key.
func MinBy[T any, K cmp.Ordered](seq iter.Seq[T], key func(T) K) T {
	return minBySeq(seq, key)
}

// GroupBy groups by key.
func GroupBy[T any, K comparable](seq iter.Seq[T], keyFn func(T) K) iter.Seq[Group[K, T]] {
	return groupBySeq(seq, keyFn)
}

// Join inner-joins with inner.
func Join[T, U, K comparable, R any](outer iter.Seq[T], inner iter.Seq[U], outerKey func(T) K, innerKey func(U) K, resultFn func(T, U) R) iter.Seq[R] {
	return joinSeq(outer, inner, outerKey, innerKey, resultFn)
}

// GroupJoin groups inner and joins with outer.
func GroupJoin[T, U, K comparable, R any](outer iter.Seq[T], inner iter.Seq[U], outerKey func(T) K, innerKey func(U) K, resultFn func(T, iter.Seq[U]) R) iter.Seq[R] {
	return groupJoinSeq(outer, inner, outerKey, innerKey, resultFn)
}

// LeftJoin left-joins with inner.
func LeftJoin[T, U, K comparable, R any](outer iter.Seq[T], inner iter.Seq[U], outerKey func(T) K, innerKey func(U) K, resultFn func(T, U) R, defaultInner U) iter.Seq[R] {
	return leftJoinSeq(outer, inner, outerKey, innerKey, resultFn, defaultInner)
}

// RightJoin right-joins with inner.
func RightJoin[T, U, K comparable, R any](outer iter.Seq[T], inner iter.Seq[U], outerKey func(T) K, innerKey func(U) K, resultFn func(T, U) R, defaultOuter T) iter.Seq[R] {
	return rightJoinSeq(outer, inner, outerKey, innerKey, resultFn, defaultOuter)
}

// FullJoin full-outer-joins with inner.
// Matched keys emit all pairings; unmatched outer rows use defaultInner;
// unmatched inner rows use defaultOuter.
func FullJoin[T, U, K comparable, R any](outer iter.Seq[T], inner iter.Seq[U], outerKey func(T) K, innerKey func(U) K, resultFn func(T, U) R, defaultOuter T, defaultInner U) iter.Seq[R] {
	return fullJoinSeq(outer, inner, outerKey, innerKey, resultFn, defaultOuter, defaultInner)
}

// TryGetNonEnumeratedCount reports a known length without enumerating.
func TryGetNonEnumeratedCount[T any](seq iter.Seq[T]) (int, bool) {
	return tryGetNonEnumeratedCountSeq(seq)
}

// Cast casts each element to U (for iter.Seq[any]).
func Cast[U any](seq iter.Seq[any]) iter.Seq[U] {
	return castSeq[U](seq)
}

// OfType filters elements assignable to U (for iter.Seq[any]).
func OfType[U any](seq iter.Seq[any]) iter.Seq[U] {
	return ofTypeSeq[U](seq)
}
