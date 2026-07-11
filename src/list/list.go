
// Package list provides a growable ordered sequence with in-place mutation.
//
// A List is backed by a slice but exposes methods so callers need not write
// s = append(s, x). Growth uses a 2× capacity strategy: whenever an append
// would exceed capacity, capacity doubles (minimum 1). Standard Go slices use
// runtime growth of 2× while capacity is below 256 elements, then ~1.25×
// (oldcap + oldcap/4) for larger capacities. list therefore reallocates less
// often on sustained growth at large sizes, at the cost of higher peak memory.
//
// Append and AddRange batch writes into the backing slice with a single capacity
// check per call; Of and FromSlice pre-size capacity to the source length.
package list

import "iter"

// List is a slice-backed growable sequence.
type List[T any] struct {
	data []T
}

// New returns an empty list.
func New[T any]() *List[T] { return &List[T]{} }

// Of returns a list containing vals, with capacity reserved for len(vals).
func Of[T any](vals ...T) *List[T] {
	if len(vals) == 0 {
		return &List[T]{}
	}
	return &List[T]{data: append(make([]T, 0, len(vals)), vals...)}
}

// FromSlice returns a list copy of s, with capacity reserved for len(s).
func FromSlice[T any](s []T) *List[T] {
	if len(s) == 0 {
		return &List[T]{}
	}
	return &List[T]{data: append(make([]T, 0, len(s)), s...)}
}

// growCap returns the smallest power-of-two capacity at least need, starting from cap (or 1).
func growCap(cap, need int) int {
	if cap == 0 {
		cap = 1
	}
	for cap < need {
		cap *= 2
	}
	return cap
}

func (l *List[T]) grow(need int) {
	newCap := growCap(cap(l.data), need)
	buf := make([]T, len(l.data), newCap)
	copy(buf, l.data)
	l.data = buf
}

// Append adds elements to the end. When the backing slice is full, capacity
// doubles before the append (see package documentation).
func (l *List[T]) Append(vals ...T) {
	n := len(vals)
	if n == 0 {
		return
	}
	need := len(l.data) + n
	if cap(l.data) < need {
		l.grow(need)
	}
	l.data = append(l.data, vals...)
}

// AddRange appends all elements from r in as few growth steps as possible.
func (l *List[T]) AddRange(r iter.Seq[T]) {
	buf := make([]T, 0, 8)
	for v := range r {
		buf = append(buf, v)
	}
	if len(buf) > 0 {
		l.Append(buf...)
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
	n := len(l.data)
	if i == n {
		l.Append(v)
		return
	}
	need := n + 1
	if cap(l.data) < need {
		l.grow(need)
	}
	l.data = append(l.data, *new(T))
	copy(l.data[i+1:], l.data[i:])
	l.data[i] = v
}

func (l *List[T]) RemoveAt(i int) {
	copy(l.data[i:], l.data[i+1:])
	l.data = l.data[:len(l.data)-1]
}
