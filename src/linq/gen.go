// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

// Empty returns an empty lazy sequence.
func Empty[T any]() Lazy[T] {
	return Lazy[T]{next: func() (T, bool) {
		var z T
		return z, false
	}}
}

// Range generates count consecutive integers starting at start.
func Range(start, count int) Lazy[int] {
	if count <= 0 {
		return Empty[int]()
	}
	i := 0
	return Lazy[int]{
		knownLen: count,
		next: func() (int, bool) {
			if i >= count {
				return 0, false
			}
			v := start + i
			i++
			return v, true
		},
	}
}

// Repeat generates count copies of element.
func Repeat[T any](element T, count int) Lazy[T] {
	if count <= 0 {
		return Empty[T]()
	}
	i := 0
	return Lazy[T]{
		knownLen: count,
		next: func() (T, bool) {
			if i >= count {
				var z T
				return z, false
			}
			i++
			return element, true
		},
	}
}

// InfiniteSequence generates an infinite sequence using generator.
func InfiniteSequence[T any](generator func() T) Lazy[T] {
	return Lazy[T]{next: func() (T, bool) {
		return generator(), true
	}}
}

// Sequence generates an infinite sequence from seed using next.
func Sequence[T any](seed T, next func(T) T) Lazy[T] {
	cur := seed
	first := true
	return Lazy[T]{next: func() (T, bool) {
		if first {
			first = false
			return cur, true
		}
		cur = next(cur)
		return cur, true
	}}
}
