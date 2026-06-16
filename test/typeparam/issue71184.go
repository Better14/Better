// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Generic method calls that return pointer-shaped type parameters must be
// reshaped back to their concrete type before field selection.

package main

type connResource struct {
	maxAgeTime int64
}

type Resource[T any] struct {
	value T
}

func (res *Resource[T]) Value() T {
	return res.value
}

func isExpired(res *Resource[*connResource]) bool {
	return res.Value().maxAgeTime > 0
}

func main() {
	r := &Resource[*connResource]{value: &connResource{maxAgeTime: 1}}
	if !isExpired(r) {
		panic("expected expired")
	}
}
