// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

// Group is one key and its elements, as produced by GroupBy.
type Group[K comparable, V any] struct {
	Key K
	items []V
}

// Count returns the number of elements in the group.
func (g Group[K, V]) Count() int { return len(g.items) }

// Values returns the elements in the group.
func (g Group[K, V]) Values() []V { return g.items }
