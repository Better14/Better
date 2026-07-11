// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "fmt"

func main() {
	price := 12.5
	name := "widget"
	msg := "Price is {price:.2f} for {name}"
	want := fmt.Sprintf("Price is %.2f for %v", price, name)
	if msg != want {
		panic("interpolation mismatch")
	}
	literal := "braces: \\{not interpolated\\}"
	if literal != "braces: {not interpolated}" {
		panic("escape mismatch")
	}
}
