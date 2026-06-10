// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

// KeyValue is a key/value pair, as produced by CountBy and join operators.
type KeyValue[K any, V any] struct {
	Key   K
	Value V
}

// Indexed pairs an element with its zero-based index.
type Indexed[T any] struct {
	Index int
	Value T
}

// Lookup maps keys to sequences of values (immutable after construction).
type Lookup[K comparable, V any] struct {
	groups map[K][]V
	keys   []K
}

// Contains reports whether key exists in the lookup.
func (lk Lookup[K, V]) Contains(key K) bool {
	_, ok := lk.groups[key]
	return ok
}

// Get returns values for key, or nil if missing.
func (lk Lookup[K, V]) Get(key K) []V {
	return lk.groups[key]
}

// Count returns the number of keys in the lookup.
func (lk Lookup[K, V]) Count() int { return len(lk.keys) }

// Keys returns all keys in insertion order.
func (lk Lookup[K, V]) Keys() []K { return append([]K(nil), lk.keys...) }

// HashSet is a set of comparable values.
type HashSet[T comparable] map[T]struct{}

// Ordered is a sorted sequence supporting ThenBy / ThenByDescending.
type Ordered[T any] struct {
	items []T
	less  func(a, b T) int
}
