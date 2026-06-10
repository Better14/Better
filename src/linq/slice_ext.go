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

func lazy[T any](s []T) Lazy[T] { return From(s) }

func (s []T) AsEnumerable[T any]() Lazy[T] {
	return From(s)
}

func (s []T) Where[T any](pred func(T) bool) Lazy[T] {
	return lazyWhere(lazy(s), pred)
}

func (s []T) Select[T, U any](fn func(T) U) Lazy[U] {
	return lazySelectBy(lazy(s), fn)
}

func (s []T) SelectMany[T, U any](fn func(T) Lazy[U]) Lazy[U] {
	return lazySelectMany(lazy(s), fn)
}

func (s []T) OrderBy[T, K cmp.Ordered](key func(T) K) Ordered[T] {
	return lazyOrderBy(lazy(s), key)
}

func (s []T) OrderByDescending[T, K cmp.Ordered](key func(T) K) Ordered[T] {
	return lazyOrderByDescending(lazy(s), key)
}

func (s []T) Order[T cmp.Ordered]() Ordered[T] {
	return lazyOrder(lazy(s))
}

func (s []T) OrderDescending[T cmp.Ordered]() Ordered[T] {
	return lazyOrderDescending(lazy(s))
}

func (s []T) Take[T any](n int) Lazy[T] {
	return lazyTake(lazy(s), n)
}

func (s []T) Skip[T any](n int) Lazy[T] {
	return lazySkip(lazy(s), n)
}

func (s []T) TakeWhile[T any](pred func(T) bool) Lazy[T] {
	return lazyTakeWhile(lazy(s), pred)
}

func (s []T) SkipWhile[T any](pred func(T) bool) Lazy[T] {
	return lazySkipWhile(lazy(s), pred)
}

func (s []T) TakeLast[T any](count int) Lazy[T] {
	return lazyTakeLast(lazy(s), count)
}

func (s []T) SkipLast[T any](count int) Lazy[T] {
	return lazySkipLast(lazy(s), count)
}

func (s []T) Append[T any](element T) Lazy[T] {
	return lazyAppend(lazy(s), element)
}

func (s []T) Prepend[T any](element T) Lazy[T] {
	return lazyPrepend(lazy(s), element)
}

func (s []T) Concat[T any](other Lazy[T]) Lazy[T] {
	return lazyConcat(lazy(s), other)
}

func (s []T) Zip[T, U, R any](other Lazy[U], fn func(T, U) R) Lazy[R] {
	return lazyZip(lazy(s), other, fn)
}

func (s []T) Reverse[T any]() Lazy[T] {
	return lazyReverse(lazy(s))
}

func (s []T) Chunk[T any](size int) Lazy[[]T] {
	return lazyChunk(lazy(s), size)
}

func (s []T) DefaultIfEmpty[T any](defaultValue T) Lazy[T] {
	return lazyDefaultIfEmpty(lazy(s), defaultValue)
}

func (s []T) Index[T any]() Lazy[Indexed[T]] {
	return lazyIndex(lazy(s))
}

func (s []T) Shuffle[T any]() Lazy[T] {
	return lazyShuffle(lazy(s))
}

func (s []T) Distinct[T comparable]() Lazy[T] {
	return lazyDistinct(lazy(s))
}

func (s []T) DistinctBy[T, K comparable](keyFn func(T) K) Lazy[T] {
	return lazyDistinctBy(lazy(s), keyFn)
}

func (s []T) Except[T comparable](other Lazy[T]) Lazy[T] {
	return lazyExcept(lazy(s), other)
}

func (s []T) ExceptBy[T, K comparable](other Lazy[T], keyFn func(T) K) Lazy[T] {
	return lazyExceptBy(lazy(s), other, keyFn)
}

func (s []T) Intersect[T comparable](other Lazy[T]) Lazy[T] {
	return lazyIntersect(lazy(s), other)
}

func (s []T) IntersectBy[T, K comparable](other Lazy[T], keyFn func(T) K) Lazy[T] {
	return lazyIntersectBy(lazy(s), other, keyFn)
}

func (s []T) Union[T comparable](other Lazy[T]) Lazy[T] {
	return lazyUnion(lazy(s), other)
}

func (s []T) UnionBy[T, K comparable](other Lazy[T], keyFn func(T) K) Lazy[T] {
	return lazyUnionBy(lazy(s), other, keyFn)
}

func (s []T) Contains[T comparable](value T) bool {
	return lazyContains(lazy(s), value)
}

func (s []T) SequenceEqual[T comparable](other Lazy[T]) bool {
	return lazySequenceEqual(lazy(s), other)
}

func (s []T) ToList[T any]() []T {
	return slices.Clone(s)
}

func (s []T) ToArray[T any]() []T {
	return lazyToArray(lazy(s))
}

func (s []T) ToHashSet[T comparable]() HashSet[T] {
	return lazyToHashSet(lazy(s))
}

func (s []T) ToDictionary[T, K comparable, V any](keyFn func(T) K, valueFn func(T) V) map[K]V {
	return lazyToDictionary(lazy(s), keyFn, valueFn)
}

func (s []T) ToLookup[T, K comparable, V any](keyFn func(T) K, valueFn func(T) V) Lookup[K, V] {
	return lazyToLookup(lazy(s), keyFn, valueFn)
}

func (s []T) First[T any]() T {
	return lazy(s).First()
}

func (s []T) FirstOrDefault[T any]() T {
	return lazy(s).FirstOrDefault()
}

func (s []T) Last[T any]() T {
	return lazyLast(lazy(s))
}

func (s []T) LastOrDefault[T any]() T {
	return lazyLastOrDefault(lazy(s))
}

func (s []T) Single[T any]() T {
	return lazySingle(lazy(s))
}

func (s []T) SingleOrDefault[T any](defaultValue T) T {
	return lazySingleOrDefault(lazy(s), defaultValue)
}

func (s []T) ElementAt[T any](index int) T {
	return lazyElementAt(lazy(s), index)
}

func (s []T) ElementAtOrDefault[T any](index int, defaultValue T) T {
	return lazyElementAtOrDefault(lazy(s), index, defaultValue)
}

func (s []T) Any[T any](pred func(T) bool) bool {
	return lazyAny(lazy(s), pred)
}

func (s []T) All[T any](pred func(T) bool) bool {
	return lazyAll(lazy(s), pred)
}

func (s []T) Count[T any]() int {
	return lazyCount(lazy(s))
}

func (s []T) LongCount[T any]() int64 {
	return lazyLongCount(lazy(s))
}

func (s []T) CountBy[T, K comparable](keyFn func(T) K) Lazy[KeyValue[K, int]] {
	return lazyCountBy(lazy(s), keyFn)
}

func (s []T) Aggregate[T any](fn func(T, T) T) T {
	return lazyAggregate(lazy(s), fn)
}

func (s []T) AggregateWithSeed[T, U any](seed U, fn func(U, T) U) U {
	return lazyAggregateWithSeed(lazy(s), seed, fn)
}

func (s []T) AggregateBy[T, K comparable, A any, R any](keyFn func(T) K, seed A, fn func(A, T) A, resultFn func(A) R) Lazy[KeyValue[K, R]] {
	return lazyAggregateBy(lazy(s), keyFn, seed, fn, resultFn)
}

func (s []T) Sum[T Number]() T {
	return lazySum(lazy(s))
}

func (s []T) Average[T Number]() float64 {
	return lazyAverage(lazy(s))
}

func (s []T) Max[T cmp.Ordered]() T {
	return lazyMax(lazy(s))
}

func (s []T) MaxBy[T, K cmp.Ordered](key func(T) K) T {
	return lazyMaxBy(lazy(s), key)
}

func (s []T) Min[T cmp.Ordered]() T {
	return lazyMin(lazy(s))
}

func (s []T) MinBy[T, K cmp.Ordered](key func(T) K) T {
	return lazyMinBy(lazy(s), key)
}

func (s []T) GroupBy[T, K comparable](keyFn func(T) K) Lazy[Group[K, T]] {
	return lazyGroupBy(lazy(s), keyFn)
}

func (s []T) Join[T, U, K comparable, R any](inner Lazy[U], outerKey func(T) K, innerKey func(U) K, resultFn func(T, U) R) Lazy[R] {
	return lazyJoin(lazy(s), inner, outerKey, innerKey, resultFn)
}

func (s []T) GroupJoin[T, U, K comparable, R any](inner Lazy[U], outerKey func(T) K, innerKey func(U) K, resultFn func(T, Lazy[U]) R) Lazy[R] {
	return lazyGroupJoin(lazy(s), inner, outerKey, innerKey, resultFn)
}

func (s []T) LeftJoin[T, U, K comparable, R any](inner Lazy[U], outerKey func(T) K, innerKey func(U) K, resultFn func(T, U) R, defaultInner U) Lazy[R] {
	return lazyLeftJoin(lazy(s), inner, outerKey, innerKey, resultFn, defaultInner)
}

func (s []T) RightJoin[T, U, K comparable, R any](inner Lazy[U], outerKey func(T) K, innerKey func(U) K, resultFn func(T, U) R, defaultOuter T) Lazy[R] {
	return lazyRightJoin(lazy(s), inner, outerKey, innerKey, resultFn, defaultOuter)
}

func (s []T) TryGetNonEnumeratedCount[T any]() (int, bool) {
	return tryGetNonEnumeratedCount(lazy(s))
}
