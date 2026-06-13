// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package length defines ILength for types with a known element count.
package length

// ILength is implemented by collections that expose a cheap Len() int.
// Slices, maps, list.List, set.Set, tree.Tree, and min/max heap types
// implement this interface.
type ILength interface {
	Len() int
}
