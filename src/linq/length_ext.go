// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

import (
	"cmp"
	"iter"
	"list"
	"maxheap"
	"minheap"
	"set"
	"tree"
)

// Fast O(1) count helpers for types with known length.
// Use CountSlice/CountMap on slices and maps; CountList/CountSet/etc. on collections;
// or CountLength for any ILength value.

func CountSlice[T any](s []T) int {
	return len(s)
}

func LongCountSlice[T any](s []T) int64 {
	return int64(len(s))
}

func TryGetSeqLenSlice[T any](s []T) (int, bool) {
	return len(s), true
}

func CountMap[K comparable, V any](m map[K]V) int {
	return len(m)
}

func LongCountMap[K comparable, V any](m map[K]V) int64 {
	return int64(len(m))
}

func TryGetSeqLenMap[K comparable, V any](m map[K]V) (int, bool) {
	return len(m), true
}

func CountList[T any](l *list.List[T]) int {
	return l.Len()
}

func LongCountList[T any](l *list.List[T]) int64 {
	return int64(l.Len())
}

func TryGetSeqLenList[T any](l *list.List[T]) (int, bool) {
	return l.Len(), true
}

func ListAsSeq[T any](l *list.List[T]) iter.Seq[T] {
	return l.All()
}

func CountSet[T comparable](s set.Set[T]) int {
	return s.Len()
}

func LongCountSet[T comparable](s set.Set[T]) int64 {
	return int64(s.Len())
}

func TryGetSeqLenSet[T comparable](s set.Set[T]) (int, bool) {
	return s.Len(), true
}

func SetAsSeq[T comparable](s set.Set[T]) iter.Seq[T] {
	return s.All()
}

func CountTree[K cmp.Ordered, V any](t *tree.Tree[K, V]) int {
	return t.Len()
}

func LongCountTree[K cmp.Ordered, V any](t *tree.Tree[K, V]) int64 {
	return int64(t.Len())
}

func TryGetSeqLenTree[K cmp.Ordered, V any](t *tree.Tree[K, V]) (int, bool) {
	return t.Len(), true
}

func TreeAsSeq[K cmp.Ordered, V any](t *tree.Tree[K, V]) iter.Seq[struct{ Key K; Value V }] {
	return t.All()
}

func CountMinHeap[T cmp.Ordered](h *minheap.Heap[T]) int {
	return h.Len()
}

func LongCountMinHeap[T cmp.Ordered](h *minheap.Heap[T]) int64 {
	return int64(h.Len())
}

func TryGetSeqLenMinHeap[T cmp.Ordered](h *minheap.Heap[T]) (int, bool) {
	return h.Len(), true
}

func MinHeapAsSeq[T cmp.Ordered](h *minheap.Heap[T]) iter.Seq[T] {
	return h.All()
}

func CountMaxHeap[T cmp.Ordered](h *maxheap.Heap[T]) int {
	return h.Len()
}

func LongCountMaxHeap[T cmp.Ordered](h *maxheap.Heap[T]) int64 {
	return int64(h.Len())
}

func TryGetSeqLenMaxHeap[T cmp.Ordered](h *maxheap.Heap[T]) (int, bool) {
	return h.Len(), true
}

func MaxHeapAsSeq[T cmp.Ordered](h *maxheap.Heap[T]) iter.Seq[T] {
	return h.All()
}

// CountLength returns the element count for any value with Len() int in O(1).
func CountLength(l interface{ Len() int }) int {
	return l.Len()
}
