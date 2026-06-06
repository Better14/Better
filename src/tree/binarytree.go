// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package tree provides container types; this file implements an ordered binary search tree.
package tree

import "cmp"

type node[K cmp.Ordered, V any] struct {
	key   K
	val   V
	left  *node[K, V]
	right *node[K, V]
}

// Tree is a binary search tree keyed by K with values V.
type Tree[K cmp.Ordered, V any] struct {
	root *node[K, V]
	size int
}

func New[K cmp.Ordered, V any]() *Tree[K, V] { return &Tree[K, V]{} }

func (t *Tree[K, V]) Len() int { return t.size }

func (t *Tree[K, V]) Insert(key K, val V) {
	t.root = t.insert(t.root, key, val)
}

func (t *Tree[K, V]) insert(n *node[K, V], key K, val V) *node[K, V] {
	if n == nil {
		t.size++
		return &node[K, V]{key: key, val: val}
	}
	switch {
	case key < n.key:
		n.left = t.insert(n.left, key, val)
	case key > n.key:
		n.right = t.insert(n.right, key, val)
	default:
		n.val = val
	}
	return n
}

func (t *Tree[K, V]) Search(key K) (V, bool) {
	n := t.root
	for n != nil {
		switch {
		case key < n.key:
			n = n.left
		case key > n.key:
			n = n.right
		default:
			return n.val, true
		}
	}
	var z V
	return z, false
}

func (t *Tree[K, V]) Inorder(fn func(K, V)) {
	var walk func(*node[K, V])
	walk = func(n *node[K, V]) {
		if n == nil {
			return
		}
		walk(n.left)
		fn(n.key, n.val)
		walk(n.right)
	}
	walk(t.root)
}

func (t *Tree[K, V]) Delete(key K) {
	t.root = t.delete(t.root, key)
}

func (t *Tree[K, V]) delete(n *node[K, V], key K) *node[K, V] {
	if n == nil {
		return nil
	}
	switch {
	case key < n.key:
		n.left = t.delete(n.left, key)
	case key > n.key:
		n.right = t.delete(n.right, key)
	default:
		t.size--
		if n.left == nil {
			return n.right
		}
		if n.right == nil {
			return n.left
		}
		succ := n.right
		for succ.left != nil {
			succ = succ.left
		}
		n.key, n.val = succ.key, succ.val
		n.right = t.delete(n.right, succ.key)
		t.size++ // delete succ compensated
	}
	return n
}
