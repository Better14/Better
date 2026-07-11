
// Package length defines ILength for types with a known element count.
package length

// ILength is implemented by collections that expose a cheap Len() int.
// Slices, maps, list.List, set.Set, tree.BinaryTree, and min/max heap types
// implement this interface.
type ILength interface {
	Len() int
}
