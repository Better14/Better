// errorcheck

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// err! is only valid in functions that return (T, error) or T!.

package main

import "errors"

func possibleError() error { return nil }

func noResultFunc() {
	err := possibleError()
	err! // ERROR "invalid operation"
}

func plainReturn() int {
	err := possibleError()
	err! // ERROR "invalid operation"
	return 1
}

func usedAsValue() int! {
	err := possibleError()
	_ = err! // ERROR "used as value"
	return 0
}

func wrongType() int! {
	x := 1
	x! // ERROR "invalid operation"
	return 0
}

func main() {
	_ = errors.New
}
