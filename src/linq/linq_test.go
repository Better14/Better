// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq_test

import (
	"linq"
	"testing"
)

func TestLazyPipeline(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	pipe := linq.LazySelect(linq.LazyWhere(linq.FromSlice(nums),
		func(n int) bool { return n%2 == 0 }),
		func(n int) int { return n * 2 })
	first, ok := linq.LazyFirst(pipe)
	if !ok || first != 4 {
		t.Fatalf("got (%v, %v), want (4, true)", first, ok)
	}
}

func TestWhereSum(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	evens := linq.Where(nums, func(n int) bool { return n%2 == 0 })
	if linq.Sum(evens) != 6 {
		t.Fatalf("sum evens: %v", evens)
	}
}
