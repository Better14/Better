// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package set provides an unordered collection of unique comparable elements.
package set

// Set is a hash set.
type Set[T comparable] map[T]struct{}

func Make[T comparable]() Set[T] { return make(Set[T]) }

func Of[T comparable](vals ...T) Set[T] {
	s := Make[T]()
	for _, v := range vals {
		s.Add(v)
	}
	return s
}

func (s Set[T]) Add(v T) { s[v] = struct{}{} }

func (s Set[T]) Delete(v T) { delete(s, v) }

func (s Set[T]) Contains(v T) bool { _, ok := s[v]; return ok }

func (s Set[T]) Len() int { return len(s) }

func (s Set[T]) Values() []T {
	out := make([]T, 0, len(s))
	for v := range s {
		out = append(out, v)
	}
	return out
}

func (a Set[T]) Union(b Set[T]) Set[T] {
	out := Make[T]()
	for v := range a {
		out.Add(v)
	}
	for v := range b {
		out.Add(v)
	}
	return out
}

func (a Set[T]) Intersect(b Set[T]) Set[T] {
	out := Make[T]()
	for v := range a {
		if b.Contains(v) {
			out.Add(v)
		}
	}
	return out
}

func (a Set[T]) Equal(b Set[T]) bool {
	if len(a) != len(b) {
		return false
	}
	for v := range a {
		if !b.Contains(v) {
			return false
		}
	}
	return true
}
