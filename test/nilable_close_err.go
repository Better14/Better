// errorcheck

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:nilable_pointers enable

package main

func main() {
	var ch chan int?
	close(ch) // ERROR "without nil check"
	if ch != nil {
		close(ch)
	}
}
