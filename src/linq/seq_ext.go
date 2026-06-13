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

func (seq iter.Seq[T]) AsSeq[T any]() iter.Seq[T] {
	return AsSeq(seq)
}

func (seq iter.Seq[T]) Where[T any](pred func(T) bool) iter.Seq[T] {
	return Where(seq, pred)
}

func (seq iter.Seq[T]) Select[T, U any](fn func(T) U) iter.Seq[U] {
	return Select(seq, fn)
}

func (seq iter.Seq[T]) SelectMany[T, U any](fn func(T) iter.Seq[U]) iter.Seq[U] {
	return SelectMany(seq, fn)
}

func (seq iter.Seq[T]) OrderBy[T, K cmp.Ordered](key func(T) K) Ordered[T] {
	return OrderBy(seq, key)
}

func (seq iter.Seq[T]) OrderByDescending[T, K cmp.Ordered](key func(T) K) Ordered[T] {
	return OrderByDescending(seq, key)
}

func (seq iter.Seq[T]) Order[T cmp.Ordered]() Ordered[T] {
	return Order(seq)
}

func (seq iter.Seq[T]) OrderDescending[T cmp.Ordered]() Ordered[T] {
	return OrderDescending(seq)
}

func (seq iter.Seq[T]) Take[T any](n int) iter.Seq[T] {
	return Take(seq, n)
}

func (seq iter.Seq[T]) Skip[T any](n int) iter.Seq[T] {
	return Skip(seq, n)
}

func (seq iter.Seq[T]) TakeWhile[T any](pred func(T) bool) iter.Seq[T] {
	return TakeWhile(seq, pred)
}

func (seq iter.Seq[T]) SkipWhile[T any](pred func(T) bool) iter.Seq[T] {
	return SkipWhile(seq, pred)
}

func (seq iter.Seq[T]) TakeLast[T any](count int) iter.Seq[T] {
	return TakeLast(seq, count)
}

func (seq iter.Seq[T]) SkipLast[T any](count int) iter.Seq[T] {
	return SkipLast(seq, count)
}

func (seq iter.Seq[T]) Append[T any](element T) iter.Seq[T] {
	return Append(seq, element)
}

func (seq iter.Seq[T]) Prepend[T any](element T) iter.Seq[T] {
	return Prepend(seq, element)
}

func (seq iter.Seq[T]) Concat[T any](other iter.Seq[T]) iter.Seq[T] {
	return Concat(seq, other)
}

func (seq iter.Seq[T]) Zip[T, U, R any](other iter.Seq[U], fn func(T, U) R) iter.Seq[R] {
	return Zip(seq, other, fn)
}

func (seq iter.Seq[T]) Reverse[T any]() iter.Seq[T] {
	return Reverse(seq)
}

func (seq iter.Seq[T]) Chunk[T any](size int) iter.Seq[[]T] {
	return Chunk(seq, size)
}

func (seq iter.Seq[T]) DefaultIfEmpty[T any](defaultValue T) iter.Seq[T] {
	return DefaultIfEmpty(seq, defaultValue)
}

func (seq iter.Seq[T]) Index[T any]() iter.Seq[Indexed[T]] {
	return Index(seq)
}

func (seq iter.Seq[T]) Shuffle[T any]() iter.Seq[T] {
	return Shuffle(seq)
}

func (seq iter.Seq[T]) Distinct[T comparable]() iter.Seq[T] {
	return Distinct(seq)
}

func (seq iter.Seq[T]) DistinctBy[T, K comparable](keyFn func(T) K) iter.Seq[T] {
	return DistinctBy(seq, keyFn)
}

func (seq iter.Seq[T]) Except[T comparable](other iter.Seq[T]) iter.Seq[T] {
	return Except(seq, other)
}

func (seq iter.Seq[T]) ExceptBy[T, K comparable](other iter.Seq[T], keyFn func(T) K) iter.Seq[T] {
	return ExceptBy(seq, other, keyFn)
}

func (seq iter.Seq[T]) Intersect[T comparable](other iter.Seq[T]) iter.Seq[T] {
	return Intersect(seq, other)
}

func (seq iter.Seq[T]) IntersectBy[T, K comparable](other iter.Seq[T], keyFn func(T) K) iter.Seq[T] {
	return IntersectBy(seq, other, keyFn)
}

func (seq iter.Seq[T]) Union[T comparable](other iter.Seq[T]) iter.Seq[T] {
	return Union(seq, other)
}

func (seq iter.Seq[T]) UnionBy[T, K comparable](other iter.Seq[T], keyFn func(T) K) iter.Seq[T] {
	return UnionBy(seq, other, keyFn)
}

func (seq iter.Seq[T]) Contains[T comparable](value T) bool {
	return Contains(seq, value)
}

func (seq iter.Seq[T]) SequenceEqual[T comparable](other iter.Seq[T]) bool {
	return SequenceEqual(seq, other)
}

func (seq iter.Seq[T]) ToList[T any]() []T {
	return ToList(seq)
}

func (seq iter.Seq[T]) ToArray[T any]() []T {
	return ToArray(seq)
}

func (seq iter.Seq[T]) ToHashSet[T comparable]() HashSet[T] {
	return ToHashSet(seq)
}

func (seq iter.Seq[T]) ToDictionary[T, K comparable, V any](keyFn func(T) K, valueFn func(T) V) map[K]V {
	return ToDictionary(seq, keyFn, valueFn)
}

func (seq iter.Seq[T]) ToLookup[T, K comparable, V any](keyFn func(T) K, valueFn func(T) V) Lookup[K, V] {
	return ToLookup(seq, keyFn, valueFn)
}

func (seq iter.Seq[T]) First[T any]() T {
	return First(seq)
}

func (seq iter.Seq[T]) FirstOrDefault[T any]() T {
	return FirstOrDefault(seq)
}

func (seq iter.Seq[T]) Last[T any]() T {
	return Last(seq)
}

func (seq iter.Seq[T]) LastOrDefault[T any]() T {
	return LastOrDefault(seq)
}

func (seq iter.Seq[T]) Single[T any]() T {
	return Single(seq)
}

func (seq iter.Seq[T]) SingleOrDefault[T any](defaultValue T) T {
	return SingleOrDefault(seq, defaultValue)
}

func (seq iter.Seq[T]) ElementAt[T any](index int) T {
	return ElementAt(seq, index)
}

func (seq iter.Seq[T]) ElementAtOrDefault[T any](index int, defaultValue T) T {
	return ElementAtOrDefault(seq, index, defaultValue)
}

func (seq iter.Seq[T]) Any[T any](pred func(T) bool) bool {
	return Any(seq, pred)
}

func (seq iter.Seq[T]) All[T any](pred func(T) bool) bool {
	return All(seq, pred)
}

func (seq iter.Seq[T]) Count[T any]() int {
	return Count(seq)
}

func (seq iter.Seq[T]) LongCount[T any]() int64 {
	return LongCount(seq)
}

func (seq iter.Seq[T]) CountBy[T, K comparable](keyFn func(T) K) iter.Seq[KeyValue[K, int]] {
	return CountBy(seq, keyFn)
}

func (seq iter.Seq[T]) Aggregate[T any](fn func(T, T) T) T {
	return Aggregate(seq, fn)
}

func (seq iter.Seq[T]) AggregateWithSeed[T, U any](seed U, fn func(U, T) U) U {
	return AggregateWithSeed(seq, seed, fn)
}

func (seq iter.Seq[T]) AggregateBy[T, K comparable, A any, R any](keyFn func(T) K, seed A, fn func(A, T) A, resultFn func(A) R) iter.Seq[KeyValue[K, R]] {
	return AggregateBy(seq, keyFn, seed, fn, resultFn)
}

func (seq iter.Seq[T]) Sum[T Number]() T {
	return Sum(seq)
}

func (seq iter.Seq[T]) Average[T Number]() float64 {
	return Average(seq)
}

func (seq iter.Seq[T]) Max[T cmp.Ordered]() T {
	return Max(seq)
}

func (seq iter.Seq[T]) MaxBy[T, K cmp.Ordered](key func(T) K) T {
	return MaxBy(seq, key)
}

func (seq iter.Seq[T]) Min[T cmp.Ordered]() T {
	return Min(seq)
}

func (seq iter.Seq[T]) MinBy[T, K cmp.Ordered](key func(T) K) T {
	return MinBy(seq, key)
}

func (seq iter.Seq[T]) GroupBy[T, K comparable](keyFn func(T) K) iter.Seq[Group[K, T]] {
	return GroupBy(seq, keyFn)
}

func (seq iter.Seq[T]) Join[T, U, K comparable, R any](inner iter.Seq[U], outerKey func(T) K, innerKey func(U) K, resultFn func(T, U) R) iter.Seq[R] {
	return Join(seq, inner, outerKey, innerKey, resultFn)
}

func (seq iter.Seq[T]) GroupJoin[T, U, K comparable, R any](inner iter.Seq[U], outerKey func(T) K, innerKey func(U) K, resultFn func(T, iter.Seq[U]) R) iter.Seq[R] {
	return GroupJoin(seq, inner, outerKey, innerKey, resultFn)
}

func (seq iter.Seq[T]) LeftJoin[T, U, K comparable, R any](inner iter.Seq[U], outerKey func(T) K, innerKey func(U) K, resultFn func(T, U) R, defaultInner U) iter.Seq[R] {
	return LeftJoin(seq, inner, outerKey, innerKey, resultFn, defaultInner)
}

func (seq iter.Seq[T]) RightJoin[T, U, K comparable, R any](inner iter.Seq[U], outerKey func(T) K, innerKey func(U) K, resultFn func(T, U) R, defaultOuter T) iter.Seq[R] {
	return RightJoin(seq, inner, outerKey, innerKey, resultFn, defaultOuter)
}

func (seq iter.Seq[T]) FullJoin[T, U, K comparable, R any](inner iter.Seq[U], outerKey func(T) K, innerKey func(U) K, resultFn func(T, U) R, defaultOuter T, defaultInner U) iter.Seq[R] {
	return FullJoin(seq, inner, outerKey, innerKey, resultFn, defaultOuter, defaultInner)
}

func (seq iter.Seq[T]) TryGetSeqLen[T any]() (int, bool) {
	return TryGetSeqLen(seq)
}

