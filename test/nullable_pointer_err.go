// errorcheck

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// *T? must bind as (*T)? and fail because pointers are already nil-able.

package main

type MyStruct struct{}

func main() {
	var _ *MyStruct? = nil // ERROR "invalid nullable type"
	var _ error? = nil     // ERROR "invalid nullable type"
}
