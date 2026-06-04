// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

import (
	"cmp"
	"slices"
)

// Where filters s, returning a new slice.
func Where[T any](s []T, pred func(T) bool) []T {
	return LazyToSlice(LazyWhere(FromSlice(s), pred))
}

// Select maps s with fn.
func Select[T any](s []T, fn func(T) T) []T {
	return LazyToSlice(LazySelect(FromSlice(s), fn))
}

// SelectBy maps s to a new element type.
func SelectBy[T, U any](s []T, fn func(T) U) []U {
	return LazyToSlice(LazySelectBy(FromSlice(s), fn))
}

// First returns the first element of s.
func First[T any](s []T) (T, bool) {
	if len(s) == 0 {
		var z T
		return z, false
	}
	return s[0], true
}


type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
	~float32 | ~float64
}

func Sum[U Number](s []U) U {
	var acc U
	for _, v := range s {
		acc += v
	}
	return acc
}

func OrderBy[T any, K cmp.Ordered](s []T, key func(T) K) []T {
	out := slices.Clone(s)
	slices.SortFunc(out, func(a, b T) int { return cmp.Compare(key(a), key(b)) })
	return out
}

func DistinctComparable[T comparable](s []T) []T {
	seen := make(map[T]struct{})
	var out []T
	for _, v := range s {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
