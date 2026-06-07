// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func (i int) Twice() int { return i + i }

func main() {
	x := 21
	if x.Twice() != 42 {
		panic("extension int failed")
	}
}
