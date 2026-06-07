// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package linq provides LINQ-style query operations with lazy evaluation.
//
// Import this package to use C#-style method chains on slices and lazy sequences:
//
//	import "linq"
//
//	nums := []int{1, 2, 3, 4, 5}
//	out := nums.Where(n => n < 5).Select(n => n + 1).ToList()
//
// Pipelines are lazy until a terminal operator (First, ToList, Sum, Any, All, etc.).
package linq
