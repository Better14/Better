// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test an internal compiler error on ? symbol in declaration
// following an empty import.

package a
var?      // ERROR "invalid character U\+003F '\?'|invalid character 0x3f in input file|unexpected \?, expected name|syntax error: unexpected \?, expected name"

var x int

func main() {
}
