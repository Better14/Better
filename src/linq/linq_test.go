// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq_test

import (
	"linq"
	"slices"
	"testing"
)

func TestChainWhereSelectToList(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	out := nums.Where(func(n int) bool { return n < 5 }).Select(func(n int) int { return n + 1 }).ToList()
	want := []int{2, 3, 4, 5}
	if !slices.Equal(out, want) {
		t.Fatalf("got %v, want %v", out, want)
	}
}

func TestChainFirst(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	first := nums.Where(func(n int) bool { return n%2 == 0 }).Select(func(n int) int { return n * 2 }).First()
	if first != 4 {
		t.Fatalf("got %v, want 4", first)
	}
}

func TestChainLambdaArrow(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	out := nums.Where(func(n int) bool { return n < 5 }).Select(func(n int) int { return n + 1 }).ToList()
	want := []int{2, 3, 4, 5}
	if !slices.Equal(out, want) {
		t.Fatalf("got %v, want %v", out, want)
	}
}

func TestLazyPackageAPI(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	pipe := linq.LazySelect(linq.LazyWhere(linq.FromSlice(nums),
		func(n int) bool { return n%2 == 0 }),
		func(n int) int { return n * 2 })
	first, ok := linq.LazyFirst(pipe)
	if !ok || first != 4 {
		t.Fatalf("got (%v, %v), want (4, true)", first, ok)
	}
}

func TestLazyReceiverSelect(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	first := linq.FromSlice(nums).Where(func(n int) bool { return n%2 == 0 }).Select(func(n int) int { return n * 2 }).First()
	if first != 4 {
		t.Fatalf("got %v, want 4", first)
	}
}

func TestSumChain(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	sum := linq.LazySum(nums.Where(func(n int) bool { return n%2 == 0 }))
	if sum != 6 {
		t.Fatalf("sum: %v", sum)
	}
}
