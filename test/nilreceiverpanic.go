// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Nil pointer method calls panic at the call site in Bow.

package main

type T struct {
	v int
}

func (t *T) M() int {
	return t.v
}

func main() {
	var p *T
	defer func() {
		if recover() == nil {
			panic("expected panic from nil receiver method call")
		}
	}()
	p.M()
}
