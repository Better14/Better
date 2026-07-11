
// Package linq provides LINQ-style query operations with lazy evaluation.
//
// Import this package to use C#-style method chains on iter.Seq[T] sequences:
//
//	import "linq"
//
//	nums := []int{1, 2, 3, 4, 5}
//	out := nums.Where(n => n < 5).Select(n => n + 1).ToSlice()
//
// Slice receivers on the first call in a chain use optimized implementations
// inside the linq package (compiler specialization).
//
// Pipelines are lazy until a terminal operator (First, ToSlice, Sum, Any, All, etc.).
package linq
