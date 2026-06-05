// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

import (
	"cmp"
	"slices"
)

// Number is a type set for numeric Sum.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
	~float32 | ~float64
}

// sortOrderBy sorts s by key (eager materialization helper).
func sortOrderBy[T any, K cmp.Ordered](s []T, key func(T) K) []T {
	out := slices.Clone(s)
	slices.SortFunc(out, func(a, b T) int { return cmp.Compare(key(a), key(b)) })
	return out
}

// DistinctComparable returns distinct elements of s (eager).
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
