// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

import (
	"cmp"
	"iter"
)

// Extension methods on iter.Seq[T] provide LINQ syntax when import "linq" is present.
// Methods delegate to package-level overloaded functions.

func (seq iter.Seq[T]) AsSeq() iter.Seq[T] {
	return AsSeq(seq)
}

func (seq iter.Seq[T]) Where(pred func(T) bool) iter.Seq[T] {
	return whereSeq(seq, pred)
}

func (seq iter.Seq[T]) Where(pred func(T, int) bool) iter.Seq[T] {
	return whereIndexedSeq(seq, pred)
}

func (seq iter.Seq[T]) Select[U any](fn func(T) U) iter.Seq[U] {
	return selectBySeq(seq, fn)
}

func (seq iter.Seq[T]) Select[U any](fn func(T, int) U) iter.Seq[U] {
	return selectIndexedSeq(seq, fn)
}

func (seq iter.Seq[T]) SelectMany[U any](fn func(T) iter.Seq[U]) iter.Seq[U] {
	return selectManySeq(seq, fn)
}

func (seq iter.Seq[T]) SelectMany[U any](fn func(T, int) iter.Seq[U]) iter.Seq[U] {
	return selectManyIndexedSeq(seq, fn)
}

func (seq iter.Seq[T]) SelectMany[C, U any](collectionFn func(T) iter.Seq[C], resultFn func(T, C) U) iter.Seq[U] {
	return selectManyResultSeq(seq, collectionFn, resultFn)
}

func (seq iter.Seq[T]) SelectMany[C, U any](collectionFn func(T) []C, resultFn func(T, C) U) iter.Seq[U] {
	return selectManySliceResultSeq(seq, collectionFn, resultFn)
}

func (seq iter.Seq[T]) SelectMany[C, U any](collectionFn func(T, int) iter.Seq[C], resultFn func(T, C) U) iter.Seq[U] {
	return selectManyIndexedResultSeq(seq, collectionFn, resultFn)
}

func (seq iter.Seq[T]) OrderBy[K cmp.Ordered](key func(T) K) Ordered[T] {
	return OrderBy(seq, key)
}

func (seq iter.Seq[T]) OrderByDescending[K cmp.Ordered](key func(T) K) Ordered[T] {
	return OrderByDescending(seq, key)
}

func (seq iter.Seq[T]) Order[T cmp.Ordered]() Ordered[T] {
	return Order(seq)
}

func (seq iter.Seq[T]) OrderDescending[T cmp.Ordered]() Ordered[T] {
	return OrderDescending(seq)
}

func (seq iter.Seq[T]) Take(n int) iter.Seq[T] {
	return Take(seq, n)
}

func (seq iter.Seq[T]) Skip(n int) iter.Seq[T] {
	return Skip(seq, n)
}

func (seq iter.Seq[T]) TakeWhile(pred func(T) bool) iter.Seq[T] {
	return TakeWhile(seq, pred)
}

func (seq iter.Seq[T]) SkipWhile(pred func(T) bool) iter.Seq[T] {
	return SkipWhile(seq, pred)
}

func (seq iter.Seq[T]) TakeLast(count int) iter.Seq[T] {
	return TakeLast(seq, count)
}

func (seq iter.Seq[T]) SkipLast(count int) iter.Seq[T] {
	return SkipLast(seq, count)
}

func (seq iter.Seq[T]) Append(element T) iter.Seq[T] {
	return Append(seq, element)
}

func (seq iter.Seq[T]) Prepend(element T) iter.Seq[T] {
	return Prepend(seq, element)
}

func (seq iter.Seq[T]) Concat(other iter.Seq[T]) iter.Seq[T] {
	return Concat(seq, other)
}

func (seq iter.Seq[T]) Zip[U, R any](other iter.Seq[U], fn func(T, U) R) iter.Seq[R] {
	return Zip(seq, other, fn)
}

func (seq iter.Seq[T]) Reverse() iter.Seq[T] {
	return Reverse(seq)
}

func (seq iter.Seq[T]) Chunk(size int) iter.Seq[[]T] {
	return Chunk(seq, size)
}

func (seq iter.Seq[T]) DefaultIfEmpty() iter.Seq[T] {
	return defaultIfEmptyZeroSeq(seq)
}

func (seq iter.Seq[T]) DefaultIfEmpty(defaultValue T) iter.Seq[T] {
	return defaultIfEmptySeq(seq, defaultValue)
}

func (seq iter.Seq[T]) Index() iter.Seq[Indexed[T]] {
	return Index(seq)
}

func (seq iter.Seq[T]) Shuffle() iter.Seq[T] {
	return Shuffle(seq)
}

func (seq iter.Seq[T]) Distinct[T comparable]() iter.Seq[T] {
	return Distinct(seq)
}

func (seq iter.Seq[T]) DistinctBy[K comparable](keyFn func(T) K) iter.Seq[T] {
	return DistinctBy(seq, keyFn)
}

func (seq iter.Seq[T]) Except[T comparable](other iter.Seq[T]) iter.Seq[T] {
	return Except(seq, other)
}

func (seq iter.Seq[T]) ExceptBy[K comparable](other iter.Seq[T], keyFn func(T) K) iter.Seq[T] {
	return ExceptBy(seq, other, keyFn)
}

func (seq iter.Seq[T]) Intersect[T comparable](other iter.Seq[T]) iter.Seq[T] {
	return Intersect(seq, other)
}

func (seq iter.Seq[T]) IntersectBy[K comparable](other iter.Seq[T], keyFn func(T) K) iter.Seq[T] {
	return IntersectBy(seq, other, keyFn)
}

func (seq iter.Seq[T]) Union[T comparable](other iter.Seq[T]) iter.Seq[T] {
	return Union(seq, other)
}

func (seq iter.Seq[T]) UnionBy[K comparable](other iter.Seq[T], keyFn func(T) K) iter.Seq[T] {
	return UnionBy(seq, other, keyFn)
}

func (seq iter.Seq[T]) Contains[T comparable](value T) bool {
	return Contains(seq, value)
}

func (seq iter.Seq[T]) SequenceEqual[T comparable](other iter.Seq[T]) bool {
	return SequenceEqual(seq, other)
}

func (seq iter.Seq[T]) ToList() []T {
	return ToList(seq)
}

func (seq iter.Seq[T]) ToArray() []T {
	return ToArray(seq)
}

func (seq iter.Seq[T]) ToHashSet[T comparable]() HashSet[T] {
	return ToHashSet(seq)
}

func (seq iter.Seq[T]) ToDictionary[K comparable, V any](keyFn func(T) K, valueFn func(T) V) map[K]V {
	return ToDictionary(seq, keyFn, valueFn)
}

func (seq iter.Seq[T]) ToLookup[K comparable, V any](keyFn func(T) K, valueFn func(T) V) Lookup[K, V] {
	return ToLookup(seq, keyFn, valueFn)
}

func (seq iter.Seq[T]) First() T {
	return firstFromSeq(seq)
}

func (seq iter.Seq[T]) First(pred func(T) bool) T {
	return firstPredSeq(seq, pred)
}

func (seq iter.Seq[T]) FirstOrDefault() T {
	return firstOrDefaultSeq(seq)
}

func (seq iter.Seq[T]) FirstOrDefault(pred func(T) bool) T {
	return firstOrDefaultPredSeq(seq, pred)
}

func (seq iter.Seq[T]) FirstOrDefault(pred func(T) bool, defaultValue T) T {
	return firstOrDefaultPredValueSeq(seq, pred, defaultValue)
}

func (seq iter.Seq[T]) Last() T {
	return lastFromSeq(seq)
}

func (seq iter.Seq[T]) Last(pred func(T) bool) T {
	return lastPredSeq(seq, pred)
}

func (seq iter.Seq[T]) LastOrDefault() T {
	return lastOrDefaultSeq(seq)
}

func (seq iter.Seq[T]) LastOrDefault(pred func(T) bool) T {
	return lastOrDefaultPredSeq(seq, pred)
}

func (seq iter.Seq[T]) LastOrDefault(pred func(T) bool, defaultValue T) T {
	return lastOrDefaultPredValueSeq(seq, pred, defaultValue)
}

func (seq iter.Seq[T]) Single() T {
	return singleSeq(seq)
}

func (seq iter.Seq[T]) Single(pred func(T) bool) T {
	return singlePredSeq(seq, pred)
}

func (seq iter.Seq[T]) SingleOrDefault(defaultValue T) T {
	return singleOrDefaultSeq(seq, defaultValue)
}

func (seq iter.Seq[T]) SingleOrDefault(pred func(T) bool, defaultValue T) T {
	return singleOrDefaultPredSeq(seq, pred, defaultValue)
}

func (seq iter.Seq[T]) ElementAt(index int) T {
	return ElementAt(seq, index)
}

func (seq iter.Seq[T]) ElementAtOrDefault(index int, defaultValue T) T {
	return ElementAtOrDefault(seq, index, defaultValue)
}

func (seq iter.Seq[T]) Any() bool {
	return anyWithoutPredSeq(seq)
}

func (seq iter.Seq[T]) Any(pred func(T) bool) bool {
	return anySeq(seq, pred)
}

func (seq iter.Seq[T]) All(pred func(T) bool) bool {
	return allSeq(seq, pred)
}

func (seq iter.Seq[T]) Count() int {
	return countSeq(seq)
}

func (seq iter.Seq[T]) Count(pred func(T) bool) int {
	return countPredSeq(seq, pred)
}

func (seq iter.Seq[T]) LongCount() int64 {
	return longCountSeq(seq)
}

func (seq iter.Seq[T]) LongCount(pred func(T) bool) int64 {
	return longCountPredSeq(seq, pred)
}

func (seq iter.Seq[T]) CountBy[K comparable](keyFn func(T) K) iter.Seq[KeyValue[K, int]] {
	return CountBy(seq, keyFn)
}

func (seq iter.Seq[T]) Aggregate(fn func(T, T) T) T {
	return aggregateSeq(seq, fn)
}

func (seq iter.Seq[T]) AggregateWithSeed[U any](seed U, fn func(U, T) U) U {
	return aggregateWithSeedSeq(seq, seed, fn)
}

func (seq iter.Seq[T]) Aggregate[U any, R any](seed U, fn func(U, T) U, resultFn func(U) R) R {
	return aggregateResultSeq(seq, seed, fn, resultFn)
}

func (seq iter.Seq[T]) AggregateBy[K comparable, A any, R any](keyFn func(T) K, seed A, fn func(A, T) A, resultFn func(A) R) iter.Seq[KeyValue[K, R]] {
	return aggregateBySeq(seq, keyFn, seed, fn, resultFn)
}

func (seq iter.Seq[T]) AggregateBy[K comparable, A any, R any](keyFn func(T) K, seedFn func(T) A, fn func(A, T) A, resultFn func(A) R) iter.Seq[KeyValue[K, R]] {
	return aggregateByFactorySeq(seq, keyFn, seedFn, fn, resultFn)
}

func (seq iter.Seq[T]) Sum[T Number]() T {
	return sumSeq(seq)
}

func (seq iter.Seq[T]) Sum[U Number](selector func(T) U) U {
	return sumBySeq(seq, selector)
}

func (seq iter.Seq[T]) Average[T Number]() float64 {
	return averageSeq(seq)
}

func (seq iter.Seq[T]) Average[U Number](selector func(T) U) float64 {
	return averageBySeq(seq, selector)
}

func (seq iter.Seq[T]) Max[T cmp.Ordered]() T {
	return maxSeq(seq)
}

func (seq iter.Seq[T]) MaxBy[K cmp.Ordered](key func(T) K) T {
	return MaxBy(seq, key)
}

func (seq iter.Seq[T]) Min[T cmp.Ordered]() T {
	return minSeq(seq)
}

func (seq iter.Seq[T]) MinBy[K cmp.Ordered](key func(T) K) T {
	return MinBy(seq, key)
}

func (seq iter.Seq[T]) GroupBy[K comparable](keyFn func(T) K) iter.Seq[Group[K, T]] {
	return GroupBy(seq, keyFn)
}

func (seq iter.Seq[T]) Join[U any, K comparable, R any](inner iter.Seq[U], outerKey func(T) K, innerKey func(U) K, resultFn func(T, U) R) iter.Seq[R] {
	return Join(seq, inner, outerKey, innerKey, resultFn)
}

func (seq iter.Seq[T]) GroupJoin[U any, K comparable, R any](inner iter.Seq[U], outerKey func(T) K, innerKey func(U) K, resultFn func(T, iter.Seq[U]) R) iter.Seq[R] {
	return GroupJoin(seq, inner, outerKey, innerKey, resultFn)
}

func (seq iter.Seq[T]) LeftJoin[U any, K comparable, R any](inner iter.Seq[U], outerKey func(T) K, innerKey func(U) K, resultFn func(T, U) R, defaultInner U) iter.Seq[R] {
	return LeftJoin(seq, inner, outerKey, innerKey, resultFn, defaultInner)
}

func (seq iter.Seq[T]) RightJoin[U any, K comparable, R any](inner iter.Seq[U], outerKey func(T) K, innerKey func(U) K, resultFn func(T, U) R, defaultOuter T) iter.Seq[R] {
	return RightJoin(seq, inner, outerKey, innerKey, resultFn, defaultOuter)
}

func (seq iter.Seq[T]) FullJoin[U any, K comparable, R any](inner iter.Seq[U], outerKey func(T) K, innerKey func(U) K, resultFn func(T, U) R, defaultOuter T, defaultInner U) iter.Seq[R] {
	return FullJoin(seq, inner, outerKey, innerKey, resultFn, defaultOuter, defaultInner)
}

func (seq iter.Seq[any]) Cast[U any]() iter.Seq[U] {
	return Cast[U](seq)
}

func (seq iter.Seq[any]) OfType[U any]() iter.Seq[U] {
	return OfType[U](seq)
}

func (seq iter.Seq[T]) TryGetSeqLen() (int, bool) {
	return tryGetSeqLenSeq(seq)
}

