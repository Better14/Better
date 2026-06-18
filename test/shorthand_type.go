// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

struct Person {
	Name string
	Age  int
}

interface Stringer {
	String() string
}

func (p Person) String() string {
	return p.Name
}

func main() {
	p := Person{Name: "Ada", Age: 42}
	var s Stringer = p
	if s.String() != "Ada" {
		panic(s.String())
	}
	println("ok")
}
