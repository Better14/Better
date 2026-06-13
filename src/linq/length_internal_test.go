// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

import (
	"list"
	"minheap"
	"set"
	"testing"
	"tree"
)

func TestSliceCountFast(t *testing.T) {
	nums := []int{1, 2, 3, 4}
	if CountSlice(nums) != 4 {
		t.Fatalf("slice CountSlice: %d", CountSlice(nums))
	}
	if LongCountSlice(nums) != 4 {
		t.Fatalf("slice LongCountSlice: %d", LongCountSlice(nums))
	}
	n, ok := TryGetSeqLenSlice(nums)
	if !ok || n != 4 {
		t.Fatalf("slice TryGetSeqLenSlice: (%d, %v)", n, ok)
	}
}

func TestMapCountFast(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	if CountMap(m) != 3 {
		t.Fatalf("map CountMap: %d", CountMap(m))
	}
}

func TestCollectionCountFast(t *testing.T) {
	l := list.Of(1, 2, 3)
	if CountList(l) != 3 {
		t.Fatalf("list CountList: %d", CountList(l))
	}
	s := set.Of(10, 20, 30, 40)
	if CountSet(s) != 4 {
		t.Fatalf("set CountSet: %d", CountSet(s))
	}
	tr := tree.New[int, string]()
	tr.Insert(1, "a")
	tr.Insert(2, "b")
	if CountTree(tr) != 2 {
		t.Fatalf("tree CountTree: %d", CountTree(tr))
	}
	h := minheap.New[int]()
	h.Push(1)
	h.Push(2)
	h.Push(3)
	if CountMinHeap(h) != 3 {
		t.Fatalf("heap CountMinHeap: %d", CountMinHeap(h))
	}
}

func TestCollectionAsSeqCount(t *testing.T) {
	l := list.Of(1, 2, 3)
	seq := ListAsSeq(l)
	if Count[int](seq) != 3 {
		t.Fatal("list AsSeq Count")
	}
	filtered := Where[int](seq, func(int) bool { return false })
	if Count[int](filtered) != 0 {
		t.Fatal("filtered Count")
	}
}

func TestCountLength(t *testing.T) {
	l := list.Of(1, 2)
	if CountLength(l) != 2 {
		t.Fatal("CountLength")
	}
}
