// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package length

// Len returns the number of elements in s.
func (s []T) Len[T any]() int {
	return len(s)
}

// MapLen returns the number of entries in m.
// Named separately from Len because slice and map extensions cannot share
// an overloaded method name until the compiler assigns distinct link symbols.
func (m map[K]V) MapLen[K comparable, V any]() int {
	return len(m)
}

// Len returns the number of entries in m (implements ILength via MapLen).
func MapLen[K comparable, V any](m map[K]V) int {
	return len(m)
}
