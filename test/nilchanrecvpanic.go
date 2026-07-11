// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Receiving from a nil channel panics in Bow instead of blocking forever.

package main

func main() {
	var ch chan int
	defer func() {
		if recover() == nil {
			panic("expected panic from nil channel receive")
		}
	}()
	<-ch
}
