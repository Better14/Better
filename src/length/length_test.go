// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package length_test

import (
	"length"
	"list"
	"maxheap"
	"minheap"
	"set"
	"testing"
	"tree"
)

func TestSliceMapLenExtension(t *testing.T) {
	s := []int{1, 2, 3}
	if s.Len() != 3 {
		t.Fatalf("slice Len: %d", s.Len())
	}
	m := map[string]int{"a": 1, "b": 2}
	if length.MapLen(m) != 2 {
		t.Fatalf("map MapLen: %d", length.MapLen(m))
	}
}

func TestCollectionsImplementILength(t *testing.T) {
	var _ length.ILength = (*list.List[int])(nil)
	var _ length.ILength = set.Set[int](nil)
	var _ length.ILength = (*tree.Tree[int, string])(nil)
	var _ length.ILength = (*minheap.Heap[int])(nil)
	var _ length.ILength = (*maxheap.Heap[int])(nil)
}
