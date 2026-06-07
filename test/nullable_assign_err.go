// errorcheck

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// T? is not assignable to T without an explicit unwrap.

package main

func myPrint(a int) {}

func main() {
	var a int?
	myPrint(a) // ERROR "cannot use"
}
