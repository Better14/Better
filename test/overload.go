// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "fmt"

func add(a int) int {
	return a + 1
}

func add(a int64) int64 {
	return a + 2
}

type S struct{}

func (s S) m(a int) int {
	return a + 10
}

func (s S) m(a int64) int64 {
	return a + 20
}

func main() {
	if add(1) != 2 {
		panic("add(int)")
	}
	if add(int64(3)) != 5 {
		panic("add(int64)")
	}
	var s S
	if s.m(1) != 11 {
		panic("m(int)")
	}
	if s.m(int64(2)) != 22 {
		panic("m(int64)")
	}
	fmt.Println("ok")
}
