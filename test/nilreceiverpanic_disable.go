// run -gcflags=all=-nilreceiverpanic=disable

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// With nil_receiver_panic disabled, nil pointer receivers may run method bodies.

package main

type T struct{}

func (t *T) M() int {
	if t == nil {
		return 42
	}
	return 1
}

func main() {
	var p *T
	if v := p.M(); v != 42 {
		panic(v)
	}
}
