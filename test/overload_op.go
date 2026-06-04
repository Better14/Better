// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type Vec int

func +(a, b Vec) Vec { return a + b }
func ==(a, b Vec) bool { return a == b }
func !=(a, b Vec) bool { return a != b }

type Counter int

func ++(c Counter) Counter {
	c++
	return c
}

func main() {
	var v Vec = 1
	if v+v != 2 {
		panic(v + v)
	}
	var n Counter = 0
	n++
	if n != 1 {
		panic(n)
	}
}
