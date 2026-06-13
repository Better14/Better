// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

import "iter"

// Empty returns an empty sequence.
func Empty[T any]() iter.Seq[T] {
	return func(yield func(T) bool) {}
}

// Range generates count consecutive integers starting at start.
func Range(start, count int) iter.Seq[int] {
	if count <= 0 {
		return Empty[int]()
	}
	end := start + count
	return func(yield func(int) bool) {
		for i := start; i < end; i++ {
			if !yield(i) {
				return
			}
		}
	}
}

// Repeat generates count copies of element.
func Repeat[T any](element T, count int) iter.Seq[T] {
	if count <= 0 {
		return Empty[T]()
	}
	return func(yield func(T) bool) {
		for i := 0; i < count; i++ {
			if !yield(element) {
				return
			}
		}
	}
}

// InfiniteSequence generates an infinite sequence using generator.
func InfiniteSequence[T any](generator func() T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for {
			if !yield(generator()) {
				return
			}
		}
	}
}

// Sequence generates an infinite sequence from seed using next.
func Sequence[T any](seed T, next func(T) T) iter.Seq[T] {
	return func(yield func(T) bool) {
		cur := seed
		for {
			if !yield(cur) {
				return
			}
			cur = next(cur)
		}
	}
}
