// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package stack provides a LIFO stack.
package stack

type Stack[T any] struct {
	data []T
}

func New[T any]() *Stack[T] { return &Stack[T]{} }

func (s *Stack[T]) Push(v T) { s.data = append(s.data, v) }

func (s *Stack[T]) Pop() (T, bool) {
	if len(s.data) == 0 {
		var z T
		return z, false
	}
	i := len(s.data) - 1
	v := s.data[i]
	s.data = s.data[:i]
	return v, true
}

func (s *Stack[T]) Peek() (T, bool) {
	if len(s.data) == 0 {
		var z T
		return z, false
	}
	return s.data[len(s.data)-1], true
}

func (s *Stack[T]) Len() int { return len(s.data) }
