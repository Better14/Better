// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

// Group is one key and its elements, as produced by GroupBy.
type Group[K comparable, T any] struct {
	Key K
	items []T
}

// Count returns the number of elements in the group.
func (g Group[K, T]) Count() int { return len(g.items) }

// Values returns the elements in the group.
func (g Group[K, T]) Values() []T { return g.items }
