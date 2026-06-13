// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

import (
	"cmp"
	"iter"
	"slices"
)

// --- core sequence operators ---

func whereSeq[T any](seq iter.Seq[T], pred func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if pred(v) && !yield(v) {
				return
			}
		}
	}
}

func selectSeq[T any](seq iter.Seq[T], fn func(T) T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if !yield(fn(v)) {
				return
			}
		}
	}
}

func selectBySeq[T, U any](seq iter.Seq[T], fn func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for v := range seq {
			if !yield(fn(v)) {
				return
			}
		}
	}
}

func skipSeq[T any](seq iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		if n < 0 {
			n = 0
		}
		for v := range seq {
			if n > 0 {
				n--
				continue
			}
			if !yield(v) {
				return
			}
		}
	}
}

func takeSeq[T any](seq iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		if n <= 0 {
			return
		}
		for v := range seq {
			if n <= 0 {
				return
			}
			n--
			if !yield(v) {
				return
			}
		}
	}
}

func firstSeq[T any](seq iter.Seq[T]) (T, bool) {
	for v := range seq {
		return v, true
	}
	var z T
	return z, false
}

func firstOrDefaultSeq[T any](seq iter.Seq[T]) T {
	if v, ok := firstSeq(seq); ok {
		return v
	}
	var z T
	return z
}

func anySeq[T any](seq iter.Seq[T], pred func(T) bool) bool {
	for v := range seq {
		if pred(v) {
			return true
		}
	}
	return false
}

func allSeq[T any](seq iter.Seq[T], pred func(T) bool) bool {
	for v := range seq {
		if !pred(v) {
			return false
		}
	}
	return true
}

func lenSeq[T any](seq iter.Seq[T]) int {
	n := 0
	for range seq {
		n++
	}
	return n
}

func longLenSeq[T any](seq iter.Seq[T]) int64 {
	var n int64
	for range seq {
		n++
	}
	return n
}

func toListSeq[T any](seq iter.Seq[T]) []T {
	return slices.Collect(seq)
}

// --- terminal operators ---

func firstFromSeq[T any](seq iter.Seq[T]) T {
	v, ok := firstSeq(seq)
	if !ok {
		panic("linq: sequence contains no elements")
	}
	return v
}

func lastSeq[T any](seq iter.Seq[T]) (T, bool) {
	var last T
	found := false
	for v := range seq {
		last = v
		found = true
	}
	return last, found
}

func lastFromSeq[T any](seq iter.Seq[T]) T {
	v, ok := lastSeq(seq)
	if !ok {
		panic("linq: sequence contains no elements")
	}
	return v
}

func lastOrDefaultSeq[T any](seq iter.Seq[T]) T {
	if v, ok := lastSeq(seq); ok {
		return v
	}
	var z T
	return z
}

func singleSeq[T any](seq iter.Seq[T]) T {
	next, stop := iter.Pull(seq)
	defer stop()
	v, ok := next()
	if !ok {
		panic("linq: sequence contains no elements")
	}
	if _, ok = next(); ok {
		panic("linq: sequence contains more than one element")
	}
	return v
}

func singleOrDefaultSeq[T any](seq iter.Seq[T], defaultValue T) T {
	next, stop := iter.Pull(seq)
	defer stop()
	v, ok := next()
	if !ok {
		return defaultValue
	}
	if _, ok = next(); ok {
		panic("linq: sequence contains more than one element")
	}
	return v
}

func elementAtSeq[T any](seq iter.Seq[T], index int) T {
	if index < 0 {
		panic("linq: index out of range")
	}
	i := 0
	for v := range seq {
		if i == index {
			return v
		}
		i++
	}
	panic("linq: index out of range")
}

func elementAtOrDefaultSeq[T any](seq iter.Seq[T], index int, defaultValue T) T {
	if index < 0 {
		return defaultValue
	}
	i := 0
	for v := range seq {
		if i == index {
			return v
		}
		i++
	}
	return defaultValue
}

func aggregateSeq[T any](seq iter.Seq[T], fn func(T, T) T) T {
	v, ok := firstSeq(seq)
	if !ok {
		panic("linq: sequence contains no elements")
	}
	acc := v
	for v := range seq {
		acc = fn(acc, v)
	}
	return acc
}

func aggregateWithSeedSeq[T any, U any](seq iter.Seq[T], seed U, fn func(U, T) U) U {
	acc := seed
	for v := range seq {
		acc = fn(acc, v)
	}
	return acc
}

func sumSeq[U Number](seq iter.Seq[U]) U {
	var acc U
	for v := range seq {
		acc += v
	}
	return acc
}

func averageSeq[U Number](seq iter.Seq[U]) float64 {
	var sum float64
	n := 0
	for v := range seq {
		sum += float64(v)
		n++
	}
	if n == 0 {
		panic("linq: sequence contains no elements")
	}
	return sum / float64(n)
}

func maxSeq[T cmp.Ordered](seq iter.Seq[T]) T {
	v, ok := firstSeq(seq)
	if !ok {
		panic("linq: sequence contains no elements")
	}
	max := v
	for v := range seq {
		if v > max {
			max = v
		}
	}
	return max
}

func minSeq[T cmp.Ordered](seq iter.Seq[T]) T {
	v, ok := firstSeq(seq)
	if !ok {
		panic("linq: sequence contains no elements")
	}
	min := v
	for v := range seq {
		if v < min {
			min = v
		}
	}
	return min
}

func maxBySeq[T any, K cmp.Ordered](seq iter.Seq[T], key func(T) K) T {
	v, ok := firstSeq(seq)
	if !ok {
		panic("linq: sequence contains no elements")
	}
	best := v
	bestKey := key(v)
	for v := range seq {
		k := key(v)
		if k > bestKey {
			best = v
			bestKey = k
		}
	}
	return best
}

func minBySeq[T any, K cmp.Ordered](seq iter.Seq[T], key func(T) K) T {
	v, ok := firstSeq(seq)
	if !ok {
		panic("linq: sequence contains no elements")
	}
	best := v
	bestKey := key(v)
	for v := range seq {
		k := key(v)
		if k < bestKey {
			best = v
			bestKey = k
		}
	}
	return best
}

func containsSeq[T comparable](seq iter.Seq[T], value T) bool {
	for v := range seq {
		if v == value {
			return true
		}
	}
	return false
}

func sequenceEqualSeq[T comparable](a, b iter.Seq[T]) bool {
	ai, aStop := iter.Pull(a)
	bi, bStop := iter.Pull(b)
	defer aStop()
	defer bStop()
	for {
		va, oka := ai()
		vb, okb := bi()
		if oka != okb {
			return false
		}
		if !oka {
			return true
		}
		if va != vb {
			return false
		}
	}
}

func toHashSetSeq[T comparable](seq iter.Seq[T]) HashSet[T] {
	set := make(HashSet[T])
	for v := range seq {
		set[v] = struct{}{}
	}
	return set
}

func toDictionarySeq[T any, K comparable, V any](seq iter.Seq[T], keyFn func(T) K, valueFn func(T) V) map[K]V {
	out := make(map[K]V)
	for v := range seq {
		k := keyFn(v)
		if _, dup := out[k]; dup {
			panic("linq: duplicate key in ToDictionary")
		}
		out[k] = valueFn(v)
	}
	return out
}

func toLookupSeq[T any, K comparable, V any](seq iter.Seq[T], keyFn func(T) K, valueFn func(T) V) Lookup[K, V] {
	groups := map[K][]V{}
	var keys []K
	for item := range seq {
		k := keyFn(item)
		if _, seen := groups[k]; !seen {
			keys = append(keys, k)
			groups[k] = nil
		}
		groups[k] = append(groups[k], valueFn(item))
	}
	return Lookup[K, V]{groups: groups, keys: keys}
}

func tryGetSeqLenSeq[T any](seq iter.Seq[T]) (int, bool) {
	return 0, false
}

// --- grouping and aggregation sequences ---

func distinctSeq[T comparable](seq iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		seen := make(map[T]struct{})
		for v := range seq {
			if _, dup := seen[v]; dup {
				continue
			}
			seen[v] = struct{}{}
			if !yield(v) {
				return
			}
		}
	}
}

func groupBySeq[T any, K comparable](seq iter.Seq[T], keyFn func(T) K) iter.Seq[Group[K, T]] {
	groups := map[K][]T{}
	var keys []K
	for v := range seq {
		k := keyFn(v)
		if _, seen := groups[k]; !seen {
			keys = append(keys, k)
			groups[k] = nil
		}
		groups[k] = append(groups[k], v)
	}
	idx := 0
	return func(yield func(Group[K, T]) bool) {
		for idx < len(keys) {
			k := keys[idx]
			g := Group[K, T]{Key: k, items: groups[k]}
			idx++
			if !yield(g) {
				return
			}
		}
	}
}

func countBySeq[T any, K comparable](seq iter.Seq[T], keyFn func(T) K) iter.Seq[KeyValue[K, int]] {
	counts := map[K]int{}
	var keys []K
	for v := range seq {
		k := keyFn(v)
		if _, seen := counts[k]; !seen {
			keys = append(keys, k)
		}
		counts[k]++
	}
	idx := 0
	return func(yield func(KeyValue[K, int]) bool) {
		for idx < len(keys) {
			k := keys[idx]
			idx++
			if !yield(KeyValue[K, int]{Key: k, Value: counts[k]}) {
				return
			}
		}
	}
}

func aggregateBySeq[T any, K comparable, A any, R any](seq iter.Seq[T], keyFn func(T) K, seed A, fn func(A, T) A, resultFn func(A) R) iter.Seq[KeyValue[K, R]] {
	groups := map[K]A{}
	var keys []K
	for v := range seq {
		k := keyFn(v)
		if _, seen := groups[k]; !seen {
			keys = append(keys, k)
			groups[k] = seed
		}
		groups[k] = fn(groups[k], v)
	}
	idx := 0
	return func(yield func(KeyValue[K, R]) bool) {
		for idx < len(keys) {
			k := keys[idx]
			idx++
			if !yield(KeyValue[K, R]{Key: k, Value: resultFn(groups[k])}) {
				return
			}
		}
	}
}

// FirstValue returns the first element of s and whether it exists.
func FirstValue[T any](s []T) (T, bool) {
	return firstSlice(s)
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
