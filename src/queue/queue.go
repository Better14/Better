// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package queue provides a FIFO queue.
package queue

// Queue is a slice-backed FIFO queue.
type Queue[T any] struct {
	data []T
}

func New[T any]() *Queue[T] { return &Queue[T]{} }

func (q *Queue[T]) Enqueue(v T) { q.data = append(q.data, v) }

func (q *Queue[T]) Dequeue() (T, bool) {
	if len(q.data) == 0 {
		var z T
		return z, false
	}
	v := q.data[0]
	q.data = q.data[1:]
	return v, true
}

func (q *Queue[T]) Peek() (T, bool) {
	if len(q.data) == 0 {
		var z T
		return z, false
	}
	return q.data[0], true
}

func (q *Queue[T]) Len() int { return len(q.data) }
