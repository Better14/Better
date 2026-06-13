// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

import (
	"iter"
	"math/rand"
	"slices"
)

func appendSeq[T any](seq iter.Seq[T], element T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if !yield(v) {
				return
			}
		}
		if !yield(element) {
			return
		}
	}
}

func prependSeq[T any](seq iter.Seq[T], element T) iter.Seq[T] {
	return func(yield func(T) bool) {
		if !yield(element) {
			return
		}
		for v := range seq {
			if !yield(v) {
				return
			}
		}
	}
}

func concatSeq[T any](a, b iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range a {
			if !yield(v) {
				return
			}
		}
		for v := range b {
			if !yield(v) {
				return
			}
		}
	}
}

func zipSeq[T, U, R any](a iter.Seq[T], b iter.Seq[U], fn func(T, U) R) iter.Seq[R] {
	return func(yield func(R) bool) {
		ai, aStop := iter.Pull(a)
		bi, bStop := iter.Pull(b)
		defer aStop()
		defer bStop()
		for {
			va, oka := ai()
			vb, okb := bi()
			if !oka || !okb {
				return
			}
			if !yield(fn(va, vb)) {
				return
			}
		}
	}
}

func selectManySeq[T, U any](seq iter.Seq[T], fn func(T) iter.Seq[U]) iter.Seq[U] {
	return func(yield func(U) bool) {
		for v := range seq {
			for u := range fn(v) {
				if !yield(u) {
					return
				}
			}
		}
	}
}

func selectManySliceSeq[T, U any](seq iter.Seq[T], fn func(T) []U) iter.Seq[U] {
	return selectManySeq(seq, func(t T) iter.Seq[U] {
		return slices.Values(fn(t))
	})
}

func reverseSeq[T any](seq iter.Seq[T]) iter.Seq[T] {
	items := slices.Collect(seq)
	slices.Reverse(items)
	return slices.Values(items)
}

func skipWhileSeq[T any](seq iter.Seq[T], pred func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		skipping := true
		for v := range seq {
			if skipping && pred(v) {
				continue
			}
			skipping = false
			if !yield(v) {
				return
			}
		}
	}
}

func takeWhileSeq[T any](seq iter.Seq[T], pred func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if !pred(v) {
				return
			}
			if !yield(v) {
				return
			}
		}
	}
}

func skipLastSeq[T any](seq iter.Seq[T], count int) iter.Seq[T] {
	items := slices.Collect(seq)
	if count >= len(items) {
		return Empty[T]()
	}
	return slices.Values(items[:len(items)-count])
}

func takeLastSeq[T any](seq iter.Seq[T], count int) iter.Seq[T] {
	items := slices.Collect(seq)
	if count <= 0 {
		return Empty[T]()
	}
	if count >= len(items) {
		return slices.Values(items)
	}
	return slices.Values(items[len(items)-count:])
}

func chunkSeq[T any](seq iter.Seq[T], size int) iter.Seq[[]T] {
	if size <= 0 {
		panic("linq: chunk size must be positive")
	}
	items := slices.Collect(seq)
	return func(yield func([]T) bool) {
		for i := 0; i < len(items); i += size {
			end := i + size
			if end > len(items) {
				end = len(items)
			}
			if !yield(slices.Clone(items[i:end])) {
				return
			}
		}
	}
}

func defaultIfEmptySeq[T any](seq iter.Seq[T], defaultValue T) iter.Seq[T] {
	return func(yield func(T) bool) {
		empty := true
		for v := range seq {
			empty = false
			if !yield(v) {
				return
			}
		}
		if empty {
			yield(defaultValue)
		}
	}
}

func indexSeq[T any](seq iter.Seq[T]) iter.Seq[Indexed[T]] {
	return func(yield func(Indexed[T]) bool) {
		i := 0
		for v := range seq {
			if !yield(Indexed[T]{Index: i, Value: v}) {
				return
			}
			i++
		}
	}
}

func castSeq[T any](seq iter.Seq[any]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			t, ok := v.(T)
			if !ok {
				panic("linq: cast failed")
			}
			if !yield(t) {
				return
			}
		}
	}
}

func ofTypeSeq[T any](seq iter.Seq[any]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			t, ok := v.(T)
			if ok && !yield(t) {
				return
			}
		}
	}
}

func shuffleSeq[T any](seq iter.Seq[T]) iter.Seq[T] {
	items := slices.Collect(seq)
	rand.Shuffle(len(items), func(i, j int) { items[i], items[j] = items[j], items[i] })
	return slices.Values(items)
}

func distinctBySeq[T any, K comparable](seq iter.Seq[T], keyFn func(T) K) iter.Seq[T] {
	return func(yield func(T) bool) {
		seen := make(map[K]struct{})
		for v := range seq {
			k := keyFn(v)
			if _, dup := seen[k]; dup {
				continue
			}
			seen[k] = struct{}{}
			if !yield(v) {
				return
			}
		}
	}
}

func exceptSeq[T comparable](seq, other iter.Seq[T]) iter.Seq[T] {
	exclude := make(map[T]struct{})
	for v := range other {
		exclude[v] = struct{}{}
	}
	return whereSeq(seq, func(v T) bool {
		_, found := exclude[v]
		return !found
	})
}

func exceptBySeq[T any, K comparable](seq, other iter.Seq[T], keyFn func(T) K) iter.Seq[T] {
	exclude := make(map[K]struct{})
	for v := range other {
		exclude[keyFn(v)] = struct{}{}
	}
	return whereSeq(seq, func(v T) bool {
		_, found := exclude[keyFn(v)]
		return !found
	})
}

func intersectSeq[T comparable](seq, other iter.Seq[T]) iter.Seq[T] {
	inOther := make(map[T]struct{})
	for v := range other {
		inOther[v] = struct{}{}
	}
	seen := make(map[T]struct{})
	return func(yield func(T) bool) {
		for v := range seq {
			if _, ok := inOther[v]; !ok {
				continue
			}
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

func intersectBySeq[T any, K comparable](seq, other iter.Seq[T], keyFn func(T) K) iter.Seq[T] {
	inOther := make(map[K]struct{})
	for v := range other {
		inOther[keyFn(v)] = struct{}{}
	}
	seen := make(map[K]struct{})
	return func(yield func(T) bool) {
		for v := range seq {
			k := keyFn(v)
			if _, ok := inOther[k]; !ok {
				continue
			}
			if _, dup := seen[k]; dup {
				continue
			}
			seen[k] = struct{}{}
			if !yield(v) {
				return
			}
		}
	}
}

func unionSeq[T comparable](seq, other iter.Seq[T]) iter.Seq[T] {
	return distinctSeq(concatSeq(seq, other))
}

func unionBySeq[T any, K comparable](seq, other iter.Seq[T], keyFn func(T) K) iter.Seq[T] {
	return distinctBySeq(concatSeq(seq, other), keyFn)
}
