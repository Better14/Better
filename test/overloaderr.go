// errorcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test overload resolution errors.

package main

type A int
type B int

func addInt(a int) int     { return 0 }
func addInt(a int64) int64 { return 0 }

func addAmbig(a A) int { return 0 }
func addAmbig(a B) int { return 0 }

func dup(int)     {}
func dup(int) {} // ERROR "redeclared function dup"

func badNoMatch() {
	addInt("x") // ERROR "no matching overload"
}

func badAmbig() {
	addAmbig(1) // ERROR "ambiguous overloaded"
}

func main() {}
