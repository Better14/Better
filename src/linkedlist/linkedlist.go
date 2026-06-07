// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package linkedlist provides a generic doubly-linked list.
package linkedlist

type element[T any] struct {
	val        T
	prev, next *element[T]
}

// List is a doubly-linked list.
type List[T any] struct {
	head, tail *element[T]
	length     int
}

func New[T any]() *List[T] { return &List[T]{} }

func FromSlice[T any](s []T) *List[T] {
	l := New[T]()
	for _, v := range s {
		l.Append(v)
	}
	return l
}

func (l *List[T]) Len() int { return l.length }

func (l *List[T]) Append(vals ...T) {
	for _, v := range vals {
		e := &element[T]{val: v}
		if l.tail == nil {
			l.head, l.tail = e, e
		} else {
			e.prev = l.tail
			l.tail.next = e
			l.tail = e
		}
		l.length++
	}
}

func (l *List[T]) Prepend(v T) {
	e := &element[T]{val: v}
	if l.head == nil {
		l.head, l.tail = e, e
	} else {
		e.next = l.head
		l.head.prev = e
		l.head = e
	}
	l.length++
}

func (l *List[T]) PopFront() (T, bool) {
	if l.head == nil {
		var z T
		return z, false
	}
	v := l.head.val
	l.head = l.head.next
	if l.head == nil {
		l.tail = nil
	} else {
		l.head.prev = nil
	}
	l.length--
	return v, true
}

func (l *List[T]) PopBack() (T, bool) {
	if l.tail == nil {
		var z T
		return z, false
	}
	v := l.tail.val
	l.tail = l.tail.prev
	if l.tail == nil {
		l.head = nil
	} else {
		l.tail.next = nil
	}
	l.length--
	return v, true
}

func (l *List[T]) ToSlice() []T {
	out := make([]T, 0, l.length)
	for e := l.head; e != nil; e = e.next {
		out = append(out, e.val)
	}
	return out
}

// Values returns elements from head to tail.
func (l *List[T]) Values() []T { return l.ToSlice() }
