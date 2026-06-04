// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"linkedlist"
	"list"
	"maxheap"
	"minheap"
	"queue"
	"set"
	"stack"
	"tree"
)

func main() {
	nums := list.Of(1, 2, 3)
	nums.Append(4)
	if nums.Len() != 4 || nums.At(0) != 1 {
		panic("list")
	}

	s := set.Of("a", "b", "a")
	if s.Len() != 2 || !s.Contains("a") {
		panic("set")
	}

	q := queue.New[int]()
	q.Enqueue(1)
	q.Enqueue(2)
	if v, _ := q.Dequeue(); v != 1 {
		panic("queue")
	}

	st := stack.New[int]()
	st.Push(10)
	if v, _ := st.Pop(); v != 10 {
		panic("stack")
	}

	h := minheap.New[int]()
	h.Push(5)
	h.Push(1)
	if v, _ := h.Pop(); v != 1 {
		panic("minheap")
	}

	mh := maxheap.New[int]()
	mh.Push(1)
	mh.Push(5)
	if v, _ := mh.Pop(); v != 5 {
		panic("maxheap")
	}

	t := tree.New[int, string]()
	t.Insert(2, "two")
	t.Insert(1, "one")
	if v, ok := t.Search(1); !ok || v != "one" {
		panic("tree")
	}

	ll := linkedlist.FromSlice([]int{1, 2})
	if v, _ := ll.PopFront(); v != 1 {
		panic("linkedlist")
	}
}
