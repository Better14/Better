
package linq

import (
	"iter"
	"slices"
)

func anyWithoutPredSeq[T any](seq iter.Seq[T]) bool {
	for range seq {
		return true
	}
	return false
}

func countPredSeq[T any](seq iter.Seq[T], pred func(T) bool) int {
	n := 0
	for v := range seq {
		if pred(v) {
			n++
		}
	}
	return n
}

func longCountPredSeq[T any](seq iter.Seq[T], pred func(T) bool) int64 {
	var n int64
	for v := range seq {
		if pred(v) {
			n++
		}
	}
	return n
}

func firstPredSeq[T any](seq iter.Seq[T], pred func(T) bool) T {
	for v := range seq {
		if pred(v) {
			return v
		}
	}
	panic("linq: sequence contains no matching element")
}

func firstOrDefaultPredSeq[T any](seq iter.Seq[T], pred func(T) bool) T {
	for v := range seq {
		if pred(v) {
			return v
		}
	}
	var z T
	return z
}

func firstOrDefaultPredValueSeq[T any](seq iter.Seq[T], pred func(T) bool, defaultValue T) T {
	for v := range seq {
		if pred(v) {
			return v
		}
	}
	return defaultValue
}

func lastPredSeq[T any](seq iter.Seq[T], pred func(T) bool) T {
	var found T
	ok := false
	for v := range seq {
		if pred(v) {
			found = v
			ok = true
		}
	}
	if !ok {
		panic("linq: sequence contains no matching element")
	}
	return found
}

func lastOrDefaultPredSeq[T any](seq iter.Seq[T], pred func(T) bool) T {
	var found T
	ok := false
	for v := range seq {
		if pred(v) {
			found = v
			ok = true
		}
	}
	if !ok {
		var z T
		return z
	}
	return found
}

func lastOrDefaultPredValueSeq[T any](seq iter.Seq[T], pred func(T) bool, defaultValue T) T {
	var found T
	ok := false
	for v := range seq {
		if pred(v) {
			found = v
			ok = true
		}
	}
	if !ok {
		return defaultValue
	}
	return found
}

func singlePredSeq[T any](seq iter.Seq[T], pred func(T) bool) T {
	next, stop := iter.Pull(seq)
	defer stop()
	var match T
	count := 0
	for {
		v, ok := next()
		if !ok {
			break
		}
		if pred(v) {
			match = v
			count++
			if count > 1 {
				panic("linq: sequence contains more than one matching element")
			}
		}
	}
	if count == 0 {
		panic("linq: sequence contains no matching element")
	}
	return match
}

func singleOrDefaultPredSeq[T any](seq iter.Seq[T], pred func(T) bool, defaultValue T) T {
	next, stop := iter.Pull(seq)
	defer stop()
	var match T
	count := 0
	for {
		v, ok := next()
		if !ok {
			break
		}
		if pred(v) {
			match = v
			count++
			if count > 1 {
				panic("linq: sequence contains more than one matching element")
			}
		}
	}
	if count == 0 {
		return defaultValue
	}
	return match
}

func defaultIfEmptyZeroSeq[T any](seq iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		var z T
		ok := false
		for v := range seq {
			ok = true
			if !yield(v) {
				return
			}
		}
		if !ok && !yield(z) {
			return
		}
	}
}

func aggregateResultSeq[T any, U any, R any](seq iter.Seq[T], seed U, fn func(U, T) U, resultFn func(U) R) R {
	return resultFn(aggregateWithSeedSeq(seq, seed, fn))
}

func aggregateByFactorySeq[T any, K comparable, A any, R any](seq iter.Seq[T], keyFn func(T) K, seedFn func(T) A, fn func(A, T) A, resultFn func(A) R) iter.Seq[KeyValue[K, R]] {
	return func(yield func(KeyValue[K, R]) bool) {
		groups := map[K]A{}
		order := make([]K, 0)
		for item := range seq {
			k := keyFn(item)
			acc, ok := groups[k]
			if !ok {
				acc = seedFn(item)
				groups[k] = acc
				order = append(order, k)
			}
			groups[k] = fn(acc, item)
		}
		for _, k := range order {
			if !yield(KeyValue[K, R]{k, resultFn(groups[k])}) {
				return
			}
		}
	}
}

func whereIndexedSeq[T any](seq iter.Seq[T], pred func(T, int) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		i := 0
		for v := range seq {
			if pred(v, i) && !yield(v) {
				return
			}
			i++
		}
	}
}

func selectIndexedSeq[T, U any](seq iter.Seq[T], fn func(T, int) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		i := 0
		for v := range seq {
			if !yield(fn(v, i)) {
				return
			}
			i++
		}
	}
}

func selectManyIndexedSeq[T, U any](seq iter.Seq[T], fn func(T, int) iter.Seq[U]) iter.Seq[U] {
	return func(yield func(U) bool) {
		i := 0
		for v := range seq {
			for u := range fn(v, i) {
				if !yield(u) {
					return
				}
			}
			i++
		}
	}
}

func selectManyResultSeq[T, C, U any](seq iter.Seq[T], collectionFn func(T) iter.Seq[C], resultFn func(T, C) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for v := range seq {
			for c := range collectionFn(v) {
				if !yield(resultFn(v, c)) {
					return
				}
			}
		}
	}
}

func selectManySliceResultSeq[T, C, U any](seq iter.Seq[T], collectionFn func(T) []C, resultFn func(T, C) U) iter.Seq[U] {
	return selectManyResultSeq(seq, func(t T) iter.Seq[C] {
		return slices.Values(collectionFn(t))
	}, resultFn)
}

func selectManyIndexedResultSeq[T, C, U any](seq iter.Seq[T], collectionFn func(T, int) iter.Seq[C], resultFn func(T, C) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		i := 0
		for v := range seq {
			for c := range collectionFn(v, i) {
				if !yield(resultFn(v, c)) {
					return
				}
			}
			i++
		}
	}
}

func sumBySeq[T any, U Number](seq iter.Seq[T], selector func(T) U) U {
	return sumSeq(selectBySeq(seq, selector))
}

func averageBySeq[T any, U Number](seq iter.Seq[T], selector func(T) U) float64 {
	return averageSeq(selectBySeq(seq, selector))
}
