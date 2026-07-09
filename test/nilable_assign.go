// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test unwrapping int? to int via nil-check, force cast, and ??.

package main

import "fmt"

func intPtr(n int) *int {
	p := new(int)
	*p = n
	return p
}

func myPrint(a int) {
	fmt.Println(a)
}

func nilCheck() {
	var a int? = intPtr(7)
	if a != nil {
		myPrint(a)
	}
}

func forceCast() {
	var a int? = intPtr(8)
	myPrint(int(a))
}

func coalesce() {
	var a int?
	myPrint(a ?? 0)
}

func main() {
	nilCheck()
	forceCast()
	coalesce()
	fmt.Println("ok")
}
