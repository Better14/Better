// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

// Lazy is a deferred sequence: chain package functions until a terminal call.
type Lazy[T any] struct {
	next func() (T, bool)
}

// FromSlice returns a lazy iterator over s.
func FromSlice[T any](s []T) Lazy[T] {
	i := 0
	return Lazy[T]{next: func() (T, bool) {
		if i >= len(s) {
			var z T
			return z, false
		}
		v := s[i]
		i++
		return v, true
	}}
}

// Where filters l and returns a lazy sequence.
func (l Lazy[T]) Where(pred func(T) bool) Lazy[T] {
	return LazyWhere(l, pred)
}

func LazyWhere[T any](l Lazy[T], pred func(T) bool) Lazy[T] {
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

func LazySelect[T any](l Lazy[T], fn func(T) T) Lazy[T] {
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

func LazySelectBy[T, U any](l Lazy[T], fn func(T) U) Lazy[U] {
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

func LazySkip[T any](l Lazy[T], n int) Lazy[T] {
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

func LazyTake[T any](l Lazy[T], n int) Lazy[T] {
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

func LazyToSlice[T any](l Lazy[T]) []T {
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

func LazyFirst[T any](l Lazy[T]) (T, bool) {
	return l.next()
}

func LazyFirstOrDefault[T any](l Lazy[T], def T) T {
	v, ok := l.next()
	if ok {
		return v
	}
	return def
}

func LazyAny[T any](l Lazy[T], pred func(T) bool) bool {
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

func LazyAll[T any](l Lazy[T], pred func(T) bool) bool {
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

func LazyCount[T any](l Lazy[T]) int {
	n := 0
	for {
		if _, ok := l.next(); !ok {
			return n
		}
		n++
	}
}
