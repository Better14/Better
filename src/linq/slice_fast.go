// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

import (
	"cmp"
	"iter"
	"slices"
)

func whereSlice[T any](s []T, pred func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range s {
			if pred(v) && !yield(v) {
				return
			}
		}
	}
}

func selectSlice[T, U any](s []T, fn func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for _, v := range s {
			if !yield(fn(v)) {
				return
			}
		}
	}
}

func selectBySlice[T, U any](s []T, fn func(T) U) iter.Seq[U] {
	return selectSlice(s, fn)
}

func takeSlice[T any](s []T, n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		if n <= 0 {
			return
		}
		for _, v := range s {
			if n <= 0 {
				return
			}
			n--
			if !yield(v) {
				return
			}
		}
	}
}

func skipSlice[T any](s []T, n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		if n < 0 {
			n = 0
		}
		for _, v := range s {
			if n > 0 {
				n--
				continue
			}
			if !yield(v) {
				return
			}
		}
	}
}

func firstSlice[T any](s []T) (T, bool) {
	if len(s) == 0 {
		var z T
		return z, false
	}
	return s[0], true
}

func firstOrDefaultSlice[T any](s []T) T {
	if v, ok := firstSlice(s); ok {
		return v
	}
	var z T
	return z
}

func lastSlice[T any](s []T) (T, bool) {
	if len(s) == 0 {
		var z T
		return z, false
	}
	return s[len(s)-1], true
}

func lastOrDefaultSlice[T any](s []T) T {
	if v, ok := lastSlice(s); ok {
		return v
	}
	var z T
	return z
}

func countSlice[T any](s []T) int {
	return len(s)
}

func anySlice[T any](s []T, pred func(T) bool) bool {
	for _, v := range s {
		if pred(v) {
			return true
		}
	}
	return false
}

func allSlice[T any](s []T, pred func(T) bool) bool {
	for _, v := range s {
		if !pred(v) {
			return false
		}
	}
	return true
}

func toListSlice[T any](s []T) []T {
	return slices.Clone(s)
}

func containsSlice[T comparable](s []T, value T) bool {
	for _, v := range s {
		if v == value {
			return true
		}
	}
	return false
}

func elementAtSlice[T any](s []T, index int) (T, bool) {
	if index < 0 || index >= len(s) {
		var z T
		return z, false
	}
	return s[index], true
}

func sumSlice[U Number](s []U) U {
	var acc U
	for _, v := range s {
		acc += v
	}
	return acc
}

func averageSlice[U Number](s []U) float64 {
	if len(s) == 0 {
		panic("linq: sequence contains no elements")
	}
	var sum float64
	for _, v := range s {
		sum += float64(v)
	}
	return sum / float64(len(s))
}

func maxSlice[T cmp.Ordered](s []T) T {
	if len(s) == 0 {
		panic("linq: sequence contains no elements")
	}
	max := s[0]
	for _, v := range s[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

func minSlice[T cmp.Ordered](s []T) T {
	if len(s) == 0 {
		panic("linq: sequence contains no elements")
	}
	min := s[0]
	for _, v := range s[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

func maxBySlice[T any, K cmp.Ordered](s []T, key func(T) K) T {
	if len(s) == 0 {
		panic("linq: sequence contains no elements")
	}
	best := s[0]
	bestKey := key(best)
	for _, v := range s[1:] {
		k := key(v)
		if k > bestKey {
			best = v
			bestKey = k
		}
	}
	return best
}

func minBySlice[T any, K cmp.Ordered](s []T, key func(T) K) T {
	if len(s) == 0 {
		panic("linq: sequence contains no elements")
	}
	best := s[0]
	bestKey := key(best)
	for _, v := range s[1:] {
		k := key(v)
		if k < bestKey {
			best = v
			bestKey = k
		}
	}
	return best
}

func aggregateSlice[T any](s []T, fn func(T, T) T) T {
	if len(s) == 0 {
		panic("linq: sequence contains no elements")
	}
	acc := s[0]
	for _, v := range s[1:] {
		acc = fn(acc, v)
	}
	return acc
}

func aggregateWithSeedSlice[T any, U any](s []T, seed U, fn func(U, T) U) U {
	acc := seed
	for _, v := range s {
		acc = fn(acc, v)
	}
	return acc
}

func reverseSlice[T any](s []T) iter.Seq[T] {
	out := slices.Clone(s)
	slices.Reverse(out)
	return slices.Values(out)
}

func distinctSlice[T comparable](s []T) iter.Seq[T] {
	return func(yield func(T) bool) {
		seen := make(map[T]struct{})
		for _, v := range s {
			if _, dup := seen[v]; dup {
				continue
			}
			seen[v] = struct{}{}
			if !yield(v) {
				return
			}
		}
	}
}

func firstFromSlice[T any](s []T) T {
	v, ok := firstSlice(s)
	if !ok {
		panic("linq: sequence contains no elements")
	}
	return v
}

func lastFromSlice[T any](s []T) T {
	v, ok := lastSlice(s)
	if !ok {
		panic("linq: sequence contains no elements")
	}
	return v
}

func singleFromSlice[T any](s []T) T {
	if len(s) != 1 {
		if len(s) == 0 {
			panic("linq: sequence contains no elements")
		}
		panic("linq: sequence contains more than one element")
	}
	return s[0]
}

func singleOrDefaultFromSlice[T any](s []T, defaultValue T) T {
	if len(s) == 0 {
		return defaultValue
	}
	if len(s) > 1 {
		panic("linq: sequence contains more than one element")
	}
	return s[0]
}

func elementAtFromSlice[T any](s []T, index int) T {
	v, ok := elementAtSlice(s, index)
	if !ok {
		panic("linq: index out of range")
	}
	return v
}

func elementAtOrDefaultFromSlice[T any](s []T, index int, defaultValue T) T {
	if v, ok := elementAtSlice(s, index); ok {
		return v
	}
	return defaultValue
}

func longCountSlice[T any](s []T) int64 {
	return int64(countSlice(s))
}

func toHashSetSlice[T comparable](s []T) HashSet[T] {
	set := make(HashSet[T])
	for _, v := range s {
		set[v] = struct{}{}
	}
	return set
}

func tryGetNonEnumeratedCountSlice[T any](s []T) (int, bool) {
	return len(s), true
}
