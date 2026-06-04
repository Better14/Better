// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package linq provides LINQ-style query operations with lazy evaluation.
//
// Lazy pipelines compose with LazyWhere, LazySelect, and terminals such as
// LazyFirst or LazyToSlice. Eager helpers include Where, Select, Sum, and OrderBy.
//
// Example:
//
//	nums := []int{1, 2, 3, 4, 5}
//	first, ok := linq.LazyFirst(linq.LazySelect(
//	    linq.LazyWhere(linq.FromSlice(nums), func(n int) bool { return n%2 == 0 }),
//	    func(n int) int { return n * 2 },
//	))
package linq
