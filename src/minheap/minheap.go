// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package minheap

import (
	"cmp"
	"iter"
)

type Heap[T cmp.Ordered] struct {
	data []T
}

func New[T cmp.Ordered]() *Heap[T] { return &Heap[T]{} }

func (h *Heap[T]) Len() int { return len(h.data) }

// All returns an iterator over heap elements in arbitrary order.
func (h *Heap[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range h.data {
			if !yield(v) {
				return
			}
		}
	}
}

func (h *Heap[T]) Push(v T) {
	h.data = append(h.data, v)
	h.up(len(h.data) - 1)
}

func (h *Heap[T]) Pop() (T, bool) {
	if len(h.data) == 0 {
		var z T
		return z, false
	}
	v := h.data[0]
	last := len(h.data) - 1
	h.data[0] = h.data[last]
	h.data = h.data[:last]
	if len(h.data) > 0 {
		h.down(0)
	}
	return v, true
}

func (h *Heap[T]) Peek() (T, bool) {
	if len(h.data) == 0 {
		var z T
		return z, false
	}
	return h.data[0], true
}

func (h *Heap[T]) up(i int) {
	for i > 0 {
		p := (i - 1) / 2
		if h.data[p] <= h.data[i] {
			break
		}
		h.data[p], h.data[i] = h.data[i], h.data[p]
		i = p
	}
}

func (h *Heap[T]) down(i int) {
	n := len(h.data)
	for {
		l, r, s := 2*i+1, 2*i+2, i
		if l < n && h.data[l] < h.data[s] {
			s = l
		}
		if r < n && h.data[r] < h.data[s] {
			s = r
		}
		if s == i {
			return
		}
		h.data[i], h.data[s] = h.data[s], h.data[i]
		i = s
	}
}
