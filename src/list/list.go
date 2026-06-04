// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package list provides a growable ordered sequence with in-place mutation.
package list

// List is a slice-backed growable sequence.
type List[T any] struct {
	data []T
}

func New[T any]() *List[T] { return &List[T]{} }

func Of[T any](vals ...T) *List[T] { return &List[T]{data: append([]T(nil), vals...)} }

func FromSlice[T any](s []T) *List[T] { return &List[T]{data: append([]T(nil), s...)} }

func (l *List[T]) Append(vals ...T) { l.data = append(l.data, vals...) }

func (l *List[T]) Len() int { return len(l.data) }

func (l *List[T]) At(i int) T { return l.data[i] }

func (l *List[T]) Set(i int, v T) { l.data[i] = v }

func (l *List[T]) ToSlice() []T { return append([]T(nil), l.data...) }

func (l *List[T]) Clear() { l.data = nil }

func (l *List[T]) Pop() (T, bool) {
	if len(l.data) == 0 {
		var z T
		return z, false
	}
	i := len(l.data) - 1
	v := l.data[i]
	l.data = l.data[:i]
	return v, true
}

func (l *List[T]) Insert(i int, v T) {
	l.data = append(l.data, *new(T))
	copy(l.data[i+1:], l.data[i:])
	l.data[i] = v
}

func (l *List[T]) RemoveAt(i int) {
	copy(l.data[i:], l.data[i+1:])
	l.data = l.data[:len(l.data)-1]
}
