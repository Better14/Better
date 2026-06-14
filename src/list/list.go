// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package list provides a growable ordered sequence with in-place mutation.
package list

import "iter"

// List is a slice-backed growable sequence.
type List[T any] struct {
	data []T
}

func New[T any]() *List[T] { return &List[T]{} }

func Of[T any](vals ...T) *List[T] {
	l := &List[T]{}
	l.Append(vals...)
	return l
}

func FromSlice[T any](s []T) *List[T] {
	l := &List[T]{}
	l.Append(s...)
	return l
}

// ensureCapacity grows the backing slice when len+additional would exceed cap.
// When already at capacity, the new capacity is double the old (minimum 1).
func (l *List[T]) ensureCapacity(additional int) {
	need := len(l.data) + additional
	if cap(l.data) >= need {
		return
	}
	newCap := cap(l.data)
	if newCap == 0 {
		newCap = 1
	}
	for newCap < need {
		newCap *= 2
	}
	buf := make([]T, len(l.data), newCap)
	copy(buf, l.data)
	l.data = buf
}

func (l *List[T]) Append(vals ...T) {
	if len(vals) == 0 {
		return
	}
	l.ensureCapacity(len(vals))
	l.data = append(l.data, vals...)
}

// AddRange appends all elements from r.
func (l *List[T]) AddRange(r iter.Seq[T]) {
	for v := range r {
		l.Append(v)
	}
}

func (l *List[T]) Len() int { return len(l.data) }

// All returns an iterator over the list elements in order.
func (l *List[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range l.data {
			if !yield(v) {
				return
			}
		}
	}
}

func (l *List[T]) Cap() int { return cap(l.data) }

func (l *List[T]) At(i int) T { return l.data[i] }

func (l *List[T]) Set(i int, v T) { l.data[i] = v }

// IndexOf returns the index of the first element equal to v, or -1.
func IndexOf[T comparable](l *List[T], v T) int {
	for i, x := range l.data {
		if x == v {
			return i
		}
	}
	return -1
}

// Contains reports whether v is present.
func Contains[T comparable](l *List[T], v T) bool {
	return IndexOf(l, v) >= 0
}

func (l *List[T]) ToSlice() []T { return append([]T(nil), l.data...) }

func (l *List[T]) Clear() { l.data = nil }

func (l *List[T]) Insert(i int, v T) {
	l.ensureCapacity(1)
	l.data = append(l.data, *new(T))
	copy(l.data[i+1:], l.data[i:])
	l.data[i] = v
}

func (l *List[T]) RemoveAt(i int) {
	copy(l.data[i:], l.data[i+1:])
	l.data = l.data[:len(l.data)-1]
}
