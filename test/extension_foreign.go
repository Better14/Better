// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"person"
	"personext"
)

func main() {
	a := person.Person{Name: "Ada"}
	if a.Hello() != "Hi, my name is Ada" {
		panic("extension call failed")
	}
}
