// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"length"
	"linq"
)

func main() {
	s := []int{1, 2, 3}
	if s.Len() != 3 {
		panic("slice Len")
	}
	m := map[string]int{"a": 1, "b": 2}
	if length.MapLen(m) != 2 {
		panic("map Len")
	}
	if linq.CountMap(m) != 2 {
		panic("map CountMap")
	}
	if linq.From(s).Count() != 3 {
		panic("seq Count")
	}
	fmt.Println("ok")
}
